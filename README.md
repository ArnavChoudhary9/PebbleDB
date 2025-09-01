# PebbleDB

A lightweight SQLite framework and REST API server written in Go, designed for rapid prototyping, embedded systems, and lightweight applications.

## 🚀 Features

- **RESTful JSON API** - Complete database operations via HTTP endpoints
- **SQLite Backend** - Fast, reliable, and file-based database
- **Custom HTTP Server** - Built-in routing and middleware support
- **ORM-like Operations** - Struct-based database operations
- **Schema Management** - Dynamic table creation and schema inference
- **Connection Pooling** - Optimized database connections
- **CORS Support** - Ready for web applications
- **Health Monitoring** - Built-in health checks and statistics

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Endpoints](#api-endpoints)
- [Usage Examples](#usage-examples)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)

## 🛠️ Installation

### Prerequisites

- Go 1.25.0 or higher
- SQLite3

### Clone and Build

```bash
git clone https://github.com/ArnavChoudhary9/PebbleDB.git
cd PebbleDB
go mod download
go build -o pebbledb
```

### Run the Server

```bash
./pebbledb
```

The server will start on `http://localhost:8080`

## ⚡ Quick Start

1. **Start the server:**

   ```bash
   go run main.go
   ```

2. **Create a table:**

   ```bash
   curl -X POST http://localhost:8080/api/db \
     -H "Content-Type: application/json" \
     -d '{
       "action": "create_table",
       "table": "users",
       "schema": {
         "id": {"type": "INTEGER", "constraints": ["PRIMARY KEY", "AUTOINCREMENT"]},
         "name": {"type": "TEXT", "constraints": ["NOT NULL"]},
         "email": {"type": "TEXT", "constraints": ["UNIQUE"]}
       }
     }'
   ```

3. **Insert data:**

   ```bash
   curl -X POST http://localhost:8080/api/db \
     -H "Content-Type: application/json" \
     -d '{
       "action": "insert",
       "table": "users",
       "data": {
         "name": "Alice",
         "email": "alice@example.com"
       }
     }'
   ```

4. **Query data:**

   ```bash
   curl -X POST http://localhost:8080/api/db \
     -H "Content-Type: application/json" \
     -d '{
       "action": "select",
       "table": "users"
     }'
   ```

## 🌐 API Endpoints

### Core Database Operations

- **POST** `/api/db` - Main database operations endpoint

### Utility Endpoints

- **GET** `/` - Welcome message
- **GET** `/api/health` - Health check
- **GET** `/api/stats` - Database statistics  
- **GET** `/api/tables` - List all tables

### Supported Actions

| Action | Description |
|--------|-------------|
| `create_table` | Create a new table |
| `insert` | Insert records |
| `select` | Query records |
| `update` | Update records |
| `delete` | Delete records |
| `count` | Count records |
| `drop_table` | Drop a table |
| `table_exists` | Check if table exists |
| `get_schema` | Get table schema |

## 📖 Usage Examples

### Creating Tables

```json
{
  "action": "create_table",
  "table": "products",
  "schema": {
    "id": {"type": "INTEGER", "constraints": ["PRIMARY KEY", "AUTOINCREMENT"]},
    "name": {"type": "TEXT", "constraints": ["NOT NULL"]},
    "price": {"type": "REAL"},
    "in_stock": {"type": "INTEGER", "constraints": ["DEFAULT 1"]}
  }
}
```

### Inserting Data

```json
{
  "action": "insert",
  "table": "products",
  "data": {
    "name": "Laptop",
    "price": 999.99,
    "in_stock": 1
  }
}
```

### Querying with Filters

```json
{
  "action": "select",
  "table": "products",
  "columns": ["id", "name", "price"],
  "where": "price > ? AND in_stock = ?",
  "where_args": [500, 1],
  "order_by": "price DESC",
  "limit": 10
}
```

### Updating Records

```json
{
  "action": "update",
  "table": "products",
  "data": {"price": 899.99},
  "where": "id = ?",
  "where_args": [1]
}
```

For more examples, see [Examples.md](Examples.md).

## ⚙️ Configuration

The database configuration is set in [`main.go`](main.go):

```go
config := Config{
    Path:            "pdb_data/pebbledb.db",  // Database file path
    MaxOpenConns:    25,                       // Max open connections
    MaxIdleConns:    10,                       // Max idle connections
    ConnMaxLifetime: time.Hour,                // Connection lifetime
    WALMode:         true,                     // Enable WAL mode
    ForeignKeys:     true,                     // Enable foreign keys
}
```

## 📁 Project Structure

```text
PebbleDB/
├── main.go           # Server bootstrap and configuration
├── server.go         # HTTP server with custom routing
├── db.go             # Database wrapper and ORM functionality
├── db_api.go         # JSON API handlers for database operations
├── go.mod            # Go module dependencies
├── go.sum            # Go module checksums
├── Documentation.md  # Detailed API documentation
├── Examples.md       # Step-by-step usage examples
├── README.md         # This file
├── LICENSE           # MIT License
└── pdb_data/         # Database storage directory
    └── .gitkeep
```

### Key Components

- **[`main.go`](main.go)** - Entry point, server initialization
- **[`server.go`](server.go)** - Custom HTTP server with middleware support
- **[`db.go`](db.go)** - SQLite wrapper with ORM capabilities
- **[`db_api.go`](db_api.go)** - JSON API handlers for database operations

## 🔧 Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -ldflags="-s -w" -o pebbledb
```

### Database Location

By default, the SQLite database is stored in `pdb_data/pebbledb.db`. This directory is created automatically on first run.

## 📚 Documentation

- **[Documentation.md](Documentation.md)** - Complete API reference
- **[Examples.md](Examples.md)** - Progressive workflow examples

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🎯 Use Cases

- **Rapid Prototyping** - Quickly spin up a database-backed API
- **Embedded Systems** - Lightweight database for IoT and edge computing
- **Microservices** - Simple data persistence layer
- **Development & Testing** - Local database for development environments
- **Small Applications** - Perfect for apps that don't need a full database server

## 🔗 Links

- **Repository:** [https://github.com/ArnavChoudhary9/PebbleDB](https://github.com/ArnavChoudhary9/PebbleDB)
- **Issues:** [Report bugs or request features](https://github.com/ArnavChoudhary9/PebbleDB/issues)

---

**PebbleDB** - *A pebble can start an avalanche.*
