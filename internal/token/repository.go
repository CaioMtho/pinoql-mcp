package token

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

type InsertTokenData struct {
	JTI           string
	TenantID      string
	ConnectionIDs string
	IssuedAt      time.Time
	ExpiresAt     time.Time
}

func (r *Repository) InsertToken(data InsertTokenData) error {
	query := `
		INSERT INTO jwt_tokens (jti, tenant_id, connection_ids, issued_at, expires_at, revoked)
		VALUES (?, ?, ?, ?, ?, 0)
	`

	_, err := r.db.Exec(query, data.JTI, data.TenantID, data.ConnectionIDs, data.IssuedAt, data.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to insert JWT token: %w", err)
	}

	return nil
}

func (r *Repository) GetTokenByJTI(jti string) (*JWTToken, error) {
	var token JWTToken

	query := `
		SELECT jti, tenant_id, connection_ids, issued_at, expires_at, revoked, revoked_at
		FROM jwt_tokens
		WHERE jti = ?
	`

	err := r.db.Get(&token, query, jti)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return &token, nil
}

func (r *Repository) IsTokenRevoked(jti string) (bool, error) {
	var revoked bool

	query := `SELECT revoked FROM jwt_tokens WHERE jti = ?`

	err := r.db.Get(&revoked, query, jti)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check revocation: %w", err)
	}

	return revoked, nil
}

func (r *Repository) RevokeToken(jti string) error {
	query := `
		UPDATE jwt_tokens 
		SET revoked = 1, revoked_at = CURRENT_TIMESTAMP
		WHERE jti = ?
	`

	result, err := r.db.Exec(query, jti)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
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

func (r *Repository) ListTokensByTenant(tenantID string, limit, offset int) ([]*JWTTokenInfo, error) {
	var tokens []*JWTTokenInfo

	query := `
		SELECT jti, tenant_id, issued_at, expires_at, revoked
		FROM jwt_tokens
		WHERE tenant_id = ?
		ORDER BY issued_at DESC
		LIMIT ? OFFSET ?
	`

	err := r.db.Select(&tokens, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tokens: %w", err)
	}

	now := time.Now()
	for _, token := range tokens {
		token.IsExpired = now.After(token.ExpiresAt)
	}

	return tokens, nil
}

func (r *Repository) DeleteExpiredTokens() (int64, error) {
	query := `
		DELETE FROM jwt_tokens
		WHERE expires_at < CURRENT_TIMESTAMP
	`

	result, err := r.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (r *Repository) CountActiveTokensByTenant(tenantID string) (int, error) {
	var count int

	query := `
		SELECT COUNT(*) 
		FROM jwt_tokens
		WHERE tenant_id = ? AND revoked = 0 AND expires_at > CURRENT_TIMESTAMP
	`

	err := r.db.Get(&count, query, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to count tokens: %w", err)
	}

	return count, nil
}

func GenerateJTI() string {
	return uuid.New().String()
}

func SerializeConnectionIDs(ids []string) (string, error) {
	data, err := json.Marshal(ids)
	if err != nil {
		return "", fmt.Errorf("failed to serialize connection IDs: %w", err)
	}
	return string(data), nil
}

func DeserializeConnectionIDs(data string) ([]string, error) {
	var ids []string
	err := json.Unmarshal([]byte(data), &ids)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize connection IDs: %w", err)
	}
	return ids, nil
}
