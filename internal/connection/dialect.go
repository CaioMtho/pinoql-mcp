package connection

type Dialect string

const (
	PostgreSQL Dialect = "postgresql"
	SQLite     Dialect = "sqlite"
	DuckDB     Dialect = "duckdb"
)

func GetDialects() []string {
	return []string{"postgresql", "sqlite", "duckdb"}
}

func IsValidDialect(s string) bool {
	switch Dialect(s){
	case PostgreSQL, SQLite, DuckDB:
		return true
	default:
		return false
	}
}