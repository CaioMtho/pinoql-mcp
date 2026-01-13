package connection_data

import "time"

type ConnectionData struct {
	ID             string    `json:"id" db:"id"`
	TenantID       string    `json:"tenant_id" db:"tenant_id"`
	Name           string    `json:"name" db:"name"`
	Description    *string   `json:"description,omitempty" db:"description"`
	DSN            string    `json:"dsn" db:"dsn"`
	Dialect        string    `json:"dialect" db:"dialect"`
	DEK            string    `json:"dek" db:"dek"`
	ReadOnly       bool      `json:"readonly" db:"readonly"`
	MaxConnections int       `json:"max_connections" db:"max_connections"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type NewConnectionData struct {
	TenantID       string  `json:"tenant_id" db:"tenant_id" validate:"required"`
	Name           string  `json:"name" db:"name" validate:"required"`
	Description    *string `json:"description,omitempty" db:"description"`
	DSN            string  `json:"dsn" db:"dsn" validate:"required"`
	Dialect        string  `json:"dialect" db:"dialect" validate:"required,oneof=postgresql mysql sqlite"`
	ReadOnly       bool    `json:"readonly" db:"readonly"`
	MaxConnections int     `json:"max_connections" db:"max_connections" validate:"min=1,max=100"`
}

type UpdateConnectionData struct {
	Name           *string `json:"name,omitempty" db:"name"`
	Description    *string `json:"description,omitempty" db:"description"`
	DSN            *string `json:"dsn,omitempty" db:"dsn"`
	Dialect        *string `json:"dialect,omitempty" db:"dialect" validate:"omitempty,oneof=postgresql mysql sqlite"`
	ReadOnly       *bool   `json:"readonly,omitempty" db:"readonly"`
	MaxConnections *int    `json:"max_connections,omitempty" db:"max_connections" validate:"omitempty,min=1,max=100"`
	IsActive       *bool   `json:"is_active,omitempty" db:"is_active"`
}

type ConnectionDataQuery struct {
	ID             string    `json:"id" db:"id"`
	TenantID       string    `json:"tenant_id" db:"tenant_id"`
	Name           string    `json:"name" db:"name"`
	Description    *string   `json:"description,omitempty" db:"description"`
	Dialect        string    `json:"dialect" db:"dialect"`
	ReadOnly       bool      `json:"readonly" db:"readonly"`
	MaxConnections int       `json:"max_connections" db:"max_connections"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

func (ConnectionData) TableName() string {
	return "connection_data"
}
