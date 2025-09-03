package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ArnavChoudhary9/PebbleDB/internal/database"
	"github.com/ArnavChoudhary9/PebbleDB/internal/server"
	"github.com/ArnavChoudhary9/PebbleDB/pkg/types"
)

// DatabaseHandler handles all database operations via JSON
func DatabaseHandler(w http.ResponseWriter, r *http.Request) error {
	var req types.JSONRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return server.BadRequest("Invalid JSON request: " + err.Error())
	}

	// Handle project management actions first (these don't require database connection)
	switch req.Action {
	case "create_project":
		return handleCreateProject(w, req, r)
	case "list_projects":
		return handleListProjects(w, req, r)
	case "delete_project":
		return handleDeleteProject(w, req, r)
	case "get_project":
		return handleGetProject(w, req, r)
	case "get_tables":
		return handleGetTables(w, req, r)
	}

	// For database operations, get the database connection
	db := database.GetDBFromContext(r)
	if db == nil {
		return server.InternalServerError("Database connection not available")
	}

	switch req.Action {
	case "create_table":
		return handleCreateTable(w, req, db)
	case "insert":
		return handleInsert(w, req, db)
	case "join":
		return handleJoin(w, req, db)
	case "select":
		return handleSelect(w, req, db)
	case "select_join":
		return handleSelectWithJoin(w, req, db)
	case "count_join":
		return handleCountWithJoin(w, req, db)
	case "query_builder":
		return handleQueryBuilder(w, req, db)
	case "update":
		return handleUpdate(w, req, db)
	case "delete":
		return handleDelete(w, req, db)
	case "count":
		return handleCount(w, req, db)
	case "drop_table":
		return handleDropTable(w, req, db)
	case "table_exists":
		return handleTableExists(w, req, db)
	case "get_schema":
		return handleGetSchema(w, req, db)
	default:
		return server.BadRequest(fmt.Sprintf("Unknown action: %s", req.Action))
	}
}

// Helper function to send success response
func sendSuccess(w http.ResponseWriter, data interface{}) error {
	response := types.JSONResponse{
		Success: true,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

// Helper function to send error response
func sendError(w http.ResponseWriter, message string, statusCode int) error {
	response := types.JSONResponse{
		Success: false,
		Error:   message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(response)
}
