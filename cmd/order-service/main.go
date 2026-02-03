package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type OrderService struct {
	config *config.Config
	logger *logger.Logger
}

func NewOrderService(cfg *config.Config, log *logger.Logger) *OrderService {
	return &OrderService{
		config: cfg,
		logger: log,
	}
}

func (s *OrderService) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readinessHandler)
	mux.HandleFunc("/live", s.livenessHandler)
	mux.HandleFunc("/metrics", s.metricsHandler)

	mux.HandleFunc("/api/v1/orders", s.handleOrders)
	mux.HandleFunc("/api/v1/orders/", s.handleOrderRouter)
	mux.HandleFunc("/api/v1/orders/status", s.handleUpdateStatus)
	mux.HandleFunc("/api/v1/orders/search", s.handleSearch)
	mux.HandleFunc("/api/v1/orders/report/summary", s.handleSummaryReport)
	mux.HandleFunc("/api/v1/orders/report/fulfillment", s.handleFulfillmentReport)

	return mux
}

func (s *OrderService) handleOrderRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	idPattern := "/api/v1/orders/"
	linesPattern := "/api/v1/orders//lines"
	paymentsPattern := "/api/v1/orders//payments"
	fulfillmentPattern := "/api/v1/orders//fulfillment"
	shipmentPattern := "/api/v1/orders//shipment"

	switch {
	case strings.HasPrefix(path, linesPattern):
		s.handleOrderLines(w, r)
	case strings.HasPrefix(path, paymentsPattern):
		s.handleOrderPayments(w, r)
	case strings.HasPrefix(path, fulfillmentPattern):
		s.handleOrderFulfillment(w, r)
	case strings.HasPrefix(path, shipmentPattern):
		s.handleOrderShipment(w, r)
	case strings.HasPrefix(path, idPattern):
		id := strings.TrimPrefix(path, idPattern)
		id = strings.Split(id, "/")[0]
		r.URL.Query().Set("orderId", id)
		s.handleOrderByID(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (s *OrderService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s", "service": "order-service"}`, time.Now().UTC())
}

func (s *OrderService) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ready", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *OrderService) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "alive", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *OrderService) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "# Order Service Metrics\n")
	fmt.Fprintf(w, "order_service_up 1\n")
	fmt.Fprintf(w, "order_service_requests_total 0\n")
	fmt.Fprintf(w, "order_service_created_total 0\n")
	fmt.Fprintf(w, "order_service_completed_total 0\n")
	fmt.Fprintf(w, "order_service_cancelled_total 0\n")
}

func (s *OrderService) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listOrders(w, r)
	case http.MethodPost:
		s.createOrder(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *OrderService) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getOrder(w, r)
	case http.MethodPut:
		s.updateOrder(w, r)
	case http.MethodDelete:
		s.cancelOrder(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *OrderService) handleOrderLines(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.addOrderLine(w, r)
	} else if r.Method == http.MethodDelete {
		s.removeOrderLine(w, r)
	} else if r.Method == http.MethodPut {
		s.updateOrderLine(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *OrderService) handleOrderPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.addPayment(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *OrderService) handleOrderFulfillment(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.fulfillOrder(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *OrderService) handleOrderShipment(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.shipOrder(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *OrderService) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		s.updateOrderStatus(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *OrderService) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.searchOrders(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *OrderService) handleSummaryReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getSummaryReport(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *OrderService) handleFulfillmentReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getFulfillmentReport(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *OrderService) listOrders(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	clientID := r.URL.Query().Get("clientId")
	status := r.URL.Query().Get("status")
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 50)

	_ = tenantID
	_ = clientID
	_ = status
	_ = page
	_ = pageSize

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"orders": [], "total": 0, "page": %d, "pageSize": %d}`, page, pageSize)
}

func (s *OrderService) createOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Order created", "id": "%s", "orderNumber": "ORD-%s"}`, generateUUID(), generateOrderNumber())
}

func (s *OrderService) getOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("orderId")
	_ = orderID

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"order": null}`)
}

func (s *OrderService) updateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Order updated"}`)
}

func (s *OrderService) cancelOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Order cancelled", "reason": "%s"}`, req.Reason)
}

func (s *OrderService) addOrderLine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Line added"}`)
}

func (s *OrderService) removeOrderLine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Line removed"}`)
}

func (s *OrderService) updateOrderLine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Line updated"}`)
}

func (s *OrderService) addPayment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Payment added"}`)
}

func (s *OrderService) fulfillOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Order fulfilled"}`)
}

func (s *OrderService) shipOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TrackingNumber string `json:"trackingNumber"`
		Carrier        string `json:"carrier"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Order shipped", "trackingNumber": "%s"}`, req.TrackingNumber)
}

func (s *OrderService) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Status updated", "newStatus": "%s"}`, req.Status)
}

func (s *OrderService) searchOrders(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	_ = query

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"results": [], "total": 0}`)
}

func (s *OrderService) getSummaryReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	_ = tenantID
	_ = startDate
	_ = endDate

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"report": "summary", "period": {"start": "%s", "end": "%s"}, "totalOrders": 0, "totalRevenue": 0, "byStatus": {}}`, startDate, endDate)
}

func (s *OrderService) getFulfillmentReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	_ = tenantID
	_ = startDate
	_ = endDate

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"report": "fulfillment", "period": {"start": "%s", "end": "%s"}, "totalShipped": 0, "totalDelivered": 0, "averageFulfillmentTime": 0}`, startDate, endDate)
}

func main() {
	cfg, err := config.Load("", "order-service")
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

	service := NewOrderService(cfg, log)
	mux := service.setupRoutes()
	handler := corsMiddleware(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      handler,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		log.Info("Starting order service", "port", cfg.App.Port)
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

func generateOrderNumber() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
