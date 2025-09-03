# Pebbledb Sequential API Examples

This document provides a progressive workflow of API usage examples.
You can run them **step by step** to see how a complete database lifecycle works.

---

## Project Management

### 1. Create a New Project

```json
POST /api/db
{
  "action": "create_project",
  "project_id": "my_blog_app",
  "project_name": "My Blog Application",
  "project_description": "A simple blog application with users and posts"
}
```

### 2. List All Projects

```json
POST /api/db
{
  "action": "list_projects"
}
```

### 3. Get Project Information

```json
POST /api/db
{
  "action": "get_project",
  "project_id": "my_blog_app"
}
```

### 4. Get Tables in Project (after creating some tables)

```json
POST /api/db
{
  "action": "get_tables",
  "project_id": "my_blog_app"
}
```

---

## Database Operations (within a project)

### 5. Create Users Table

```json
POST /api/db
{
  "action": "create_table",
  "project_id": "my_blog_app",
  "table": "users",
  "schema": {
    "id": {"type": "INTEGER", "constraints": ["PRIMARY KEY", "AUTOINCREMENT"]},
    "name": {"type": "TEXT", "constraints": ["NOT NULL"]},
    "email": {"type": "TEXT", "constraints": ["UNIQUE", "NOT NULL"]},
    "age": {"type": "INTEGER"},
    "created_at": {"type": "DATETIME", "constraints": ["DEFAULT CURRENT_TIMESTAMP"]}
  }
}
```

### 6. Create Posts Table

```json
POST /api/db
{
  "action": "create_table",
  "project_id": "my_blog_app",
  "table": "posts",
  "schema": {
    "id": {"type": "INTEGER", "constraints": ["PRIMARY KEY", "AUTOINCREMENT"]},
    "user_id": {"type": "INTEGER", "constraints": ["NOT NULL"]},
    "title": {"type": "TEXT", "constraints": ["NOT NULL"]},
    "content": {"type": "TEXT"},
    "created_at": {"type": "DATETIME", "constraints": ["DEFAULT CURRENT_TIMESTAMP"]}
  }
}
```

### 7. Insert a User

```json
POST /api/db
{
  "action": "insert",
  "project_id": "my_blog_app",
  "table": "users",
  "data": {
    "name": "Alice",
    "email": "alice@example.com",
    "age": 28
  }
}
```

### 8. Insert Another User

```json
POST /api/db
{
  "action": "insert",
  "project_id": "my_blog_app",
  "table": "users",
  "data": {
    "name": "Bob",
    "email": "bob@example.com",
    "age": 34
  }
}
```

### 9. Insert a Post for Alice

```json
POST /api/db
{
  "action": "insert",
  "project_id": "my_blog_app",
  "table": "posts",
  "data": {
    "user_id": 1,
    "title": "My First Post",
    "content": "Hello world from Alice!"
  }
}
```

### 10. Insert a Post for Bob

```json
POST /api/db
{
  "action": "insert",
  "project_id": "my_blog_app",
  "table": "posts",
  "data": {
    "user_id": 2,
    "title": "Bob's Thoughts",
    "content": "Pebbledb is awesome!"
  }
}
```

### 11. Select All Users

```json
POST /api/db
{
  "action": "select",
  "project_id": "my_blog_app",
  "table": "users",
  "columns": ["id", "name", "email", "age"]
}
```

### 12. Select Posts with User Info (JOIN)

```json
POST /api/db
{
  "action": "join",
  "project_id": "my_blog_app",
  "tables": ["posts", "users"],
  "on": "posts.user_id = users.id",
  "columns": ["posts.id", "posts.title", "users.name"]
}
```

### 13. Update Alice's Age

```json
POST /api/db
{
  "action": "update",
  "project_id": "my_blog_app",
  "table": "users",
  "data": {"age": 29},
  "where": "id = ?",
  "where_args": [1]
}
```

### 14. Count Posts by Bob

```json
POST /api/db
{
  "action": "count",
  "project_id": "my_blog_app",
  "table": "posts",
  "where": "user_id = ?",
  "where_args": [2]
}
```

### 15. Paginate Posts

```json
POST /api/db
{
  "action": "select",
  "project_id": "my_blog_app",
  "table": "posts",
  "columns": ["id", "title"],
  "order_by": "created_at DESC",
  "limit": 1,
  "offset": 0
}
```

