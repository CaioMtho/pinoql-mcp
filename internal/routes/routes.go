package routes

import (
	"net/http"

	"github.com/CaioMtho/pinoql-mcp/internal/audit"
	"github.com/CaioMtho/pinoql-mcp/internal/connection_data"
	"github.com/CaioMtho/pinoql-mcp/internal/middleware"
	"github.com/CaioMtho/pinoql-mcp/internal/tenant"
	"github.com/CaioMtho/pinoql-mcp/internal/token"
	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	ConnectionDataHandler *connection_data.ConnectionHandler
	TokenHandler          *token.JWTHandler
	TenantHandler         *tenant.Handler
	AuditHandler          *audit.Handler
	AuthMiddleware        *middleware.AuthMiddleware
	MCPHandler            http.Handler
	APIKey                string
	EnableAuth            bool
}

func SetupRoutes(r *gin.Engine, cfg *RouterConfig) {
	api := r.Group("/api/v1")

	if cfg.EnableAuth {
		api.Use(cfg.AuthMiddleware.RequireAPIKey())
	}

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	tenants := api.Group("/tenants")
	{
		tenants.POST("", cfg.TenantHandler.CreateTenant)
		tenants.GET("", cfg.TenantHandler.ListTenants)
		tenants.GET("/:id", cfg.TenantHandler.GetTenant)
		tenants.PUT("/:id", cfg.TenantHandler.UpdateTenant)
		tenants.DELETE("/:id", cfg.TenantHandler.DeleteTenant)
	}

	connections := api.Group("/connections")
	connections.Use()
	{
		connections.POST("", cfg.ConnectionDataHandler.CreateConnection)
		connections.GET("", cfg.ConnectionDataHandler.ListConnections)
		connections.GET("/:id", cfg.ConnectionDataHandler.GetConnection)
		connections.PUT("/:id", cfg.ConnectionDataHandler.UpdateConnection)
		connections.DELETE("/:id", cfg.ConnectionDataHandler.DeleteConnection)
	}

	jwt := api.Group("/jwt")
	jwt.Use()
	{
		jwt.POST("/issue", cfg.TokenHandler.Issue)
		jwt.POST("/revoke", cfg.TokenHandler.Revoke)
		jwt.GET("/list", cfg.TokenHandler.List)
	}

	auditRoutes := api.Group("/audit")
	auditRoutes.Use()
	{
		auditRoutes.GET("/logs", cfg.AuditHandler.ListLogs)
		auditRoutes.GET("/stats", cfg.AuditHandler.GetStats)
	}

	mcpGroup := r.Group("/mcp")
	mcpGroup.Use(cfg.AuthMiddleware.RequireAuth())
	{
		mcpGroup.Any("", gin.WrapH(cfg.MCPHandler))
	}
}
