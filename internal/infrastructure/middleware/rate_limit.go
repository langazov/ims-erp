package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

// tokenBucket implements a simple token bucket rate limiter
type tokenBucket struct {
	tokens   int
	lastFill time.Time
	rate     int // tokens per second
	burst    int // max tokens
	mu       sync.Mutex
}

func newTokenBucket(rate, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:   burst,
		lastFill: time.Now(),
		rate:     rate,
		burst:    burst,
	}
}

func (tb *tokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Add tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastFill).Seconds()
	tb.tokens += int(elapsed * float64(tb.rate))
	if tb.tokens > tb.burst {
		tb.tokens = tb.burst
	}
	tb.lastFill = now

	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// RateLimiter implements token bucket rate limiting per client
type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*clientLimiter
	rate     int
	burst    int
}

type clientLimiter struct {
	limiter  *tokenBucket
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(ratePerSecond, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*clientLimiter),
		rate:     ratePerSecond,
		burst:    burst,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Handler returns the middleware handler
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := rl.getClientIP(r)

		// Get or create limiter for this client
		limiter := rl.getLimiter(clientIP)

		// Check rate limit
		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts client IP from request
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Use RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// getLimiter gets or creates a rate limiter for a client
func (rl *RateLimiter) getLimiter(clientIP string) *tokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if cl, exists := rl.limiters[clientIP]; exists {
		cl.lastSeen = time.Now()
		return cl.limiter
	}

	// Create new limiter
	limiter := newTokenBucket(rl.rate, rl.burst)
	rl.limiters[clientIP] = &clientLimiter{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

// cleanup removes old limiters periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, cl := range rl.limiters {
			if time.Since(cl.lastSeen) > 5*time.Minute {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	failureCount     int
	lastFailureTime  time.Time
	threshold        int
	timeout          time.Duration
	halfOpenMaxCalls int
	halfOpenCalls    int
}

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		threshold:        threshold,
		timeout:          timeout,
		halfOpenMaxCalls: 3,
	}
}

// Handler wraps an HTTP handler with circuit breaker
func (cb *CircuitBreaker) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cb.CanExecute() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Service temporarily unavailable (circuit breaker open)",
			})
			return
		}

		// Wrap response writer to detect failures
		rw := &circuitBreakerResponseWriter{
			ResponseWriter: w,
			cb:             cb,
		}

		next.ServeHTTP(rw, r)
	})
}

// CanExecute returns true if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.halfOpenCalls = 0
			return true
		}
		return false
	case StateHalfOpen:
		if cb.halfOpenCalls < cb.halfOpenMaxCalls {
			cb.halfOpenCalls++
			return true
		}
		return false
	}

	return false
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.halfOpenCalls++
		if cb.halfOpenCalls >= cb.halfOpenMaxCalls {
			// Close the circuit
			cb.state = StateClosed
			cb.failureCount = 0
			cb.halfOpenCalls = 0
		}
	} else {
		cb.failureCount = 0
	}
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen {
		// Open the circuit again
		cb.state = StateOpen
	} else if cb.failureCount >= cb.threshold {
		// Open the circuit
		cb.state = StateOpen
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// circuitBreakerResponseWriter wraps response writer to detect failures
type circuitBreakerResponseWriter struct {
	http.ResponseWriter
	cb     *CircuitBreaker
	status int
}

func (w *circuitBreakerResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *circuitBreakerResponseWriter) Write(b []byte) (int, error) {
	// Record result based on status code
	if w.status == 0 {
		w.status = http.StatusOK
	}

	if w.status >= 500 {
		w.cb.RecordFailure()
	} else {
		w.cb.RecordSuccess()
	}

	return w.ResponseWriter.Write(b)
}

// ConnectionPoolManager manages connection pools for various services
type ConnectionPoolManager struct {
	pools map[string]*PoolConfig
}

// PoolConfig holds connection pool configuration
type PoolConfig struct {
	MaxConnections    int
	MinConnections    int
	MaxIdleTime       time.Duration
	HealthCheckPeriod time.Duration
}

// NewConnectionPoolManager creates a new pool manager
func NewConnectionPoolManager() *ConnectionPoolManager {
	return &ConnectionPoolManager{
		pools: make(map[string]*PoolConfig),
	}
}

// RegisterPool registers a pool configuration
func (m *ConnectionPoolManager) RegisterPool(name string, config *PoolConfig) {
	m.pools[name] = config
}

// GetPool returns pool configuration
func (m *ConnectionPoolManager) GetPool(name string) *PoolConfig {
	return m.pools[name]
}

// GetRecommendedPool returns recommended pool settings for a service type
func GetRecommendedPool(serviceType string) *PoolConfig {
	switch serviceType {
	case "mongodb":
		return &PoolConfig{
			MaxConnections:    100,
			MinConnections:    10,
			MaxIdleTime:       30 * time.Minute,
			HealthCheckPeriod: 10 * time.Second,
		}
	case "redis":
		return &PoolConfig{
			MaxConnections:    50,
			MinConnections:    5,
			MaxIdleTime:       10 * time.Minute,
			HealthCheckPeriod: 5 * time.Second,
		}
	case "http":
		return &PoolConfig{
			MaxConnections:    200,
			MinConnections:    20,
			MaxIdleTime:       90 * time.Second,
			HealthCheckPeriod: 30 * time.Second,
		}
	default:
		return &PoolConfig{
			MaxConnections:    50,
			MinConnections:    5,
			MaxIdleTime:       30 * time.Minute,
			HealthCheckPeriod: 30 * time.Second,
		}
	}
}
