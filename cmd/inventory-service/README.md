# Inventory Service

Inventory and warehouse management service.

## API Endpoints

### Inventory

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/inventory` | List inventory items |
| GET | `/api/v1/inventory/:id` | Get inventory by ID |
| GET | `/api/v1/inventory/product/:productId` | Get inventory by product |
| POST | `/api/v1/inventory/adjust` | Adjust inventory |
| POST | `/api/v1/inventory/transfer` | Transfer between warehouses |
| GET | `/api/v1/inventory/levels` | Get stock levels |

### Warehouses

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/warehouses` | List warehouses |
| POST | `/api/v1/warehouses` | Create warehouse |
| GET | `/api/v1/warehouses/:id` | Get warehouse by ID |
| PUT | `/api/v1/warehouses/:id` | Update warehouse |
| DELETE | `/api/v1/warehouses/:id` | Delete warehouse |

### Reservations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/inventory/reserve` | Reserve inventory |
| POST | `/api/v1/inventory/release` | Release reservation |
| POST | `/api/v1/inventory/commit` | Commit reservation |
| GET | `/api/v1/inventory/reservations` | List reservations |

### Stock Movements

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/inventory/movements` | List movements |
| POST | `/api/v1/inventory/movements` | Create movement |

## Adjust Inventory

```json
POST /api/v1/inventory/adjust
{
  "productId": "uuid",
  "variantId": "uuid",
  "warehouseId": "uuid",
  "quantity": 10,
  "type": "addition",
  "reason": "Stock received from supplier",
  "reference": "PO-2024-001"
}
```

## Transfer Inventory

```json
POST /api/v1/inventory/transfer
{
  "productId": "uuid",
  "variantId": "uuid",
  "fromWarehouseId": "uuid-1",
  "toWarehouseId": "uuid-2",
  "quantity": 5,
  "reference": "TR-2024-001"
}
```

## Reserve Inventory

```json
POST /api/v1/inventory/reserve
{
  "productId": "uuid",
  "variantId": "uuid",
  "warehouseId": "uuid",
  "quantity": 2,
  "orderId": "uuid",
  "expiresAt": "2024-01-02T00:00:00Z"
}
```

## Running

```bash
go run cmd/inventory-service/main.go
```

## Testing

```bash
make test
```
