package database

import (
	"database/sql"
	"fmt"
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
		conn.Close()
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
	if db.conn == nil {
		return fmt.Errorf("database connection is nil")
	}
	return db.conn.Ping()
}

// Exec executes a query without returning any rows
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	if db.conn == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	return db.conn.Exec(query, args...)
}

// Query executes a query that returns rows
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if db.conn == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	return db.conn.Query(query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	if db.conn == nil {
		// Return a row that will return an error when scanned
		return &sql.Row{}
	}
	return db.conn.QueryRow(query, args...)
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

// Prepare creates a prepared statement for later queries or executions
func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	if db.conn == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	return db.conn.Prepare(query)
}

// Transaction represents a database transaction
type Transaction struct {
	tx *sql.Tx
}

// Begin starts a new transaction
func (db *DB) Begin() (*Transaction, error) {
	if db.conn == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	if t.tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	if t.tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	return t.tx.Rollback()
}

// Exec executes a query within the transaction
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	if t.tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	return t.tx.Exec(query, args...)
}

// Query executes a query within the transaction
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if t.tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	return t.tx.Query(query, args...)
}

// QueryRow executes a query within the transaction
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	if t.tx == nil {
		return &sql.Row{}
	}
	return t.tx.QueryRow(query, args...)
}
