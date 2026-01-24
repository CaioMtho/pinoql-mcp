package token

import "time"

type JWTToken struct {
	JTI           string     `json:"jti" db:"jti"`
	TenantID      string     `json:"tenant_id" db:"tenant_id"`
	ConnectionIDs string     `json:"connection_ids" db:"connection_ids"`
	IssuedAt      time.Time  `json:"issued_at" db:"issued_at"`
	ExpiresAt     time.Time  `json:"expires_at" db:"expires_at"`
	Revoked       bool       `json:"revoked" db:"revoked"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

type NewJWTToken struct {
	TenantID      string   `json:"tenant_id" validate:"required"`
	ConnectionIDs []string `json:"connection_ids" validate:"required,min=1"`
	TTLSeconds    int      `json:"ttl_seconds" validate:"required,min=60,max=86400"`
}

type JWTTokenQuery struct {
	TenantID      *string    `json:"tenant_id,omitempty"`
	Revoked       *bool      `json:"revoked,omitempty"`
	ExpiredBefore *time.Time `json:"expired_before,omitempty"`
	IssuedAfter   *time.Time `json:"issued_after,omitempty"`
	Limit         int        `json:"limit" validate:"min=1,max=1000"`
	Offset        int        `json:"offset" validate:"min=0"`
}

type RevokeJWTToken struct {
	JTI string `json:"jti" validate:"required"`
}

type JWTTokenInfo struct {
	JTI       string    `json:"jti" db:"jti"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	IssuedAt  time.Time `json:"issued_at" db:"issued_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Revoked   bool      `json:"revoked" db:"revoked"`
	IsExpired bool      `json:"is_expired"`
}

func (t *JWTToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *JWTToken) IsValid() bool {
	return !t.Revoked && !t.IsExpired()
}

func (t *JWTToken) TableName() string {
	return "jwt_tokens"
}
