package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading env file: %v", err.Error())
	}

	server := mcp.NewServer(&mcp.Implementation{
		Title:   "Pinoql MCP Server",
		Version: "v0.1.0",
	}, nil)

	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{},
	)

	r := gin.Default()

	r.Any("/mcp", gin.WrapF(handler.ServeHTTP))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
