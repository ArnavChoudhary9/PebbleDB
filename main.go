package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

const workingDirectoryContextKey contextKey = "working_directory"

func main() {
	// Create server instance
	server := NewServer()

	// Add global middleware
	server.Use(authMiddleware)
	server.Use(LoggingMiddleware)
	server.Use(CORSMiddleware)
	server.Use(WorkingDirectoryMiddleware("pdb_data"))
	server.Use(dbMiddleware)

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

func WorkingDirectoryMiddleware(basePath string) func(HTTPHandlerFunc) HTTPHandlerFunc {
	return func(next HTTPHandlerFunc) HTTPHandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx := context.WithValue(r.Context(), workingDirectoryContextKey, basePath)
			return next(w, r.WithContext(ctx))
		}
	}
}
