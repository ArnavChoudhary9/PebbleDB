# 📖 PebbleDB Server Documentation

## Overview

PebbleDB Server is a lightweight HTTP server written in Go, built on top of **SQLite** with custom routing and middleware. It provides a JSON-based API for managing databases (creating tables, inserting, selecting, updating, deleting data, etc.), as well as utility endpoints for health checks and statistics.

The server runs on **port 8080** by default and organizes routes into:

* **Root routes** (e.g., `/`)
* **API routes** (under `/api`)

---

## ⚙️ Server Structure

* **main.go**

  * Bootstraps the server.
  * Initializes SQLite database.
  * Registers middleware.
  * Defines routes (`/`, `/api/...`).

* **server.go**

  * Implements a minimal HTTP server with:

    * Custom router.
    * Route grouping (`/api`).
    * Middleware support (logging, CORS, DB injection).

* **db.go**

  * Provides a full wrapper around SQLite:

    * Connection pooling.
    * CRUD operations (`Insert`, `Update`, `Delete`, `Select`).
    * Table management (`CreateTable`, `DropTable`, `ListTables`, `GetTableSchema`).
    * Maintenance (`Vacuum`, `Analyze`, `Backup`).
    * ORM-style struct operations.

* **db\_api.go**

  * Implements JSON-based API endpoints for database operations.
  * Provides a middleware (`dbMiddleware`) to inject the database into requests.
  * Handles JSON request parsing and response formatting.

---

## 🌐 API Endpoints

### Root Endpoints

#### **`GET /`**

* **Description:** Root endpoint of the server.
* **Handler:** `homeHandler`
* **Response:**

  ```text
  Welcome to PebbleDB Server!
  ```

* **Use case:** Quick check to confirm the server is running.

---

### API Endpoints (`/api/...`)

#### **`POST /api/db`**

* **Description:** Central endpoint for all database operations.
  Expects a **JSON body** specifying an `action` and parameters.
* **Handler:** `handleDatabaseRequest`
* **Supported actions:**

| Action         | Description                                                                  |
| -------------- | ---------------------------------------------------------------------------- |
| `create_table` | Creates a new table (schema can be provided manually or inferred from data). |
| `insert`       | Inserts a new record into a table.                                           |
| `select`       | Selects records from a table (with optional filters, ordering, pagination).  |
| `update`       | Updates records in a table based on a condition.                             |
| `delete`       | Deletes records from a table based on a condition.                           |
| `count`        | Returns the number of rows in a table or matching a condition.               |
| `drop_table`   | Drops a table if it exists.                                                  |
| `table_exists` | Checks if a given table exists.                                              |
| `get_schema`   | Returns the SQL schema of a given table.                                     |

* **Request format (example – insert):**

  ```json
  {
    "action": "insert",
    "table": "users",
    "data": {
      "name": "Alice",
      "age": 25
    }
  }
  ```

* **Response format (success):**

  ```json
  {
    "success": true,
    "id": 1,
    "data": {
      "inserted_id": 1
    }
  }
  ```

* **Response format (error):**

  ```json
  {
    "success": false,
    "error": "Failed to insert record: no such table: users"
  }
  ```

---

#### **`GET /api/health`**

* **Description:** Health check endpoint.
* **Handler:** `handleHealth`
* **Response (example):**

  ```json
  {
    "success": true,
    "data": "OK"
  }
  ```

* **Use case:** Used by monitoring tools to check if the server and DB are responsive.

---

#### **`GET /api/stats`**

* **Description:** Returns database statistics.
* **Handler:** `handleStats`
* **Response (example):**

  ```json
  {
    "success": true,
    "data": {
      "page_count": 32,
      "page_size": 4096,
      "database_size": 131072,
      "user_version": 0
    }
  }
  ```

* **Use case:** Monitor DB size, performance, and health.

---

#### **`GET /api/tables`**

* **Description:** Lists all tables in the database.
* **Handler:** `handleTables`
* **Response (example):**

  ```json
  {
    "success": true,
    "data": ["users", "orders", "products"]
  }
  ```
  
* **Use case:** Discover database schema dynamically.

---

## 🛡️ Middleware

* **LoggingMiddleware**

  * Logs every request method, path, and client address.

* **CORSMiddleware**

  * Enables cross-origin requests with standard headers.
  * Handles `OPTIONS` preflight requests.

* **dbMiddleware**

  * Injects database connection into request context.
  * Makes DB available in handlers via `getDB(r)`.

---

✅ With this setup, PebbleDB Server acts as a **SQLite-backed REST API**, useful for lightweight apps, prototyping, and embedded systems.
