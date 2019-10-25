package expl

import (
	"context"
	"database/sql"
	"expl/model"

	"fmt"

	"github.com/pkg/errors"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	dbMap = map[string]*sql.DB{}
	dbType = "mysql"
	officialDB = "mysql"
}

var dbMap map[string]*sql.DB

type dbm struct{}

var dbType string

var officialDB string

type explainResult struct {
	ID           int64
	SelectType   sql.NullString
	Table        sql.NullString
	Partitions   sql.NullInt64
	Type         sql.NullString
	PossibleKeys sql.NullString
	Key          sql.NullString
	KeyLen       sql.NullInt64
	Ref          sql.NullString
	Rows         sql.NullInt64
	Filtered     sql.NullFloat64
	Extra        sql.NullString
}

// openAdditional
func openAdditional(ctx context.Context, user, pass, address, database string) error {

	db, err := open(user, pass, address, database)
	if err != nil {
		return ErrWrap(err, OtherError)
	}

	dbMap[database] = db

	return nil
}

// open
func open(user, pass, address, database string) (*sql.DB, error) {
	if address == "localhost" {
		address = ""
	}

	dataSourse := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", user, pass, address, database)
	return sql.Open(dbType, dataSourse)
}

// explain
func explain(ctx context.Context, database, sql string) ([]*model.Explain, error) {
	list := make([]*model.Explain, 0)

	exQuery := fmt.Sprintf("explain %s", sql)

	rows, err := query(database, exQuery)
	if err != nil {
		return list, ErrWrap(err, ExeExplainError)
	}

	for rows.Next() {
		ex := &explainResult{}
		err := rows.Scan(&ex.ID, &ex.SelectType, &ex.Table, &ex.Partitions, &ex.Type, &ex.PossibleKeys, &ex.Key, &ex.KeyLen, &ex.Ref, &ex.Rows, &ex.Filtered, &ex.Extra)
		if err != nil {
			return list, ErrWrap(err, ExeExplainError)
		}
		exm := &model.Explain{
			ID:           ex.ID,
			SelectType:   ex.SelectType.String,
			Table:        ex.Table.String,
			Partitions:   ex.Partitions.Int64,
			Type:         ex.Type.String,
			PossibleKeys: ex.PossibleKeys.String,
			Key:          ex.Key.String,
			KeyLen:       ex.KeyLen.Int64,
			Ref:          ex.Ref.String,
			Rows:         ex.Rows.Int64,
			Filtered:     ex.Filtered.Float64,
			Extra:        ex.Extra.String,
		}

		list = append(list, exm)
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return list, ErrWrap(err, ExeExplainError)
	}

	return list, nil
}

// query
func query(database, querySQL string) (*sql.Rows, error) {

	var db *sql.DB
	var ok bool
	if db, ok = dbMap[database]; !ok {
		return nil, errors.Errorf("database is none. database:%s", database)
	}

	return db.Query(querySQL)
}

func showtables(database string) ([]string, error) {
	tables := []string{}

	rows, err := query(database, "show tables")
	if err != nil {
		return nil, ErrWrap(err, ShowTablesError)
	}

	for rows.Next() {
		tb := ""
		err := rows.Scan(&tb)
		if err != nil {
			return nil, ErrWrap(err, ShowTablesError)
		}
		tables = append(tables, tb)
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return tables, ErrWrap(err, ShowTablesError)
	}

	return tables, nil
}

// close
func closeAll(ctx context.Context) {
	for _, db := range dbMap {
		db.Close()
	}
}
