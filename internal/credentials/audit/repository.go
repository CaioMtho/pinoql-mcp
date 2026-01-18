package audit

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewAuditLogRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertLog(data NewConnectionAuditLog) error {
	query := `
		INSERT INTO connection_audit_log (
			tenant_id, connection_id, action, query_hash, success, 
			error_message, execution_time_ms, rows_affected
		)
		VALUES (
			:tenant_id, :connection_id, :action, :query_hash, :success,
			:error_message, :execution_time_ms, :rows_affected
		)
	`

	_, err := r.db.NamedExec(query, data)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	return nil
}

func (r *Repository) ListLogs(filter AuditLogQuery) ([]*ConnectionAuditLog, error) {
	var logs []*ConnectionAuditLog

	var whereClauses []string
	params := map[string]interface{}{}

	if filter.TenantID != nil {
		whereClauses = append(whereClauses, "tenant_id = :tenant_id")
		params["tenant_id"] = *filter.TenantID
	}

	if filter.ConnectionID != nil {
		whereClauses = append(whereClauses, "connection_id = :connection_id")
		params["connection_id"] = *filter.ConnectionID
	}

	if filter.Action != nil {
		whereClauses = append(whereClauses, "action = :action")
		params["action"] = *filter.Action
	}

	if filter.Success != nil {
		whereClauses = append(whereClauses, "success = :success")
		params["success"] = *filter.Success
	}

	if filter.StartTime != nil {
		whereClauses = append(whereClauses, "timestamp >= :start_time")
		params["start_time"] = *filter.StartTime
	}

	if filter.EndTime != nil {
		whereClauses = append(whereClauses, "timestamp <= :end_time")
		params["end_time"] = *filter.EndTime
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	params["limit"] = filter.Limit
	params["offset"] = filter.Offset

	query := fmt.Sprintf(`
		SELECT 
			id, tenant_id, connection_id, action, query_hash, 
			success, error_message, execution_time_ms, rows_affected, timestamp
		FROM connection_audit_log
		%s
		ORDER BY timestamp DESC
		LIMIT :limit OFFSET :offset
	`, whereClause)

	rows, err := r.db.NamedQuery(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
		}
	}(rows)

	for rows.Next() {
		var log ConnectionAuditLog
		err := rows.StructScan(&log)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

func (r *Repository) GetStatsByConnection(tenantID, connectionID string) (*AuditLogStats, error) {
	var stats AuditLogStats

	query := `
		SELECT 
			tenant_id,
			connection_id,
			COUNT(*) as total_queries,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful_queries,
			SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failed_queries,
			AVG(COALESCE(execution_time_ms, 0)) as avg_execution_ms,
			SUM(COALESCE(rows_affected, 0)) as total_rows_affected
		FROM connection_audit_log
		WHERE tenant_id = ? AND connection_id = ?
		GROUP BY tenant_id, connection_id
	`

	err := r.db.Get(&stats, query, tenantID, connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit stats: %w", err)
	}

	return &stats, nil
}

func (r *Repository) GetStatsByTenant(tenantID string) ([]*AuditLogStats, error) {
	var stats []*AuditLogStats

	query := `
		SELECT 
			tenant_id,
			connection_id,
			COUNT(*) as total_queries,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful_queries,
			SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failed_queries,
			AVG(COALESCE(execution_time_ms, 0)) as avg_execution_ms,
			SUM(COALESCE(rows_affected, 0)) as total_rows_affected
		FROM connection_audit_log
		WHERE tenant_id = ?
		GROUP BY tenant_id, connection_id
		ORDER BY total_queries DESC
	`

	err := r.db.Select(&stats, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant stats: %w", err)
	}

	return stats, nil
}

func (r *Repository) DeleteOldLogs(daysToKeep int) (int64, error) {
	query := `
		DELETE FROM connection_audit_log
		WHERE timestamp < datetime('now', '-' || ? || ' days')
	`

	result, err := r.db.Exec(query, daysToKeep)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
