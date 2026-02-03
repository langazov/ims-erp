package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/internal/queries"
	"github.com/shopspring/decimal"
)

// Resolver is the root resolver for the GraphQL schema
type Resolver struct {
	ClientHandler    *queries.ClientQueryHandler
	WarehouseHandler *queries.WarehouseQueryHandler
	InventoryHandler *queries.InventoryQueryHandler
	DocumentHandler  *queries.DocumentQueryHandler
	InvoiceHandler   *queries.InvoiceQueryHandler
	PaymentHandler   *queries.PaymentQueryHandler
	EventPublisher   events.Publisher
	EventSubscriber  EventSubscriber
}

// EventSubscriber interface for subscriptions
type EventSubscriber interface {
	Subscribe(ctx context.Context, eventType string) (<-chan *events.EventEnvelope, error)
}

// Query resolvers
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Mutation resolvers
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Subscription resolvers
func (r *Resolver) Subscription() SubscriptionResolver {
	return &subscriptionResolver{r}
}

// queryResolver handles Query operations
type queryResolver struct {
	*Resolver
}

func (r *queryResolver) Client(ctx context.Context, id string) (*Client, error) {
	clientID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	tenantID := getTenantID(ctx)

	client, err := r.ClientHandler.GetClientByID(ctx, &queries.GetClientByIDQuery{
		ClientID: clientID.String(),
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	return mapClientSummaryToGraphQL(client), nil
}

func (r *queryResolver) Clients(ctx context.Context, filter *ClientFilter, first *int, after *string, last *int, before *string) (*ClientConnection, error) {
	tenantID := getTenantID(ctx)

	pageSize := 20
	if first != nil {
		pageSize = *first
	}

	page := 1
	if after != nil {
		// Parse cursor to get page number
		page = parseCursor(*after) + 1
	}

	result, err := r.ClientHandler.ListClients(ctx, &queries.ListClientsQuery{
		TenantID: tenantID,
		Page:     page,
		PageSize: pageSize,
		Search:   getString(filter.Search),
		Status:   getString(filter.Status),
	})
	if err != nil {
		return nil, err
	}

	return mapClientConnectionToGraphQL(result), nil
}

func (r *queryResolver) Invoice(ctx context.Context, id string) (*Invoice, error) {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	tenantID := getTenantID(ctx)

	result, err := r.InvoiceHandler.GetInvoiceByID(ctx, &queries.GetInvoiceByIDQuery{
		InvoiceID: invoiceID.String(),
		TenantID:  tenantID,
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return mapInvoiceSummaryToGraphQL(result), nil
}

func (r *queryResolver) Invoices(ctx context.Context, filter *InvoiceFilter, first *int, after *string, last *int, before *string) (*InvoiceConnection, error) {
	tenantID := getTenantID(ctx)

	pageSize := 20
	if first != nil {
		pageSize = *first
	}

	page := 1
	if after != nil {
		page = parseCursor(*after) + 1
	}

	query := &queries.ListInvoicesQuery{
		TenantID: tenantID,
		Page:     page,
		PageSize: pageSize,
	}

	if filter != nil {
		if filter.ClientID != nil {
			query.ClientID = *filter.ClientID
		}
		if filter.Status != nil {
			query.Status = *filter.Status
		}
		if filter.DateFrom != nil {
			query.StartDate = *filter.DateFrom
		}
		if filter.DateTo != nil {
			query.EndDate = *filter.DateTo
		}
	}

	result, err := r.InvoiceHandler.ListInvoices(ctx, query)
	if err != nil {
		return nil, err
	}

	return mapInvoiceConnectionToGraphQL(result), nil
}

func (r *queryResolver) Payment(ctx context.Context, id string) (*Payment, error) {
	paymentID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}

	tenantID := getTenantID(ctx)

	result, err := r.PaymentHandler.GetPaymentByID(ctx, &queries.GetPaymentByIDQuery{
		PaymentID: paymentID.String(),
		TenantID:  tenantID,
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return mapPaymentSummaryToGraphQL(result), nil
}

func (r *queryResolver) PaymentsByInvoice(ctx context.Context, invoiceID string) ([]*Payment, error) {
	invID, err := uuid.Parse(invoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	tenantID := getTenantID(ctx)

	result, err := r.PaymentHandler.GetPaymentsByInvoice(ctx, &queries.GetPaymentsByInvoiceQuery{
		InvoiceID: invID.String(),
		TenantID:  tenantID,
		Page:      1,
		PageSize:  100,
	})
	if err != nil {
		return nil, err
	}

	payments := make([]*Payment, 0, len(result.Payments))
	for _, p := range result.Payments {
		payments = append(payments, mapPaymentSummaryToGraphQL(&p))
	}

	return payments, nil
}

func (r *queryResolver) Warehouse(ctx context.Context, id string) (*Warehouse, error) {
	tenantID := getTenantID(ctx)

	result, err := r.WarehouseHandler.GetWarehouseByID(ctx, &queries.GetWarehouseByIDQuery{
		WarehouseID: id,
		TenantID:    tenantID,
	})
	if err != nil {
		return nil, err
	}

	return mapWarehouseDetailToGraphQL(result), nil
}

func (r *queryResolver) Warehouses(ctx context.Context, status *WarehouseStatus, first *int, after *string) (*WarehouseConnection, error) {
	tenantID := getTenantID(ctx)

	pageSize := 20
	if first != nil {
		pageSize = *first
	}

	page := 1
	if after != nil {
		page = parseCursor(*after) + 1
	}

	result, err := r.WarehouseHandler.ListWarehouses(ctx, &queries.ListWarehousesQuery{
		TenantID: tenantID,
		Page:     page,
		PageSize: pageSize,
		Status:   getWarehouseStatusString(status),
	})
	if err != nil {
		return nil, err
	}

	return mapWarehouseConnectionToGraphQL(result), nil
}

func (r *queryResolver) InventoryItem(ctx context.Context, id string) (*InventoryItem, error) {
	return &InventoryItem{}, nil
}

func (r *queryResolver) Inventory(ctx context.Context, filter *InventoryFilter, first *int, after *string) (*InventoryConnection, error) {
	tenantID := getTenantID(ctx)

	pageSize := 20
	if first != nil {
		pageSize = *first
	}

	result, err := r.InventoryHandler.ListInventory(ctx, &queries.ListInventoryQuery{
		TenantID:    tenantID,
		WarehouseID: getString(filter.WarehouseID),
		ProductID:   getString(filter.ProductID),
		Status:      getString(filter.Status),
		Page:        1,
		PageSize:    pageSize,
	})
	if err != nil {
		return nil, err
	}

	return mapInventoryConnectionToGraphQL(result), nil
}

func (r *queryResolver) Product(ctx context.Context, id string) (*Product, error) {
	return &Product{}, nil
}

func (r *queryResolver) Products(ctx context.Context, category *string, status *string, first *int, after *string) (*ProductConnection, error) {
	return &ProductConnection{}, nil
}

func (r *queryResolver) Document(ctx context.Context, id string) (*Document, error) {
	tenantID := getTenantID(ctx)

	doc, err := r.DocumentHandler.GetDocumentByID(ctx, &queries.GetDocumentByIDQuery{
		DocumentID: id,
		TenantID:   tenantID,
	})
	if err != nil {
		return nil, err
	}

	return mapDocumentDetailToGraphQL(doc), nil
}

func (r *queryResolver) Documents(ctx context.Context, docType *DocumentType, status *string, first *int, after *string) (*DocumentConnection, error) {
	tenantID := getTenantID(ctx)

	pageSize := 20
	if first != nil {
		pageSize = *first
	}

	typeStr := ""
	if docType != nil {
		typeStr = string(*docType)
	}

	result, err := r.DocumentHandler.ListDocuments(ctx, &queries.ListDocumentsQuery{
		TenantID: tenantID,
		Page:     1,
		PageSize: pageSize,
		Type:     typeStr,
		Status:   getString(status),
	})
	if err != nil {
		return nil, err
	}

	return mapDocumentConnectionToGraphQL(result), nil
}

func (r *queryResolver) Search(ctx context.Context, query string, types []string, first *int, after *string) (*SearchConnection, error) {
	return &SearchConnection{}, nil
}

// mutationResolver handles Mutation operations
type mutationResolver struct {
	*Resolver
}

func (r *mutationResolver) CreateClient(ctx context.Context, input CreateClientInput) (*Client, error) {
	// Implementation would call command handler
	return &Client{}, nil
}

func (r *mutationResolver) UpdateClient(ctx context.Context, id string, input UpdateClientInput) (*Client, error) {
	return &Client{}, nil
}

func (r *mutationResolver) DeleteClient(ctx context.Context, id string) (bool, error) {
	return true, nil
}

func (r *mutationResolver) ActivateClient(ctx context.Context, id string) (*Client, error) {
	return &Client{}, nil
}

func (r *mutationResolver) DeactivateClient(ctx context.Context, id string) (*Client, error) {
	return &Client{}, nil
}

func (r *mutationResolver) CreateInvoice(ctx context.Context, input CreateInvoiceInput) (*Invoice, error) {
	// Implementation would call command handler to create invoice
	// For now, return a mock invoice
	return &Invoice{
		ID:            uuid.New().String(),
		TenantID:      getTenantID(ctx),
		ClientID:      input.ClientID,
		InvoiceNumber: "INV-001",
		Status:        InvoiceStatusDraft,
		IssueDate:     time.Now(),
		DueDate:       input.DueDate,
		Total:         decimal.Zero,
		BalanceDue:    decimal.Zero,
		Currency:      getString(input.Currency),
		Notes:         getString(input.Notes),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

func (r *mutationResolver) IssueInvoice(ctx context.Context, id string) (*Invoice, error) {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	tenantID := getTenantID(ctx)

	// Fetch the invoice first
	result, err := r.InvoiceHandler.GetInvoiceByID(ctx, &queries.GetInvoiceByIDQuery{
		InvoiceID: invoiceID.String(),
		TenantID:  tenantID,
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Update status to sent
	result.Status = "sent"
	result.UpdatedAt = time.Now()

	return mapInvoiceSummaryToGraphQL(result), nil
}

func (r *mutationResolver) PayInvoice(ctx context.Context, id string, input CreatePaymentInput) (*Invoice, error) {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	tenantID := getTenantID(ctx)

	// Create payment
	payment := &Payment{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		InvoiceID: invoiceID.String(),
		Amount:    input.Amount,
		Currency:  "USD",
		Method:    input.Method,
		Status:    PaymentStatusCompleted,
		Reference: getString(input.Reference),
		Notes:     getString(input.Notes),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Fetch and update invoice
	result, err := r.InvoiceHandler.GetInvoiceByID(ctx, &queries.GetInvoiceByIDQuery{
		InvoiceID: invoiceID.String(),
		TenantID:  tenantID,
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Update invoice status based on payment
	// Parse amounts from strings
	amountPaid, _ := decimal.NewFromString(result.AmountPaid)
	total, _ := decimal.NewFromString(result.Total)
	amountDue, _ := decimal.NewFromString(result.AmountDue)

	amountPaid = amountPaid.Add(input.Amount)
	amountDue = total.Sub(amountPaid)

	result.AmountPaid = amountPaid.String()
	result.AmountDue = amountDue.String()

	if amountDue.LessThanOrEqual(decimal.Zero) {
		result.Status = "paid"
	} else {
		result.Status = "partial"
	}
	result.UpdatedAt = time.Now()

	// Store payment (in real implementation, call payment handler)
	_ = payment

	return mapInvoiceSummaryToGraphQL(result), nil
}

func (r *mutationResolver) VoidInvoice(ctx context.Context, id string) (*Invoice, error) {
	return &Invoice{}, nil
}

func (r *mutationResolver) CreatePayment(ctx context.Context, input CreatePaymentInput) (*Payment, error) {
	return &Payment{}, nil
}

func (r *mutationResolver) RefundPayment(ctx context.Context, id string, reason *string) (*Payment, error) {
	return &Payment{}, nil
}

func (r *mutationResolver) CreateWarehouse(ctx context.Context, input CreateWarehouseInput) (*Warehouse, error) {
	return &Warehouse{}, nil
}

func (r *mutationResolver) UpdateWarehouse(ctx context.Context, id string, name *string, address *AddressInput) (*Warehouse, error) {
	return &Warehouse{}, nil
}

func (r *mutationResolver) ActivateWarehouse(ctx context.Context, id string) (*Warehouse, error) {
	return &Warehouse{}, nil
}

func (r *mutationResolver) DeactivateWarehouse(ctx context.Context, id string) (*Warehouse, error) {
	return &Warehouse{}, nil
}

func (r *mutationResolver) ReceiveInventory(ctx context.Context, productID string, warehouseID string, quantity int, unitCost *decimal.Decimal) (*InventoryItem, error) {
	return &InventoryItem{}, nil
}

func (r *mutationResolver) ShipInventory(ctx context.Context, productID string, warehouseID string, quantity int) (*InventoryItem, error) {
	return &InventoryItem{}, nil
}

func (r *mutationResolver) AdjustInventory(ctx context.Context, productID string, warehouseID string, quantity int, reason string) (*InventoryItem, error) {
	return &InventoryItem{}, nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, input CreateProductInput) (*Product, error) {
	return &Product{}, nil
}

func (r *mutationResolver) UpdateProduct(ctx context.Context, id string, name *string, description *string, listPrice *decimal.Decimal) (*Product, error) {
	return &Product{}, nil
}

func (r *mutationResolver) ActivateProduct(ctx context.Context, id string) (*Product, error) {
	return &Product{}, nil
}

func (r *mutationResolver) DeactivateProduct(ctx context.Context, id string) (*Product, error) {
	return &Product{}, nil
}

func (r *mutationResolver) UploadDocument(ctx context.Context, file string, docType DocumentType, tags []string) (*Document, error) {
	return &Document{}, nil
}

func (r *mutationResolver) DeleteDocument(ctx context.Context, id string) (bool, error) {
	return true, nil
}

func (r *mutationResolver) ReprocessDocument(ctx context.Context, id string) (*Document, error) {
	return &Document{}, nil
}

// subscriptionResolver handles Subscription operations
type subscriptionResolver struct {
	*Resolver
}

func (r *subscriptionResolver) ClientCreated(ctx context.Context) (<-chan *Client, error) {
	ch := make(chan *Client)
	// Implementation would subscribe to events
	return ch, nil
}

func (r *subscriptionResolver) ClientUpdated(ctx context.Context, clientID *string) (<-chan *Client, error) {
	ch := make(chan *Client)
	return ch, nil
}

func (r *subscriptionResolver) InvoiceCreated(ctx context.Context) (<-chan *Invoice, error) {
	ch := make(chan *Invoice, 1)

	if r.EventSubscriber == nil {
		return ch, nil
	}

	// Subscribe to invoice.created events
	eventCh, err := r.EventSubscriber.Subscribe(ctx, "invoice.created")
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		for {
			select {
			case event := <-eventCh:
				if event == nil {
					return
				}
				invoice := mapEventToInvoice(event)
				if invoice != nil {
					select {
					case ch <- invoice:
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func (r *subscriptionResolver) InvoiceUpdated(ctx context.Context, invoiceID *string) (<-chan *Invoice, error) {
	ch := make(chan *Invoice, 1)

	if r.EventSubscriber == nil {
		return ch, nil
	}

	// Build filter
	filter := "invoice.updated"
	if invoiceID != nil {
		filter = fmt.Sprintf("invoice.updated.%s", *invoiceID)
	}

	eventCh, err := r.EventSubscriber.Subscribe(ctx, filter)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		for {
			select {
			case event := <-eventCh:
				if event == nil {
					return
				}
				invoice := mapEventToInvoice(event)
				if invoice != nil {
					select {
					case ch <- invoice:
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func (r *subscriptionResolver) InvoicePaid(ctx context.Context, invoiceID *string) (<-chan *Invoice, error) {
	ch := make(chan *Invoice, 1)

	if r.EventSubscriber == nil {
		return ch, nil
	}

	// Build filter
	filter := "invoice.paid"
	if invoiceID != nil {
		filter = fmt.Sprintf("invoice.paid.%s", *invoiceID)
	}

	eventCh, err := r.EventSubscriber.Subscribe(ctx, filter)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		for {
			select {
			case event := <-eventCh:
				if event == nil {
					return
				}
				invoice := mapEventToInvoice(event)
				if invoice != nil {
					select {
					case ch <- invoice:
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func (r *subscriptionResolver) PaymentReceived(ctx context.Context) (<-chan *Payment, error) {
	ch := make(chan *Payment, 1)

	if r.EventSubscriber == nil {
		return ch, nil
	}

	eventCh, err := r.EventSubscriber.Subscribe(ctx, "payment.received")
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		for {
			select {
			case event := <-eventCh:
				if event == nil {
					return
				}
				payment := mapEventToPayment(event)
				if payment != nil {
					select {
					case ch <- payment:
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func (r *subscriptionResolver) PaymentFailed(ctx context.Context) (<-chan *Payment, error) {
	ch := make(chan *Payment, 1)

	if r.EventSubscriber == nil {
		return ch, nil
	}

	eventCh, err := r.EventSubscriber.Subscribe(ctx, "payment.failed")
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		for {
			select {
			case event := <-eventCh:
				if event == nil {
					return
				}
				payment := mapEventToPayment(event)
				if payment != nil {
					select {
					case ch <- payment:
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func (r *subscriptionResolver) InventoryUpdated(ctx context.Context, productID *string, warehouseID *string) (<-chan *InventoryItem, error) {
	ch := make(chan *InventoryItem)
	return ch, nil
}

func (r *subscriptionResolver) LowStockAlert(ctx context.Context, threshold *int) (<-chan []*InventoryItem, error) {
	ch := make(chan []*InventoryItem)
	return ch, nil
}

func (r *subscriptionResolver) DocumentUploaded(ctx context.Context) (<-chan *Document, error) {
	ch := make(chan *Document)
	return ch, nil
}

func (r *subscriptionResolver) DocumentProcessed(ctx context.Context, documentID *string) (<-chan *Document, error) {
	ch := make(chan *Document)
	return ch, nil
}

// Helper functions
func getTenantID(ctx context.Context) string {
	// Extract tenant ID from context
	return "default-tenant"
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getWarehouseStatusString(s *WarehouseStatus) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func parseCursor(cursor string) int {
	// Parse cursor to get page number
	return 0
}

// Mapping functions
func mapClientSummaryToGraphQL(c *events.ClientSummary) *Client {
	return &Client{
		ID:        c.ID,
		TenantID:  c.TenantID,
		Name:      c.Name,
		Email:     c.Email,
		Status:    ClientStatus(c.Status),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func mapClientConnectionToGraphQL(result *queries.ListClientsResult) *ClientConnection {
	edges := make([]*ClientEdge, 0, len(result.Clients))
	for i, client := range result.Clients {
		edges = append(edges, &ClientEdge{
			Node:   mapClientSummaryToGraphQL(&client),
			Cursor: string(rune(i)),
		})
	}

	return &ClientConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage: result.Page < result.TotalPages,
			HasPrevPage: result.Page > 1,
			StartCursor: string(rune(0)),
			EndCursor:   string(rune(len(edges) - 1)),
			TotalCount:  int(result.Total),
		},
	}
}

func mapWarehouseDetailToGraphQL(w *queries.WarehouseDetail) *Warehouse {
	return &Warehouse{
		ID:        w.ID,
		TenantID:  w.TenantID,
		Name:      w.Name,
		Code:      w.Code,
		Type:      string(w.Type),
		Status:    WarehouseStatus(w.Status),
		IsActive:  w.IsActive,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func mapWarehouseConnectionToGraphQL(result *queries.ListWarehousesResult) *WarehouseConnection {
	edges := make([]*WarehouseEdge, 0, len(result.Warehouses))
	for i, wh := range result.Warehouses {
		edges = append(edges, &WarehouseEdge{
			Node:   mapWarehouseSummaryToGraphQL(&wh),
			Cursor: string(rune(i)),
		})
	}

	return &WarehouseConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage: result.Page < result.TotalPages,
			HasPrevPage: result.Page > 1,
			StartCursor: string(rune(0)),
			EndCursor:   string(rune(len(edges) - 1)),
			TotalCount:  int(result.Total),
		},
	}
}

func mapWarehouseSummaryToGraphQL(w *queries.WarehouseSummary) *Warehouse {
	return &Warehouse{
		ID:        w.ID,
		TenantID:  w.TenantID,
		Name:      w.Name,
		Code:      w.Code,
		Type:      w.Type,
		Status:    WarehouseStatus(w.Status),
		IsActive:  w.IsActive,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func mapInventoryItemToGraphQL(item *queries.InventoryItemDetail) *InventoryItem {
	return &InventoryItem{
		ID:           item.ID,
		TenantID:     item.TenantID,
		ProductID:    item.ProductID,
		WarehouseID:  item.WarehouseID,
		SKU:          item.SKU,
		Quantity:     item.Quantity,
		ReservedQty:  item.ReservedQty,
		AvailableQty: item.AvailableQty,
		Status:       InventoryStatus(item.Status),
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func mapInventorySummaryToDetail(summary *queries.InventoryItemSummary) *queries.InventoryItemDetail {
	return &queries.InventoryItemDetail{
		ID:            summary.ID,
		TenantID:      summary.TenantID,
		ProductID:     summary.ProductID,
		SKU:           summary.SKU,
		WarehouseID:   summary.WarehouseID,
		LocationID:    summary.LocationID,
		Quantity:      summary.Quantity,
		ReservedQty:   summary.ReservedQty,
		AvailableQty:  summary.AvailableQty,
		AllocatedQty:  summary.AllocatedQty,
		Status:        summary.Status,
		UnitCost:      summary.UnitCost,
		TotalValue:    summary.TotalValue,
		LastCountedAt: summary.LastCountedAt,
		CreatedAt:     summary.CreatedAt,
		UpdatedAt:     summary.UpdatedAt,
	}
}

func mapInventoryConnectionToGraphQL(result *queries.ListInventoryResult) *InventoryConnection {
	edges := make([]*InventoryEdge, 0, len(result.Items))
	for i, item := range result.Items {
		detail := mapInventorySummaryToDetail(&item)
		edges = append(edges, &InventoryEdge{
			Node:   mapInventoryItemToGraphQL(detail),
			Cursor: string(rune(i)),
		})
	}

	return &InventoryConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage: result.Page < result.TotalPages,
			HasPrevPage: result.Page > 1,
			StartCursor: string(rune(0)),
			EndCursor:   string(rune(len(edges) - 1)),
			TotalCount:  int(result.Total),
		},
	}
}

func mapDocumentDetailToGraphQL(doc *queries.DocumentDetail) *Document {
	return &Document{
		ID:               doc.ID,
		TenantID:         doc.TenantID,
		Type:             DocumentType(doc.Type),
		FileName:         doc.FileName,
		MimeType:         doc.MimeType,
		Size:             doc.Size,
		ProcessingStatus: doc.ProcessingStatus,
		PageCount:        doc.PageCount,
		Tags:             doc.Tags,
		UploadedBy:       doc.UploadedBy,
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        doc.UpdatedAt,
	}
}

func mapDocumentSummaryToDetail(summary *queries.DocumentSummary) *queries.DocumentDetail {
	return &queries.DocumentDetail{
		ID:               summary.ID,
		TenantID:         summary.TenantID,
		Type:             summary.Type,
		FileName:         summary.FileName,
		MimeType:         summary.MimeType,
		Size:             summary.Size,
		Checksum:         summary.Checksum,
		ProcessingStatus: summary.ProcessingStatus,
		PageCount:        summary.PageCount,
		Tags:             summary.Tags,
		UploadedBy:       summary.UploadedBy,
		CreatedAt:        summary.CreatedAt,
		UpdatedAt:        summary.UpdatedAt,
	}
}

func mapDocumentConnectionToGraphQL(result *queries.ListDocumentsResult) *DocumentConnection {
	edges := make([]*DocumentEdge, 0, len(result.Documents))
	for i, doc := range result.Documents {
		detail := mapDocumentSummaryToDetail(&doc)
		edges = append(edges, &DocumentEdge{
			Node:   mapDocumentDetailToGraphQL(detail),
			Cursor: string(rune(i)),
		})
	}

	return &DocumentConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage: result.Page < result.TotalPages,
			HasPrevPage: result.Page > 1,
			StartCursor: string(rune(0)),
			EndCursor:   string(rune(len(edges) - 1)),
			TotalCount:  int(result.Total),
		},
	}
}

// Invoice and Payment mapping functions
func mapInvoiceSummaryToGraphQL(inv *events.InvoiceSummary) *Invoice {
	if inv == nil {
		return nil
	}

	subtotal, _ := decimal.NewFromString(inv.Subtotal)
	taxTotal, _ := decimal.NewFromString(inv.TaxTotal)
	discountTotal, _ := decimal.NewFromString(inv.DiscountTotal)
	total, _ := decimal.NewFromString(inv.Total)
	balanceDue, _ := decimal.NewFromString(inv.AmountDue)

	return &Invoice{
		ID:            inv.ID,
		TenantID:      inv.TenantID,
		ClientID:      inv.ClientID,
		InvoiceNumber: inv.InvoiceNumber,
		Status:        InvoiceStatus(inv.Status),
		IssueDate:     inv.IssueDate,
		DueDate:       inv.DueDate,
		Subtotal:      subtotal,
		TaxTotal:      taxTotal,
		DiscountTotal: discountTotal,
		Total:         total,
		BalanceDue:    balanceDue,
		Currency:      inv.Currency,
		Notes:         inv.Notes,
		CreatedAt:     inv.CreatedAt,
		UpdatedAt:     inv.UpdatedAt,
	}
}

func mapPaymentSummaryToGraphQL(payment *events.PaymentSummary) *Payment {
	if payment == nil {
		return nil
	}

	amount, _ := decimal.NewFromString(payment.Amount)

	return &Payment{
		ID:        payment.ID,
		TenantID:  payment.TenantID,
		InvoiceID: payment.InvoiceID,
		Amount:    amount,
		Currency:  payment.Currency,
		Method:    payment.Method,
		Status:    PaymentStatus(payment.Status),
		Reference: payment.Reference,
		Notes:     payment.Description,
		CreatedAt: payment.CreatedAt,
		UpdatedAt: payment.UpdatedAt,
	}
}

func mapInvoiceConnectionToGraphQL(result *queries.ListInvoicesResult) *InvoiceConnection {
	edges := make([]*InvoiceEdge, 0, len(result.Invoices))
	for i, inv := range result.Invoices {
		edges = append(edges, &InvoiceEdge{
			Node:   mapInvoiceSummaryToGraphQL(&inv),
			Cursor: string(rune(i)),
		})
	}

	return &InvoiceConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage: result.Page < result.TotalPages,
			HasPrevPage: result.Page > 1,
			StartCursor: string(rune(0)),
			EndCursor:   string(rune(len(edges) - 1)),
			TotalCount:  int(result.Total),
		},
	}
}

func mapEventToInvoice(event *events.EventEnvelope) *Invoice {
	if event == nil {
		return nil
	}

	var invoice events.InvoiceSummary
	data, err := json.Marshal(event.Data)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(data, &invoice); err != nil {
		return nil
	}

	return mapInvoiceSummaryToGraphQL(&invoice)
}

func mapEventToPayment(event *events.EventEnvelope) *Payment {
	if event == nil {
		return nil
	}

	var payment events.PaymentSummary
	data, err := json.Marshal(event.Data)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(data, &payment); err != nil {
		return nil
	}

	return mapPaymentSummaryToGraphQL(&payment)
}
