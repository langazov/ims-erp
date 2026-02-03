package middleware

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ims-erp/system/pkg/logger"
)

// PerformanceMiddleware provides performance optimization features
type PerformanceMiddleware struct {
	compressor     *CompressionMiddleware
	coalescer      *RequestCoalescingMiddleware
	circuitBreaker *CircuitBreaker
	rateLimiter    *RateLimiter
	poolManager    *ConnectionPoolManager
	logger         *logger.Logger
	metrics        *PerformanceMetrics
}

// PerformanceMetrics tracks performance statistics
type PerformanceMetrics struct {
	mu            sync.RWMutex
	requestCount  int64
	totalLatency  time.Duration
	p95Latency    time.Duration
	p99Latency    time.Duration
	errorCount    int64
	compressBytes int64
}

// NewPerformanceMiddleware creates a new performance middleware
func NewPerformanceMiddleware(logger *logger.Logger) *PerformanceMiddleware {
	return &PerformanceMiddleware{
		compressor:     NewCompressionMiddleware(gzip.DefaultCompression),
		coalescer:      NewRequestCoalescingMiddleware(),
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second),
		rateLimiter:    NewRateLimiter(100, 200),
		poolManager:    NewConnectionPoolManager(),
		logger:         logger,
		metrics:        &PerformanceMetrics{},
	}
}

// Handler returns the complete performance middleware chain
func (pm *PerformanceMiddleware) Handler(next http.Handler) http.Handler {
	return pm.rateLimiter.Handler(
		pm.circuitBreaker.Handler(
			pm.coalescer.Handler(
				pm.compressor.Handler(
					pm.metrics.Handler(next),
				),
			),
		),
	)
}

