# Payment Service

Payment processing service supporting multiple payment providers.

## API Endpoints

### Payments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/payments` | List payments |
| POST | `/api/v1/payments` | Create payment |
| GET | `/api/v1/payments/:id` | Get payment by ID |
| POST | `/api/v1/payments/:id/refund` | Refund payment |
| POST | `/api/v1/payments/:id/capture` | Capture authorized payment |
| POST | `/api/v1/payments/:id/void` | Void payment |

### Payment Methods

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/payment-methods` | List payment methods |
| POST | `/api/v1/payment-methods` | Add payment method |
| DELETE | `/api/v1/payment-methods/:id` | Remove payment method |
| PUT | `/api/v1/payment-methods/:id/default` | Set as default |

### Refunds

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/refunds` | List refunds |
| GET | `/api/v1/refunds/:id` | Get refund by ID |

## Create Payment

```json
POST /api/v1/payments
{
  "amount": 100.00,
  "currency": "USD",
  "clientId": "uuid",
  "invoiceId": "uuid",
  "paymentMethod": {
    "type": "card",
    "provider": "stripe",
    "token": "tok_visa"
  },
  "description": "Payment for invoice INV-2024-001"
}
```

## Supported Providers

### Stripe

```json
{
  "provider": "stripe",
  "token": "tok_visa",
  "returnUrl": "https://example.com/payment/complete"
}
```

### PayPal

```json
{
  "provider": "paypal",
  "orderId": "PAYPAL_ORDER_ID",
  "returnUrl": "https://example.com/payment/complete",
  "cancelUrl": "https://example.com/payment/cancel"
}
```

## Payment Status

- `pending` - Payment initiated
- `processing` - Payment being processed
- `succeeded` - Payment successful
- `failed` - Payment failed
- `refunded` - Payment refunded
- `voided` - Payment voided

## Running

```bash
go run cmd/payment-service/main.go
```

## Testing

```bash
make test
```
