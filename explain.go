package expl

import (
	"context"
	"expl/model"
	"strings"
)

// ExplainChannels execute explain queries
func ExplainChannels(
	ctx context.Context,
	queryCh <-chan string,
	option *model.ExplainOption,
	fi *model.ExplainFilter,
) (<-chan *model.ExplainInfo, <-chan error) {

	exCh := make(chan *model.ExplainInfo)
	errCh := make(chan error)

	option.TableMap = GetTableDBMap(ctx)

	go func() {
		defer func() {
			close(exCh)
			close(errCh)
		}()

		if err := openAdditonal(ctx, GetDBInfo(ctx)); err != nil {
			errCh <- err
			return
		}

		ech, erch := exeExplainChannels(ctx, queryCh, option)
		var err error
		for {
			select {
			case exp, ok := <-ech:
				if !ok {
					return
				} else if getAdditionalFlgInFilterResult(exp, fi) {
					exCh <- exp
				}
			case err = <-erch:
				errCh <- err
				return
			}
		}

	}()

	return exCh, errCh
}

// Explain execute explain query
func Explain(
	ctx context.Context,
	query string,
	option *model.ExplainOption,
	fi *model.ExplainFilter,
) (*model.ExplainInfo, error) {
	expIno := new(model.ExplainInfo)

	option.TableMap = GetTableDBMap(ctx)

	if err := openAdditonal(ctx, GetDBInfo(ctx)); err != nil {
		return expIno, err
	}

	exp, err := exeExplainOne(ctx, query, option)
	if err != nil {
		return expIno, err
	}

	if !getAdditionalFlgInFilterResult(exp, fi) {
		return expIno, nil
	}

	return exp, nil
}

func openAdditonal(ctx context.Context, dbi *model.DBInfo) error {
	for _, h := range dbi.Hosts {
		for _, db := range h.Databases {
			err := openAdditional(ctx, h.User, h.Password, h.Address, db.Name, h.Port, h.Protocol)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func exeExplainOne(
	ctx context.Context, query string, option *model.ExplainOption,
) (*model.ExplainInfo, error) {

	expInfo := new(model.ExplainInfo)

	// SQL Parse
	info, err := getSQLInfo(ctx, query)
	if err != nil {
		if !option.NoError || ErrCode(err) != int(SQLParseError) {
			return nil, err
		}
	}
	if info.Table == "" {
		return expInfo, nil
	}

	if option.UseTableMap {
		for _, db := range option.TableMap[info.Table] {
			// Explain実行
			expInfo, err = exeExplain(ctx, db, query, info.PrepareSQL)
			if err != nil {
				if !option.NoError || ErrCode(err) != int(ExeExplainError) {
					return nil, err
				}
			}
		}
	} else {
		// Explain実行
		db := option.DB
		expInfo, err = exeExplain(ctx, db, query, info.PrepareSQL)
		if err != nil {
			if !option.NoError || ErrCode(err) != int(ExeExplainError) {
				return nil, err
			}
		}
	}

	return expInfo, nil
}

func exeExplainChannels(
	ctx context.Context, queryCh <-chan string, option *model.ExplainOption,
) (<-chan *model.ExplainInfo, <-chan error) {

	exCh := make(chan *model.ExplainInfo)
	errCh := make(chan error)

	queryMap := map[string]*model.SQLInfo{}

	go func() {
		defer func() {
			errCh <- nil
			close(exCh)
			close(errCh)
		}()

		for q := range queryCh {
			// SQL Parse
			info, err := getSQLInfo(ctx, q)
			if err != nil {
				if option.NoError {
					if ErrCode(err) == int(SQLParseError) {
						continue
					}
				}
				errCh <- err
				return
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
						errCh <- err
						return
					}

					exCh <- expInfo
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
					errCh <- err
					return
				}

				exCh <- expInfo
			}
		}
	}()

	return exCh, errCh
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

func getAdditionalFlgInFilterResult(info *model.ExplainInfo, fi *model.ExplainFilter) bool {
	if fi == (&model.ExplainFilter{}) {
		return true
	}

	add := true
	for _, exp := range info.Values {

		// SelectType
		if add = getAddFlagForFiltering(fi.SelectType, exp.SelectType, false, false); !add {
			continue
		}

		// Table
		if add = getAddFlagForFiltering(fi.Table, exp.Table, false, false); !add {
			continue
		}

		// Type
		if add = getAddFlagForFiltering(fi.Type, exp.Type, false, false); !add {
			continue
		}

		// PossibleKey
		if add = getAddFlagForFiltering(fi.PossibleKeys, exp.PossibleKeys, false, true); !add {
			continue
		}

		// Key
		if add = getAddFlagForFiltering(fi.Key, exp.Key, false, false); !add {
			continue
		}

		// Extra
		if add = getAddFlagForFiltering(fi.Extra, exp.Extra, false, true); !add {
			continue
		}

		// SelectTypeNot
		if add = getAddFlagForFiltering(fi.SelectTypeNot, exp.SelectType, true, false); !add {
			continue
		}

		// TableNot
		if add = getAddFlagForFiltering(fi.TableNot, exp.Table, true, false); !add {
			continue
		}

		// TypeNot
		if add = getAddFlagForFiltering(fi.TypeNot, exp.Type, true, false); !add {
			continue
		}

		// PossibleKeysNot
		if add = getAddFlagForFiltering(fi.PossibleKeysNot, exp.PossibleKeys, true, true); !add {
			continue
		}

		// KeyNot
		if add = getAddFlagForFiltering(fi.KeyNot, exp.Key, true, false); !add {
			continue
		}

		// ExtraNot
		if add = getAddFlagForFiltering(fi.ExtraNot, exp.Extra, true, true); !add {
			continue
		}

		if add {
			break
		}
	}
	return add
}

func canGetAddFlagForFiltering(list []string) bool {
	return list != nil && len(list) > 0
}

func getAddFlagForFiltering(list []string, target string, not, isExp bool) bool {

	if list == nil || len(list) == 0 {
		return true
	}

	var add bool

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
	if isExp && len(target) > 0 {
		return (strings.Index(strings.ToLower(val), strings.ToLower(target)) > -1)
	} else {
		return (val == target)
	}
}
