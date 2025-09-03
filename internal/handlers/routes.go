package handlers

import (
	"fmt"
	"net/http"

	"github.com/ArnavChoudhary9/PebbleDB/internal/auth"
	"github.com/ArnavChoudhary9/PebbleDB/internal/config"
	"github.com/ArnavChoudhary9/PebbleDB/internal/database"
	"github.com/ArnavChoudhary9/PebbleDB/internal/server"
)

// SetupRoutes configures all routes and middleware for the server
func SetupRoutes(srv *server.Server, cfg *config.Config) {
	// Add global middleware
	srv.Use(server.LoggingMiddleware)
	srv.Use(server.CORSMiddleware)
	srv.Use(server.WorkingDirectoryMiddleware("pdb_data"))
	srv.Use(auth.Middleware(cfg))
	srv.Use(database.Middleware())

	// Add root routes
	srv.GET("/", homeHandler)

	// Create API route group
	apiGroup := srv.Group("/api")
	apiGroup.POST("/db", DatabaseHandler)
	apiGroup.GET("/health", HealthHandler)
	apiGroup.GET("/stats", statsHandler)
	apiGroup.GET("/tables", tablesHandler)
}

// homeHandler handles the root endpoint
func homeHandler(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintf(w, "Welcome to PebbleDB Server!")
	return nil
}

// statsHandler handles database statistics requests
func statsHandler(w http.ResponseWriter, r *http.Request) error {
	// TODO: Implement database statistics
	return sendError(w, "Statistics endpoint not yet implemented", http.StatusNotImplemented)
}

// tablesHandler handles table listing requests
func tablesHandler(w http.ResponseWriter, r *http.Request) error {
	// TODO: Implement table listing
	return sendError(w, "Tables endpoint not yet implemented", http.StatusNotImplemented)
}
