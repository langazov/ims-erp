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

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/commands"
	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/queries"
	"github.com/ims-erp/system/pkg/errors"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/ims-erp/system/pkg/tracer"
)

type InvoiceService struct {
	config         *config.Config
	logger         *logger.Logger
	invoiceHandler *commands.InvoiceCommandHandler
	queryHandler   *queries.InvoiceQueryHandler
	invoiceRepo    commands.InvoiceRepository
	publisher      commands.Publisher
}

func NewInvoiceService(
	cfg *config.Config,
	log *logger.Logger,
	invoiceHandler *commands.InvoiceCommandHandler,
	queryHandler *queries.InvoiceQueryHandler,
	invoiceRepo commands.InvoiceRepository,
	publisher commands.Publisher,
) *InvoiceService {
	return &InvoiceService{
		config:         cfg,
		logger:         log,
		invoiceHandler: invoiceHandler,
		queryHandler:   queryHandler,
		invoiceRepo:    invoiceRepo,
		publisher:      publisher,
	}
}

func (s *InvoiceService) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readinessHandler)
	mux.HandleFunc("/live", s.livenessHandler)
	mux.HandleFunc("/metrics", s.metricsHandler)

	mux.HandleFunc("/api/v1/invoices", s.handleInvoices)
	mux.HandleFunc("/api/v1/invoices/", s.handleInvoiceOperations)
	mux.HandleFunc("/api/v1/invoices/report/outstanding", s.handleOutstandingReport)
	mux.HandleFunc("/api/v1/invoices/report/overdue", s.handleOverdueReport)
	mux.HandleFunc("/api/v1/invoices/report/summary", s.handleSummaryReport)

	return mux
}

func (s *InvoiceService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s", "service": "invoice-service"}`, time.Now().UTC())
}

func (s *InvoiceService) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ready", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *InvoiceService) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "alive", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *InvoiceService) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "# Invoice Service Metrics\n")
	fmt.Fprintf(w, "invoice_service_up 1\n")
	fmt.Fprintf(w, "invoice_service_requests_total 0\n")
	fmt.Fprintf(w, "invoice_service_created_total 0\n")
	fmt.Fprintf(w, "invoice_service_sent_total 0\n")
	fmt.Fprintf(w, "invoice_service_paid_total 0\n")
	fmt.Fprintf(w, "invoice_service_overdue_total 0\n")
}

