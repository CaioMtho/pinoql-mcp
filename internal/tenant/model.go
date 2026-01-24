package tenant

import "time"

type Tenant struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type NewTenant struct {
	Name string `json:"name" db:"name" validate:"required"`
}

type UpdateTenant struct {
	Name     *string `json:"name,omitempty" db:"name"`
	IsActive *bool   `json:"is_active,omitempty" db:"is_active"`
}

type TenantQuery struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (Tenant) TableName() string {
	return "tenants"
}
