package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/CaioMtho/pinoql-mcp/internal/credentials/audit"
	"github.com/CaioMtho/pinoql-mcp/internal/credentials/connection_data"
	"github.com/CaioMtho/pinoql-mcp/internal/credentials/middleware"
	"github.com/CaioMtho/pinoql-mcp/internal/credentials/tenant"
	"github.com/CaioMtho/pinoql-mcp/internal/credentials/token"
	"github.com/CaioMtho/pinoql-mcp/internal/crypto"
	"github.com/CaioMtho/pinoql-mcp/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err.Error())
	}

	masterKeyB64 := os.Getenv("MASTER_KEY")
	if masterKeyB64 == "" {
		log.Fatal("MASTER_KEY environment variable is required")
	}

	masterKey, err := base64.StdEncoding.DecodeString(masterKeyB64)
	if err != nil {
		log.Fatalf("Invalid MASTER_KEY: %v", err)
	}

	cryptoManager, err := crypto.NewCryptoManager(masterKey)
	if err != nil {
		fmt.Println("Decoded master key length:", len(masterKey))
		log.Fatalf("Failed to create crypto manager: %v", err)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./db/pinoql.sqlite"
	}

	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close connection: %v", err.Error())
		}
	}(db)

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	connDataRepo := connection_data.NewConnectionDataRepository(db, cryptoManager)
	tenantRepo := tenant.NewTenantRepository(db)
	tokenRepo := token.NewJWTTokenRepository(db)
	auditRepo := audit.NewAuditLogRepository(db)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	connDataHandler := connection_data.NewConnectionHandler(connDataRepo)
	tokenHandler := token.NewJWTHandler(tokenRepo, connDataRepo, jwtSecret)
	tenantHandler := tenant.NewTenantHandler(tenantRepo)
	auditHandler := audit.NewAuditHandler(auditRepo)

	authMiddleware := middleware.NewAuthMiddleware(jwtSecret, tokenRepo)

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Title:   "Pinoql MCP Server",
		Version: "v0.1.0",
	}, nil)

	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return mcpServer },
		&mcp.StreamableHTTPOptions{},
	)

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	routerConfig := &routes.RouterConfig{
		ConnectionDataHandler: connDataHandler,
		TokenHandler:          tokenHandler,
		TenantHandler:         tenantHandler,
		AuditHandler:          auditHandler,
		AuthMiddleware:        authMiddleware,
		MCPHandler:            mcpHandler,
	}

	routes.SetupRoutes(r, routerConfig)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Pinoql MCP Server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
