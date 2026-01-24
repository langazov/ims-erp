package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type HealthChecker struct {
	config  *config.Config
	mongodb *repository.MongoDB
	redis   *repository.Redis
	logger  *logger.Logger
	tracer  trace.Tracer
}

func NewHealthChecker(
	cfg *config.Config,
	mongodb *repository.MongoDB,
	redis *repository.Redis,
	log *logger.Logger,
) *HealthChecker {
	return &HealthChecker{
		config:  cfg,
		mongodb: mongodb,
		redis:   redis,
		logger:  log,
		tracer:  otel.Tracer("health"),
	}
}

type HealthStatus struct {
	Status    string           `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Version   string           `json:"version"`
	Uptime    string           `json:"uptime"`
	Checks    map[string]Check `json:"checks"`
}

type Check struct {
	Status  string `json:"status"`
	Latency string `json:"latency"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Component struct {
	Name    string
	Checker func(ctx context.Context) Check
}

func (h *HealthChecker) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		status := h.GetHealth(ctx)
		statusJSON, err := json.Marshal(status)
		if err != nil {
			h.logger.Error("Failed to marshal health status", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if status.Status == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		w.Write(statusJSON)
	})
}

func (h *HealthChecker) GetHealth(ctx context.Context) HealthStatus {
	start := time.Now()

	checks := make(map[string]Check)

	var overallStatus string = "healthy"

	components := []Component{
		{"mongodb", h.checkMongoDB},
		{"redis", h.checkRedis},
	}

	for _, comp := range components {
		check := comp.Checker(ctx)
		checks[comp.Name] = check
		if check.Status != "healthy" {
			overallStatus = "unhealthy"
		}
	}

	return HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Version:   h.config.App.Version,
		Uptime:    time.Since(start).String(),
		Checks:    checks,
	}
}

func (h *HealthChecker) checkMongoDB(ctx context.Context) Check {
	start := time.Now()
	defer func() { _ = time.Since(start) }()

	if err := h.mongodb.Health(ctx); err != nil {
		return Check{
			Status:  "unhealthy",
			Latency: time.Since(start).String(),
			Error:   err.Error(),
		}
	}

	return Check{
		Status:  "healthy",
		Latency: time.Since(start).String(),
		Message: "Connected",
	}
}

func (h *HealthChecker) checkRedis(ctx context.Context) Check {
	start := time.Now()
	defer func() { _ = time.Since(start) }()

	if err := h.redis.Health(ctx); err != nil {
		return Check{
			Status:  "unhealthy",
			Latency: time.Since(start).String(),
			Error:   err.Error(),
		}
	}

	return Check{
		Status:  "healthy",
		Latency: time.Since(start).String(),
		Message: "Connected",
	}
}

type ReadinessChecker struct {
	components []Component
	logger     *logger.Logger
}

func NewReadinessChecker(log *logger.Logger) *ReadinessChecker {
	return &ReadinessChecker{
		components: make([]Component, 0),
		logger:     log,
	}
}

func (r *ReadinessChecker) AddComponent(name string, checker func(ctx context.Context) Check) {
	r.components = append(r.components, Component{name, checker})
}

func (r *ReadinessChecker) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer cancel()

		allReady := true
		checks := make(map[string]Check)

		for _, comp := range r.components {
			check := comp.Checker(ctx)
			checks[comp.Name] = check
			if check.Status != "healthy" {
				allReady = false
			}
		}

		status := struct {
			Status string           `json:"status"`
			Checks map[string]Check `json:"checks"`
		}{
			Status: "ready",
			Checks: checks,
		}

		statusJSON, err := json.Marshal(status)
		if err != nil {
			r.logger.Error("Failed to marshal readiness status", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if allReady {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		w.Write(statusJSON)
	})
}

type LivenessChecker struct {
	startTime time.Time
}

func NewLivenessChecker() *LivenessChecker {
	return &LivenessChecker{
		startTime: time.Now(),
	}
}

func (l *LivenessChecker) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "alive",
			"uptime":    time.Since(l.startTime).String(),
			"timestamp": time.Now().UTC(),
		})
	})
}
