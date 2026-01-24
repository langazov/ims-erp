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

type ProductService struct {
	config *config.Config
	logger *logger.Logger
}

func NewProductService(cfg *config.Config, log *logger.Logger) *ProductService {
	return &ProductService{
		config: cfg,
		logger: log,
	}
}

func (s *ProductService) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readinessHandler)
	mux.HandleFunc("/live", s.livenessHandler)
	mux.HandleFunc("/metrics", s.metricsHandler)

	mux.HandleFunc("/api/v1/products", s.handleProducts)
	mux.HandleFunc("/api/v1/products/", s.handleProductByID)
	mux.HandleFunc("/api/v1/products/", s.handleProductVariants)
	mux.HandleFunc("/api/v1/products/", s.handleProductPricing)
	mux.HandleFunc("/api/v1/products/", s.handleProductInventory)
	mux.HandleFunc("/api/v1/products/", s.handleProductImages)
	mux.HandleFunc("/api/v1/products/search", s.handleSearch)
	mux.HandleFunc("/api/v1/products/categories", s.handleCategories)
	mux.HandleFunc("/api/v1/products/brands", s.handleBrands)
	mux.HandleFunc("/api/v1/products/report/valuation", s.handleValuationReport)

	return mux
}

func (s *ProductService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s", "service": "product-service"}`, time.Now().UTC())
}

func (s *ProductService) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ready", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *ProductService) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "alive", "timestamp": "%s"}`, time.Now().UTC())
}

func (s *ProductService) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "# Product Service Metrics\n")
	fmt.Fprintf(w, "product_service_up 1\n")
	fmt.Fprintf(w, "product_service_requests_total 0\n")
	fmt.Fprintf(w, "product_service_created_total 0\n")
	fmt.Fprintf(w, "product_service_active_total 0\n")
}

func (s *ProductService) handleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listProducts(w, r)
	case http.MethodPost:
		s.createProduct(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *ProductService) handleProductByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getProduct(w, r)
	case http.MethodPut:
		s.updateProduct(w, r)
	case http.MethodDelete:
		s.deleteProduct(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *ProductService) handleProductVariants(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.createVariant(w, r)
	} else if r.Method == http.MethodGet {
		s.listVariants(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *ProductService) handleProductPricing(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		s.updatePricing(w, r)
	} else if r.Method == http.MethodGet {
		s.getPricing(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *ProductService) handleProductInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getInventory(w, r)
	} else if r.Method == http.MethodPut {
		s.updateInventory(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *ProductService) handleProductImages(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.uploadImage(w, r)
	} else if r.Method == http.MethodDelete {
		s.deleteImage(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *ProductService) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.searchProducts(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *ProductService) handleCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listCategories(w, r)
	case http.MethodPost:
		s.createCategory(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *ProductService) handleBrands(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listBrands(w, r)
	case http.MethodPost:
		s.createBrand(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *ProductService) handleValuationReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getValuationReport(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *ProductService) listProducts(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 50)

	_ = tenantID
	_ = category
	_ = status
	_ = page
	_ = pageSize

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"products": [], "total": 0, "page": %d, "pageSize": %d}`, page, pageSize)
}

func (s *ProductService) createProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Product created", "id": "%s"}`, generateUUID())
}

func (s *ProductService) getProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("productId")
	_ = productID

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"product": null}`)
}

func (s *ProductService) updateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Product updated"}`)
}

func (s *ProductService) deleteProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Product deleted"}`)
}

func (s *ProductService) createVariant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Variant created"}`)
}

func (s *ProductService) listVariants(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"variants": []}`)
}

func (s *ProductService) updatePricing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Pricing updated"}`)
}

func (s *ProductService) getPricing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"pricing": null}`)
}

func (s *ProductService) getInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"inventory": null}`)
}

func (s *ProductService) updateInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Inventory updated"}`)
}

func (s *ProductService) uploadImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Image uploaded"}`)
}

func (s *ProductService) deleteImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Image deleted"}`)
}

func (s *ProductService) searchProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	_ = query

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"results": [], "total": 0}`)
}

func (s *ProductService) listCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"categories": []}`)
}

func (s *ProductService) createCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Category created"}`)
}

func (s *ProductService) listBrands(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"brands": []}`)
}

func (s *ProductService) createBrand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Brand created"}`)
}

func (s *ProductService) getValuationReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	_ = tenantID

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"report": "valuation", "generatedAt": "%s", "totalValue": 0, "byCategory": {}, "byWarehouse": {}}`, time.Now().UTC())
}

func main() {
	cfg, err := config.Load("", "product-service")
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

	service := NewProductService(cfg, log)
	mux := service.setupRoutes()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      mux,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		log.Info("Starting product service", "port", cfg.App.Port)
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
