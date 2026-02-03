package events

import (
	"time"

	"github.com/ims-erp/system/internal/domain"
)

type DocumentUploadedEvent struct {
	EventEnvelope
}

func NewDocumentUploadedEvent(doc *domain.Document, userID string) *DocumentUploadedEvent {
	event := NewEvent(
		doc.ID.String(),
		"Document",
		"document.uploaded",
		doc.TenantID.String(),
		userID,
		map[string]interface{}{
			"fileName":  doc.FileName,
			"type":      string(doc.Type),
			"mimeType":  doc.MimeType,
			"size":      doc.Size,
			"bucket":    doc.Bucket,
			"objectKey": doc.ObjectKey,
			"tags":      doc.Tags,
		},
	)
	return &DocumentUploadedEvent{*event}
}

type DocumentProcessingStartedEvent struct {
	EventEnvelope
}

func NewDocumentProcessingStartedEvent(doc *domain.Document, userID string) *DocumentProcessingStartedEvent {
	event := NewEvent(
		doc.ID.String(),
		"Document",
		"document.processing.started",
		doc.TenantID.String(),
		userID,
		map[string]interface{}{
			"fileName": doc.FileName,
			"type":     string(doc.Type),
		},
	)
	return &DocumentProcessingStartedEvent{*event}
}

type DocumentProcessingCompletedEvent struct {
	EventEnvelope
}

func NewDocumentProcessingCompletedEvent(doc *domain.Document, userID string) *DocumentProcessingCompletedEvent {
	event := NewEvent(
		doc.ID.String(),
		"Document",
		"document.processing.completed",
		doc.TenantID.String(),
		userID,
		map[string]interface{}{
			"fileName":      doc.FileName,
			"type":          string(doc.Type),
			"pageCount":     doc.PageCount,
			"extractedText": doc.ExtractedText != "",
			"thumbnailKey":  doc.ThumbnailKey,
			"metadata": map[string]interface{}{
				"invoiceNumber": doc.ExtractedMetadata.InvoiceNumber,
				"vendorName":    doc.ExtractedMetadata.VendorName,
				"totalAmount":   doc.ExtractedMetadata.TotalAmount.String(),
			},
		},
	)
	return &DocumentProcessingCompletedEvent{*event}
}

type DocumentProcessingFailedEvent struct {
	EventEnvelope
}

func NewDocumentProcessingFailedEvent(docID, tenantID, userID string, reason string) *DocumentProcessingFailedEvent {
	event := NewEvent(
		docID,
		"Document",
		"document.processing.failed",
		tenantID,
		userID,
		map[string]interface{}{
			"reason":   reason,
			"failedAt": time.Now().UTC(),
		},
	)
	return &DocumentProcessingFailedEvent{*event}
}

type DocumentDeletedEvent struct {
	EventEnvelope
}

func NewDocumentDeletedEvent(docID, tenantID, userID string) *DocumentDeletedEvent {
	event := NewEvent(
		docID,
		"Document",
		"document.deleted",
		tenantID,
		userID,
		map[string]interface{}{
			"deletedAt": time.Now().UTC(),
		},
	)
	return &DocumentDeletedEvent{*event}
}

type DocumentUpdatedEvent struct {
	EventEnvelope
}

func NewDocumentUpdatedEvent(doc *domain.Document, userID string) *DocumentUpdatedEvent {
	event := NewEvent(
		doc.ID.String(),
		"Document",
		"document.updated",
		doc.TenantID.String(),
		userID,
		map[string]interface{}{
			"fileName":      doc.FileName,
			"type":          string(doc.Type),
			"tags":          doc.Tags,
			"updatedFields": []string{"tags"},
		},
	)
	return &DocumentUpdatedEvent{*event}
}

type DocumentIndexedEvent struct {
	EventEnvelope
}

func NewDocumentIndexedEvent(docID, tenantID, userID string) *DocumentIndexedEvent {
	event := NewEvent(
		docID,
		"Document",
		"document.indexed",
		tenantID,
		userID,
		map[string]interface{}{
			"indexedAt": time.Now().UTC(),
		},
	)
	return &DocumentIndexedEvent{*event}
}

type DocumentSearchIndexDeletedEvent struct {
	EventEnvelope
}

func NewDocumentSearchIndexDeletedEvent(docID, tenantID, userID string) *DocumentSearchIndexDeletedEvent {
	event := NewEvent(
		docID,
		"Document",
		"document.index.deleted",
		tenantID,
		userID,
		map[string]interface{}{
			"deletedAt": time.Now().UTC(),
		},
	)
	return &DocumentSearchIndexDeletedEvent{*event}
}
