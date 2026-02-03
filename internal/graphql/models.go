package graphql

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// Scalar types

// Client represents a customer in the system
type Client struct {
	ID            string
	TenantID      string
	Name          string
	Email         string
	Phone         string
	Mobile        string
	Status        ClientStatus
	Type          string
	TaxID         string
	CreditLimit   decimal.Decimal
	CurrentCredit decimal.Decimal
	Address       *Address
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Metadata      map[string]string
}

// Invoice represents a billing document
type Invoice struct {
	ID            string
	TenantID      string
	ClientID      string
	Client        *Client
	InvoiceNumber string
	Status        InvoiceStatus
	IssueDate     time.Time
	DueDate       time.Time
	Subtotal      decimal.Decimal
	TaxTotal      decimal.Decimal
	DiscountTotal decimal.Decimal
	Total         decimal.Decimal
	BalanceDue    decimal.Decimal
	Currency      string
	LineItems     []*InvoiceLineItem
	Payments      []*Payment
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Metadata      map[string]string
}

// InvoiceLineItem represents a single line on an invoice
type InvoiceLineItem struct {
	ID          string
	ProductID   string
	Product     *Product
	Description string
	Quantity    decimal.Decimal
	UnitPrice   decimal.Decimal
	Discount    decimal.Decimal
	TaxRate     decimal.Decimal
	TaxAmount   decimal.Decimal
	LineTotal   decimal.Decimal
}

// Payment represents a payment made on an invoice
type Payment struct {
	ID          string
	TenantID    string
	InvoiceID   string
	Invoice     *Invoice
	Amount      decimal.Decimal
	Currency    string
	Method      string
	Status      PaymentStatus
	Reference   string
	Notes       string
	ProcessedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Metadata    map[string]string
}

// Warehouse represents a storage location
type Warehouse struct {
	ID        string
	TenantID  string
	Name      string
	Code      string
	Type      string
	Status    WarehouseStatus
	IsActive  bool
	Address   *Address
	Locations []*WarehouseLocation
	ManagerID string
	CreatedAt time.Time
	UpdatedAt time.Time
	Metadata  map[string]string
}

// WarehouseLocation represents a specific location within a warehouse
type WarehouseLocation struct {
	ID          string
	WarehouseID string
	Name        string
	Code        string
	Zone        string
	Row         string
	Section     string
	Shelf       string
	Bin         string
	IsActive    bool
	IsDefault   bool
	Capacity    int
}

// InventoryItem represents stock quantity for a product in a warehouse
type InventoryItem struct {
	ID           string
	TenantID     string
	ProductID    string
	Product      *Product
	WarehouseID  string
	Warehouse    *Warehouse
	LocationID   string
	Location     *WarehouseLocation
	SKU          string
	Quantity     int
	ReservedQty  int
	AvailableQty int
	MinStock     int
	MaxStock     int
	ReorderPoint int
	Status       InventoryStatus
	UnitCost     decimal.Decimal
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Metadata     map[string]string
}

// Product represents a catalog item
type Product struct {
	ID            string
	TenantID      string
	SKU           string
	Name          string
	Description   string
	Category      string
	Type          string
	Status        string
	UnitOfMeasure string
	ListPrice     decimal.Decimal
	CostPrice     decimal.Decimal
	Barcode       string
	Weight        decimal.Decimal
	Dimensions    *ProductDimensions
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Metadata      map[string]string
}

// ProductDimensions represents physical dimensions
type ProductDimensions struct {
	Length decimal.Decimal
	Width  decimal.Decimal
	Height decimal.Decimal
	Weight decimal.Decimal
}

// Document represents an uploaded file
type Document struct {
	ID               string
	TenantID         string
	Type             DocumentType
	FileName         string
	OriginalName     string
	MimeType         string
	Size             int64
	Hash             string
	StoragePath      string
	ProcessingStatus string
	ExtractedText    string
	PageCount        int
	Tags             []string
	RelatedTo        *RelatedEntity
	UploadedBy       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Metadata         map[string]string
}

// RelatedEntity links a document to another entity
type RelatedEntity struct {
	Type       string
	EntityID   string
	EntityType string
}

// Address represents a physical address
type Address struct {
	Street1    string
	Street2    string
	City       string
	State      string
	PostalCode string
	Country    string
	Latitude   decimal.Decimal
	Longitude  decimal.Decimal
}

// Enum types

// ClientStatus represents the status of a client
type ClientStatus string

