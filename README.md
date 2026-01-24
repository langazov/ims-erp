# IMS ERP System

A comprehensive Enterprise Resource Planning (ERP) system built with Go microservices architecture.

## Architecture Overview

```
                                    ┌─────────────────┐
                                    │   API Gateway   │
                                    │    (Port 8080)  │
                                    └────────┬────────┘
                                             │
          ┌──────────────┬──────────────┬──────┴──────┬──────────────┐
          │              │              │              │              │
          ▼              ▼              ▼              ▼              ▼
    ┌──────────┐  ┌──────────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐
    │  Auth    │  │  Client      │  │ Invoice  │  │ Payment  │  │  Product    │
    │ Service  │  │  Services    │  │ Service  │  │ Service  │  │  Service    │
    │ (8081)   │  │  (8082)      │  │ (8083)   │  │ (8084)   │  │  (8085)     │
    └──────────┘  └──────────────┘  └──────────┘  └──────────┘  └──────────────┘
          │              │              │              │              │
          └──────────────┴──────┬──────┴──────────────┴──────────────┘
                                 │
          ┌──────────────────────┴──────────────────────┐
          │                                              │
          ▼                                              ▼
    ┌──────────────┐                            ┌──────────────┐
    │  Order       │                            │  Inventory  │
    │  Service     │                            │  Service    │
    │  (8086)      │                            │  (8087)      │
    └──────────────┘                            └──────────────┘
```

## Technology Stack

- **Language**: Go 1.25.6
- **Database**: MongoDB (event store, read models)
- **Cache**: Redis (caching, rate limiting, sessions)
- **Messaging**: NATS with JetStream (CQRS events)
- **Monitoring**: Prometheus metrics, OpenTelemetry tracing
- **API Gateway**: Reverse proxy with authentication
- **Configuration**: Viper with YAML support

## Services

| Service | Port | Description |
|---------|------|-------------|
| API Gateway | 8080 | Central entry point, routing, authentication |
| Auth Service | 8081 | User authentication, JWT tokens, sessions |
| Client Query Service | 8082 | Client read operations, event projection |
| Client Command Service | 8080 | Client write operations (CQRS) |
| Invoice Service | 8083 | Invoice creation, management, PDF generation |
| Payment Service | 8084 | Payment processing, refunds, webhooks |
| Product Service | 8085 | Product catalog, variants, pricing |
| Order Service | 8086 | Order management, fulfillment, shipping |
| Inventory Service | 8087 | Stock control, warehouses, reservations |

## Quick Start

### Prerequisites

- Go 1.25.6+
- MongoDB 7+
- Redis 7+
- NATS 2.10+

### Configuration

Copy the example configuration and customize:

```bash
cp config.yaml.example config.yaml
```

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Run Services

```bash
# Start infrastructure
docker-compose -f docker-compose.integration.yml up -d

# Start all services
make run-services

# Or start individual services
./bin/auth-service
./bin/client-command-service
./bin/client-query-service
./bin/invoice-service
./bin/payment-service
./bin/product-service
./bin/order-service
./bin/inventory-service
./bin/api-gateway
```

## API Documentation

### OpenAPI Spec

The complete API specification is available at:

- [OpenAPI 3.0 Specification](api/openapi.yaml)
- Swagger UI (when running docs service)

### Authentication

All protected endpoints require Bearer token authentication:

```http
Authorization: Bearer <your-jwt-token>
```

### Rate Limiting

- 1000 requests per minute per tenant
- Rate limit headers included in responses

## Project Structure

```
ims-erp/
├── api/                    # API specifications
│   └── openapi.yaml       # OpenAPI 3.0 spec
├── cmd/                    # Service entry points
│   ├── api-gateway/
│   ├── auth-service/
│   ├── client-command-service/
│   ├── client-query-service/
│   ├── invoice-service/
│   ├── payment-service/
│   ├── product-service/
│   ├── order-service/
│   └── inventory-service/
├── internal/               # Application logic
│   ├── auth/              # Authentication
│   ├── commands/          # Command handlers
│   ├── config/            # Configuration
│   ├── domain/            # Domain models
│   ├── events/            # Event handlers
│   ├── integration/       # Integration tests
│   ├── messaging/         # NATS messaging
│   ├── middleware/        # HTTP middleware
│   ├── queries/           # Query handlers
│   ├── rbac/              # Role-based access control
│   └── repository/        # Data access
├── pkg/                    # Shared libraries
│   ├── errors/
│   ├── logger/
│   ├── metrics/
│   └── tracer/
├── deployments/           # Kubernetes/Helm
├── scripts/              # Build scripts
├── Makefile
├── docker-compose.integration.yml
└── config.yaml.example
```

## Key Features

### CQRS Pattern
- Command and Query Responsibility Segregation
- Event sourcing for audit trail
- Read models optimized for queries

### Multi-Tenancy
- Tenant isolation at data level
- Tenant-scoped RBAC
- Tenant-specific configurations

### Event-Driven Architecture
- NATS JetStream for messaging
- Event projections for read models
- eventual consistency

### Monitoring
- Prometheus metrics endpoints
- OpenTelemetry distributed tracing
- Health, readiness, liveness probes

### Security
- JWT-based authentication
- Refresh token rotation
- Rate limiting
- CORS support

## Development

### Adding a New Service

1. Create service directory in `cmd/`
2. Implement main.go with health endpoints
3. Add domain models in `internal/domain/`
4. Add command/query handlers if needed
5. Update API gateway routing
6. Add to docker-compose and Kubernetes manifests

### Running Tests

```bash
# Unit tests
make test

# Integration tests
docker-compose -f docker-compose.integration.yml up -d
make test-integration

# With coverage
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run vet
make vet
```

## Deployment

### Kubernetes

```bash
# Deploy to Kubernetes
kubectl apply -f deployments/kubernetes/

# Or use Helm
helm install erp-system deployments/helm/erp-system/
```

### Docker

```bash
# Build images
make docker-build

# Push images
make docker-push
```

## License

MIT
