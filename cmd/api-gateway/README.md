# API Gateway

Reverse proxy and routing service for the ERP microservices.

## Overview

The API Gateway serves as the single entry point for all client requests. It handles:
- Request routing to backend services
- Authentication and authorization
- Rate limiting
- Request/response transformation
- Load balancing

## Architecture

```
                    +-----------------+
                    |   API Gateway   |
                    |     (8080)      |
                    +--------+--------+
                             |
        +--------------------+--------------------+
        |                    |                    |
+--------+--------+  +-------+--------+  +--------+--------+
| Auth Service    |  | Client Command  |  | Client Query   |
|    (8081)       |  |    Service      |  |    Service     |
+-----------------+  |    (8082)       |  |    (8083)      |
                     +-----------------+  +----------------+
        +--------------------+--------------------+
        |                    |                    |
+--------+--------+  +-------+--------+  +--------+--------+
| Invoice Service |  | Payment Service |  | Product Service |
|    (8084)       |  |    (8085)       |  |    (8086)       |
+-----------------+  +-----------------+  +----------------+
        +--------------------+--------------------+
        |                    |
+--------+--------+  +-------+--------+
| Order Service   |  | Inventory       |
|    (8087)       |  |    Service      |
+-----------------+  |    (8088)       |
                     +-----------------+
```

## Routes

### Authentication

| Method | Path | Service | Description |
|--------|------|---------|-------------|
| POST | `/api/v1/auth/*` | auth-service | Authentication endpoints |
| GET | `/api/v1/roles/*` | auth-service | RBAC endpoints |

### Clients

| Method | Path | Service | Description |
|--------|------|---------|-------------|
| POST | `/api/v1/commands` | client-command-service | Client commands |
| GET | `/api/v1/clients/*` | client-query-service | Client queries |

### Invoicing

| Method | Path | Service | Description |
|--------|------|---------|-------------|
| GET/POST/PUT/DELETE | `/api/v1/invoices/*` | invoice-service | Invoice management |
| GET/POST/PUT/DELETE | `/api/v1/invoice-lines/*` | invoice-service | Invoice line items |

### Payments

| Method | Path | Service | Description |
|--------|------|---------|-------------|
| GET/POST/PUT/DELETE | `/api/v1/payments/*` | payment-service | Payment processing |
| GET/POST/PUT/DELETE | `/api/v1/payment-methods/*` | payment-service | Payment methods |
| GET/POST/PUT/DELETE | `/api/v1/refunds/*` | payment-service | Refunds |

### Products

| Method | Path | Service | Description |
|--------|------|---------|-------------|
| GET/POST/PUT/DELETE | `/api/v1/products/*` | product-service | Product catalog |
| GET/POST/PUT/DELETE | `/api/v1/categories/*` | product-service | Categories |
| GET/POST/PUT/DELETE | `/api/v1/product-pricing/*` | product-service | Pricing |

### Orders

| Method | Path | Service | Description |
|--------|------|---------|-------------|
| GET/POST/PUT/DELETE | `/api/v1/orders/*` | order-service | Order management |
| GET/POST/PUT/DELETE | `/api/v1/fulfillments/*` | order-service | Fulfillment |

### Inventory

| Method | Path | Service | Description |
|--------|------|---------|-------------|
| GET/POST/PUT/DELETE | `/api/v1/inventory/*` | inventory-service | Inventory management |
| GET/POST/PUT/DELETE | `/api/v1/warehouses/*` | inventory-service | Warehouses |

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| AUTH_SERVICE_URL | Auth service URL | http://localhost:8081 |
| CLIENT_COMMAND_SERVICE_URL | Client command service URL | http://localhost:8082 |
| CLIENT_QUERY_SERVICE_URL | Client query service URL | http://localhost:8083 |
| INVOICE_SERVICE_URL | Invoice service URL | http://localhost:8084 |
| PAYMENT_SERVICE_URL | Payment service URL | http://localhost:8085 |
| PRODUCT_SERVICE_URL | Product service URL | http://localhost:8086 |
| ORDER_SERVICE_URL | Order service URL | http://localhost:8087 |
| INVENTORY_SERVICE_URL | Inventory service URL | http://localhost:8088 |
| JWT_SECRET | JWT signing secret | - |
| RATE_LIMIT | Requests per minute | 1000 |

## Middleware

1. **Authentication** - Validates JWT tokens
2. **Authorization** - Checks RBAC permissions
3. **Rate Limiting** - Controls request rate
4. **Logging** - Logs all requests
5. **Tracing** - OpenTelemetry tracing
6. **Metrics** - Prometheus metrics

## Running

```bash
go run cmd/api-gateway/main.go
```

## Testing

```bash
make test
```
