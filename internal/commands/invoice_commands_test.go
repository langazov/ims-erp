package commands

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	eventpkg "github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockInvoiceRepo struct {
	invoices map[uuid.UUID]*domain.Invoice
}

func newMockInvoiceRepo() *mockInvoiceRepo {
	return &mockInvoiceRepo{
		invoices: make(map[uuid.UUID]*domain.Invoice),
	}
}

func (r *mockInvoiceRepo) Create(ctx context.Context, invoice *domain.Invoice) error {
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *mockInvoiceRepo) Update(ctx context.Context, invoice *domain.Invoice) error {
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *mockInvoiceRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	if invoice, ok := r.invoices[id]; ok {
		return invoice, nil
	}
	return nil, nil
}

func (r *mockInvoiceRepo) FindByInvoiceNumber(ctx context.Context, tenantID uuid.UUID, invoiceNumber string) (*domain.Invoice, error) {
	for _, inv := range r.invoices {
		if inv.TenantID == tenantID && inv.InvoiceNumber == invoiceNumber {
			return inv, nil
		}
	}
	return nil, nil
}

func (r *mockInvoiceRepo) FindByClientID(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*domain.Invoice, error) {
	var result []*domain.Invoice
	for _, inv := range r.invoices {
		if inv.ClientID == clientID {
			result = append(result, inv)
		}
	}
	return result, nil
}

type mockInvoiceCounter struct {
	counter int
}

func (c *mockInvoiceCounter) GetNextInvoiceNumber(ctx context.Context, tenantID uuid.UUID, year int) (string, error) {
	c.counter++
	return uuid.New().String(), nil
}

type mockPublisher struct {
	events []*eventpkg.EventEnvelope
}

func (p *mockPublisher) PublishEvent(ctx context.Context, event *eventpkg.EventEnvelope) error {
	p.events = append(p.events, event)
	return nil
}

func TestInvoiceCommandHandler_HandleCreateInvoice(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	cmd := &CommandEnvelope{
		Type:     "createInvoice",
		TenantID: uuid.New().String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"clientId":    uuid.New().String(),
			"type":        "standard",
			"currency":    "USD",
			"paymentTerm": "net_30",
			"notes":       "Test invoice",
			"terms":       "Net 30 days",
		},
	}

	invoice, err := handler.HandleCreateInvoice(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, invoice)
	assert.NotEqual(t, uuid.Nil, invoice.ID)
	assert.Equal(t, domain.InvoiceStatusDraft, invoice.Status)
	assert.Equal(t, "USD", invoice.Currency)
	assert.Equal(t, "Test invoice", invoice.Notes)
	assert.Equal(t, "Net 30 days", invoice.Terms)
	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "invoice.created", publisher.events[0].Type)
}

func TestInvoiceCommandHandler_HandleCreateInvoice_InvalidTenantID(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	cmd := &CommandEnvelope{
		Type:     "createInvoice",
		TenantID: "invalid-uuid",
		UserID:   uuid.New().String(),
		Data:     map[string]interface{}{},
	}

	invoice, err := handler.HandleCreateInvoice(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "invalid tenant ID")
}

func TestInvoiceCommandHandler_HandleAddLineItem(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	// First create an invoice
	tenantID := uuid.New()
	clientID := uuid.New()
	userID := uuid.New()

	invoice := &domain.Invoice{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ClientID:  clientID,
		Status:    domain.InvoiceStatusDraft,
		Currency:  "USD",
		Lines:     []domain.InvoiceLine{},
		Subtotal:  decimal.Zero,
		TaxTotal:  decimal.Zero,
		Total:     decimal.Zero,
		AmountDue: decimal.Zero,
		Version:   0,
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "addLineItem",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   userID.String(),
		Data: map[string]interface{}{
			"description": "Test product",
			"quantity":    "5",
			"unitPrice":   "100.00",
			"discount":    "0",
			"taxRate":     "20",
			"sortOrder":   1,
		},
	}

	updatedInvoice, err := handler.HandleAddLineItem(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, updatedInvoice)
	assert.Len(t, updatedInvoice.Lines, 1)
	assert.Equal(t, "Test product", updatedInvoice.Lines[0].Description)
	assert.Equal(t, 0, updatedInvoice.Lines[0].Quantity.Cmp(decimal.NewFromInt(5)))
	assert.Equal(t, 0, updatedInvoice.Lines[0].Total.Cmp(decimal.NewFromInt(500)))
	assert.Equal(t, 0, updatedInvoice.Total.Cmp(decimal.NewFromInt(600))) // 500 + 20% tax
}

