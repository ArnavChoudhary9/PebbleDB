package server

import (
	"log"
	"net/http"
	"strings"
)

// HTTPHandlerFunc is a handler function that can return an error
type HTTPHandlerFunc func(http.ResponseWriter, *http.Request) error

// Middleware function type updated to work with HTTPHandlerFunc
type Middleware func(HTTPHandlerFunc) HTTPHandlerFunc

// Route represents a single route
type Route struct {
	Method  string
	Pattern string
	Handler HTTPHandlerFunc
}

// Server represents the HTTP server with routing capabilities
type Server struct {
	routes      map[string][]Route // Group routes by pattern
	middlewares []Middleware
	mux         *http.ServeMux
}

// NewServer creates a new server instance
func NewServer() *Server {
	return &Server{
		routes:      make(map[string][]Route),
		middlewares: make([]Middleware, 0),
		mux:         http.NewServeMux(),
	}
}

// Use adds middleware to the server
func (s *Server) Use(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

// addRoute adds a route to the routes map
func (s *Server) addRoute(method, pattern string, handler HTTPHandlerFunc) {
	route := Route{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	}
	s.routes[pattern] = append(s.routes[pattern], route)
}

// Handle adds a route with any HTTP method
func (s *Server) Handle(pattern string, handler HTTPHandlerFunc) {
	s.addRoute("", pattern, handler)
}

// GET adds a GET route
func (s *Server) GET(pattern string, handler HTTPHandlerFunc) {
	s.addRoute("GET", pattern, handler)
}

// POST adds a POST route
func (s *Server) POST(pattern string, handler HTTPHandlerFunc) {
	s.addRoute("POST", pattern, handler)
}

// PUT adds a PUT route
func (s *Server) PUT(pattern string, handler HTTPHandlerFunc) {
	s.addRoute("PUT", pattern, handler)
}

// DELETE adds a DELETE route
func (s *Server) DELETE(pattern string, handler HTTPHandlerFunc) {
	s.addRoute("DELETE", pattern, handler)
}

// Group creates a route group with common prefix
func (s *Server) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		server: s,
		prefix: prefix,
	}
}

// Start starts the server with all registered routes and middleware
func (s *Server) Start(port string) error {
	// Register each unique pattern once with a method dispatcher
	for pattern, routes := range s.routes {
		httpHandler := s.createMethodDispatcher(routes)

		// Convert http.HandlerFunc to HTTPHandlerFunc for middleware processing
		handler := func(w http.ResponseWriter, r *http.Request) error {
			httpHandler(w, r)
			return nil
		}

		// Apply middleware in reverse order
		for i := len(s.middlewares) - 1; i >= 0; i-- {
			handler = s.middlewares[i](handler)
		}

		// Convert back to http.HandlerFunc for registration
		finalHandler := func(w http.ResponseWriter, r *http.Request) {
			if err := handler(w, r); err != nil {
				s.handleError(w, err)
			}
		}

		s.mux.HandleFunc(pattern, finalHandler)
	}

	log.Printf("Server starting on port %s\n", port)
	return http.ListenAndServe(port, s.mux)
}

// createMethodDispatcher creates a handler that dispatches based on HTTP method
func (s *Server) createMethodDispatcher(routes []Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Find matching route for the HTTP method
		for _, route := range routes {
			if route.Method == "" || route.Method == r.Method {
				if err := route.Handler(w, r); err != nil {
					s.handleError(w, err)
				}
				return
			}
		}

		// If no route matches, return method not allowed
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleError handles errors returned by handlers
func (s *Server) handleError(w http.ResponseWriter, err error) {
	if httpErr, ok := err.(HTTPError); ok {
		http.Error(w, httpErr.Message, httpErr.Code)
	} else {
		// Handle regular errors as internal server errors
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Internal server error: %v", err)
	}
}

// RouteGroup represents a group of routes with common prefix
type RouteGroup struct {
	server *Server
	prefix string
}

// Handle adds a route to the group (any method)
func (rg *RouteGroup) Handle(pattern string, handler HTTPHandlerFunc) {
	fullPattern := rg.buildFullPattern(pattern)
	rg.server.Handle(fullPattern, handler)
}

// GET adds a GET route to the group
func (rg *RouteGroup) GET(pattern string, handler HTTPHandlerFunc) {
	fullPattern := rg.buildFullPattern(pattern)
	rg.server.GET(fullPattern, handler)
}

// POST adds a POST route to the group
func (rg *RouteGroup) POST(pattern string, handler HTTPHandlerFunc) {
	fullPattern := rg.buildFullPattern(pattern)
	rg.server.POST(fullPattern, handler)
}

// PUT adds a PUT route to the group
func (rg *RouteGroup) PUT(pattern string, handler HTTPHandlerFunc) {
	fullPattern := rg.buildFullPattern(pattern)
	rg.server.PUT(fullPattern, handler)
}

// DELETE adds a DELETE route to the group
func (rg *RouteGroup) DELETE(pattern string, handler HTTPHandlerFunc) {
	fullPattern := rg.buildFullPattern(pattern)
	rg.server.DELETE(fullPattern, handler)
}

// buildFullPattern constructs the full pattern with prefix
func (rg *RouteGroup) buildFullPattern(pattern string) string {
	prefix := strings.TrimSuffix(rg.prefix, "/")
	pattern = strings.TrimPrefix(pattern, "/")

	if pattern == "" {
		return prefix + "/"
	}
	return prefix + "/" + pattern
}
