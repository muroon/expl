package expl

import (
	"context"
	"expl/model"
	"fmt"

	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Hosts     []*host     `yaml:"hosts"`
	Databases []*database `yaml:"databases"`
}

type host struct {
	Key      int    `yaml:"key"`
	Address  string `yaml:"address"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type database struct {
	HostKey int      `yaml:"hostkey"`
	Name    string   `yaml:"name"`
	Tables  []string `yaml:"tables"`
}

var dbInfo *model.DBInfo

// setConfig write config
func setConfig(ctx context.Context, conf *config, filePath string) error {

	filePath, err := getPath(filePath)
	if err != nil {
		return err
	}

	buf, err := yaml.Marshal(conf)
	if err != nil {
		return ErrWrap(err, UserInputError)
	}

	err = ioutil.WriteFile(filePath, buf, os.ModePerm)
	if err != nil {
		return ErrWrap(err, UserInputError)
	}

	return nil
}

// getConfig get config
func getConfig(ctx context.Context, filePath string) (*config, error) {
	// 外部からconfの中身を参照できるようにする
	var c config

	filePath, err := getPath(filePath)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, ErrWrap(err, UserInputError)
	}

	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		return nil, ErrWrap(err, UserInputError)
	}

	return &c, nil
}

func AddHostAndDatabase(ctx context.Context, user, pass, address, dbName, filePath string) error {

	conf := new(config)
	if _, err := os.Stat(filePath); err == nil {
		conf, err = getConfig(ctx, filePath)
		if err != nil {
			return err
		}
	}

	if conf.Hosts == nil {
		conf.Hosts = []*host{}
	}

	if conf.Databases == nil {
		conf.Databases = []*database{}
	}

	// add Host Info
	var ho *host
	for _, h := range conf.Hosts {
		if h.User == user && h.Password == pass && h.Address == address {
			ho = h
			break
		}
	}

	if ho == nil {
		ho = &host{
			Key:      len(conf.Hosts) + 1,
			User:     user,
			Password: pass,
			Address:  address,
		}

		conf.Hosts = append(conf.Hosts, ho)
	}

	hostKey := ho.Key

	var db *database

	// add Database Info
	for _, d := range conf.Databases {
		if d.HostKey == hostKey && d.Name == dbName {
			db = d
			break
		}
	}

	if db == nil {
		db = &database{
			HostKey: hostKey,
			Name:    dbName,
		}

		conf.Databases = append(conf.Databases, db)
	}

	return setConfig(ctx, conf, filePath)
}

func RemoveHostAndDatabase(ctx context.Context, user, pass, address, dbName, filePath string) error {

	conf, err := getConfig(ctx, filePath)
	if err != nil {
		return err
	}

	// add Host Info
	var ho *host
	for _, h := range conf.Hosts {
		if h.User == user && h.Password == pass && h.Address == address {
			ho = h
			break
		}
	}

	if ho == nil {
		return ErrWrap(
			fmt.Errorf("none data user:%s, pass:%s, address:%s", user, pass, address),
			UserInputError,
		)
	}

	hostKey := ho.Key

	var db *database

	dbs := make([]*database, 0, len(conf.Databases))

	// add Database Info
	for _, d := range conf.Databases {
		if d.HostKey == hostKey && d.Name == dbName {
			db = d
			continue
		}
		dbs = append(dbs, d)
	}

	if db == nil {
		return ErrWrap(
			fmt.Errorf("none data user:%s, pass:%s, address:%s, database:%s",
				user, pass, address, dbName,
			),
			UserInputError,
		)
	}
	conf.Databases = dbs

	return setConfig(ctx, conf, filePath)
}

func ReloadAllTableInfo(ctx context.Context, filePath string) error {
	conf, err := getConfig(ctx, filePath)
	if err != nil {
		return err
	}

	for _, db := range conf.Databases {
		for _, h := range conf.Hosts {
			if db.HostKey != h.Key {
				continue
			}

			err := openAdditional(ctx, h.User, h.Password, h.Address, db.Name)
			if err != nil {
				return err
			}

			tables, err := showtables(db.Name)
			if err != nil {
				return err
			}

			db.Tables = tables
		}
	}

	return setConfig(ctx, conf, filePath)
}

func GetDBInfo(ctx context.Context) *model.DBInfo {
	return dbInfo
}

func SetDBInfo(ctx context.Context, info *model.DBInfo) {
	dbInfo = info
}

func SetDBOne(host, database, user, pass string) {
	dbInfo = &model.DBInfo{
		Hosts: []*model.DBHost{
			&model.DBHost{
				Address:   host,
				Databases: []*model.DBDatabase{&model.DBDatabase{Name: database}},
				User:      user,
				Password:  pass,
			},
		},
	}
}

func LoadDBInfo(ctx context.Context, filePath string) error {
	conf, err := getConfig(ctx, filePath)
	if err != nil {
		return err
	}

	dbs := []*model.DBHost{}
	for _, ho := range conf.Hosts {
		dbh := &model.DBHost{Address: ho.Address, User: ho.User, Password: ho.Password, Databases: []*model.DBDatabase{}}
		for _, d := range conf.Databases {
			if d.HostKey != ho.Key {
				continue
			}

			dbd := &model.DBDatabase{Name: d.Name, Tables: d.Tables}
			dbh.Databases = append(dbh.Databases, dbd)
		}
		dbs = append(dbs, dbh)
	}

	SetDBInfo(ctx, &model.DBInfo{dbs})

	return nil
}

func GetTableDBMap(ctx context.Context) model.TableDBMap {
	info := GetDBInfo(ctx)

	tbMap := map[string][]string{}

	for _, h := range info.Hosts {
		for _, db := range h.Databases {
			for _, tb := range db.Tables {
				tDbs := []string{}
				if v, ok := tbMap[tb]; ok {
					tDbs = v
				}
				tDbs = append(tDbs, db.Name)
				tbMap[tb] = tDbs
			}
		}
	}

	return model.TableDBMap(tbMap)
}
