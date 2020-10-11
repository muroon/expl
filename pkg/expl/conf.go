package expl

import (
	"context"
	"github.com/muroon/expl/pkg/expl/model"
	"fmt"
	"path/filepath"

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
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"`
}

type database struct {
	HostKey int      `yaml:"hostkey"`
	Name    string   `yaml:"name"`
	Tables  []string `yaml:"tables"`
}

type param struct {
	Address      string
	User         string
	Password     string
	Database     string
	Port         int
	Protocol     string
	ConfFilePath string
}

var dbInfo *model.DBInfo

// DBOption function of database setting option
type DBOption func(pm *param) *param

// DBUser user setting function
func DBUser(user string) DBOption {
	return func(pm *param) *param {
		pm.User = user
		return pm
	}
}

// DBPass password setting function
func DBPass(pass string) DBOption {
	return func(pm *param) *param {
		pm.Password = pass
		return pm
	}
}

// DBHost setting host function
func DBHost(address string) DBOption {
	return func(pm *param) *param {
		pm.Address = address
		return pm
	}
}

// DBDatabase database setting function
func DBDatabase(database string) DBOption {
	return func(pm *param) *param {
		pm.Database = database
		return pm
	}
}

// DBPort setting port function
func DBPort(port int) DBOption {
	return func(pm *param) *param {
		pm.Port = port
		return pm
	}
}

// DBProtocol setting protocol function
func DBProtocol(protocol string) DBOption {
	return func(pm *param) *param {
		pm.Protocol = protocol
		return pm
	}
}

// ConfFilePath setting config path function
func ConfFilePath(path string) DBOption {
	return func(pm *param) *param {
		pm.ConfFilePath = path
		return pm
	}
}

func getParam(pmfs ...DBOption) *param {
	pm := &param{
		Address:  "localhost",
		Port:     3306,
		Protocol: "tcp",
	}
	for _, pmf := range pmfs {
		pm = pmf(pm)
	}
	return pm
}

// AddHostAndDatabase adding host and database
func AddHostAndDatabase(ctx context.Context, pmfs ...DBOption) error {
	pm := getParam(pmfs...)

	conf := new(config)
	if _, err := os.Stat(pm.ConfFilePath); err == nil {
		conf, err = getConfig(ctx, pm.ConfFilePath)
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
		if h.User == pm.User &&
			h.Password == pm.Password &&
			h.Address == pm.Address &&
			h.Port == pm.Port &&
			h.Protocol == pm.Protocol {
			ho = h
			break
		}
	}

	if ho == nil {
		ho = &host{
			Key:      len(conf.Hosts) + 1,
			User:     pm.User,
			Password: pm.Password,
			Address:  pm.Address,
			Port:     pm.Port,
			Protocol: pm.Protocol,
		}

		conf.Hosts = append(conf.Hosts, ho)
	}

	hostKey := ho.Key

	var db *database

	// add Database Info
	for _, d := range conf.Databases {
		if d.HostKey == hostKey && d.Name == pm.Database {
			db = d
			break
		}
	}

	if db == nil {
		db = &database{
			HostKey: hostKey,
			Name:    pm.Database,
		}

		conf.Databases = append(conf.Databases, db)
	}

	return setConfig(ctx, conf, pm.ConfFilePath)
}

// RemoveHostAndDatabase removing host and database
func RemoveHostAndDatabase(ctx context.Context, pmfs ...DBOption) error {
	pm := getParam(pmfs...)

	conf, err := getConfig(ctx, pm.ConfFilePath)
	if err != nil {
		return err
	}

	var ho *host
	for _, h := range conf.Hosts {
		if h.User == pm.User &&
			h.Password == pm.Password &&
			h.Address == pm.Address &&
			h.Port == pm.Port &&
			h.Protocol == pm.Protocol {
			ho = h
			break
		}
	}

	if ho == nil {
		return ErrWrap(
			fmt.Errorf("none data parameter:%#v", pm),
			UserInputError,
		)
	}

	hostKey := ho.Key

	var db *database

	dbs := make([]*database, 0, len(conf.Databases))

	// add Database Info
	for _, d := range conf.Databases {
		if d.HostKey == hostKey && d.Name == pm.Database {
			db = d
			continue
		}
		dbs = append(dbs, d)
	}

	if db == nil {
		return ErrWrap(
			fmt.Errorf("none database data parameter:%#v", pm),
			UserInputError,
		)
	}
	conf.Databases = dbs

	return setConfig(ctx, conf, pm.ConfFilePath)
}

// ReloadAllTableInfo reload all setting from file
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

			err := openAdditional(ctx, h.User, h.Password, h.Address, db.Name, h.Port, h.Protocol)
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

// GetDBInfo getting database information
func GetDBInfo(ctx context.Context) *model.DBInfo {
	return dbInfo
}

// SetDBInfo setting database information
func SetDBInfo(ctx context.Context, info *model.DBInfo) {
	dbInfo = info
}

// SetDBOne setting single db setting
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

// LoadDBInfo loading db information
func LoadDBInfo(ctx context.Context, filePath string) error {
	conf, err := getConfig(ctx, filePath)
	if err != nil {
		return err
	}

	dbs := []*model.DBHost{}
	for _, ho := range conf.Hosts {
		dbh := &model.DBHost{
			Address:   ho.Address,
			User:      ho.User,
			Password:  ho.Password,
			Port:      ho.Port,
			Protocol:  ho.Protocol,
			Databases: []*model.DBDatabase{},
		}
		for _, d := range conf.Databases {
			if d.HostKey != ho.Key {
				continue
			}

			dbd := &model.DBDatabase{Name: d.Name, Tables: d.Tables}
			dbh.Databases = append(dbh.Databases, dbd)
		}
		dbs = append(dbs, dbh)
	}

	SetDBInfo(ctx, &model.DBInfo{Hosts:dbs})

	return nil
}

// GetTableDBMap getting table-database mapping
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

	buf, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, ErrWrap(err, UserInputError)
	}

	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		return nil, ErrWrap(err, UserInputError)
	}

	return &c, nil
}
