package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/ArnavChoudhary9/PebbleDB/internal/database"
	"github.com/ArnavChoudhary9/PebbleDB/internal/server"
	"github.com/ArnavChoudhary9/PebbleDB/pkg/types"
)

// Basic CRUD operations for database handlers

// handleInsert handles record insertion
func handleInsert(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" || req.Data == nil {
		return server.BadRequest("Table name and data are required")
	}

	id, err := db.Insert(req.Table, req.Data)
	if err != nil {
		return server.InternalServerError("Failed to insert record: " + err.Error())
	}

	response := types.JSONResponse{
		Success: true,
		ID:      id,
		Data:    map[string]interface{}{"inserted_id": id},
	}

	return sendJSONResponse(w, response)
}

// handleSelect handles record selection
func handleSelect(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" {
		return server.BadRequest("Table name is required")
	}

	// Build query using the database Select method
	rows, err := db.Select(req.Table, req.Columns, req.Where, req.WhereArgs...)
	if err != nil {
		return server.InternalServerError("Failed to execute query: " + err.Error())
	}
	defer rows.Close()

	data, err := rowsToMap(rows)
	if err != nil {
		return server.InternalServerError("Failed to process results: " + err.Error())
	}

	// Apply ORDER BY, LIMIT, OFFSET at application level if needed
	// Note: For better performance, these should be handled in the database layer
	if req.OrderBy != "" || req.Limit > 0 || req.Offset > 0 {
		// Fall back to building a custom query for these advanced features
		return handleSelectWithCustomQuery(w, req, db)
	}

	response := types.JSONResponse{
		Success: true,
		Data:    data,
		Count:   int64(len(data)),
	}

	return sendJSONResponse(w, response)
}

// handleSelectWithCustomQuery handles SELECT with ORDER BY, LIMIT, OFFSET
func handleSelectWithCustomQuery(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	// Build columns
	columns := "*"
	if len(req.Columns) > 0 {
		columns = strings.Join(req.Columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", columns, req.Table)
	args := []interface{}{}

	// Add WHERE clause
	if req.Where != "" {
		query += " WHERE " + req.Where
		args = append(args, req.WhereArgs...)
	}

	// Add ORDER BY
	if req.OrderBy != "" {
		query += " ORDER BY " + req.OrderBy
	}

	// Add LIMIT and OFFSET
	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", req.Limit)
	}
	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", req.Offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return server.InternalServerError("Failed to execute query: " + err.Error())
	}
	defer rows.Close()

	data, err := rowsToMap(rows)
	if err != nil {
		return server.InternalServerError("Failed to process results: " + err.Error())
	}

	response := types.JSONResponse{
		Success: true,
		Data:    data,
		Count:   int64(len(data)),
		Query:   query,
	}

	return sendJSONResponse(w, response)
}

// handleUpdate handles record updates
func handleUpdate(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" || req.Data == nil {
		return server.BadRequest("Table name and data are required")
	}

	rowsAffected, err := db.Update(req.Table, req.Data, req.Where, req.WhereArgs...)
	if err != nil {
		return server.InternalServerError("Failed to update records: " + err.Error())
	}

	response := types.JSONResponse{
		Success: true,
		Count:   rowsAffected,
		Data:    map[string]interface{}{"rows_affected": rowsAffected},
	}

	return sendJSONResponse(w, response)
}

// handleDelete handles record deletion
func handleDelete(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" {
		return server.BadRequest("Table name is required")
	}

	rowsAffected, err := db.Delete(req.Table, req.Where, req.WhereArgs...)
	if err != nil {
		return server.InternalServerError("Failed to delete records: " + err.Error())
	}

	response := types.JSONResponse{
		Success: true,
		Count:   rowsAffected,
		Data:    map[string]interface{}{"rows_affected": rowsAffected},
	}

	return sendJSONResponse(w, response)
}

// handleCount handles record counting
func handleCount(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" {
		return server.BadRequest("Table name is required")
	}

	count, err := db.Count(req.Table, req.Where, req.WhereArgs...)
	if err != nil {
		return server.InternalServerError("Failed to count records: " + err.Error())
	}

	response := types.JSONResponse{
		Success: true,
		Count:   count,
		Data:    map[string]interface{}{"count": count},
	}

	return sendJSONResponse(w, response)
}

// Helper function to convert SQL rows to map slice
func rowsToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Create map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Convert []byte to string if necessary
			if b, ok := val.([]byte); ok {
				val = string(b)
			}

			rowMap[col] = val
		}

		results = append(results, rowMap)
	}

	return results, rows.Err()
}

// Helper function to send JSON response
func sendJSONResponse(w http.ResponseWriter, response types.JSONResponse) error {
	w.Header().Set("Content-Type", "application/json")
	return sendSuccess(w, response.Data)
}