const (
	ClientStatusActive    ClientStatus = "active"
	ClientStatusInactive  ClientStatus = "inactive"
	ClientStatusSuspended ClientStatus = "suspended"
	ClientStatusPending   ClientStatus = "pending"
)

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft   InvoiceStatus = "draft"
	InvoiceStatusSent    InvoiceStatus = "sent"
	InvoiceStatusPartial InvoiceStatus = "partial"
	InvoiceStatusPaid    InvoiceStatus = "paid"
	InvoiceStatusOverdue InvoiceStatus = "overdue"
	InvoiceStatusVoid    InvoiceStatus = "void"
	InvoiceStatusPending InvoiceStatus = "pending"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending         PaymentStatus = "pending"
	PaymentStatusCompleted       PaymentStatus = "completed"
	PaymentStatusFailed          PaymentStatus = "failed"
	PaymentStatusRefunded        PaymentStatus = "refunded"
	PaymentStatusPartialRefunded PaymentStatus = "partial_refunded"
)

// WarehouseStatus represents the status of a warehouse
type WarehouseStatus string

const (
	WarehouseStatusActive      WarehouseStatus = "active"
	WarehouseStatusInactive    WarehouseStatus = "inactive"
	WarehouseStatusFull        WarehouseStatus = "full"
	WarehouseStatusMaintenance WarehouseStatus = "maintenance"
)

// InventoryStatus represents the status of inventory
type InventoryStatus string

const (
	InventoryStatusAvailable  InventoryStatus = "available"
	InventoryStatusLow        InventoryStatus = "low"
	InventoryStatusOutOfStock InventoryStatus = "out_of_stock"
	InventoryReserved         InventoryStatus = "reserved"
)

// DocumentType represents the type of document
type DocumentType string

const (
	DocumentTypeInvoice   DocumentType = "invoice"
	DocumentTypeReceipt   DocumentType = "receipt"
	DocumentTypeContract  DocumentType = "contract"
	DocumentTypeReport    DocumentType = "report"
	DocumentTypeID        DocumentType = "id"
	DocumentTypeOther     DocumentType = "other"
	DocumentTypeTaxForm   DocumentType = "tax_form"
	DocumentTypeStatement DocumentType = "statement"
)

// Input types

// CreateClientInput represents input for creating a client
type CreateClientInput struct {
	Name        string
	Email       string
	Phone       *string
	Mobile      *string
	TaxID       *string
	CreditLimit *decimal.Decimal
	Address     *AddressInput
	Metadata    map[string]string
}

// UpdateClientInput represents input for updating a client
type UpdateClientInput struct {
	Name        *string
	Email       *string
	Phone       *string
	Mobile      *string
	TaxID       *string
	CreditLimit *decimal.Decimal
	Address     *AddressInput
	Metadata    map[string]string
}

// CreateInvoiceInput represents input for creating an invoice
type CreateInvoiceInput struct {
	ClientID  string
	DueDate   time.Time
	Currency  *string
	LineItems []InvoiceLineItemInput
	Notes     *string
	Metadata  map[string]string
}

// InvoiceLineItemInput represents a line item input
type InvoiceLineItemInput struct {
	ProductID   string
	Description string
	Quantity    decimal.Decimal
	UnitPrice   decimal.Decimal
	Discount    *decimal.Decimal
	TaxRate     *decimal.Decimal
}

// CreatePaymentInput represents input for creating a payment
type CreatePaymentInput struct {
	InvoiceID string
	Amount    decimal.Decimal
	Method    string
	Reference *string
	Notes     *string
}

// CreateWarehouseInput represents input for creating a warehouse
type CreateWarehouseInput struct {
	Name      string
	Code      string
	Type      string
	Address   *AddressInput
	ManagerID *string
}

// CreateProductInput represents input for creating a product
type CreateProductInput struct {
	SKU           string
	Name          string
	Description   *string
	Category      string
	Type          string
	UnitOfMeasure string
	ListPrice     decimal.Decimal
	CostPrice     decimal.Decimal
	Barcode       *string
	Weight        *decimal.Decimal
	Dimensions    *DimensionsInput
}

// DimensionsInput represents physical dimensions input
type DimensionsInput struct {
	Length decimal.Decimal
	Width  decimal.Decimal
	Height decimal.Decimal
	Weight decimal.Decimal
}

// AddressInput represents address input
type AddressInput struct {
	Street1    string
	Street2    *string
	City       string
	State      string
	PostalCode string
	Country    string
}

// Filter types

// ClientFilter represents filter options for clients
type ClientFilter struct {
	Search   *string
	Status   *string
	Type     *string
	DateFrom *time.Time
	DateTo   *time.Time
}

// InvoiceFilter represents filter options for invoices
type InvoiceFilter struct {
	ClientID  *string
	Status    *string
	DateFrom  *time.Time
	DateTo    *time.Time
	MinAmount *decimal.Decimal
	MaxAmount *decimal.Decimal
}

