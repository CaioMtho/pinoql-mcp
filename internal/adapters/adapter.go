package adapters

import "github.com/jmoiron/sqlx"

type Adapter interface {
	HealthCheck() error
	RunQuery(query string, args ...any) (*sqlx.Rows, error)
	DescribeSchema() (string, error)
	GetDB() *sqlx.DB
	Close() error
}
