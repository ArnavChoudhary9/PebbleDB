package main

import (
    "database/sql"
    "fmt"
    "reflect"
    "strings"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

// DB represents a SQLite database wrapper
type DB struct {
    conn *sql.DB
    path string
}

// Config holds database configuration options
type Config struct {
    Path            string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    WALMode         bool
    ForeignKeys     bool
}

// NewDB creates a new database connection with the given configuration
func NewDB(config Config) (*DB, error) {
    dsn := config.Path

    // Add query parameters for SQLite optimizations
    params := []string{}
    if config.WALMode {
        params = append(params, "_journal_mode=WAL")
    }
    if config.ForeignKeys {
        params = append(params, "_foreign_keys=on")
    }

    if len(params) > 0 {
        dsn += "?" + strings.Join(params, "&")
    }

    conn, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Set connection pool settings
    if config.MaxOpenConns > 0 {
        conn.SetMaxOpenConns(config.MaxOpenConns)
    }
    if config.MaxIdleConns > 0 {
        conn.SetMaxIdleConns(config.MaxIdleConns)
    }
    if config.ConnMaxLifetime > 0 {
        conn.SetConnMaxLifetime(config.ConnMaxLifetime)
    }

    // Test the connection
    if err := conn.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return &DB{
        conn: conn,
        path: config.Path,
    }, nil
}

// Close closes the database connection
func (db *DB) Close() error {
    if db.conn != nil {
        return db.conn.Close()
    }
    return nil
}

// Ping verifies the database connection is alive
func (db *DB) Ping() error {
    return db.conn.Ping()
}

// Exec executes a query without returning any rows
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
    return db.conn.Exec(query, args...)
}

// Query executes a query that returns rows
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
    return db.conn.Query(query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
    return db.conn.QueryRow(query, args...)
}

// Prepare creates a prepared statement for later queries or executions
func (db *DB) Prepare(query string) (*sql.Stmt, error) {
    return db.conn.Prepare(query)
}

// Transaction represents a database transaction
type Transaction struct {
    tx *sql.Tx
}

// Begin starts a new transaction
func (db *DB) Begin() (*Transaction, error) {
    tx, err := db.conn.Begin()
    if err != nil {
        return nil, err
    }
    return &Transaction{tx: tx}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
    return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
    return t.tx.Rollback()
}

// Exec executes a query within the transaction
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
    return t.tx.Exec(query, args...)
}

// Query executes a query within the transaction
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
    return t.tx.Query(query, args...)
}

// QueryRow executes a query within the transaction
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
    return t.tx.QueryRow(query, args...)
}

// CreateTable creates a table with the given schema
func (db *DB) CreateTable(tableName string, schema string) error {
    query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, schema)
    _, err := db.Exec(query)
    return err
}

// DropTable drops a table
func (db *DB) DropTable(tableName string) error {
    query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
    _, err := db.Exec(query)
    return err
}

// TableExists checks if a table exists
func (db *DB) TableExists(tableName string) (bool, error) {
    query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
    var name string
    err := db.QueryRow(query, tableName).Scan(&name)
    if err == sql.ErrNoRows {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

// GetTableSchema returns the schema of a table
func (db *DB) GetTableSchema(tableName string) (string, error) {
    query := "SELECT sql FROM sqlite_master WHERE type='table' AND name=?"
    var schema string
    err := db.QueryRow(query, tableName).Scan(&schema)
    return schema, err
}

// ListTables returns a list of all tables in the database
func (db *DB) ListTables() ([]string, error) {
    query := "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
            return nil, err
        }
        tables = append(tables, name)
    }
    return tables, rows.Err()
}

// Insert inserts a new record into the specified table
func (db *DB) Insert(tableName string, data map[string]interface{}) (int64, error) {
    columns := make([]string, 0, len(data))
    placeholders := make([]string, 0, len(data))
    values := make([]interface{}, 0, len(data))

    for column, value := range data {
        columns = append(columns, column)
        placeholders = append(placeholders, "?")
        values = append(values, value)
    }

    query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
        tableName,
        strings.Join(columns, ", "),
        strings.Join(placeholders, ", "))

    result, err := db.Exec(query, values...)
    if err != nil {
        return 0, err
    }

    return result.LastInsertId()
}

