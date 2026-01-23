package middleware

import (
	"net/http"
	"strings"

	"github.com/CaioMtho/pinoql-mcp/internal/credentials/claims"
	"github.com/CaioMtho/pinoql-mcp/internal/credentials/token"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	jwtSecret string
	tokenRepo *token.Repository
	apiKey    string
}

func NewAuthMiddleware(jwtSecret string, tokenRepo *token.Repository, apiKey string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
		tokenRepo: tokenRepo,
		apiKey:    apiKey,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		authToken, err := jwt.ParseWithClaims(tokenString, &claims.PinoQLClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		pinoqlClaims, ok := authToken.Claims.(*claims.PinoQLClaims)
		if !ok || !authToken.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		revoked, err := m.tokenRepo.IsTokenRevoked(pinoqlClaims.ID)
		if err != nil || revoked {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
			c.Abort()
			return
		}

		c.Set("tenant_id", pinoqlClaims.TenantID)
		c.Set("claims", pinoqlClaims)

		c.Next()
	}
}

func (m *AuthMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			c.Abort()
			return
		}

		if !m.validateAPIKey(apiKey) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) validateAPIKey(apiKey string) bool {
	return m.apiKey != "" && m.apiKey == apiKey
}
