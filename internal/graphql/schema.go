package graphql

// Schema contains GraphQL type definitions for the ERP system
const Schema = `
# Scalar types
scalar Time
scalar UUID
scalar Decimal
scalar JSON

# Enums
enum ClientStatus {
  ACTIVE
  INACTIVE
  SUSPENDED
  MERGED
}

enum InvoiceStatus {
  DRAFT
  ISSUED
  PAID
  OVERDUE
  VOID
}

enum PaymentStatus {
  PENDING
  COMPLETED
  FAILED
  REFUNDED
}

enum WarehouseStatus {
  ACTIVE
  INACTIVE
  PENDING
}

enum InventoryStatus {
  AVAILABLE
  RESERVED
  ALLOCATED
  IN_TRANSIT
  QUARANTINE
  DAMAGED
  EXPIRED
}

enum DocumentType {
  INVOICE
  PURCHASE_ORDER
  RECEIPT
  CONTRACT
  SCANNED
  OTHER
}

# Interface for common fields
interface Node {
  id: UUID!
  createdAt: Time!
  updatedAt: Time!
}

# Client types
type Client implements Node {
  id: UUID!
  tenantId: UUID!
  name: String!
  email: String!
  phone: String
  status: ClientStatus!
  creditLimit: Decimal
  currentBalance: Decimal
  availableCredit: Decimal
  tags: [String!]
  billingAddress: Address
  shippingAddresses: [Address!]
  customFields: JSON
  createdAt: Time!
  updatedAt: Time!
}

type Address {
  street: String!
  city: String!
  state: String
  postalCode: String!
  country: String!
}

# Invoice types
type Invoice implements Node {
  id: UUID!
  tenantId: UUID!
  clientId: UUID!
  client: Client
  invoiceNumber: String!
  status: InvoiceStatus!
  issueDate: Time!
  dueDate: Time!
  subtotal: Decimal!
  taxTotal: Decimal!
  total: Decimal!
  amountPaid: Decimal!
  amountDue: Decimal!
  lineItems: [InvoiceLineItem!]!
  notes: String
  createdAt: Time!
  updatedAt: Time!
}

type InvoiceLineItem {
  id: UUID!
  productId: UUID
  description: String!
  quantity: Int!
  unitPrice: Decimal!
  total: Decimal!
}

# Payment types
type Payment implements Node {
  id: UUID!
  tenantId: UUID!
  invoiceId: UUID!
  invoice: Invoice
  amount: Decimal!
  currency: String!
  status: PaymentStatus!
  method: String!
  reference: String
  processedAt: Time
  createdAt: Time!
}

# Warehouse types
type Warehouse implements Node {
  id: UUID!
  tenantId: UUID!
  name: String!
  code: String!
  type: String!
  status: WarehouseStatus!
  address: Address
  capacity: Int
  isActive: Boolean!
  locations: [Location!]
  createdAt: Time!
  updatedAt: Time!
}

type Location {
  id: UUID!
  name: String!
  code: String!
  zone: String!
  aisle: String
  rack: String
  bin: String
  capacity: Int
  currentStock: Int
  isActive: Boolean!
}

# Inventory types
type InventoryItem implements Node {
  id: UUID!
  tenantId: UUID!
  productId: UUID!
  warehouseId: UUID!
  warehouse: Warehouse
  sku: String!
  quantity: Int!
  reservedQty: Int!
  availableQty: Int!
  status: InventoryStatus!
  reorderPoint: Int
  reorderQuantity: Int
  unitCost: Decimal
  totalValue: Decimal
  createdAt: Time!
  updatedAt: Time!
}

# Product types
type Product implements Node {
  id: UUID!
  tenantId: UUID!
  sku: String!
  name: String!
  description: String
  type: String!
  category: String!
  status: String!
  pricing: ProductPricing!
  inventory: ProductInventory!
  createdAt: Time!
  updatedAt: Time!
}

type ProductPricing {
  listPrice: Decimal!
  salePrice: Decimal
  costPrice: Decimal
  currency: String!
}

type ProductInventory {
  trackInventory: Boolean!
  quantityOnHand: Int
  quantityReserved: Int
  quantityAvailable: Int
  reorderPoint: Int
}

# Document types
type Document implements Node {
  id: UUID!
  tenantId: UUID!
  type: DocumentType!
  fileName: String!
  mimeType: String!
  size: Int!
  processingStatus: String!
  pageCount: Int
  tags: [String!]
  metadata: DocumentMetadata
  uploadedBy: UUID!
  createdAt: Time!
  updatedAt: Time!
}

type DocumentMetadata {
  invoiceNumber: String
  invoiceDate: Time
  totalAmount: Decimal
  vendorName: String
}

# User types
type User implements Node {
  id: UUID!
  tenantId: UUID!
  email: String!
  firstName: String!
  lastName: String!
  status: String!
  roles: [String!]!
  mfaEnabled: Boolean!
  lastLoginAt: Time
  createdAt: Time!
  updatedAt: Time!
}

# Pagination types
type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
  totalCount: Int!
}

type ClientConnection {
  edges: [ClientEdge!]!
  pageInfo: PageInfo!
}

type ClientEdge {
  node: Client!
  cursor: String!
}

type InvoiceConnection {
  edges: [InvoiceEdge!]!
  pageInfo: PageInfo!
}

type InvoiceEdge {
  node: Invoice!
  cursor: String!
}

type WarehouseConnection {
  edges: [WarehouseEdge!]!
  pageInfo: PageInfo!
}

type WarehouseEdge {
  node: Warehouse!
  cursor: String!
}

type InventoryConnection {
  edges: [InventoryEdge!]!
  pageInfo: PageInfo!
}

type InventoryEdge {
  node: InventoryItem!
  cursor: String!
}

type ProductConnection {
  edges: [ProductEdge!]!
  pageInfo: PageInfo!
}

type ProductEdge {
  node: Product!
  cursor: String!
}

type DocumentConnection {
  edges: [DocumentEdge!]!
  pageInfo: PageInfo!
}

type DocumentEdge {
  node: Document!
  cursor: String!
}

# Input types for mutations
input CreateClientInput {
  name: String!
  email: String!
  phone: String
  creditLimit: Decimal
  billingAddress: AddressInput
  tags: [String!]
}

input AddressInput {
  street: String!
  city: String!
  state: String
  postalCode: String!
  country: String!
}

input UpdateClientInput {
  name: String
  email: String
  phone: String
  status: ClientStatus
  creditLimit: Decimal
  tags: [String!]
}

input CreateInvoiceInput {
  clientId: UUID!
  issueDate: Time
  dueDate: Time
  lineItems: [InvoiceLineItemInput!]!
  notes: String
}

input InvoiceLineItemInput {
  productId: UUID
  description: String!
  quantity: Int!
  unitPrice: Decimal!
}

input CreatePaymentInput {
  invoiceId: UUID!
  amount: Decimal!
  method: String!
  reference: String
}

input CreateWarehouseInput {
  name: String!
  code: String!
  type: String!
  address: AddressInput
  capacity: Int
}

input CreateProductInput {
  sku: String!
  name: String!
  description: String
  type: String!
  category: String!
  listPrice: Decimal!
  costPrice: Decimal
}

# Search and filter input types
input ClientFilter {
  status: ClientStatus
  search: String
  tags: [String!]
  minCreditLimit: Decimal
  maxCreditLimit: Decimal
}

input InvoiceFilter {
  status: InvoiceStatus
  clientId: UUID
  dateFrom: Time
  dateTo: Time
  minAmount: Decimal
  maxAmount: Decimal
}

input InventoryFilter {
  warehouseId: UUID
  productId: UUID
  status: InventoryStatus
  lowStock: Boolean
}

# Query root type
type Query {
  # Client queries
  client(id: UUID!): Client
  clients(
    filter: ClientFilter
    first: Int
    after: String
    last: Int
    before: String
  ): ClientConnection!
  
  # Invoice queries
  invoice(id: UUID!): Invoice
  invoices(
    filter: InvoiceFilter
    first: Int
    after: String
    last: Int
    before: String
  ): InvoiceConnection!
  
  # Payment queries
  payment(id: UUID!): Payment
  paymentsByInvoice(invoiceId: UUID!): [Payment!]!
  
  # Warehouse queries
  warehouse(id: UUID!): Warehouse
  warehouses(
    status: WarehouseStatus
    first: Int
    after: String
  ): WarehouseConnection!
  
  # Inventory queries
  inventoryItem(id: UUID!): InventoryItem
  inventory(
    filter: InventoryFilter
    first: Int
    after: String
  ): InventoryConnection!
  
  # Product queries
  product(id: UUID!): Product
  products(
    category: String
    status: String
    first: Int
    after: String
  ): ProductConnection!
  
  # Document queries
  document(id: UUID!): Document
  documents(
    type: DocumentType
    status: String
    first: Int
    after: String
  ): DocumentConnection!
  
  # Search
  search(
    query: String!
    types: [String!]
    first: Int
    after: String
  ): SearchConnection!
}

type SearchConnection {
  edges: [SearchEdge!]!
  pageInfo: PageInfo!
}

type SearchEdge {
  node: SearchResult!
  cursor: String!
}

union SearchResult = Client | Invoice | Product | Document

# Mutation root type
type Mutation {
  # Client mutations
  createClient(input: CreateClientInput!): Client!
  updateClient(id: UUID!, input: UpdateClientInput!): Client!
  deleteClient(id: UUID!): Boolean!
  activateClient(id: UUID!): Client!
  deactivateClient(id: UUID!): Client!
  
  # Invoice mutations
  createInvoice(input: CreateInvoiceInput!): Invoice!
  issueInvoice(id: UUID!): Invoice!
  payInvoice(id: UUID!, input: CreatePaymentInput!): Invoice!
  voidInvoice(id: UUID!): Invoice!
  
  # Payment mutations
  createPayment(input: CreatePaymentInput!): Payment!
  refundPayment(id: UUID!, reason: String): Payment!
  
  # Warehouse mutations
  createWarehouse(input: CreateWarehouseInput!): Warehouse!
  updateWarehouse(id: UUID!, name: String, address: AddressInput): Warehouse!
  activateWarehouse(id: UUID!): Warehouse!
  deactivateWarehouse(id: UUID!): Warehouse!
  
  # Inventory mutations
  receiveInventory(
    productId: UUID!
    warehouseId: UUID!
    quantity: Int!
    unitCost: Decimal
  ): InventoryItem!
  
  shipInventory(
    productId: UUID!
    warehouseId: UUID!
    quantity: Int!
  ): InventoryItem!
  
  adjustInventory(
    productId: UUID!
    warehouseId: UUID!
    quantity: Int!
    reason: String!
  ): InventoryItem!
  
  # Product mutations
  createProduct(input: CreateProductInput!): Product!
  updateProduct(id: UUID!, name: String, description: String, listPrice: Decimal): Product!
  activateProduct(id: UUID!): Product!
  deactivateProduct(id: UUID!): Product!
  
  # Document mutations
  uploadDocument(file: Upload!, type: DocumentType!, tags: [String!]): Document!
  deleteDocument(id: UUID!): Boolean!
  reprocessDocument(id: UUID!): Document!
}

# Upload scalar for file uploads
scalar Upload

# Subscription root type for real-time updates
type Subscription {
  # Client events
  clientCreated: Client!
  clientUpdated(clientId: UUID): Client!
  
  # Invoice events
  invoiceCreated: Invoice!
  invoiceUpdated(invoiceId: UUID): Invoice!
  invoicePaid(invoiceId: UUID): Invoice!
  
  # Payment events
  paymentReceived: Payment!
  paymentFailed: Payment!
  
  # Inventory events
  inventoryUpdated(productId: UUID, warehouseId: UUID): InventoryItem!
  lowStockAlert(threshold: Int): [InventoryItem!]!
  
  # Document events
  documentUploaded: Document!
  documentProcessed(documentId: UUID): Document!
}

# Error types
type Error {
  message: String!
  code: String!
  field: String
}

# Response wrapper with errors
type MutationResponse {
  success: Boolean!
  data: JSON
  errors: [Error!]
}
`
