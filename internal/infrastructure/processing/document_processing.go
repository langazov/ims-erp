package processing

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"regexp"
	"strings"
	"time"

	"github.com/ims-erp/system/internal/domain"
	"github.com/shopspring/decimal"
)

// DocumentProcessingService implements domain.ProcessingService
type DocumentProcessingService struct {
	storageService domain.StorageService
	extractors     map[domain.DocumentType]MetadataExtractor
}

// MetadataExtractor extracts metadata from document text
type MetadataExtractor func(text string) domain.DocumentMetadata

// NewDocumentProcessingService creates a new document processing service
func NewDocumentProcessingService(storageService domain.StorageService) *DocumentProcessingService {
	service := &DocumentProcessingService{
		storageService: storageService,
		extractors:     make(map[domain.DocumentType]MetadataExtractor),
	}

	// Register extractors
	service.extractors[domain.DocTypeInvoice] = extractInvoiceMetadata
	service.extractors[domain.DocTypeReceipt] = extractReceiptMetadata
	service.extractors[domain.DocTypePurchaseOrder] = extractPurchaseOrderMetadata
	service.extractors[domain.DocTypeContract] = extractContractMetadata

	return service
}

// ProcessDocument processes a document: extracts text, metadata, and generates thumbnail
func (s *DocumentProcessingService) ProcessDocument(ctx context.Context, doc *domain.Document, data []byte) (*domain.Document, error) {
	processedDoc := &domain.Document{
		ID:               doc.ID,
		TenantID:         doc.TenantID,
		Type:             doc.Type,
		FileName:         doc.FileName,
		MimeType:         doc.MimeType,
		Size:             doc.Size,
		Checksum:         doc.Checksum,
		Bucket:           doc.Bucket,
		ObjectKey:        doc.ObjectKey,
		VersionID:        doc.VersionID,
		ProcessingStatus: doc.ProcessingStatus,
		Tags:             doc.Tags,
		UploadedBy:       doc.UploadedBy,
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        time.Now().UTC(),
	}

	// Extract text based on document type
	text, err := s.ExtractText(ctx, data, doc.MimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}
	processedDoc.ExtractedText = text

	// Extract metadata
	processedDoc.ExtractedMetadata = s.ExtractMetadata(ctx, doc.Type, text)

	// Generate thumbnail for images and PDFs
	if isImage(doc.MimeType) || doc.MimeType == "application/pdf" {
		thumbnail, err := s.GenerateThumbnail(ctx, data, doc.MimeType)
		if err == nil && thumbnail != nil {
			thumbnailKey := fmt.Sprintf("thumbnails/%s.jpg", doc.ID.String())
			err = s.storageService.Upload(ctx, doc.Bucket+"-processed", thumbnailKey, thumbnail, "image/jpeg")
			if err == nil {
				processedDoc.ThumbnailKey = thumbnailKey
			}
		}
	}

	// Estimate page count
	processedDoc.PageCount = estimatePageCount(data, doc.MimeType, text)

	return processedDoc, nil
}

// ExtractText extracts text from document data
func (s *DocumentProcessingService) ExtractText(ctx context.Context, data []byte, mimeType string) (string, error) {
	switch {
	case mimeType == "text/plain":
		return string(data), nil
	case mimeType == "application/pdf":
		// In a real implementation, use a PDF library to extract text
		// For now, return a placeholder
		return extractTextFromPDF(data)
	case isImage(mimeType):
		// In a real implementation, use OCR (Tesseract, etc.)
		// For now, return a placeholder
		return performOCR(data, mimeType)
	case mimeType == "application/msword" || mimeType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		// In a real implementation, use a DOCX library
		return extractTextFromWord(data)
	default:
		return "", fmt.Errorf("unsupported mime type for text extraction: %s", mimeType)
	}
}

// ExtractMetadata extracts structured metadata from document text
func (s *DocumentProcessingService) ExtractMetadata(ctx context.Context, docType domain.DocumentType, text string) domain.DocumentMetadata {
	// Use the registered extractor if available
	if extractor, ok := s.extractors[docType]; ok {
		return extractor(text)
	}

	// Default extraction
	return extractGenericMetadata(text)
}

