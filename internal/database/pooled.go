package database

import (
	"fmt"
	"sync"
	"time"
)

var projectDBs = struct {
	sync.RWMutex
	conns map[string]*DB
}{conns: make(map[string]*DB)}

// GetProjectDB returns a database connection for a specific project
// It uses connection pooling to reuse existing connections
func GetProjectDB(basePath, key string) (*DB, error) {
	projectDBs.RLock()
	if db, ok := projectDBs.conns[key]; ok {
		projectDBs.RUnlock()
		return db, nil
	}
	projectDBs.RUnlock()

	projectDBs.Lock()
	defer projectDBs.Unlock()

	// Double check after upgrade to write lock
	if db, ok := projectDBs.conns[key]; ok {
		return db, nil
	}

	dbPath := fmt.Sprintf("%s/%s.db", basePath, key)
	cfg := Config{
		Path:            dbPath,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		WALMode:         true,
		ForeignKeys:     true,
	}

	db, err := NewDB(cfg)
	if err != nil {
		return nil, err
	}

	projectDBs.conns[key] = db
	return db, nil
}

// CloseProjectDB closes a specific project database connection
func CloseProjectDB(key string) error {
	projectDBs.Lock()
	defer projectDBs.Unlock()

	if db, ok := projectDBs.conns[key]; ok {
		err := db.Close()
		delete(projectDBs.conns, key)
		return err
	}
	return nil
}

// CloseAllProjectDBs closes all project database connections
func CloseAllProjectDBs() error {
	projectDBs.Lock()
	defer projectDBs.Unlock()

	var lastErr error
	for key, db := range projectDBs.conns {
		if err := db.Close(); err != nil {
			lastErr = err
		}
		delete(projectDBs.conns, key)
	}
	return lastErr
}
