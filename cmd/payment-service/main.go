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

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/commands"
	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/messaging"
	"github.com/ims-erp/system/internal/queries"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/errors"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/ims-erp/system/pkg/tracer"
)

type PaymentService struct {
	config         *config.Config
	logger         *logger.Logger
	paymentHandler *commands.PaymentCommandHandler
	queryHandler   *queries.PaymentQueryHandler
	webhookHandler *commands.WebhookHandler
	paymentRepo    commands.PaymentRepository
	invoiceRepo    commands.InvoiceRepository
	publisher      commands.Publisher
	processors     *domain.ProcessorRegistry
}

func NewPaymentService(
	cfg *config.Config,
	log *logger.Logger,
	paymentHandler *commands.PaymentCommandHandler,
	queryHandler *queries.PaymentQueryHandler,
	webhookHandler *commands.WebhookHandler,
	paymentRepo commands.PaymentRepository,
	invoiceRepo commands.InvoiceRepository,
	publisher commands.Publisher,
	processors *domain.ProcessorRegistry,
) *PaymentService {
	return &PaymentService{
		config:         cfg,
		logger:         log,
		paymentHandler: paymentHandler,
		queryHandler:   queryHandler,
		webhookHandler: webhookHandler,
		paymentRepo:    paymentRepo,
		invoiceRepo:    invoiceRepo,
		publisher:      publisher,
		processors:     processors,
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
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *PaymentService) handlePaymentByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getPayment(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *PaymentService) handleProcessPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.processPayment(w, r)
	} else {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *PaymentService) handleRefund(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.processRefund(w, r)
	} else {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *PaymentService) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.processWebhook(w, r)
	} else {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *PaymentService) handlePaymentMethods(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getPaymentMethods(w, r)
	} else {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *PaymentService) handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getTransactions(w, r)
	} else {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *PaymentService) handleDailyReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getDailyReport(w, r)
	} else {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *PaymentService) handleSummaryReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getSummaryReport(w, r)
	} else {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *PaymentService) listPayments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, http.StatusBadRequest, "tenantId is required")
		return
	}

	clientID := r.URL.Query().Get("clientId")
	status := r.URL.Query().Get("status")
	method := r.URL.Query().Get("method")
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 20)

	startDate, _ := time.Parse(time.RFC3339, r.URL.Query().Get("startDate"))
	endDate, _ := time.Parse(time.RFC3339, r.URL.Query().Get("endDate"))

	query := &queries.ListPaymentsQuery{
		TenantID:  tenantID,
		ClientID:  clientID,
		Status:    status,
		Method:    method,
		Page:      page,
		PageSize:  pageSize,
		StartDate: startDate,
		EndDate:   endDate,
		SortBy:    r.URL.Query().Get("sortBy"),
		SortOrder: r.URL.Query().Get("sortOrder"),
	}

	result, err := s.queryHandler.ListPayments(ctx, query)
	if err != nil {
		s.logger.New(ctx).Error("Failed to list payments", "error", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to list payments")
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}

func (s *PaymentService) getPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	paymentID := r.URL.Path[len("/api/v1/payments/"):]
	if paymentID == "" {
		s.writeError(w, http.StatusBadRequest, "payment ID is required")
		return
	}

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, http.StatusBadRequest, "tenantId is required")
		return
	}

	query := &queries.GetPaymentByIDQuery{
		PaymentID: paymentID,
		TenantID:  tenantID,
	}

	payment, err := s.queryHandler.GetPaymentByID(ctx, query)
	if err != nil {
		s.logger.New(ctx).Error("Failed to get payment", "error", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to get payment")
		return
	}

	if payment == nil {
		s.writeError(w, http.StatusNotFound, "Payment not found")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{"payment": payment})
}

