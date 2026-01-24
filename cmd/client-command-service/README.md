# Client Service

Client management service implementing CQRS pattern.

## API Endpoints

### Queries

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/clients` | List clients with pagination |
| GET | `/api/v1/clients/search?q=` | Search clients |
| GET | `/api/v1/clients/id/?clientId=` | Get client by ID |
| GET | `/api/v1/clients/detail/?clientId=` | Get client detail |
| GET | `/api/v1/clients/credit/?clientId=` | Get credit status |

### Commands

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/commands` | Execute client command |

## Commands

### CreateClient
```json
{
  "type": "client.create",
  "tenantId": "uuid",
  "userId": "uuid",
  "data": {
    "name": "Client Name",
    "email": "client@example.com",
    "phone": "+1234567890",
    "creditLimit": 10000,
    "billingAddress": {
      "street": "123 Main St",
      "city": "New York",
      "state": "NY",
      "postalCode": "10001",
      "country": "USA"
    },
    "tags": ["vip", "enterprise"]
  }
}
```

### UpdateClient
```json
{
  "type": "client.update",
  "tenantId": "uuid",
  "userId": "uuid",
  "aggregateId": "client-uuid",
  "data": {
    "name": "Updated Name",
    "email": "updated@example.com"
  }
}
```

### DeactivateClient
```json
{
  "type": "client.deactivate",
  "tenantId": "uuid",
  "userId": "uuid",
  "aggregateId": "client-uuid",
  "data": {
    "reason": "Customer request"
  }
}
```

### AssignCreditLimit
```json
{
  "type": "client.assign_credit_limit",
  "tenantId": "uuid",
  "userId": "uuid",
  "aggregateId": "client-uuid",
  "data": {
    "newLimit": 20000,
    "reason": "Increased creditworthiness"
  }
}
```

## Events

The service emits the following events:

- `ClientCreated` - When a new client is created
- `ClientUpdated` - When client details are updated
- `ClientDeactivated` - When a client is deactivated
- `CreditLimitAssigned` - When credit limit changes
- `BillingInfoUpdated` - When billing address changes
- `ClientsMerged` - When clients are merged

## Running

```bash
go run cmd/client-command-service/main.go
```

## Testing

```bash
make test
```
