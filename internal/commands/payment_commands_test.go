package commands

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPaymentRepo struct {
	payments map[uuid.UUID]*domain.Payment
}

func newMockPaymentRepo() *mockPaymentRepo {
	return &mockPaymentRepo{
		payments: make(map[uuid.UUID]*domain.Payment),
	}
}

func (r *mockPaymentRepo) Create(ctx context.Context, payment *domain.Payment) error {
	r.payments[payment.ID] = payment
	return nil
}

func (r *mockPaymentRepo) Update(ctx context.Context, payment *domain.Payment) error {
	r.payments[payment.ID] = payment
	return nil
}

func (r *mockPaymentRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	if payment, ok := r.payments[id]; ok {
		return payment, nil
	}
	return nil, nil
}

func (r *mockPaymentRepo) FindByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]*domain.Payment, error) {
	var result []*domain.Payment
	for _, p := range r.payments {
		if p.InvoiceID == invoiceID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *mockPaymentRepo) FindByProviderID(ctx context.Context, providerID string) (*domain.Payment, error) {
	for _, p := range r.payments {
		if p.ProviderID == providerID {
			return p, nil
		}
	}
	return nil, nil
}

type mockInvoiceRepoForPayment struct {
	invoices map[uuid.UUID]*domain.Invoice
}

func newMockInvoiceRepoForPayment() *mockInvoiceRepoForPayment {
	return &mockInvoiceRepoForPayment{
		invoices: make(map[uuid.UUID]*domain.Invoice),
	}
}

func (r *mockInvoiceRepoForPayment) Create(ctx context.Context, invoice *domain.Invoice) error {
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *mockInvoiceRepoForPayment) Update(ctx context.Context, invoice *domain.Invoice) error {
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *mockInvoiceRepoForPayment) FindByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	if invoice, ok := r.invoices[id]; ok {
		return invoice, nil
	}
	return nil, nil
}

func (r *mockInvoiceRepoForPayment) FindByInvoiceNumber(ctx context.Context, tenantID uuid.UUID, invoiceNumber string) (*domain.Invoice, error) {
	return nil, nil
}

func (r *mockInvoiceRepoForPayment) FindByClientID(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*domain.Invoice, error) {
	return nil, nil
}

func TestPaymentCommandHandler_HandleCreatePayment(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepoForPayment()
	publisher := &mockPublisher{}
	processors := domain.NewProcessorRegistry()
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewPaymentCommandHandler(paymentRepo, invoiceRepo, nil, publisher, log, processors)

	cmd := &CommandEnvelope{
		Type:     "createPayment",
		TenantID: uuid.New().String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"invoiceId":   uuid.New().String(),
			"clientId":    uuid.New().String(),
			"amount":      "500.00",
			"currency":    "USD",
			"method":      "credit_card",
			"provider":    "stripe",
			"reference":   "ref-12345",
			"description": "Test payment",
		},
	}

	payment, err := handler.HandleCreatePayment(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, payment)
	assert.NotEqual(t, uuid.Nil, payment.ID)
	assert.Equal(t, domain.PaymentStatusPending, payment.Status)
	assert.Equal(t, "USD", payment.Currency)
	assert.Equal(t, domain.PaymentMethodCreditCard, payment.Method)
	assert.Equal(t, "stripe", payment.Provider)
	assert.Equal(t, "ref-12345", payment.Reference)
	assert.Equal(t, 0, payment.Amount.Cmp(decimal.NewFromInt(500)))
	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "payment.created", publisher.events[0].Type)
}

func TestPaymentCommandHandler_HandleCreatePayment_InvalidAmount(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepoForPayment()
	publisher := &mockPublisher{}
	processors := domain.NewProcessorRegistry()
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})

	handler := NewPaymentCommandHandler(paymentRepo, invoiceRepo, nil, publisher, log, processors)

	cmd := &CommandEnvelope{
		Type:     "createPayment",
		TenantID: uuid.New().String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"invoiceId": uuid.New().String(),
			"clientId":  uuid.New().String(),
			"amount":    "0",
			"currency":  "USD",
			"method":    "credit_card",
		},
	}

	payment, err := handler.HandleCreatePayment(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Contains(t, err.Error(), "payment amount must be greater than zero")
}