// InventoryFilter represents filter options for inventory
type InventoryFilter struct {
	WarehouseID *string
	ProductID   *string
	Status      *string
	LowStock    *bool
}

// Connection types

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage bool
	HasPrevPage bool
	StartCursor string
	EndCursor   string
	TotalCount  int
}

// ClientConnection represents a paginated list of clients
type ClientConnection struct {
	Edges    []*ClientEdge
	PageInfo *PageInfo
}

// ClientEdge represents a single edge in a client connection
type ClientEdge struct {
	Node   *Client
	Cursor string
}

// InvoiceConnection represents a paginated list of invoices
type InvoiceConnection struct {
	Edges    []*InvoiceEdge
	PageInfo *PageInfo
}

// InvoiceEdge represents a single edge in an invoice connection
type InvoiceEdge struct {
	Node   *Invoice
	Cursor string
}

// PaymentConnection represents a paginated list of payments
type PaymentConnection struct {
	Edges    []*PaymentEdge
	PageInfo *PageInfo
}

// PaymentEdge represents a single edge in a payment connection
type PaymentEdge struct {
	Node   *Payment
	Cursor string
}

// WarehouseConnection represents a paginated list of warehouses
type WarehouseConnection struct {
	Edges    []*WarehouseEdge
	PageInfo *PageInfo
}

// WarehouseEdge represents a single edge in a warehouse connection
type WarehouseEdge struct {
	Node   *Warehouse
	Cursor string
}

// InventoryConnection represents a paginated list of inventory items
type InventoryConnection struct {
	Edges    []*InventoryEdge
	PageInfo *PageInfo
}

// InventoryEdge represents a single edge in an inventory connection
type InventoryEdge struct {
	Node   *InventoryItem
	Cursor string
}

// ProductConnection represents a paginated list of products
type ProductConnection struct {
	Edges    []*ProductEdge
	PageInfo *PageInfo
}

// ProductEdge represents a single edge in a product connection
type ProductEdge struct {
	Node   *Product
	Cursor string
}

// DocumentConnection represents a paginated list of documents
type DocumentConnection struct {
	Edges    []*DocumentEdge
	PageInfo *PageInfo
}

// DocumentEdge represents a single edge in a document connection
type DocumentEdge struct {
	Node   *Document
	Cursor string
}

// SearchConnection represents a paginated list of search results
type SearchConnection struct {
	Edges    []*SearchEdge
	PageInfo *PageInfo
}

// SearchEdge represents a single edge in a search connection
type SearchEdge struct {
	Node   SearchResult
	Cursor string
}

// SearchResult is an interface for search results
type SearchResult interface {
	GetID() string
	GetTenantID() string
	GetSearchType() string
}

// Ensure scalar types implement SearchResult
func (c *Client) GetID() string         { return c.ID }
func (c *Client) GetTenantID() string   { return c.TenantID }
func (c *Client) GetSearchType() string { return "client" }

func (i *Invoice) GetID() string         { return i.ID }
func (i *Invoice) GetTenantID() string   { return i.TenantID }
func (i *Invoice) GetSearchType() string { return "invoice" }

func (p *Payment) GetID() string         { return p.ID }
func (p *Payment) GetTenantID() string   { return p.TenantID }
func (p *Payment) GetSearchType() string { return "payment" }

func (w *Warehouse) GetID() string         { return w.ID }
func (w *Warehouse) GetTenantID() string   { return w.TenantID }
func (w *Warehouse) GetSearchType() string { return "warehouse" }

func (i *InventoryItem) GetID() string         { return i.ID }
func (i *InventoryItem) GetTenantID() string   { return i.TenantID }
func (i *InventoryItem) GetSearchType() string { return "inventory" }

func (p *Product) GetID() string         { return p.ID }
func (p *Product) GetTenantID() string   { return p.TenantID }
func (p *Product) GetSearchType() string { return "product" }

func (d *Document) GetID() string         { return d.ID }
func (d *Document) GetTenantID() string   { return d.TenantID }
func (d *Document) GetSearchType() string { return "document" }

// Resolver interfaces

