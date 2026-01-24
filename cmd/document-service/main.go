package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/pkg/logger"
)

var (
	Version   = "dev"
	BuildDate = time.Now().Format(time.RFC3339)
)

type Config struct {
	ServiceName      string        `mapstructure:"SERVICE_NAME"`
	ServicePort      int           `mapstructure:"SERVICE_PORT"`
	MongoURI         string        `mapstructure:"MONGO_URI"`
	MongoDatabase    string        `mapstructure:"MONGO_DATABASE"`
	RedisAddr        string        `mapstructure:"REDIS_ADDR"`
	RedisPassword    string        `mapstructure:"REDIS_PASSWORD"`
	MinIOEndpoint    string        `mapstructure:"MINIO_ENDPOINT"`
	MinIOAccessKey   string        `mapstructure:"MINIO_ACCESS_KEY"`
	MinIOSecretKey   string        `mapstructure:"MINIO_SECRET_KEY"`
	MinIOUseSSL      bool          `mapstructure:"MINIO_USE_SSL"`
	ElasticsearchURL string        `mapstructure:"ELASTICSEARCH_URL"`
	MaxFileSize      int64         `mapstructure:"MAX_FILE_SIZE"`
	PresignedExpiry  time.Duration `mapstructure:"PRESIGNED_EXPIRY"`
	LogLevel         string        `mapstructure:"LOG_LEVEL"`
}

type Service struct {
	config   *Config
	logger   *logger.Logger
	mongo    *mongo.Client
	mongoDb  *mongo.Database
	redis    redis.UniversalClient
	minio    *minio.Client
	esClient *http.Client
	repo     domain.DocumentRepository
	storage  domain.StorageService
	search   domain.SearchService
}

type UploadRequest struct {
	Type       string    `json:"type"`
	Tags       []string  `json:"tags"`
	UploadedBy uuid.UUID `json:"uploadedBy"`
}

type UploadResponse struct {
	DocumentID      uuid.UUID         `json:"documentId"`
	PresignedURL    string            `json:"presignedUrl"`
	ObjectKey       string            `json:"objectKey"`
	RequiredHeaders map[string]string `json:"requiredHeaders"`
}

type SearchRequest struct {
	Query    string   `json:"query"`
	Type     string   `json:"type,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	DateFrom string   `json:"dateFrom,omitempty"`
	DateTo   string   `json:"dateTo,omitempty"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
}

func NewConfig() *Config {
	return &Config{
		ServiceName:     "document-service",
		ServicePort:     8080,
		MongoURI:        "mongodb://localhost:27017",
		MongoDatabase:   "erp_documents",
		RedisAddr:       "localhost:6379",
		MinIOEndpoint:   "localhost:9000",
		MaxFileSize:     50 * 1024 * 1024,
		PresignedExpiry: 1 * time.Hour,
		LogLevel:        "info",
	}
}

func NewService(cfg *Config) (*Service, error) {
	logConfig := logger.Config{
		Level:       cfg.LogLevel,
		ServiceName: cfg.ServiceName,
	}
	log, err := logger.New(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	svc := &Service{
		config:   cfg,
		logger:   log,
		esClient: &http.Client{Timeout: 10 * time.Second},
	}

	if err := svc.connectMongo(); err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := svc.connectRedis(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	if err := svc.connectMinIO(); err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO: %w", err)
	}

	svc.repo = NewMongoDocumentRepository(svc.mongoDb)
	svc.storage = NewMinIOStorageService(svc.minio)
	svc.search = NewElasticsearchService(svc.esClient, cfg.ElasticsearchURL)

	return svc, nil
}

func (s *Service) connectMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(s.config.MongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	s.mongo = client
	s.mongoDb = client.Database(s.config.MongoDatabase)
	s.logger.Info("Connected to MongoDB")
	return nil
}

func (s *Service) connectRedis() error {
	s.redis = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{s.config.RedisAddr},
		Password: s.config.RedisPassword,
		DB:       0,
	})

	ctx := context.Background()
	if _, err := s.redis.Ping(ctx).Result(); err != nil {
		return err
	}

	s.logger.Info("Connected to Redis")
	return nil
}

