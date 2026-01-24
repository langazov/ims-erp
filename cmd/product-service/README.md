# Product Service

Product catalog and inventory management service.

## API Endpoints

### Products

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/products` | List products |
| POST | `/api/v1/products` | Create product |
| GET | `/api/v1/products/:id` | Get product by ID |
| PUT | `/api/v1/products/:id` | Update product |
| DELETE | `/api/v1/products/:id` | Delete product |
| GET | `/api/v1/products/search` | Search products |
| GET | `/api/v1/products/:id/variants` | Get product variants |
| POST | `/api/v1/products/:id/variants` | Add variant |
| PUT | `/api/v1/products/:id/variants/:variantId` | Update variant |
| DELETE | `/api/v1/products/:id/variants/:variantId` | Delete variant |

### Categories

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/categories` | List categories |
| POST | `/api/v1/categories` | Create category |
| GET | `/api/v1/categories/:id` | Get category by ID |
| PUT | `/api/v1/categories/:id` | Update category |
| DELETE | `/api/v1/categories/:id` | Delete category |

### Pricing

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/products/:id/pricing` | Get product pricing |
| PUT | `/api/v1/products/:id/pricing` | Update pricing |
| POST | `/api/v1/products/:id/pricing/tiers` | Add price tier |

## Create Product

```json
POST /api/v1/products
{
  "name": "Product Name",
  "sku": "SKU-001",
  "description": "Product description",
  "basePrice": 99.99,
  "currency": "USD",
  "categoryId": "uuid",
  "tags": ["tag1", "tag2"],
  "attributes": {
    "color": "red",
    "size": "large"
  },
  "variants": [
    {
      "name": "Red Large",
      "sku": "SKU-001-RED-L",
      "price": 99.99,
      "attributes": {
        "color": "red",
        "size": "large"
      }
    }
  ],
  "inventory": {
    "trackInventory": true,
    "lowStockThreshold": 10
  }
}
```

## Search Products

```
GET /api/v1/products/search?q=product&category=uuid&minPrice=10&maxPrice=100&inStock=true
```

## Events

The service emits the following events:

- `ProductCreated` - When product is created
- `ProductUpdated` - When product is modified
- `ProductDeleted` - When product is deleted
- `VariantCreated` - When variant is added
- `VariantUpdated` - When variant is modified
- `PriceUpdated` - When pricing changes
- `InventoryUpdated` - When stock levels change

## Running

```bash
go run cmd/product-service/main.go
```

## Testing

```bash
make test
```
