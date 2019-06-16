package service

import (
	"context"
	"expl/model"
	"strings"
)

// Explains execute explain queries
func Explains(
	ctx context.Context,
	queries []string,
	option *model.ExplainOption,
	fi *model.ExplainFilter,
) ([]*model.ExplainInfo, error) {
	infos := make([]*model.ExplainInfo, 0)

	option.TableMap = GetTableDBMap(ctx) // TODO: ここでやるべき？

	if err := openAdditonal(ctx, GetDBInfo(ctx)); err != nil {
		return infos, err
	}

	infos, err := exeExplains(ctx, queries, option)
	if err != nil {
		return infos, err
	}

	return filterResults(infos, fi), nil
}

func openAdditonal(ctx context.Context, dbi *model.DBInfo) error {
	for _, h := range dbi.Hosts {
		for _, db := range h.Databases {
			err := openAdditional(ctx, h.User, h.Password, h.Address, db.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func exeExplains(
	ctx context.Context, queries []string, option *model.ExplainOption,
) ([]*model.ExplainInfo, error) {

	list := []*model.ExplainInfo{}

	queryMap := map[string]*model.SQLInfo{}

	for _, q := range queries {
		// SQL Parse
		info, err := getSQLInfo(ctx, q)
		if err != nil {
			if option.NoError {
				if ErrCode(err) == int(SQLParseError) {
					continue
				}
			}
			return nil, err
		}
		if info.Table == "" {
			continue
		}

		// uniqフラグ指定の場合、重複SQLの除外
		if _, ok := queryMap[info.PrepareSQL]; ok && option.Uniq {
			continue
		}
		queryMap[info.PrepareSQL] = info

		if option.UseTableMap {
			for _, db := range option.TableMap[info.Table] {
				// Explain実行
				expInfo, err := exeExplain(ctx, db, q, info.PrepareSQL)
				if err != nil {
					if option.NoError {
						if ErrCode(err) == int(ExeExplainError) {
							continue
						}
					}
					return nil, err
				}

				list = append(list, expInfo)
			}
		} else {
			// Explain実行
			db := option.DB
			expInfo, err := exeExplain(ctx, db, q, info.PrepareSQL)
			if err != nil {
				if option.NoError {
					if ErrCode(err) == int(ExeExplainError) {
						continue
					}
				}
				return nil, err
			}

			list = append(list, expInfo)
		}

	}

	return list, nil
}

func exeExplain(ctx context.Context, db, sql, prepareSQL string) (*model.ExplainInfo, error) {
	exps, err := explain(ctx, db, sql)
	if err != nil {
		return nil, err
	}

	return &model.ExplainInfo{
		DataBase:   db,
		PrepareSQL: prepareSQL,
		SQL:        sql,
		Values:     exps,
	}, nil
}

func filterResults(infos []*model.ExplainInfo, fi *model.ExplainFilter) []*model.ExplainInfo {

	if fi == (&model.ExplainFilter{}) {
		return infos
	}

	list := make([]*model.ExplainInfo, 0, len(infos))

	for _, info := range infos {

		add := true
		for i, exp := range info.Values {

			if i == 0 || !add {
				// SelectType
				add = getAddFlagForFiltering(add, fi.SelectType, exp.SelectType, false, false)

				// Table
				add = getAddFlagForFiltering(add, fi.Table, exp.Table, false, false)

				// Type
				add = getAddFlagForFiltering(add, fi.Type, exp.Type, false, false)

				// Extra
				add = getAddFlagForFiltering(add, fi.Extra, exp.Extra, false, true)
			}

			// SelectTypeNot
			add = getAddFlagForFiltering(add, fi.SelectTypeNot, exp.SelectType, true, false)

			// TableNot
			add = getAddFlagForFiltering(add, fi.TableNot, exp.Table, true, false)

			// TypeNot
			add = getAddFlagForFiltering(add, fi.TypeNot, exp.Type, true, false)

			// ExtraNot
			add = getAddFlagForFiltering(add, fi.ExtraNot, exp.Extra, true, true)
		}

		if add {
			list = append(list, info)
		}
	}

	return list
}

func getAddFlagForFiltering(add bool, list []string, target string, not, isExp bool) bool {

	if list == nil {
		return add
	}

	for _, val := range list {
		if isTrueForFiltering(val, target, isExp) == not {
			add = false
			if not {
				break
			}
		} else {
			add = true
			if !not {
				break
			}
		}
	}
	return add
}

func isTrueForFiltering(val, target string, isExp bool) bool {
	if isExp {
		return (strings.Index(strings.ToLower(val), strings.ToLower(target)) > -1)
	} else {
		return (val == target)
	}
}
