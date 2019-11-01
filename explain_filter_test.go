package expl

import (
	"expl/model"
	"fmt"
	"testing"
)

func TestUseCaseExplain_IsTrueForFiltering(t *testing.T) {

	res := false
	res = isTrueForFiltering("SIMPLE", "SIMPLE", false)
	if !res {
		t.Errorf("failed TestUseCaseExplain_IsTrueForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%s, %s, %v)", "SIMPLE", "SIMPLE", false),
		)
	}

	res = isTrueForFiltering("SIMPLE", "!SIMPLE", false)
	if res {
		t.Errorf("failed TestUseCaseExplain_IsTrueForFiltering %v",
			fmt.Errorf("isTrueForFiltering Not (%s, %s, %v)", "SIMPLE", "!SIMPLE", false),
		)
	}

	res = isTrueForFiltering("Using where; Using index", "Using where", true)
	if !res {
		t.Errorf("failed TestUseCaseExplain_IsTrueForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%s, %s, %v)", "Using where; Using index", "Using where", true),
		)
	}

	res = isTrueForFiltering("Using where; Using index", "Using filesourt", true)
	if res {
		t.Errorf("failed TestUseCaseExplain_IsTrueForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%s, %s, %v)", "Using where; Using index", "Using filesourt", true),
		)
	}
}

func TestUseCaseExplain_GetAddFlagForFiltering(t *testing.T) {

	res := true
	res = getAddFlagForFiltering([]string{"SIMPLE", "PRIMARY"}, "SIMPLE", false, false)
	if !res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{"SIMPLE", "ALL"}, "SIMPLE", false, false),
		)
	}

	res = true
	res = getAddFlagForFiltering([]string{}, "SIMPLE", false, false)
	if !res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{}, "SIMPLE", false, false),
		)
	}

	res = true
	res = getAddFlagForFiltering([]string{}, "SIMPLE", true, false)
	if !res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{}, "SIMPLE", true, false),
		)
	}

	res = true
	res = getAddFlagForFiltering([]string{"PRIMARY"}, "SIMPLE", false, false)
	if res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{"ALL"}, "SIMPLE", false, false),
		)
	}

	res = true
	res = getAddFlagForFiltering([]string{"SIMPLE"}, "SIMPLE", true, false)
	if res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{}, "SIMPLE", true, false),
		)
	}

	res = true
	res = getAddFlagForFiltering([]string{"PRIMARY"}, "SIMPLE", true, false)
	if !res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{}, "SIMPLE", true, false),
		)
	}
}

