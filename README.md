# kai

A minimal Go HTTP framework with routing, middleware, and a small context API.

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

## Context helpers

- `Param(key)` for path params.
- `Query(key)` and `QueryDefault(key, fallback)`.
- `BodyBytes()` and `BodyString()`.
- `JSON(code, obj)`, `String(code, message)`, `Status(code)`.
- `Set(key, value)` / `Get(key)` for request-scoped data.
- `ServeFile(path)`, `Redirect(code, location)`.

## Example routes

See the example handlers in [cmd/example/test_routes.go](cmd/example/test_routes.go).

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