// GenerateThumbnail creates a thumbnail image from document data
func (s *DocumentProcessingService) GenerateThumbnail(ctx context.Context, data []byte, mimeType string) ([]byte, error) {
	if !isImage(mimeType) {
		return nil, fmt.Errorf("cannot generate thumbnail for non-image type: %s", mimeType)
	}

	// Decode image
	var img image.Image
	var err error

	switch {
	case mimeType == "image/jpeg":
		img, err = jpeg.Decode(bytes.NewReader(data))
	case mimeType == "image/png":
		img, err = png.Decode(bytes.NewReader(data))
	default:
		return nil, fmt.Errorf("unsupported image format: %s", mimeType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize to thumbnail (max 200x200)
	thumb := resizeImage(img, 200, 200)

	// Encode to JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// Helper functions

func isImage(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

func extractTextFromPDF(data []byte) (string, error) {
	// Placeholder for PDF text extraction
	// In production, use a library like pdfcpu or unidoc
	return "PDF content placeholder", nil
}

func performOCR(data []byte, mimeType string) (string, error) {
	// Placeholder for OCR
	// In production, use Tesseract via a Go wrapper or call external service
	return "OCR text placeholder", nil
}

func extractTextFromWord(data []byte) (string, error) {
	// Placeholder for Word document text extraction
	// In production, use a library like unioffice
	return "Word document content placeholder", nil
}

func estimatePageCount(data []byte, mimeType string, text string) int {
	switch {
	case mimeType == "application/pdf":
		// Rough estimate: average PDF page is ~50KB
		return len(data)/(50*1024) + 1
	case isImage(mimeType):
		return 1
	default:
		// Text-based: estimate ~3000 chars per page
		return len(text)/3000 + 1
	}
}

func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate new dimensions while maintaining aspect ratio
	if width > maxWidth || height > maxHeight {
		scaleX := float64(maxWidth) / float64(width)
		scaleY := float64(maxHeight) / float64(height)
		scale := scaleX
		if scaleY < scale {
			scale = scaleY
		}

		newWidth := int(float64(width) * scale)
		newHeight := int(float64(height) * scale)

		// Create new image
		newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

		// Simple nearest-neighbor resize
		for y := 0; y < newHeight; y++ {
			for x := 0; x < newWidth; x++ {
				srcX := int(float64(x) / scale)
				srcY := int(float64(y) / scale)
				newImg.Set(x, y, img.At(srcX+bounds.Min.X, srcY+bounds.Min.Y))
			}
		}

		return newImg
	}

	return img
}

// Metadata extractors

func extractInvoiceMetadata(text string) domain.DocumentMetadata {
	metadata := domain.DocumentMetadata{
		Dates:   []string{},
		Amounts: []string{},
		Emails:  []string{},
	}

	// Extract invoice number
	invoicePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)invoice\s*(?:#|number|num|no)?[:\s]*(\w+[-]?\d+)`),
		regexp.MustCompile(`(?i)inv[.#\s]*(\w+[-]?\d+)`),
		regexp.MustCompile(`(?i)bill\s*(?:#|number)?[:\s]*(\w+[-]?\d+)`),
	}

	for _, pattern := range invoicePatterns {
		if matches := pattern.FindStringSubmatch(text); len(matches) > 1 {
			metadata.InvoiceNumber = matches[1]
			break
		}
	}

	// Extract dates
	datePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:invoice\s*)?date[:\s]*(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})`),
		regexp.MustCompile(`(?i)(?:invoice\s*)?date[:\s]*(\d{4}[/-]\d{1,2}[/-]\d{1,2})`),
		regexp.MustCompile(`(\d{1,2}\s+(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\.?\s+\d{4})`),
	}

	for _, pattern := range datePatterns {
		if matches := pattern.FindStringSubmatch(text); len(matches) > 1 {
			metadata.Dates = append(metadata.Dates, matches[1])
			// Try to parse the first date as invoice date
			if t, err := parseDate(matches[1]); err == nil {
				metadata.InvoiceDate = t
				break
			}
		}
	}

	// Extract amounts
	amountPattern := regexp.MustCompile(`(?i)(?:total|amount|sum|balance)[\s:]*[$€£]?\s*([\d,]+\.?\d*)`)
	if matches := amountPattern.FindStringSubmatch(text); len(matches) > 1 {
		amount := strings.ReplaceAll(matches[1], ",", "")
		if d, err := decimal.NewFromString(amount); err == nil {
			metadata.TotalAmount = d
		}
		metadata.Amounts = append(metadata.Amounts, matches[1])
	}

	// Extract vendor name (typically at the top of invoice)
	lines := strings.Split(text, "\n")
	if len(lines) > 0 {
		metadata.VendorName = strings.TrimSpace(lines[0])
	}

	// Extract emails
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailPattern.FindAllString(text, -1)
	metadata.Emails = emails

	return metadata
}

func extractReceiptMetadata(text string) domain.DocumentMetadata {
	metadata := domain.DocumentMetadata{
		Dates:   []string{},
		Amounts: []string{},
	}

	// Extract date
	datePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:date|time)[:\s]*(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})`),
		regexp.MustCompile(`(\d{1,2}\s+(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\.?\s+\d{4})`),
	}

	for _, pattern := range datePatterns {
		if matches := pattern.FindStringSubmatch(text); len(matches) > 1 {
			metadata.Dates = append(metadata.Dates, matches[1])
			if t, err := parseDate(matches[1]); err == nil {
				metadata.InvoiceDate = t
			}
			break
		}
	}

	// Extract total
	totalPattern := regexp.MustCompile(`(?i)(?:total|amount)[\s:]*[$€£]?\s*([\d,]+\.?\d*)`)
	if matches := totalPattern.FindStringSubmatch(text); len(matches) > 1 {
		amount := strings.ReplaceAll(matches[1], ",", "")
		if d, err := decimal.NewFromString(amount); err == nil {
			metadata.TotalAmount = d
		}
		metadata.Amounts = append(metadata.Amounts, matches[1])
	}

	return metadata
}

