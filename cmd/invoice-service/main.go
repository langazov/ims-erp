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

type InvoiceService struct {
	config *config.Config
	logger *logger.Logger
}

func NewInvoiceService(cfg *config.Config, log *logger.Logger) *InvoiceService {
	return &InvoiceService{
		config: cfg,
		logger: log,
	}
}

func (s *InvoiceService) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readinessHandler)
	mux.HandleFunc("/live", s.livenessHandler)
	mux.HandleFunc("/metrics", s.metricsHandler)

	mux.HandleFunc("/api/v1/invoices", s.handleInvoices)
	mux.HandleFunc("/api/v1/invoices/", s.handleInvoiceByID)
	mux.HandleFunc("/api/v1/invoices/", s.handleInvoiceLines)
	mux.HandleFunc("/api/v1/invoices/", s.handleInvoicePayments)
	mux.HandleFunc("/api/v1/invoices/", s.handleInvoiceSend)
	mux.HandleFunc("/api/v1/invoices/", s.handleInvoicePDF)
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

func (s *InvoiceService) handleInvoiceByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getInvoice(w, r)
	case http.MethodPut:
		s.updateInvoice(w, r)
	case http.MethodDelete:
		s.deleteInvoice(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleInvoiceLines(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.addInvoiceLine(w, r)
	} else if r.Method == http.MethodDelete {
		s.removeInvoiceLine(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleInvoicePayments(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.recordPayment(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleInvoiceSend(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.sendInvoice(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleInvoicePDF(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.generatePDF(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleOutstandingReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getOutstandingReport(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleOverdueReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getOverdueReport(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) handleSummaryReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getSummaryReport(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *InvoiceService) listInvoices(w http.ResponseWriter, r *http.Request) {
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
	fmt.Fprintf(w, `{"invoices": [], "total": 0, "page": %d, "pageSize": %d}`, page, pageSize)
}

func (s *InvoiceService) createInvoice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Invoice created", "id": "%s"}`, generateUUID())
}

func (s *InvoiceService) getInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID := r.URL.Query().Get("invoiceId")
	_ = invoiceID

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"invoice": null}`)
}

func (s *InvoiceService) updateInvoice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Invoice updated"}`)
}

func (s *InvoiceService) deleteInvoice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Invoice deleted"}`)
}

func (s *InvoiceService) addInvoiceLine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Line added"}`)
}

func (s *InvoiceService) removeInvoiceLine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Line removed"}`)
}

func (s *InvoiceService) recordPayment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Payment recorded"}`)
}

func (s *InvoiceService) sendInvoice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Invoice sent"}`)
}

func (s *InvoiceService) generatePDF(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("%PDF-1.4 Invoice PDF Placeholder"))
}

func (s *InvoiceService) getOutstandingReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	_ = tenantID

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"report": "outstanding", "generatedAt": "%s", "totalOutstanding": 0, "invoices": []}`, time.Now().UTC())
}

func (s *InvoiceService) getOverdueReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	_ = tenantID

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"report": "overdue", "generatedAt": "%s", "totalOverdue": 0, "invoices": []}`, time.Now().UTC())
}

func (s *InvoiceService) getSummaryReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	_ = tenantID
	_ = startDate
	_ = endDate

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"report": "summary", "generatedAt": "%s", "totalInvoiced": 0, "totalPaid": 0, "totalOutstanding": 0, "totalOverdue": 0}`, time.Now().UTC())
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

	service := NewInvoiceService(cfg, log)
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

func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