func TestUseCaseExplain_FilterResults(t *testing.T) {

	expModel0 := getExplainModel(
		setExplainID(1),
		setExplainSelectedType("SIMPLE"),
		setExplainTable("tag_memo"),
		setExplainType("ref"),
		setExplainPossibleKeys("PRIMARY,memo_id"),
		setExplainKey("memo_id"),
		setExplainKeyLen(4),
		setExplainRef("const"),
		setExplainRows(2),
		setExplainFiltered(100.00),
		setExplainExtra("Using index"),
	)

	expModel1 := getExplainModel(
		setExplainID(1),
		setExplainSelectedType("SIMPLE"),
		setExplainTable("tag"),
		setExplainType("eq_ref"),
		setExplainPossibleKeys("PRIMARY"),
		setExplainKey("PRIMARY"),
		setExplainKeyLen(4),
		setExplainRef("memo_sample.tag_memo.tag_id"),
		setExplainRows(1),
		setExplainFiltered(100.00),
	)

	expModel2 := getExplainModel(
		setExplainID(2),
		setExplainSelectedType("SIMPLE"),
		setExplainTable("tag"),
		setExplainType("ALL"),
		setExplainRows(23),
		setExplainFiltered(11.00),
		setExplainExtra("Using where"),
	)

	infos := []*model.ExplainInfo{
		&model.ExplainInfo{
			PrepareSQL: "select tag.* from tag, tag_memo where tag_memo.memo_id = :1 and tag.id = tag_memo.tag_id",
			SQL:        "select tag.* from tag, tag_memo where tag_memo.memo_id = 12 and tag.id = tag_memo.tag_id",
			Values: []*model.Explain{
				expModel0,
				expModel1,
			},
		},
		&model.ExplainInfo{
			PrepareSQL: "select tag.* from tag where title like '%:1%'",
			SQL:        "select tag.* from tag where title like '%ok%'",
			Values: []*model.Explain{
				expModel2,
			},
		},
	}

	list := make([]*model.ExplainInfo, 0)

	for _, info := range infos {
		if !getAdditionalFlgInFilterResult(info, &model.ExplainFilter{}) {
			continue
		}
		list = append(list, info)
	}
	if len(list) != len(infos) {
		t.Error("failed TestUseCaseExplain_FilterResults")
	}

	filter := &model.ExplainFilter{
		SelectType: []string{"SIMPLE"},
		Extra:      []string{"Using index"},
	}

	list = make([]*model.ExplainInfo, 0)
	for _, info := range infos {
		if !getAdditionalFlgInFilterResult(info, filter) {
			continue
		}
		list = append(list, info)
	}
	if len(list) != 1 {
		t.Errorf("failed TestUseCaseExplain_FilterResults: %d\n", len(list))
	}


	filter = &model.ExplainFilter{
		SelectType: []string{"SIMPLE"},
		TypeNot:    []string{"const", "eq_ref", "ref"},
	}

	list = make([]*model.ExplainInfo, 0)

	for _, info := range infos {
		if !getAdditionalFlgInFilterResult(info, filter) {
			continue
		}
		list = append(list, info)
	}
	if len(list) != 1 {
		t.Errorf("failed TestUseCaseExplain_FilterResults len: %d", len(list))
	}

	filter = &model.ExplainFilter{
		SelectType: []string{"SIMPLE"},
	}

	list = make([]*model.ExplainInfo, 0)

	for _, info := range infos {
		if !getAdditionalFlgInFilterResult(info, filter) {
			continue
		}
		list = append(list, info)
	}

	if len(list) != 2 {
		t.Errorf("failed TestUseCaseExplain_FilterResults len: %d", len(list))
	}


	filter = &model.ExplainFilter{
		SelectType: []string{"SIMPLE"},
		Extra: []string{"Using where"},
	}

	list = make([]*model.ExplainInfo, 0)

	for _, info := range infos {
		if !getAdditionalFlgInFilterResult(info, filter) {
			continue
		}
		list = append(list, info)
	}

	if len(list) != 1 {
		t.Errorf("failed TestUseCaseExplain_FilterResults len: %d", len(list))
	}
}

type option func(exp *model.Explain) *model.Explain

func setExplainID(id int64) option {
	return func(exp *model.Explain) *model.Explain {
		exp.ID = id
		return exp
	}
}

func setExplainSelectedType(seletedType string) option {
	return func(exp *model.Explain) *model.Explain {
		exp.SelectType = seletedType
		return exp
	}
}

func setExplainTable(table string) option {
	return func(exp *model.Explain) *model.Explain {
		exp.Table = table
		return exp
	}
}

func setExplainPartitions(partitions int64) option {
	return func(exp *model.Explain) *model.Explain {
		exp.Partitions = partitions
		return exp
	}
}

func setExplainType(tp string) option {
	return func(exp *model.Explain) *model.Explain {
		exp.Type = tp
		return exp
	}
}

func setExplainPossibleKeys(possibleKeys string) option {
	return func(exp *model.Explain) *model.Explain {
		exp.PossibleKeys = possibleKeys
		return exp
	}
}

func setExplainKey(key string) option {
	return func(exp *model.Explain) *model.Explain {
		exp.Key = key
		return exp
	}
}

func setExplainKeyLen(keyLen int64) option {
	return func(exp *model.Explain) *model.Explain {
		exp.KeyLen = keyLen
		return exp
	}
}

func setExplainRef(ref string) option {
	return func(exp *model.Explain) *model.Explain {
		exp.Ref = ref
		return exp
	}
}

func setExplainRows(rows int64) option {
	return func(exp *model.Explain) *model.Explain {
		exp.Rows = rows
		return exp
	}
}

func setExplainFiltered(filtered float64) option {
	return func(exp *model.Explain) *model.Explain {
		exp.Filtered = filtered
		return exp
	}
}


func setExplainExtra(extra string) option {
	return func(exp *model.Explain) *model.Explain {
		exp.Extra = extra
		return exp
	}
}

func getExplainModel(opts ...option) *model.Explain {
	exp := new(model.Explain)
	for _, opt := range opts {
		exp = opt(exp)
	}

	return exp
}

