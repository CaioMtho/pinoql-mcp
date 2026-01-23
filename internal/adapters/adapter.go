package adapters

import (
	"context"
	"database/sql"
)

type Adapter interface {
	Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	ListTables(ctx context.Context, schema *string) ([]TableInfo, error)
	DescribeTable(ctx context.Context, tableName string, schema *string) (*TableSchema, error)
	GetDB() *sql.DB
	Close() error
}

type TableInfo struct {
	Name   string
	Schema *string
	Type   string
}

type TableSchema struct {
	TableName string
	Schema    *string
	Columns   []ColumnInfo
	Indexes   []IndexInfo
}

type ColumnInfo struct {
	Name         string
	Type         string
	Nullable     bool
	DefaultValue *string
	IsPrimaryKey bool
	IsUnique     bool
}

type IndexInfo struct {
	Name     string
	Columns  []string
	IsUnique bool
	Type     string
}
