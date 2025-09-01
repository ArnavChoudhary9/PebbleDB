# Pebbledb Sequential API Examples

This document provides a progressive workflow of API usage examples.
You can run them **step by step** to see how a complete database lifecycle works.

---

## 1. Create Users Table

```json
POST /api/db
{
  "action": "create_table",
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

---

## 2. Create Posts Table

```json
POST /api/db
{
  "action": "create_table",
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

---

## 3. Insert a User

```json
POST /api/db
{
  "action": "insert",
  "table": "users",
  "data": {
    "name": "Alice",
    "email": "alice@example.com",
    "age": 28
  }
}
```

---

## 4. Insert Another User

```json
POST /api/db
{
  "action": "insert",
  "table": "users",
  "data": {
    "name": "Bob",
    "email": "bob@example.com",
    "age": 34
  }
}
```

---

## 5. Insert a Post for Alice

```json
POST /api/db
{
  "action": "insert",
  "table": "posts",
  "data": {
    "user_id": 1,
    "title": "My First Post",
    "content": "Hello world from Alice!"
  }
}
```

---

## 6. Insert a Post for Bob

```json
POST /api/db
{
  "action": "insert",
  "table": "posts",
  "data": {
    "user_id": 2,
    "title": "Bob’s Thoughts",
    "content": "Pebbledb is awesome!"
  }
}
```

---

## 7. Select All Users

```json
POST /api/db
{
  "action": "select",
  "table": "users",
  "columns": ["id", "name", "email", "age"]
}
```

---

## 8. Select Posts with User Info (JOIN)

```json
POST /api/db
{
  "action": "join",
  "tables": ["posts", "users"],
  "on": "posts.user_id = users.id",
  "columns": ["posts.id", "posts.title", "users.name"]
}
```

---

## 9. Update Alice’s Age

```json
POST /api/db
{
  "action": "update",
  "table": "users",
  "data": {"age": 29},
  "where": "id = ?",
  "where_args": [1]
}
```

---

## 10. Count Posts by Bob

```json
POST /api/db
{
  "action": "count",
  "table": "posts",
  "where": "user_id = ?",
  "where_args": [2]
}
```

---

## 11. Paginate Posts

```json
POST /api/db
{
  "action": "select",
  "table": "posts",
  "columns": ["id", "title"],
  "order_by": "created_at DESC",
  "limit": 1,
  "offset": 0
}
```

---

## 12. Add Comments Table

```json
POST /api/db
{
  "action": "create_table",
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

---

## 13. Insert Comment on Alice’s Post

```json
POST /api/db
{
  "action": "insert",
  "table": "comments",
  "data": {
    "post_id": 1,
    "author": "Bob",
    "content": "Nice post Alice!"
  }
}
```

---

## 14. Select Comments with Post Titles

```json
POST /api/db
{
  "action": "join",
  "tables": ["comments", "posts"],
  "on": "comments.post_id = posts.id",
  "columns": ["comments.author", "comments.content", "posts.title"]
}
```

---

## 15. Delete Bob’s Post

```json
POST /api/db
{
  "action": "delete",
  "table": "posts",
  "where": "id = ?",
  "where_args": [2]
}
```

---

## 16. Select Remaining Posts

```json
POST /api/db
{
  "action": "select",
  "table": "posts",
  "columns": ["id", "title", "content"]
}
```

---

## 17. Drop Comments Table

```json
POST /api/db
{
  "action": "drop_table",
  "table": "comments"
}
```

---

## 18. Add “last\_login” Column to Users

```json
POST /api/db
{
  "action": "alter_table",
  "table": "users",
  "add_column": {
    "last_login": {"type": "DATETIME"}
  }
}
```

---

## 19. Update Alice’s Last Login

```json
POST /api/db
{
  "action": "update",
  "table": "users",
  "data": {"last_login": "2025-09-02T10:00:00Z"},
  "where": "id = ?",
  "where_args": [1]
}
```

---

## 20. Export Users Table

```json
POST /api/db
{
  "action": "export",
  "table": "users",
  "format": "json"
}
```

---

✅ By following these 20 steps in sequence, you’ll simulate a **full project lifecycle**:

* Creating tables
* Inserting records
* Querying and joining data
* Updating and deleting
* Altering schema
* Exporting data
