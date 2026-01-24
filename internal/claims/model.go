package claims

import (
	"github.com/golang-jwt/jwt/v5"
)

type PinoQLClaims struct {
	jwt.RegisteredClaims
	TenantID      string                `json:"tenant_id"`
	ConnectionIDs []string              `json:"connection_ids"`
	Permissions   ConnectionPermissions `json:"permissions"`
}

type ConnectionPermissions struct {
	Read       bool     `json:"read"`
	Write      bool     `json:"write"`
	Schema     bool     `json:"schema"`
	DDL        bool     `json:"ddl"`
	MaxRows    int      `json:"max_rows"`
	AllowedOps []string `json:"allowed_ops"`
}

func DefaultReadOnlyPermissions() ConnectionPermissions {
	return ConnectionPermissions{
		Read:       true,
		Write:      false,
		Schema:     true,
		DDL:        false,
		MaxRows:    1000,
		AllowedOps: []string{"SELECT"},
	}
}

func DefaultReadWritePermissions() ConnectionPermissions {
	return ConnectionPermissions{
		Read:       true,
		Write:      true,
		Schema:     true,
		DDL:        false,
		MaxRows:    0,
		AllowedOps: []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
	}
}

func FullPermissions() ConnectionPermissions {
	return ConnectionPermissions{
		Read:       true,
		Write:      true,
		Schema:     true,
		DDL:        true,
		MaxRows:    0,
		AllowedOps: []string{"*"},
	}
}

func (c *PinoQLClaims) HasAccessToConnection(connectionID string) bool {
	for _, id := range c.ConnectionIDs {
		if id == connectionID || id == "*" {
			return true
		}
	}
	return false
}

func (c *PinoQLClaims) CanExecuteOperation(operation string) bool {
	if len(c.Permissions.AllowedOps) == 0 {
		return false
	}

	for _, op := range c.Permissions.AllowedOps {
		if op == "*" || op == operation {
			return true
		}
	}
	return false
}

func (c *PinoQLClaims) CanRead() bool {
	return c.Permissions.Read
}

func (c *PinoQLClaims) CanWrite() bool {
	return c.Permissions.Write
}

func (c *PinoQLClaims) CanAccessSchema() bool {
	return c.Permissions.Schema
}

func (c *PinoQLClaims) CanExecuteDDL() bool {
	return c.Permissions.DDL
}

func (c *PinoQLClaims) GetMaxRows() int {
	return c.Permissions.MaxRows
}

type JWTIssueRequest struct {
	TenantID      string                `json:"tenant_id" validate:"required"`
	ConnectionIDs []string              `json:"connection_ids" validate:"required,min=1"`
	Permissions   ConnectionPermissions `json:"permissions"`
	TTLSeconds    int                   `json:"ttl_seconds" validate:"min=60,max=86400"`
}

type JWTIssueResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	JTI       string `json:"jti"`
}
