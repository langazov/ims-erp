# Document Service

Document management service for the ERP system with MinIO storage, Elasticsearch search, and OCR processing capabilities.

## Features

- **Document Upload**: Presigned URL uploads for direct browser-to-storage transfers
- **Document Storage**: MinIO/S3-compatible object storage with versioning
- **Full-Text Search**: Elasticsearch-powered document search with highlighting
- **Metadata Extraction**: Automatic extraction of invoice numbers, dates, amounts
- **OCR Processing**: Tesseract integration for scanned documents
- **Thumbnail Generation**: Automatic preview image generation
- **Tagging**: Flexible document tagging and categorization

## Architecture

```
cmd/document-service/
├── main.go              # Service entry point and HTTP handlers
internal/
└── domain/
    └── document.go      # Domain models and interfaces
```

## Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `SERVICE_PORT` | HTTP server port | `8086` |
| `MONGO_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `MONGO_DATABASE` | MongoDB database name | `erp_documents` |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | `` |
| `MINIO_ENDPOINT` | MinIO server endpoint | `localhost:9000` |
| `MINIO_ACCESS_KEY` | MinIO access key | `` |
| `MINIO_SECRET_KEY` | MinIO secret key | `` |
| `MINIO_USE_SSL` | Use SSL for MinIO | `false` |
| `ELASTICSEARCH_URL` | Elasticsearch URL | `` |
| `MAX_FILE_SIZE` | Maximum upload size (bytes) | `52428800` (50MB) |
| `PRESIGNED_EXPIRY` | Presigned URL expiry duration | `1h` |
| `LOG_LEVEL` | Logging level | `info` |

## API Endpoints

### Health & Monitoring

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
| GET | `/metrics` | Prometheus metrics |

### Document Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/documents/upload` | Initiate presigned URL upload |
| POST | `/api/v1/documents` | Create document metadata |
| GET | `/api/v1/documents` | List documents |
| GET | `/api/v1/documents/{id}` | Get document metadata |
| PUT | `/api/v1/documents/{id}` | Update document metadata |
| DELETE | `/api/v1/documents/{id}` | Delete document |
| GET | `/api/v1/documents/{id}/download` | Download document |
| GET | `/api/v1/documents/{id}/thumbnail` | Get document thumbnail |
| GET | `/api/v1/documents/{id}/presigned-url` | Get download URL |
| PUT | `/api/v1/documents/{id}/tags` | Update document tags |
| POST | `/api/v1/documents/{id}/reprocess` | Reprocess document |

### Search

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/documents/search` | Full-text search |
| GET | `/api/v1/documents/search/suggest` | Autocomplete suggestions |

### Multipart Upload (Coming Soon)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/documents/multipart/start` | Start multipart upload |
| PUT | `/api/v1/documents/multipart/{uploadId}/part` | Upload part |
| POST | `/api/v1/documents/multipart/{uploadId}/complete` | Complete upload |

## Document Types

| Type | Description |
|------|-------------|
| `invoice` | Invoice documents |
| `purchase_order` | Purchase orders |
| `receipt` | Receipts |
| `contract` | Contracts |
| `scanned` | Scanned documents |
| `other` | Other document types |

## Processing Status

| Status | Description |
|--------|-------------|
| `pending` | Awaiting processing |
| `processing` | Currently being processed |
| `completed` | Processing complete |
| `failed` | Processing failed |

## Quick Start

```bash
# Start with default configuration
go run cmd/document-service/main.go

# With custom configuration
SERVICE_PORT=8086 \
MONGO_URI=mongodb://localhost:27017 \
MINIO_ENDPOINT=localhost:9000 \
go run cmd/document-service/main.go
```

## Upload Flow

1. Client sends upload request to `/api/v1/documents/upload`
2. Server returns presigned URL for direct upload
3. Client uploads file directly to MinIO
4. Client creates document record via `/api/v1/documents`
5. Document is queued for processing (OCR, text extraction)

## Building

```bash
# Build binary
go build -o bin/document-service ./cmd/document-service

# Build with version info
go build -ldflags="-X main.Version=1.0.0" -o bin/document-service ./cmd/document-service
```

## Dependencies

- **MongoDB**: Document metadata storage
- **Redis**: Caching and rate limiting
- **MinIO**: Object storage for files
- **Elasticsearch**: Full-text search
- **Tesseract**: OCR processing (optional)

## Related Services

- [Warehouse Service](../warehouse-service/) - For document-linked warehouse operations
- [Invoice Service](../invoice-service/) - Invoice document generation
