package model

// Explain struct of expl query result
type Explain struct {
	ID           int64
	SelectType   string
	Table        string
	Partitions   int64
	Type         string
	PossibleKeys string
	Key          string
	KeyLen       int64
	Ref          string
	Rows         int64
	Filtered     float64
	Extra        string
}

// ExplainInfo struct of expl query result
type ExplainInfo struct {
	DataBase   string
	PrepareSQL string
	SQL        string
	Values     []*Explain
}

// ExplainFilter filter option
type ExplainFilter struct {
	SelectType    []string
	SelectTypeNot []string
	Table         []string
	TableNot      []string
	Type          []string
	TypeNot       []string
	Extra         []string
	ExtraNot      []string
}

// ExplainOption option of service.Explain
type ExplainOption struct {
	// useTableMap
	UseTableMap bool

	// table-database map
	TableMap TableDBMap

	// Config path
	Config string

	// database (no use table mapping)
	DB string

	// database host (no use table mapping)
	DBHost string

	// database user (no use table mapping)
	DBUser string

	// database password (no use table mapping)
	DBPass string

	// uniq
	Uniq bool
	// no error
	NoError bool
	// table
	Table map[string]uint
	// without table
	TableNot map[string]uint
}
