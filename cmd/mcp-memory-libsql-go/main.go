package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ZanzyTHEbar/mcp-memory-libsql-go/internal/database"
	"github.com/ZanzyTHEbar/mcp-memory-libsql-go/internal/server"
)

var (
	libsqlURL   = flag.String("libsql-url", "", "libSQL database URL (default: file:./libsql.db)")
	authToken   = flag.String("auth-token", "", "Authentication token for remote databases")
	projectsDir = flag.String("projects-dir", "", "Base directory for projects. Enables multi-project mode.")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, closing server...")
		cancel()
	}()

	// Initialize database configuration
	config := database.NewConfig()

	// Override with command line flags if provided
	if *libsqlURL != "" {
		config.URL = *libsqlURL
	}
	if *authToken != "" {
		config.AuthToken = *authToken
	}
	if *projectsDir != "" {
		config.ProjectsDir = *projectsDir
		config.MultiProjectMode = true
	}

	// Create database manager
	db, err := database.NewDBManager(config)
	if err != nil {
		log.Fatalf("Failed to create database manager: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Create MCP server
	mcpServer := server.NewMCPServer(db)

	// Run the server
	log.Println("Starting MCP Memory LibSQL server...")
	go func() {
		if err := mcpServer.Run(ctx); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Server stopped")
}
