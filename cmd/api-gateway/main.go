package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/ims-erp/system/pkg/tracer"
)

type ServiceConfig struct {
	Name    string
	URL     string
	Paths   map[string]string
	Methods map[string]string
}

type APIGateway struct {
	config   *config.Config
	logger   *logger.Logger
	services map[string]ServiceConfig
}

func NewAPIGateway(cfg *config.Config, log *logger.Logger) *APIGateway {
	return &APIGateway{
		config:   cfg,
		logger:   log,
		services: make(map[string]ServiceConfig),
	}
}

func (g *APIGateway) AddService(name string, svcURL string, paths map[string]string, methods map[string]string) {
	g.services[name] = ServiceConfig{
		Name:    name,
		URL:     svcURL,
		Paths:   paths,
		Methods: methods,
	}
}

func (g *APIGateway) buildRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", g.healthHandler)
	mux.HandleFunc("/ready", g.readinessHandler)
	mux.HandleFunc("/live", g.livenessHandler)
	mux.HandleFunc("/api/v1/auth/", g.authHandler)
	mux.HandleFunc("/api/v1/clients/", g.clientsHandler)
	mux.HandleFunc("/api/v1/invoices/", g.invoicesHandler)
	mux.HandleFunc("/api/v1/payments/", g.paymentsHandler)
	mux.HandleFunc("/api/v1/products/", g.productsHandler)
	mux.HandleFunc("/api/v1/orders/", g.ordersHandler)

	return mux
}

func (g *APIGateway) createProxy(targetURL string) http.Handler {
	target, _ := url.Parse(targetURL)
	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Request-ID", generateRequestID())
	}

	return proxy
}

func (g *APIGateway) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   g.config.App.Name,
	})
}

func (g *APIGateway) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().UTC(),
		"services":  g.checkServices(),
	})
}

func (g *APIGateway) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
	})
}

func (g *APIGateway) checkServices() map[string]string {
	status := make(map[string]string)
	for name, svc := range g.services {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", svc.URL+"/health", nil)
		if err != nil {
			status[name] = "unknown"
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode >= 500 {
			status[name] = "unhealthy"
		} else {
			status[name] = "healthy"
		}
	}
	return status
}

func (g *APIGateway) authHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/auth")

	switch {
	case strings.HasPrefix(path, "/register"):
		g.proxyRequest(w, r, "http://localhost:8081")
	case strings.HasPrefix(path, "/login"):
		g.proxyRequest(w, r, "http://localhost:8081")
	case strings.HasPrefix(path, "/refresh"):
		g.proxyRequest(w, r, "http://localhost:8081")
	case strings.HasPrefix(path, "/me"):
		g.proxyRequest(w, r, "http://localhost:8081")
	default:
		g.proxyRequest(w, r, "http://localhost:8081")
	}
}

func (g *APIGateway) clientsHandler(w http.ResponseWriter, r *http.Request) {
	g.proxyRequest(w, r, "http://localhost:8082")
}

func (g *APIGateway) invoicesHandler(w http.ResponseWriter, r *http.Request) {
	g.proxyRequest(w, r, "http://localhost:8083")
}

func (g *APIGateway) paymentsHandler(w http.ResponseWriter, r *http.Request) {
	g.proxyRequest(w, r, "http://localhost:8084")
}

func (g *APIGateway) productsHandler(w http.ResponseWriter, r *http.Request) {
	g.proxyRequest(w, r, "http://localhost:8085")
}

func (g *APIGateway) ordersHandler(w http.ResponseWriter, r *http.Request) {
	g.proxyRequest(w, r, "http://localhost:8086")
}

func (g *APIGateway) proxyRequest(w http.ResponseWriter, r *http.Request, target string) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	r = r.WithContext(ctx)
	r.Header.Set("X-Forwarded-Host", r.Host)
	r.Header.Set("X-Request-ID", generateRequestID())

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		r.Header.Set("X-Authorization", authHeader)
	}

	proxy := g.createProxy(target)
	proxy.ServeHTTP(w, r)
}

func (g *APIGateway) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/auth/") || r.URL.Path == "/health" || r.URL.Path == "/ready" || r.URL.Path == "/live" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg, err := config.Load("", "api-gateway")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(logger.Config{
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		ServiceName: cfg.App.Name,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	tr, err := tracer.New(tracer.Config{
		Enabled:      cfg.Tracing.Enabled,
		ServiceName:  cfg.App.Name,
		ExporterType: cfg.Tracing.ExporterType,
		Endpoint:     cfg.Tracing.Endpoint,
		SamplerType:  cfg.Tracing.SamplerType,
		SamplerRatio: cfg.Tracing.SamplerRatio,
	})
	if err != nil {
		log.Error("Failed to create tracer", "error", err)
		os.Exit(1)
	}
	defer tr.Shutdown(context.Background())

	gateway := NewAPIGateway(cfg, log)

	mux := gateway.buildRouter()
	mux = gateway.authenticationMiddleware(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      mux,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		log.Info("Starting API Gateway", "port", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server stopped")
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
