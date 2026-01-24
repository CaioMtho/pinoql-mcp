package mcp

import (
	"crypto/sha256"
	"fmt"

	"github.com/CaioMtho/pinoql-mcp/internal/audit"
)

func NewAuditLog(tenantID, connectionID, action, query string, success bool, errorMsg string, execTime, rowsAffected int) audit.NewConnectionAuditLog {
	queryHash := hashQuery(query)
	execTimePtr := &execTime
	rowsAffectedPtr := &rowsAffected
	errorMsgPtr := &errorMsg

	if errorMsg == "" {
		errorMsgPtr = nil
	}

	return audit.NewConnectionAuditLog{
		TenantID:        tenantID,
		ConnectionID:    connectionID,
		Action:          action,
		QueryHash:       &queryHash,
		Success:         success,
		ErrorMessage:    errorMsgPtr,
		ExecutionTimeMs: execTimePtr,
		RowsAffected:    rowsAffectedPtr,
	}
}

func hashQuery(query string) string {
	hash := sha256.Sum256([]byte(query))
	return fmt.Sprintf("%x", hash[:8])
}
