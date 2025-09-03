package handlers

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/ArnavChoudhary9/PebbleDB/internal/database"
	"github.com/ArnavChoudhary9/PebbleDB/internal/server"
	"github.com/ArnavChoudhary9/PebbleDB/pkg/types"
)

// handleCreateTable handles table creation from JSON schema
func handleCreateTable(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" {
		return server.BadRequest("Table name is required")
	}

	var schema string
	if req.Schema != nil {
		schema = generateSchemaFromJSON(req.Schema)
	} else if req.Data != nil {
		// Auto-generate schema from sample data
		schema = inferSchemaFromData(req.Data)
	} else {
		return server.BadRequest("Schema or sample data is required")
	}

	err := db.CreateTable(req.Table, schema)
	if err != nil {
		return server.InternalServerError("Failed to create table: " + err.Error())
	}

	return sendSuccess(w, map[string]string{"message": "Table created successfully"})
}

// handleDropTable handles table deletion
func handleDropTable(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" {
		return server.BadRequest("Table name is required")
	}

	err := db.DropTable(req.Table)
	if err != nil {
		return server.InternalServerError("Failed to drop table: " + err.Error())
	}

	return sendSuccess(w, map[string]string{"message": "Table dropped successfully"})
}

// handleTableExists checks if table exists
func handleTableExists(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" {
		return server.BadRequest("Table name is required")
	}

	exists, err := db.TableExists(req.Table)
	if err != nil {
		return server.InternalServerError("Failed to check table existence: " + err.Error())
	}

	return sendSuccess(w, map[string]interface{}{
		"table":  req.Table,
		"exists": exists,
	})
}

// handleGetSchema gets table schema
func handleGetSchema(w http.ResponseWriter, req types.JSONRequest, db *database.DB) error {
	if req.Table == "" {
		return server.BadRequest("Table name is required")
	}

	schema, err := db.GetTableSchema(req.Table)
	if err != nil {
		return server.InternalServerError("Failed to get table schema: " + err.Error())
	}

	return sendSuccess(w, map[string]interface{}{
		"table":  req.Table,
		"schema": schema,
	})
}

// Helper functions for schema generation

// generateSchemaFromJSON creates SQL schema from JSON schema definition
func generateSchemaFromJSON(schema map[string]interface{}) string {
	var parts []string

	for column, def := range schema {
		var columnDef string

		switch typeDef := def.(type) {
		case string:
			// Simple type definition like "TEXT", "INTEGER", etc.
			columnDef = fmt.Sprintf("%s %s", column, typeDef)
		case map[string]interface{}:
			// Complex definition with type and constraints
			if columnType, ok := typeDef["type"].(string); ok {
				columnDef = fmt.Sprintf("%s %s", column, strings.ToUpper(columnType))

				// Add constraints
				if primaryKey, ok := typeDef["primary_key"].(bool); ok && primaryKey {
					columnDef += " PRIMARY KEY"
				}
				if autoIncrement, ok := typeDef["auto_increment"].(bool); ok && autoIncrement {
					columnDef += " AUTOINCREMENT"
				}
				if notNull, ok := typeDef["not_null"].(bool); ok && notNull {
					columnDef += " NOT NULL"
				}
				if unique, ok := typeDef["unique"].(bool); ok && unique {
					columnDef += " UNIQUE"
				}
				if defaultVal, ok := typeDef["default"]; ok {
					columnDef += fmt.Sprintf(" DEFAULT %v", defaultVal)
				}
			}
		}

		if columnDef != "" {
			parts = append(parts, columnDef)
		}
	}

	return strings.Join(parts, ", ")
}

// inferSchemaFromData infers SQL schema from sample data
func inferSchemaFromData(data map[string]interface{}) string {
	var parts []string

	for column, value := range data {
		sqlType := getSQLTypeFromValue(value)
		parts = append(parts, fmt.Sprintf("%s %s", column, sqlType))
	}

	return strings.Join(parts, ", ")
}

// getSQLTypeFromValue maps Go values to SQLite types
func getSQLTypeFromValue(value interface{}) string {
	if value == nil {
		return "TEXT"
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.String:
		return "TEXT"
	default:
		return "TEXT"
	}
}