func TestInvoiceCommandHandler_HandleAddLineItem_NotDraft(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	tenantID := uuid.New()
	invoice := &domain.Invoice{
		ID:       uuid.New(),
		TenantID: tenantID,
		Status:   domain.InvoiceStatusSent,
		Lines:    []domain.InvoiceLine{},
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "addLineItem",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"description": "Test product",
			"quantity":    "5",
			"unitPrice":   "100.00",
		},
	}

	updatedInvoice, err := handler.HandleAddLineItem(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, updatedInvoice)
	assert.Contains(t, err.Error(), "can only add lines to draft invoices")
}

func TestInvoiceCommandHandler_HandleFinalizeInvoice(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	tenantID := uuid.New()
	line := domain.InvoiceLine{
		ID:          uuid.New(),
		Description: "Test product",
		Quantity:    decimal.NewFromInt(1),
		UnitPrice:   decimal.NewFromInt(100),
		TaxRate:     decimal.Zero,
		TaxAmount:   decimal.Zero,
		Total:       decimal.NewFromInt(100),
	}

	invoice := &domain.Invoice{
		ID:       uuid.New(),
		TenantID: tenantID,
		Status:   domain.InvoiceStatusDraft,
		Lines:    []domain.InvoiceLine{line},
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "finalizeInvoice",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   uuid.New().String(),
		Data:     map[string]interface{}{},
	}

	updatedInvoice, err := handler.HandleFinalizeInvoice(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, updatedInvoice)
	assert.Equal(t, domain.InvoiceStatusPending, updatedInvoice.Status)
	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "invoice.finalized", publisher.events[0].Type)
}

func TestInvoiceCommandHandler_HandleFinalizeInvoice_NoLines(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	tenantID := uuid.New()
	invoice := &domain.Invoice{
		ID:       uuid.New(),
		TenantID: tenantID,
		Status:   domain.InvoiceStatusDraft,
		Lines:    []domain.InvoiceLine{},
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "finalizeInvoice",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   uuid.New().String(),
		Data:     map[string]interface{}{},
	}

	updatedInvoice, err := handler.HandleFinalizeInvoice(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, updatedInvoice)
	assert.Contains(t, err.Error(), "cannot finalize invoice with no line items")
}

func TestInvoiceCommandHandler_HandleSendInvoice(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	tenantID := uuid.New()
	invoice := &domain.Invoice{
		ID:       uuid.New(),
		TenantID: tenantID,
		Status:   domain.InvoiceStatusPending,
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "sendInvoice",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   uuid.New().String(),
		Data:     map[string]interface{}{},
	}

	updatedInvoice, err := handler.HandleSendInvoice(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, updatedInvoice)
	assert.Equal(t, domain.InvoiceStatusSent, updatedInvoice.Status)
	assert.NotNil(t, updatedInvoice.SentDate)
	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "invoice.sent", publisher.events[0].Type)
}

func TestInvoiceCommandHandler_HandleVoidInvoice(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	tenantID := uuid.New()
	invoice := &domain.Invoice{
		ID:       uuid.New(),
		TenantID: tenantID,
		Status:   domain.InvoiceStatusSent,
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "voidInvoice",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"reason": "Customer cancellation",
		},
	}

	updatedInvoice, err := handler.HandleVoidInvoice(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, updatedInvoice)
	assert.Equal(t, domain.InvoiceStatusCancelled, updatedInvoice.Status)
	assert.Contains(t, updatedInvoice.Notes, "Customer cancellation")
	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "invoice.voided", publisher.events[0].Type)
}

