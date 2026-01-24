package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/pkg/logger"
)

type WarehouseService struct {
	config *config.Config
	logger *logger.Logger
}

func NewWarehouseService(cfg *config.Config, log *logger.Logger) *WarehouseService {
	return &WarehouseService{
		config: cfg,
		logger: log,
	}
}

func (s *WarehouseService) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readinessHandler)
	mux.HandleFunc("/live", s.livenessHandler)
	mux.HandleFunc("/metrics", s.metricsHandler)

	mux.HandleFunc("/api/v1/warehouses", s.handleWarehouses)
	mux.HandleFunc("/api/v1/warehouses/", s.handleWarehouseByID)
	mux.HandleFunc("/api/v1/warehouses/", s.handleWarehouseLocations)
	mux.HandleFunc("/api/v1/warehouses/", s.handleWarehouseOperations)
	mux.HandleFunc("/api/v1/warehouses/", s.handleWarehouseCapacity)
	mux.HandleFunc("/api/v1/locations", s.handleLocations)
	mux.HandleFunc("/api/v1/locations/", s.handleLocationByID)
	mux.HandleFunc("/api/v1/operations", s.handleOperations)
	mux.HandleFunc("/api/v1/operations/", s.handleOperationByID)
	mux.HandleFunc("/api/v1/inventory/adjust", s.handleInventoryAdjust)
	mux.HandleFunc("/api/v1/inventory/transfer", s.handleInventoryTransfer)
	mux.HandleFunc("/api/v1/inventory/reserve", s.handleReserveStock)
	mux.HandleFunc("/api/v1/inventory/release", s.handleReleaseStock)
	mux.HandleFunc("/api/v1/inventory/commit", s.handleCommitStock)
	mux.HandleFunc("/api/v1/inventory/levels", s.handleInventoryLevels)
	mux.HandleFunc("/api/v1/inventory/movements", s.handleInventoryMovements)

	return mux
}

func (s *WarehouseService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s", "service": "warehouse-service"}`, time.Now().UTC())
}

func (s *WarehouseService) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ready", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *WarehouseService) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "alive", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *WarehouseService) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "# Warehouse Service Metrics\n")
	fmt.Fprintf(w, "warehouse_service_up 1\n")
	fmt.Fprintf(w, "warehouse_service_requests_total 0\n")
	fmt.Fprintf(w, "warehouse_service_created_total 0\n")
	fmt.Fprintf(w, "warehouse_service_locations_total 0\n")
	fmt.Fprintf(w, "warehouse_service_operations_total 0\n")
}

func (s *WarehouseService) handleWarehouses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listWarehouses(w, r)
	case http.MethodPost:
		s.createWarehouse(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WarehouseService) handleWarehouseByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getWarehouse(w, r)
	case http.MethodPut:
		s.updateWarehouse(w, r)
	case http.MethodDelete:
		s.deleteWarehouse(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WarehouseService) handleWarehouseLocations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getWarehouseLocations(w, r)
	case http.MethodPost:
		s.createLocation(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WarehouseService) handleWarehouseOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getWarehouseOperations(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WarehouseService) handleWarehouseCapacity(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		s.updateCapacity(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WarehouseService) handleLocations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listLocations(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WarehouseService) handleLocationByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getLocation(w, r)
	case http.MethodPut:
		s.updateLocation(w, r)
	case http.MethodDelete:
		s.deleteLocation(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WarehouseService) handleOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listOperations(w, r)
	case http.MethodPost:
		s.createOperation(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WarehouseService) handleOperationByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getOperation(w, r)
	case http.MethodPut:
		s.updateOperation(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WarehouseService) handleInventoryAdjust(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.adjustInventory(w, r)
}

func (s *WarehouseService) handleInventoryTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.transferInventory(w, r)
}

func (s *WarehouseService) handleReserveStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.reserveStock(w, r)
}

func (s *WarehouseService) handleReleaseStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.releaseStock(w, r)
}

func (s *WarehouseService) handleCommitStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.commitStock(w, r)
}

func (s *WarehouseService) handleInventoryLevels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.getInventoryLevels(w, r)
}

func (s *WarehouseService) handleInventoryMovements(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listMovements(w, r)
	case http.MethodPost:
		s.createMovement(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WarehouseService) listWarehouses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data": [], "meta": {"page": 1, "limit": 20, "total": 0}}`)
}

func (s *WarehouseService) createWarehouse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id": "%s", "status": "created"}`, uuid.New())
}

func (s *WarehouseService) getWarehouse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s", "name": "Warehouse", "isActive": true}`, uuid.New())
}

func (s *WarehouseService) updateWarehouse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s", "status": "updated"}`, uuid.New())
}

func (s *WarehouseService) deleteWarehouse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s", "status": "deleted"}`, uuid.New())
}

func (s *WarehouseService) getWarehouseLocations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data": [], "meta": {"page": 1, "limit": 50, "total": 0}}`)
}

func (s *WarehouseService) createLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id": "%s", "status": "created"}`, uuid.New())
}

func (s *WarehouseService) getWarehouseOperations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data": [], "meta": {"page": 1, "limit": 50, "total": 0}}`)
}

func (s *WarehouseService) updateCapacity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s", "status": "capacity_updated"}`, uuid.New())
}

func (s *WarehouseService) listLocations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data": [], "meta": {"page": 1, "limit": 50, "total": 0}}`)
}

func (s *WarehouseService) getLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s", "code": "A-01-01-01", "isActive": true}`, uuid.New())
}

func (s *WarehouseService) updateLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s", "status": "updated"}`, uuid.New())
}

func (s *WarehouseService) deleteLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s", "status": "deleted"}`, uuid.New())
}

func (s *WarehouseService) listOperations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data": [], "meta": {"page": 1, "limit": 50, "total": 0}}`)
}

func (s *WarehouseService) createOperation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id": "%s", "status": "pending"}`, uuid.New())
}

func (s *WarehouseService) getOperation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s", "status": "pending", "type": "pick"}`, uuid.New())
}

func (s *WarehouseService) updateOperation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s", "status": "updated"}`, uuid.New())
}

func (s *WarehouseService) adjustInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"id": "%s", "status": "adjusted"}`, uuid.New())
}

func (s *WarehouseService) transferInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"id": "%s", "status": "transferred"}`, uuid.New())
}

func (s *WarehouseService) reserveStock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"id": "%s", "status": "reserved"}`, uuid.New())
}

func (s *WarehouseService) releaseStock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"id": "%s", "status": "released"}`, uuid.New())
}

func (s *WarehouseService) commitStock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"id": "%s", "status": "committed"}`, uuid.New())
}

func (s *WarehouseService) getInventoryLevels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data": [], "meta": {"page": 1, "limit": 50, "total": 0}}`)
}

func (s *WarehouseService) listMovements(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data": [], "meta": {"page": 1, "limit": 50, "total": 0}}`)
}

func (s *WarehouseService) createMovement(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id": "%s", "status": "created"}`, uuid.New())
}

func (s *WarehouseService) runServer() {
	port := 8087

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.setupRoutes(),
	}

	go func() {
		s.logger.Info("Starting warehouse service", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server failed", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down warehouse service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", err)
	}
}

func main() {
	cfg := &config.Config{}

	log, err := logger.New(logger.Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "warehouse-service",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	service := NewWarehouseService(cfg, log)
	service.runServer()
}
