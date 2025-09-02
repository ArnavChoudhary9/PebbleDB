package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Create server instance
	server := NewServer()

	// Create database configuration
	config := Config{
		Path:            "pdb_data/pebbledb.db",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		WALMode:         true,
		ForeignKeys:     true,
	}

	// Create database connection
	db, err := NewDB(config)
	if err != nil {
		log.Fatal("Failed to create database:", err)
	}
	defer db.Close()

	// Add global middleware
	server.Use(authMiddleware)
	server.Use(LoggingMiddleware)
	server.Use(CORSMiddleware)
	server.Use(dbMiddleware(db))

	// Add root routes
	server.GET("/", homeHandler)

	// Create API route group
	apiGroup := server.Group("/api")
	apiGroup.POST("/db", handleDatabaseRequest)
	apiGroup.GET("/health", handleHealth)
	apiGroup.GET("/stats", handleStats)
	apiGroup.GET("/tables", handleTables)

	// Start server
	log.Fatal(server.Start(":8080"))
}

// Handler functions
func homeHandler(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintf(w, "Welcome to PebbleDB Server!")
	return nil
}
