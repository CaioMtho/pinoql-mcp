package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main(){
	godotenv.Load()

	server := mcp.NewServer(&mcp.Implementation{Title: "Pinoql MCP Server", Version: "v0.1.0"}, nil)
	
	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server {return server},
		&mcp.StreamableHTTPOptions{},
	)
	
	http.HandleFunc("/mcp", handler.ServeHTTP)

	log.Println("Starting server on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}