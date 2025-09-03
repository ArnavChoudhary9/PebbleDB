package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/ArnavChoudhary9/PebbleDB/pkg/types"
)

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) error {
	healthData := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"system": map[string]interface{}{
			"go_version":  runtime.Version(),
			"go_routines": runtime.NumGoroutine(),
			"memory_used": getMemoryUsage(),
		},
	}

	response := types.JSONResponse{
		Success: true,
		Data:    healthData,
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

// getMemoryUsage returns memory usage statistics
func getMemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc_mb":       m.Alloc / 1024 / 1024,
		"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
		"sys_mb":         m.Sys / 1024 / 1024,
		"num_gc":         m.NumGC,
	}
}
