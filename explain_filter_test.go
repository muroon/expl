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
	res = getAddFlagForFiltering(res, []string{"SIMPLE", "PRIMARY"}, "SIMPLE", false, false)
	if !res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{"SIMPLE", "ALL"}, "SIMPLE", false, false),
		)
	}

	res = true
	res = getAddFlagForFiltering(res, []string{}, "SIMPLE", false, false)
	if !res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{}, "SIMPLE", false, false),
		)
	}

	res = true
	res = getAddFlagForFiltering(res, []string{}, "SIMPLE", true, false)
	if !res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{}, "SIMPLE", true, false),
		)
	}

	res = true
	res = getAddFlagForFiltering(res, []string{"PRIMARY"}, "SIMPLE", false, false)
	if res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{"ALL"}, "SIMPLE", false, false),
		)
	}

	res = true
	res = getAddFlagForFiltering(res, []string{"SIMPLE"}, "SIMPLE", true, false)
	if res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{}, "SIMPLE", true, false),
		)
	}

	res = true
	res = getAddFlagForFiltering(res, []string{"PRIMARY"}, "SIMPLE", true, false)
	if !res {
		t.Errorf("failed TestUseCaseExplain_GetAddFlagForFiltering %v",
			fmt.Errorf("isTrueForFiltering (%v, %s, %v, %v)", []string{}, "SIMPLE", true, false),
		)
	}
}

func TestUseCaseExplain_FilterResults(t *testing.T) {

	expModel0 := getExplainModel(1, "SIMPLE", "tag_memo", 0, "ref", "PRIMARY,memo_id", "memo_id", 4, "const", 2, 100.00, "Using index")
	expModel1 := getExplainModel(1, "SIMPLE", "tag", 0, "eq_ref", "PRIMARY", "PRIMARY", 4, "memo_sample.tag_memo.tag_id", 1, 100.00, "")
	expModel2 := getExplainModel(1, "SIMPLE", "tag", 0, "ALL", "", "", 0, "", 23, 11.00, "Using where")

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
		TypeNot:    []string{"const", "eq_ref", "req"},
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

func getExplainModel(
	id int64,
	seletedType, table string,
	partitions int64,
	tp, possibleKeys, key string,
	keyLen int64,
	ref string,
	rows int64,
	filtered float64,
	extra string,
) *model.Explain {
	return &model.Explain{
		ID:           id,
		SelectType:   seletedType,
		Table:        table,
		Partitions:   partitions,
		Type:         tp,
		PossibleKeys: possibleKeys,
		Key:          key,
		KeyLen:       keyLen,
		Ref:          ref,
		Rows:         rows,
		Filtered:     filtered,
		Extra:        extra,
	}
}
