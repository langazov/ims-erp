package commands

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/events"
)

// Command types
type UploadDocument struct {
	Type     string
	FileName string
	MimeType string
	Size     int64
	Data     []byte
	Tags     []string
	Metadata map[string]interface{}
}

type CreateDocument struct {
	Type      string
	FileName  string
	MimeType  string
	Size      int64
	Checksum  string
	Bucket    string
	ObjectKey string
	Tags      []string
}

type DeleteDocument struct {
	DocumentID uuid.UUID
	Force      bool
}

type UpdateDocumentMetadata struct {
	DocumentID uuid.UUID
	Tags       []string
	Metadata   map[string]interface{}
}

type ReprocessDocument struct {
	DocumentID uuid.UUID
}

type GeneratePresignedUploadURL struct {
	Type          string
	FileName      string
	MimeType      string
	Size          int64
	ExpiryMinutes int
}

type CompleteUpload struct {
	UploadID  string
	Bucket    string
	ObjectKey string
	Checksum  string
	Size      int64
}

// DocumentCommandHandler handles document-related commands
type DocumentCommandHandler struct {
	docRepo           domain.DocumentRepository
	storageService    domain.StorageService
	processingService domain.ProcessingService
	searchService     domain.SearchService
	publisher         events.Publisher
}

func NewDocumentCommandHandler(
	docRepo domain.DocumentRepository,
	storageService domain.StorageService,
	processingService domain.ProcessingService,
	searchService domain.SearchService,
	publisher events.Publisher,
) *DocumentCommandHandler {
	return &DocumentCommandHandler{
		docRepo:           docRepo,
		storageService:    storageService,
		processingService: processingService,
		searchService:     searchService,
		publisher:         publisher,
	}
}

// HandleUploadDocument handles direct document upload with data
func (h *DocumentCommandHandler) HandleUploadDocument(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input UploadDocument
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	docType := domain.DocumentType(input.Type)
	if !docType.IsValid() {
		return nil, domain.ErrInvalidDocument
	}

	// Validate file size (max 100MB)
	if input.Size > 100*1024*1024 {
		return nil, domain.ErrDocumentTooLarge
	}

	// Calculate checksum
	hash := sha256.Sum256(input.Data)
	checksum := hex.EncodeToString(hash[:])

	// Check for duplicate
	existing, err := h.docRepo.GetByChecksum(ctx, tenantID, checksum)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("document already exists: %s", existing.ID)
	}

	// Generate bucket and object key
	bucket := fmt.Sprintf("%s-documents", tenantID.String())
	objectKey := generateObjectKey(docType, input.FileName)

	// Ensure bucket exists
	exists, err := h.storageService.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := h.storageService.CreateBucket(ctx, bucket); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// Upload to storage
	if err := h.storageService.Upload(ctx, bucket, objectKey, input.Data, input.MimeType); err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	// Create document record
	doc := &domain.Document{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Type:             docType,
		FileName:         input.FileName,
		MimeType:         input.MimeType,
		Size:             input.Size,
		Checksum:         checksum,
		Bucket:           bucket,
		ObjectKey:        objectKey,
		ProcessingStatus: domain.ProcessingStatusPending,
		Tags:             input.Tags,
		UploadedBy:       userID,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := h.docRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	// Process document asynchronously
	go func() {
		processCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		doc.ProcessingStatus = domain.ProcessingStatusProcessing
		h.docRepo.Update(processCtx, doc)

		processedDoc, err := h.processingService.ProcessDocument(processCtx, doc, input.Data)
		if err != nil {
			doc.ProcessingStatus = domain.ProcessingStatusFailed
			h.docRepo.Update(processCtx, doc)
			return
		}

		doc.ExtractedText = processedDoc.ExtractedText
		doc.ThumbnailKey = processedDoc.ThumbnailKey
		doc.PageCount = processedDoc.PageCount
		doc.ExtractedMetadata = processedDoc.ExtractedMetadata
		doc.ProcessingStatus = domain.ProcessingStatusCompleted
		doc.UpdatedAt = time.Now().UTC()

		h.docRepo.Update(processCtx, doc)
		h.searchService.IndexDocument(processCtx, doc)

		evt := events.NewDocumentProcessingCompletedEvent(doc, cmd.UserID)
		h.publisher.PublishEvent(processCtx, &evt.EventEnvelope)
	}()

	// Publish event
	evt := events.NewDocumentUploadedEvent(doc, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    doc,
		Events:  []interface{}{evt},
	}, nil
}

