package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type DocumentQueryHandler struct {
	docRepo       domain.DocumentRepository
	searchService domain.SearchService
	cache         *repository.Cache
	logger        *logger.Logger
	tracer        trace.Tracer
}

func NewDocumentQueryHandler(
	docRepo domain.DocumentRepository,
	searchService domain.SearchService,
	cache *repository.Cache,
	log *logger.Logger,
) *DocumentQueryHandler {
	return &DocumentQueryHandler{
		docRepo:       docRepo,
		searchService: searchService,
		cache:         cache,
		logger:        log,
		tracer:        otel.Tracer("document-query-handler"),
	}
}

// Query types
type GetDocumentByIDQuery struct {
	DocumentID string
	TenantID   string
}

type ListDocumentsQuery struct {
	TenantID  string
	Page      int
	PageSize  int
	Type      string
	Status    string
	Tags      []string
	Search    string
	DateFrom  *time.Time
	DateTo    *time.Time
	SortBy    string
	SortOrder string
}

type SearchDocumentsQuery struct {
	TenantID string
	Query    string
	Type     string
	Tags     []string
	DateFrom *time.Time
	DateTo   *time.Time
	Page     int
	PageSize int
}

type GetDocumentDownloadURLQuery struct {
	DocumentID    string
	TenantID      string
	ExpiryMinutes int
}

type GetDocumentsByTypeQuery struct {
	TenantID string
	Type     string
	Page     int
	PageSize int
}

// Result types
type DocumentSummary struct {
	ID               string    `json:"id" bson:"_id"`
	TenantID         string    `json:"tenantId" bson:"tenantId"`
	Type             string    `json:"type" bson:"type"`
	FileName         string    `json:"fileName" bson:"fileName"`
	MimeType         string    `json:"mimeType" bson:"mimeType"`
	Size             int64     `json:"size" bson:"size"`
	Checksum         string    `json:"checksum" bson:"checksum"`
	ProcessingStatus string    `json:"processingStatus" bson:"processingStatus"`
	PageCount        int       `json:"pageCount" bson:"pageCount"`
	Tags             []string  `json:"tags" bson:"tags"`
	UploadedBy       string    `json:"uploadedBy" bson:"uploadedBy"`
	CreatedAt        time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt" bson:"updatedAt"`
}

type DocumentDetail struct {
	ID               string                  `json:"id" bson:"_id"`
	TenantID         string                  `json:"tenantId" bson:"tenantId"`
	Type             string                  `json:"type" bson:"type"`
	FileName         string                  `json:"fileName" bson:"fileName"`
	MimeType         string                  `json:"mimeType" bson:"mimeType"`
	Size             int64                   `json:"size" bson:"size"`
	Checksum         string                  `json:"checksum" bson:"checksum"`
	Bucket           string                  `json:"bucket" bson:"bucket"`
	ObjectKey        string                  `json:"objectKey" bson:"objectKey"`
	VersionID        string                  `json:"versionId" bson:"versionId"`
	ProcessingStatus string                  `json:"processingStatus" bson:"processingStatus"`
	ExtractedText    string                  `json:"extractedText" bson:"extractedText"`
	ThumbnailKey     string                  `json:"thumbnailKey" bson:"thumbnailKey"`
	PageCount        int                     `json:"pageCount" bson:"pageCount"`
	Metadata         domain.DocumentMetadata `json:"metadata" bson:"metadata"`
	Tags             []string                `json:"tags" bson:"tags"`
	UploadedBy       string                  `json:"uploadedBy" bson:"uploadedBy"`
	CreatedAt        time.Time               `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time               `json:"updatedAt" bson:"updatedAt"`
}

type ListDocumentsResult struct {
	Documents  []DocumentSummary `json:"documents"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int               `json:"totalPages"`
}

type SearchDocumentsResult struct {
	Documents []DocumentSearchResult `json:"documents"`
	Total     int64                  `json:"total"`
	Page      int                    `json:"page"`
	PageSize  int                    `json:"pageSize"`
}