func TestPaymentCommandHandler_HandleProcessPayment_Success(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepoForPayment()
	publisher := &mockPublisher{}

	// Register a mock Stripe processor
	processors := domain.NewProcessorRegistry()
	processors.Register("stripe", func(provider string, config interface{}) (domain.PaymentProcessor, error) {
		return &domain.StripeProcessor{}, nil
	})

	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})
	handler := NewPaymentCommandHandler(paymentRepo, invoiceRepo, nil, publisher, log, processors)

	tenantID := uuid.New()
	invoiceID := uuid.New()
	clientID := uuid.New()

	// Create an invoice
	invoice := &domain.Invoice{
		ID:        invoiceID,
		TenantID:  tenantID,
		ClientID:  clientID,
		Status:    domain.InvoiceStatusSent,
		Total:     decimal.NewFromInt(500),
		AmountDue: decimal.NewFromInt(500),
	}
	invoiceRepo.Create(context.Background(), invoice)

	// Create a pending payment
	payment := domain.NewPayment(tenantID, invoiceID, clientID, decimal.NewFromInt(500), "USD", domain.PaymentMethodCreditCard)
	payment.Provider = "stripe"
	paymentRepo.Create(context.Background(), payment)

	cmd := &CommandEnvelope{
		Type:     "processPayment",
		TenantID: tenantID.String(),
		TargetID: payment.ID.String(),
		UserID:   uuid.New().String(),
		Data:     map[string]interface{}{},
	}

	processedPayment, err := handler.HandleProcessPayment(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, processedPayment)
	assert.Equal(t, domain.PaymentStatusCompleted, processedPayment.Status)
	assert.NotNil(t, processedPayment.ProcessedAt)
	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "payment.processed", publisher.events[0].Type)

	// Verify invoice was updated
	updatedInvoice, _ := invoiceRepo.FindByID(context.Background(), invoiceID)
	assert.Equal(t, domain.InvoiceStatusPaid, updatedInvoice.Status)
}

func TestPaymentCommandHandler_HandleProcessPayment_NotPending(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepoForPayment()
	publisher := &mockPublisher{}
	processors := domain.NewProcessorRegistry()
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})
	handler := NewPaymentCommandHandler(paymentRepo, invoiceRepo, nil, publisher, log, processors)

	tenantID := uuid.New()
	payment := domain.NewPayment(tenantID, uuid.New(), uuid.New(), decimal.NewFromInt(500), "USD", domain.PaymentMethodCreditCard)
	payment.Provider = "stripe"
	payment.MarkAsCompleted(time.Now().UTC())
	paymentRepo.Create(context.Background(), payment)

	cmd := &CommandEnvelope{
		Type:     "processPayment",
		TenantID: tenantID.String(),
		TargetID: payment.ID.String(),
		UserID:   uuid.New().String(),
		Data:     map[string]interface{}{},
	}

	processedPayment, err := handler.HandleProcessPayment(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, processedPayment)
	assert.Contains(t, err.Error(), "payment is not in pending status")
}

func TestPaymentCommandHandler_HandleRefundPayment(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepoForPayment()
	publisher := &mockPublisher{}

	processors := domain.NewProcessorRegistry()
	processors.Register("stripe", func(provider string, config interface{}) (domain.PaymentProcessor, error) {
		return &domain.StripeProcessor{}, nil
	})

	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})
	handler := NewPaymentCommandHandler(paymentRepo, invoiceRepo, nil, publisher, log, processors)

	tenantID := uuid.New()
	payment := domain.NewPayment(tenantID, uuid.New(), uuid.New(), decimal.NewFromInt(500), "USD", domain.PaymentMethodCreditCard)
	payment.Provider = "stripe"
	payment.MarkAsProcessing("pi_test", "tx_test")
	payment.MarkAsCompleted(time.Now().UTC())
	paymentRepo.Create(context.Background(), payment)

	cmd := &CommandEnvelope{
		Type:     "refundPayment",
		TenantID: tenantID.String(),
		TargetID: payment.ID.String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"amount": "500.00",
			"reason": "Customer request",
		},
	}

	refundedPayment, err := handler.HandleRefundPayment(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, refundedPayment)
	assert.Equal(t, domain.PaymentStatusRefunded, refundedPayment.Status)
	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "payment.refunded", publisher.events[0].Type)
}

