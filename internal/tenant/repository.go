package tenant

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewTenantRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertTenant(data NewTenant) (*TenantQuery, error) {
	id := generateTenantID()

	query := `
		INSERT INTO tenants (id, name, is_active)
		VALUES (?, ?, 1)
	`

	_, err := r.db.Exec(query, id, data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to insert tenant: %w", err)
	}

	return r.GetTenantByID(id)
}

func (r *Repository) GetTenantByID(id string) (*TenantQuery, error) {
	var tenant Tenant

	query := `
		SELECT id, name, is_active, created_at
		FROM tenants
		WHERE id = ? AND is_active = 1
	`

	err := r.db.Get(&tenant, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &TenantQuery{
		ID:        tenant.ID,
		Name:      tenant.Name,
		IsActive:  tenant.IsActive,
		CreatedAt: tenant.CreatedAt,
	}, nil
}

func (r *Repository) ListTenants() ([]*TenantQuery, error) {
	var tenants []*TenantQuery

	query := `
		SELECT id, name, is_active, created_at
		FROM tenants
		WHERE is_active = 1
		ORDER BY created_at DESC
	`

	err := r.db.Select(&tenants, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, nil
}

func (r *Repository) UpdateTenant(id string, update UpdateTenant) error {
	query := `
		UPDATE tenants SET
			name = COALESCE(:name, name),
			is_active = COALESCE(:is_active, is_active),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`

	params := map[string]interface{}{
		"id":        id,
		"name":      update.Name,
		"is_active": update.IsActive,
	}

	result, err := r.db.NamedExec(query, params)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
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

func (r *Repository) DeleteTenant(id string) error {
	query := `
		UPDATE tenants 
		SET is_active = 0, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete enant: %w", err)
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

func (r *Repository) HardDeleteTenant(id string) error {
	query := `DELETE FROM tenants WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete tenant: %w", err)
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

func generateTenantID() string {
	return fmt.Sprintf("tenant_%s", uuid.New().String()[:12])
}