type DocumentSearchResult struct {
	ID         string                 `json:"id"`
	FileName   string                 `json:"fileName"`
	Type       string                 `json:"type"`
	Score      float64                `json:"score"`
	Highlights map[string][]string    `json:"highlights"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type DocumentDownloadURLResult struct {
	URL       string        `json:"url"`
	ExpiresIn time.Duration `json:"expiresIn"`
}

// GetDocumentByID retrieves a single document by ID
func (h *DocumentQueryHandler) GetDocumentByID(ctx context.Context, query *GetDocumentByIDQuery) (*DocumentDetail, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_document_by_id",
		trace.WithAttributes(
			attribute.String("document_id", query.DocumentID),
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("document:detail:%s", query.DocumentID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var doc DocumentDetail
		if err := json.Unmarshal(cached, &doc); err == nil {
			return &doc, nil
		}
	}

	tenantID, err := uuid.Parse(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	docID, err := uuid.Parse(query.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}

	doc, err := h.docRepo.GetByID(ctx, tenantID, docID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	result := mapDocumentToDetail(doc)

	if data, err := json.Marshal(result); err == nil {
		h.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return result, nil
}

// ListDocuments retrieves a paginated list of documents
func (h *DocumentQueryHandler) ListDocuments(ctx context.Context, query *ListDocumentsQuery) (*ListDocumentsResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.list_documents",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
			attribute.Int("page", query.Page),
			attribute.Int("page_size", query.PageSize),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("document:list:%s:%s:%s:%d:%d:%s",
		query.TenantID, query.Type, query.Status, query.Page, query.PageSize, query.Search)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var result ListDocumentsResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	tenantID, err := uuid.Parse(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	filter := domain.DocumentFilter{
		TenantID: tenantID,
		Page:     query.Page,
		PageSize: query.PageSize,
	}

	if query.Type != "" {
		filter.Type = domain.DocumentType(query.Type)
	}
	if query.Status != "" {
		filter.Status = domain.ProcessingStatus(query.Status)
	}
	if len(query.Tags) > 0 {
		filter.Tags = query.Tags
	}
	if query.DateFrom != nil {
		filter.DateFrom = query.DateFrom
	}
	if query.DateTo != nil {
		filter.DateTo = query.DateTo
	}
	if query.PageSize <= 0 {
		filter.PageSize = 20
	}
	if query.Page <= 0 {
		filter.Page = 1
	}

	docs, total, err := h.docRepo.List(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	summaries := make([]DocumentSummary, 0, len(docs))
	for _, doc := range docs {
		summaries = append(summaries, *mapDocumentToSummary(&doc))
	}

	pageSize := filter.PageSize
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	result := &ListDocumentsResult{
		Documents:  summaries,
		Total:      total,
		Page:       filter.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	if data, err := json.Marshal(result); err == nil {
		h.cache.Set(ctx, cacheKey, data, 2*time.Minute)
	}

	return result, nil
}

// SearchDocuments performs full-text search on documents
func (h *DocumentQueryHandler) SearchDocuments(ctx context.Context, query *SearchDocumentsQuery) (*SearchDocumentsResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.search_documents",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
			attribute.String("query", query.Query),
		),
	)
	defer span.End()

	tenantID, err := uuid.Parse(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	filters := make(map[string]interface{})
	if query.Type != "" {
		filters["type"] = query.Type
	}
	if len(query.Tags) > 0 {
		filters["tags"] = query.Tags
	}
	if query.DateFrom != nil {
		filters["dateFrom"] = query.DateFrom
	}
	if query.DateTo != nil {
		filters["dateTo"] = query.DateTo
	}

	searchResults, err := h.searchService.Search(ctx, tenantID, query.Query, filters)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("search failed: %w", err)
	}

	docs := make([]DocumentSearchResult, 0, len(searchResults))
	for _, sr := range searchResults {
		docs = append(docs, DocumentSearchResult{
			ID:         sr.ID.String(),
			Score:      sr.Score,
			Highlights: sr.Highlights,
			Metadata:   sr.Metadata,
		})
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}

	return &SearchDocumentsResult{
		Documents: docs,
		Total:     int64(len(docs)),
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// GetDocumentDownloadURL generates a presigned download URL
func (h *DocumentQueryHandler) GetDocumentDownloadURL(ctx context.Context, query *GetDocumentDownloadURLQuery) (*DocumentDownloadURLResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_document_download_url",
		trace.WithAttributes(
			attribute.String("document_id", query.DocumentID),
		),
	)
	defer span.End()

	// This would typically call the storage service to generate a presigned URL
	// For now, return a placeholder
	return &DocumentDownloadURLResult{
		URL:       "",
		ExpiresIn: time.Duration(query.ExpiryMinutes) * time.Minute,
	}, nil
}

// GetDocumentsByType retrieves documents filtered by type
func (h *DocumentQueryHandler) GetDocumentsByType(ctx context.Context, query *GetDocumentsByTypeQuery) (*ListDocumentsResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_documents_by_type",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
			attribute.String("type", query.Type),
		),
	)
	defer span.End()

	return h.ListDocuments(ctx, &ListDocumentsQuery{
		TenantID: query.TenantID,
		Type:     query.Type,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
}

// GetProcessingStatus retrieves processing status for multiple documents
func (h *DocumentQueryHandler) GetProcessingStatus(ctx context.Context, tenantID string, documentIDs []string) (map[string]string, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_processing_status",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID),
		),
	)
	defer span.End()

	status := make(map[string]string)
	for _, docID := range documentIDs {
		doc, err := h.GetDocumentByID(ctx, &GetDocumentByIDQuery{
			DocumentID: docID,
			TenantID:   tenantID,
		})
		if err != nil {
			status[docID] = "error"
			continue
		}
		status[docID] = doc.ProcessingStatus
	}

	return status, nil
}

// Helper functions

func mapDocumentToSummary(doc *domain.Document) *DocumentSummary {
	return &DocumentSummary{
		ID:               doc.ID.String(),
		TenantID:         doc.TenantID.String(),
		Type:             string(doc.Type),
		FileName:         doc.FileName,
		MimeType:         doc.MimeType,
		Size:             doc.Size,
		Checksum:         doc.Checksum,
		ProcessingStatus: string(doc.ProcessingStatus),
		PageCount:        doc.PageCount,
		Tags:             doc.Tags,
		UploadedBy:       doc.UploadedBy.String(),
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        doc.UpdatedAt,
	}
}

func mapDocumentToDetail(doc *domain.Document) *DocumentDetail {
	return &DocumentDetail{
		ID:               doc.ID.String(),
		TenantID:         doc.TenantID.String(),
		Type:             string(doc.Type),
		FileName:         doc.FileName,
		MimeType:         doc.MimeType,
		Size:             doc.Size,
		Checksum:         doc.Checksum,
		Bucket:           doc.Bucket,
		ObjectKey:        doc.ObjectKey,
		VersionID:        doc.VersionID,
		ProcessingStatus: string(doc.ProcessingStatus),
		ExtractedText:    doc.ExtractedText,
		ThumbnailKey:     doc.ThumbnailKey,
		PageCount:        doc.PageCount,
		Metadata:         doc.ExtractedMetadata,
		Tags:             doc.Tags,
		UploadedBy:       doc.UploadedBy.String(),
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        doc.UpdatedAt,
	}
}
