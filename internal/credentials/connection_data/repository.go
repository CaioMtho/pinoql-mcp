package connection_data

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/CaioMtho/pinoql-mcp/internal/crypto"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ConnectionDataRepository struct {
	db *sqlx.DB
	cm *crypto.CryptoManager
}

func NewConnectionDataRepository(db *sqlx.DB, cm *crypto.CryptoManager) *ConnectionDataRepository {
	return &ConnectionDataRepository{
		db: db,
		cm: cm,
	}
}

func (r *ConnectionDataRepository) InsertConnection(data NewConnectionData) (*ConnectionDataQuery, error) {
	envelope, err := r.cm.Encrypt([]byte(data.DSN))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt DSN: %w", err)
	}

	id := generateConnectionID()

	params := map[string]interface{}{
		"id":              id,
		"tenant_id":       data.TenantID,
		"name":            data.Name,
		"description":     data.Description,
		"dsn":             envelope.CiphertextHex,
		"dialect":         data.Dialect,
		"dek":             envelope.WrappedDEKHex,
		"readonly":        data.ReadOnly,
		"max_connections": data.MaxConnections,
		"is_active":       true,
	}

	query := `
		INSERT INTO connection_data (
			id, tenant_id, name, description, dsn, dialect, dek, 
			readonly, max_connections, is_active
		)
		VALUES (
			:id, :tenant_id, :name, :description, :dsn, :dialect, :dek,
			:readonly, :max_connections, :is_active
		)`

	_, err = r.db.NamedExec(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to insert connection: %w", err)
	}

	return r.GetConnectionByID(data.TenantID, id)
}

func (r *ConnectionDataRepository) GetConnectionByID(tenantID, connectionID string) (*ConnectionDataQuery, error) {
	var cred ConnectionData

	query := `
		SELECT
			id, tenant_id, name, description, dialect, 
			readonly, max_connections, is_active, created_at, updated_at
		FROM connection_data
		WHERE id = ? AND tenant_id = ? AND is_active = 1
	`

	err := r.db.Get(&cred, query, connectionID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("connection not found or access denied")
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	return &ConnectionDataQuery{
		ID:             cred.ID,
		TenantID:       cred.TenantID,
		Name:           cred.Name,
		Description:    cred.Description,
		Dialect:        cred.Dialect,
		ReadOnly:       cred.ReadOnly,
		MaxConnections: cred.MaxConnections,
		IsActive:       cred.IsActive,
		CreatedAt:      cred.CreatedAt,
		UpdatedAt:      cred.UpdatedAt,
	}, nil
}

func (r *ConnectionDataRepository) GetConnectionWithDSN(tenantID, connectionID string) (*ConnectionData, error) {
	var cred ConnectionData

	query := `
		SELECT
			id, tenant_id, name, description, dsn, dialect, dek,
			readonly, max_connections, is_active, created_at, updated_at
		FROM connection_data
		WHERE id = ? AND tenant_id = ? AND is_active = 1
	`

	err := r.db.Get(&cred, query, connectionID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("connection not found or access denied")
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	envelope := &crypto.Envelope{
		CiphertextHex: cred.DSN,
		WrappedDEKHex: cred.DEK,
	}

	plainDSN, err := r.cm.Decrypt(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt DSN: %w", err)
	}

	cred.DSN = string(plainDSN)
	return &cred, nil
}

func (r *ConnectionDataRepository) ListConnections(tenantID string) ([]*ConnectionDataQuery, error) {
	var connections []*ConnectionDataQuery

	query := `
		SELECT
			id, tenant_id, name, description, dialect, 
			readonly, max_connections, is_active, created_at, updated_at
		FROM connection_data
		WHERE tenant_id = ? AND is_active = 1
		ORDER BY created_at DESC
	`

	err := r.db.Select(&connections, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list connections: %w", err)
	}

	return connections, nil
}

func (r *ConnectionDataRepository) UpdateConnection(tenantID, connectionID string, update UpdateConnectionData) error {
	if update.DSN != nil {
		envelope, err := r.cm.Encrypt([]byte(*update.DSN))
		if err != nil {
			return fmt.Errorf("failed to encrypt DSN: %w", err)
		}
		encryptedDSN := envelope.CiphertextHex
		encryptedDEK := envelope.WrappedDEKHex
		update.DSN = &encryptedDSN

		query := `
			UPDATE connection_data SET
				name = COALESCE(:name, name),
				description = COALESCE(:description, description),
				dsn = COALESCE(:dsn, dsn),
				dek = :dek,
				dialect = COALESCE(:dialect, dialect),
				readonly = COALESCE(:readonly, readonly),
				max_connections = COALESCE(:max_connections, max_connections),
				is_active = COALESCE(:is_active, is_active),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = :id AND tenant_id = :tenant_id
		`

		params := map[string]interface{}{
			"id":              connectionID,
			"tenant_id":       tenantID,
			"name":            update.Name,
			"description":     update.Description,
			"dsn":             update.DSN,
			"dek":             encryptedDEK,
			"dialect":         update.Dialect,
			"readonly":        update.ReadOnly,
			"max_connections": update.MaxConnections,
			"is_active":       update.IsActive,
		}

		result, err := r.db.NamedExec(query, params)
		if err != nil {
			return fmt.Errorf("failed to update connection: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	}

	query := `
		UPDATE connection_data SET
			name = COALESCE(:name, name),
			description = COALESCE(:description, description),
			dialect = COALESCE(:dialect, dialect),
			readonly = COALESCE(:readonly, readonly),
			max_connections = COALESCE(:max_connections, max_connections),
			is_active = COALESCE(:is_active, is_active),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND tenant_id = :tenant_id
	`

	params := map[string]interface{}{
		"id":              connectionID,
		"tenant_id":       tenantID,
		"name":            update.Name,
		"description":     update.Description,
		"dialect":         update.Dialect,
		"readonly":        update.ReadOnly,
		"max_connections": update.MaxConnections,
		"is_active":       update.IsActive,
	}

	result, err := r.db.NamedExec(query, params)
	if err != nil {
		return fmt.Errorf("failed to update connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *ConnectionDataRepository) DeleteConnection(tenantID, connectionID string) error {
	query := `
		UPDATE connection_data 
		SET is_active = 0, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND tenant_id = ?
	`

	result, err := r.db.Exec(query, connectionID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *ConnectionDataRepository) HardDeleteConnection(tenantID, connectionID string) error {
	query := `
		DELETE FROM connection_data
		WHERE id = ? AND tenant_id = ?
	`

	result, err := r.db.Exec(query, connectionID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to hard delete connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func generateConnectionID() string {
	return fmt.Sprintf("conn_%s", uuid.New().String()[:8])
}
