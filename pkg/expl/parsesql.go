package expl

import (
	"context"
	"expl/pkg/expl/model"
	"fmt"

	"strings"

	"github.com/xwb1989/sqlparser"
	querypb "github.com/xwb1989/sqlparser/dependency/querypb"
)

// getSQLInfo SQL関連情報を取得
func getSQLInfo(ctx context.Context, query string) (*model.SQLInfo, error) {
	// Parse
	bv := map[string]*querypb.BindVariable{}
	sqlStripped, comments := sqlparser.SplitMarginComments(query)

	stmt, err := sqlparser.Parse(sqlStripped)
	if err != nil {
		msg := fmt.Sprintf("error query line:%s", query)
		return nil, ErrWrapWithMessage(err, SQLParseError, msg)
	}

	// Normalize
	prefix := ""
	sqlparser.Normalize(stmt, bv, prefix)

	// query
	preareQuery := comments.Leading + sqlparser.String(stmt) + comments.Trailing

	// table
	table := ""
	switch stmt.(type) {
	case *sqlparser.Select, *sqlparser.Update, *sqlparser.Delete:

		table = getFirstTableName(stmt)
		if table == "" {
			table = getTableNameFromDBDot(query)
		}
	}

	return &model.SQLInfo{PrepareSQL: preareQuery, Table: table}, nil
}

// getFirstTableName SQLから最初のTable名を取得
func getFirstTableName(stmt sqlparser.Statement) string {
	var expr sqlparser.SimpleTableExpr

	switch stmt.(type) {
	case *sqlparser.Select:
		expr = stmt.(*sqlparser.Select).From[0].(*sqlparser.AliasedTableExpr).Expr

	case *sqlparser.Update:
		expr = stmt.(*sqlparser.Update).TableExprs[0].(*sqlparser.AliasedTableExpr).Expr

	case *sqlparser.Delete:
		expr = stmt.(*sqlparser.Delete).TableExprs[0].(*sqlparser.AliasedTableExpr).Expr
	}

	switch expr.(type) {
	case sqlparser.TableName:
		out := sqlparser.GetTableName(expr)
		if out.String() != "" {
			return out.String()
		}

	case *sqlparser.Subquery:
		return getFirstTableName(expr.(*sqlparser.Subquery).Select)
	}

	return ""
}

// getTableNameFromDBDot DB.Table名の形式の場合、SQLからTable名を取得
func getTableNameFromDBDot(query string) string {
	tokens := strings.Fields(query)
	index := 0
	for i, v := range tokens {
		if strings.ToLower(v) == "from" {
			index = i
			break
		}

		if strings.ToLower(v) == "update" {
			index = i
			break
		}
	}
	if len(tokens) <= index+1 {
		return ""
	}
	db_tb := tokens[index+1]

	ind := strings.Index(db_tb, ".")
	if ind == -1 {
		return ""
	}

	ind = ind + 1
	table := db_tb[ind:]

	return table
}