func (s *PaymentService) processPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		InvoiceID   string  `json:"invoiceId"`
		ClientID    string  `json:"clientId"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		Method      string  `json:"method"`
		Provider    string  `json:"provider"`
		Reference   string  `json:"reference"`
		Description string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.InvoiceID == "" {
		s.writeError(w, http.StatusBadRequest, "invoiceId is required")
		return
	}
	if req.ClientID == "" {
		s.writeError(w, http.StatusBadRequest, "clientId is required")
		return
	}
	if req.Amount <= 0 {
		s.writeError(w, http.StatusBadRequest, "amount must be greater than zero")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	if tenantID == "" {
		s.writeError(w, http.StatusBadRequest, "tenantId is required")
		return
	}

	data := map[string]interface{}{
		"invoiceId":   req.InvoiceID,
		"clientId":    req.ClientID,
		"amount":      fmt.Sprintf("%.2f", req.Amount),
		"currency":    req.Currency,
		"method":      req.Method,
		"provider":    req.Provider,
		"reference":   req.Reference,
		"description": req.Description,
	}

	cmd := commands.NewCommand("createPayment", tenantID, "", userID, data)

	payment, err := s.paymentHandler.HandleCreatePayment(ctx, cmd)
	if err != nil {
		s.logger.New(ctx).Error("Failed to create payment", "error", err)
		s.writeErrorFromAppError(w, err)
		return
	}

	processCmd := commands.NewCommand("processPayment", tenantID, payment.ID.String(), userID, nil)

	payment, err = s.paymentHandler.HandleProcessPayment(ctx, processCmd)
	if err != nil {
		s.logger.New(ctx).Error("Failed to process payment", "error", err)
		s.writeErrorFromAppError(w, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":     "Payment processed",
		"paymentId":   payment.ID.String(),
		"status":      string(payment.Status),
		"amount":      payment.Amount.String(),
		"currency":    payment.Currency,
		"method":      string(payment.Method),
		"provider":    payment.Provider,
		"processedAt": payment.ProcessedAt,
	})
}

func (s *PaymentService) processRefund(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		PaymentID string  `json:"paymentId"`
		Amount    float64 `json:"amount"`
		Reason    string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PaymentID == "" {
		s.writeError(w, http.StatusBadRequest, "paymentId is required")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	if tenantID == "" {
		s.writeError(w, http.StatusBadRequest, "tenantId is required")
		return
	}

	data := map[string]interface{}{
		"reason": req.Reason,
	}
	if req.Amount > 0 {
		data["amount"] = fmt.Sprintf("%.2f", req.Amount)
	}

	cmd := commands.NewCommand("refundPayment", tenantID, req.PaymentID, userID, data)

	payment, err := s.paymentHandler.HandleRefundPayment(ctx, cmd)
	if err != nil {
		s.logger.New(ctx).Error("Failed to process refund", "error", err)
		s.writeErrorFromAppError(w, err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Refund processed",
		"paymentId": payment.ID.String(),
		"status":    string(payment.Status),
		"refunded":  true,
	})
}

func (s *PaymentService) processWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "stripe"
	}

	var result *commands.WebhookResult
	var err error

	switch provider {
	case "stripe":
		payload, signature, parseErr := s.webhookHandler.ParseStripeWebhook(r)
		if parseErr != nil {
			s.logger.New(ctx).Error("Failed to parse Stripe webhook", "error", parseErr)
			s.writeError(w, http.StatusBadRequest, "Invalid webhook payload")
			return
		}

		result, err = s.webhookHandler.HandleStripeWebhook(ctx, payload, signature)

	case "paypal":
		payload, headers, parseErr := s.webhookHandler.ParsePayPalWebhook(r)
		if parseErr != nil {
			s.logger.New(ctx).Error("Failed to parse PayPal webhook", "error", parseErr)
			s.writeError(w, http.StatusBadRequest, "Invalid webhook payload")
			return
		}

		result, err = s.webhookHandler.HandlePayPalWebhook(ctx, payload, headers)

	default:
		s.writeError(w, http.StatusBadRequest, "Invalid provider")
		return
	}

	if err != nil {
		s.logger.New(ctx).Error("Failed to process webhook", "provider", provider, "error", err)
		s.writeErrorFromAppError(w, err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   result.Success,
		"eventId":   result.EventID,
		"eventType": result.EventType,
		"paymentId": result.PaymentID,
	})
}

func (s *PaymentService) getPaymentMethods(w http.ResponseWriter, r *http.Request) {
	methods := []map[string]interface{}{
		{
			"id":        "credit_card",
			"name":      "Credit Card",
			"providers": []string{"stripe", "adyen"},
		},
		{
			"id":        "debit_card",
			"name":      "Debit Card",
			"providers": []string{"stripe"},
		},
		{
			"id":        "bank_transfer",
			"name":      "Bank Transfer",
			"providers": []string{"ach", "sepa"},
		},
		{
			"id":        "paypal",
			"name":      "PayPal",
			"providers": []string{"paypal"},
		},
		{
			"id":        "cash",
			"name":      "Cash",
			"providers": []string{},
		},
		{
			"id":        "check",
			"name":      "Check",
			"providers": []string{},
		},
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{"methods": methods})
}

func (s *PaymentService) getTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, http.StatusBadRequest, "tenantId is required")
		return
	}

	startDate, _ := time.Parse(time.RFC3339, r.URL.Query().Get("startDate"))
	endDate, _ := time.Parse(time.RFC3339, r.URL.Query().Get("endDate"))

	query := &queries.GetPaymentStatsQuery{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	stats, err := s.queryHandler.GetPaymentStats(ctx, query)
	if err != nil {
		s.logger.New(ctx).Error("Failed to get transactions", "error", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to get transactions")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": stats,
	})
}

func (s *PaymentService) getDailyReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, http.StatusBadRequest, "tenantId is required")
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	query := &queries.GetPaymentStatsQuery{
		TenantID:  tenantID,
		StartDate: parsedDate,
		EndDate:   parsedDate.Add(24 * time.Hour),
	}

	stats, err := s.queryHandler.GetPaymentStats(ctx, query)
	if err != nil {
		s.logger.New(ctx).Error("Failed to get daily report", "error", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to get daily report")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"report":            "daily",
		"date":              date,
		"totalTransactions": stats.TotalPayments,
		"totalVolume":       stats.TotalAmount,
		"totalRefunds":      stats.TotalRefunded,
		"pendingCount":      stats.PendingCount,
		"completedCount":    stats.CompletedCount,
		"failedCount":       stats.FailedCount,
		"refundedCount":     stats.RefundedCount,
	})
}

func (s *PaymentService) getSummaryReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, http.StatusBadRequest, "tenantId is required")
		return
	}

	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid startDate format")
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid endDate format")
			return
		}
	}

	query := &queries.GetPaymentStatsQuery{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	stats, err := s.queryHandler.GetPaymentStats(ctx, query)
	if err != nil {
		s.logger.New(ctx).Error("Failed to get summary report", "error", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to get summary report")
		return
	}

	successRate := 0.0
	if stats.TotalPayments > 0 {
		successRate = float64(stats.CompletedCount) / float64(stats.TotalPayments) * 100
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"report":            "summary",
		"period":            map[string]interface{}{"start": startDateStr, "end": endDateStr},
		"totalVolume":       stats.TotalAmount,
		"totalTransactions": stats.TotalPayments,
		"successRate":       successRate,
		"pendingCount":      stats.PendingCount,
		"completedCount":    stats.CompletedCount,
		"failedCount":       stats.FailedCount,
		"refundedCount":     stats.RefundedCount,
		"cancelledCount":    stats.CancelledCount,
	})
}

func (s *PaymentService) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.New(context.Background()).Error("Failed to encode JSON response", "error", err)
	}
}

func (s *PaymentService) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]interface{}{
		"error":   message,
		"status":  status,
		"success": false,
	})
}

func (s *PaymentService) writeErrorFromAppError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*errors.Error); ok {
		switch appErr.Code {
		case errors.CodeNotFound:
			s.writeError(w, http.StatusNotFound, appErr.Message)
		case errors.CodeInvalidArgument:
			s.writeError(w, http.StatusBadRequest, appErr.Message)
		case errors.CodeForbidden:
			s.writeError(w, http.StatusForbidden, appErr.Message)
		case errors.CodeUnauthorized:
			s.writeError(w, http.StatusUnauthorized, appErr.Message)
		default:
			s.writeError(w, http.StatusInternalServerError, appErr.Message)
		}
	} else {
		s.writeError(w, http.StatusInternalServerError, err.Error())
	}
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

	// Initialize MongoDB connection
	mongoDB, err := repository.NewMongoDB(cfg.MongoDB, log)
	if err != nil {
		log.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer mongoDB.Close(context.Background())

	// Initialize repositories
	paymentRepo := repository.NewMongoPaymentRepository(mongoDB, log)
	invoiceRepo := repository.NewMongoInvoiceRepository(mongoDB, log)
	eventStore := repository.NewEventStore(mongoDB, log)

	// Initialize read model store (using MongoDB for simplicity)
	readModelStore := repository.NewReadModelStore(mongoDB, "payment_read_models", log)

	// Initialize Redis connection
	redisClient, err := repository.NewRedis(cfg.Redis, log)
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// Initialize cache (using Redis)
	cache := repository.NewCache(redisClient, "payment_cache", log)

	// Initialize publisher (using NATS)
	natsConfig := messaging.NATSConfig{
		URLs:           cfg.NATS.URLs,
		Username:       cfg.NATS.Username,
		Password:       cfg.NATS.Password,
		Token:          cfg.NATS.Token,
		MaxReconnect:   cfg.NATS.MaxReconnect,
		ReconnectWait:  cfg.NATS.ReconnectWait,
		ConnectTimeout: cfg.NATS.ConnectTimeout,
		JetStream:      cfg.NATS.JetStream.Enabled,
		Domain:         cfg.NATS.JetStream.Domain,
		StreamPrefix:   cfg.NATS.JetStream.StreamPrefix,
	}
	publisher, err := messaging.NewPublisher(natsConfig, log)
	if err != nil {
		log.Error("Failed to connect to NATS", "error", err)
		os.Exit(1)
	}

	// Initialize processor registry
	processors := domain.NewProcessorRegistry()
	processors.Register("stripe", func(name string, config interface{}) (domain.PaymentProcessor, error) {
		return domain.NewStripeProcessor("stripe_key", "stripe_secret"), nil
	})
	processors.Register("paypal", func(name string, config interface{}) (domain.PaymentProcessor, error) {
		return domain.NewPayPalProcessor("paypal_client_id", "paypal_secret", "sandbox"), nil
	})

	// Initialize handlers
	paymentHandler := commands.NewPaymentCommandHandler(
		paymentRepo,
		invoiceRepo,
		eventStore,
		publisher,
		log,
		processors,
	)

	queryHandler := queries.NewPaymentQueryHandler(
		readModelStore,
		cache,
		log,
	)

	webhookHandler := commands.NewWebhookHandler(
		paymentRepo,
		invoiceRepo,
		publisher,
		log,
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
		os.Getenv("PAYPAL_WEBHOOK_ID"),
	)

	service := NewPaymentService(
		cfg,
		log,
		paymentHandler,
		queryHandler,
		webhookHandler,
		paymentRepo,
		invoiceRepo,
		publisher,
		processors,
	)

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
	return uuid.New().String()
}
