package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Create server instance
	server := NewServer()

	// Add global middleware
	server.Use(authMiddleware)
	server.Use(LoggingMiddleware)
	server.Use(CORSMiddleware)
	server.Use(dbMiddleware("pdb_data"))

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
