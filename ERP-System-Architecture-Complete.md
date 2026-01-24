# Enterprise ERP System Architecture
## Complete Technical Specification

**Version:** 3.0  
**Last Updated:** January 2026  
**Status:** Production Ready

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [System Overview](#2-system-overview)
3. [Core Architecture Patterns](#3-core-architecture-patterns)
4. [Module Specifications](#4-module-specifications)
5. [Messaging Infrastructure (NATS)](#5-messaging-infrastructure-nats)
6. [Data Layer](#6-data-layer)
7. [File Storage & Document Processing](#7-file-storage--document-processing)
8. [Plugin System](#8-plugin-system)
9. [Caching Strategy](#9-caching-strategy)
10. [Security Architecture](#10-security-architecture)
11. [Scaling & Multi-Tenancy](#11-scaling--multi-tenancy)
12. [Deployment & Operations](#12-deployment--operations)
13. [Observability](#13-observability)
14. [Disaster Recovery](#14-disaster-recovery)
15. [Capacity Planning](#15-capacity-planning)
16. [Technology Stack](#16-technology-stack)

---

## 1. Executive Summary

This document describes a scalable, event-driven Enterprise Resource Planning (ERP) system built on microservices architecture using the CQRS (Command Query Responsibility Segregation) pattern. The system is designed to handle millions of clients and enterprise-grade workloads.

### Key Features

| Feature | Description |
|---------|-------------|
| **CQRS Pattern** | Separate read/write models for optimal performance |
| **Event Sourcing** | Complete audit trail with event replay capability |
| **Multi-Tenancy** | 4-tier isolation from shared to dedicated infrastructure |
| **Customer 360°** | Complete CRM with leads, opportunities, tickets, health scoring |
| **Pluggable Payments** | Support for Stripe, PayPal, Adyen, and custom processors |
| **Document Management** | MinIO storage with OCR and full-text search |
| **Plugin Architecture** | Extensible system for custom modules |
| **Global Scale** | Multi-region deployment with NATS supercluster |

### Core Modules

- **Client Module** - Basic client/account management with credit limits
- **Customer Management (CRM)** - Full CRM with contacts, leads, opportunities, tickets, health scoring
- **Invoicing Module** - Invoice lifecycle with credit notes and aging reports
- **Payment Module** - Multi-processor payments with routing rules
- **Warehouse Module** - Multi-location warehouse operations
- **Inventory Module** - Stock management with reservations
- **Document Module** - File storage, OCR, and search

---

## 2. System Overview

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              API GATEWAY LAYER                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   REST API  │  │  GraphQL    │  │  WebSocket  │  │  Plugin API Gateway     │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           AUTHENTICATION & AUTHORIZATION                         │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────────┐  │
│  │    Auth Service     │  │   RBAC Service      │  │   Tenant Service        │  │
│  │   (OAuth 2.0/OIDC)  │  │  (Permissions)      │  │  (Multi-tenancy)        │  │
│  └─────────────────────┘  └─────────────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              NATS JETSTREAM                                      │
│                    (Event Bus / Message Broker / Command Bus)                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │   COMMANDS   │  │   EVENTS     │  │   QUERIES    │  │   DOCUMENT_EVENTS    │ │
│  │   Stream     │  │   Stream     │  │   Stream     │  │   Stream             │ │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
        ┌───────────────┬───────────────┼───────────────┬───────────────┬─────────────────┬─────────────────┐
        ▼               ▼               ▼               ▼               ▼                 ▼                 ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   CLIENT     │ │  CUSTOMER    │ │  INVOICING   │ │   PAYMENT    │ │  WAREHOUSE   │ │  INVENTORY   │ │  DOCUMENT    │
│   MODULE     │ │  MGMT (CRM)  │ │   MODULE     │ │   MODULE     │ │   MODULE     │ │   MODULE     │ │   MODULE     │
│              │ │              │ │              │ │              │ │              │ │              │ │              │
│ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │
│ │ Command  │ │ │ │ Command  │ │ │ │ Command  │ │ │ │ Command  │ │ │ │ Command  │ │ │ │ Command  │ │ │ │ Upload   │ │
│ │ Service  │ │ │ │ Service  │ │ │ │ Service  │ │ │ │ Service  │ │ │ │ Service  │ │ │ │ Service  │ │ │ │ Service  │ │
│ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │
│ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │
│ │  Query   │ │ │ │  Query   │ │ │ │  Query   │ │ │ │  Query   │ │ │ │  Query   │ │ │ │  Query   │ │ │ │Processor │ │
│ │ Service  │ │ │ │ Service  │ │ │ │ Service  │ │ │ │ Service  │ │ │ │ Service  │ │ │ │ Service  │ │ │ │ Service  │ │
│ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
        │               │               │               │               │                 │                 │
        ▼               ▼               ▼               ▼               ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                          DATA LAYER                                                  │
│                                                                                                      │
│  ┌───────────────────────────────────────────────────────────────────────────────────────────────┐  │
│  │                              MONGODB CLUSTER (Sharded)                                        │  │
│  │  • Event Store (Write Models)     • Read Models (Projected)     • Document Metadata          │  │
│  │  • Shard Key: { tenantId: 1, _id: 1 }                                                        │  │
│  └───────────────────────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                                      │
│  ┌────────────────────────────────────┐  ┌───────────────────────────────────────────────────────┐  │
│  │         REDIS CLUSTER              │  │                 MINIO CLUSTER                         │  │
│  │  • Cache    • Sessions   • Locks   │  │  • Invoices   • Receipts   • Scanned Documents       │  │
│  │  • Rate Limiting                   │  │  • Erasure Coding (EC:4)  • Versioning               │  │
│  └────────────────────────────────────┘  └───────────────────────────────────────────────────────┘  │
│                                                                                                      │
│  ┌───────────────────────────────────────────────────────────────────────────────────────────────┐  │
│  │                              ELASTICSEARCH CLUSTER                                            │  │
│  │  • Full-text Document Search      • Per-tenant Indexes      • Hot/Warm/Cold Tiering          │  │
│  └───────────────────────────────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Core Architecture Patterns

### 3.1 CQRS (Command Query Responsibility Segregation)

The system separates read and write operations into distinct services for optimal performance and scalability.

#### Command Flow (Write Path)

```
┌───────────────┐     ┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│  API Gateway  │────▶│   Command     │────▶│   Command     │────▶│   Aggregate   │
│               │     │   Validator   │     │   Handler     │     │    Root       │
└───────────────┘     └───────────────┘     └───────────────┘     └───────────────┘
                                                                          │
                                                                          ▼
                      ┌───────────────┐     ┌───────────────┐     ┌───────────────┐
                      │    NATS       │◀────│  Event Store  │◀────│    Domain     │
                      │  (Publish)    │     │   (MongoDB)   │     │    Events     │
                      └───────────────┘     └───────────────┘     └───────────────┘
```

#### Query Flow (Read Path)

```
┌───────────────┐     ┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│  API Gateway  │────▶│ Redis Cache   │────▶│  Cache Hit?   │─YES─▶│   Return      │
│               │     │   Lookup      │     │               │     │   Cached      │
└───────────────┘     └───────────────┘     └───────────────┘     └───────────────┘
                                                    │
                                                   NO
                                                    │
                                                    ▼
                                           ┌───────────────┐     ┌───────────────┐
                                           │  Read Model   │────▶│ Cache Result  │
                                           │   (MongoDB)   │     │  + Return     │
                                           └───────────────┘     └───────────────┘
```

#### Event Projection Flow

```
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│    NATS       │────▶│   Event       │────▶│  Projector    │
│  (Subscribe)  │     │   Handler     │     │               │
└───────────────┘     └───────────────┘     └───────────────┘
                                                    │
                                    ┌───────────────┴───────────────┐
                                    ▼                               ▼
                            ┌───────────────┐               ┌───────────────┐
                            │  Read Model   │               │ Invalidate    │
                            │   Update      │               │ Redis Cache   │
                            └───────────────┘               └───────────────┘
```

### 3.2 Event Sourcing

All state changes are captured as immutable events, enabling:
- Complete audit trail
- Event replay for debugging
- Temporal queries ("state at time X")
- Easy integration with other systems

### 3.3 Message Schemas

```go
// Event Envelope - All events follow this structure
type EventEnvelope struct {
    ID            string                 `json:"id"`
    Type          string                 `json:"type"`
    AggregateID   string                 `json:"aggregateId"`
    AggregateType string                 `json:"aggregateType"`
    TenantID      string                 `json:"tenantId"`
    Version       int64                  `json:"version"`
    Timestamp     time.Time              `json:"timestamp"`
    CorrelationID string                 `json:"correlationId"`
    CausationID   string                 `json:"causationId"`
    UserID        string                 `json:"userId"`
    Data          map[string]interface{} `json:"data"`
    Metadata      map[string]string      `json:"metadata"`
}

// Command Envelope - All commands follow this structure
type CommandEnvelope struct {
    ID              string                 `json:"id"`
    Type            string                 `json:"type"`
    TenantID        string                 `json:"tenantId"`
    TargetID        string                 `json:"targetId,omitempty"`
    Timestamp       time.Time              `json:"timestamp"`
    CorrelationID   string                 `json:"correlationId"`
    UserID          string                 `json:"userId"`
    ExpectedVersion int64                  `json:"expectedVersion,omitempty"`
    Data            map[string]interface{} `json:"data"`
    Metadata        map[string]string      `json:"metadata"`
}
```

---

## 4. Module Specifications

### 4.1 Client Module

Manages customer data, credit limits, and billing information.

| Commands | Events | Read Models |
|----------|--------|-------------|
| CreateClient | ClientCreated | ClientSummary |
| UpdateClient | ClientUpdated | ClientDetail |
| DeactivateClient | ClientDeactivated | ClientCreditStatus |
| AssignCreditLimit | CreditLimitAssigned | ClientActivityLog |
| UpdateBillingInfo | BillingInfoUpdated | |
| MergeClients | ClientsMerged | |

**MongoDB Schema (Read Model):**

```javascript
db.clients_read = {
  _id: UUID,
  tenantId: UUID,
  code: String,
  name: String,
  email: String,
  phone: String,
  billingAddress: {
    street: String,
    city: String,
    country: String,
    postalCode: String
  },
  shippingAddresses: [Address],
  creditLimit: Decimal128,
  currentBalance: Decimal128,
  status: String,              // "active", "inactive", "suspended"
  tags: [String],
  customFields: Object,
  createdAt: ISODate,
  updatedAt: ISODate,
  version: Number
}
// Shard Key: { tenantId: 1, _id: 1 }
```

### 4.2 Invoicing Module

Handles invoice lifecycle from creation to payment reconciliation.

| Commands | Events | Read Models |
|----------|--------|-------------|
| CreateInvoice | InvoiceCreated | InvoiceList |
| AddLineItem | LineItemAdded | InvoiceDetail |
| RemoveLineItem | LineItemRemoved | ClientInvoiceSummary |
| ApplyDiscount | DiscountApplied | RevenueReport |
| FinalizeInvoice | InvoiceFinalized | AgingReport |
| VoidInvoice | InvoiceVoided | |
| CreateCreditNote | CreditNoteCreated | |
| RecordPayment | PaymentRecorded | |

**Saga:** `InvoicePaymentSaga` - Coordinates payment processing across modules

**MongoDB Schema (Read Model):**

```javascript
db.invoices_read = {
  _id: UUID,
  tenantId: UUID,
  invoiceNumber: String,
  clientId: UUID,
  clientName: String,          // Denormalized
  status: String,              // "draft", "finalized", "paid", "partial", "voided"
  lineItems: [{
    productId: UUID,
    sku: String,
    description: String,
    quantity: Decimal128,
    unitPrice: Decimal128,
    taxRate: Decimal128,
    discount: Decimal128,
    total: Decimal128
  }],
  subtotal: Decimal128,
  taxTotal: Decimal128,
  discountTotal: Decimal128,
  grandTotal: Decimal128,
  amountPaid: Decimal128,
  amountDue: Decimal128,
  currency: String,
  dueDate: ISODate,
  issuedAt: ISODate,
  paidAt: ISODate,
  payments: [{
    paymentId: UUID,
    amount: Decimal128,
    method: String,
    paidAt: ISODate
  }],
  version: Number
}
```

### 4.3 Payment Module (Pluggable Processors)

Supports multiple payment processors with intelligent routing.

```
┌─────────────────────────────────────────────────────────────────┐
│                       PAYMENT MODULE                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                  Payment Gateway Router                    │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────┐   │  │
│  │  │ Stripe  │  │ PayPal  │  │ Adyen   │  │   Custom    │   │  │
│  │  │Processor│  │Processor│  │Processor│  │  Processor  │   │  │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────────┘   │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Commands:          Events:              Read Models:            │
│  • InitiatePayment  • PaymentInitiated   • PaymentHistory        │
│  • ProcessPayment   • PaymentProcessing  • PaymentDetail         │
│  • CapturePayment   • PaymentSucceeded   • ReconciliationReport  │
│  • RefundPayment    • PaymentFailed                              │
│  • VoidPayment      • PaymentCaptured                            │
│  • ReconcilePayment • PaymentRefunded                            │
│                     • PaymentReconciled                          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Payment Processor Interface (Go):**

```go
// PaymentProcessor defines the interface for pluggable payment processors
type PaymentProcessor interface {
    // Identification
    Name() string
    Version() string
    SupportedCurrencies() []string
    SupportedPaymentMethods() []PaymentMethod
    
    // Core Operations
    Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error)
    Capture(ctx context.Context, req *CaptureRequest) (*CaptureResponse, error)
    Charge(ctx context.Context, req *ChargeRequest) (*ChargeResponse, error)
    Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)
    Void(ctx context.Context, req *VoidRequest) (*VoidResponse, error)
    
    // Webhooks
    ParseWebhook(req *http.Request) (*WebhookEvent, error)
    VerifyWebhookSignature(payload []byte, signature string) error
    
    // Health & Status
    HealthCheck(ctx context.Context) error
    GetTransactionStatus(ctx context.Context, txnID string) (*TransactionStatus, error)
}

// PaymentProcessorRegistry manages processor plugins
type PaymentProcessorRegistry struct {
    processors map[string]PaymentProcessor
    mu         sync.RWMutex
}

func (r *PaymentProcessorRegistry) Register(processor PaymentProcessor) error
func (r *PaymentProcessorRegistry) Get(name string) (PaymentProcessor, error)
func (r *PaymentProcessorRegistry) List() []ProcessorInfo
func (r *PaymentProcessorRegistry) Unregister(name string) error
```

**Routing Configuration (Per Tenant):**

```javascript
db.payment_processor_configs = {
  _id: UUID,
  tenantId: UUID,
  processorName: String,       // "stripe", "paypal", "adyen"
  isEnabled: Boolean,
  isDefault: Boolean,
  priority: Number,            // For fallback routing
  credentials: {
    // Encrypted at rest
    apiKey: String,
    secretKey: String,
    webhookSecret: String,
    merchantId: String
  },
  settings: {
    environment: String,       // "sandbox", "production"
    supportedMethods: [String],
    currencies: [String],
    maxTransactionAmount: Decimal128,
    autoCapture: Boolean
  },
  routingRules: [{
    condition: {
      field: String,           // "amount", "currency", "paymentMethod"
      operator: String,        // "eq", "gt", "lt", "in"
      value: Mixed
    },
    action: String             // "use", "skip", "fallback"
  }]
}
```

### 4.4 Warehouse Module

Manages multi-location warehouse operations.

| Commands | Events | Read Models |
|----------|--------|-------------|
| CreateWarehouse | WarehouseCreated | WarehouseOverview |
| UpdateWarehouse | LocationCreated | LocationInventory |
| CreateLocation | GoodsReceived | PickingQueue |
| ReceiveGoods | ItemsPutAway | ShippingQueue |
| PutAway | ItemsPicked | StockMovementHistory |
| Pick | OrderPacked | |
| Pack | OrderShipped | |
| Ship | StockTransferred | |
| TransferStock | StockAdjusted | |
| AdjustStock | | |

### 4.5 Inventory Module

Handles stock levels, reservations, and reorder management.

| Commands | Events | Read Models |
|----------|--------|-------------|
| CreateProduct | ProductCreated | ProductCatalog |
| UpdateProduct | ProductUpdated | StockLevels |
| SetReorderPoint | StockReserved | ReservationStatus |
| ReserveStock | ReservationReleased | LowStockReport |
| ReleaseReservation | ReservationCommitted | InventoryValuation |
| CommitReservation | InventoryAdjusted | |
| AdjustInventory | LowStockAlert | |
| RecountInventory | StockRecounted | |

**Saga:** `OrderFulfillmentSaga` - Coordinates reservation → pick → ship

**MongoDB Schema (Read Model):**

```javascript
db.inventory_read = {
  _id: UUID,
  tenantId: UUID,
  productId: UUID,
  sku: String,
  name: String,
  warehouses: [{
    warehouseId: UUID,
    warehouseName: String,
    locations: [{
      locationId: UUID,
      locationCode: String,
      quantity: Number,
      reservedQuantity: Number,
      availableQuantity: Number
    }],
    totalQuantity: Number,
    totalReserved: Number,
    totalAvailable: Number
  }],
  globalQuantity: Number,
  globalReserved: Number,
  globalAvailable: Number,
  reorderPoint: Number,
  reorderQuantity: Number,
  unitCost: Decimal128,
  lastReceivedAt: ISODate,
  lastSoldAt: ISODate,
  version: Number
}
// Shard Key: { tenantId: 1, productId: 1 }
```

### 4.6 Customer Management Module (CRM)

Comprehensive customer relationship management including contacts, interactions, segmentation, support tickets, and customer health scoring.

```
┌─────────────────────────────────────────────────────────────────┐
│                 CUSTOMER MANAGEMENT MODULE (CRM)                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    Customer 360° View                      │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────┐   │  │
│  │  │ Profile │  │Contacts │  │ Orders  │  │ Interactions│   │  │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────────┘   │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────┐   │  │
│  │  │Invoices │  │ Tickets │  │Documents│  │Health Score │   │  │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────────┘   │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    Lead Pipeline                           │  │
│  │  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐        │  │
│  │  │ Lead │─▶│Qual. │─▶│Prop. │─▶│Nego. │─▶│ Won  │        │  │
│  │  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘        │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                  Support Ticketing                         │  │
│  │  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐        │  │
│  │  │ New  │─▶│ Open │─▶│Pending│─▶│Resolved│─▶│Closed│      │  │
│  │  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘        │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

#### Commands

| Command | Description |
|---------|-------------|
| **Customer Lifecycle** | |
| CreateCustomer | Create new customer with profile data |
| UpdateCustomer | Update customer information |
| MergeCustomers | Merge duplicate customer records |
| ArchiveCustomer | Archive inactive customer |
| DeleteCustomer | GDPR-compliant customer deletion |
| **Contact Management** | |
| AddContact | Add contact person to customer |
| UpdateContact | Update contact information |
| RemoveContact | Remove contact from customer |
| SetPrimaryContact | Designate primary contact |
| **Lead/Opportunity** | |
| CreateLead | Create new sales lead |
| QualifyLead | Move lead to qualified stage |
| ConvertLeadToCustomer | Convert lead to full customer |
| CreateOpportunity | Create sales opportunity |
| UpdateOpportunityStage | Progress opportunity through pipeline |
| CloseOpportunity | Close as won/lost |
| **Interactions** | |
| LogInteraction | Record customer interaction (call, email, meeting) |
| ScheduleFollowUp | Schedule follow-up activity |
| CompleteActivity | Mark activity as completed |
| **Segmentation** | |
| CreateSegment | Define customer segment with rules |
| UpdateSegment | Modify segment criteria |
| AssignToSegment | Manually assign customer to segment |
| RunSegmentation | Execute segmentation rules |
| **Support Tickets** | |
| CreateTicket | Create support ticket |
| AssignTicket | Assign ticket to agent |
| UpdateTicketStatus | Change ticket status |
| AddTicketComment | Add comment/reply to ticket |
| EscalateTicket | Escalate to higher tier |
| ResolveTicket | Mark ticket as resolved |
| ReopenTicket | Reopen closed ticket |
| **Health Scoring** | |
| RecalculateHealthScore | Trigger health score recalculation |
| SetHealthScoreOverride | Manually override health score |
| ConfigureHealthModel | Update health scoring model |

#### Events

| Event | Triggered When |
|-------|----------------|
| **Customer Events** | |
| CustomerCreated | New customer created |
| CustomerUpdated | Customer profile updated |
| CustomersMerged | Duplicate customers merged |
| CustomerArchived | Customer archived |
| CustomerDeleted | Customer data deleted (GDPR) |
| **Contact Events** | |
| ContactAdded | New contact added |
| ContactUpdated | Contact information changed |
| ContactRemoved | Contact removed |
| PrimaryContactChanged | Primary contact designation changed |
| **Lead Events** | |
| LeadCreated | New lead created |
| LeadQualified | Lead moved to qualified |
| LeadConverted | Lead converted to customer |
| LeadLost | Lead marked as lost |
| **Opportunity Events** | |
| OpportunityCreated | New opportunity created |
| OpportunityStageChanged | Stage progression |
| OpportunityWon | Deal closed won |
| OpportunityLost | Deal closed lost |
| OpportunityValueChanged | Deal value updated |
| **Interaction Events** | |
| InteractionLogged | Customer interaction recorded |
| FollowUpScheduled | Follow-up activity scheduled |
| FollowUpCompleted | Activity completed |
| FollowUpOverdue | Activity past due date |
| **Segment Events** | |
| SegmentCreated | New segment defined |
| SegmentUpdated | Segment rules changed |
| CustomerSegmentChanged | Customer moved between segments |
| SegmentationCompleted | Batch segmentation finished |
| **Ticket Events** | |
| TicketCreated | New ticket opened |
| TicketAssigned | Ticket assigned to agent |
| TicketStatusChanged | Ticket status updated |
| TicketCommentAdded | New comment on ticket |
| TicketEscalated | Ticket escalated |
| TicketResolved | Ticket resolved |
| TicketReopened | Closed ticket reopened |
| TicketSLABreached | SLA threshold exceeded |
| **Health Events** | |
| HealthScoreChanged | Customer health score changed |
| HealthScoreCritical | Health dropped to critical level |
| ChurnRiskDetected | High churn probability detected |

#### Read Models

| Read Model | Purpose |
|------------|---------|
| Customer360View | Complete customer profile with all related data |
| CustomerList | Paginated customer listing with filters |
| CustomerTimeline | Chronological activity timeline |
| ContactDirectory | All contacts across customers |
| LeadPipeline | Lead funnel visualization |
| OpportunityBoard | Kanban-style opportunity view |
| OpportunityForecast | Revenue forecast by stage |
| InteractionHistory | Customer interaction log |
| ActivityCalendar | Scheduled activities view |
| SegmentMembership | Customers per segment |
| SegmentAnalytics | Segment performance metrics |
| TicketQueue | Support ticket queue view |
| TicketDetails | Full ticket with history |
| AgentWorkload | Tickets per agent |
| SLADashboard | SLA compliance metrics |
| HealthScoreReport | Customer health overview |
| ChurnRiskReport | At-risk customers list |
| CustomerLifetimeValue | CLV calculations |

#### Sagas

| Saga | Coordinates |
|------|-------------|
| LeadConversionSaga | Lead → Customer conversion with data migration |
| CustomerMergeSaga | Merge duplicates, consolidate history |
| TicketEscalationSaga | Auto-escalation based on SLA rules |
| ChurnPreventionSaga | Trigger interventions for at-risk customers |
| OnboardingSaga | New customer onboarding workflow |

**MongoDB Schema - Customer (Read Model):**

```javascript
db.customers_read = {
  _id: UUID,
  tenantId: UUID,
  
  // Basic Info
  customerNumber: String,          // Unique identifier (e.g., "CUST-00001")
  type: String,                    // "individual", "business"
  status: String,                  // "lead", "prospect", "active", "inactive", "archived"
  
  // Business Customer
  companyName: String,
  industry: String,
  employeeCount: String,           // "1-10", "11-50", "51-200", "201-500", "500+"
  annualRevenue: Decimal128,
  website: String,
  taxId: String,
  
  // Individual Customer
  firstName: String,
  lastName: String,
  dateOfBirth: ISODate,
  
  // Primary Contact (denormalized)
  primaryContact: {
    contactId: UUID,
    name: String,
    email: String,
    phone: String,
    title: String
  },
  
  // All Contacts
  contacts: [{
    _id: UUID,
    firstName: String,
    lastName: String,
    email: String,
    phone: String,
    mobile: String,
    title: String,
    department: String,
    isPrimary: Boolean,
    isActive: Boolean,
    preferences: {
      preferredChannel: String,    // "email", "phone", "sms"
      doNotContact: Boolean,
      marketingOptIn: Boolean
    },
    createdAt: ISODate
  }],
  
  // Addresses
  addresses: [{
    type: String,                  // "billing", "shipping", "headquarters"
    street: String,
    city: String,
    state: String,
    postalCode: String,
    country: String,
    isPrimary: Boolean
  }],
  
  // Source & Attribution
  source: String,                  // "website", "referral", "trade_show", "cold_call"
  sourceDetail: String,            // Campaign name, referrer name, etc.
  assignedTo: UUID,                // Sales rep / Account manager
  
  // Segmentation
  segments: [String],              // ["enterprise", "high-value", "at-risk"]
  tags: [String],
  customFields: Object,
  
  // Health & Scoring
  healthScore: {
    score: Number,                 // 0-100
    grade: String,                 // "A", "B", "C", "D", "F"
    trend: String,                 // "improving", "stable", "declining"
    factors: [{
      name: String,
      weight: Number,
      score: Number,
      impact: String               // "positive", "neutral", "negative"
    }],
    lastCalculatedAt: ISODate
  },
  
  // Lifetime Value
  lifetimeValue: {
    totalRevenue: Decimal128,
    totalOrders: Number,
    averageOrderValue: Decimal128,
    firstPurchaseDate: ISODate,
    lastPurchaseDate: ISODate,
    predictedNextPurchase: ISODate,
    predictedAnnualValue: Decimal128
  },
  
  // Engagement Metrics
  engagement: {
    lastInteractionDate: ISODate,
    lastInteractionType: String,
    interactionCount30Days: Number,
    emailOpenRate: Number,
    responseTime: Number,          // Average response time in hours
    npsScore: Number,
    lastNPSSurveyDate: ISODate
  },
  
  // Support Metrics
  support: {
    openTickets: Number,
    totalTickets: Number,
    averageResolutionTime: Number, // Hours
    lastTicketDate: ISODate,
    satisfactionScore: Number
  },
  
  // Financial Summary (denormalized from Invoicing)
  financials: {
    creditLimit: Decimal128,
    currentBalance: Decimal128,
    overdueAmount: Decimal128,
    paymentTerms: String,
    preferredPaymentMethod: String,
    lastPaymentDate: ISODate
  },
  
  // Preferences
  preferences: {
    language: String,
    timezone: String,
    currency: String,
    communicationPreferences: {
      email: Boolean,
      phone: Boolean,
      sms: Boolean,
      mail: Boolean
    },
    billingPreferences: {
      invoiceDelivery: String,     // "email", "mail", "portal"
      consolidateInvoices: Boolean
    }
  },
  
  // Timestamps
  convertedFromLeadAt: ISODate,
  firstContactDate: ISODate,
  createdAt: ISODate,
  updatedAt: ISODate,
  archivedAt: ISODate,
  version: Number
}

// Indexes
db.customers_read.createIndex({ tenantId: 1, _id: 1 })                    // Shard key
db.customers_read.createIndex({ tenantId: 1, customerNumber: 1 }, { unique: true })
db.customers_read.createIndex({ tenantId: 1, status: 1, "healthScore.score": -1 })
db.customers_read.createIndex({ tenantId: 1, segments: 1 })
db.customers_read.createIndex({ tenantId: 1, assignedTo: 1, status: 1 })
db.customers_read.createIndex({ tenantId: 1, "contacts.email": 1 })
db.customers_read.createIndex({ tenantId: 1, tags: 1 })
db.customers_read.createIndex({ "lifetimeValue.totalRevenue": -1 })
```

**MongoDB Schema - Lead:**

```javascript
db.leads_read = {
  _id: UUID,
  tenantId: UUID,
  
  // Lead Info
  leadNumber: String,
  status: String,                  // "new", "contacted", "qualified", "unqualified", "converted", "lost"
  stage: String,                   // Pipeline stage
  
  // Contact Info
  firstName: String,
  lastName: String,
  email: String,
  phone: String,
  company: String,
  title: String,
  
  // Source
  source: String,
  sourceDetail: String,
  campaign: String,
  utmParams: {
    source: String,
    medium: String,
    campaign: String,
    content: String,
    term: String
  },
  
  // Qualification
  score: Number,                   // Lead score (0-100)
  rating: String,                  // "hot", "warm", "cold"
  budget: String,
  authority: String,
  need: String,
  timeline: String,
  
  // Assignment
  assignedTo: UUID,
  assignedAt: ISODate,
  
  // Conversion
  convertedToCustomerId: UUID,
  convertedAt: ISODate,
  convertedBy: UUID,
  
  // Lost
  lostReason: String,
  lostAt: ISODate,
  competitor: String,
  
  // Activity
  lastActivityDate: ISODate,
  lastActivityType: String,
  touchCount: Number,
  
  createdAt: ISODate,
  updatedAt: ISODate
}
```

**MongoDB Schema - Opportunity:**

```javascript
db.opportunities_read = {
  _id: UUID,
  tenantId: UUID,
  
  // Basic Info
  opportunityNumber: String,
  name: String,
  customerId: UUID,
  customerName: String,
  
  // Pipeline
  stage: String,                   // "discovery", "qualification", "proposal", "negotiation", "closed_won", "closed_lost"
  probability: Number,             // 0-100%
  
  // Value
  amount: Decimal128,
  currency: String,
  weightedAmount: Decimal128,      // amount * probability
  
  // Timeline
  expectedCloseDate: ISODate,
  actualCloseDate: ISODate,
  daysInStage: Number,
  totalDaysOpen: Number,
  
  // Products/Services
  lineItems: [{
    productId: UUID,
    productName: String,
    quantity: Number,
    unitPrice: Decimal128,
    total: Decimal128
  }],
  
  // Competition
  competitors: [{
    name: String,
    strengths: [String],
    weaknesses: [String],
    status: String                 // "active", "eliminated"
  }],
  
  // Assignment
  ownerId: UUID,
  ownerName: String,
  teamMembers: [UUID],
  
  // Close Info
  closedReason: String,
  lostReason: String,
  wonAgainst: String,
  
  // Source
  sourceLeadId: UUID,
  source: String,
  campaign: String,
  
  createdAt: ISODate,
  updatedAt: ISODate
}
```

**MongoDB Schema - Support Ticket:**

```javascript
db.tickets_read = {
  _id: UUID,
  tenantId: UUID,
  
  // Basic Info
  ticketNumber: String,            // "TKT-00001"
  subject: String,
  description: String,
  
  // Classification
  type: String,                    // "question", "incident", "problem", "feature_request"
  category: String,
  subcategory: String,
  priority: String,                // "low", "medium", "high", "urgent"
  severity: String,                // "minor", "major", "critical"
  
  // Status
  status: String,                  // "new", "open", "pending", "on_hold", "resolved", "closed"
  resolution: String,
  resolutionNotes: String,
  
  // Customer
  customerId: UUID,
  customerName: String,
  contactId: UUID,
  contactName: String,
  contactEmail: String,
  
  // Assignment
  assignedTo: UUID,
  assignedToName: String,
  assignedAt: ISODate,
  team: String,
  tier: Number,                    // Support tier (1, 2, 3)
  
  // SLA
  sla: {
    policy: String,
    responseDeadline: ISODate,
    resolutionDeadline: ISODate,
    responseBreached: Boolean,
    resolutionBreached: Boolean,
    firstResponseAt: ISODate,
    resolvedAt: ISODate
  },
  
  // Related
  relatedTickets: [UUID],
  relatedOrderId: UUID,
  relatedInvoiceId: UUID,
  relatedProductId: UUID,
  
  // Communication
  channel: String,                 // "email", "phone", "chat", "portal", "social"
  comments: [{
    _id: UUID,
    authorId: UUID,
    authorName: String,
    authorType: String,            // "agent", "customer", "system"
    content: String,
    isPublic: Boolean,
    attachments: [{
      fileName: String,
      fileSize: Number,
      mimeType: String,
      objectKey: String
    }],
    createdAt: ISODate
  }],
  
  // History
  statusHistory: [{
    from: String,
    to: String,
    changedBy: UUID,
    changedAt: ISODate,
    reason: String
  }],
  
  // Satisfaction
  satisfaction: {
    rating: Number,                // 1-5
    feedback: String,
    surveyedAt: ISODate
  },
  
  // Metrics
  firstResponseTime: Number,       // Minutes
  resolutionTime: Number,          // Minutes
  reopenCount: Number,
  touchCount: Number,
  
  createdAt: ISODate,
  updatedAt: ISODate,
  closedAt: ISODate
}

// Indexes
db.tickets_read.createIndex({ tenantId: 1, _id: 1 })
db.tickets_read.createIndex({ tenantId: 1, ticketNumber: 1 }, { unique: true })
db.tickets_read.createIndex({ tenantId: 1, status: 1, priority: -1 })
db.tickets_read.createIndex({ tenantId: 1, customerId: 1, createdAt: -1 })
db.tickets_read.createIndex({ tenantId: 1, assignedTo: 1, status: 1 })
db.tickets_read.createIndex({ tenantId: 1, "sla.resolutionDeadline": 1, status: 1 })
```

**MongoDB Schema - Interaction:**

```javascript
db.interactions_read = {
  _id: UUID,
  tenantId: UUID,
  
  // Type
  type: String,                    // "call", "email", "meeting", "note", "chat", "sms"
  direction: String,               // "inbound", "outbound"
  
  // Related
  customerId: UUID,
  customerName: String,
  contactId: UUID,
  contactName: String,
  opportunityId: UUID,
  ticketId: UUID,
  
  // Content
  subject: String,
  description: String,
  outcome: String,                 // "positive", "neutral", "negative"
  
  // Call specific
  callDetails: {
    phoneNumber: String,
    duration: Number,              // Seconds
    recordingUrl: String,
    voicemailLeft: Boolean
  },
  
  // Email specific
  emailDetails: {
    from: String,
    to: [String],
    cc: [String],
    messageId: String,
    threadId: String,
    hasAttachments: Boolean
  },
  
  // Meeting specific
  meetingDetails: {
    location: String,
    meetingUrl: String,
    startTime: ISODate,
    endTime: ISODate,
    attendees: [{
      email: String,
      name: String,
      status: String               // "accepted", "declined", "tentative"
    }]
  },
  
  // Assignment
  ownerId: UUID,
  ownerName: String,
  
  // Follow-up
  followUp: {
    required: Boolean,
    dueDate: ISODate,
    type: String,
    assignedTo: UUID,
    completed: Boolean,
    completedAt: ISODate
  },
  
  // Sentiment (AI-analyzed)
  sentiment: {
    score: Number,                 // -1 to 1
    label: String,                 // "positive", "neutral", "negative"
    keywords: [String]
  },
  
  occurredAt: ISODate,
  createdAt: ISODate,
  createdBy: UUID
}
```

**MongoDB Schema - Segment:**

```javascript
db.segments_read = {
  _id: UUID,
  tenantId: UUID,
  
  name: String,
  description: String,
  type: String,                    // "static", "dynamic"
  
  // Dynamic segment rules
  rules: {
    operator: String,              // "and", "or"
    conditions: [{
      field: String,
      operator: String,            // "eq", "ne", "gt", "lt", "contains", "in", "between"
      value: Mixed,
      dataType: String             // "string", "number", "date", "boolean"
    }]
  },
  
  // Members (for static segments or cached for dynamic)
  memberCount: Number,
  members: [UUID],                 // Only for static segments
  
  // Automation
  automations: [{
    trigger: String,               // "on_enter", "on_exit"
    action: String,                // "send_email", "assign_tag", "notify_owner"
    config: Object
  }],
  
  // Metadata
  isActive: Boolean,
  lastRefreshedAt: ISODate,
  refreshFrequency: String,        // "realtime", "hourly", "daily"
  
  createdBy: UUID,
  createdAt: ISODate,
  updatedAt: ISODate
}
```

**Health Score Calculation Model:**

```go
// Health Score Configuration
type HealthScoreModel struct {
    TenantID    string              `bson:"tenantId"`
    Name        string              `bson:"name"`
    IsDefault   bool                `bson:"isDefault"`
    Factors     []HealthFactor      `bson:"factors"`
    Thresholds  HealthThresholds    `bson:"thresholds"`
    UpdatedAt   time.Time           `bson:"updatedAt"`
}

type HealthFactor struct {
    Name        string              `bson:"name"`
    Weight      float64             `bson:"weight"`       // Sum of all weights = 1.0
    DataSource  string              `bson:"dataSource"`   // "engagement", "financial", "support", "usage"
    Metric      string              `bson:"metric"`       // Specific metric name
    Scoring     []ScoringRule       `bson:"scoring"`
}

type ScoringRule struct {
    Condition   string              `bson:"condition"`    // "gt", "lt", "between", "eq"
    Value       interface{}         `bson:"value"`
    Score       int                 `bson:"score"`        // 0-100
}

type HealthThresholds struct {
    Excellent   int                 `bson:"excellent"`    // >= 80 = A
    Good        int                 `bson:"good"`         // >= 60 = B
    Fair        int                 `bson:"fair"`         // >= 40 = C
    Poor        int                 `bson:"poor"`         // >= 20 = D
    // < 20 = F
}

// Default health factors
var DefaultHealthFactors = []HealthFactor{
    {
        Name:       "Payment Behavior",
        Weight:     0.25,
        DataSource: "financial",
        Metric:     "payment_timeliness",
        Scoring: []ScoringRule{
            {Condition: "eq", Value: "always_on_time", Score: 100},
            {Condition: "eq", Value: "usually_on_time", Score: 75},
            {Condition: "eq", Value: "sometimes_late", Score: 50},
            {Condition: "eq", Value: "often_late", Score: 25},
        },
    },
    {
        Name:       "Engagement Frequency",
        Weight:     0.20,
        DataSource: "engagement",
        Metric:     "interactions_30d",
        Scoring: []ScoringRule{
            {Condition: "gt", Value: 10, Score: 100},
            {Condition: "between", Value: []int{5, 10}, Score: 75},
            {Condition: "between", Value: []int{1, 4}, Score: 50},
            {Condition: "eq", Value: 0, Score: 25},
        },
    },
    {
        Name:       "Support Satisfaction",
        Weight:     0.20,
        DataSource: "support",
        Metric:     "csat_score",
        Scoring: []ScoringRule{
            {Condition: "gt", Value: 4.5, Score: 100},
            {Condition: "between", Value: []float64{4.0, 4.5}, Score: 75},
            {Condition: "between", Value: []float64{3.0, 4.0}, Score: 50},
            {Condition: "lt", Value: 3.0, Score: 25},
        },
    },
    {
        Name:       "Revenue Trend",
        Weight:     0.20,
        DataSource: "financial",
        Metric:     "revenue_growth_qoq",
        Scoring: []ScoringRule{
            {Condition: "gt", Value: 0.10, Score: 100},      // >10% growth
            {Condition: "between", Value: []float64{0, 0.10}, Score: 75},
            {Condition: "between", Value: []float64{-0.10, 0}, Score: 50},
            {Condition: "lt", Value: -0.10, Score: 25},      // >10% decline
        },
    },
    {
        Name:       "Product Adoption",
        Weight:     0.15,
        DataSource: "usage",
        Metric:     "feature_adoption_rate",
        Scoring: []ScoringRule{
            {Condition: "gt", Value: 0.80, Score: 100},
            {Condition: "between", Value: []float64{0.50, 0.80}, Score: 75},
            {Condition: "between", Value: []float64{0.25, 0.50}, Score: 50},
            {Condition: "lt", Value: 0.25, Score: 25},
        },
    },
}
```

**Customer Management API Endpoints:**

```
# Customers
POST   /api/v1/customers                         # Create customer
GET    /api/v1/customers                         # List customers
GET    /api/v1/customers/{id}                    # Get customer
GET    /api/v1/customers/{id}/360                # Get Customer 360 view
PATCH  /api/v1/customers/{id}                    # Update customer
DELETE /api/v1/customers/{id}                    # Archive/delete customer
POST   /api/v1/customers/merge                   # Merge customers
POST   /api/v1/customers/{id}/gdpr-delete        # GDPR deletion

# Contacts
POST   /api/v1/customers/{id}/contacts           # Add contact
GET    /api/v1/customers/{id}/contacts           # List contacts
PATCH  /api/v1/customers/{id}/contacts/{cid}     # Update contact
DELETE /api/v1/customers/{id}/contacts/{cid}     # Remove contact
POST   /api/v1/customers/{id}/contacts/{cid}/primary  # Set primary

# Leads
POST   /api/v1/leads                             # Create lead
GET    /api/v1/leads                             # List leads (pipeline view)
GET    /api/v1/leads/{id}                        # Get lead
PATCH  /api/v1/leads/{id}                        # Update lead
POST   /api/v1/leads/{id}/qualify                # Qualify lead
POST   /api/v1/leads/{id}/convert                # Convert to customer
POST   /api/v1/leads/{id}/lose                   # Mark as lost

# Opportunities
POST   /api/v1/opportunities                     # Create opportunity
GET    /api/v1/opportunities                     # List opportunities
GET    /api/v1/opportunities/{id}                # Get opportunity
PATCH  /api/v1/opportunities/{id}                # Update opportunity
POST   /api/v1/opportunities/{id}/stage          # Change stage
POST   /api/v1/opportunities/{id}/close          # Close won/lost
GET    /api/v1/opportunities/forecast            # Revenue forecast

# Interactions
POST   /api/v1/interactions                      # Log interaction
GET    /api/v1/interactions                      # List interactions
GET    /api/v1/customers/{id}/interactions       # Customer interactions
GET    /api/v1/customers/{id}/timeline           # Activity timeline

# Segments
POST   /api/v1/segments                          # Create segment
GET    /api/v1/segments                          # List segments
GET    /api/v1/segments/{id}                     # Get segment
PATCH  /api/v1/segments/{id}                     # Update segment
DELETE /api/v1/segments/{id}                     # Delete segment
POST   /api/v1/segments/{id}/refresh             # Refresh members
GET    /api/v1/segments/{id}/members             # Get members

# Tickets
POST   /api/v1/tickets                           # Create ticket
GET    /api/v1/tickets                           # List tickets
GET    /api/v1/tickets/{id}                      # Get ticket
PATCH  /api/v1/tickets/{id}                      # Update ticket
POST   /api/v1/tickets/{id}/assign               # Assign ticket
POST   /api/v1/tickets/{id}/escalate             # Escalate ticket
POST   /api/v1/tickets/{id}/resolve              # Resolve ticket
POST   /api/v1/tickets/{id}/comments             # Add comment
GET    /api/v1/customers/{id}/tickets            # Customer tickets

# Health Scores
GET    /api/v1/customers/{id}/health             # Get health score
POST   /api/v1/customers/{id}/health/recalculate # Recalculate
GET    /api/v1/health/at-risk                    # At-risk customers
GET    /api/v1/health/dashboard                  # Health dashboard
```

**NATS Subjects for Customer Management:**

```javascript
// Commands
"cmd.customer.create"
"cmd.customer.update"
"cmd.customer.merge"
"cmd.customer.archive"
"cmd.contact.add"
"cmd.contact.update"
"cmd.lead.create"
"cmd.lead.qualify"
"cmd.lead.convert"
"cmd.opportunity.create"
"cmd.opportunity.stage"
"cmd.opportunity.close"
"cmd.interaction.log"
"cmd.segment.create"
"cmd.segment.refresh"
"cmd.ticket.create"
"cmd.ticket.assign"
"cmd.ticket.resolve"
"cmd.health.recalculate"

// Events
"evt.customer.created"
"evt.customer.updated"
"evt.customer.merged"
"evt.customer.archived"
"evt.contact.added"
"evt.lead.created"
"evt.lead.qualified"
"evt.lead.converted"
"evt.opportunity.created"
"evt.opportunity.stage_changed"
"evt.opportunity.won"
"evt.opportunity.lost"
"evt.interaction.logged"
"evt.segment.membership_changed"
"evt.ticket.created"
"evt.ticket.resolved"
"evt.ticket.sla_breached"
"evt.health.score_changed"
"evt.health.churn_risk_detected"
```

---

## 5. Messaging Infrastructure (NATS)

### 5.1 Stream Definitions

```javascript
const streams = {
  // Command streams (work queues)
  COMMANDS: {
    name: "COMMANDS",
    subjects: [
      "cmd.client.*",
      "cmd.customer.*",
      "cmd.contact.*",
      "cmd.lead.*",
      "cmd.opportunity.*",
      "cmd.interaction.*",
      "cmd.segment.*",
      "cmd.ticket.*",
      "cmd.health.*",
      "cmd.invoice.*",
      "cmd.payment.*",
      "cmd.warehouse.*",
      "cmd.inventory.*",
      "cmd.plugin.*"
    ],
    retention: "workqueue",
    maxAge: 24 * 60 * 60 * 1e9,  // 24 hours
    storage: "file",
    replicas: 3,
    maxMsgsPerSubject: 100000,
    discard: "old"
  },
  
  // Event streams (log/replay)
  EVENTS: {
    name: "EVENTS",
    subjects: [
      "evt.client.*",
      "evt.customer.*",
      "evt.contact.*",
      "evt.lead.*",
      "evt.opportunity.*",
      "evt.interaction.*",
      "evt.segment.*",
      "evt.ticket.*",
      "evt.health.*",
      "evt.invoice.*",
      "evt.payment.*",
      "evt.warehouse.*",
      "evt.inventory.*",
      "evt.plugin.*"
    ],
    retention: "limits",
    maxAge: 365 * 24 * 60 * 60 * 1e9,  // 1 year
    storage: "file",
    replicas: 3,
    maxBytes: 1024 * 1024 * 1024 * 100,  // 100GB
    discard: "old"
  },
  
  // Query streams (request/reply)
  QUERIES: {
    name: "QUERIES",
    subjects: ["qry.*.*"],
    retention: "workqueue",
    maxAge: 60 * 1e9,  // 1 minute
    storage: "memory",
    replicas: 3
  },
  
  // Document events
  DOCUMENT_EVENTS: {
    name: "DOCUMENT_EVENTS",
    subjects: [
      "doc.uploaded.>",
      "doc.processed.>",
      "doc.deleted.>",
      "doc.indexed.>",
      "doc.failed.>"
    ],
    retention: "limits",
    maxAge: 30 * 24 * 60 * 60 * 1e9,  // 30 days
    storage: "file",
    replicas: 3,
    maxBytes: 1024 * 1024 * 1024 * 10  // 10GB
  },
  
  // Dead Letter Queue
  DLQ: {
    name: "DLQ",
    subjects: ["dlq.*"],
    retention: "limits",
    maxAge: 30 * 24 * 60 * 60 * 1e9,  // 30 days
    storage: "file",
    replicas: 3
  }
}
```

### 5.2 Consumer Groups

```javascript
const consumers = {
  // Module command handlers
  CLIENT_COMMAND_HANDLER: {
    stream: "COMMANDS",
    durable: "client-cmd-handler",
    filterSubject: "cmd.client.*",
    ackPolicy: "explicit",
    maxDeliver: 5,
    maxAckPending: 1000
  },
  
  // CRM command handlers
  CUSTOMER_COMMAND_HANDLER: {
    stream: "COMMANDS",
    durable: "customer-cmd-handler",
    filterSubject: "cmd.customer.*",
    ackPolicy: "explicit",
    maxDeliver: 5,
    maxAckPending: 1000
  },
  
  LEAD_COMMAND_HANDLER: {
    stream: "COMMANDS",
    durable: "lead-cmd-handler",
    filterSubject: "cmd.lead.*",
    ackPolicy: "explicit",
    maxDeliver: 5,
    maxAckPending: 500
  },
  
  OPPORTUNITY_COMMAND_HANDLER: {
    stream: "COMMANDS",
    durable: "opportunity-cmd-handler",
    filterSubject: "cmd.opportunity.*",
    ackPolicy: "explicit",
    maxDeliver: 5,
    maxAckPending: 500
  },
  
  TICKET_COMMAND_HANDLER: {
    stream: "COMMANDS",
    durable: "ticket-cmd-handler",
    filterSubject: "cmd.ticket.*",
    ackPolicy: "explicit",
    maxDeliver: 5,
    maxAckPending: 1000
  },
  
  HEALTH_SCORE_CALCULATOR: {
    stream: "EVENTS",
    durable: "health-score-calc",
    filterSubject: "evt.>",  // Listens to all events for health calculation
    ackPolicy: "explicit",
    maxDeliver: 3,
    maxAckPending: 500
  },
  
  // Event projectors
  CLIENT_EVENT_PROJECTOR: {
    stream: "EVENTS",
    durable: "client-evt-projector",
    filterSubject: "evt.client.*",
    ackPolicy: "explicit",
    maxDeliver: 10
  },
  
  CUSTOMER_EVENT_PROJECTOR: {
    stream: "EVENTS",
    durable: "customer-evt-projector",
    filterSubject: "evt.customer.*",
    ackPolicy: "explicit",
    maxDeliver: 10
  },
  
  TICKET_SLA_MONITOR: {
    stream: "EVENTS",
    durable: "ticket-sla-monitor",
    filterSubject: "evt.ticket.*",
    ackPolicy: "explicit",
    maxDeliver: 5
  },
  
  // Document processors
  DOCUMENT_PROCESSOR: {
    stream: "DOCUMENT_EVENTS",
    durable: "doc-processor",
    filterSubject: "doc.uploaded.>",
    ackPolicy: "explicit",
    maxDeliver: 5,
    maxAckPending: 100,
    ackWait: 10 * 60 * 1e9  // 10 minutes
  },
  
  ELASTICSEARCH_INDEXER: {
    stream: "DOCUMENT_EVENTS",
    durable: "es-indexer",
    filterSubject: "doc.processed.>",
    ackPolicy: "explicit",
    maxDeliver: 5,
    maxAckPending: 200
  }
}
```

### 5.3 Multi-Region Supercluster

```
┌─────────────────────────────────────────────────────────────────┐
│                    NATS SUPERCLUSTER                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Region: US-EAST                    Region: EU-WEST             │
│  ┌────────────────────────┐        ┌────────────────────────┐   │
│  │  NATS Cluster          │◄──────►│  NATS Cluster          │   │
│  │  ┌────┐ ┌────┐ ┌────┐  │ Gateway│  ┌────┐ ┌────┐ ┌────┐  │   │
│  │  │ N1 │ │ N2 │ │ N3 │  │        │  │ N1 │ │ N2 │ │ N3 │  │   │
│  │  └────┘ └────┘ └────┘  │        │  └────┘ └────┘ └────┘  │   │
│  │  JetStream R3          │        │  JetStream R3          │   │
│  └────────────────────────┘        └────────────────────────┘   │
│           │                                   │                  │
│           └───────────────┬───────────────────┘                  │
│                           │                                      │
│                  Region: APAC                                    │
│                  ┌────────────────────────┐                      │
│                  │  NATS Cluster          │                      │
│                  │  ┌────┐ ┌────┐ ┌────┐  │                      │
│                  │  │ N1 │ │ N2 │ │ N3 │  │                      │
│                  │  └────┘ └────┘ └────┘  │                      │
│                  │  JetStream R3          │                      │
│                  └────────────────────────┘                      │
│                                                                  │
│  Stream Mirroring:                                               │
│  • EVENTS stream → mirrored across all regions                  │
│  • COMMANDS stream → regional (routed by tenant)                │
│  • QUERIES stream → regional only                               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. Data Layer

### 6.1 MongoDB Sharding Strategy

```
┌─────────────────────────────────────────────────────────────────┐
│                  MONGODB SHARDING TOPOLOGY                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Config Servers (Replica Set)                                    │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                          │
│  │Config 1 │  │Config 2 │  │Config 3 │                          │
│  └─────────┘  └─────────┘  └─────────┘                          │
│                                                                  │
│  Mongos Routers                                                  │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐             │
│  │Mongos 1 │  │Mongos 2 │  │Mongos 3 │  │Mongos N │             │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘             │
│                                                                  │
│  Shard Clusters                                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ Shard 1 (RS)        Shard 2 (RS)        Shard N (RS)        ││
│  │ ┌───┐┌───┐┌───┐    ┌───┐┌───┐┌───┐    ┌───┐┌───┐┌───┐      ││
│  │ │P  ││S  ││S  │    │P  ││S  ││S  │    │P  ││S  ││S  │      ││
│  │ └───┘└───┘└───┘    └───┘└───┘└───┘    └───┘└───┘└───┘      ││
│  │ Tenants: A-F        Tenants: G-M        Tenants: N-Z        ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                  │
│  Shard Key: { tenantId: 1, _id: 1 }                             │
│  • Ensures tenant isolation                                      │
│  • Enables tenant-level migrations                               │
│  • Supports zone sharding for data residency                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 6.2 Event Store Schema

```javascript
// Generic event store collection (per module)
db.{module}_events = {
  _id: ObjectId,
  aggregateId: UUID,
  aggregateType: String,
  eventType: String,
  eventData: Object,
  metadata: {
    tenantId: UUID,
    userId: UUID,
    correlationId: UUID,
    causationId: UUID,
    timestamp: ISODate,
    version: Number
  }
}

// Indexes
db.{module}_events.createIndex({ aggregateId: 1, version: 1 })
db.{module}_events.createIndex({ "metadata.tenantId": 1, eventType: 1, "metadata.timestamp": -1 })
```

---

## 7. File Storage & Document Processing

### 7.1 Document Processing Pipeline

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                      DOCUMENT PROCESSING PIPELINE                                │
└─────────────────────────────────────────────────────────────────────────────────┘

  User Upload                                                      Search Query
       │                                                                 │
       ▼                                                                 ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  API Gateway │────▶│   Document   │────▶│    MinIO     │     │  Search API  │
│              │     │   Service    │     │   Cluster    │     │              │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                            │                    │                      │
                            │                    │ Bucket               │
                            │                    │ Notification         │
                            ▼                    ▼                      ▼
                     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
                     │   MongoDB    │     │    NATS      │     │Elasticsearch │
                     │  (Metadata)  │     │  JetStream   │     │   Cluster    │
                     └──────────────┘     └──────────────┘     └──────────────┘
                                                │                      ▲
                                                │                      │
                                                ▼                      │
                                         ┌──────────────┐              │
                                         │  Document    │              │
                                         │  Processor   │──────────────┘
                                         │   Service    │
                                         └──────────────┘
                                                │
                                    ┌───────────┴───────────┐
                                    ▼                       ▼
                             ┌──────────────┐       ┌──────────────┐
                             │     OCR      │       │    Text      │
                             │  (Tesseract) │       │  Extraction  │
                             └──────────────┘       └──────────────┘
```

### 7.2 MinIO Cluster Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         MINIO DISTRIBUTED CLUSTER                                │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────────┐│
│  │                         Load Balancer (HAProxy/Nginx)                       ││
│  └─────────────────────────────────────────────────────────────────────────────┘│
│                                        │                                         │
│           ┌────────────────────────────┼────────────────────────────┐           │
│           ▼                            ▼                            ▼           │
│  ┌─────────────────┐          ┌─────────────────┐          ┌─────────────────┐  │
│  │   MinIO Node 1  │          │   MinIO Node 2  │          │   MinIO Node 3  │  │
│  │  ┌───────────┐  │          │  ┌───────────┐  │          │  ┌───────────┐  │  │
│  │  │  Disk 1-4 │  │          │  │  Disk 1-4 │  │          │  │  Disk 1-4 │  │  │
│  │  └───────────┘  │          │  └───────────┘  │          │  └───────────┘  │  │
│  └─────────────────┘          └─────────────────┘          └─────────────────┘  │
│                                                                                  │
│  ┌─────────────────┐          ┌─────────────────┐                               │
│  │   MinIO Node 4  │          │   MinIO Node N  │    (Minimum 4 nodes for EC)   │
│  └─────────────────┘          └─────────────────┘                               │
│                                                                                  │
│  • Erasure Coding: EC:4 (4 data + 4 parity) - survives 4 drive/node failures   │
│  • Encryption: Server-side encryption with Vault integration                    │
│  • Versioning: Enabled for audit trail                                          │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 7.3 Bucket Structure

```
{tenant-id}-documents/
├── invoices/
│   └── {year}/{month}/
├── purchase-orders/
│   └── {year}/{month}/
├── receipts/
│   └── {year}/{month}/
├── contracts/
│   └── {client-id}/
└── scanned/
    └── {year}/{month}/{day}/

{tenant-id}-processed/
├── text/
│   └── {document-id}.txt
├── thumbnails/
│   └── {document-id}.jpg
└── metadata/
    └── {document-id}.json

{tenant-id}-temp/
└── {upload-session-id}/
```

**Lifecycle Policies:**
- `temp bucket`: Delete after 24 hours
- `processed/thumbnails`: Move to IA tier after 30 days
- `documents`: Retain based on compliance rules (7 years default)

### 7.4 Document Service (Go Implementation)

```go
type Document struct {
    ID              string            `bson:"_id" json:"id"`
    TenantID        string            `bson:"tenantId" json:"tenantId"`
    Type            DocumentType      `bson:"type" json:"type"`
    FileName        string            `bson:"fileName" json:"fileName"`
    MimeType        string            `bson:"mimeType" json:"mimeType"`
    Size            int64             `bson:"size" json:"size"`
    Checksum        string            `bson:"checksum" json:"checksum"`
    
    // Storage
    Bucket          string            `bson:"bucket" json:"bucket"`
    ObjectKey       string            `bson:"objectKey" json:"objectKey"`
    VersionID       string            `bson:"versionId,omitempty"`
    
    // Processing
    ProcessingStatus ProcessingStatus `bson:"processingStatus"`
    ExtractedText    string           `bson:"extractedText,omitempty"`
    ThumbnailKey     string           `bson:"thumbnailKey,omitempty"`
    ExtractedMetadata struct {
        PageCount     int       `bson:"pageCount,omitempty"`
        InvoiceNumber string    `bson:"invoiceNumber,omitempty"`
        InvoiceDate   time.Time `bson:"invoiceDate,omitempty"`
        TotalAmount   float64   `bson:"totalAmount,omitempty"`
        VendorName    string    `bson:"vendorName,omitempty"`
        Dates         []string  `bson:"dates,omitempty"`
        Amounts       []string  `bson:"amounts,omitempty"`
        Emails        []string  `bson:"emails,omitempty"`
    } `bson:"extractedMetadata"`
    
    // Audit
    UploadedBy      string    `bson:"uploadedBy"`
    CreatedAt       time.Time `bson:"createdAt"`
    UpdatedAt       time.Time `bson:"updatedAt"`
}

type DocumentType string

const (
    DocTypeInvoice       DocumentType = "invoice"
    DocTypePurchaseOrder DocumentType = "purchase_order"
    DocTypeReceipt       DocumentType = "receipt"
    DocTypeContract      DocumentType = "contract"
    DocTypeScanned       DocumentType = "scanned"
)
```

### 7.5 Elasticsearch Index Mapping

```json
{
  "settings": {
    "number_of_shards": 5,
    "number_of_replicas": 1,
    "analysis": {
      "analyzer": {
        "document_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "asciifolding", "snowball"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "tenant_id": { "type": "keyword" },
      "type": { "type": "keyword" },
      "file_name": { "type": "text", "analyzer": "document_analyzer" },
      "content": { 
        "type": "text",
        "analyzer": "document_analyzer",
        "term_vector": "with_positions_offsets"
      },
      "invoice_number": { "type": "keyword" },
      "invoice_date": { "type": "date" },
      "total_amount": { "type": "float" },
      "vendor_name": { "type": "text" },
      "tags": { "type": "keyword" },
      "created_at": { "type": "date" }
    }
  }
}
```

### 7.6 Document API Endpoints

```
POST   /api/v1/documents                    # Upload document
POST   /api/v1/documents/multipart          # Multipart upload
GET    /api/v1/documents/{id}               # Get document metadata
GET    /api/v1/documents/{id}/download      # Download document
GET    /api/v1/documents/{id}/thumbnail     # Get thumbnail
GET    /api/v1/documents/{id}/presigned-url # Get presigned URL
PATCH  /api/v1/documents/{id}               # Update metadata
DELETE /api/v1/documents/{id}               # Delete document
PUT    /api/v1/documents/{id}/tags          # Update tags
POST   /api/v1/documents/search             # Full-text search
GET    /api/v1/documents/search/suggest     # Search suggestions
POST   /api/v1/documents/{id}/reprocess     # Reprocess document
```

---

## 8. Plugin System

### 8.1 Plugin Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      PLUGIN SYSTEM                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                   Plugin Registry                          │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐│  │
│  │  │ Discovery   │  │ Lifecycle   │  │   Health Monitor    ││  │
│  │  │ Service     │  │ Manager     │  │                     ││  │
│  │  └─────────────┘  └─────────────┘  └─────────────────────┘│  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                   Plugin Types                             │  │
│  │  ┌─────────────────┐  ┌─────────────────┐                 │  │
│  │  │  Event Handlers │  │ Command Handlers│                 │  │
│  │  │  (React to      │  │ (Extend core    │                 │  │
│  │  │   core events)  │  │  operations)    │                 │  │
│  │  └─────────────────┘  └─────────────────┘                 │  │
│  │  ┌─────────────────┐  ┌─────────────────┐                 │  │
│  │  │  API Extensions │  │  Scheduled Jobs │                 │  │
│  │  │  (New endpoints)│  │  (Cron tasks)   │                 │  │
│  │  └─────────────────┘  └─────────────────┘                 │  │
│  │  ┌─────────────────┐  ┌─────────────────┐                 │  │
│  │  │  Report         │  │  Integration    │                 │  │
│  │  │  Generators     │  │  Connectors     │                 │  │
│  │  └─────────────────┘  └─────────────────┘                 │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 8.2 Plugin Interface

```go
// Plugin manifest (plugin.yaml)
type PluginManifest struct {
    Name               string            `yaml:"name"`
    Version            string            `yaml:"version"`
    Description        string            `yaml:"description"`
    Author             string            `yaml:"author"`
    EntryPoint         string            `yaml:"entryPoint"`
    Permissions        []Permission      `yaml:"permissions"`
    EventSubscriptions []string          `yaml:"eventSubscriptions"`
    CommandExtensions  []CommandExt      `yaml:"commandExtensions"`
    APIRoutes          []APIRoute        `yaml:"apiRoutes"`
    ScheduledTasks     []ScheduledTask   `yaml:"scheduledTasks"`
    Dependencies       []Dependency      `yaml:"dependencies"`
}

// Plugin runtime interface
type Plugin interface {
    Initialize(ctx context.Context, config PluginConfig) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    HealthCheck(ctx context.Context) HealthStatus
    HandleEvent(ctx context.Context, event EventEnvelope) error
    HandleCommand(ctx context.Context, cmd CommandEnvelope) (interface{}, error)
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Plugin SDK - provided to plugins
type PluginSDK interface {
    PublishEvent(ctx context.Context, event EventEnvelope) error
    PublishCommand(ctx context.Context, cmd CommandEnvelope) error
    RequestReply(ctx context.Context, subject string, data interface{}) (interface{}, error)
    GetCollection(name string) *mongo.Collection
    GetCache() redis.Cmdable
    Logger() *slog.Logger
    Metrics() MetricsCollector
    GetConfig(key string) interface{}
    GetSecret(key string) (string, error)
}
```

### 8.3 Plugin Communication

```
NATS Subject Patterns for Plugins:

• cmd.plugin.<plugin-name>.<action>  → Commands TO plugin
• evt.plugin.<plugin-name>.<event>   → Events FROM plugin
• qry.plugin.<plugin-name>.<query>   → Queries TO plugin
• rpc.plugin.<plugin-name>.<method>  → RPC calls TO plugin
```

---

## 9. Caching Strategy

### 9.1 Redis Key Patterns

```
Pattern: {tenant}:{module}:{entity}:{id}:{view}

Examples:
t:abc123:client:detail:uuid-xxx          → Client detail
t:abc123:client:list:page:1:size:50      → Client list page
t:abc123:invoice:summary:uuid-xxx        → Invoice summary
t:abc123:inventory:stock:sku:ABC123      → Stock level by SKU

Global (no tenant prefix):
g:product:catalog:uuid-xxx               → Product info
g:exchange:rate:USD:EUR                  → Exchange rates

Sessions & Locks:
sess:{session-id}                        → User session
lock:invoice:uuid-xxx                    → Distributed lock
ratelimit:api:tenant:abc123              → Rate limit counter
```

### 9.2 Caching Patterns by Data Type

| Pattern | Use Case | Examples |
|---------|----------|----------|
| **Write-Through** | Strong consistency | Payment configs, Sessions, Feature flags |
| **Cache-Aside** | Read-heavy | Client details, Products, Invoice summaries |
| **Write-Behind** | Eventual consistency OK | Analytics, Activity logs, Report aggregations |
| **No Cache** | Always fresh | Payments, Reservations, Order state |

### 9.3 Cache Invalidation

Event-driven invalidation rules automatically clear cache when data changes:

```go
var invalidationRules = map[string][]string{
    "ClientCreated":    {"t:{tenant}:client:list:*"},
    "ClientUpdated":    {"t:{tenant}:client:detail:{id}", "t:{tenant}:client:list:*"},
    "InvoiceFinalized": {"t:{tenant}:invoice:detail:{id}", "t:{tenant}:invoice:list:*"},
    "PaymentSucceeded": {"t:{tenant}:invoice:detail:{invoiceId}"},
    "StockReserved":    {"t:{tenant}:inventory:stock:*:{productId}"},
}
```

---

## 10. Security Architecture

### 10.1 Authentication & Authorization

```
┌───────────────────────────────────────────────────────────────┐
│                    SECURITY LAYERS                             │
├───────────────────────────────────────────────────────────────┤
│                                                                │
│  API Gateway                                                   │
│  • JWT validation                                              │
│  • Rate limiting (per tenant, per user)                        │
│  • IP allowlisting                                             │
│  • Request signing verification                                │
│                                                                │
│  Auth Service                                                  │
│  • OAuth 2.0 / OIDC provider                                   │
│  • SAML integration (enterprise SSO)                           │
│  • MFA enforcement                                             │
│  • API key management                                          │
│  • Service-to-service auth (mTLS)                              │
│                                                                │
│  RBAC Service                                                  │
│  • Role definitions per module                                 │
│  • Permission inheritance                                      │
│  • Attribute-based access control (ABAC)                       │
│  • Row-level security policies                                 │
│  • Audit logging                                               │
│                                                                │
│  Data Security                                                 │
│  • Encryption at rest (MongoDB, Redis, MinIO)                  │
│  • Encryption in transit (TLS 1.3)                             │
│  • Field-level encryption (PII, payment data)                  │
│  • Key rotation (AWS KMS / HashiCorp Vault)                    │
│  • Data masking in logs                                        │
│                                                                │
└───────────────────────────────────────────────────────────────┘
```

---

## 11. Scaling & Multi-Tenancy

### 11.1 Multi-Tenancy Tiers

| Tier | Name | Infrastructure | Cost |
|------|------|----------------|------|
| **1** | Shared (SMB) | Shared MongoDB/Redis/NATS, namespace isolation | $ |
| **2** | Dedicated Data (Mid-Market) | Dedicated MongoDB/Redis, shared compute | $$ |
| **3** | Dedicated Compute (Enterprise) | Dedicated clusters, node affinity | $$$ |
| **4** | Isolated (Large Enterprise) | Separate account/project, on-premise option | $$$$ |

### 11.2 Horizontal Scaling

```
┌─────────────────────────────────────────────────────────────────┐
│                    SCALING TOPOLOGY                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Load Balancer Layer                                             │
│  • HAProxy / AWS ALB / Kubernetes Ingress                        │
│                                                                  │
│  API Gateway Layer (Stateless)                                   │
│  • HPA: min=3, max=50, target CPU=70%                            │
│                                                                  │
│  Service Layer (Per Module)                                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Client Service     │ HPA: 3-30 pods  │ CPU target: 60%   │   │
│  │ Customer/CRM Svc   │ HPA: 5-50 pods  │ CPU target: 65%   │   │
│  │ Lead/Opp Service   │ HPA: 3-20 pods  │ CPU target: 60%   │   │
│  │ Ticket Service     │ HPA: 5-40 pods  │ CPU target: 65%   │   │
│  │ Invoice Service    │ HPA: 3-50 pods  │ CPU target: 70%   │   │
│  │ Payment Service    │ HPA: 5-100 pods │ CPU target: 50%   │   │
│  │ Warehouse Service  │ HPA: 3-30 pods  │ CPU target: 65%   │   │
│  │ Inventory Service  │ HPA: 5-80 pods  │ CPU target: 60%   │   │
│  │ Document Processor │ HPA: 10-50 pods │ CPU target: 70%   │   │
│  │ Health Score Svc   │ HPA: 3-15 pods  │ CPU target: 70%   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 12. Deployment & Operations

### 12.1 Kubernetes Deployment Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: client-command-service
  namespace: erp-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: client-command-service
  template:
    metadata:
      labels:
        app: client-command-service
        module: client
        tier: command
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchLabels:
                  app: client-command-service
              topologyKey: kubernetes.io/hostname
      containers:
      - name: client-command-service
        image: erp/client-command:v1.2.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc
        - containerPort: 9091
          name: metrics
        env:
        - name: NATS_URL
          valueFrom:
            configMapKeyRef:
              name: nats-config
              key: url
        - name: MONGODB_URI
          valueFrom:
            secretKeyRef:
              name: mongodb-secrets
              key: uri
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 15
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: client-command-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: client-command-service
  minReplicas: 3
  maxReplicas: 30
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 60
```

### 12.2 Service Mesh (Istio)

Features enabled:
- mTLS between all services
- Traffic routing (canary, blue-green)
- Request retries and timeouts
- Circuit breaking
- Distributed tracing

---

## 13. Observability

### 13.1 Observability Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Metrics** | Prometheus + Grafana | Latency, throughput, error rates |
| **Logs** | Loki / ELK | Structured JSON logs, correlation IDs |
| **Traces** | Jaeger / Tempo | End-to-end request tracing |
| **Alerting** | PagerDuty / OpsGenie | SLO-based alerts, on-call rotation |

### 13.2 Key Metrics

- Service latency (p50, p95, p99)
- Request rates by module
- Error rates by endpoint
- NATS message lag
- MongoDB operation latency
- Redis cache hit ratio
- MinIO storage utilization
- Elasticsearch indexing rate
- Business metrics (invoices/day, payments/hour)

---

## 14. Disaster Recovery

### 14.1 Recovery Objectives

| Component | RTO | RPO | Strategy |
|-----------|-----|-----|----------|
| MongoDB | < 1 min | 0 | 3-node RS, cross-AZ |
| Redis | < 30 sec | < 1 sec | Cluster + Sentinel |
| NATS | < 30 sec | 0 | JetStream R3 |
| MinIO | < 2 min | 0 | Erasure coding + replication |
| Elasticsearch | < 5 min | < 1 min | Cross-cluster replication |
| Services | < 2 min | N/A | K8s self-healing |

### 14.2 Backup Schedule

- **MongoDB**: Continuous oplog streaming + daily snapshots
- **Event Store**: Real-time replication + hourly S3 archival
- **Redis**: AOF persistence + hourly RDB snapshots
- **MinIO**: Cross-region replication
- **Configuration**: GitOps (ArgoCD) - version controlled

---

## 15. Capacity Planning

### Scale Tier: 1M Active Clients, 10K Concurrent Users

| Component | Specification | Cost (Monthly) |
|-----------|--------------|----------------|
| **API Gateway** | 10-20 instances, 2 cores, 4GB each | |
| **Module Services** | 5-30 instances each, 1-2 cores, 1-2GB | |
| **Document Processors** | 10-50 instances, 4 cores, 8GB (CPU for OCR) | |
| **MongoDB** | 5-10 shards, 3-node RS, 16 cores, 64GB, NVMe | $10-20K |
| **Redis Cluster** | 6 nodes, 8 cores, 32GB each | $3-5K |
| **MinIO Cluster** | 4-8 nodes, 4+ NVMe SSDs, 10TB+ per node | $5-15K |
| **Elasticsearch** | 3 master + 5-10 hot + 3-5 warm, 16 cores, 64GB | $8-15K |
| **NATS Cluster** | 5-7 nodes per region, 8 cores, 16GB | $2-4K |
| **Compute** | Kubernetes nodes | $20-35K |
| **Network** | Data transfer, load balancers | $3-8K |
| **Total** | | **$49K - $98K** |

---

## 16. Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| **API Gateway** | Kong / AWS API Gateway | Routing, rate limiting, auth |
| **Service Framework** | Go + gRPC | High-performance microservices |
| **Message Broker** | NATS JetStream | Event streaming, commands, queries |
| **Primary Database** | MongoDB (Sharded) | Event store, read models |
| **Object Storage** | MinIO (Distributed) | Documents, invoices, files |
| **Cache** | Redis Cluster | Caching, sessions, distributed locks |
| **Search** | Elasticsearch | Full-text document search |
| **OCR Engine** | Tesseract | Image/PDF text extraction |
| **Container Orchestration** | Kubernetes | Deployment, scaling, self-healing |
| **Service Mesh** | Istio | mTLS, traffic management |
| **CI/CD** | ArgoCD + GitHub Actions | GitOps deployment |
| **Monitoring** | Prometheus + Grafana | Metrics, dashboards |
| **Logging** | Loki / ELK | Centralized logging |
| **Tracing** | Jaeger / Tempo | Distributed tracing |
| **Secrets** | HashiCorp Vault | Secrets management, encryption |

---

## Implementation Roadmap

### Phase 1: Foundation (Months 1-3)
- Core infrastructure (K8s, MongoDB, Redis, NATS)
- Auth/RBAC service
- Client module (full CQRS)
- Basic API gateway

### Phase 2: Core Modules (Months 4-6)
- Invoicing module
- Payment module (Stripe/PayPal)
- Basic reporting

### Phase 3: Customer Management (Months 7-9)
- Customer 360° view
- Contact management
- Lead & opportunity pipeline
- Basic interactions logging

### Phase 4: Warehouse & Inventory (Months 10-12)
- Warehouse module
- Inventory module
- Cross-module sagas

### Phase 5: Document Management (Months 13-14)
- MinIO setup
- Document service
- OCR pipeline
- Elasticsearch integration

### Phase 6: CRM Advanced Features (Months 15-16)
- Support ticketing system
- SLA management
- Health scoring engine
- Customer segmentation
- Churn prediction

### Phase 7: Plugin System (Months 17-18)
- Plugin framework
- Plugin SDK
- Example plugins

### Phase 8: Scale & Optimize (Months 19-21)
- Performance optimization
- Multi-region deployment
- Advanced analytics
- Enterprise features

---

**Document Version:** 3.0  
**Architecture Status:** Production Ready  
**Last Updated:** January 2026

---

*This document is maintained by the Platform Engineering team.*
