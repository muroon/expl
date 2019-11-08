package view

import (
	"github.com/muroon/expl/pkg/expl/model"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

var tableHeader []string = []string{
	"ID",
	"SelectType",
	"Table",
	"Partitions",
	"Type",
	"PossibleKeys",
	"Key",
	"KeyLen",
	"Ref",
	"Rows",
	"Filtered",
	"Extra",
}

var optionHeader []string = []string{
	"inpnut option",
	"value",
}

// RenderExplain render expl result
func RenderExplain(info *model.ExplainInfo, isParseEnable bool) {

	if info == nil || len(info.Values) == 0 {
		return
	}

	headerDatas := make([][]string, 0, 3)
	headerDatas = append(headerDatas, []string{"DataBase:", info.DataBase})
	if isParseEnable {
		headerDatas = append(headerDatas, []string{"ParsedSQL:", info.PrepareSQL})
	}
	headerDatas = append(headerDatas, []string{"SQL:", info.SQL})

	headerTable := tablewriter.NewWriter(os.Stdout)
	headerTable.SetBorder(false)
	headerTable.SetColumnSeparator("")
	headerTable.SetAutoWrapText(false)
	headerTable.SetAlignment(tablewriter.ALIGN_LEFT)
	headerTable.AppendBulk(headerDatas)
	headerTable.Render()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(tableHeader)

	for _, v := range info.Values {
		table.Append(getRecord(v))
	}

	table.Render()

	fmt.Print("\n")
}

func getRecord(e *model.Explain) []string {
	rec := []string{
		strconv.FormatInt(e.ID, 10),
		e.SelectType,
		e.Table,
		strconv.FormatInt(e.Partitions, 10),
		e.Type,
		e.PossibleKeys,
		e.Key,
		strconv.FormatInt(e.KeyLen, 10),
		e.Ref,
		strconv.FormatInt(e.Rows, 10),
		strconv.FormatFloat(e.Filtered, 'f', 4, 64),
		e.Extra,
	}

	return rec
}

// RenderOptions render expl options
func RenderOptions(
	expOpt *model.ExplainOption,
	fiOpt *model.ExplainFilter,
	logPath, format, formatCmd string,
) {

	optionDatas := [][]string{
		[]string{"database", expOpt.DB},
		[]string{"host", expOpt.DBHost},
		[]string{"user", expOpt.DBUser},
		[]string{"pass", expOpt.DBPass},
		[]string{"conf", expOpt.Config},
		[]string{"log", logPath},
		[]string{"format", format},
		[]string{"format-cmd", formatCmd},
		[]string{"filter-select-type", strings.Join(fiOpt.SelectType, ",")},
		[]string{"filter-no-select-type", strings.Join(fiOpt.SelectTypeNot, ",")},
		[]string{"filter-table", strings.Join(fiOpt.Table, ",")},
		[]string{"filter-no-table", strings.Join(fiOpt.TableNot, ",")},
		[]string{"filter-type", strings.Join(fiOpt.Type, ",")},
		[]string{"filter-no-type", strings.Join(fiOpt.TypeNot, ",")},
		[]string{"filter-extra", strings.Join(fiOpt.Extra, ",")},
		[]string{"filter-no-extra", strings.Join(fiOpt.ExtraNot, ",")},
		[]string{"update-table-map", fmt.Sprintf("%v", expOpt.UseTableMap)},
		[]string{"ignore-error", fmt.Sprintf("%v", expOpt.NoError)},
		[]string{"combine-sql", fmt.Sprintf("%v", expOpt.Uniq)},
	}

	optionTable := tablewriter.NewWriter(os.Stdout)
	optionTable.SetBorder(false)
	optionTable.SetColumnSeparator("")
	optionTable.SetAutoWrapText(false)
	optionTable.SetAlignment(tablewriter.ALIGN_LEFT)
	optionTable.SetHeader(optionHeader)
	optionTable.AppendBulk(optionDatas)
	optionTable.Render()

	fmt.Println("")
}
