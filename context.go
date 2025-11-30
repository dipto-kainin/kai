/*
   notes :
       1. Wrap the Request & Response Store original http.Request Store original http.ResponseWriter Allow easy access for the user

       2. Manage path parameters Example route: /users/:id Extracted params: { "id": "32" } Exposed via ctx.Param("id")

       3. Manage query parameters Parse ?page=2&search=a Store internally on first access Exposed via ctx.Query("page")

       4. Read request body Read full body once Cache it so multiple middlewares can access it

           Provide:
               BodyBytes()
               BodyString() (later) JSON binding

       5. Response helpers JSON response Text response Raw bytes Status code control

       6. Middleware data store A map to save temporary data:
           ctx.Set("user", userObj)
           ctx.Get("user")

       7. Abort mechanism Allow middleware to stop the chain E.g., in auth middleware:
           ctx.Abort()
           ctx.JSON(401, {"error":"unauthorized"})
       Chain must stop executing further handlers

       8. Error collection Middleware/handlers can push errors Error-handling middleware processes them at the end

       9. Track whether response is already written Prevent double writes

   Track:
       status written
       header written
       body written

*/

package kai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

/*
    Input: ResponseWriter, Request
Output: Context object

Steps:
1. Create a new Context struct.
2. Store Request and ResponseWriter.
3. Initialize:
      params map
      query map
      store map
      errors list
      status code = 200 (default)
      wroteHeader = false
      aborted = false
4. Return Context

*/

type HandlerFunc func(*Context)

type Context struct {
    // Core
    Writer          http.ResponseWriter
    Request         *http.Request

    // Routing
    Path            string
    Method          string
    Params          map[string]string
    Route           string // route pattern e.g. "/users/:id" set by router

    // Middleware
    MiddlewareIndex int
    Handlers        []HandlerFunc
    aborted         bool

    // Internal state
    StatusCode      int
    wroteHeader     bool
    wroteBody       bool

    // Body cache
    bodyBytes       []byte

    // Query cache
    queryCache      map[string][]string

    // User storage
    Keys            map[string]any

    // Errors
    Errors          []error
}

func (c *Context) AddError(err error) {
    c.Errors = append(c.Errors, err)
}

func (c *Context) Abort() {
    c.aborted = true
}

// IsAborted reports whether the context has been aborted.
func (c *Context) IsAborted() bool {
    return c.aborted
}

// Next executes the next handler in the middleware chain.
// MiddlewareIndex uses -1 as the "not started" value.
// This implementation increments before running each handler so it works with -1 init.
func (c *Context) Next() {
    for {
        c.MiddlewareIndex++
        if c.MiddlewareIndex >= len(c.Handlers) || c.aborted {
            return
        }
        handler := c.Handlers[c.MiddlewareIndex]
        handler(c)
    }
}

// FullPath returns the route pattern assigned by the router (e.g. "/users/:id").
func (c *Context) FullPath() string {
    return c.Route
}

// Reset re-initializes the context for reuse (useful when adding a sync.Pool).
func (c *Context) Reset(w http.ResponseWriter, req *http.Request) {
    c.Writer = w
    c.Request = req
    c.Path = req.URL.Path
    c.Method = req.Method

    c.Params = make(map[string]string)
    c.Route = ""

    c.MiddlewareIndex = -1
    c.Handlers = c.Handlers[:0]
    c.aborted = false

    c.StatusCode = http.StatusOK
    c.wroteHeader = false
    c.wroteBody = false

    c.bodyBytes = nil
    c.queryCache = nil

    // leave Keys nil to lazily initialize on Set
    c.Keys = make(map[string]any)

    c.Errors = c.Errors[:0]
}

// NewContext is the exported constructor used by router/http server code.
func NewContext(w http.ResponseWriter, req *http.Request) *Context {
    ctx := new(Context)
    ctx.Reset(w, req)
    return ctx
}

func (c *Context) Status(code int) {
    c.StatusCode = code
    if !c.wroteHeader {
        c.Writer.WriteHeader(code)
        c.wroteHeader = true
    } 	
}

func (c *Context) Write(data []byte) {
    if !c.wroteHeader {
        c.Status(c.StatusCode)
    }
    c.Writer.Write(data)
    c.wroteBody = true
}

func (c *Context) String(code int, message string) {
    c.Writer.Header().Set("Content-Type", "text/plain")
    c.Status(code)
    c.Write([]byte(message))
}

func (c *Context) JSON(code int, obj any) {
    jsonBytes, err := json.Marshal(obj)
    if err != nil {
        c.Status(http.StatusInternalServerError)
        c.Write([]byte(`{"error":"Internal Server Error"}`))
        c.AddError(err)
        return
    }
    c.Writer.Header().Set("Content-Type", "application/json")
    c.Status(code)
    c.Write(jsonBytes)
}

func (c *Context) AbortWithStatusJSON(code int, obj any) {
    c.Abort()
    c.JSON(code, obj)
}

func (c *Context) Param(key string) string {
    return c.Params[key]
}

func (c *Context) Query(key string) string {
    if c.queryCache == nil {
        c.queryCache = c.Request.URL.Query()
    }
    values := c.queryCache[key]
    if len(values) > 0 {
        return values[0]
    }
    return ""
}

func (c *Context) QueryDefault(key, defaultValue string) string {
    if value := c.Query(key); value != "" {
        return value
    }
    return defaultValue
}

func (c *Context) BodyBytes() ([]byte, error) {
    if c.bodyBytes == nil {
        body, err := io.ReadAll(c.Request.Body)
        if err != nil {
            c.bodyBytes = []byte{}
            return c.bodyBytes, err
        }
        c.bodyBytes = body
        c.Request.Body = io.NopCloser(bytes.NewBuffer(c.bodyBytes))
    }
    return c.bodyBytes, nil
}

func (c *Context) BodyString() (string, error) {
    body, err := c.BodyBytes()
    return string(body), err
}

func (c *Context) BindJSON(obj any) error {
    body, err := c.BodyBytes()
    if err != nil {
        return err
    }
    return json.Unmarshal(body, obj)
}

func (c *Context) Set(key string, value any) {
    if c.Keys == nil {
        c.Keys = make(map[string]any)
    }
    c.Keys[key] = value
}

func (c *Context) Get(key string) (any, bool) {
	if c.Keys == nil {
		return nil, false
	}
	value, exists := c.Keys[key]
	return value, exists
}

// ServeFile serves a single file
func (c *Context) ServeFile(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// Redirect performs an HTTP redirect
func (c *Context) Redirect(code int, location string) {
	if code < 300 || code > 308 {
		panic("invalid redirect code")
	}
	c.Writer.Header().Set("Location", location)
	c.Status(code)
}

// Header returns the request header value
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}
