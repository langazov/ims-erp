package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DocumentType string

const (
	DocTypeInvoice       DocumentType = "invoice"
	DocTypePurchaseOrder DocumentType = "purchase_order"
	DocTypeReceipt       DocumentType = "receipt"
	DocTypeContract      DocumentType = "contract"
	DocTypeScanned       DocumentType = "scanned"
	DocTypeOther         DocumentType = "other"
)

type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

type Document struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	Type              DocumentType
	FileName          string
	MimeType          string
	Size              int64
	Checksum          string
	Bucket            string
	ObjectKey         string
	VersionID         string
	ProcessingStatus  ProcessingStatus
	ExtractedText     string
	ThumbnailKey      string
	PageCount         int
	ExtractedMetadata DocumentMetadata
	Tags              []string
	UploadedBy        uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type DocumentMetadata struct {
	InvoiceNumber string          `json:"invoiceNumber,omitempty"`
	InvoiceDate   time.Time       `json:"invoiceDate,omitempty"`
	TotalAmount   decimal.Decimal `json:"totalAmount,omitempty"`
	VendorName    string          `json:"vendorName,omitempty"`
	Dates         []string        `json:"dates,omitempty"`
	Amounts       []string        `json:"amounts,omitempty"`
	Emails        []string        `json:"emails,omitempty"`
}

type DocumentFilter struct {
	TenantID   uuid.UUID
	Type       DocumentType
	Status     ProcessingStatus
	Tags       []string
	UploadedBy uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
	FileName   string
	Page       int
	PageSize   int
}

func (d *Document) IsValid() bool {
	return d.TenantID != uuid.Nil &&
		d.FileName != "" &&
		d.MimeType != "" &&
		d.Bucket != "" &&
		d.ObjectKey != ""
}

func (t DocumentType) IsValid() bool {
	switch t {
	case DocTypeInvoice, DocTypePurchaseOrder, DocTypeReceipt,
		DocTypeContract, DocTypeScanned, DocTypeOther:
		return true
	}
	return false
}

func (s ProcessingStatus) IsValid() bool {
	switch s {
	case ProcessingStatusPending, ProcessingStatusProcessing,
		ProcessingStatusCompleted, ProcessingStatusFailed:
		return true
	}
	return false
}

type DocumentError struct {
	Code    string
	Message string
	Err     error
}

func (e *DocumentError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *DocumentError) Unwrap() error {
	return e.Err
}

func (e *DocumentError) Is(target error) bool {
	_, ok := target.(*DocumentError)
	return ok
}

var (
	ErrDocumentNotFound    = &DocumentError{Code: "DOCUMENT_NOT_FOUND", Message: "document not found"}
	ErrInvalidDocument     = &DocumentError{Code: "INVALID_DOCUMENT", Message: "invalid document data"}
	ErrDocumentTooLarge    = &DocumentError{Code: "DOCUMENT_TOO_LARGE", Message: "document exceeds maximum size"}
	ErrUnsupportedFileType = &DocumentError{Code: "UNSUPPORTED_FILE_TYPE", Message: "file type not supported"}
	ErrUploadFailed        = &DocumentError{Code: "UPLOAD_FAILED", Message: "document upload failed"}
	ErrDownloadFailed      = &DocumentError{Code: "DOWNLOAD_FAILED", Message: "document download failed"}
	ErrProcessingFailed    = &DocumentError{Code: "PROCESSING_FAILED", Message: "document processing failed"}
	ErrChecksumMismatch    = &DocumentError{Code: "CHECKSUM_MISMATCH", Message: "document checksum mismatch"}
	ErrBucketNotFound      = &DocumentError{Code: "BUCKET_NOT_FOUND", Message: "bucket not found"}
	ErrPresignedURLExpired = &DocumentError{Code: "PRESIGNED_URL_EXPIRED", Message: "presigned URL has expired"}
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Document, error)
	Update(ctx context.Context, doc *Document) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, filter DocumentFilter) ([]Document, int64, error)
	GetByChecksum(ctx context.Context, tenantID uuid.UUID, checksum string) (*Document, error)
}

type StorageService interface {
	Upload(ctx context.Context, bucket, objectKey string, data []byte, contentType string) error
	Download(ctx context.Context, bucket, objectKey string) ([]byte, error)
	Delete(ctx context.Context, bucket, objectKey string) error
	GetPresignedUploadURL(ctx context.Context, bucket, objectKey string, contentType string, expiry time.Duration) (string, error)
	GetPresignedDownloadURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)
	BucketExists(ctx context.Context, bucket string) (bool, error)
	CreateBucket(ctx context.Context, bucket string) error
}

type ProcessingService interface {
	ProcessDocument(ctx context.Context, doc *Document, data []byte) (*Document, error)
	ExtractText(ctx context.Context, data []byte, mimeType string) (string, error)
	ExtractMetadata(ctx context.Context, docType DocumentType, text string) DocumentMetadata
	GenerateThumbnail(ctx context.Context, data []byte, mimeType string) ([]byte, error)
}

type SearchService interface {
	IndexDocument(ctx context.Context, doc *Document) error
	DeleteFromIndex(ctx context.Context, tenantID, id uuid.UUID) error
	Search(ctx context.Context, tenantID uuid.UUID, query string, filters map[string]interface{}) ([]SearchResult, error)
	Suggest(ctx context.Context, tenantID uuid.UUID, prefix string) ([]string, error)
}

type SearchResult struct {
	ID         uuid.UUID
	Score      float64
	Highlights map[string][]string
	Metadata   map[string]interface{}
}
