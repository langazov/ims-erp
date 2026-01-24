# Order Service

Order management service for the ERP system.

## API Endpoints

### Orders

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/orders` | List orders |
| POST | `/api/v1/orders` | Create order |
| GET | `/api/v1/orders/:id` | Get order by ID |
| PUT | `/api/v1/orders/:id` | Update order |
| POST | `/api/v1/orders/:id/cancel` | Cancel order |
| POST | `/api/v1/orders/:id/fulfill` | Fulfill order |
| POST | `/api/v1/orders/:id/return` | Return order |

### Order Lines

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/orders/:id/lines` | Add order line |
| PUT | `/api/v1/orders/:id/lines/:lineId` | Update order line |
| DELETE | `/api/v1/orders/:id/lines/:lineId` | Remove order line |

### Fulfillment

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/orders/:id/fulfillments` | Get fulfillments |
| POST | `/api/v1/orders/:id/fulfillments` | Create fulfillment |
| PUT | `/api/v1/fulfillments/:id` | Update fulfillment status |

## Create Order

```json
POST /api/v1/orders
{
  "clientId": "uuid",
  "currency": "USD",
  "shippingAddress": {
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "postalCode": "10001",
    "country": "USA"
  },
  "billingAddress": {
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "postalCode": "10001",
    "country": "USA"
  },
  "lines": [
    {
      "productId": "uuid",
      "variantId": "uuid",
      "quantity": 2,
      "unitPrice": 99.99
    }
  ],
  "shippingMethod": "standard",
  "notes": "Gift wrap please"
}
```

## Order Status

- `draft` - Order being created
- `pending` - Order confirmed, awaiting payment
- `paid` - Payment received
- `processing` - Being prepared
- `shipped` - Shipped to customer
- `delivered` - Delivered
- `completed` - Order complete
- `cancelled` - Order cancelled
- `returned` - Items returned

## Running

```bash
go run cmd/order-service/main.go
```

## Testing

```bash
make test
```
