package model

type DBInfo struct {
	Hosts []*DBHost
}

type DBHost struct {
	Address   string
	User      string
	Password  string
	Port      int
	Protocol  string
	Databases []*DBDatabase
}

type DBDatabase struct {
	Name   string
	Tables []string
}

// key:table value:database
type TableDBMap map[string][]string
