package token

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/CaioMtho/pinoql-mcp/internal/credentials/claims"
	"github.com/CaioMtho/pinoql-mcp/internal/credentials/connection_data"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTHandler struct {
	tokenRepo *Repository
	connRepo  *connection_data.Repository
	jwtSecret string
}

func NewJWTHandler(
	tokenRepo *Repository,
	connRepo *connection_data.Repository,
	jwtSecret string,
) *JWTHandler {
	return &JWTHandler{
		tokenRepo: tokenRepo,
		connRepo:  connRepo,
		jwtSecret: jwtSecret,
	}
}

func (h *JWTHandler) IssueToken(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	var req claims.JWTIssueRequest
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
	jti := uuid.New().String()

	pinoqlClaims := claims.PinoQLClaims{
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
		Permissions:   req.Permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, pinoqlClaims)
	signedToken, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	connectionIDsJSON, _ := json.Marshal(req.ConnectionIDs)
	newToken := NewJWTToken{
		TenantID:      tenantID,
		ConnectionIDs: string(connectionIDsJSON),
		IssuedAt:      now,
		ExpiresAt:     expiresAt,
	}

	err = h.tokenRepo.InsertToken(jti, newToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store token"})
		return
	}

	response := claims.JWTIssueResponse{
		Token:     signedToken,
		ExpiresAt: expiresAt.Unix(),
		JTI:       jti,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *JWTHandler) RevokeToken(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	var req RevokeJWTToken
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.tokenRepo.GetTokenByJTI(req.JTI)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}

	if token.TenantID != tenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	err = h.tokenRepo.RevokeToken(req.JTI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token revoked successfully"})
}

func (h *JWTHandler) ListTokens(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	limit := 50
	offset := 0

	tokens, err := h.tokenRepo.ListTokensByTenant(tenantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}
