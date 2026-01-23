package mcp

import (
	"context"

	"github.com/CaioMtho/pinoql-mcp/internal/connection"
	"github.com/CaioMtho/pinoql-mcp/internal/credentials/audit"
	"github.com/CaioMtho/pinoql-mcp/internal/credentials/connection_data"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ToolDependencies struct {
	ConnManager *connection.Manager
	ConnRepo    *connection_data.Repository
	AuditRepo   *audit.Repository
}

type ExecuteQueryInput struct {
	ConnectionID string        `json:"connection_id" jsonschema:"ID of the database connection,required"`
	Query        string        `json:"query" jsonschema:"SQL query to execute,required"`
	Params       []interface{} `json:"params" jsonschema:"Query parameters for prepared statements"`
}

type ExecuteQueryOutput struct {
	Success      bool                     `json:"success"`
	Rows         []map[string]interface{} `json:"rows,omitempty"`
	RowCount     int                      `json:"row_count"`
	RowsAffected int                      `json:"rows_affected,omitempty"`
	ExecutionMs  int64                    `json:"execution_ms"`
	Error        string                   `json:"error,omitempty"`
}

type ListTablesInput struct {
	ConnectionID string  `json:"connection_id" jsonschema:"ID of the database connection,required"`
	Schema       *string `json:"schema" jsonschema:"Database schema name (optional)"`
}

type ListTablesOutput struct {
	Success bool        `json:"success"`
	Tables  []TableInfo `json:"tables,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type TableInfo struct {
	Name   string  `json:"name"`
	Schema *string `json:"schema,omitempty"`
	Type   string  `json:"type"`
}

type DescribeTableInput struct {
	ConnectionID string  `json:"connection_id" jsonschema:"ID of the database connection,required"`
	TableName    string  `json:"table_name" jsonschema:"Name of the table to describe,required"`
	Schema       *string `json:"schema" jsonschema:"Database schema name (optional)"`
}

type DescribeTableOutput struct {
	Success bool         `json:"success"`
	Table   string       `json:"table,omitempty"`
	Columns []ColumnInfo `json:"columns,omitempty"`
	Indexes []IndexInfo  `json:"indexes,omitempty"`
	Error   string       `json:"error,omitempty"`
}

type ColumnInfo struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Nullable     bool    `json:"nullable"`
	DefaultValue *string `json:"default_value,omitempty"`
	IsPrimaryKey bool    `json:"is_primary_key"`
	IsUnique     bool    `json:"is_unique"`
}

type IndexInfo struct {
	Name     string   `json:"name"`
	Columns  []string `json:"columns"`
	IsUnique bool     `json:"is_unique"`
	Type     string   `json:"type"`
}

func RegisterTools(server *mcp.Server, deps *ToolDependencies) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "execute_query",
		Description: "Execute SQL query on a database connection",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ExecuteQueryInput) (*mcp.CallToolResult, ExecuteQueryOutput, error) {
		return ExecuteQueryTool(ctx, req, input, deps)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tables",
		Description: "List all tables in a database connection",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListTablesInput) (*mcp.CallToolResult, ListTablesOutput, error) {
		return ListTablesTool(ctx, req, input, deps)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "describe_table",
		Description: "Get schema information for a specific table",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DescribeTableInput) (*mcp.CallToolResult, DescribeTableOutput, error) {
		return DescribeTableTool(ctx, req, input, deps)
	})
}
