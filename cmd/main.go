package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	audit2 "github.com/CaioMtho/pinoql-mcp/internal/audit"
	"github.com/CaioMtho/pinoql-mcp/internal/connection"
	connection_data2 "github.com/CaioMtho/pinoql-mcp/internal/connection_data"
	"github.com/CaioMtho/pinoql-mcp/internal/crypto"
	"github.com/CaioMtho/pinoql-mcp/internal/mcp"
	"github.com/CaioMtho/pinoql-mcp/internal/middleware"
	"github.com/CaioMtho/pinoql-mcp/internal/routes"
	tenant2 "github.com/CaioMtho/pinoql-mcp/internal/tenant"
	token2 "github.com/CaioMtho/pinoql-mcp/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/pflag"
)

var apiKey string
var enableAuth bool

func init() {
	pflag.StringVar(&apiKey, "api-key", "", "API key for authentication")
	pflag.BoolVar(&enableAuth, "enable-auth", false, "Enable authentication")
	pflag.Parse()
}

func generateAPIKey() string {
	key := uuid.New().String()
	encodedKey := base64.StdEncoding.EncodeToString([]byte(key))
	fmt.Println("Generated API Key:", encodedKey)
	os.Setenv("API_KEY", encodedKey)
	return encodedKey
}

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

	connManager := connection.NewConnectionManager()

	connDataRepo := connection_data2.NewConnectionDataRepository(db, cryptoManager)
	tenantRepo := tenant2.NewTenantRepository(db)
	tokenRepo := token2.NewRepository(db)
	auditRepo := audit2.NewAuditLogRepository(db)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	connDataHandler := connection_data2.NewConnectionHandler(connDataRepo)
	tokenHandler := token2.NewHandler(tokenRepo, connDataRepo, jwtSecret)
	tenantHandler := tenant2.NewTenantHandler(tenantRepo)
	auditHandler := audit2.NewAuditHandler(auditRepo)

	var generatedAPIKey string
	if enableAuth {
		generatedAPIKey = generateAPIKey()
	}

	authMiddleware := middleware.NewAuthMiddleware(jwtSecret, tokenRepo, generatedAPIKey)

	mcpServer := mcpsdk.NewServer(&mcpsdk.Implementation{
		Title:   "Pinoql MCP Server",
		Version: "v0.1.0",
	}, nil)

	toolDeps := &mcp.ToolDependencies{
		ConnManager: connManager,
		ConnRepo:    connDataRepo,
		AuditRepo:   auditRepo,
	}

	mcp.RegisterTools(mcpServer, toolDeps)

	baseMCPHandler := mcpsdk.NewStreamableHTTPHandler(
		func(*http.Request) *mcpsdk.Server { return mcpServer },
		&mcpsdk.StreamableHTTPOptions{},
	)

	mcpHandler := mcp.NewMCPContextHandler(baseMCPHandler)

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
		APIKey:                generatedAPIKey,
		EnableAuth:            enableAuth,
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
