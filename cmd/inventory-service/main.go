package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/ims-erp/system/pkg/tracer"
)

var allowedOrigins = []string{
	"http://localhost:5173",
	"http://localhost:5178",
	"http://localhost:5174",
	"http://localhost:5175",
	"http://localhost:5176",
	"http://localhost:5177",
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		isAllowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func corsOptionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		isAllowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type InventoryService struct {
	config *config.Config
	logger *logger.Logger
}

func NewInventoryService(cfg *config.Config, log *logger.Logger) *InventoryService {
	return &InventoryService{
		config: cfg,
		logger: log,
	}
}

func (s *InventoryService) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readinessHandler)
	mux.HandleFunc("/live", s.livenessHandler)
	mux.HandleFunc("/metrics", s.metricsHandler)

	mux.HandleFunc("/api/v1/inventory/items", s.handleInventoryItems)
	mux.HandleFunc("/api/v1/inventory/transactions", s.handleTransactions)
	mux.HandleFunc("/api/v1/inventory/warehouses", s.handleWarehouses)
	mux.HandleFunc("/api/v1/inventory/reservations", s.handleReservations)
	mux.HandleFunc("/api/v1/inventory/adjustments", s.handleAdjustments)
	mux.HandleFunc("/api/v1/inventory/levels", s.handleLevels)
	mux.HandleFunc("/api/v1/inventory/reports/stock", s.handleStockReport)
	mux.HandleFunc("/api/v1/inventory/reports/movements", s.handleMovementsReport)

	return mux
}

func (s *InventoryService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s", "service": "inventory-service"}`, time.Now().UTC())
}

func (s *InventoryService) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ready", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *InventoryService) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "alive", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *InventoryService) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "# Inventory Service Metrics\n")
	fmt.Fprintf(w, "inventory_service_up 1\n")
	fmt.Fprintf(w, "inventory_service_requests_total 0\n")
	fmt.Fprintf(w, "inventory_service_items_total 0\n")
	fmt.Fprintf(w, "inventory_service_transactions_total 0\n")
	fmt.Fprintf(w, "inventory_service_warehouses_total 0\n")
}

func (s *InventoryService) handleInventoryItems(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listInventoryItems(w, r)
	case http.MethodPost:
		s.createInventoryItem(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InventoryService) handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTransactions(w, r)
	case http.MethodPost:
		s.createTransaction(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InventoryService) handleWarehouses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listWarehouses(w, r)
	case http.MethodPost:
		s.createWarehouse(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InventoryService) handleReservations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listReservations(w, r)
	case http.MethodPost:
		s.createReservation(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InventoryService) handleAdjustments(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.createAdjustment(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InventoryService) handleLevels(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getInventoryLevels(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InventoryService) handleStockReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.generateStockReport(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InventoryService) handleMovementsReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.generateMovementsReport(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InventoryService) listInventoryItems(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	productID := r.URL.Query().Get("productId")
	warehouseID := r.URL.Query().Get("warehouseId")
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 50)

	_ = tenantID
	_ = productID
	_ = warehouseID
	_ = page
	_ = pageSize

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"items": [], "total": 0, "page": %d, "pageSize": %d}`, page, pageSize)
}

func (s *InventoryService) createInventoryItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Inventory item created", "id": "%s"}`, generateUUID())
}

func (s *InventoryService) listTransactions(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	productID := r.URL.Query().Get("productId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	_ = tenantID
	_ = productID
	_ = startDate
	_ = endDate

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"transactions": [], "total": 0}`)
}

func (s *InventoryService) createTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Transaction recorded", "id": "%s"}`, generateUUID())
}

func (s *InventoryService) listWarehouses(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")

	_ = tenantID

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"warehouses": [], "total": 0}`)
}

func (s *InventoryService) createWarehouse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Warehouse created", "id": "%s"}`, generateUUID())
}

func (s *InventoryService) listReservations(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	status := r.URL.Query().Get("status")

	_ = tenantID
	_ = status

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"reservations": [], "total": 0}`)
}

func (s *InventoryService) createReservation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Reservation created", "id": "%s"}`, generateUUID())
}

func (s *InventoryService) createAdjustment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Adjustment recorded", "id": "%s"}`, generateUUID())
}

func (s *InventoryService) getInventoryLevels(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	productID := r.URL.Query().Get("productId")
	warehouseID := r.URL.Query().Get("warehouseId")

	_ = tenantID
	_ = productID
	_ = warehouseID

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"levels": [], "total": 0}`)
}

func (s *InventoryService) generateStockReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	warehouseID := r.URL.Query().Get("warehouseId")
	includeZeroStock := r.URL.Query().Get("includeZeroStock") == "true"

	_ = tenantID
	_ = warehouseID
	_ = includeZeroStock

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"report": "stock", "generatedAt": "%s", "items": []}`, time.Now().UTC())
}

func (s *InventoryService) generateMovementsReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	groupBy := r.URL.Query().Get("groupBy")

	_ = tenantID
	_ = startDate
	_ = endDate
	_ = groupBy

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"report": "movements", "generatedAt": "%s", "summary": {}}`, time.Now().UTC())
}

func main() {
	cfg, err := config.Load("", "inventory-service")
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

	service := NewInventoryService(cfg, log)
	mux := service.setupRoutes()
	handler := corsMiddleware(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      handler,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		log.Info("Starting inventory service", "port", cfg.App.Port)
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

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
