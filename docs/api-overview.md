# ERP System OpenAPI Specification

This document describes the REST API for the ERP System.

## Overview

The ERP System API provides a comprehensive set of endpoints for managing enterprise resources including:

- **Authentication** (`/api/v1/auth/*`) - User registration, login, token management
- **Clients** (`/api/v1/clients/*`) - Client management with CQRS pattern
- **Invoices** (`/api/v1/invoices/*`) - Invoice creation and management
- **Payments** (`/api/v1/payments/*`) - Payment processing with multiple providers
- **Products** (`/api/v1/products/*`) - Product catalog management
- **Orders** (`/api/v1/orders/*`) - Order processing and fulfillment
- **Inventory** (`/api/v1/inventory/*`) - Stock and warehouse management

## Authentication

All API endpoints require authentication via JWT tokens. Include the token in the Authorization header:

```
Authorization: Bearer <token>
```

## Rate Limiting

- 1000 requests per minute for authenticated requests
- 100 requests per minute for unauthenticated requests

## Response Format

All responses follow a consistent format:

```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100
  }
}
```

## Error Handling

Errors return appropriate HTTP status codes:

- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `422` - Validation Error
- `429` - Rate Limit Exceeded
- `500` - Internal Server Error

## Versioning

The API uses path versioning: `/api/v1/`

---

## Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/refresh` | Refresh tokens |
| POST | `/api/v1/auth/logout` | User logout |
| GET | `/api/v1/auth/me` | Get current user |

### Clients

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/clients` | List clients |
| POST | `/api/v1/commands` | Execute command |
| GET | `/api/v1/clients/:id` | Get client |

### Invoices

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/invoices` | List invoices |
| POST | `/api/v1/invoices` | Create invoice |
| GET | `/api/v1/invoices/:id` | Get invoice |
| PUT | `/api/v1/invoices/:id` | Update invoice |
| POST | `/api/v1/invoices/:id/send` | Send invoice |

### Payments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/payments` | List payments |
| POST | `/api/v1/payments` | Create payment |
| POST | `/api/v1/payments/:id/refund` | Refund payment |

### Products

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/products` | List products |
| POST | `/api/v1/products` | Create product |
| GET | `/api/v1/products/:id` | Get product |
| PUT | `/api/v1/products/:id` | Update product |

### Orders

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/orders` | List orders |
| POST | `/api/v1/orders` | Create order |
| GET | `/api/v1/orders/:id` | Get order |
| POST | `/api/v1/orders/:id/fulfill` | Fulfill order |

### Inventory

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/inventory` | List inventory |
| POST | `/api/v1/inventory/adjust` | Adjust stock |
| POST | `/api/v1/inventory/reserve` | Reserve stock |
