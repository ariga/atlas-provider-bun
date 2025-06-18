package dialect

// Dialect is the type for supported dialects.
type Dialect string

const (
	MSSQL    Dialect = "mssql"
	MySQL    Dialect = "mysql"
	Oracle   Dialect = "oracle"
	Postgres Dialect = "postgres"
	SQLite   Dialect = "sqlite"
)