func (s *Service) connectMinIO() error {
	minioClient, err := minio.New(s.config.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.config.MinIOAccessKey, s.config.MinIOSecretKey, ""),
		Secure: s.config.MinIOUseSSL,
	})
	if err != nil {
		return err
	}

	s.minio = minioClient
	s.logger.Info("Connected to MinIO")
	return nil
}

func (s *Service) Start() error {
	router := mux.NewRouter()

	s.setupMiddleware(router)
	s.setupRoutes(router)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.ServicePort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		s.logger.Info("Starting document-service", "port", s.config.ServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	return nil
}

func (s *Service) setupMiddleware(router *mux.Router) {
	router.Use(loggerMiddleware)
	router.Use(corsMiddleware)
	router.Use(requestIDMiddleware)
}

func (s *Service) setupRoutes(router *mux.Router) {
	router.HandleFunc("/health", s.healthHandler).Methods("GET")
	router.HandleFunc("/ready", s.readyHandler).Methods("GET")
	router.HandleFunc("/metrics", s.metricsHandler).Methods("GET")

	api := router.PathPrefix("/api/v1/documents").Subrouter()

	api.HandleFunc("/upload", s.initiateUploadHandler).Methods("POST")
	api.HandleFunc("/multipart/start", s.startMultipartUploadHandler).Methods("POST")
	api.HandleFunc("/multipart/{uploadId}/part", s.uploadPartHandler).Methods("PUT")
	api.HandleFunc("/multipart/{uploadId}/complete", s.completeMultipartUploadHandler).Methods("POST")
	api.HandleFunc("", s.createDocumentHandler).Methods("POST")
	api.HandleFunc("", s.listDocumentsHandler).Methods("GET")
	api.HandleFunc("/{id}", s.getDocumentHandler).Methods("GET")
	api.HandleFunc("/{id}", s.updateDocumentHandler).Methods("PUT")
	api.HandleFunc("/{id}", s.deleteDocumentHandler).Methods("DELETE")
	api.HandleFunc("/{id}/download", s.downloadDocumentHandler).Methods("GET")
	api.HandleFunc("/{id}/thumbnail", s.getThumbnailHandler).Methods("GET")
	api.HandleFunc("/{id}/presigned-url", s.getPresignedURLHandler).Methods("GET")
	api.HandleFunc("/{id}/tags", s.updateTagsHandler).Methods("PUT")
	api.HandleFunc("/{id}/reprocess", s.reprocessHandler).Methods("POST")

	api.HandleFunc("/search", s.searchDocumentsHandler).Methods("POST")
	api.HandleFunc("/search/suggest", s.suggestHandler).Methods("GET")
}

func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   Version,
	})
}

func (s *Service) readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	ready := true
	status := "ready"

	if err := s.mongo.Ping(ctx, nil); err != nil {
		ready = false
		status = "MongoDB not ready"
	}

	if err := s.redis.Ping(ctx).Err(); err != nil {
		ready = false
		status = "Redis not ready"
	}

	if ready {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(map[string]string{
		"status": status,
	})
}

func (s *Service) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("# Prometheus metrics endpoint\n"))
}

func (s *Service) initiateUploadHandler(w http.ResponseWriter, r *http.Request) {
	var req UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tenantID := getTenantID(r)
	docID := uuid.New()
	objectKey := s.generateObjectKey(tenantID, req.Type, docID)

	presignedURL, err := s.storage.GetPresignedUploadURL(
		r.Context(),
		tenantID.String(),
		objectKey,
		r.Header.Get("Content-Type"),
		s.config.PresignedExpiry,
	)
	if err != nil {
		s.logger.Error("Failed to generate presigned URL", "error", err)
		http.Error(w, "Failed to generate upload URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadResponse{
		DocumentID:   docID,
		PresignedURL: presignedURL,
		ObjectKey:    objectKey,
	})
}

