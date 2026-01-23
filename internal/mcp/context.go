package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/CaioMtho/pinoql-mcp/internal/connection"
	credClaims "github.com/CaioMtho/pinoql-mcp/internal/credentials/claims"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type contextKey string

const (
	claimsContextKey contextKey = "pinoql_claims"
	tenantContextKey contextKey = "pinoql_tenant_id"
)

func WithClaims(ctx context.Context, claims *credClaims.PinoQLClaims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

func GetClaims(ctx context.Context) (*credClaims.PinoQLClaims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*credClaims.PinoQLClaims)
	return claims, ok
}

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantContextKey, tenantID)
}

func GetTenantID(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(tenantContextKey).(string)
	return tenantID, ok
}

func ExecuteQueryTool(ctx context.Context, req *mcp.CallToolRequest, input ExecuteQueryInput, deps *ToolDependencies) (*mcp.CallToolResult, ExecuteQueryOutput, error) {
	claims, ok := GetClaims(ctx)
	if !ok {
		return nil, ExecuteQueryOutput{
			Success: false,
			Error:   "unauthorized: missing claims",
		}, nil
	}

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		return nil, ExecuteQueryOutput{
			Success: false,
			Error:   "unauthorized: missing tenant_id",
		}, nil
	}

	if !claims.HasAccessToConnection(input.ConnectionID) {
		if logErr := deps.AuditRepo.InsertLog(NewAuditLog(tenantID, input.ConnectionID, "query", input.Query, false, "access denied", 0, 0)); logErr != nil {
			fmt.Printf("audit log error: %v\n", logErr)
		}
		return nil, ExecuteQueryOutput{
			Success: false,
			Error:   fmt.Sprintf("access denied to connection: %s", input.ConnectionID),
		}, nil
	}

	startTime := time.Now()

	connData, err := deps.ConnRepo.GetConnectionWithDSN(tenantID, input.ConnectionID)
	if err != nil {
		if logErr := deps.AuditRepo.InsertLog(NewAuditLog(tenantID, input.ConnectionID, "query", input.Query, false, err.Error(), 0, 0)); logErr != nil {
			fmt.Printf("audit log error: %v\n", logErr)
		}
		return nil, ExecuteQueryOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to get connection: %v", err),
		}, nil
	}

	cfg, err := connection.FromRaw(connData.Dialect, connData.DSN, fmt.Sprintf("%t", connData.ReadOnly))
	if err != nil {
		return nil, ExecuteQueryOutput{
			Success: false,
			Error:   fmt.Sprintf("invalid connection config: %v", err),
		}, nil
	}

	adapter, err := deps.ConnManager.GetAdapter(*cfg)
	if err != nil {
		if logErr := deps.AuditRepo.InsertLog(NewAuditLog(tenantID, input.ConnectionID, "query", input.Query, false, err.Error(), 0, 0)); logErr != nil {
			fmt.Printf("audit log error: %v\n", logErr)
		}
		return nil, ExecuteQueryOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to get adapter: %v", err),
		}, nil
	}

	rows, err := adapter.Query(ctx, input.Query, input.Params...)
	execTime := int(time.Since(startTime).Milliseconds())

	if err != nil {
		if logErr := deps.AuditRepo.InsertLog(NewAuditLog(tenantID, input.ConnectionID, "query", input.Query, false, err.Error(), execTime, 0)); logErr != nil {
			fmt.Printf("audit log error: %v\n", logErr)
		}
		return nil, ExecuteQueryOutput{
			Success:     false,
			Error:       fmt.Sprintf("query failed: %v", err),
			ExecutionMs: int64(execTime),
		}, nil
	}

	if logErr := deps.AuditRepo.InsertLog(NewAuditLog(tenantID, input.ConnectionID, "query", input.Query, true, "", execTime, len(rows))); logErr != nil {
		fmt.Printf("audit log error: %v\n", logErr)
	}

	return nil, ExecuteQueryOutput{
		Success:     true,
		Rows:        rows,
		RowCount:    len(rows),
		ExecutionMs: int64(execTime),
	}, nil
}

