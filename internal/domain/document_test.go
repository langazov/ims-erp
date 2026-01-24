package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDocumentType_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		docType DocumentType
		want    bool
	}{
		{"valid invoice", DocTypeInvoice, true},
		{"valid purchase order", DocTypePurchaseOrder, true},
		{"valid receipt", DocTypeReceipt, true},
		{"valid contract", DocTypeContract, true},
		{"valid scanned", DocTypeScanned, true},
		{"valid other", DocTypeOther, true},
		{"invalid type", DocumentType("invalid"), false},
		{"empty type", DocumentType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.docType.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProcessingStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status ProcessingStatus
		want   bool
	}{
		{"valid pending", ProcessingStatusPending, true},
		{"valid processing", ProcessingStatusProcessing, true},
		{"valid completed", ProcessingStatusCompleted, true},
		{"valid failed", ProcessingStatusFailed, true},
		{"invalid status", ProcessingStatus("invalid"), false},
		{"empty status", ProcessingStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDocument_IsValid(t *testing.T) {
	tenantID := uuid.New()
	docID := uuid.New()

	tests := []struct {
		name string
		doc  *Document
		want bool
	}{
		{
			name: "valid document",
			doc: &Document{
				ID:        docID,
				TenantID:  tenantID,
				FileName:  "test.pdf",
				MimeType:  "application/pdf",
				Bucket:    "documents",
				ObjectKey: "tenant/doc.pdf",
			},
			want: true,
		},
		{
			name: "nil tenant ID",
			doc: &Document{
				ID:        docID,
				TenantID:  uuid.Nil,
				FileName:  "test.pdf",
				MimeType:  "application/pdf",
				Bucket:    "documents",
				ObjectKey: "tenant/doc.pdf",
			},
			want: false,
		},
		{
			name: "empty file name",
			doc: &Document{
				ID:        docID,
				TenantID:  tenantID,
				FileName:  "",
				MimeType:  "application/pdf",
				Bucket:    "documents",
				ObjectKey: "tenant/doc.pdf",
			},
			want: false,
		},
		{
			name: "empty mime type",
			doc: &Document{
				ID:        docID,
				TenantID:  tenantID,
				FileName:  "test.pdf",
				MimeType:  "",
				Bucket:    "documents",
				ObjectKey: "tenant/doc.pdf",
			},
			want: false,
		},
		{
			name: "empty bucket",
			doc: &Document{
				ID:        docID,
				TenantID:  tenantID,
				FileName:  "test.pdf",
				MimeType:  "application/pdf",
				Bucket:    "",
				ObjectKey: "tenant/doc.pdf",
			},
			want: false,
		},
		{
			name: "empty object key",
			doc: &Document{
				ID:        docID,
				TenantID:  tenantID,
				FileName:  "test.pdf",
				MimeType:  "application/pdf",
				Bucket:    "documents",
				ObjectKey: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.doc.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDocumentError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		err := &DocumentError{
			Code:    "TEST_ERROR",
			Message: "Test error message",
		}
		assert.Equal(t, "Test error message", err.Error())
	})

	t.Run("error with wrapped error", func(t *testing.T) {
		wrappedErr := assert.AnError
		err := &DocumentError{
			Code:    "WRAPPED_ERROR",
			Message: "Wrapped error",
			Err:     wrappedErr,
		}
		assert.Contains(t, err.Error(), "Wrapped error")
		assert.Contains(t, err.Error(), wrappedErr.Error())
		assert.Equal(t, wrappedErr, err.Unwrap())
	})

	t.Run("Is method with errors.Is", func(t *testing.T) {
		err1 := &DocumentError{Code: "ERR1", Message: "Error 1"}
		err2 := &DocumentError{Code: "ERR2", Message: "Error 2"}
		var target *DocumentError

		assert.True(t, errors.Is(err1, target))
		assert.True(t, errors.Is(err1, err2))
		assert.True(t, errors.Is(err2, err1))
	})
}

func TestDocumentMetadata(t *testing.T) {
	t.Run("invoice metadata", func(t *testing.T) {
		now := time.Now()
		metadata := DocumentMetadata{
			InvoiceNumber: "INV-001",
			InvoiceDate:   now,
			TotalAmount:   decimal.NewFromFloat(1500.50),
			VendorName:    "Acme Corp",
			Dates:         []string{"2024-01-15", "2024-02-01"},
			Amounts:       []string{"$1,500.50", "$500.00"},
			Emails:        []string{"vendor@example.com"},
		}
		assert.Equal(t, "INV-001", metadata.InvoiceNumber)
		assert.Equal(t, now, metadata.InvoiceDate)
		assert.True(t, metadata.TotalAmount.Equal(decimal.NewFromFloat(1500.50)))
		assert.Equal(t, "Acme Corp", metadata.VendorName)
		assert.Len(t, metadata.Dates, 2)
		assert.Len(t, metadata.Amounts, 2)
		assert.Len(t, metadata.Emails, 1)
	})

	t.Run("empty metadata", func(t *testing.T) {
		metadata := DocumentMetadata{}
		assert.Empty(t, metadata.InvoiceNumber)
		assert.True(t, metadata.TotalAmount.IsZero())
	})
}

func TestSearchResult(t *testing.T) {
	id := uuid.New()
	result := SearchResult{
		ID:    id,
		Score: 0.95,
		Highlights: map[string][]string{
			"content": {"highlighted text"},
		},
		Metadata: map[string]interface{}{
			"type": "invoice",
		},
	}
	assert.Equal(t, id, result.ID)
	assert.Equal(t, 0.95, result.Score)
	assert.Contains(t, result.Highlights, "content")
	assert.Contains(t, result.Metadata, "type")
}

func TestDocumentFilter(t *testing.T) {
	tenantID := uuid.New()
	dateFrom := time.Now().AddDate(0, -1, 0)
	dateTo := time.Now()

	filter := DocumentFilter{
		TenantID:   tenantID,
		Type:       DocTypeInvoice,
		Status:     ProcessingStatusCompleted,
		Tags:       []string{"urgent", "finance"},
		UploadedBy: uuid.New(),
		DateFrom:   &dateFrom,
		DateTo:     &dateTo,
		FileName:   "invoice",
		Page:       1,
		PageSize:   20,
	}

	assert.Equal(t, tenantID, filter.TenantID)
	assert.Equal(t, DocTypeInvoice, filter.Type)
	assert.Equal(t, ProcessingStatusCompleted, filter.Status)
	assert.Equal(t, 2, len(filter.Tags))
	assert.Equal(t, 1, filter.Page)
	assert.Equal(t, 20, filter.PageSize)
}

func TestDocumentLifecycle(t *testing.T) {
	tenantID := uuid.New()
	docID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	doc := &Document{
		ID:               docID,
		TenantID:         tenantID,
		Type:             DocTypeInvoice,
		FileName:         "invoice-001.pdf",
		MimeType:         "application/pdf",
		Size:             1024,
		Checksum:         "abc123",
		Bucket:           "documents",
		ObjectKey:        "invoices/2024/01/invoice-001.pdf",
		ProcessingStatus: ProcessingStatusPending,
		UploadedBy:       userID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	assert.Equal(t, ProcessingStatusPending, doc.ProcessingStatus)
	assert.Empty(t, doc.ExtractedText)
	assert.Empty(t, doc.Tags)
	assert.Zero(t, doc.PageCount)

	doc.ProcessingStatus = ProcessingStatusProcessing
	assert.Equal(t, ProcessingStatusProcessing, doc.ProcessingStatus)

	doc.ExtractedText = "Extracted invoice text..."
	doc.ExtractedMetadata = DocumentMetadata{
		InvoiceNumber: "INV-001",
		TotalAmount:   decimal.NewFromFloat(500.00),
	}
	doc.PageCount = 3
	doc.Tags = []string{"urgent", "Q1"}

	assert.Equal(t, ProcessingStatusProcessing, doc.ProcessingStatus)
	assert.NotEmpty(t, doc.ExtractedText)
	assert.Equal(t, "INV-001", doc.ExtractedMetadata.InvoiceNumber)
	assert.Equal(t, 3, doc.PageCount)
	assert.Len(t, doc.Tags, 2)

	doc.ProcessingStatus = ProcessingStatusCompleted
	assert.Equal(t, ProcessingStatusCompleted, doc.ProcessingStatus)
}

func TestDocumentErrorTypes(t *testing.T) {
	assert.Equal(t, "DOCUMENT_NOT_FOUND", ErrDocumentNotFound.Code)
	assert.Equal(t, "document not found", ErrDocumentNotFound.Message)

	assert.Equal(t, "INVALID_DOCUMENT", ErrInvalidDocument.Code)
	assert.Equal(t, "invalid document data", ErrInvalidDocument.Message)

	assert.Equal(t, "DOCUMENT_TOO_LARGE", ErrDocumentTooLarge.Code)
	assert.Equal(t, "document exceeds maximum size", ErrDocumentTooLarge.Message)

	assert.Equal(t, "UNSUPPORTED_FILE_TYPE", ErrUnsupportedFileType.Code)
	assert.Equal(t, "file type not supported", ErrUnsupportedFileType.Message)

	assert.Equal(t, "UPLOAD_FAILED", ErrUploadFailed.Code)
	assert.Equal(t, "document upload failed", ErrUploadFailed.Message)

	assert.Equal(t, "DOWNLOAD_FAILED", ErrDownloadFailed.Code)
	assert.Equal(t, "document download failed", ErrDownloadFailed.Message)

	assert.Equal(t, "PROCESSING_FAILED", ErrProcessingFailed.Code)
	assert.Equal(t, "document processing failed", ErrProcessingFailed.Message)

	assert.Equal(t, "CHECKSUM_MISMATCH", ErrChecksumMismatch.Code)
	assert.Equal(t, "document checksum mismatch", ErrChecksumMismatch.Message)

	assert.Equal(t, "BUCKET_NOT_FOUND", ErrBucketNotFound.Code)
	assert.Equal(t, "bucket not found", ErrBucketNotFound.Message)

	assert.Equal(t, "PRESIGNED_URL_EXPIRED", ErrPresignedURLExpired.Code)
	assert.Equal(t, "presigned URL has expired", ErrPresignedURLExpired.Message)
}
