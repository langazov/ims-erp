package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// CompressionMiddleware compresses HTTP responses using gzip
type CompressionMiddleware struct {
	level   int
	minSize int
	pool    sync.Pool
}

// NewCompressionMiddleware creates a new compression middleware
func NewCompressionMiddleware(level int) *CompressionMiddleware {
	return &CompressionMiddleware{
		level:   level,
		minSize: 1024, // Only compress responses > 1KB
		pool: sync.Pool{
			New: func() interface{} {
				w, _ := gzip.NewWriterLevel(io.Discard, level)
				return w
			},
		},
	}
}

// Handler returns the middleware handler
func (m *CompressionMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for small responses or already compressed content
		if r.Header.Get("Content-Encoding") != "" {
			next.ServeHTTP(w, r)
			return
		}

		// Wrap response writer
		cw := &compressResponseWriter{
			ResponseWriter: w,
			middleware:     m,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(cw, r)
	})
}

// compressResponseWriter wraps http.ResponseWriter to compress output
type compressResponseWriter struct {
	http.ResponseWriter
	middleware *CompressionMiddleware
	writer     *gzip.Writer
	statusCode int
	written    bool
}

func (w *compressResponseWriter) WriteHeader(code int) {
	w.statusCode = code
}

func (w *compressResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true

		// Only compress if response is large enough and successful
		if len(b) >= w.middleware.minSize && w.statusCode < 300 {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")
			w.ResponseWriter.WriteHeader(w.statusCode)

			// Get writer from pool
			w.writer = w.middleware.pool.Get().(*gzip.Writer)
			w.writer.Reset(w.ResponseWriter)

			return w.writer.Write(b)
		}

		w.ResponseWriter.WriteHeader(w.statusCode)
	}

	return w.ResponseWriter.Write(b)
}

func (w *compressResponseWriter) Close() {
	if w.writer != nil {
		w.writer.Close()
		w.middleware.pool.Put(w.writer)
	}
}

// RequestCoalescingMiddleware deduplicates concurrent identical requests
type RequestCoalescingMiddleware struct {
	mu      sync.RWMutex
	pending map[string]*pendingRequest
}

type pendingRequest struct {
	response []byte
	err      error
	done     chan struct{}
	refCount int
}

// NewRequestCoalescingMiddleware creates a new request coalescing middleware
func NewRequestCoalescingMiddleware() *RequestCoalescingMiddleware {
	return &RequestCoalescingMiddleware{
		pending: make(map[string]*pendingRequest),
	}
}

// Handler returns the middleware handler
func (m *RequestCoalescingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only coalesce GET requests
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Create key from request
		key := fmt.Sprintf("%s:%s:%s", r.Method, r.URL.Path, r.URL.RawQuery)

		m.mu.Lock()
		if pending, exists := m.pending[key]; exists {
			// Wait for existing request
			pending.refCount++
			m.mu.Unlock()

			<-pending.done

			if pending.err != nil {
				http.Error(w, pending.err.Error(), http.StatusInternalServerError)
				return
			}

			w.Write(pending.response)
			return
		}

		// Create new pending request
		pending := &pendingRequest{
			done:     make(chan struct{}),
			refCount: 1,
		}
		m.pending[key] = pending
		m.mu.Unlock()

		// Wrap response writer to capture output
		rw := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		// Store result
		pending.response = []byte(rw.body.String())
		close(pending.done)

		// Clean up
		m.mu.Lock()
		delete(m.pending, key)
		m.mu.Unlock()
	})
}

// responseRecorder captures response for coalescing
type responseRecorder struct {
	http.ResponseWriter
	body   strings.Builder
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
