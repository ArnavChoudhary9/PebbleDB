package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ArnavChoudhary9/PebbleDB/internal/database"
	"github.com/ArnavChoudhary9/PebbleDB/internal/server"
	"github.com/ArnavChoudhary9/PebbleDB/pkg/types"
)

// handleCreateProject creates a new project
func handleCreateProject(w http.ResponseWriter, req types.JSONRequest, r *http.Request) error {
	if req.ProjectName == "" {
		return server.BadRequest("Project name is required")
	}

	// Get user ID from context
	userID, ok := r.Context().Value(types.UserContextKey).(string)
	if !ok || userID == "" {
		return server.BadRequest("User context required")
	}

	// Get working directory from context
	basePath, ok := r.Context().Value(types.WorkingDirectoryContextKey).(string)
	if !ok || basePath == "" {
		return server.InternalServerError("Working directory context required")
	}

	projectsBasePath := filepath.Join(basePath, "projects")
	userProjectsPath := filepath.Join(projectsBasePath, userID)

	// Create user projects directory if it doesn't exist
	if err := os.MkdirAll(userProjectsPath, 0755); err != nil {
		return server.InternalServerError("Failed to create user projects directory: " + err.Error())
	}

	// Generate a unique project ID
	projectID := generateProjectID()
	projectPath := filepath.Join(userProjectsPath, projectID)

	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return server.InternalServerError("Failed to create project directory: " + err.Error())
	}

	// Create project metadata
	project := types.Project{
		ID:          projectID,
		Name:        req.ProjectName,
		Description: req.ProjectDescription,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		Path:        projectPath,
	}

	// Save project metadata to JSON file
	metadataPath := filepath.Join(projectPath, fmt.Sprintf("%s.json", req.ProjectName))
	metadataFile, err := os.Create(metadataPath)
	if err != nil {
		return server.InternalServerError("Failed to create project metadata file: " + err.Error())
	}
	defer metadataFile.Close()

	if err := json.NewEncoder(metadataFile).Encode(project); err != nil {
		return server.InternalServerError("Failed to write project metadata: " + err.Error())
	}

	return sendSuccess(w, project)
}

// handleListProjects lists all projects for a user
func handleListProjects(w http.ResponseWriter, req types.JSONRequest, r *http.Request) error {
	// Get user ID from context
	userID, ok := r.Context().Value(types.UserContextKey).(string)
	if !ok || userID == "" {
		return server.BadRequest("User context required")
	}

	// Get working directory from context
	basePath, ok := r.Context().Value(types.WorkingDirectoryContextKey).(string)
	if !ok || basePath == "" {
		return server.InternalServerError("Working directory context required")
	}

	projectsBasePath := filepath.Join(basePath, "projects")
	userProjectsPath := filepath.Join(projectsBasePath, userID)

	// Check if user projects directory exists
	if _, err := os.Stat(userProjectsPath); os.IsNotExist(err) {
		return sendSuccess(w, []types.Project{})
	}

	// Read project directories
	entries, err := os.ReadDir(userProjectsPath)
	if err != nil {
		return server.InternalServerError("Failed to read projects directory: " + err.Error())
	}

	projects := []types.Project{}
	for _, entry := range entries {
		if entry.IsDir() {
			projectPath := filepath.Join(userProjectsPath, entry.Name())

			// Look for JSON metadata file
			jsonFiles, err := filepath.Glob(filepath.Join(projectPath, "*.json"))
			if err != nil || len(jsonFiles) == 0 {
				continue
			}

			// Read the first JSON file found
			metadataFile, err := os.Open(jsonFiles[0])
			if err != nil {
				continue
			}

			var project types.Project
			if err := json.NewDecoder(metadataFile).Decode(&project); err != nil {
				metadataFile.Close()
				continue
			}
			metadataFile.Close()

			projects = append(projects, project)
		}
	}

	return sendSuccess(w, projects)
}

// handleDeleteProject deletes a project
func handleDeleteProject(w http.ResponseWriter, req types.JSONRequest, r *http.Request) error {
	if req.ProjectID == "" {
		return server.BadRequest("Project ID is required")
	}

	// Get user ID from context
	userID, ok := r.Context().Value(types.UserContextKey).(string)
	if !ok || userID == "" {
		return server.BadRequest("User context required")
	}

	// Get working directory from context
	basePath, ok := r.Context().Value(types.WorkingDirectoryContextKey).(string)
	if !ok || basePath == "" {
		return server.InternalServerError("Working directory context required")
	}

	projectsBasePath := filepath.Join(basePath, "projects")
	userProjectsPath := filepath.Join(projectsBasePath, userID)
	projectPath := filepath.Join(userProjectsPath, req.ProjectID)

	// Check if project exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return server.NotFound("Project not found")
	}

	// Remove project directory
	if err := os.RemoveAll(projectPath); err != nil {
		return server.InternalServerError("Failed to delete project: " + err.Error())
	}

	return sendSuccess(w, map[string]string{"message": "Project deleted successfully"})
}

// handleGetProject gets project information
func handleGetProject(w http.ResponseWriter, req types.JSONRequest, r *http.Request) error {
	if req.ProjectID == "" {
		return server.BadRequest("Project ID is required")
	}

	// Get user ID from context
	userID, ok := r.Context().Value(types.UserContextKey).(string)
	if !ok || userID == "" {
		return server.BadRequest("User context required")
	}

	// Get working directory from context
	basePath, ok := r.Context().Value(types.WorkingDirectoryContextKey).(string)
	if !ok || basePath == "" {
		return server.InternalServerError("Working directory context required")
	}

	projectsBasePath := filepath.Join(basePath, "projects")
	userProjectsPath := filepath.Join(projectsBasePath, userID)
	projectPath := filepath.Join(userProjectsPath, req.ProjectID)

	// Check if project exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return server.NotFound("Project not found")
	}

	// Look for JSON metadata file
	jsonFiles, err := filepath.Glob(filepath.Join(projectPath, "*.json"))
	if err != nil || len(jsonFiles) == 0 {
		return server.NotFound("Project metadata not found")
	}

	// Read the first JSON file found
	metadataFile, err := os.Open(jsonFiles[0])
	if err != nil {
		return server.InternalServerError("Failed to open project metadata: " + err.Error())
	}
	defer metadataFile.Close()

	var project types.Project
	if err := json.NewDecoder(metadataFile).Decode(&project); err != nil {
		return server.InternalServerError("Failed to parse project metadata: " + err.Error())
	}

	return sendSuccess(w, project)
}

// handleGetTables gets all tables for a project (requires DB connection)
func handleGetTables(w http.ResponseWriter, req types.JSONRequest, r *http.Request) error {
	db := database.GetDBFromContext(r)
	if db == nil {
		return server.InternalServerError("Database connection not available")
	}

	tables, err := db.ListTables()
	if err != nil {
		return server.InternalServerError("Failed to list tables: " + err.Error())
	}

	return sendSuccess(w, map[string]interface{}{
		"tables": tables,
		"count":  len(tables),
	})
}

// generateProjectID generates a unique project ID
func generateProjectID() string {
	// For now, use timestamp + random string
	// In production, you might want to use UUID
	return fmt.Sprintf("proj_%d", time.Now().Unix())
}
