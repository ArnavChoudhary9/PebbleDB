package database

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ArnavChoudhary9/PebbleDB/internal/server"
	"github.com/ArnavChoudhary9/PebbleDB/pkg/types"
)

// Actions that don't require database middleware
var skipDBActions = map[string]bool{
	"create_project": true,
	"list_projects":  true,
	"delete_project": true,
	"get_project":    true,
}

// Middleware creates a middleware that injects database connections into the request context
func Middleware() func(server.HTTPHandlerFunc) server.HTTPHandlerFunc {
	return func(next server.HTTPHandlerFunc) server.HTTPHandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			// Parse JSON to check if we should skip DB middleware
			var req types.JSONRequest
			if r.Method == "POST" && r.Body != nil && r.ContentLength > 0 {
				// Read the body
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					return server.BadRequest("Failed to read request body")
				}

				// Try to parse JSON, but don't fail if it's invalid
				json.Unmarshal(bodyBytes, &req)

				// Reset body for downstream handlers
				r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
			}

			// Skip database middleware for certain actions
			if skipDBActions[req.Action] {
				return next(w, r)
			}

			userID, ok := r.Context().Value(types.UserContextKey).(string)
			if !ok || userID == "" {
				return server.BadRequest("Missing user context")
			}

			projectID := req.ProjectID
			if projectID == "" {
				projectID = r.URL.Query().Get("project")
			}

			if projectID == "" {
				return server.BadRequest("Missing project ID")
			}

			// Get working directory from context
			basePath, ok := r.Context().Value(types.WorkingDirectoryContextKey).(string)
			if !ok || basePath == "" {
				return server.InternalServerError("Missing working directory context")
			}

			// Keep user/project format as requested
			dbKey := fmt.Sprintf("%s/%s", userID, projectID)
			projectsBasePath := filepath.Join(basePath, "projects")

			log.Printf("Establishing database connection for project: %s (user: %s)", projectID, userID)
			db, err := GetProjectDB(projectsBasePath, dbKey)
			if err != nil {
				return server.InternalServerError("Failed to load database: " + err.Error())
			}

			ctx := context.WithValue(r.Context(), types.DatabaseContextKey, db)
			return next(w, r.WithContext(ctx))
		}
	}
}

// GetDBFromContext retrieves the database connection from the request context
func GetDBFromContext(r *http.Request) *DB {
	db, ok := r.Context().Value(types.DatabaseContextKey).(*DB)
	if !ok {
		return nil
	}
	return db
}
