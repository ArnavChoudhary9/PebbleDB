package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type contextKey string

const dbContextKey contextKey = "database"

// Project represents a database project
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	Path        string `json:"path,omitempty"`
}

// Actions that don't require database middleware
var skipDBActions = map[string]bool{
	"create_project": true,
	"list_projects":  true,
	"delete_project": true,
	"get_project":    true,
}

// Middleware to inject database into context
func dbMiddleware(next HTTPHandlerFunc) HTTPHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Parse JSON to check if we should skip DB middleware
		var req JSONRequest
		if r.Method == "POST" && r.Body != nil && r.ContentLength > 0 {
			// Read the body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				return BadRequest("Failed to read request body")
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

		userID := r.Context().Value(userContextKey).(string)
		projectID := req.ProjectID
		if projectID == "" {
			projectID = r.URL.Query().Get("project")
		}

		if userID == "" || projectID == "" {
			return BadRequest("Missing user/project")
		}

		// Keep user/project format as requested
		basePath := r.Context().Value(workingDirectoryContextKey).(string)
		dbKey := fmt.Sprintf("%s/%s", userID, projectID)
		projectsBasePath := filepath.Join(basePath, "projects")

		log.Println("Database connection established for project:", projectID)
		db, err := getProjectDB(projectsBasePath, dbKey)
		if err != nil {
			return InternalServerError("Failed to load DB: " + err.Error())
		}

		ctx := context.WithValue(r.Context(), dbContextKey, db)
		return next(w, r.WithContext(ctx))
	}
}

// Helper to get database from context
func getDB(r *http.Request) *DB {
	return r.Context().Value(dbContextKey).(*DB)
}

// JSONJoin represents a join operation in JSON
type JSONJoin struct {
	Type      string `json:"type"`      // "INNER", "LEFT", "RIGHT", "FULL"
	Table     string `json:"table"`     // Table to join
	Condition string `json:"condition"` // Join condition (e.g., "users.id = profiles.user_id")
}