func (s *InvoiceService) handleInvoices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listInvoices(w, r)
	case http.MethodPost:
		s.createInvoice(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleInvoiceOperations(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/invoices/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	invoiceID := parts[0]

	if len(parts) > 1 {
		switch parts[1] {
		case "lines":
			s.handleInvoiceLines(w, r, invoiceID)
		case "payments":
			s.handleInvoicePayments(w, r, invoiceID)
		case "send":
			s.handleInvoiceSend(w, r, invoiceID)
		case "pdf":
			s.handleInvoicePDF(w, r, invoiceID)
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getInvoice(w, r, invoiceID)
	case http.MethodPut, http.MethodPatch:
		s.updateInvoice(w, r, invoiceID)
	case http.MethodDelete:
		s.deleteInvoice(w, r, invoiceID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleInvoiceLines(w http.ResponseWriter, r *http.Request, invoiceID string) {
	if r.Method == http.MethodPost {
		s.addInvoiceLine(w, r, invoiceID)
	} else if r.Method == http.MethodDelete {
		s.removeInvoiceLine(w, r, invoiceID)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleInvoicePayments(w http.ResponseWriter, r *http.Request, invoiceID string) {
	if r.Method == http.MethodPost {
		s.recordPayment(w, r, invoiceID)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleInvoiceSend(w http.ResponseWriter, r *http.Request, invoiceID string) {
	if r.Method == http.MethodPost {
		s.sendInvoice(w, r, invoiceID)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleInvoicePDF(w http.ResponseWriter, r *http.Request, invoiceID string) {
	if r.Method == http.MethodGet {
		s.generatePDF(w, r, invoiceID)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) listInvoices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}

	clientID := r.URL.Query().Get("clientId")
	status := r.URL.Query().Get("status")
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 20)

	query := &queries.ListInvoicesQuery{
		TenantID: tenantID,
		ClientID: clientID,
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	}

	result, err := s.queryHandler.ListInvoices(ctx, query)
	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}

func (s *InvoiceService) createInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		TenantID    string                 `json:"tenantId"`
		ClientID    string                 `json:"clientId"`
		UserID      string                 `json:"userId"`
		Type        string                 `json:"type"`
		Currency    string                 `json:"currency"`
		PaymentTerm string                 `json:"paymentTerm"`
		IssueDate   string                 `json:"issueDate"`
		DueDate     string                 `json:"dueDate"`
		Notes       string                 `json:"notes"`
		Terms       string                 `json:"terms"`
		Data        map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, errors.InvalidArgument("invalid request body"))
		return
	}

	if req.TenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}
	if req.ClientID == "" {
		s.writeError(w, errors.InvalidArgument("clientId is required"))
		return
	}
	if req.UserID == "" {
		req.UserID = "system"
	}

	data := req.Data
	if data == nil {
		data = make(map[string]interface{})
	}
	data["clientId"] = req.ClientID
	data["type"] = req.Type
	data["currency"] = req.Currency
	data["paymentTerm"] = req.PaymentTerm
	data["issueDate"] = req.IssueDate
	data["dueDate"] = req.DueDate
	data["notes"] = req.Notes
	data["terms"] = req.Terms

	cmd := commands.NewCommand("createInvoice", req.TenantID, "", req.UserID, data)

	invoice, err := s.invoiceHandler.HandleCreateInvoice(ctx, cmd)
	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, invoice)
}

func (s *InvoiceService) getInvoice(w http.ResponseWriter, r *http.Request, invoiceID string) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}

	query := &queries.GetInvoiceByIDQuery{
		InvoiceID: invoiceID,
		TenantID:  tenantID,
	}

	invoice, err := s.queryHandler.GetInvoiceByID(ctx, query)
	if err != nil {
		s.writeError(w, err)
		return
	}

	if invoice == nil {
		s.writeError(w, errors.NotFound("invoice not found"))
		return
	}

	s.writeJSON(w, http.StatusOK, invoice)
}

func (s *InvoiceService) updateInvoice(w http.ResponseWriter, r *http.Request, invoiceID string) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}

	var req struct {
		UserID string                 `json:"userId"`
		Action string                 `json:"action"`
		Data   map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, errors.InvalidArgument("invalid request body"))
		return
	}

	if req.UserID == "" {
		req.UserID = "system"
	}

	cmd := commands.NewCommand("", tenantID, invoiceID, req.UserID, req.Data)

	var invoice *domain.Invoice
	var err error

	switch req.Action {
	case "finalize":
		cmd.Type = "finalizeInvoice"
		invoice, err = s.invoiceHandler.HandleFinalizeInvoice(ctx, cmd)
	case "void", "cancel":
		cmd.Type = "voidInvoice"
		invoice, err = s.invoiceHandler.HandleVoidInvoice(ctx, cmd)
	case "send":
		cmd.Type = "sendInvoice"
		invoice, err = s.invoiceHandler.HandleSendInvoice(ctx, cmd)
	default:
		s.writeError(w, errors.InvalidArgument("invalid action: must be 'finalize', 'void', 'cancel', or 'send'"))
		return
	}

	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeJSON(w, http.StatusOK, invoice)
}

func (s *InvoiceService) deleteInvoice(w http.ResponseWriter, r *http.Request, invoiceID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Invoice deleted"}`)
}

func (s *InvoiceService) addInvoiceLine(w http.ResponseWriter, r *http.Request, invoiceID string) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}

	var req struct {
		UserID      string                 `json:"userId"`
		Description string                 `json:"description"`
		Quantity    string                 `json:"quantity"`
		UnitPrice   string                 `json:"unitPrice"`
		Discount    string                 `json:"discount"`
		TaxRate     string                 `json:"taxRate"`
		ProductID   string                 `json:"productId"`
		SortOrder   int                    `json:"sortOrder"`
		Data        map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, errors.InvalidArgument("invalid request body"))
		return
	}

	if req.UserID == "" {
		req.UserID = "system"
	}

	data := req.Data
	if data == nil {
		data = make(map[string]interface{})
	}
	data["description"] = req.Description
	data["quantity"] = req.Quantity
	data["unitPrice"] = req.UnitPrice
	data["discount"] = req.Discount
	data["taxRate"] = req.TaxRate
	data["productId"] = req.ProductID
	data["sortOrder"] = float64(req.SortOrder)

	cmd := commands.NewCommand("addLineItem", tenantID, invoiceID, req.UserID, data)

	invoice, err := s.invoiceHandler.HandleAddLineItem(ctx, cmd)
	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, invoice)
}