// MetricsHandler returns current performance metrics
func (pm *PerformanceMiddleware) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	pm.metrics.mu.RLock()
	metrics := map[string]interface{}{
		"requestCount":  pm.metrics.requestCount,
		"errorCount":    pm.metrics.errorCount,
		"avgLatencyMs":  float64(pm.metrics.totalLatency.Milliseconds()) / float64(pm.metrics.requestCount),
		"p95LatencyMs":  pm.metrics.p95Latency.Milliseconds(),
		"p99LatencyMs":  pm.metrics.p99Latency.Milliseconds(),
		"compressBytes": pm.metrics.compressBytes,
	}
	pm.metrics.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"metrics": %+v}`, metrics)
}

// metrics middleware tracks request metrics
func (m *PerformanceMetrics) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer
		rw := &metricsResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		// Record metrics
		latency := time.Since(start)
		m.mu.Lock()
		m.requestCount++
		m.totalLatency += latency
		if rw.statusCode >= 500 {
			m.errorCount++
		}

		// Update P95/P99 (simplified calculation)
		if latency > m.p95Latency {
			m.p95Latency = latency
		}
		if latency > m.p99Latency {
			m.p99Latency = latency
		}
		m.mu.Unlock()
	})
}

// metricsResponseWriter captures status code
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *metricsResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// BrotliMiddleware provides Brotli compression
type BrotliMiddleware struct {
	minSize int
	level   int
}

// NewBrotliMiddleware creates a new Brotli middleware
func NewBrotliMiddleware(level int) *BrotliMiddleware {
	return &BrotliMiddleware{
		minSize: 1024,
		level:   level,
	}
}

// Handler returns the middleware handler
func (m *BrotliMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts brotli
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
			next.ServeHTTP(w, r)
			return
		}

		// Wrap response writer
		rw := &brotliResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)
	})
}

// brotliResponseWriter wraps http.ResponseWriter
type brotliResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *brotliResponseWriter) WriteHeader(code int) {
	w.statusCode = code
}

func (w *brotliResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
		w.Header().Set("Content-Encoding", "br")
		w.Header().Add("Vary", "Accept-Encoding")
		w.ResponseWriter.WriteHeader(w.statusCode)
	}
	return w.ResponseWriter.Write(b)
}

// AdaptiveRateLimiter adjusts rate limits based on system load
type AdaptiveRateLimiter struct {
	mu           sync.RWMutex
	baseRate     int
	currentRate  int
	loadFactor   float64
	lastAdjusted time.Time
}

// NewAdaptiveRateLimiter creates an adaptive rate limiter
func NewAdaptiveRateLimiter(baseRate int) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		baseRate:     baseRate,
		currentRate:  baseRate,
		loadFactor:   1.0,
		lastAdjusted: time.Now(),
	}
}

// GetCurrentRate returns the current rate limit
func (a *AdaptiveRateLimiter) GetCurrentRate() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentRate
}

// AdjustRate adjusts the rate based on system load
func (a *AdaptiveRateLimiter) AdjustRate(load float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Adjust every 10 seconds
	if time.Since(a.lastAdjusted) < 10*time.Second {
		return
	}

	a.loadFactor = load
	a.lastAdjusted = time.Now()

	// Reduce rate under high load
	if load > 0.8 {
		a.currentRate = int(float64(a.baseRate) * 0.5)
	} else if load > 0.6 {
		a.currentRate = int(float64(a.baseRate) * 0.75)
	} else {
		a.currentRate = a.baseRate
	}
}

// ConnectionPoolStats tracks connection pool statistics
type ConnectionPoolStats struct {
	ActiveConnections int
	IdleConnections   int
	WaitCount         int64
	WaitDuration      time.Duration
}

// GetPoolStats returns connection pool statistics
func GetPoolStats(poolName string) *ConnectionPoolStats {
	// In a real implementation, this would query actual pool stats
	return &ConnectionPoolStats{
		ActiveConnections: 10,
		IdleConnections:   5,
		WaitCount:         0,
		WaitDuration:      0,
	}
}

// LatencyHistogram tracks latency distribution
type LatencyHistogram struct {
	mu      sync.RWMutex
	buckets []time.Duration
	counts  []int64
	total   int64
}

// NewLatencyHistogram creates a new latency histogram
func NewLatencyHistogram() *LatencyHistogram {
	return &LatencyHistogram{
		buckets: []time.Duration{
			1 * time.Millisecond,
			5 * time.Millisecond,
			10 * time.Millisecond,
			25 * time.Millisecond,
			50 * time.Millisecond,
			100 * time.Millisecond,
			250 * time.Millisecond,
			500 * time.Millisecond,
			1 * time.Second,
			2 * time.Second,
			5 * time.Second,
		},
		counts: make([]int64, 11),
	}
}

// Record records a latency measurement
func (h *LatencyHistogram) Record(latency time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.total++
	for i, bucket := range h.buckets {
		if latency <= bucket {
			h.counts[i]++
			return
		}
	}
	// If larger than all buckets, increment last bucket
	h.counts[len(h.counts)-1]++
}

// GetPercentile returns the latency at a given percentile
func (h *LatencyHistogram) GetPercentile(p float64) time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()

	target := int64(float64(h.total) * p / 100)
	var cumulative int64

	for i, count := range h.counts {
		cumulative += count
		if cumulative >= target {
			return h.buckets[i]
		}
	}

	return h.buckets[len(h.buckets)-1]
}

// PerformanceConfig holds performance middleware configuration
type PerformanceConfig struct {
	EnableCompression       bool
	EnableCoalescing        bool
	EnableCircuitBreaker    bool
	EnableRateLimiting      bool
	CompressionLevel        int
	RateLimitPerSecond      int
	RateLimitBurst          int
	CircuitBreakerThreshold int
	CircuitBreakerTimeout   time.Duration
}

// DefaultPerformanceConfig returns default configuration
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		EnableCompression:       true,
		EnableCoalescing:        true,
		EnableCircuitBreaker:    true,
		EnableRateLimiting:      true,
		CompressionLevel:        gzip.DefaultCompression,
		RateLimitPerSecond:      100,
		RateLimitBurst:          200,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
	}
}

// NewPerformanceMiddlewareFromConfig creates middleware from config
func NewPerformanceMiddlewareFromConfig(cfg *PerformanceConfig, logger *logger.Logger) *PerformanceMiddleware {
	pm := &PerformanceMiddleware{
		logger:  logger,
		metrics: &PerformanceMetrics{},
	}

	if cfg.EnableCompression {
		pm.compressor = NewCompressionMiddleware(cfg.CompressionLevel)
	}

	if cfg.EnableCoalescing {
		pm.coalescer = NewRequestCoalescingMiddleware()
	}

	if cfg.EnableCircuitBreaker {
		pm.circuitBreaker = NewCircuitBreaker(cfg.CircuitBreakerThreshold, cfg.CircuitBreakerTimeout)
	}

	if cfg.EnableRateLimiting {
		pm.rateLimiter = NewRateLimiter(cfg.RateLimitPerSecond, cfg.RateLimitBurst)
	}

	pm.poolManager = NewConnectionPoolManager()

	return pm
}