// JSONRequest represents a generic JSON request
type JSONRequest struct {
	Action    string                 `json:"action"`
	ProjectID string                 `json:"project_id,omitempty"` // Project identifier
	Table     string                 `json:"table"`
	Tables    []string               `json:"tables,omitempty"`    // For join action
	On        string                 `json:"on,omitempty"`        // For join condition
	JoinType  string                 `json:"join_type,omitempty"` // Optional join type
	Data      map[string]interface{} `json:"data,omitempty"`
	Where     string                 `json:"where,omitempty"`
	WhereArgs []interface{}          `json:"where_args,omitempty"`
	Columns   []string               `json:"columns,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
	Offset    int                    `json:"offset,omitempty"`
	OrderBy   string                 `json:"order_by,omitempty"`
	GroupBy   string                 `json:"group_by,omitempty"`
	Having    string                 `json:"having,omitempty"`
	Schema    map[string]interface{} `json:"schema,omitempty"`
	Joins     []JSONJoin             `json:"joins,omitempty"`
	// Project-specific fields
	ProjectName        string `json:"project_name,omitempty"`
	ProjectDescription string `json:"project_description,omitempty"`
}

// JSONResponse represents a generic JSON response
type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Count   int64       `json:"count,omitempty"`
	ID      int64       `json:"id,omitempty"`
	Query   string      `json:"query,omitempty"` // Optional: show generated query for debugging
}

// handleDatabaseRequest handles all database operations via JSON
func handleDatabaseRequest(w http.ResponseWriter, r *http.Request) error {
	var req JSONRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return err
	}

	// Handle project management actions first
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
	db := getDB(r)
	switch req.Action {
	case "create_table":
		handleCreateTable(w, req, db)
	case "insert":
		handleInsert(w, req, db)
	case "join":
		handleJoin(w, req, db)
	case "select":
		handleSelect(w, req, db)
	case "select_join":
		handleSelectWithJoin(w, req, db)
	case "count_join":
		handleCountWithJoin(w, req, db)
	case "query_builder":
		handleQueryBuilder(w, req, db)
	case "update":
		handleUpdate(w, req, db)
	case "delete":
		handleDelete(w, req, db)
	case "count":
		handleCount(w, req, db)
	case "drop_table":
		handleDropTable(w, req, db)
	case "table_exists":
		handleTableExists(w, req, db)
	case "get_schema":
		handleGetSchema(w, req, db)
	default:
		err := fmt.Errorf("unknown action: %s", req.Action)
		sendError(w, err.Error(), http.StatusBadRequest)
		return err
	}

	return nil
}

// handleCreateTable handles table creation from JSON schema
func handleCreateTable(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return fmt.Errorf("table name is required")
	}

	var schema string
	if req.Schema != nil {
		schema = generateSchemaFromJSON(req.Schema)
	} else if req.Data != nil {
		// Auto-generate schema from sample data
		schema = inferSchemaFromData(req.Data)
	} else {
		sendError(w, "Schema or sample data is required", http.StatusBadRequest)
		return fmt.Errorf("schema or sample data is required")
	}

	err := db.CreateTable(req.Table, schema)
	if err != nil {
		sendError(w, "Failed to create table: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to create table: %w", err)
	}

	sendSuccess(w, map[string]string{"message": "Table created successfully"})
	return nil
}

// handleInsert handles record insertion
func handleInsert(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" || req.Data == nil {
		sendError(w, "Table name and data are required", http.StatusBadRequest)
		return fmt.Errorf("table name and data are required")
	}

	id, err := db.Insert(req.Table, req.Data)
	if err != nil {
		sendError(w, "Failed to insert record: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to insert record: %w", err)
	}

	response := JSONResponse{
		Success: true,
		ID:      id,
		Data:    map[string]interface{}{"inserted_id": id},
	}
	json.NewEncoder(w).Encode(response)

	return nil
}

// handleJoin handles simple join queries using the Examples.md format
func handleJoin(w http.ResponseWriter, req JSONRequest, db *DB) error {
	// Validate required fields
	if len(req.Tables) < 2 {
		sendError(w, "At least two tables are required for join", http.StatusBadRequest)
		return fmt.Errorf("at least two tables are required for join")
	}

	if req.On == "" {
		sendError(w, "Join condition (on) is required", http.StatusBadRequest)
		return fmt.Errorf("join condition (on) is required")
	}

	// Build the join query
	baseTable := req.Tables[0]
	joinTable := req.Tables[1]

	// Default to INNER JOIN if not specified
	joinType := InnerJoin
	if req.JoinType != "" {
		var err error
		joinType, err = parseJoinType(req.JoinType)
		if err != nil {
			sendError(w, "Invalid join type: "+req.JoinType, http.StatusBadRequest)
			return fmt.Errorf("invalid join type: %w", err)
		}
	}

	joins := []Join{
		{
			Type:      joinType,
			Table:     joinTable,
			Condition: req.On,
		},
	}

	// Execute the join query
	rows, err := db.SelectWithJoin(baseTable, req.Columns, joins, req.Where, req.WhereArgs...)
	if err != nil {
		sendError(w, "Failed to execute join query: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to execute join query: %w", err)
	}
	defer rows.Close()

	results, err := rowsToJSON(rows)
	if err != nil {
		sendError(w, "Failed to process results: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to process results: %w", err)
	}

	response := JSONResponse{
		Success: true,
		Data:    results,
		Count:   int64(len(results)),
	}
	json.NewEncoder(w).Encode(response)

	return nil
}

// handleSelect handles record selection
func handleSelect(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return fmt.Errorf("table name is required")
	}

	// Build query
	query := buildSelectQuery(req)
	args := req.WhereArgs

	rows, err := db.Query(query, args...)
	if err != nil {
		sendError(w, "Failed to select records: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to select records: %w", err)
	}
	defer rows.Close()

	results, err := rowsToJSON(rows)
	if err != nil {
		sendError(w, "Failed to process results: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to process results: %w", err)
	}

	response := JSONResponse{
		Success: true,
		Data:    results,
		Count:   int64(len(results)),
		Query:   query, // Optional: for debugging
	}
	json.NewEncoder(w).Encode(response)
	return nil
}

// handleSelectWithJoin handles SELECT queries with joins
func handleSelectWithJoin(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return fmt.Errorf("table name is required")
	}

	if len(req.Joins) == 0 {
		sendError(w, "At least one join is required for select_join action", http.StatusBadRequest)
		return fmt.Errorf("at least one join is required for select_join action")
	}

	// Convert JSON joins to internal Join struct
	joins := make([]Join, 0, len(req.Joins))
	for _, jsonJoin := range req.Joins {
		joinType, err := parseJoinType(jsonJoin.Type)
		if err != nil {
			sendError(w, "Invalid join type: "+jsonJoin.Type, http.StatusBadRequest)
			return fmt.Errorf("invalid join type: %w", err)
		}

		joins = append(joins, Join{
			Type:      joinType,
			Table:     jsonJoin.Table,
			Condition: jsonJoin.Condition,
		})
	}

	rows, err := db.SelectWithJoin(req.Table, req.Columns, joins, req.Where, req.WhereArgs...)
	if err != nil {
		sendError(w, "Failed to execute join query: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to execute join query: %w", err)
	}
	defer rows.Close()

	results, err := rowsToJSON(rows)
	if err != nil {
		sendError(w, "Failed to process results: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to process results: %w", err)
	}

	response := JSONResponse{
		Success: true,
		Data:    results,
		Count:   int64(len(results)),
	}
	json.NewEncoder(w).Encode(response)
	return nil
}

// handleCountWithJoin handles COUNT queries with joins
func handleCountWithJoin(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return fmt.Errorf("table name is required")
	}

	if len(req.Joins) == 0 {
		sendError(w, "At least one join is required for count_join action", http.StatusBadRequest)
		return fmt.Errorf("at least one join is required for count_join action")
	}

	// Convert JSON joins to internal Join struct
	joins := make([]Join, 0, len(req.Joins))
	for _, jsonJoin := range req.Joins {
		joinType, err := parseJoinType(jsonJoin.Type)
		if err != nil {
			sendError(w, "Invalid join type: "+jsonJoin.Type, http.StatusBadRequest)
			return fmt.Errorf("invalid join type: %w", err)
		}

		joins = append(joins, Join{
			Type:      joinType,
			Table:     jsonJoin.Table,
			Condition: jsonJoin.Condition,
		})
	}

	count, err := db.CountWithJoin(req.Table, joins, req.Where, req.WhereArgs...)
	if err != nil {
		sendError(w, "Failed to execute count join query: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to execute count join query: %w", err)
	}

	response := JSONResponse{
		Success: true,
		Count:   count,
		Data:    map[string]interface{}{"count": count},
	}
	json.NewEncoder(w).Encode(response)
	return nil
}

// handleQueryBuilder handles complex queries using QueryBuilder
func handleQueryBuilder(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return fmt.Errorf("table name is required")
	}

	// Build query using QueryBuilder
	qb := db.QueryBuilder(req.Table)

	// Add columns if specified
	if len(req.Columns) > 0 {
		qb.Select(req.Columns...)
	}

	// Add joins
	for _, jsonJoin := range req.Joins {
		joinType, err := parseJoinType(jsonJoin.Type)
		if err != nil {
			sendError(w, "Invalid join type: "+jsonJoin.Type, http.StatusBadRequest)
			return fmt.Errorf("invalid join type: %w", err)
		}

		qb.Join(joinType, jsonJoin.Table, jsonJoin.Condition)
	}

	// Add WHERE clause
	if req.Where != "" {
		qb.Where(req.Where, req.WhereArgs...)
	}

	// Add GROUP BY
	if req.GroupBy != "" {
		qb.GroupBy(req.GroupBy)
	}

	// Add HAVING
	if req.Having != "" {
		qb.Having(req.Having)
	}

	// Add ORDER BY
	if req.OrderBy != "" {
		qb.OrderBy(req.OrderBy)
	}

	// Add LIMIT and OFFSET
	if req.Limit > 0 {
		qb.Limit(req.Limit)
	}
	if req.Offset > 0 {
		qb.Offset(req.Offset)
	}

	// Execute query
	rows, err := qb.Execute(db)
	if err != nil {
		sendError(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	results, err := rowsToJSON(rows)
	if err != nil {
		sendError(w, "Failed to process results: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to process results: %w", err)
	}

	// Get the generated query for debugging
	query, _ := qb.Build()

	response := JSONResponse{
		Success: true,
		Data:    results,
		Count:   int64(len(results)),
		Query:   query, // Show generated query
	}
	json.NewEncoder(w).Encode(response)

	return nil
}

// handleUpdate handles record updates
func handleUpdate(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" || req.Data == nil {
		sendError(w, "Table name and data are required", http.StatusBadRequest)
		return fmt.Errorf("table name and data are required")
	}

	rowsAffected, err := db.Update(req.Table, req.Data, req.Where, req.WhereArgs...)
	if err != nil {
		sendError(w, "Failed to update records: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to update records: %w", err)
	}

	response := JSONResponse{
		Success: true,
		Count:   rowsAffected,
		Data:    map[string]interface{}{"rows_affected": rowsAffected},
	}
	json.NewEncoder(w).Encode(response)
	return nil
}

// handleDelete handles record deletion
func handleDelete(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return fmt.Errorf("table name is required")
	}

	rowsAffected, err := db.Delete(req.Table, req.Where, req.WhereArgs...)
	if err != nil {
		sendError(w, "Failed to delete records: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to delete records: %w", err)
	}

	response := JSONResponse{
		Success: true,
		Count:   rowsAffected,
		Data:    map[string]interface{}{"rows_affected": rowsAffected},
	}
	json.NewEncoder(w).Encode(response)
	return nil
}

// handleCount handles record counting
func handleCount(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return fmt.Errorf("table name is required")
	}

	count, err := db.Count(req.Table, req.Where, req.WhereArgs...)
	if err != nil {
		sendError(w, "Failed to count records: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to count records: %w", err)
	}

	response := JSONResponse{
		Success: true,
		Count:   count,
		Data:    map[string]interface{}{"count": count},
	}
	json.NewEncoder(w).Encode(response)
	return nil
}

// handleDropTable handles table deletion
func handleDropTable(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return fmt.Errorf("table name is required")
	}

	err := db.DropTable(req.Table)
	if err != nil {
		sendError(w, "Failed to drop table: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to drop table: %w", err)
	}

	sendSuccess(w, map[string]string{"message": "Table dropped successfully"})
	return nil
}

// handleTableExists checks if table exists
func handleTableExists(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return fmt.Errorf("table name is required")
	}

	exists, err := db.TableExists(req.Table)
	if err != nil {
		sendError(w, "Failed to check table existence: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	sendSuccess(w, map[string]interface{}{"exists": exists})
	return nil
}

// handleGetSchema gets table schema
func handleGetSchema(w http.ResponseWriter, req JSONRequest, db *DB) error {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return fmt.Errorf("table name is required")
	}

	schema, err := db.GetTableSchema(req.Table)
	if err != nil {
		sendError(w, "Failed to get table schema: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to get table schema: %w", err)
	}

	sendSuccess(w, map[string]interface{}{"schema": schema})
	return nil
}

// handleHealth returns server health status
func handleHealth(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	db := getDB(r)
	err := db.Ping()
	if err != nil {
		sendError(w, "Database connection failed: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("database connection failed: %w", err)
	}

	sendSuccess(w, map[string]string{"status": "healthy"})
	return nil
}

// handleStats returns database statistics
func handleStats(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	db := getDB(r)
	stats, err := db.GetStats()
	if err != nil {
		sendError(w, "Failed to get stats: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to get stats: %w", err)
	}

	sendSuccess(w, stats)
	return nil
}

// handleTables returns list of tables
func handleTables(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	db := getDB(r)
	tables, err := db.ListTables()
	if err != nil {
		sendError(w, "Failed to list tables: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to list tables: %w", err)
	}

	sendSuccess(w, map[string]interface{}{"tables": tables})
	return nil
}

// Helper functions

// parseJoinType converts string join type to JoinType
func parseJoinType(joinTypeStr string) (JoinType, error) {
	switch strings.ToUpper(joinTypeStr) {
	case "INNER":
		return InnerJoin, nil
	case "LEFT":
		return LeftJoin, nil
	case "RIGHT":
		return RightJoin, nil
	case "FULL":
		return FullJoin, nil
	default:
		return "", fmt.Errorf("invalid join type: %s", joinTypeStr)
	}
}

// buildSelectQuery builds a SELECT query from JSON request
func buildSelectQuery(req JSONRequest) string {
	columns := "*"
	if len(req.Columns) > 0 {
		columns = strings.Join(req.Columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", columns, req.Table)

	// Add joins if any
	for _, join := range req.Joins {
		joinType := strings.ToUpper(join.Type) + " JOIN"
		query += fmt.Sprintf(" %s %s ON %s", joinType, join.Table, join.Condition)
	}

	if req.Where != "" {
		query += " WHERE " + req.Where
	}

	if req.GroupBy != "" {
		query += " GROUP BY " + req.GroupBy
	}

	if req.Having != "" {
		query += " HAVING " + req.Having
	}

	if req.OrderBy != "" {
		query += " ORDER BY " + req.OrderBy
	}

	if req.Limit > 0 {
		query += " LIMIT " + strconv.Itoa(req.Limit)
		if req.Offset > 0 {
			query += " OFFSET " + strconv.Itoa(req.Offset)
		}
	}

	return query
}

// rowsToJSON converts SQL rows to JSON-compatible format
func rowsToJSON(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Convert byte arrays to strings
			if b, ok := val.([]byte); ok {
				val = string(b)
			}

			row[col] = val
		}

		results = append(results, row)
	}

	return results, rows.Err()
}

// generateSchemaFromJSON creates SQL schema from JSON schema definition
func generateSchemaFromJSON(schema map[string]interface{}) string {
	var columns []string

	for name, def := range schema {
		column := name

		if defMap, ok := def.(map[string]interface{}); ok {
			if sqlType, ok := defMap["type"].(string); ok {
				column += " " + strings.ToUpper(sqlType)
			}

			if constraints, ok := defMap["constraints"].([]interface{}); ok {
				for _, constraint := range constraints {
					if constraintStr, ok := constraint.(string); ok {
						column += " " + strings.ToUpper(constraintStr)
					}
				}
			}
		}

		columns = append(columns, column)
	}

	return strings.Join(columns, ",\n    ")
}

// inferSchemaFromData infers SQL schema from sample data
func inferSchemaFromData(data map[string]interface{}) string {
	var columns []string

	for name, value := range data {
		column := name + " "

		switch v := value.(type) {
		case int, int32, int64:
			column += "INTEGER"
		case float32, float64:
			column += "REAL"
		case bool:
			column += "INTEGER"
		case string:
			column += "TEXT"
		case nil:
			column += "TEXT"
		default:
			_ = v
			column += "TEXT"
		}

		columns = append(columns, column)
	}

	return strings.Join(columns, ",\n    ")
}

// sendSuccess sends a success response
func sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := JSONResponse{
		Success: true,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

// sendError sends an error response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := JSONResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(response)
}

// Project Management Handlers

// handleCreateProject creates a new project
func handleCreateProject(w http.ResponseWriter, req JSONRequest, r *http.Request) error {
	if req.ProjectID == "" || req.ProjectName == "" {
		sendError(w, "Project ID and name are required", http.StatusBadRequest)
		return fmt.Errorf("project ID and name are required")
	}

	userID := r.Context().Value(userContextKey).(string)
	if userID == "" {
		sendError(w, "User ID is required", http.StatusBadRequest)
		return fmt.Errorf("user ID is required")
	}

	// Create project directory
	basePath := r.Context().Value(workingDirectoryContextKey).(string)
	projectsDir := filepath.Join(basePath, "projects", userID)
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		sendError(w, "Failed to create project directory: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create project database - keep user/project format as requested
	dbKey := fmt.Sprintf("%s/%s", userID, req.ProjectID)
	projectPath := filepath.Join(projectsDir, req.ProjectID+".db")

	// Check if project already exists
	if _, err := os.Stat(projectPath); err == nil {
		sendError(w, "Project already exists", http.StatusConflict)
		return fmt.Errorf("project already exists")
	}

	// Initialize the database - ensure projects directory exists
	projectsBasePath := filepath.Join(basePath, "projects")
	if err := os.MkdirAll(projectsBasePath, 0755); err != nil {
		sendError(w, "Failed to create projects base directory: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to create projects base directory: %w", err)
	}

	_, err := getProjectDB(projectsBasePath, dbKey)
	if err != nil {
		sendError(w, "Failed to create project database: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to create project database: %w", err)
	}

	// Create project metadata
	project := Project{
		ID:          req.ProjectID,
		Name:        req.ProjectName,
		Description: req.ProjectDescription,
		CreatedAt:   fmt.Sprintf("%d", os.Getpid()), // Simplified timestamp
		Path:        projectPath,
	}

	// Save project metadata (you might want to store this in a separate metadata DB)
	metadataPath := filepath.Join(projectsDir, req.ProjectID+".json")
	metadataFile, err := os.Create(metadataPath)
	if err != nil {
		sendError(w, "Failed to create project metadata: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to create project metadata: %w", err)
	}
	defer metadataFile.Close()

	if err := json.NewEncoder(metadataFile).Encode(project); err != nil {
		sendError(w, "Failed to save project metadata: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to save project metadata: %w", err)
	}

	sendSuccess(w, map[string]interface{}{
		"message": "Project created successfully",
		"project": project,
	})
	return nil
}

// handleListProjects lists all projects for a user
func handleListProjects(w http.ResponseWriter, req JSONRequest, r *http.Request) error {
	userID := r.Context().Value(userContextKey).(string)
	if userID == "" {
		sendError(w, "User ID is required", http.StatusBadRequest)
		return fmt.Errorf("user ID is required")
	}

	basePath := r.Context().Value(workingDirectoryContextKey).(string)
	projectsDir := filepath.Join(basePath, "projects", userID)
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		sendSuccess(w, map[string]interface{}{"projects": []Project{}})
		return nil
	}

	files, err := os.ReadDir(projectsDir)
	if err != nil {
		sendError(w, "Failed to read projects directory: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to read projects directory: %w", err)
	}

	var projects []Project
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			metadataPath := filepath.Join(projectsDir, file.Name())
			metadataFile, err := os.Open(metadataPath)
			if err != nil {
				continue
			}

			var project Project
			if err := json.NewDecoder(metadataFile).Decode(&project); err != nil {
				metadataFile.Close()
				continue
			}
			metadataFile.Close()

			projects = append(projects, project)
		}
	}

	sendSuccess(w, map[string]interface{}{"projects": projects})
	return nil
}

// handleDeleteProject deletes a project
func handleDeleteProject(w http.ResponseWriter, req JSONRequest, r *http.Request) error {
	if req.ProjectID == "" {
		sendError(w, "Project ID is required", http.StatusBadRequest)
		return fmt.Errorf("project ID is required")
	}

	userID := r.Context().Value(userContextKey).(string)
	if userID == "" {
		sendError(w, "User ID is required", http.StatusBadRequest)
		return fmt.Errorf("user ID is required")
	}

	basePath := r.Context().Value(workingDirectoryContextKey).(string)
	projectsDir := filepath.Join(basePath, "projects", userID)
	dbPath := filepath.Join(projectsDir, req.ProjectID+".db")
	metadataPath := filepath.Join(projectsDir, req.ProjectID+".json")

	// Remove database file
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		sendError(w, "Failed to delete project database: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to delete project database: %w", err)
	}

	// Remove metadata file
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		sendError(w, "Failed to delete project metadata: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to delete project metadata: %w", err)
	}

	sendSuccess(w, map[string]interface{}{"message": "Project deleted successfully"})
	return nil
}

// handleGetProject gets project information
func handleGetProject(w http.ResponseWriter, req JSONRequest, r *http.Request) error {
	if req.ProjectID == "" {
		sendError(w, "Project ID is required", http.StatusBadRequest)
		return fmt.Errorf("project ID is required")
	}

	userID := r.Context().Value(userContextKey).(string)
	if userID == "" {
		sendError(w, "User ID is required", http.StatusBadRequest)
		return fmt.Errorf("user ID is required")
	}

	basePath := r.Context().Value(workingDirectoryContextKey).(string)
	projectsDir := filepath.Join(basePath, "projects", userID)
	metadataPath := filepath.Join(projectsDir, req.ProjectID+".json")

	metadataFile, err := os.Open(metadataPath)
	if err != nil {
		sendError(w, "Project not found", http.StatusNotFound)
		return fmt.Errorf("project not found")
	}
	defer metadataFile.Close()

	var project Project
	if err := json.NewDecoder(metadataFile).Decode(&project); err != nil {
		sendError(w, "Failed to read project metadata: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to read project metadata: %w", err)
	}

	sendSuccess(w, map[string]interface{}{"project": project})
	return nil
}

// handleGetTables gets all tables for a project (requires DB connection)
func handleGetTables(w http.ResponseWriter, req JSONRequest, r *http.Request) error {
	if req.ProjectID == "" {
		sendError(w, "Project ID is required", http.StatusBadRequest)
		return fmt.Errorf("project ID is required")
	}

	// This requires database connection, so get it from context
	db := getDB(r)
	tables, err := db.ListTables()
	if err != nil {
		sendError(w, "Failed to list tables: "+err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to list tables: %w", err)
	}

	sendSuccess(w, map[string]interface{}{
		"project_id": req.ProjectID,
		"tables":     tables,
	})
	return nil
}
