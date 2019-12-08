package model

// DBInfo struct of database info
type DBInfo struct {
	Hosts []*DBHost
}

// DBHost struct of database host
type DBHost struct {
	Address   string
	User      string
	Password  string
	Port      int
	Protocol  string
	Databases []*DBDatabase
}

// DBDatabase struct of database
type DBDatabase struct {
	Name   string
	Tables []string
}

// TableDBMap table-database mapping (key:table value:database)
type TableDBMap map[string][]string
