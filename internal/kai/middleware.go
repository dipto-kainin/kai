package kai

import (
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

func DamageControl() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if r := recover(); r != nil {
				c.Status(500)
				c.JSON(500, map[string]any{
					"error": "Internal Server Error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
func Logger() HandlerFunc {
	return func(c *Context) {
		startTime := time.Now()
		c.Next()
		duration := time.Since(startTime)
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.StatusCode

		fmt.Println("[Kai Logger]", "Method:", method, "Path:", path, "Status:", statusCode, "Duration:", duration)
	}
}

type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposeHeaders    []string
	AllowCredentials bool
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
func Pain_of_CORS(opts ...CORSOptions) HandlerFunc {
    var cfg CORSOptions

    // Defaults
    if len(opts) == 0 {
        cfg = CORSOptions{
            AllowedOrigins:   []string{"*"},
            AllowedMethods:   []string{"*"},
            AllowedHeaders:   []string{"*"},
            AllowCredentials: true,
        }
    } else {
        cfg = opts[0]

        if len(cfg.AllowedOrigins) == 0 {
            panic("Kai CORS: AllowedOrigins cannot be empty")
        }
        if len(cfg.AllowedMethods) == 0 {
            cfg.AllowedMethods = []string{"*"}
        }
        if len(cfg.AllowedHeaders) == 0 {
            cfg.AllowedHeaders = []string{"*"}
        }

        // SPEC: Wildcard origin + credentials is NOT allowed
        if cfg.AllowCredentials && contains(cfg.AllowedOrigins, "*") {
            panic("Kai CORS: wildcard origin cannot be used when AllowCredentials=true")
        }
    }

    allowedMethodsHeader :=
        strings.Join(cfg.AllowedMethods, ", ")

    return func(c *Context) {
        origin := c.Request.Header.Get("Origin")
        reqMethod := c.Request.Method
        preflightMethod := c.Request.Header.Get("Access-Control-Request-Method")
        preflightHeaders := c.Request.Header.Get("Access-Control-Request-Headers")

        // --- Add Vary headers (required) ---
        h := c.Writer.Header()
        h.Add("Vary", "Origin")
        h.Add("Vary", "Access-Control-Request-Method")
        h.Add("Vary", "Access-Control-Request-Headers")

        // 1. Origin validation (with wildcard domain support)
        originAllowed := false
        if contains(cfg.AllowedOrigins, "*") {
            originAllowed = true
        } else if contains(cfg.AllowedOrigins, origin) {
            originAllowed = true
        } else {
            // Check wildcard patterns like "*.example.com"
            for _, allowed := range cfg.AllowedOrigins {
                if strings.HasPrefix(allowed, "*.") {
                    domain := allowed[2:]
                    if strings.HasSuffix(origin, domain) || strings.HasSuffix(origin, "."+domain) {
                        originAllowed = true
                        break
                    }
                }
            }
        }
        if !originAllowed {
            c.Status(403)
            c.JSON(403, map[string]string{"error": "CORS origin not allowed"})
            c.Abort()
            return
        }

        // 2. Method validation (preflight)
        if reqMethod == "OPTIONS" && preflightMethod != "" {
            if !contains(cfg.AllowedMethods, "*") &&
               !contains(cfg.AllowedMethods, preflightMethod) {
                c.Status(403)
                c.JSON(403, map[string]string{"error": "CORS method not allowed"})
                c.Abort()
                return
            }
        }

        // 2b. Method validation (normal requests)
        if reqMethod != "OPTIONS" &&
           !contains(cfg.AllowedMethods, "*") &&
           !contains(cfg.AllowedMethods, reqMethod) {
            c.Status(405)
            c.JSON(405, map[string]string{"error": "method not allowed"})
            c.Abort()
            return
        }

        // 3. Header validation
        if reqMethod == "OPTIONS" &&
           preflightHeaders != "" &&
           !contains(cfg.AllowedHeaders, "*") {

            for _, hdr := range strings.Split(preflightHeaders, ",") {
                hdr = strings.TrimSpace(strings.ToLower(hdr))
                hdrAllowed := false

                for _, allowed := range cfg.AllowedHeaders {
                    if strings.ToLower(allowed) == hdr {
                        hdrAllowed = true
                        break
                    }
                }

                if !hdrAllowed {
                    c.Status(403)
                    c.JSON(403, map[string]string{"error": "CORS header not allowed: " + hdr})
                    c.Abort()
                    return
                }
            }
        }

        // 4. Set Allow-Origin
        if contains(cfg.AllowedOrigins, "*") {
            h.Set("Access-Control-Allow-Origin", "*")
        } else {
            h.Set("Access-Control-Allow-Origin", origin)
        }

        // Allow-Methods
        if contains(cfg.AllowedMethods, "*") {
            h.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
        } else {
            h.Set("Access-Control-Allow-Methods", allowedMethodsHeader)
        }

        // Allow-Headers
        if contains(cfg.AllowedHeaders, "*") {
            if preflightHeaders != "" {
                h.Set("Access-Control-Allow-Headers", preflightHeaders)
            } else {
                h.Set("Access-Control-Allow-Headers",
                    "Content-Type, Authorization, X-Requested-With, Accept, Origin")
            }
        } else {
            h.Set("Access-Control-Allow-Headers",
                strings.Join(cfg.AllowedHeaders, ", "))
        }

        if cfg.AllowCredentials {
            h.Set("Access-Control-Allow-Credentials", "true")
        }

        // Expose-Headers
        if len(cfg.ExposeHeaders) > 0 {
            h.Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
        }

        // Preflight end
        if reqMethod == "OPTIONS" {
            h.Set("Access-Control-Max-Age", "86400")
            c.Status(204)
            c.Abort()
            return
        }

        c.Next()
    }
}

func RequestID() HandlerFunc {
	return func(c *Context) {
		id := make([]byte, 16)
		_, err := rand.Read(id)
		if err != nil {
			id = []byte(fmt.Sprintf("%d", time.Now().UnixNano()))
		}
		rid := hex.EncodeToString(id)
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Set("RequestID", rid)
		c.Next()
	}
}
// Timeout enforces a timeout for the request handler chain.
// When timeout occurs, sends 504 Gateway Timeout and aborts the chain.
// 
// IMPORTANT: Handlers should check c.Request.Context().Err() to detect cancellation
// and stop long-running operations. Example:
//
//   if c.Request.Context().Err() != nil {
//       return // context cancelled or timed out
//   }
func Timeout(d time.Duration) HandlerFunc {
	return func(c *Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-ctx.Done():
			// If response already started, don't attempt to write timeout
			if c.wroteHeader {
				c.Abort()
				return
			}
			c.Status(504)
			c.JSON(504, map[string]any{"error": "timeout"})
			c.Abort()
			return
		case <-done:
			return
		}
	}
}
type rateBucket struct {
	mu          sync.Mutex
	windowStart time.Time
	count       int
}

var rateStore sync.Map
var rateLimitCleanupOnce sync.Once

func RateLimit(max int, per time.Duration) HandlerFunc {
	// Start cleanup goroutine once
	rateLimitCleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				now := time.Now()
				rateStore.Range(func(key, value any) bool {
					b := value.(*rateBucket)
					b.mu.Lock()
					if now.Sub(b.windowStart) > 24*time.Hour {
						rateStore.Delete(key)
					}
					b.mu.Unlock()
					return true
				})
			}
		}()
	})

	return func(c *Context) {
		ip := clientIP(c.Request)
		val, _ := rateStore.LoadOrStore(ip, &rateBucket{windowStart: time.Now(), count: 0})
		b := val.(*rateBucket)

		b.mu.Lock()
		now := time.Now()
		if now.Sub(b.windowStart) > per {
			b.windowStart = now
			b.count = 0
		}
		b.count++
		count := b.count
		b.mu.Unlock()

		if count > max {
			c.Status(429)
			c.JSON(429, map[string]any{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func clientIP(r *http.Request) string {
	if h := r.Header.Get("X-Forwarded-For"); h != "" {
		parts := strings.Split(h, ",")
		return strings.TrimSpace(parts[0])
	}
	if h := r.Header.Get("X-Real-IP"); h != "" {
		return strings.TrimSpace(h)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
func SecureHeaders() HandlerFunc {
	return func(c *Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-XSS-Protection", "1; mode=block")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "geolocation=(), microphone=()")
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'")
		c.Next()
	}
}
func GZip(level int) HandlerFunc {
	return func(c *Context) {
		ae := c.Request.Header.Get("Accept-Encoding")
		if !strings.Contains(ae, "gzip") {
			c.Next()
			return
		}
		// Initialize gzip writer with the actual ResponseWriter (not nil)
		gw, err := gzip.NewWriterLevel(c.Writer, level)
		if err != nil {
			c.Next()
			return
		}

		grw := &gzipResponseWriter{
			ResponseWriter: c.Writer,
			gz:             gw,
			wroteHeader:    false,
		}
		c.Writer = grw
		defer func() {
			_ = grw.Close()
		}()

		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Del("Content-Length")

		c.Next()
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz          *gzip.Writer
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if !w.wroteHeader {
		w.ResponseWriter.WriteHeader(status)
		w.wroteHeader = true
	}
}

func (w *gzipResponseWriter) Write(p []byte) (int, error) {
	if w.gz == nil {
		w.gz = gzip.NewWriter(w.ResponseWriter)
	}
	return w.gz.Write(p)
}

// Flush flushes the gzip writer and underlying ResponseWriter if it supports flushing
func (w *gzipResponseWriter) Flush() {
	if w.gz != nil {
		w.gz.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *gzipResponseWriter) Close() error {
	if w.gz != nil {
		return w.gz.Close()
	}
	return nil
}
func BodyLimit(maxBytes int64) HandlerFunc {
	return func(c *Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}