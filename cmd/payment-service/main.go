package main

import (
	"context"
	"encoding/json"
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

type PaymentService struct {
	config *config.Config
	logger *logger.Logger
}

func NewPaymentService(cfg *config.Config, log *logger.Logger) *PaymentService {
	return &PaymentService{
		config: cfg,
		logger: log,
	}
}

func (s *PaymentService) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readinessHandler)
	mux.HandleFunc("/live", s.livenessHandler)
	mux.HandleFunc("/metrics", s.metricsHandler)

	mux.HandleFunc("/api/v1/payments", s.handlePayments)
	mux.HandleFunc("/api/v1/payments/", s.handlePaymentByID)
	mux.HandleFunc("/api/v1/payments/process", s.handleProcessPayment)
	mux.HandleFunc("/api/v1/payments/refund", s.handleRefund)
	mux.HandleFunc("/api/v1/payments/webhook", s.handleWebhook)
	mux.HandleFunc("/api/v1/payments/methods", s.handlePaymentMethods)
	mux.HandleFunc("/api/v1/payments/transactions", s.handleTransactions)
	mux.HandleFunc("/api/v1/payments/report/daily", s.handleDailyReport)
	mux.HandleFunc("/api/v1/payments/report/summary", s.handleSummaryReport)

	return mux
}

func (s *PaymentService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s", "service": "payment-service"}`, time.Now().UTC())
}

func (s *PaymentService) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ready", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *PaymentService) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "alive", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *PaymentService) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "# Payment Service Metrics\n")
	fmt.Fprintf(w, "payment_service_up 1\n")
	fmt.Fprintf(w, "payment_service_requests_total 0\n")
	fmt.Fprintf(w, "payment_service_processed_total 0\n")
	fmt.Fprintf(w, "payment_service_failed_total 0\n")
	fmt.Fprintf(w, "payment_service_refunded_total 0\n")
	fmt.Fprintf(w, "payment_service_volume_total 0\n")
}

func (s *PaymentService) handlePayments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listPayments(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *PaymentService) handlePaymentByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getPayment(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *PaymentService) handleProcessPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.processPayment(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *PaymentService) handleRefund(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.processRefund(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *PaymentService) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.processWebhook(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *PaymentService) handlePaymentMethods(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getPaymentMethods(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *PaymentService) handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getTransactions(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *PaymentService) handleDailyReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getDailyReport(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *PaymentService) handleSummaryReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getSummaryReport(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *PaymentService) listPayments(w http.ResponseWriter, r *http.Request) {
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
	fmt.Fprintf(w, `{"payments": [], "total": 0, "page": %d, "pageSize": %d}`, page, pageSize)
}

func (s *PaymentService) getPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := r.URL.Query().Get("paymentId")
	_ = paymentID

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"payment": null}`)
}

func (s *PaymentService) processPayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InvoiceID   string  `json:"invoiceId"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		Method      string  `json:"method"`
		Provider    string  `json:"provider"`
		Description string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Payment processed", "paymentId": "%s", "status": "completed"}`, generateUUID())
}

func (s *PaymentService) processRefund(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PaymentID string  `json:"paymentId"`
		Amount    float64 `json:"amount"`
		Reason    string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Refund processed", "refundId": "%s"}`, generateUUID())
}

func (s *PaymentService) processWebhook(w http.ResponseWriter, r *http.Request) {
	var webhook struct {
		Event string                 `json:"event"`
		Data  map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Webhook processed"}`)
}

func (s *PaymentService) getPaymentMethods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"methods": [
		{"id": "credit_card", "name": "Credit Card", "providers": ["stripe", "adyen"]},
		{"id": "debit_card", "name": "Debit Card", "providers": ["stripe"]},
		{"id": "bank_transfer", "name": "Bank Transfer", "providers": ["ach", "sepa"]},
		{"id": "paypal", "name": "PayPal", "providers": ["paypal"]},
		{"id": "cash", "name": "Cash", "providers": []},
		{"id": "check", "name": "Check", "providers": []}
	]}`)
}

func (s *PaymentService) getTransactions(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	_ = tenantID
	_ = startDate
	_ = endDate

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"transactions": [], "total": 0}`)
}

func (s *PaymentService) getDailyReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	date := r.URL.Query().Get("date")

	_ = tenantID
	_ = date

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"report": "daily", "date": "%s", "totalTransactions": 0, "totalVolume": 0, "totalRefunds": 0}`, date)
}

func (s *PaymentService) getSummaryReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	_ = tenantID
	_ = startDate
	_ = endDate

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"report": "summary", "period": {"start": "%s", "end": "%s"}, "totalVolume": 0, "totalTransactions": 0, "successRate": 0}`, startDate, endDate)
}

func main() {
	cfg, err := config.Load("", "payment-service")
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

	service := NewPaymentService(cfg, log)
	mux := service.setupRoutes()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      mux,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		log.Info("Starting payment service", "port", cfg.App.Port)
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
