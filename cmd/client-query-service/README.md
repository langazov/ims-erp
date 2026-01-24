# Client Query Service

Read model service for client data implementing CQRS query side.

## API Endpoints

### Queries

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/clients` | List clients with pagination |
| GET | `/api/v1/clients/search?q=` | Search clients |
| GET | `/api/v1/clients/:id` | Get client by ID |
| GET | `/api/v1/clients/:id/detail` | Get full client details |
| GET | `/api/v1/clients/:id/credit` | Get credit status |
| GET | `/api/v1/clients/:id/orders` | Get client orders |
| GET | `/api/v1/clients/:id/invoices` | Get client invoices |
| GET | `/api/v1/clients/:id/payments` | Get client payments |

## Query Parameters

### List Clients

```
GET /api/v1/clients?page=1&limit=20&sort=name&status=active
```

| Parameter | Type | Description |
|-----------|------|-------------|
| page | int | Page number (default: 1) |
| limit | int | Items per page (default: 20, max: 100) |
| sort | string | Sort field |
| order | string | Sort order (asc/desc) |
| status | string | Filter by status |
| tags | string | Filter by tags (comma-separated) |

### Search Clients

```
GET /api/v1/clients/search?q=company&field=name,email
```

## Response Format

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Client Name",
      "email": "client@example.com",
      "phone": "+1234567890",
      "creditLimit": 10000,
      "creditUsed": 2500,
      "status": "active",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

## Running

```bash
go run cmd/client-query-service/main.go
```

## Testing

```bash
make test
```
