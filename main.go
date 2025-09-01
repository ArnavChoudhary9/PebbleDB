package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

func main() {
    // Create server instance
    server := NewServer()
    
    // Add global middleware
    server.Use(LoggingMiddleware)
    server.Use(CORSMiddleware)
    
    // Add root routes
    server.GET("/", homeHandler)
    
    // Create API route group
    apiGroup := server.Group("/api")
    apiGroup.GET("/health", healthHandler)
    apiGroup.GET("/status", statusHandler)
    
    // Create users route group with different HTTP methods
    usersGroup := server.Group("/api/users")
    usersGroup.GET("/", getAllUsersHandler)
    usersGroup.POST("/", createUserHandler)
    usersGroup.GET("/{id}", getUserHandler)
    usersGroup.PUT("/{id}", updateUserHandler)
    usersGroup.DELETE("/{id}", deleteUserHandler)
    
    // Start server
    log.Fatal(server.Start(":8080"))
}

// Handler functions
func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to PebbleDB Server!")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    response := map[string]interface{}{
        "status": "healthy",
        "service": "PebbleDB",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
    response := map[string]interface{}{
        "uptime": "24h",
        "connections": 42,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
    users := []map[string]string{
        {"id": "1", "name": "John Doe"},
        {"id": "2", "name": "Jane Smith"},
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
    // Extract ID from URL path (simple implementation)
    user := map[string]string{
        "id": "1",
        "name": "John Doe",
        "email": "john@example.com",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    // Parse request body here
    response := map[string]string{
        "message": "User created successfully",
        "id": "3",
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
    // Parse request body and update user
    response := map[string]string{
        "message": "User updated successfully",
        "id": "1",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
    response := map[string]string{
        "message": "User deleted successfully",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