// Update updates records in the specified table
func (db *DB) Update(tableName string, data map[string]interface{}, where string, whereArgs ...interface{}) (int64, error) {
    setParts := make([]string, 0, len(data))
    values := make([]interface{}, 0, len(data))

    for column, value := range data {
        setParts = append(setParts, column+" = ?")
        values = append(values, value)
    }

    query := fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(setParts, ", "))
    if where != "" {
        query += " WHERE " + where
        values = append(values, whereArgs...)
    }

    result, err := db.Exec(query, values...)
    if err != nil {
        return 0, err
    }

    return result.RowsAffected()
}

// Delete deletes records from the specified table
func (db *DB) Delete(tableName string, where string, whereArgs ...interface{}) (int64, error) {
    query := fmt.Sprintf("DELETE FROM %s", tableName)
    if where != "" {
        query += " WHERE " + where
    }

    result, err := db.Exec(query, whereArgs...)
    if err != nil {
        return 0, err
    }

    return result.RowsAffected()
}

// Select performs a SELECT query and returns the results
func (db *DB) Select(tableName string, columns []string, where string, whereArgs ...interface{}) (*sql.Rows, error) {
    columnStr := "*"
    if len(columns) > 0 {
        columnStr = strings.Join(columns, ", ")
    }

    query := fmt.Sprintf("SELECT %s FROM %s", columnStr, tableName)
    if where != "" {
        query += " WHERE " + where
    }

    return db.Query(query, whereArgs...)
}

// SelectOne performs a SELECT query that returns a single row
func (db *DB) SelectOne(tableName string, columns []string, where string, whereArgs ...interface{}) *sql.Row {
    columnStr := "*"
    if len(columns) > 0 {
        columnStr = strings.Join(columns, ", ")
    }

    query := fmt.Sprintf("SELECT %s FROM %s", columnStr, tableName)
    if where != "" {
        query += " WHERE " + where
    }
    query += " LIMIT 1"

    return db.QueryRow(query, whereArgs...)
}

// Count returns the number of rows in a table or matching a condition
func (db *DB) Count(tableName string, where string, whereArgs ...interface{}) (int64, error) {
    query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
    if where != "" {
        query += " WHERE " + where
    }

    var count int64
    err := db.QueryRow(query, whereArgs...).Scan(&count)
    return count, err
}

// Backup creates a backup of the database
func (db *DB) Backup(backupPath string) error {
    query := fmt.Sprintf("VACUUM INTO '%s'", backupPath)
    _, err := db.Exec(query)
    return err
}

// GetVersion returns the SQLite version
func (db *DB) GetVersion() (string, error) {
    var version string
    err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
    return version, err
}

// Vacuum performs a VACUUM operation to reclaim space
func (db *DB) Vacuum() error {
    _, err := db.Exec("VACUUM")
    return err
}

// Analyze updates the internal statistics used by the query planner
func (db *DB) Analyze() error {
    _, err := db.Exec("ANALYZE")
    return err
}

// GetStats returns basic database statistics
func (db *DB) GetStats() (map[string]interface{}, error) {
    stats := make(map[string]interface{})

    // Get page count
    var pageCount int64
    err := db.QueryRow("PRAGMA page_count").Scan(&pageCount)
    if err != nil {
        return nil, err
    }
    stats["page_count"] = pageCount

    // Get page size
    var pageSize int64
    err = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
    if err != nil {
        return nil, err
    }
    stats["page_size"] = pageSize

    // Calculate database size
    stats["database_size"] = pageCount * pageSize

    // Get user version
    var userVersion int64
    err = db.QueryRow("PRAGMA user_version").Scan(&userVersion)
    if err != nil {
        return nil, err
    }
    stats["user_version"] = userVersion

    return stats, nil
}

// STRUCT-BASED ORM FUNCTIONALITY

// getStructFields extracts field information from a struct
func getStructFields(v reflect.Value) ([]string, []interface{}, error) {
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }

    if v.Kind() != reflect.Struct {
        return nil, nil, fmt.Errorf("expected struct, got %s", v.Kind())
    }

    t := v.Type()
    var fields []string
    var values []interface{}

    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        value := v.Field(i)

        // Skip unexported fields
        if !field.IsExported() {
            continue
        }

        // Get field name from tag or use field name
        fieldName := field.Name
        if tag := field.Tag.Get("db"); tag != "" && tag != "-" {
            fieldName = tag
        }

        // Skip auto-increment id fields for inserts
        if strings.ToLower(fieldName) == "id" && field.Tag.Get("auto") == "true" {
            continue
        }

        fields = append(fields, fieldName)
        values = append(values, value.Interface())
    }

    return fields, values, nil
}