func (s *InvoiceService) removeInvoiceLine(w http.ResponseWriter, r *http.Request, invoiceID string) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}

	lineID := r.URL.Query().Get("lineId")
	if lineID == "" {
		var req struct {
			LineID string `json:"lineId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			lineID = req.LineID
		}
	}

	if lineID == "" {
		s.writeError(w, errors.InvalidArgument("lineId is required"))
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = "system"
	}

	data := map[string]interface{}{
		"lineId": lineID,
	}

	cmd := commands.NewCommand("removeLineItem", tenantID, invoiceID, userID, data)

	invoice, err := s.invoiceHandler.HandleRemoveLineItem(ctx, cmd)
	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeJSON(w, http.StatusOK, invoice)
}

func (s *InvoiceService) recordPayment(w http.ResponseWriter, r *http.Request, invoiceID string) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}

	var req struct {
		UserID        string `json:"userId"`
		Amount        string `json:"amount"`
		PaymentMethod string `json:"paymentMethod"`
		Reference     string `json:"reference"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, errors.InvalidArgument("invalid request body"))
		return
	}

	if req.Amount == "" {
		s.writeError(w, errors.InvalidArgument("amount is required"))
		return
	}

	if req.UserID == "" {
		req.UserID = "system"
	}

	data := map[string]interface{}{
		"amount":        req.Amount,
		"paymentMethod": req.PaymentMethod,
		"reference":     req.Reference,
	}

	cmd := commands.NewCommand("recordPayment", tenantID, invoiceID, req.UserID, data)

	invoice, err := s.invoiceHandler.HandleRecordPayment(ctx, cmd)
	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, invoice)
}

func (s *InvoiceService) sendInvoice(w http.ResponseWriter, r *http.Request, invoiceID string) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}

	var req struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.UserID = r.URL.Query().Get("userId")
	}

	if req.UserID == "" {
		req.UserID = "system"
	}

	data := make(map[string]interface{})

	cmd := commands.NewCommand("sendInvoice", tenantID, invoiceID, req.UserID, data)

	invoice, err := s.invoiceHandler.HandleSendInvoice(ctx, cmd)
	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeJSON(w, http.StatusOK, invoice)
}

func (s *InvoiceService) generatePDF(w http.ResponseWriter, r *http.Request, invoiceID string) {
	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("%PDF-1.4 Invoice PDF Placeholder"))
}

func (s *InvoiceService) handleOutstandingReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}

	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 20)

	query := &queries.GetOverdueInvoicesQuery{
		TenantID: tenantID,
		Page:     page,
		PageSize: pageSize,
	}

	result, err := s.queryHandler.GetOverdueInvoices(ctx, query)
	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}

func (s *InvoiceService) handleOverdueReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}

	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 20)

	query := &queries.GetOverdueInvoicesQuery{
		TenantID: tenantID,
		Page:     page,
		PageSize: pageSize,
	}

	result, err := s.queryHandler.GetOverdueInvoices(ctx, query)
	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}

func (s *InvoiceService) handleSummaryReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.URL.Query().Get("tenantId")
	if tenantID == "" {
		s.writeError(w, errors.InvalidArgument("tenantId is required"))
		return
	}

	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	query := &queries.GetInvoiceStatsQuery{
		TenantID: tenantID,
	}

	if startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			query.StartDate = startDate
		}
	}

	if endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			query.EndDate = endDate
		}
	}

	stats, err := s.queryHandler.GetInvoiceStats(ctx, query)
	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeJSON(w, http.StatusOK, stats)
}

func (s *InvoiceService) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response", "error", err)
	}
}

func (s *InvoiceService) writeError(w http.ResponseWriter, err error) {
	var statusCode int
	var errorResponse map[string]interface{}

	if appErr, ok := err.(*errors.Error); ok {
		statusCode = appErr.StatusCode()
		errorResponse = map[string]interface{}{
			"error":   appErr.Code,
			"message": appErr.Message,
		}
		if appErr.Details != nil {
			errorResponse["details"] = appErr.Details
		}
	} else {
		statusCode = http.StatusInternalServerError
		errorResponse = map[string]interface{}{
			"error":   "INTERNAL_ERROR",
			"message": err.Error(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

func main() {
	cfg, err := config.Load("", "invoice-service")
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

	var invoiceRepo commands.InvoiceRepository
	var publisher commands.Publisher

	invoiceCounter := &invoiceNumberCounter{}

	invoiceHandler := commands.NewInvoiceCommandHandler(
		invoiceRepo,
		nil,
		publisher,
		log,
		invoiceCounter,
	)

	queryHandler := queries.NewInvoiceQueryHandler(
		nil,
		nil,
		log,
	)

	service := NewInvoiceService(cfg, log, invoiceHandler, queryHandler, invoiceRepo, publisher)
	mux := service.setupRoutes()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      mux,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		log.Info("Starting invoice service", "port", cfg.App.Port)
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

type invoiceNumberCounter struct{}

func (c *invoiceNumberCounter) GetNextInvoiceNumber(ctx context.Context, tenantID uuid.UUID, year int) (string, error) {
	return fmt.Sprintf("INV-%d-%06d", year, time.Now().UnixNano()%1000000), nil
}
