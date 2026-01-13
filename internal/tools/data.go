package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type QueryInput struct {
	SQL string `json:"sql" jsonschema:"o sql a ser executado"`
}

type QueryOutput struct {
	Rows []map[string]any `json:"rows"`
}

func RunQuery(ctx context.Context, req *mcp.CallToolRequest, input QueryInput) (*mcp.CallToolResult, *QueryOutput, error) {
	return nil, nil, nil
}
