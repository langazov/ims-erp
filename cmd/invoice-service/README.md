# Invoice Service

Invoice management service for the ERP system.

## API Endpoints

### Invoices

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/invoices` | List invoices |
| POST | `/api/v1/invoices` | Create invoice |
| GET | `/api/v1/invoices/:id` | Get invoice by ID |
| PUT | `/api/v1/invoices/:id` | Update invoice |
| DELETE | `/api/v1/invoices/:id` | Delete invoice |
| POST | `/api/v1/invoices/:id/send` | Send invoice to client |
| POST | `/api/v1/invoices/:id/void` | Void invoice |
| POST | `/api/v1/invoices/:id/refund` | Issue refund |

### Invoice Lines

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/invoices/:id/lines` | Add line item |
| PUT | `/api/v1/invoices/:id/lines/:lineId` | Update line item |
| DELETE | `/api/v1/invoices/:id/lines/:lineId` | Remove line item |

### Payments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/invoices/:id/payments` | Get invoice payments |
| POST | `/api/v1/invoices/:id/payments` | Record payment |

## Create Invoice

```json
POST /api/v1/invoices
{
  "clientId": "uuid",
  "invoiceNumber": "INV-2024-001",
  "currency": "USD",
  "taxRate": 10,
  "paymentTerms": "net30",
  "dueDate": "2024-02-01",
  "lines": [
    {
      "description": "Product A",
      "quantity": 2,
      "unitPrice": 100.00,
      "productId": "uuid"
    }
  ],
  "notes": "Thank you for your business"
}
```

## Events

The service emits the following events:

- `InvoiceCreated` - When invoice is created
- `InvoiceUpdated` - When invoice is modified
- `InvoiceSent` - When invoice is sent to client
- `InvoicePaid` - When payment is received
- `InvoiceVoided` - When invoice is voided
- `InvoiceRefunded` - When refund is issued

## Running

```bash
go run cmd/invoice-service/main.go
```

## Testing

```bash
make test
```
