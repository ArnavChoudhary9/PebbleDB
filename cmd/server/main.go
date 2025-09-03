package main

import (
	"log"

	"github.com/ArnavChoudhary9/PebbleDB/internal/config"
	"github.com/ArnavChoudhary9/PebbleDB/internal/handlers"
	"github.com/ArnavChoudhary9/PebbleDB/internal/server"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatal("Configuration validation failed:", err)
	}

	// Create server instance
	srv := server.NewServer()

	// Setup routes and middleware
	handlers.SetupRoutes(srv, cfg)

	// Start server
	log.Printf("Starting PebbleDB server on port :8080")
	if err := srv.Start(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
