# ERP System Development Environment

A complete local development environment for the ERP System microservices architecture.

## Quick Start

### Option A: Full Docker Compose Dev Stack (Backend + Frontend)

```bash
docker compose -f docker-compose.dev.yml up --build
```

This starts MongoDB, Redis, NATS, backend services, API Gateway, and frontend.

To stop:

```bash
docker compose -f docker-compose.dev.yml down
```

### Option B: Infrastructure + Local Services

### 1. Start Infrastructure (MongoDB, Redis, NATS)

```bash
docker-compose -f docker-compose.integration.yml up -d
```

### 2. Start All Backend Services

```bash
# Using the dev script (recommended)
./scripts/dev.sh start backend

# Or manually start each service:
cd cmd/auth-service && go run main.go &
cd cmd/client-query-service && go run main.go &
cd cmd/inventory-service && go run main.go &
cd cmd/order-service && go run main.go &
cd cmd/product-service && go run main.go &
cd cmd/api-gateway && go run main.go &
```

### 3. Start Frontend

```bash
cd frontend
npm run dev
```

### 4. Login to the Application

Open http://localhost:5173/login and use the default admin credentials:

| Field | Value |
|-------|-------|
| Email | admin@erp.local |
| Password | Admin123! |
| Tenant ID | (leave empty - defaults to demo tenant) |

**Note:** The Tenant ID field can be left empty for the demo tenant. The demo tenant is automatically used if no tenant ID is provided.

## Default Admin Account

A default admin account is automatically seeded when running the seed script:

```bash
go run scripts/seed-admin.go -mongodb "mongodb://localhost:27017" -database erp_system
```

**Credentials:**
- **Email:** admin@erp.local
- **Password:** Admin123!
- **Tenant ID:** 00000000-0000-0000-0000-000000000001
- **Role:** admin (full system access)

## Available Commands

```bash
./scripts/dev.sh start                 # Start full environment (infra + backend + frontend)
./scripts/dev.sh stop                  # Stop full environment
./scripts/dev.sh restart               # Restart full environment
./scripts/dev.sh start infra           # Start infrastructure only
./scripts/dev.sh start backend         # Start all backend services
./scripts/dev.sh start frontend        # Start frontend only
./scripts/dev.sh start api-gateway     # Start a single backend service
./scripts/dev.sh stop order-service    # Stop a single backend service
./scripts/dev.sh status backend        # Check backend status
./scripts/dev.sh logs frontend         # View frontend logs
./scripts/dev.sh logs order-service    # View service logs
```

## Service Ports

| Service | Port | Description |
|---------|------|-------------|
| API Gateway | 8080 | Main API entry point |
| Auth Service | 8081 | Authentication & authorization |
| Client Query Service | 8082 | Client data queries |
| Product Service | 8085 | Product catalog management |
| Inventory Service | 8084 | Inventory management |
| Order Service | 8086 | Order processing |

## Infrastructure Ports

| Service | Port | Protocol |
|---------|------|----------|
| MongoDB | 27017 | TCP |
| Redis | 6379 | TCP |
| NATS | 4222 | TCP |

## API Endpoints

### Health Checks

```bash
curl http://localhost:8080/health  # API Gateway
curl http://localhost:8081/health  # Auth Service
curl http://localhost:8082/health  # Client Query Service
curl http://localhost:8084/health  # Inventory Service
curl http://localhost:8085/health  # Product Service
curl http://localhost:8086/health  # Order Service
```

### Backend Service APIs

```bash
# Clients
curl http://localhost:8082/api/v1/clients?tenantId=test

# Inventory
curl http://localhost:8084/api/v1/inventory/items?tenantId=test

# Products
curl http://localhost:8085/api/v1/products

# Orders
curl http://localhost:8086/api/v1/orders
```

### Frontend URLs

| Route | URL |
|-------|-----|
| Login | http://localhost:5173/login |
| Dashboard | http://localhost:5173/ |
| Clients | http://localhost:5173/clients |
| Products | http://localhost:5173/products |
| Inventory | http://localhost:5173/inventory |
| Orders | http://localhost:5173/orders |

## Configuration

Each service has its own configuration file:

- `cmd/api-gateway/api-gateway.yaml`
- `cmd/auth-service/auth-service.yaml`
- `cmd/client-query-service/client-query-service.yaml`
- `cmd/inventory-service/inventory-service.yaml`
- `cmd/order-service/order-service.yaml`
- `cmd/product-service/product-service.yaml`

## Troubleshooting

### Services not starting

1. Check if ports are already in use:
   ```bash
   lsof -i :8080
   ```

2. Check service logs:
   ```bash
   ./scripts/dev.sh logs
   ```

### MongoDB connection failed

Ensure MongoDB is running:
```bash
docker ps | grep mongo
```

### NATS connection issues

Ensure NATS is running on port 4222:
```bash
nats-server -p 4222 &
```

## Environment Variables

For frontend, create `frontend/.env`:

```env
VITE_API_URL=http://localhost:8080/api/v1
VITE_AUTH_ENABLED=true
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend (SvelteKit)                      │
│                      http://localhost:5173                   │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway                               │
│                      http://localhost:8080                   │
└─────────┬─────────┬─────────┬─────────┬─────────┬───────────┘
          │         │         │         │         │
          ▼         ▼         ▼         ▼         ▼
    ┌──────────┐┌──────────┐┌──────────┐┌──────────┐┌──────────┐
    │  Auth    ││  Client  ││Inventory ││ Product  ││  Order   │
    │ Service  ││ Service  ││ Service  ││ Service  ││ Service  │
    │ :8081    ││ :8082    ││ :8084    ││ :8085    ││ :8086    │
    └────┬─────┘└────┬─────┘└────┬─────┘└────┬─────┘└────┬─────┘
         │           │           │           │           │
         └───────────┴─────┬─────┴───────────┴───────────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
         ▼                 ▼                 ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │ MongoDB  │    │  Redis   │    │   NATS   │
    │ :27017   │    │  :6379   │    │  :4222   │
    └──────────┘    └──────────┘    └──────────┘
```