// HandleGeneratePresignedUploadURL generates a presigned URL for client-side upload
func (h *DocumentCommandHandler) HandleGeneratePresignedUploadURL(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input GeneratePresignedUploadURL
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	docType := domain.DocumentType(input.Type)
	if !docType.IsValid() {
		return nil, domain.ErrInvalidDocument
	}

	// Validate file size
	if input.Size > 100*1024*1024 {
		return nil, domain.ErrDocumentTooLarge
	}

	// Generate bucket and object key
	bucket := fmt.Sprintf("%s-documents", tenantID.String())
	objectKey := generateObjectKey(docType, input.FileName)

	// Ensure bucket exists
	exists, err := h.storageService.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := h.storageService.CreateBucket(ctx, bucket); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// Generate presigned URL
	expiry := time.Duration(input.ExpiryMinutes) * time.Minute
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}

	uploadURL, err := h.storageService.GetPresignedUploadURL(ctx, bucket, objectKey, input.MimeType, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Create temporary document record
	doc := &domain.Document{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Type:             docType,
		FileName:         input.FileName,
		MimeType:         input.MimeType,
		Size:             input.Size,
		Bucket:           bucket,
		ObjectKey:        objectKey,
		ProcessingStatus: domain.ProcessingStatusPending,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	return &CommandResult{
		Success: true,
		Data: map[string]interface{}{
			"uploadUrl":  uploadURL,
			"documentId": doc.ID.String(),
			"bucket":     bucket,
			"objectKey":  objectKey,
			"expiresIn":  expiry.Seconds(),
		},
	}, nil
}

// HandleCompleteUpload completes the upload process after client-side upload
func (h *DocumentCommandHandler) HandleCompleteUpload(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input CompleteUpload
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	docID, err := uuid.Parse(input.UploadID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}

	// Download to verify
	data, err := h.storageService.Download(ctx, input.Bucket, input.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to verify upload: %w", err)
	}

	// Verify checksum
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])
	if checksum != input.Checksum {
		return nil, domain.ErrChecksumMismatch
	}

	// Create document record
	doc := &domain.Document{
		ID:               docID,
		TenantID:         tenantID,
		Checksum:         checksum,
		Size:             input.Size,
		ProcessingStatus: domain.ProcessingStatusPending,
		UploadedBy:       userID,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := h.docRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	// Process document asynchronously
	go func() {
		processCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		doc.ProcessingStatus = domain.ProcessingStatusProcessing
		h.docRepo.Update(processCtx, doc)

		processedDoc, err := h.processingService.ProcessDocument(processCtx, doc, data)
		if err != nil {
			doc.ProcessingStatus = domain.ProcessingStatusFailed
			h.docRepo.Update(processCtx, doc)
			return
		}

		doc.ExtractedText = processedDoc.ExtractedText
		doc.ThumbnailKey = processedDoc.ThumbnailKey
		doc.PageCount = processedDoc.PageCount
		doc.ExtractedMetadata = processedDoc.ExtractedMetadata
		doc.ProcessingStatus = domain.ProcessingStatusCompleted
		doc.UpdatedAt = time.Now().UTC()

		h.docRepo.Update(processCtx, doc)
		h.searchService.IndexDocument(processCtx, doc)

		evt := events.NewDocumentProcessingCompletedEvent(doc, cmd.UserID)
		h.publisher.PublishEvent(processCtx, &evt.EventEnvelope)
	}()

	// Publish event
	evt := events.NewDocumentUploadedEvent(doc, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    doc,
		Events:  []interface{}{evt},
	}, nil
}