func (s *Service) createDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var doc domain.Document
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	doc.TenantID = getTenantID(r)
	doc.ID = uuid.New()
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	doc.ProcessingStatus = domain.ProcessingStatusPending

	if err := s.repo.Create(r.Context(), &doc); err != nil {
		s.logger.Error("Failed to create document", "error", err)
		http.Error(w, "Failed to create document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

func (s *Service) listDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)

	filter := domain.DocumentFilter{
		TenantID: tenantID,
		Page:     1,
		PageSize: 20,
	}

	if docType := r.URL.Query().Get("type"); docType != "" {
		filter.Type = domain.DocumentType(docType)
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = domain.ProcessingStatus(status)
	}

	docs, total, err := s.repo.List(r.Context(), filter)
	if err != nil {
		s.logger.Error("Failed to list documents", "error", err)
		http.Error(w, "Failed to list documents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": docs,
		"total":     total,
		"page":      filter.Page,
		"pageSize":  filter.PageSize,
	})
}

func (s *Service) getDocumentHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	docID := getIDParam(r)

	doc, err := s.repo.GetByID(r.Context(), tenantID, docID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		s.logger.Error("Failed to get document", "error", err)
		http.Error(w, "Failed to get document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (s *Service) updateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	docID := getIDParam(r)

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	doc, err := s.repo.GetByID(r.Context(), tenantID, docID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		s.logger.Error("Failed to get document", "error", err)
		http.Error(w, "Failed to get document", http.StatusInternalServerError)
		return
	}

	if tags, ok := updates["tags"].([]interface{}); ok {
		doc.Tags = make([]string, len(tags))
		for i, t := range tags {
			doc.Tags[i] = t.(string)
		}
	}
	doc.UpdatedAt = time.Now()

	if err := s.repo.Update(r.Context(), doc); err != nil {
		s.logger.Error("Failed to update document", "error", err)
		http.Error(w, "Failed to update document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (s *Service) deleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	docID := getIDParam(r)

	doc, err := s.repo.GetByID(r.Context(), tenantID, docID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		s.logger.Error("Failed to get document", "error", err)
		http.Error(w, "Failed to get document", http.StatusInternalServerError)
		return
	}

	if err := s.storage.Delete(r.Context(), doc.Bucket, doc.ObjectKey); err != nil {
		s.logger.Error("Failed to delete from storage", "error", err)
	}

	if err := s.repo.Delete(r.Context(), tenantID, docID); err != nil {
		s.logger.Error("Failed to delete document", "error", err)
		http.Error(w, "Failed to delete document", http.StatusInternalServerError)
		return
	}

	s.search.DeleteFromIndex(r.Context(), tenantID, docID)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) downloadDocumentHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	docID := getIDParam(r)

	doc, err := s.repo.GetByID(r.Context(), tenantID, docID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		s.logger.Error("Failed to get document", "error", err)
		http.Error(w, "Failed to get document", http.StatusInternalServerError)
		return
	}

	data, err := s.storage.Download(r.Context(), doc.Bucket, doc.ObjectKey)
	if err != nil {
		s.logger.Error("Failed to download document", "error", err)
		http.Error(w, "Failed to download document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", doc.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", doc.FileName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", doc.Size))
	w.Write(data)
}

func (s *Service) getThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	docID := getIDParam(r)

	doc, err := s.repo.GetByID(r.Context(), tenantID, docID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	if doc.ThumbnailKey == "" {
		http.Error(w, "No thumbnail available", http.StatusNotFound)
		return
	}

	data, err := s.storage.Download(r.Context(), doc.Bucket, doc.ThumbnailKey)
	if err != nil {
		http.Error(w, "Failed to get thumbnail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(data)
}

func (s *Service) getPresignedURLHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	docID := getIDParam(r)

	doc, err := s.repo.GetByID(r.Context(), tenantID, docID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	downloadURL, err := s.storage.GetPresignedDownloadURL(
		r.Context(),
		doc.Bucket,
		doc.ObjectKey,
		15*time.Minute,
	)
	if err != nil {
		s.logger.Error("Failed to generate presigned URL", "error", err)
		http.Error(w, "Failed to generate URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": downloadURL})
}

func (s *Service) updateTagsHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	docID := getIDParam(r)

	var req struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	doc, err := s.repo.GetByID(r.Context(), tenantID, docID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	doc.Tags = req.Tags
	doc.UpdatedAt = time.Now()

	if err := s.repo.Update(r.Context(), doc); err != nil {
		s.logger.Error("Failed to update tags", "error", err)
		http.Error(w, "Failed to update tags", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (s *Service) reprocessHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	docID := getIDParam(r)

	doc, err := s.repo.GetByID(r.Context(), tenantID, docID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	doc.ProcessingStatus = domain.ProcessingStatusPending
	doc.UpdatedAt = time.Now()

	if err := s.repo.Update(r.Context(), doc); err != nil {
		s.logger.Error("Failed to reprocess document", "error", err)
		http.Error(w, "Failed to reprocess document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
}

func (s *Service) searchDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	tenantID := getTenantID(r)
	filters := make(map[string]interface{})
	if req.Type != "" {
		filters["type"] = req.Type
	}
	if len(req.Tags) > 0 {
		filters["tags"] = req.Tags
	}

	results, err := s.search.Search(r.Context(), tenantID, req.Query, filters)
	if err != nil {
		s.logger.Error("Search failed", "error", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results":  results,
		"query":    req.Query,
		"page":     req.Page,
		"pageSize": req.PageSize,
	})
}

func (s *Service) suggestHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	prefix := r.URL.Query().Get("prefix")

	suggestions, err := s.search.Suggest(r.Context(), tenantID, prefix)
	if err != nil {
		s.logger.Error("Suggest failed", "error", err)
		http.Error(w, "Suggest failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"suggestions": suggestions})
}

func (s *Service) startMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Service) uploadPartHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Service) completeMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Service) generateObjectKey(tenantID uuid.UUID, docType string, docID uuid.UUID) string {
	now := time.Now()
	return fmt.Sprintf("%s/%s/%s/%s/%s",
		tenantID.String(),
		docType,
		now.Format("2006"),
		now.Format("01"),
		docID.String(),
	)
}

func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func getTenantID(r *http.Request) uuid.UUID {
	return uuid.MustParse(r.Header.Get("X-Tenant-ID"))
}

func getIDParam(r *http.Request) uuid.UUID {
	vars := mux.Vars(r)
	return uuid.MustParse(vars["id"])
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID, X-Request-ID")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

type MongoDocumentRepository struct {
	collection *mongo.Collection
}

func NewMongoDocumentRepository(db *mongo.Database) *MongoDocumentRepository {
	return &MongoDocumentRepository{
		collection: db.Collection("documents"),
	}
}

func (r *MongoDocumentRepository) Create(ctx context.Context, doc *domain.Document) error {
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *MongoDocumentRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Document, error) {
	var doc domain.Document
	err := r.collection.FindOne(ctx, bson.M{
		"_id":      id,
		"tenantId": tenantID,
	}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *MongoDocumentRepository) Update(ctx context.Context, doc *domain.Document) error {
	_, err := r.collection.ReplaceOne(ctx, bson.M{
		"_id":      doc.ID,
		"tenantId": doc.TenantID,
	}, doc)
	return err
}

func (r *MongoDocumentRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{
		"_id":      id,
		"tenantId": tenantID,
	})
	return err
}

func (r *MongoDocumentRepository) List(ctx context.Context, filter domain.DocumentFilter) ([]domain.Document, int64, error) {
	filterCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	query := bson.M{"tenantId": filter.TenantID}
	if filter.Type != "" {
		query["type"] = filter.Type
	}
	if filter.Status != "" {
		query["processingStatus"] = filter.Status
	}

	opts := options.Find().
		SetSkip(int64((filter.Page - 1) * filter.PageSize)).
		SetLimit(int64(filter.PageSize)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(filterCtx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(filterCtx)

	var docs []domain.Document
	if err := cursor.All(filterCtx, &docs); err != nil {
		return nil, 0, err
	}

	total, _ := r.collection.CountDocuments(ctx, query)
	return docs, total, nil
}

func (r *MongoDocumentRepository) GetByChecksum(ctx context.Context, tenantID uuid.UUID, checksum string) (*domain.Document, error) {
	var doc domain.Document
	err := r.collection.FindOne(ctx, bson.M{
		"checksum": checksum,
		"tenantId": tenantID,
	}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

type MinIOStorageService struct {
	client *minio.Client
}

func NewMinIOStorageService(client *minio.Client) *MinIOStorageService {
	return &MinIOStorageService{client: client}
}

func (s *MinIOStorageService) Upload(ctx context.Context, bucket, objectKey string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, bucket, objectKey, io.NopCloser(io.LimitReader(
		&byteReader{data: data}, int64(len(data)))), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *MinIOStorageService) Download(ctx context.Context, bucket, objectKey string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func (s *MinIOStorageService) Delete(ctx context.Context, bucket, objectKey string) error {
	return s.client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{})
}

func (s *MinIOStorageService) GetPresignedUploadURL(ctx context.Context, bucket, objectKey, contentType string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expiry)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (s *MinIOStorageService) GetPresignedDownloadURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, bucket, objectKey, expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (s *MinIOStorageService) BucketExists(ctx context.Context, bucket string) (bool, error) {
	return s.client.BucketExists(ctx, bucket)
}

func (s *MinIOStorageService) CreateBucket(ctx context.Context, bucket string) error {
	return s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
}

type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

type ProcessingService struct{}

func NewProcessingService() *ProcessingService {
	return &ProcessingService{}
}

func (s *ProcessingService) ProcessDocument(ctx context.Context, doc *domain.Document, data []byte) (*domain.Document, error) {
	text, err := s.ExtractText(ctx, data, doc.MimeType)
	if err != nil {
		doc.ProcessingStatus = domain.ProcessingStatusFailed
		return doc, err
	}
	doc.ExtractedText = text
	doc.ExtractedMetadata = s.ExtractMetadata(ctx, doc.Type, text)
	doc.ProcessingStatus = domain.ProcessingStatusCompleted
	return doc, nil
}

func (s *ProcessingService) ExtractText(ctx context.Context, data []byte, mimeType string) (string, error) {
	return "", nil
}

func (s *ProcessingService) ExtractMetadata(ctx context.Context, docType domain.DocumentType, text string) domain.DocumentMetadata {
	return domain.DocumentMetadata{}
}

func (s *ProcessingService) GenerateThumbnail(ctx context.Context, data []byte, mimeType string) ([]byte, error) {
	return nil, nil
}

type ElasticsearchService struct {
	client *http.Client
	url    string
}

func NewElasticsearchService(client *http.Client, url string) *ElasticsearchService {
	return &ElasticsearchService{client: client, url: url}
}

func (s *ElasticsearchService) IndexDocument(ctx context.Context, doc *domain.Document) error {
	return nil
}

func (s *ElasticsearchService) DeleteFromIndex(ctx context.Context, tenantID, id uuid.UUID) error {
	return nil
}

func (s *ElasticsearchService) Search(ctx context.Context, tenantID uuid.UUID, query string, filters map[string]interface{}) ([]domain.SearchResult, error) {
	return []domain.SearchResult{}, nil
}

func (s *ElasticsearchService) Suggest(ctx context.Context, tenantID uuid.UUID, prefix string) ([]string, error) {
	return []string{}, nil
}

func main() {
	cfg := NewConfig()
	cfg.ServicePort = 8086

	svc, err := NewService(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create service: %v\n", err)
		os.Exit(1)
	}

	if err := svc.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Service failed: %v\n", err)
		os.Exit(1)
	}
}

var _ = url.Parse
