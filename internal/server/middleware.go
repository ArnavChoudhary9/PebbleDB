package server

import (
	"context"
	"log"
	"net/http"

	"github.com/ArnavChoudhary9/PebbleDB/pkg/types"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next HTTPHandlerFunc) HTTPHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Printf("[%s] %s %s\n", r.Method, r.URL.Path, r.RemoteAddr)
		return next(w, r)
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next HTTPHandlerFunc) HTTPHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return nil
		}

		return next(w, r)
	}
}

// WorkingDirectoryMiddleware adds working directory to request context
func WorkingDirectoryMiddleware(basePath string) func(HTTPHandlerFunc) HTTPHandlerFunc {
	return func(next HTTPHandlerFunc) HTTPHandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx := context.WithValue(r.Context(), types.WorkingDirectoryContextKey, basePath)
			return next(w, r.WithContext(ctx))
		}
	}
}
