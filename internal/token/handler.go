package token

import (
	"net/http"
	"time"

	"github.com/CaioMtho/pinoql-mcp/internal/claims"
	"github.com/CaioMtho/pinoql-mcp/internal/connection_data"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTHandler struct {
	repo      *Repository
	connRepo  *connection_data.Repository
	jwtSecret string
}

func NewHandler(repo *Repository, connRepo *connection_data.Repository, jwtSecret string) *JWTHandler {
	return &JWTHandler{
		repo:      repo,
		connRepo:  connRepo,
		jwtSecret: jwtSecret,
	}
}

func (h *JWTHandler) Issue(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	var req NewJWTToken
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.TenantID = tenantID

	for _, connID := range req.ConnectionIDs {
		_, err := h.connRepo.GetConnectionByID(tenantID, connID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied to connection: " + connID})
			return
		}
	}

	now := time.Now()
	ttl := time.Duration(req.TTLSeconds) * time.Second
	expiresAt := now.Add(ttl)
	jti := GenerateJTI()

	permissions := claims.DefaultReadWritePermissions()

	claimsData := claims.PinoQLClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   tenantID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
			Issuer:    "pinoql-mcp",
		},
		TenantID:      tenantID,
		ConnectionIDs: req.ConnectionIDs,
		Permissions:   permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsData)
	signedToken, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	connectionIDsJSON, err := SerializeConnectionIDs(req.ConnectionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize connection IDs"})
		return
	}

	insertData := InsertTokenData{
		JTI:           jti,
		TenantID:      tenantID,
		ConnectionIDs: connectionIDsJSON,
		IssuedAt:      now,
		ExpiresAt:     expiresAt,
	}

	err = h.repo.InsertToken(insertData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store token"})
		return
	}

	response := map[string]interface{}{
		"token":      signedToken,
		"expires_at": expiresAt.Unix(),
		"jti":        jti,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *JWTHandler) Revoke(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	var req RevokeJWTToken
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.repo.GetTokenByJTI(req.JTI)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}

	if token.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	err = h.repo.RevokeToken(req.JTI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token revoked successfully"})
}

func (h *JWTHandler) List(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	limit := 50
	offset := 0

	tokens, err := h.repo.ListTokensByTenant(tenantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}
