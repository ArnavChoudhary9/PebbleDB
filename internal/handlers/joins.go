package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ArnavChoudhary9/PebbleDB/internal/database"
	"github.com/ArnavChoudhary9/PebbleDB/internal/server"
	"github.com/ArnavChoudhary9/PebbleDB/pkg/types"
)

// handleJoin handles simple join queries
func handleJoin(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	// Validate required fields
	if len(req.Tables) < 2 {
		return server.BadRequest("At least two tables are required for join")
	}

	if req.On == "" {
		return server.BadRequest("Join condition (on) is required")
	}

	// Build the join query
	baseTable := req.Tables[0]
	joinTable := req.Tables[1]

	// Default to INNER JOIN if not specified
	joinType := "INNER JOIN"
	if req.JoinType != "" {
		joinType = strings.ToUpper(req.JoinType) + " JOIN"
	}

	// Build columns to select
	columns := "*"
	if len(req.Columns) > 0 {
		columns = strings.Join(req.Columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s %s %s ON %s",
		columns, baseTable, joinType, joinTable, req.On)

	// Add WHERE clause
	args := []interface{}{}
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

	// Execute the join query
	rows, err := db.Query(query, args...)
	if err != nil {
		return server.InternalServerError("Failed to execute join query: " + err.Error())
	}
	defer rows.Close()

	data, err := rowsToMap(rows)
	if err != nil {
		return server.InternalServerError("Failed to process join results: " + err.Error())
	}

	response := types.JSONResponse{
		Success: true,
		Data:    data,
		Count:   int64(len(data)),
		Query:   query,
	}

	return sendJSONResponse(w, response)
}

// handleSelectWithJoin handles SELECT queries with joins using the Joins array
func handleSelectWithJoin(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" {
		return server.BadRequest("Base table name is required")
	}

	if len(req.Joins) == 0 {
		return server.BadRequest("At least one join is required")
	}

	// Build columns to select
	columns := "*"
	if len(req.Columns) > 0 {
		columns = strings.Join(req.Columns, ", ")
	}

	// Start building the query
	query := fmt.Sprintf("SELECT %s FROM %s", columns, req.Table)

	// Add joins
	for _, join := range req.Joins {
		joinType := "INNER JOIN"
		if join.Type != "" {
			joinType = strings.ToUpper(join.Type) + " JOIN"
		}
		query += fmt.Sprintf(" %s %s ON %s", joinType, join.Table, join.Condition)
	}

	// Add WHERE clause
	args := []interface{}{}
	if req.Where != "" {
		query += " WHERE " + req.Where
		args = append(args, req.WhereArgs...)
	}

	// Add GROUP BY
	if req.GroupBy != "" {
		query += " GROUP BY " + req.GroupBy
	}

	// Add HAVING
	if req.Having != "" {
		query += " HAVING " + req.Having
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

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		return server.InternalServerError("Failed to execute select with joins: " + err.Error())
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

// handleCountWithJoin handles COUNT queries with joins
func handleCountWithJoin(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" {
		return server.BadRequest("Base table name is required")
	}

	if len(req.Joins) == 0 {
		return server.BadRequest("At least one join is required")
	}

	// Start building the query
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", req.Table)

	// Add joins
	for _, join := range req.Joins {
		joinType := "INNER JOIN"
		if join.Type != "" {
			joinType = strings.ToUpper(join.Type) + " JOIN"
		}
		query += fmt.Sprintf(" %s %s ON %s", joinType, join.Table, join.Condition)
	}

	// Add WHERE clause
	args := []interface{}{}
	if req.Where != "" {
		query += " WHERE " + req.Where
		args = append(args, req.WhereArgs...)
	}

	// Add GROUP BY
	if req.GroupBy != "" {
		query += " GROUP BY " + req.GroupBy
	}

	// Add HAVING
	if req.Having != "" {
		query += " HAVING " + req.Having
	}

	// Execute the count query
	var count int64
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return server.InternalServerError("Failed to execute count with joins: " + err.Error())
	}

	response := types.JSONResponse{
		Success: true,
		Count:   count,
		Data:    map[string]interface{}{"count": count},
		Query:   query,
	}

	return sendJSONResponse(w, response)
}

// handleQueryBuilder handles complex queries using a query builder approach
func handleQueryBuilder(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" {
		return server.BadRequest("Base table name is required")
	}

	// Build columns to select
	columns := "*"
	if len(req.Columns) > 0 {
		columns = strings.Join(req.Columns, ", ")
	}

	// Start building the query
	query := fmt.Sprintf("SELECT %s FROM %s", columns, req.Table)

	// Add joins if specified
	for _, join := range req.Joins {
		joinType := "INNER JOIN"
		if join.Type != "" {
			joinType = strings.ToUpper(join.Type) + " JOIN"
		}
		query += fmt.Sprintf(" %s %s ON %s", joinType, join.Table, join.Condition)
	}

	// Add WHERE clause
	args := []interface{}{}
	if req.Where != "" {
		query += " WHERE " + req.Where
		args = append(args, req.WhereArgs...)
	}

	// Add GROUP BY
	if req.GroupBy != "" {
		query += " GROUP BY " + req.GroupBy
	}

	// Add HAVING
	if req.Having != "" {
		query += " HAVING " + req.Having
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

	// Execute the query
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