// HandleDeleteDocument handles document deletion
func (h *DocumentCommandHandler) HandleDeleteDocument(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input DeleteDocument
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Get document
	doc, err := h.docRepo.GetByID(ctx, tenantID, input.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Delete from storage if force delete
	if input.Force {
		if err := h.storageService.Delete(ctx, doc.Bucket, doc.ObjectKey); err != nil {
			return nil, fmt.Errorf("failed to delete from storage: %w", err)
		}

		if doc.ThumbnailKey != "" {
			h.storageService.Delete(ctx, doc.Bucket+"-thumbnails", doc.ThumbnailKey)
		}
	}

	// Delete from search index
	if err := h.searchService.DeleteFromIndex(ctx, tenantID, input.DocumentID); err != nil {
		// Log but don't fail
		_ = err
	}

	// Delete document record
	if err := h.docRepo.Delete(ctx, tenantID, input.DocumentID); err != nil {
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}

	// Publish event
	evt := events.NewDocumentDeletedEvent(input.DocumentID.String(), tenantID.String(), cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    map[string]string{"id": input.DocumentID.String()},
		Events:  []interface{}{evt},
	}, nil
}

// HandleUpdateDocumentMetadata updates document metadata
func (h *DocumentCommandHandler) HandleUpdateDocumentMetadata(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input UpdateDocumentMetadata
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Get document
	doc, err := h.docRepo.GetByID(ctx, tenantID, input.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Update tags
	if input.Tags != nil {
		doc.Tags = input.Tags
	}

	doc.UpdatedAt = time.Now().UTC()

	if err := h.docRepo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	// Re-index in search
	h.searchService.IndexDocument(ctx, doc)

	// Publish event
	evt := events.NewDocumentUpdatedEvent(doc, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    doc,
		Events:  []interface{}{evt},
	}, nil
}

// HandleReprocessDocument triggers document reprocessing
func (h *DocumentCommandHandler) HandleReprocessDocument(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input ReprocessDocument
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Get document
	doc, err := h.docRepo.GetByID(ctx, tenantID, input.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Download document data
	data, err := h.storageService.Download(ctx, doc.Bucket, doc.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download document: %w", err)
	}

	// Reset processing status
	doc.ProcessingStatus = domain.ProcessingStatusProcessing
	doc.UpdatedAt = time.Now().UTC()
	h.docRepo.Update(ctx, doc)

	// Process asynchronously
	go func() {
		processCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		processedDoc, err := h.processingService.ProcessDocument(processCtx, doc, data)
		if err != nil {
			doc.ProcessingStatus = domain.ProcessingStatusFailed
			h.docRepo.Update(processCtx, doc)
			return
		}

		doc.ExtractedText = processedDoc.ExtractedText
		doc.ThumbnailKey = processedDoc.ThumbnailKey
		doc.PageCount = processedDoc.PageCount
		doc.ExtractedMetadata = processedDoc.ExtractedMetadata
		doc.ProcessingStatus = domain.ProcessingStatusCompleted
		doc.UpdatedAt = time.Now().UTC()

		h.docRepo.Update(processCtx, doc)
		h.searchService.IndexDocument(processCtx, doc)

		evt := events.NewDocumentProcessingCompletedEvent(doc, cmd.UserID)
		h.publisher.PublishEvent(processCtx, &evt.EventEnvelope)
	}()

	// Publish event
	evt := events.NewDocumentProcessingStartedEvent(doc, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    doc,
		Events:  []interface{}{evt},
	}, nil
}

// Helper function to generate object key
func generateObjectKey(docType domain.DocumentType, fileName string) string {
	now := time.Now().UTC()
	ext := getFileExtension(fileName)
	return fmt.Sprintf("%s/%d/%02d/%s.%s",
		docType,
		now.Year(),
		now.Month(),
		uuid.New().String(),
		ext,
	)
}

func getFileExtension(fileName string) string {
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			return fileName[i+1:]
		}
	}
	return "bin"
}
