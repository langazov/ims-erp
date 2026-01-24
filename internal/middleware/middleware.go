package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type AuthContextKey string

const (
	UserContextKey        AuthContextKey = "user"
	TenantContextKey      AuthContextKey = "tenant"
	PermissionsContextKey AuthContextKey = "permissions"
)

func GetUserID(ctx context.Context) string {
	if v := ctx.Value(UserContextKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func GetTenantID(ctx context.Context) string {
	if v := ctx.Value(TenantContextKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func GetPermissions(ctx context.Context) []string {
	if v := ctx.Value(PermissionsContextKey); v != nil {
		if s, ok := v.([]string); ok {
			return s
		}
	}
	return nil
}

type LoggingMiddleware struct {
	logger *logger.Logger
}

func NewLoggingMiddleware(log *logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: log}
}

func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)

		userID := GetUserID(r.Context())
		tenantID := GetTenantID(r.Context())

		m.logger.New(r.Context()).Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", w.Header().Get("X-Status-Code"),
			"duration_ms", duration.Milliseconds(),
			"user_id", userID,
			"tenant_id", tenantID,
			"ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

type TracingMiddleware struct {
	tracer trace.Tracer
}

func NewTracingMiddleware() *TracingMiddleware {
	return &TracingMiddleware{
		tracer: otel.Tracer("http-server"),
	}
}

func (m *TracingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		ctx, span := m.tracer.Start(ctx, r.Method+" "+r.URL.Path,
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.host", r.Host),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
			),
		)
		defer span.End()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type RecoveryMiddleware struct {
	logger *logger.Logger
}

func NewRecoveryMiddleware(log *logger.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{logger: log}
}

func (m *RecoveryMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				m.logger.Error("Panic recovered",
					"error", r,
				)

				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Internal server error",
					"code":  "INTERNAL_ERROR",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type CORSMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
}

func NewCORSMiddleware(cfg *config.SecurityConfig) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: cfg.CORSDomain,
		allowedMethods: cfg.AllowedMethods,
		allowedHeaders: cfg.AllowedHeaders,
	}
}

func (m *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != "" {
			allowed := false
			for _, o := range m.allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(m.allowedMethods, ","))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(m.allowedHeaders, ","))
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type SecurityHeadersMiddleware struct{}

func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{}
}

func (m *SecurityHeadersMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")

		next.ServeHTTP(w, r)
	})
}

type RequestIDMiddleware struct {
	generator func() string
}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{
		generator: func() string {
			return time.Now().Format("20060102150405") + "-" + randomString(8)
		},
	}
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func (m *RequestIDMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = m.generator()
		}

		ctx := logger.WithRequestID(r.Context(), requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type TimeoutMiddleware struct {
	timeout time.Duration
}

func NewTimeoutMiddleware(timeout time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{timeout: timeout}
}

func (m *TimeoutMiddleware) Handler(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, m.timeout, "Request timeout")
}

type ValidatorMiddleware struct{}

func NewValidatorMiddleware() *ValidatorMiddleware {
	return &ValidatorMiddleware{}
}

func (m *ValidatorMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "" && !strings.Contains(contentType, "application/json") {
			http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
			return
		}

		if r.ContentLength > 10*1024*1024 {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func ChainMiddleware(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
