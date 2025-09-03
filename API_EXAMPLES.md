# PebbleDB API Test Examples

This document provides examples of all implemented API functions for the `/api/db` endpoint.

## Project Management APIs

### 1. Create Project

```json
{
  "action": "create_project",
  "project_name": "my_test_app",
  "project_description": "A test application database"
}
```

### 2. List Projects

```json
{
  "action": "list_projects"
}
```

### 3. Get Project

```json
{
  "action": "get_project",
  "project_id": "proj_1725360000"
}
```

### 4. Delete Project

```json
{
  "action": "delete_project",
  "project_id": "proj_1725360000"
}
```

## Table Management APIs

### 5. Create Table (with explicit schema)

```json
{
  "action": "create_table",
  "project_id": "proj_1725360000",
  "table": "users",
  "schema": {
    "id": {
      "type": "INTEGER",
      "primary_key": true,
      "auto_increment": true
    },
    "name": {
      "type": "TEXT",
      "not_null": true
    },
    "email": {
      "type": "TEXT",
      "unique": true,
      "not_null": true
    },
    "age": "INTEGER",
    "created_at": {
      "type": "DATETIME",
      "default": "CURRENT_TIMESTAMP"
    }
  }
}
```

### 6. Create Table (infer from data)

```json
{
  "action": "create_table",
  "project_id": "proj_1725360000",
  "table": "products",
  "data": {
    "name": "Sample Product",
    "price": 29.99,
    "category": "Electronics",
    "in_stock": true,
    "quantity": 100
  }
}
```

### 7. Get Tables List

```json
{
  "action": "get_tables",
  "project_id": "proj_1725360000"
}
```

### 8. Table Exists Check

```json
{
  "action": "table_exists",
  "project_id": "proj_1725360000",
  "table": "users"
}
```

### 9. Get Table Schema

```json
{
  "action": "get_schema",
  "project_id": "proj_1725360000",
  "table": "users"
}
```

### 10. Drop Table

```json
{
  "action": "drop_table",
  "project_id": "proj_1725360000",
  "table": "old_table"
}
```

## CRUD Operations

### 11. Insert Record

```json
{
  "action": "insert",
  "project_id": "proj_1725360000",
  "table": "users",
  "data": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }
}
```

### 12. Select Records (basic)

```json
{
  "action": "select",
  "project_id": "proj_1725360000",
  "table": "users"
}
```

### 13. Select Records (with conditions)

```json
{
  "action": "select",
  "project_id": "proj_1725360000",
  "table": "users",
  "columns": ["id", "name", "email"],
  "where": "age > ?",
  "where_args": [25],
  "order_by": "name ASC",
  "limit": 10,
  "offset": 0
}
```

### 14. Update Records

```json
{
  "action": "update",
  "project_id": "proj_1725360000",
  "table": "users",
  "data": {
    "age": 31
  },
  "where": "email = ?",
  "where_args": ["john@example.com"]
}
```

### 15. Delete Records

```json
{
  "action": "delete",
  "project_id": "proj_1725360000",
  "table": "users",
  "where": "age < ?",
  "where_args": [18]
}
```

### 16. Count Records

```json
{
  "action": "count",
  "project_id": "proj_1725360000",
  "table": "users",
  "where": "age >= ?",
  "where_args": [18]
}
```

## Join Operations

### 17. Simple Join

```json
{
  "action": "join",
  "project_id": "proj_1725360000",
  "tables": ["users", "orders"],
  "on": "users.id = orders.user_id",
  "join_type": "INNER",
  "columns": ["users.name", "users.email", "orders.total", "orders.created_at"],
  "where": "orders.total > ?",
  "where_args": [100],
  "order_by": "orders.created_at DESC"
}
```

### 18. Select with Multiple Joins

```json
{
  "action": "select_join",
  "project_id": "proj_1725360000",
  "table": "users",
  "joins": [
    {
      "type": "INNER",
      "table": "orders",
      "condition": "users.id = orders.user_id"
    },
    {
      "type": "LEFT",
      "table": "order_items",
      "condition": "orders.id = order_items.order_id"
    }
  ],
  "columns": ["users.name", "orders.total", "order_items.quantity"],
  "where": "orders.status = ?",
  "where_args": ["completed"],
  "group_by": "users.id",
  "having": "COUNT(order_items.id) > 1",
  "order_by": "orders.total DESC",
  "limit": 20
}
```

### 19. Count with Joins

```json
{
  "action": "count_join",
  "project_id": "proj_1725360000",
  "table": "users",
  "joins": [
    {
      "type": "INNER",
      "table": "orders",
      "condition": "users.id = orders.user_id"
    }
  ],
  "where": "orders.status = ?",
  "where_args": ["completed"]
}
```

### 20. Query Builder (Complex Query)

```json
{
  "action": "query_builder",
  "project_id": "proj_1725360000",
  "table": "products",
  "columns": ["category", "AVG(price) as avg_price", "COUNT(*) as product_count"],
  "joins": [
    {
      "type": "LEFT",
      "table": "product_reviews",
      "condition": "products.id = product_reviews.product_id"
    }
  ],
  "where": "products.in_stock = ?",
  "where_args": [true],
  "group_by": "category",
  "having": "COUNT(*) > 5",
  "order_by": "avg_price DESC"
}
```

## Response Format

All API responses follow this format:

```json
{
  "success": true,
  "data": { /* response data */ },
  "count": 10,
  "id": 123,
  "query": "SELECT * FROM users WHERE age > 25"
}
```

Or for errors:

```json
{
  "success": false,
  "error": "Error message here"
}
```

## Testing with curl

Here's an example curl command to test the API:

```bash
curl -X POST http://localhost:8080/api/db \
  -H "Content-Type: application/json" \
  -H "Cookie: your-auth-token=base64-encoded-token" \
  -d '{
    "action": "create_table",
    "project_id": "proj_1725360000",
    "table": "test_table",
    "data": {
      "name": "Test",
      "value": 123
    }
  }'
```

## Notes

- All database operations (except project management) require a valid `project_id`
- Authentication is required via JWT token in cookies
- The server automatically creates database connections for each project
- Schema can be explicitly defined or inferred from sample data
- All SQL queries support parameterized arguments to prevent injection attacks
