package audit

import "time"

type ConnectionAuditLog struct {
	ID              int64     `json:"id" db:"id"`
	TenantID        string    `json:"tenant_id" db:"tenant_id"`
	ConnectionID    string    `json:"connection_id" db:"connection_id"`
	Action          string    `json:"action" db:"action"` // 'query', 'schema', 'connect'
	QueryHash       *string   `json:"query_hash,omitempty" db:"query_hash"`
	Success         bool      `json:"success" db:"success"`
	ErrorMessage    *string   `json:"error_message,omitempty" db:"error_message"`
	ExecutionTimeMs *int      `json:"execution_time_ms,omitempty" db:"execution_time_ms"`
	RowsAffected    *int      `json:"rows_affected,omitempty" db:"rows_affected"`
	Timestamp       time.Time `json:"timestamp" db:"timestamp"`
}

type NewConnectionAuditLog struct {
	TenantID        string  `json:"tenant_id" db:"tenant_id" validate:"required"`
	ConnectionID    string  `json:"connection_id" db:"connection_id" validate:"required"`
	Action          string  `json:"action" db:"action" validate:"required,oneof=query schema connect"`
	QueryHash       *string `json:"query_hash,omitempty" db:"query_hash"`
	Success         bool    `json:"success" db:"success"`
	ErrorMessage    *string `json:"error_message,omitempty" db:"error_message"`
	ExecutionTimeMs *int    `json:"execution_time_ms,omitempty" db:"execution_time_ms"`
	RowsAffected    *int    `json:"rows_affected,omitempty" db:"rows_affected"`
}

type AuditLogQuery struct {
	TenantID     *string    `json:"tenant_id,omitempty"`
	ConnectionID *string    `json:"connection_id,omitempty"`
	Action       *string    `json:"action,omitempty"`
	Success      *bool      `json:"success,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Limit        int        `json:"limit" validate:"min=1,max=1000"`
	Offset       int        `json:"offset" validate:"min=0"`
}

type AuditLogStats struct {
	TenantID          string  `json:"tenant_id"`
	ConnectionID      string  `json:"connection_id"`
	TotalQueries      int64   `json:"total_queries"`
	SuccessfulQueries int64   `json:"successful_queries"`
	FailedQueries     int64   `json:"failed_queries"`
	AvgExecutionMs    float64 `json:"avg_execution_ms"`
	TotalRowsAffected int64   `json:"total_rows_affected"`
}

func (ConnectionAuditLog) TableName() string {
	return "connection_audit_log"
}
