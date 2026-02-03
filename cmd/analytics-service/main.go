package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ims-erp/system/internal/analytics"
	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
)

// AnalyticsServer provides real-time analytics dashboard
type AnalyticsServer struct {
	service    *analytics.ReportingService
	cache      *repository.Cache
	logger     *logger.Logger
	clients    map[string]*DashboardClient
	mu         sync.RWMutex
	aggregated *DashboardData
}

// DashboardClient represents a connected WebSocket client
type DashboardClient struct {
	id       string
	tenantID string
	conn     *websocket.Conn
	send     chan []byte
	server   *AnalyticsServer
}

// DashboardData contains aggregated dashboard metrics
type DashboardData struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
	Revenue   *analytics.RevenueSummary
	Aging     *analytics.AgingReport
	Payments  *analytics.PaymentSummary
}

func main() {
	// Load configuration
	cfg, err := config.Load("analytics-service", "./configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logCfg := logger.Config{
		Level:       cfg.Logging.Level,
		Format:      "json",
		ServiceName: "analytics-service",
	}
	logr, err := logger.New(logCfg)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Initialize repositories
	mongoDB, err := repository.NewMongoDB(cfg.MongoDB, logr)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	readModelStore := repository.NewReadModelStore(mongoDB, cfg.MongoDB.Database, logr)

	redisClient, err := repository.NewRedis(cfg.Redis, logr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	cache := repository.NewCache(redisClient, "analytics", logr)

	// Initialize reporting service
	service := analytics.NewReportingService(readModelStore, cache, logr)

	// Create server
	server := NewAnalyticsServer(service, cache, logr)

	// Start background aggregation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.startAggregation(ctx)
	go server.startCacheWarming(ctx)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/dashboard", server.handleDashboard)
	mux.HandleFunc("/api/v1/dashboard/ws", server.handleWebSocket)
	mux.HandleFunc("/api/v1/health", server.handleHealth)
	mux.HandleFunc("/api/v1/metrics/revenue", server.handleRevenueMetrics)
	mux.HandleFunc("/api/v1/metrics/aging", server.handleAgingMetrics)
	mux.HandleFunc("/api/v1/metrics/payments", server.handlePaymentMetrics)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logr.Info("Starting analytics service", "port", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logr.Fatal("Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logr.Info("Shutting down analytics service...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logr.Error("Server shutdown error", "error", err)
	}

	// Close all WebSocket connections
	server.closeAllClients()

	logr.Info("Analytics service stopped")
}

// NewAnalyticsServer creates a new analytics server
func NewAnalyticsServer(service *analytics.ReportingService, cache *repository.Cache, log *logger.Logger) *AnalyticsServer {
	return &AnalyticsServer{
		service: service,
		cache:   cache,
		logger:  log,
		clients: make(map[string]*DashboardClient),
	}
}

// handleDashboard returns current dashboard data
func (s *AnalyticsServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID from request
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	// Parse tenant UUID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Get dashboard data
	data, err := s.service.GetDashboardData(ctx, tenantUUID)
	if err != nil {
		s.logger.Error("Failed to get dashboard data", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handleWebSocket handles WebSocket connections for real-time updates
func (s *AnalyticsServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Get tenant ID from request
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade failed", "error", err)
		return
	}

	// Create client
	client := &DashboardClient{
		id:       uuid.New().String(),
		tenantID: tenantID,
		conn:     conn,
		send:     make(chan []byte, 256),
		server:   s,
	}

	// Register client
	s.mu.Lock()
	s.clients[client.id] = client
	s.mu.Unlock()

	// Start goroutines
	go client.writePump()
	go client.readPump()

	// Send initial data
	s.sendInitialData(client)

	s.logger.Info("WebSocket client connected", "client_id", client.id, "tenant_id", tenantID)
}

// handleHealth returns health status
func (s *AnalyticsServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleRevenueMetrics returns revenue analytics
func (s *AnalyticsServer) handleRevenueMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Parse date range
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	if start := r.URL.Query().Get("start"); start != "" {
		if parsed, err := time.Parse(time.RFC3339, start); err == nil {
			startDate = parsed
		}
	}

	if end := r.URL.Query().Get("end"); end != "" {
		if parsed, err := time.Parse(time.RFC3339, end); err == nil {
			endDate = parsed
		}
	}

	summary, err := s.service.GetRevenueSummary(ctx, tenantUUID, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get revenue summary", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// handleAgingMetrics returns aging report
func (s *AnalyticsServer) handleAgingMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	asOfDate := time.Now()
	if date := r.URL.Query().Get("as_of"); date != "" {
		if parsed, err := time.Parse(time.RFC3339, date); err == nil {
			asOfDate = parsed
		}
	}

	report, err := s.service.GetAgingReport(ctx, tenantUUID, asOfDate)
	if err != nil {
		s.logger.Error("Failed to get aging report", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// handlePaymentMetrics returns payment analytics
func (s *AnalyticsServer) handlePaymentMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Parse date range
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	if start := r.URL.Query().Get("start"); start != "" {
		if parsed, err := time.Parse(time.RFC3339, start); err == nil {
			startDate = parsed
		}
	}

	if end := r.URL.Query().Get("end"); end != "" {
		if parsed, err := time.Parse(time.RFC3339, end); err == nil {
			endDate = parsed
		}
	}

	summary, err := s.service.GetPaymentSummary(ctx, tenantUUID, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get payment summary", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// startAggregation runs background job to aggregate metrics every 30 seconds
func (s *AnalyticsServer) startAggregation(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.aggregateMetrics(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// aggregateMetrics aggregates metrics from all sources
func (s *AnalyticsServer) aggregateMetrics(ctx context.Context) {
	// Get aggregated metrics for default tenant
	tenantUUID := uuid.MustParse("default")

	dashboard, err := s.service.GetDashboardData(ctx, tenantUUID)
	if err != nil {
		s.logger.Error("Failed to aggregate metrics", "error", err)
		return
	}

	s.mu.Lock()
	s.aggregated = &DashboardData{
		Timestamp: time.Now(),
		Revenue:   &dashboard.Revenue,
		Aging:     &dashboard.Aging,
		Payments:  &dashboard.Payments,
		Metrics:   dashboard.KeyMetrics,
	}
	s.mu.Unlock()

	// Broadcast to all connected clients
	s.broadcastUpdate()
}

// startCacheWarming warms up cache with dashboard data
func (s *AnalyticsServer) startCacheWarming(ctx context.Context) {
	// Initial warm-up
	s.warmCache(ctx)

	// Periodic warm-up every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.warmCache(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// warmCache pre-computes and caches dashboard data
func (s *AnalyticsServer) warmCache(ctx context.Context) {
	// Warm up dashboard data for common tenants
	tenants := []string{"default", "tenant-1", "tenant-2"}

	for _, tenantID := range tenants {
		tenantUUID, err := uuid.Parse(tenantID)
		if err != nil {
			continue
		}

		_, err = s.service.GetDashboardData(ctx, tenantUUID)
		if err != nil {
			s.logger.Error("Failed to warm cache", "tenant", tenantID, "error", err)
		}
	}

	s.logger.Info("Cache warming completed")
}

// sendInitialData sends initial dashboard data to a new client
func (s *AnalyticsServer) sendInitialData(client *DashboardClient) {
	s.mu.RLock()
	data := s.aggregated
	s.mu.RUnlock()

	if data == nil {
		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"type": "initial",
		"data": data,
	})
	if err != nil {
		s.logger.Error("Failed to marshal initial data", "error", err)
		return
	}

	select {
	case client.send <- payload:
	default:
		s.logger.Warn("Client send buffer full", "client_id", client.id)
	}
}

// broadcastUpdate sends updates to all connected clients
func (s *AnalyticsServer) broadcastUpdate() {
	s.mu.RLock()
	data := s.aggregated
	clients := make([]*DashboardClient, 0, len(s.clients))
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.RUnlock()

	if data == nil {
		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"type": "update",
		"data": data,
	})
	if err != nil {
		s.logger.Error("Failed to marshal update", "error", err)
		return
	}

	for _, client := range clients {
		select {
		case client.send <- payload:
		default:
			// Client buffer full, will catch up on next update
		}
	}
}

// closeAllClients closes all WebSocket connections
func (s *AnalyticsServer) closeAllClients() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		close(client.send)
		client.conn.Close()
	}
}

// readPump pumps messages from the WebSocket connection
func (c *DashboardClient) readPump() {
	defer func() {
		c.server.mu.Lock()
		delete(c.server.clients, c.id)
		c.server.mu.Unlock()
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.server.logger.Error("WebSocket read error", "client_id", c.id, "error", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *DashboardClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