func extractPurchaseOrderMetadata(text string) domain.DocumentMetadata {
	metadata := domain.DocumentMetadata{
		Dates:   []string{},
		Amounts: []string{},
	}

	// Extract PO number
	poPattern := regexp.MustCompile(`(?i)(?:purchase\s*order|p\.?o\.?)[\s#:]*(\w+[-]?\d+)`)
	if matches := poPattern.FindStringSubmatch(text); len(matches) > 1 {
		metadata.InvoiceNumber = matches[1]
	}

	return metadata
}

func extractContractMetadata(text string) domain.DocumentMetadata {
	metadata := domain.DocumentMetadata{
		Dates:   []string{},
		Amounts: []string{},
	}

	// Extract contract value
	valuePattern := regexp.MustCompile(`(?i)(?:value|amount|price)[\s:]*[$€£]?\s*([\d,]+\.?\d*)`)
	if matches := valuePattern.FindStringSubmatch(text); len(matches) > 1 {
		amount := strings.ReplaceAll(matches[1], ",", "")
		if d, err := decimal.NewFromString(amount); err == nil {
			metadata.TotalAmount = d
		}
	}

	return metadata
}

func extractGenericMetadata(text string) domain.DocumentMetadata {
	metadata := domain.DocumentMetadata{
		Dates:   []string{},
		Amounts: []string{},
		Emails:  []string{},
	}

	// Extract emails
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	metadata.Emails = emailPattern.FindAllString(text, -1)

	return metadata
}

func parseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02",
		"2006/01/02",
		"01-02-2006",
		"01/02/2006",
		"02-01-2006",
		"02/01/2006",
		"Jan 2, 2006",
		"January 2, 2006",
		"2 Jan 2006",
		"2 January 2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
