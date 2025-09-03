# PebbleDB - Refactored Structure

This document describes the new, improved structure of the PebbleDB project after refactoring for better organization and maintainability.

## Project Structure

```text
PebbleDB/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── auth/                       # Authentication & Authorization
│   │   ├── auth.go                 # Main auth middleware
│   │   ├── jwt.go                  # JWT token handling
│   │   └── cookie.go               # Cookie utilities
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── database/                   # Database layer
│   │   ├── db.go                   # Core database wrapper
│   │   ├── pooled.go               # Connection pooling
│   │   ├── query_builder.go        # Query builder with JOIN support
│   │   └── middleware.go           # Database middleware
│   ├── handlers/                   # HTTP handlers
│   │   ├── api.go                  # Main API handler dispatcher
│   │   ├── crud.go                 # Basic CRUD operations
│   │   ├── health.go               # Health check endpoint
│   │   ├── projects.go             # Project management
│   │   └── routes.go               # Route setup and configuration
│   └── server/                     # HTTP server core
│       ├── server.go               # Server implementation
│       ├── middleware.go           # Common middleware
│       └── errors.go               # Error handling
├── pkg/
│   └── types/
│       └── types.go                # Shared types and constants
├── pdb_data/                       # Data directory
├── go.mod
├── go.sum
├── README.md
├── Documentation.md
└── Examples.md
```

## Key Improvements

### 1. **Clear Separation of Concerns**

- **cmd/**: Application entry points
- **internal/**: Private application code
- **pkg/**: Public, reusable packages

### 2. **Modular Architecture**

- **auth/**: All authentication logic isolated
- **config/**: Centralized configuration management
- **database/**: Database layer with connection pooling
- **handlers/**: HTTP request handling logic
- **server/**: Core HTTP server functionality

### 3. **Better Code Organization**

- Single responsibility for each package
- Clear dependencies between packages
- Shared types in pkg/types for consistency

### 4. **Improved Maintainability**

- Smaller, focused files (< 200 lines each)
- Clear naming conventions
- Better error handling
- Comprehensive middleware system

## Package Descriptions

### cmd/server

Contains the main application entry point. Minimal code that just wires everything together.

### internal/auth

Handles all authentication and authorization:

- JWT token verification
- Cookie management
- User context injection
- JWKS key fetching and caching

### internal/config

Manages application configuration:

- Environment variable loading
- Configuration validation
- Centralized config struct

### internal/database

Database layer with:

- SQLite connection management
- Connection pooling for projects
- Query builder for complex queries
- Database middleware for request injection

### internal/handlers

HTTP request handlers:

- Main API dispatcher
- CRUD operations (Create, Read, Update, Delete)
- Project management
- Health checks
- Route configuration

### internal/server

Core HTTP server functionality:

- Custom HTTP handler interface with error returns
- Middleware system
- Route grouping
- Error handling

### pkg/types

Shared types and constants used across packages.

## Migration Benefits

1. **Easier Testing**: Each package can be tested independently
2. **Better Code Reuse**: Shared types and utilities in pkg/
3. **Cleaner Dependencies**: Clear import paths and dependency direction
4. **Easier Maintenance**: Smaller files are easier to understand and modify
5. **Better Onboarding**: New developers can understand the structure quickly
6. **Scalability**: Easy to add new features without affecting existing code

## Usage

To run the refactored server:

```bash
go run ./cmd/server
```

To build:

```bash
go build ./cmd/server
```

## Configuration

The server still uses the same environment variables as before:

- `JWKS_URL`
- `AUTH_TOKEN_NAME`
- `TOKEN_REFRESH_URL`
- `TOKEN_REFRESH_KEY`
- `COOKIE_DOMAIN`

## API Compatibility

The refactored version maintains full API compatibility with the previous version. All existing endpoints and request/response formats remain unchanged.

## Future Enhancements

The new structure makes it easy to add:

- More authentication providers
- Additional database backends
- New API endpoints
- Enhanced middleware
- Better testing coverage
- Metrics and monitoring
