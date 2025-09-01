package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type contextKey string

const dbContextKey contextKey = "database"

// Middleware to inject database into context
func dbMiddleware(db *DB) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), dbContextKey, db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper to get database from context
func getDB(r *http.Request) *DB {
	return r.Context().Value(dbContextKey).(*DB)
}

// JSONRequest represents a generic JSON request
type JSONRequest struct {
	Action    string                 `json:"action"`
	Table     string                 `json:"table"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Where     string                 `json:"where,omitempty"`
	WhereArgs []interface{}          `json:"where_args,omitempty"`
	Columns   []string               `json:"columns,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
	Offset    int                    `json:"offset,omitempty"`
	OrderBy   string                 `json:"order_by,omitempty"`
	Schema    map[string]interface{} `json:"schema,omitempty"`
}

// JSONResponse represents a generic JSON response
type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Count   int64       `json:"count,omitempty"`
	ID      int64       `json:"id,omitempty"`
}

// handleDatabaseRequest handles all database operations via JSON
func handleDatabaseRequest(w http.ResponseWriter, r *http.Request) {
	var req JSONRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	db := getDB(r)
	switch req.Action {
	case "create_table":
		handleCreateTable(w, req, db)
	case "insert":
		handleInsert(w, req, db)
	case "select":
		handleSelect(w, req, db)
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
		sendError(w, "Unknown action: "+req.Action, http.StatusBadRequest)
	}
}

// handleCreateTable handles table creation from JSON schema
func handleCreateTable(w http.ResponseWriter, req JSONRequest, db *DB) {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return
	}

	var schema string
	if req.Schema != nil {
		schema = generateSchemaFromJSON(req.Schema)
	} else if req.Data != nil {
		// Auto-generate schema from sample data
		schema = inferSchemaFromData(req.Data)
	} else {
		sendError(w, "Schema or sample data is required", http.StatusBadRequest)
		return
	}

	err := db.CreateTable(req.Table, schema)
	if err != nil {
		sendError(w, "Failed to create table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]string{"message": "Table created successfully"})
}

// handleInsert handles record insertion
func handleInsert(w http.ResponseWriter, req JSONRequest, db *DB) {
	if req.Table == "" || req.Data == nil {
		sendError(w, "Table name and data are required", http.StatusBadRequest)
		return
	}

	id, err := db.Insert(req.Table, req.Data)
	if err != nil {
		sendError(w, "Failed to insert record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := JSONResponse{
		Success: true,
		ID:      id,
		Data:    map[string]interface{}{"inserted_id": id},
	}
	json.NewEncoder(w).Encode(response)
}

// handleSelect handles record selection
func handleSelect(w http.ResponseWriter, req JSONRequest, db *DB) {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return
	}

	// Build query
	query := buildSelectQuery(req)
	args := req.WhereArgs

	rows, err := db.Query(query, args...)
	if err != nil {
		sendError(w, "Failed to select records: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results, err := rowsToJSON(rows)
	if err != nil {
		sendError(w, "Failed to process results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := JSONResponse{
		Success: true,
		Data:    results,
		Count:   int64(len(results)),
	}
	json.NewEncoder(w).Encode(response)
}

// handleUpdate handles record updates
func handleUpdate(w http.ResponseWriter, req JSONRequest, db *DB) {
	if req.Table == "" || req.Data == nil {
		sendError(w, "Table name and data are required", http.StatusBadRequest)
		return
	}

	rowsAffected, err := db.Update(req.Table, req.Data, req.Where, req.WhereArgs...)
	if err != nil {
		sendError(w, "Failed to update records: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := JSONResponse{
		Success: true,
		Count:   rowsAffected,
		Data:    map[string]interface{}{"rows_affected": rowsAffected},
	}
	json.NewEncoder(w).Encode(response)
}

// handleDelete handles record deletion
func handleDelete(w http.ResponseWriter, req JSONRequest, db *DB) {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return
	}

	rowsAffected, err := db.Delete(req.Table, req.Where, req.WhereArgs...)
	if err != nil {
		sendError(w, "Failed to delete records: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := JSONResponse{
		Success: true,
		Count:   rowsAffected,
		Data:    map[string]interface{}{"rows_affected": rowsAffected},
	}
	json.NewEncoder(w).Encode(response)
}

// handleCount handles record counting
func handleCount(w http.ResponseWriter, req JSONRequest, db *DB) {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return
	}

	count, err := db.Count(req.Table, req.Where, req.WhereArgs...)
	if err != nil {
		sendError(w, "Failed to count records: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := JSONResponse{
		Success: true,
		Count:   count,
		Data:    map[string]interface{}{"count": count},
	}
	json.NewEncoder(w).Encode(response)
}

// handleDropTable handles table deletion
func handleDropTable(w http.ResponseWriter, req JSONRequest, db *DB) {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return
	}

	err := db.DropTable(req.Table)
	if err != nil {
		sendError(w, "Failed to drop table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]string{"message": "Table dropped successfully"})
}

// handleTableExists checks if table exists
func handleTableExists(w http.ResponseWriter, req JSONRequest, db *DB) {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return
	}

	exists, err := db.TableExists(req.Table)
	if err != nil {
		sendError(w, "Failed to check table existence: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]interface{}{"exists": exists})
}

// handleGetSchema gets table schema
func handleGetSchema(w http.ResponseWriter, req JSONRequest, db *DB) {
	if req.Table == "" {
		sendError(w, "Table name is required", http.StatusBadRequest)
		return
	}

	schema, err := db.GetTableSchema(req.Table)
	if err != nil {
		sendError(w, "Failed to get table schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]interface{}{"schema": schema})
}

// handleHealth returns server health status
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := getDB(r)
	err := db.Ping()
	if err != nil {
		sendError(w, "Database connection failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]string{"status": "healthy"})
}

// handleStats returns database statistics
func handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := getDB(r)
	stats, err := db.GetStats()
	if err != nil {
		sendError(w, "Failed to get stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, stats)
}

// handleTables returns list of tables
func handleTables(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := getDB(r)
	tables, err := db.ListTables()
	if err != nil {
		sendError(w, "Failed to list tables: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]interface{}{"tables": tables})
}

// Helper functions

// buildSelectQuery builds a SELECT query from JSON request
func buildSelectQuery(req JSONRequest) string {
	columns := "*"
	if len(req.Columns) > 0 {
		columns = strings.Join(req.Columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", columns, req.Table)

	if req.Where != "" {
		query += " WHERE " + req.Where
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
	response := JSONResponse{
		Success: true,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

// sendError sends an error response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := JSONResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(response)
}