### 16. Add Comments Table

```json
POST /api/db
{
  "action": "create_table",
  "project_id": "my_blog_app",
  "table": "comments",
  "schema": {
    "id": {"type": "INTEGER", "constraints": ["PRIMARY KEY", "AUTOINCREMENT"]},
    "post_id": {"type": "INTEGER", "constraints": ["NOT NULL"]},
    "author": {"type": "TEXT"},
    "content": {"type": "TEXT"},
    "created_at": {"type": "DATETIME", "constraints": ["DEFAULT CURRENT_TIMESTAMP"]}
  }
}
```

### 17. Insert Comment on Alice's Post

```json
POST /api/db
{
  "action": "insert",
  "project_id": "my_blog_app",
  "table": "comments",
  "data": {
    "post_id": 1,
    "author": "Bob",
    "content": "Nice post Alice!"
  }
}
```

### 18. Select Comments with Post Titles (Advanced Join)

```json
POST /api/db
{
  "action": "select_join",
  "project_id": "my_blog_app",
  "table": "comments",
  "joins": [
    {
      "type": "INNER",
      "table": "posts",
      "condition": "comments.post_id = posts.id"
    },
    {
      "type": "INNER", 
      "table": "users",
      "condition": "posts.user_id = users.id"
    }
  ],
  "columns": ["comments.author", "comments.content", "posts.title", "users.name as post_author"]
}
```

### 19. Get Table Schema

```json
POST /api/db
{
  "action": "get_schema",
  "project_id": "my_blog_app",
  "table": "users"
}
```

### 20. Check if Table Exists

```json
POST /api/db
{
  "action": "table_exists",
  "project_id": "my_blog_app",
  "table": "categories"
}
```

### 21. Delete Bob's Post

```json
POST /api/db
{
  "action": "delete",
  "project_id": "my_blog_app",
  "table": "posts",
  "where": "id = ?",
  "where_args": [2]
}
```

### 22. Select Remaining Posts

```json
POST /api/db
{
  "action": "select",
  "project_id": "my_blog_app",
  "table": "posts",
  "columns": ["id", "title", "content"]
}
```

### 23. Drop Comments Table

```json
POST /api/db
{
  "action": "drop_table",
  "project_id": "my_blog_app",
  "table": "comments"
}
```

### 24. Get Updated Tables List

```json
POST /api/db
{
  "action": "get_tables",
  "project_id": "my_blog_app"
}
```

---

## Project Cleanup

### 25. Delete Project (Clean Up)

```json
POST /api/db
{
  "action": "delete_project",
  "project_id": "my_blog_app"
}
```

### 26. Verify Project Deletion

```json
POST /api/db
{
  "action": "list_projects"
}
```

---

## Advanced Examples

### Query Builder Example

```json
POST /api/db
{
  "action": "query_builder",
  "project_id": "my_blog_app",
  "table": "posts",
  "columns": ["posts.title", "users.name", "COUNT(comments.id) as comment_count"],
  "joins": [
    {
      "type": "INNER",
      "table": "users", 
      "condition": "posts.user_id = users.id"
    },
    {
      "type": "LEFT",
      "table": "comments",
      "condition": "posts.id = comments.post_id"
    }
  ],
  "group_by": "posts.id, users.name",
  "having": "COUNT(comments.id) > 0",
  "order_by": "comment_count DESC",
  "limit": 10
}
```

### Complex Select with Multiple Joins

```json
POST /api/db
{
  "action": "select_join",
  "project_id": "my_blog_app", 
  "table": "posts",
  "joins": [
    {
      "type": "INNER",
      "table": "users",
      "condition": "posts.user_id = users.id"
    }
  ],
  "columns": ["posts.*", "users.name as author_name"],
  "where": "users.age > ?",
  "where_args": [25],
  "order_by": "posts.created_at DESC"
}
```

---

✅ By following these steps in sequence, you'll experience a **complete project and database lifecycle**:

* **Project Management**: Creating, listing, and managing database projects
* **Schema Design**: Creating tables with proper relationships
* **Data Operations**: Inserting, updating, deleting records
* **Complex Queries**: Joins, aggregations, and advanced filtering
* **Maintenance**: Schema inspection and table management
* **Cleanup**: Project deletion and resource management

Each example builds upon the previous ones, demonstrating real-world usage patterns for a blog application with users, posts, and comments.