func TestPaymentCommandHandler_HandleRefundPayment_NotCompleted(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepoForPayment()
	publisher := &mockPublisher{}
	processors := domain.NewProcessorRegistry()
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})
	handler := NewPaymentCommandHandler(paymentRepo, invoiceRepo, nil, publisher, log, processors)

	tenantID := uuid.New()
	payment := domain.NewPayment(tenantID, uuid.New(), uuid.New(), decimal.NewFromInt(500), "USD", domain.PaymentMethodCreditCard)
	payment.Provider = "stripe"
	// Leave as pending status
	paymentRepo.Create(context.Background(), payment)

	cmd := &CommandEnvelope{
		Type:     "refundPayment",
		TenantID: tenantID.String(),
		TargetID: payment.ID.String(),
		UserID:   uuid.New().String(),
		Data: map[string]interface{}{
			"reason": "Customer request",
		},
	}

	refundedPayment, err := handler.HandleRefundPayment(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, refundedPayment)
	assert.Contains(t, err.Error(), "can only refund completed payments")
}

func TestPaymentCommandHandler_HandleCancelPayment(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepoForPayment()
	publisher := &mockPublisher{}
	processors := domain.NewProcessorRegistry()
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})
	handler := NewPaymentCommandHandler(paymentRepo, invoiceRepo, nil, publisher, log, processors)

	tenantID := uuid.New()
	payment := domain.NewPayment(tenantID, uuid.New(), uuid.New(), decimal.NewFromInt(500), "USD", domain.PaymentMethodCreditCard)
	payment.Provider = "stripe"
	paymentRepo.Create(context.Background(), payment)

	cmd := &CommandEnvelope{
		Type:     "cancelPayment",
		TenantID: tenantID.String(),
		TargetID: payment.ID.String(),
		UserID:   uuid.New().String(),
		Data:     map[string]interface{}{},
	}

	cancelledPayment, err := handler.HandleCancelPayment(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, cancelledPayment)
	assert.Equal(t, domain.PaymentStatusCancelled, cancelledPayment.Status)
	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "payment.cancelled", publisher.events[0].Type)
}

func TestPaymentCommandHandler_HandleCancelPayment_AlreadyCompleted(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepoForPayment()
	publisher := &mockPublisher{}
	processors := domain.NewProcessorRegistry()
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})
	handler := NewPaymentCommandHandler(paymentRepo, invoiceRepo, nil, publisher, log, processors)

	tenantID := uuid.New()
	payment := domain.NewPayment(tenantID, uuid.New(), uuid.New(), decimal.NewFromInt(500), "USD", domain.PaymentMethodCreditCard)
	payment.Provider = "stripe"
	payment.MarkAsProcessing("pi_test", "tx_test")
	payment.MarkAsCompleted(time.Now().UTC())
	paymentRepo.Create(context.Background(), payment)

	cmd := &CommandEnvelope{
		Type:     "cancelPayment",
		TenantID: tenantID.String(),
		TargetID: payment.ID.String(),
		UserID:   uuid.New().String(),
		Data:     map[string]interface{}{},
	}

	cancelledPayment, err := handler.HandleCancelPayment(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, cancelledPayment)
	assert.Contains(t, err.Error(), "cannot cancel completed payments")
}

func TestPaymentCommandHandler_HandleCancelPayment_AlreadyRefunded(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepoForPayment()
	publisher := &mockPublisher{}
	processors := domain.NewProcessorRegistry()
	log, _ := logger.New(logger.Config{Level: "error", Format: "json", ServiceName: "test"})
	handler := NewPaymentCommandHandler(paymentRepo, invoiceRepo, nil, publisher, log, processors)

	tenantID := uuid.New()
	payment := domain.NewPayment(tenantID, uuid.New(), uuid.New(), decimal.NewFromInt(500), "USD", domain.PaymentMethodCreditCard)
	payment.Provider = "stripe"
	payment.MarkAsRefunded()
	paymentRepo.Create(context.Background(), payment)

	cmd := &CommandEnvelope{
		Type:     "cancelPayment",
		TenantID: tenantID.String(),
		TargetID: payment.ID.String(),
		UserID:   uuid.New().String(),
		Data:     map[string]interface{}{},
	}

	cancelledPayment, err := handler.HandleCancelPayment(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, cancelledPayment)
	assert.Contains(t, err.Error(), "payment is already refunded")
}
