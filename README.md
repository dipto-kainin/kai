# kai

Kai is a tiny Go HTTP framework for building clean REST APIs.
It offers method-based routing with path params and simple route groups.
Middleware chaining is first-class with `Next()`/`Abort()` control flow.
The context API keeps handlers short with helpers for JSON, text, files, and redirects.
Built-in utilities cover logging, recovery, CORS, request IDs, timeouts, and rate limits.

## Features

- Method-based routing with path params (e.g. `/users/:id`).
- Global and per-route middleware with `Next()` and `Abort()`.
- Context helpers for JSON, text, status, headers, and redirects.
- Query parsing and request body caching.
- File upload helpers and simple file serving.
- Built-in middleware: logger, panic recovery, CORS, request ID, timeout, rate limit, secure headers, gzip.

## Install

```bash
go get github.com/dipto-kainin/kai
```

## Quick start

```go
package main

import (
    "github.com/dipto-kainin/kai"
)

func main() {
    app := kai.NewApp()

    // Global middleware
    app.Use(kai.Logger(), kai.DamageControl())

    app.GET("/hello", func(c *kai.Context) {
        c.JSON(200, map[string]string{
            "message": "Hello, World!",
        })
    })

    _ = app.Play(8000)
}
```

Run it:

```bash
go run ./cmd
```

## Routes and groups

```go
app := kai.NewApp()

api := app.Group("/api")
api.GET("/users/:id", func(c *kai.Context) {
    id := c.Param("id")
    c.JSON(200, map[string]string{"id": id})
})
```

## Middleware

Middleware can call `c.Next()` to continue or `c.Abort()` to stop the chain.

```go
app.Use(kai.RequestID(), kai.Timeout(5*time.Second))
```

## Full CRUD and file example

The project now includes a fuller example in `cmd/example/crud_showcase.go`. It demonstrates:

- `GET /api/posts` to list records with optional query filters.
- `GET /api/posts/:id` to fetch one record by path param.
- `POST /api/posts` to create a record from JSON.
- `PUT /api/posts/:id` to fully update a record from JSON.
- `DELETE /api/posts/:id` to remove a record.
- `POST /api/posts/:id/file` to upload a multipart file.
- `GET /api/posts/:id/file` to download the uploaded file.
- `DELETE /api/posts/:id/file` to remove the uploaded file.

Run the example server:

```bash
go run ./cmd
```

List posts:

```bash
curl "http://localhost:8000/api/posts?limit=10&published=true"
```

Get one post:

```bash
curl http://localhost:8000/api/posts/1
```

Create a post:

```bash
curl -X POST http://localhost:8000/api/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Write docs",
    "content": "Document the framework with realistic examples.",
    "published": false
  }'
```

Update a post:

```bash
curl -X PUT http://localhost:8000/api/posts/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Ship the first endpoint",
    "content": "The example now shows full CRUD plus files.",
    "published": true
  }'
```

Delete a post:

```bash
curl -X DELETE http://localhost:8000/api/posts/1
```

Upload a file to a post:

```bash
curl -X POST http://localhost:8000/api/posts/2/file \
  -F "file=@./README.md"
```

Download that file:

```bash
curl -OJ http://localhost:8000/api/posts/2/file
```

Delete that file:

```bash
curl -X DELETE http://localhost:8000/api/posts/2/file
```

Example JSON response from `GET /api/posts/1`:

```json
{
  "id": 1,
  "title": "Ship the first endpoint",
  "content": "This seeded record helps you try GET and PUT immediately.",
  "published": true,
  "created_at": "2026-04-14T10:00:00Z",
  "updated_at": "2026-04-14T10:00:00Z"
}
```

## Context helpers

- `Param(key)` for path params.
- `Query(key)` and `QueryDefault(key, fallback)`.
- `BodyBytes()` and `BodyString()`.
- `JSON(code, obj)`, `String(code, message)`, `Status(code)`.
- `Set(key, value)` / `Get(key)` for request-scoped data.
- `GetJSON()` for simple JSON request parsing.
- `GetFileBytes(fieldName)`, `SaveToDest(dest, fieldName)` for multipart uploads.
- `ServeFile(path)`, `Redirect(code, location)`.

## Example routes

See the example handlers in [cmd/example/test_routes.go](cmd/example/test_routes.go).

## License

Kai is licensed under the MIT License. The canonical license text lives in `LICENSE`, and compatible alias filenames are included so Go license detection tools can recognize the module reliably.

## Project layout

```
.
├── app.go
├── context.go
├── middleware.go
├── router.go
├── utils/
│   ├── errors.go
│   └── path.go
└── cmd/
    ├── main.go
    └── example/
        ├── test_control.go
        ├── test_control1.go
        └── test_routes.go
```