// QueryResolver defines the Query root type
type QueryResolver interface {
	// Client queries
	Client(ctx context.Context, id string) (*Client, error)
	Clients(ctx context.Context, filter *ClientFilter, first *int, after *string, last *int, before *string) (*ClientConnection, error)

	// Invoice queries
	Invoice(ctx context.Context, id string) (*Invoice, error)
	Invoices(ctx context.Context, filter *InvoiceFilter, first *int, after *string, last *int, before *string) (*InvoiceConnection, error)

	// Payment queries
	Payment(ctx context.Context, id string) (*Payment, error)
	PaymentsByInvoice(ctx context.Context, invoiceID string) ([]*Payment, error)

	// Warehouse queries
	Warehouse(ctx context.Context, id string) (*Warehouse, error)
	Warehouses(ctx context.Context, status *WarehouseStatus, first *int, after *string) (*WarehouseConnection, error)

	// Inventory queries
	InventoryItem(ctx context.Context, id string) (*InventoryItem, error)
	Inventory(ctx context.Context, filter *InventoryFilter, first *int, after *string) (*InventoryConnection, error)

	// Product queries
	Product(ctx context.Context, id string) (*Product, error)
	Products(ctx context.Context, category *string, status *string, first *int, after *string) (*ProductConnection, error)

	// Document queries
	Document(ctx context.Context, id string) (*Document, error)
	Documents(ctx context.Context, docType *DocumentType, status *string, first *int, after *string) (*DocumentConnection, error)

	// Search
	Search(ctx context.Context, query string, types []string, first *int, after *string) (*SearchConnection, error)
}

// MutationResolver defines the Mutation root type
type MutationResolver interface {
	// Client mutations
	CreateClient(ctx context.Context, input CreateClientInput) (*Client, error)
	UpdateClient(ctx context.Context, id string, input UpdateClientInput) (*Client, error)
	DeleteClient(ctx context.Context, id string) (bool, error)
	ActivateClient(ctx context.Context, id string) (*Client, error)
	DeactivateClient(ctx context.Context, id string) (*Client, error)

	// Invoice mutations
	CreateInvoice(ctx context.Context, input CreateInvoiceInput) (*Invoice, error)
	IssueInvoice(ctx context.Context, id string) (*Invoice, error)
	PayInvoice(ctx context.Context, id string, input CreatePaymentInput) (*Invoice, error)
	VoidInvoice(ctx context.Context, id string) (*Invoice, error)

	// Payment mutations
	CreatePayment(ctx context.Context, input CreatePaymentInput) (*Payment, error)
	RefundPayment(ctx context.Context, id string, reason *string) (*Payment, error)

	// Warehouse mutations
	CreateWarehouse(ctx context.Context, input CreateWarehouseInput) (*Warehouse, error)
	UpdateWarehouse(ctx context.Context, id string, name *string, address *AddressInput) (*Warehouse, error)
	ActivateWarehouse(ctx context.Context, id string) (*Warehouse, error)
	DeactivateWarehouse(ctx context.Context, id string) (*Warehouse, error)

	// Inventory mutations
	ReceiveInventory(ctx context.Context, productID string, warehouseID string, quantity int, unitCost *decimal.Decimal) (*InventoryItem, error)
	ShipInventory(ctx context.Context, productID string, warehouseID string, quantity int) (*InventoryItem, error)
	AdjustInventory(ctx context.Context, productID string, warehouseID string, quantity int, reason string) (*InventoryItem, error)

	// Product mutations
	CreateProduct(ctx context.Context, input CreateProductInput) (*Product, error)
	UpdateProduct(ctx context.Context, id string, name *string, description *string, listPrice *decimal.Decimal) (*Product, error)
	ActivateProduct(ctx context.Context, id string) (*Product, error)
	DeactivateProduct(ctx context.Context, id string) (*Product, error)

	// Document mutations
	UploadDocument(ctx context.Context, file string, docType DocumentType, tags []string) (*Document, error)
	DeleteDocument(ctx context.Context, id string) (bool, error)
	ReprocessDocument(ctx context.Context, id string) (*Document, error)
}

// SubscriptionResolver defines the Subscription root type
type SubscriptionResolver interface {
	// Client subscriptions
	ClientCreated(ctx context.Context) (<-chan *Client, error)
	ClientUpdated(ctx context.Context, clientID *string) (<-chan *Client, error)

	// Invoice subscriptions
	InvoiceCreated(ctx context.Context) (<-chan *Invoice, error)
	InvoiceUpdated(ctx context.Context, invoiceID *string) (<-chan *Invoice, error)
	InvoicePaid(ctx context.Context, invoiceID *string) (<-chan *Invoice, error)

	// Payment subscriptions
	PaymentReceived(ctx context.Context) (<-chan *Payment, error)
	PaymentFailed(ctx context.Context) (<-chan *Payment, error)

	// Inventory subscriptions
	InventoryUpdated(ctx context.Context, productID *string, warehouseID *string) (<-chan *InventoryItem, error)
	LowStockAlert(ctx context.Context, threshold *int) (<-chan []*InventoryItem, error)

	// Document subscriptions
	DocumentUploaded(ctx context.Context) (<-chan *Document, error)
	DocumentProcessed(ctx context.Context, documentID *string) (<-chan *Document, error)
}