// getStructFieldsForSelect gets field names for SELECT queries
func getStructFieldsForSelect(structType reflect.Type) []string {
    if structType.Kind() == reflect.Ptr {
        structType = structType.Elem()
    }

    var fields []string
    for i := 0; i < structType.NumField(); i++ {
        field := structType.Field(i)

        if !field.IsExported() {
            continue
        }

        fieldName := field.Name
        if tag := field.Tag.Get("db"); tag != "" && tag != "-" {
            fieldName = tag
        }

        fields = append(fields, fieldName)
    }

    return fields
}

// scanIntoStruct scans a row into a struct
func scanIntoStruct(rows *sql.Rows, dest interface{}) error {
    v := reflect.ValueOf(dest)
    if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
        return fmt.Errorf("dest must be a pointer to struct")
    }

    v = v.Elem()
    t := v.Type()

    // Get column names
    columns, err := rows.Columns()
    if err != nil {
        return err
    }

    // Create slice of pointers to scan into
    values := make([]interface{}, len(columns))
    fieldMap := make(map[string]int)

    // Map struct fields to column indices
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        if !field.IsExported() {
            continue
        }

        fieldName := strings.ToLower(field.Name)
        if tag := field.Tag.Get("db"); tag != "" && tag != "-" {
            fieldName = strings.ToLower(tag)
        }

        fieldMap[fieldName] = i
    }

    // Set up scan destinations
    for i, column := range columns {
        columnName := strings.ToLower(column)
        if fieldIndex, exists := fieldMap[columnName]; exists {
            field := v.Field(fieldIndex)
            values[i] = field.Addr().Interface()
        } else {
            // Use a dummy variable for unknown columns
            var dummy interface{}
            values[i] = &dummy
        }
    }

    return rows.Scan(values...)
}

// generateTableSchema generates CREATE TABLE SQL from a struct
func generateTableSchema(structType reflect.Type) string {
    if structType.Kind() == reflect.Ptr {
        structType = structType.Elem()
    }

    var columns []string

    for i := 0; i < structType.NumField(); i++ {
        field := structType.Field(i)

        if !field.IsExported() {
            continue
        }

        fieldName := field.Name
        if tag := field.Tag.Get("db"); tag != "" && tag != "-" {
            fieldName = tag
        }

        // Determine SQL type based on Go type
        sqlType := getSQLType(field.Type)

        // Add constraints from tags
        constraints := getConstraints(field)

        column := fieldName + " " + sqlType + constraints
        columns = append(columns, column)
    }

    return strings.Join(columns, ",\n    ")
}

// getSQLType maps Go types to SQLite types
func getSQLType(goType reflect.Type) string {
    switch goType.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        return "INTEGER"
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        return "INTEGER"
    case reflect.Float32, reflect.Float64:
        return "REAL"
    case reflect.Bool:
        return "INTEGER"
    case reflect.String:
        return "TEXT"
    case reflect.Slice:
        if goType.Elem().Kind() == reflect.Uint8 {
            return "BLOB"
        }
        return "TEXT"
    default:
        if goType == reflect.TypeOf(time.Time{}) {
            return "DATETIME"
        }
        return "TEXT"
    }
}

// getConstraints extracts SQL constraints from struct tags
func getConstraints(field reflect.StructField) string {
    var constraints []string

    if field.Tag.Get("primary") == "true" {
        constraints = append(constraints, "PRIMARY KEY")
    }

    if field.Tag.Get("auto") == "true" {
        constraints = append(constraints, "AUTOINCREMENT")
    }

    if field.Tag.Get("unique") == "true" {
        constraints = append(constraints, "UNIQUE")
    }

    if field.Tag.Get("notnull") == "true" {
        constraints = append(constraints, "NOT NULL")
    }

    if defaultVal := field.Tag.Get("default"); defaultVal != "" {
        constraints = append(constraints, "DEFAULT "+defaultVal)
    }

    if len(constraints) > 0 {
        return " " + strings.Join(constraints, " ")
    }
    return ""
}

// CreateTableFromStruct creates a table based on a struct definition
func (db *DB) CreateTableFromStruct(tableName string, structType interface{}) error {
    t := reflect.TypeOf(structType)
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }

    schema := generateTableSchema(t)
    query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n    %s\n)", tableName, schema)

    _, err := db.Exec(query)
    return err
}