func TestInvoiceCommandHandler_HandleVoidInvoice_PaidInvoice(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	tenantID := uuid.New()
	invoice := &domain.Invoice{
		ID:       uuid.New(),
		TenantID: tenantID,
		Status:   domain.InvoiceStatusPaid,
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "voidInvoice",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   uuid.New().String(),
		Data:     map[string]interface{}{},
	}

	updatedInvoice, err := handler.HandleVoidInvoice(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, updatedInvoice)
	assert.Contains(t, err.Error(), "cannot void a paid invoice")
}

func TestInvoiceCommandHandler_HandleRecordPayment(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	tenantID := uuid.New()
	invoice := &domain.Invoice{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Status:     domain.InvoiceStatusSent,
		Total:      decimal.NewFromInt(500),
		AmountPaid: decimal.Zero,
		AmountDue:  decimal.NewFromInt(500),
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "recordPayment",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"amount":        "500.00",
			"paymentMethod": "credit_card",
			"reference":     "ref-12345",
		},
	}

	updatedInvoice, err := handler.HandleRecordPayment(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, updatedInvoice)
	assert.Equal(t, domain.InvoiceStatusPaid, updatedInvoice.Status)
	assert.Equal(t, 0, updatedInvoice.AmountPaid.Cmp(decimal.NewFromInt(500)))
	assert.Equal(t, 0, updatedInvoice.AmountDue.Cmp(decimal.Zero))
	assert.NotNil(t, updatedInvoice.PaidDate)
	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "invoice.payment_recorded", publisher.events[0].Type)
}

func TestInvoiceCommandHandler_HandleRecordPayment_Partial(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	tenantID := uuid.New()
	invoice := &domain.Invoice{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Status:     domain.InvoiceStatusSent,
		Total:      decimal.NewFromInt(500),
		AmountPaid: decimal.Zero,
		AmountDue:  decimal.NewFromInt(500),
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "recordPayment",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"amount": "200.00",
		},
	}

	updatedInvoice, err := handler.HandleRecordPayment(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, updatedInvoice)
	assert.Equal(t, domain.InvoiceStatusSent, updatedInvoice.Status) // Not fully paid yet
	assert.Equal(t, 0, updatedInvoice.AmountPaid.Cmp(decimal.NewFromInt(200)))
	assert.Equal(t, 0, updatedInvoice.AmountDue.Cmp(decimal.NewFromInt(300)))
}

func TestInvoiceCommandHandler_HandleRecordPayment_ExceedsAmount(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	tenantID := uuid.New()
	invoice := &domain.Invoice{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Status:    domain.InvoiceStatusSent,
		Total:     decimal.NewFromInt(500),
		AmountDue: decimal.NewFromInt(500),
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "recordPayment",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"amount": "600.00",
		},
	}

	updatedInvoice, err := handler.HandleRecordPayment(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, updatedInvoice)
	assert.Contains(t, err.Error(), "payment amount exceeds amount due")
}

func TestInvoiceCommandHandler_HandleRemoveLineItem(t *testing.T) {
	repo := newMockInvoiceRepo()
	publisher := &mockPublisher{}
	counter := &mockInvoiceCounter{}
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewInvoiceCommandHandler(repo, nil, publisher, log, counter)

	tenantID := uuid.New()
	lineID := uuid.New()
	line := domain.InvoiceLine{
		ID:          lineID,
		Description: "Test product",
		Quantity:    decimal.NewFromInt(1),
		UnitPrice:   decimal.NewFromInt(100),
		Total:       decimal.NewFromInt(100),
	}

	invoice := &domain.Invoice{
		ID:       uuid.New(),
		TenantID: tenantID,
		Status:   domain.InvoiceStatusDraft,
		Lines:    []domain.InvoiceLine{line},
		Total:    decimal.NewFromInt(100),
	}
	repo.Create(context.Background(), invoice)

	cmd := &CommandEnvelope{
		Type:     "removeLineItem",
		TenantID: tenantID.String(),
		TargetID: invoice.ID.String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"lineId": lineID.String(),
		},
	}

	updatedInvoice, err := handler.HandleRemoveLineItem(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, updatedInvoice)
	assert.Len(t, updatedInvoice.Lines, 0)
	assert.Equal(t, 0, updatedInvoice.Total.Cmp(decimal.Zero))
	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "invoice.line_removed", publisher.events[0].Type)
}