func ListTablesTool(ctx context.Context, req *mcp.CallToolRequest, input ListTablesInput, deps *ToolDependencies) (*mcp.CallToolResult, ListTablesOutput, error) {
	claims, ok := GetClaims(ctx)
	if !ok {
		return nil, ListTablesOutput{
			Success: false,
			Error:   "unauthorized: missing claims",
		}, nil
	}

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		return nil, ListTablesOutput{
			Success: false,
			Error:   "unauthorized: missing tenant_id",
		}, nil
	}

	if !claims.HasAccessToConnection(input.ConnectionID) {
		return nil, ListTablesOutput{
			Success: false,
			Error:   fmt.Sprintf("access denied to connection: %s", input.ConnectionID),
		}, nil
	}

	connData, err := deps.ConnRepo.GetConnectionWithDSN(tenantID, input.ConnectionID)
	if err != nil {
		return nil, ListTablesOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to get connection: %v", err),
		}, nil
	}

	cfg, err := connection.FromRaw(connData.Dialect, connData.DSN, fmt.Sprintf("%t", connData.ReadOnly))
	if err != nil {
		return nil, ListTablesOutput{
			Success: false,
			Error:   fmt.Sprintf("invalid connection config: %v", err),
		}, nil
	}

	adapter, err := deps.ConnManager.GetAdapter(*cfg)
	if err != nil {
		return nil, ListTablesOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to get adapter: %v", err),
		}, nil
	}

	tables, err := adapter.ListTables(ctx, input.Schema)
	if err != nil {
		return nil, ListTablesOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to list tables: %v", err),
		}, nil
	}

	tableInfos := make([]TableInfo, len(tables))
	for i, t := range tables {
		tableInfos[i] = TableInfo{
			Name:   t.Name,
			Schema: t.Schema,
			Type:   t.Type,
		}
	}

	return nil, ListTablesOutput{
		Success: true,
		Tables:  tableInfos,
	}, nil
}

func DescribeTableTool(ctx context.Context, req *mcp.CallToolRequest, input DescribeTableInput, deps *ToolDependencies) (*mcp.CallToolResult, DescribeTableOutput, error) {
	claims, ok := GetClaims(ctx)
	if !ok {
		return nil, DescribeTableOutput{
			Success: false,
			Error:   "unauthorized: missing claims",
		}, nil
	}

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		return nil, DescribeTableOutput{
			Success: false,
			Error:   "unauthorized: missing tenant_id",
		}, nil
	}

	if !claims.HasAccessToConnection(input.ConnectionID) {
		return nil, DescribeTableOutput{
			Success: false,
			Error:   fmt.Sprintf("access denied to connection: %s", input.ConnectionID),
		}, nil
	}

	connData, err := deps.ConnRepo.GetConnectionWithDSN(tenantID, input.ConnectionID)
	if err != nil {
		return nil, DescribeTableOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to get connection: %v", err),
		}, nil
	}

	cfg, err := connection.FromRaw(connData.Dialect, connData.DSN, fmt.Sprintf("%t", connData.ReadOnly))
	if err != nil {
		return nil, DescribeTableOutput{
			Success: false,
			Error:   fmt.Sprintf("invalid connection config: %v", err),
		}, nil
	}

	adapter, err := deps.ConnManager.GetAdapter(*cfg)
	if err != nil {
		return nil, DescribeTableOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to get adapter: %v", err),
		}, nil
	}

	schema, err := adapter.DescribeTable(ctx, input.TableName, input.Schema)
	if err != nil {
		return nil, DescribeTableOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to describe table: %v", err),
		}, nil
	}

	columns := make([]ColumnInfo, len(schema.Columns))
	for i, c := range schema.Columns {
		columns[i] = ColumnInfo{
			Name:         c.Name,
			Type:         c.Type,
			Nullable:     c.Nullable,
			DefaultValue: c.DefaultValue,
			IsPrimaryKey: c.IsPrimaryKey,
			IsUnique:     c.IsUnique,
		}
	}

	indexes := make([]IndexInfo, len(schema.Indexes))
	for i, idx := range schema.Indexes {
		indexes[i] = IndexInfo{
			Name:     idx.Name,
			Columns:  idx.Columns,
			IsUnique: idx.IsUnique,
			Type:     idx.Type,
		}
	}

	return nil, DescribeTableOutput{
		Success: true,
		Table:   schema.TableName,
		Columns: columns,
		Indexes: indexes,
	}, nil
}
