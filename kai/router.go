package kai

import (
	"net/http"
	"strings"

	"github.com/dipto-kainin/kai-rest-server/kai/utils"
)

type Router struct {
	routes           map[string][]routeEntry // method => list of routes
	globalMiddleware []HandlerFunc
	NotFoundHandler  HandlerFunc
}

type routeEntry struct {
	pattern  string
	segments []segment
	handlers []HandlerFunc
}

type segment struct {
	literal   string
	isParam   bool
	paramName string
}

func NewRouter() *Router {
	r := &Router{}
	r.routes = make(map[string][]routeEntry)
	r.globalMiddleware = []HandlerFunc{}
	r.NotFoundHandler = default404Handler
	return r
}

func default404Handler(c *Context) {
	c.JSON(http.StatusNotFound, map[string]any{
		"error": "route not found",
	})
}

// ---------------------------
// Middleware Registration
// ---------------------------

func (r *Router) Use(mw ...HandlerFunc) {
	r.globalMiddleware = append(r.globalMiddleware, mw...)
}

// ---------------------------
// Route Registration
// ---------------------------

func (r *Router) GET(pattern string, handlers ...HandlerFunc) {
	r.addRoute("GET", pattern, handlers)
}

func (r *Router) POST(pattern string, handlers ...HandlerFunc) {
	r.addRoute("POST", pattern, handlers)
}

func (r *Router) PUT(pattern string, handlers ...HandlerFunc) {
	r.addRoute("PUT", pattern, handlers)
}

func (r *Router) DELETE(pattern string, handlers ...HandlerFunc) {
	r.addRoute("DELETE", pattern, handlers)
}

func (r *Router) addRoute(method string, pattern string, handlers []HandlerFunc) {
	segments := parsePattern(pattern)

	entry := routeEntry{
		pattern:  pattern,
		segments: segments,
		handlers: handlers,
	}

	r.routes[method] = append(r.routes[method], entry)
}

// ---------------------------
// Pattern Parsing
// ---------------------------

func parsePattern(pattern string) []segment {
	segments := []segment{}

	parts := strings.Split(pattern, "/")

	for _, part := range parts {
		if part == "" {
			continue
		}

		var seg segment

		if part[0] == ':' {
			seg.isParam = true
			seg.paramName = part[1:]
		} else {
			seg.literal = part
			seg.isParam = false
		}

		segments = append(segments, seg)
	}

	return segments
}

// ---------------------------
// Route Matching
// ---------------------------

func (r *Router) findRoute(method string, path string) (routeEntry, map[string]string, bool) {
	reqSegments := utils.SplitPath(path)

	entries, ok := r.routes[method]
	if !ok {
		return routeEntry{}, nil, false
	}

	for _, entry := range entries {
		if len(entry.segments) != len(reqSegments) {
			continue
		}

		params := make(map[string]string)
		matched := true

		for i, seg := range entry.segments {
			reqSeg := reqSegments[i]

			if seg.isParam {
				params[seg.paramName] = reqSeg
			} else if seg.literal != reqSeg {
				matched = false
				break
			}
		}

		if matched {
			return entry, params, true
		}
	}

	return routeEntry{}, nil, false
}

// ---------------------------
// HTTP Handling
// ---------------------------

func (r *Router) HandlerHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := NewContext(w, req)

	entry, params, found := r.findRoute(req.Method, req.URL.Path)

	if !found {
		// Run dynamic 404 handler
		ctx.Handlers = append(r.globalMiddleware, r.NotFoundHandler)
		ctx.Next()
		return
	}

	ctx.Path = req.URL.Path
	ctx.Method = req.Method
	ctx.Params = params
	ctx.Route = entry.pattern

	// merge global + route handlers
	ctx.Handlers = append(r.globalMiddleware, entry.handlers...)

	ctx.Next()
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.HandlerHTTP(w, req)
}
