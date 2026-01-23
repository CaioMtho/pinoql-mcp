package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/CaioMtho/pinoql-mcp/internal/adapters"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Adapter struct {
	DB *sqlx.DB
}

func NewPostgresAdapter(dsn string) (*Adapter, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Adapter{DB: db}, nil
}

func (p *Adapter) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := p.DB.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}

	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (p *Adapter) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.DB.ExecContext(ctx, query, args...)
}

func (p *Adapter) ListTables(ctx context.Context, schema *string) ([]adapters.TableInfo, error) {
	targetSchema := "public"
	if schema != nil {
		targetSchema = *schema
	}

	query := `
		SELECT 
			table_name,
			table_schema,
			table_type
		FROM information_schema.tables
		WHERE table_schema = $1
		ORDER BY table_name
	`

	rows, err := p.DB.QueryxContext(ctx, query, targetSchema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []adapters.TableInfo

	for rows.Next() {
		var tableName, tableSchema, tableType string
		if err := rows.Scan(&tableName, &tableSchema, &tableType); err != nil {
			return nil, err
		}

		tables = append(tables, adapters.TableInfo{
			Name:   tableName,
			Schema: &tableSchema,
			Type:   tableType,
		})
	}

	return tables, nil
}

func (p *Adapter) DescribeTable(ctx context.Context, tableName string, schema *string) (*adapters.TableSchema, error) {
	targetSchema := "public"
	if schema != nil {
		targetSchema = *schema
	}

	columns, err := p.getColumns(ctx, tableName, targetSchema)
	if err != nil {
		return nil, err
	}

	indexes, err := p.getIndexes(ctx, tableName, targetSchema)
	if err != nil {
		return nil, err
	}

	return &adapters.TableSchema{
		TableName: tableName,
		Schema:    &targetSchema,
		Columns:   columns,
		Indexes:   indexes,
	}, nil
}

func (p *Adapter) getColumns(ctx context.Context, tableName, schema string) ([]adapters.ColumnInfo, error) {
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key,
			CASE WHEN u.column_name IS NOT NULL THEN true ELSE false END as is_unique
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku
				ON tc.constraint_name = ku.constraint_name
				AND tc.table_schema = ku.table_schema
			WHERE tc.constraint_type = 'PRIMARY KEY'
				AND tc.table_name = $1
				AND tc.table_schema = $2
		) pk ON c.column_name = pk.column_name
		LEFT JOIN (
			SELECT ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku
				ON tc.constraint_name = ku.constraint_name
				AND tc.table_schema = ku.table_schema
			WHERE tc.constraint_type = 'UNIQUE'
				AND tc.table_name = $1
				AND tc.table_schema = $2
		) u ON c.column_name = u.column_name
		WHERE c.table_name = $1
			AND c.table_schema = $2
		ORDER BY c.ordinal_position
	`

	rows, err := p.DB.QueryxContext(ctx, query, tableName, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []adapters.ColumnInfo

	for rows.Next() {
		var (
			name         string
			dataType     string
			isNullable   string
			defaultValue sql.NullString
			isPrimaryKey bool
			isUnique     bool
		)

		if err := rows.Scan(&name, &dataType, &isNullable, &defaultValue, &isPrimaryKey, &isUnique); err != nil {
			return nil, err
		}

		var defaultPtr *string
		if defaultValue.Valid {
			defaultPtr = &defaultValue.String
		}

		columns = append(columns, adapters.ColumnInfo{
			Name:         name,
			Type:         dataType,
			Nullable:     isNullable == "YES",
			DefaultValue: defaultPtr,
			IsPrimaryKey: isPrimaryKey,
			IsUnique:     isUnique,
		})
	}

	return columns, nil
}

func (p *Adapter) getIndexes(ctx context.Context, tableName, schema string) ([]adapters.IndexInfo, error) {
	query := `
		SELECT
			i.relname as index_name,
			array_agg(a.attname ORDER BY a.attnum) as column_names,
			ix.indisunique as is_unique,
			am.amname as index_type
		FROM pg_class t
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_am am ON i.relam = am.oid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		JOIN pg_namespace n ON t.relnamespace = n.oid
		WHERE t.relname = $1
			AND n.nspname = $2
			AND t.relkind = 'r'
		GROUP BY i.relname, ix.indisunique, am.amname
		ORDER BY i.relname
	`

	rows, err := p.DB.QueryxContext(ctx, query, tableName, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []adapters.IndexInfo

	for rows.Next() {
		var (
			indexName   string
			columnNames []string
			isUnique    bool
			indexType   string
		)

		if err := rows.Scan(&indexName, &columnNames, &isUnique, &indexType); err != nil {
			return nil, err
		}

		indexes = append(indexes, adapters.IndexInfo{
			Name:     indexName,
			Columns:  columnNames,
			IsUnique: isUnique,
			Type:     indexType,
		})
	}

	return indexes, nil
}

func (p *Adapter) GetDB() *sql.DB {
	return p.DB.DB
}

func (p *Adapter) Close() error {
	return p.DB.Close()
}