// InsertStruct inserts a struct into the database
func (db *DB) InsertStruct(tableName string, data interface{}) (int64, error) {
    v := reflect.ValueOf(data)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }

    fields, values, err := getStructFields(v)
    if err != nil {
        return 0, err
    }

    placeholders := make([]string, len(fields))
    for i := range placeholders {
        placeholders[i] = "?"
    }

    query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
        tableName,
        strings.Join(fields, ", "),
        strings.Join(placeholders, ", "))

    result, err := db.Exec(query, values...)
    if err != nil {
        return 0, err
    }

    return result.LastInsertId()
}

// UpdateStruct updates records using a struct
func (db *DB) UpdateStruct(tableName string, data interface{}, where string, whereArgs ...interface{}) (int64, error) {
    v := reflect.ValueOf(data)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }

    fields, values, err := getStructFields(v)
    if err != nil {
        return 0, err
    }

    setParts := make([]string, len(fields))
    for i, field := range fields {
        setParts[i] = field + " = ?"
    }

    query := fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(setParts, ", "))
    if where != "" {
        query += " WHERE " + where
        values = append(values, whereArgs...)
    }

    result, err := db.Exec(query, values...)
    if err != nil {
        return 0, err
    }

    return result.RowsAffected()
}

// SelectStruct performs a SELECT query and scans results into structs
func (db *DB) SelectStruct(tableName string, dest interface{}, where string, whereArgs ...interface{}) error {
    destValue := reflect.ValueOf(dest)
    if destValue.Kind() != reflect.Ptr {
        return fmt.Errorf("dest must be a pointer")
    }

    destValue = destValue.Elem()
    if destValue.Kind() != reflect.Slice {
        return fmt.Errorf("dest must be a pointer to slice")
    }

    // Get the element type of the slice
    elementType := destValue.Type().Elem()

    // Get field names for SELECT
    fields := getStructFieldsForSelect(elementType)

    query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ", "), tableName)
    if where != "" {
        query += " WHERE " + where
    }

    rows, err := db.Query(query, whereArgs...)
    if err != nil {
        return err
    }
    defer rows.Close()

    // Create slice to hold results
    results := reflect.MakeSlice(destValue.Type(), 0, 0)

    for rows.Next() {
        // Create new instance of element type
        elem := reflect.New(elementType).Interface()

        if err := scanIntoStruct(rows, elem); err != nil {
            return err
        }

        // Append to results slice
        results = reflect.Append(results, reflect.ValueOf(elem).Elem())
    }

    // Set the destination slice
    destValue.Set(results)
    return rows.Err()
}

// SelectOneStruct performs a SELECT query that returns a single struct
func (db *DB) SelectOneStruct(tableName string, dest interface{}, where string, whereArgs ...interface{}) error {
    destValue := reflect.ValueOf(dest)
    if destValue.Kind() != reflect.Ptr {
        return fmt.Errorf("dest must be a pointer")
    }

    elementType := destValue.Type().Elem()
    fields := getStructFieldsForSelect(elementType)

    query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ", "), tableName)
    if where != "" {
        query += " WHERE " + where
    }
    query += " LIMIT 1"

    rows, err := db.Query(query, whereArgs...)
    if err != nil {
        return err
    }
    defer rows.Close()

    if !rows.Next() {
        return sql.ErrNoRows
    }

    return scanIntoStruct(rows, dest)
}

// Transaction methods for struct operations

// InsertStruct inserts a struct within a transaction
func (t *Transaction) InsertStruct(tableName string, data interface{}) (int64, error) {
    v := reflect.ValueOf(data)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }

    fields, values, err := getStructFields(v)
    if err != nil {
        return 0, err
    }

    placeholders := make([]string, len(fields))
    for i := range placeholders {
        placeholders[i] = "?"
    }

    query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
        tableName,
        strings.Join(fields, ", "),
        strings.Join(placeholders, ", "))

    result, err := t.Exec(query, values...)
    if err != nil {
        return 0, err
    }

    return result.LastInsertId()
}

// UpdateStruct updates records using a struct within a transaction
func (t *Transaction) UpdateStruct(tableName string, data interface{}, where string, whereArgs ...interface{}) (int64, error) {
    v := reflect.ValueOf(data)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }

    fields, values, err := getStructFields(v)
    if err != nil {
        return 0, err
    }

    setParts := make([]string, len(fields))
    for i, field := range fields {
        setParts[i] = field + " = ?"
    }

    query := fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(setParts, ", "))
    if where != "" {
        query += " WHERE " + where
        values = append(values, whereArgs...)
    }

    result, err := t.Exec(query, values...)
    if err != nil {
        return 0, err
    }

    return result.RowsAffected()
}
