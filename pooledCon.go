package main

import (
	"fmt"
	"sync"
	"time"
)

var projectDBs = struct {
	sync.RWMutex
	conns map[string]*DB
}{conns: make(map[string]*DB)}

func getProjectDB(basePath, key string) (*DB, error) {
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
