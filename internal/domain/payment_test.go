package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPayment(t *testing.T) {
	tenantID := uuid.New()
	invoiceID := uuid.New()
	clientID := uuid.New()
	amount := decimal.NewFromFloat(100.00)

	payment := NewPayment(tenantID, invoiceID, clientID, amount, "USD", PaymentMethodCreditCard)

	assert.NotEmpty(t, payment.ID)
	assert.Equal(t, tenantID, payment.TenantID)
	assert.Equal(t, invoiceID, payment.InvoiceID)
	assert.Equal(t, clientID, payment.ClientID)
	assert.True(t, payment.Amount.Equal(amount))
	assert.Equal(t, "USD", payment.Currency)
	assert.Equal(t, PaymentMethodCreditCard, payment.Method)
	assert.Equal(t, PaymentStatusPending, payment.Status)
	assert.False(t, payment.CreatedAt.IsZero())
}

func TestPaymentMarkAsProcessing(t *testing.T) {
	payment := NewPayment(uuid.New(), uuid.New(), uuid.New(), decimal.NewFromFloat(100.00), "USD", PaymentMethodCreditCard)

	payment.MarkAsProcessing("proc_123", "txn_123")

	assert.Equal(t, PaymentStatusProcessing, payment.Status)
	assert.Equal(t, "proc_123", payment.ProviderID)
	assert.Equal(t, "txn_123", payment.TransactionID)
}

func TestPaymentMarkAsCompleted(t *testing.T) {
	payment := NewPayment(uuid.New(), uuid.New(), uuid.New(), decimal.NewFromFloat(100.00), "USD", PaymentMethodCreditCard)

	processedAt := time.Now()
	payment.MarkAsCompleted(processedAt)

	assert.Equal(t, PaymentStatusCompleted, payment.Status)
	assert.NotNil(t, payment.ProcessedAt)
}

func TestPaymentMarkAsFailed(t *testing.T) {
	payment := NewPayment(uuid.New(), uuid.New(), uuid.New(), decimal.NewFromFloat(100.00), "USD", PaymentMethodCreditCard)

	payment.MarkAsFailed("card_declined", "Your card was declined")

	assert.Equal(t, PaymentStatusFailed, payment.Status)
	assert.Equal(t, "card_declined", payment.FailureCode)
	assert.Equal(t, "Your card was declined", payment.FailureMessage)
}

func TestPaymentMarkAsRefunded(t *testing.T) {
	payment := NewPayment(uuid.New(), uuid.New(), uuid.New(), decimal.NewFromFloat(100.00), "USD", PaymentMethodCreditCard)

	payment.MarkAsRefunded()

	assert.Equal(t, PaymentStatusRefunded, payment.Status)
}

func TestPaymentSetReference(t *testing.T) {
	payment := NewPayment(uuid.New(), uuid.New(), uuid.New(), decimal.NewFromFloat(100.00), "USD", PaymentMethodCreditCard)

	payment.SetReference("REF-2024-001")

	assert.Equal(t, "REF-2024-001", payment.Reference)
}

func TestPaymentSetMetadata(t *testing.T) {
	payment := NewPayment(uuid.New(), uuid.New(), uuid.New(), decimal.NewFromFloat(100.00), "USD", PaymentMethodCreditCard)

	metadata := map[string]string{
		"order_id": "ord_123",
		"vip":      "true",
	}
	payment.SetMetadata(metadata)

	assert.Equal(t, metadata, payment.Metadata)
}

func TestPaymentStatusTransitions(t *testing.T) {
	payment := NewPayment(uuid.New(), uuid.New(), uuid.New(), decimal.NewFromFloat(100.00), "USD", PaymentMethodCreditCard)

	assert.Equal(t, PaymentStatusPending, payment.Status)

	payment.MarkAsProcessing("proc", "txn")
	assert.Equal(t, PaymentStatusProcessing, payment.Status)

	processedAt := time.Now()
	payment.MarkAsCompleted(processedAt)
	assert.Equal(t, PaymentStatusCompleted, payment.Status)
}

func TestNewStripeProcessor(t *testing.T) {
	processor := NewStripeProcessor("sk_test_123", "whsec_123")

	assert.Equal(t, "sk_test_123", processor.apiKey)
	assert.Equal(t, "whsec_123", processor.webhookSecret)
	assert.Equal(t, "2023-10-16", processor.version)
}

func TestStripeProcessorProcessPayment(t *testing.T) {
	processor := NewStripeProcessor("sk_test_123", "whsec_123")

	req := &PaymentRequest{
		InvoiceID: uuid.New(),
		Amount:    decimal.NewFromFloat(100.00),
		Currency:  "USD",
		Method:    PaymentMethodCreditCard,
	}

	result, err := processor.ProcessPayment(nil, req)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, PaymentStatusCompleted, result.Status)
}

func TestStripeProcessorProcessRefund(t *testing.T) {
	processor := NewStripeProcessor("sk_test_123", "whsec_123")

	req := &RefundRequest{
		PaymentID: uuid.New(),
		Amount:    decimal.NewFromFloat(50.00),
		Reason:    "Customer request",
	}

	result, err := processor.ProcessRefund(nil, req)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, PaymentStatusRefunded, result.Status)
}

func TestStripeProcessorGetPaymentStatus(t *testing.T) {
	processor := NewStripeProcessor("sk_test_123", "whsec_123")

	result, err := processor.GetPaymentStatus(nil, "pi_123")

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, PaymentStatusCompleted, result.Status)
}

func TestNewPayPalProcessor(t *testing.T) {
	processor := NewPayPalProcessor("client_123", "secret_123", "sandbox")

	assert.Equal(t, "client_123", processor.clientID)
	assert.Equal(t, "secret_123", processor.clientSecret)
	assert.Equal(t, "sandbox", processor.mode)
}

func TestPayPalProcessorProcessPayment(t *testing.T) {
	processor := NewPayPalProcessor("client_123", "secret_123", "sandbox")

	req := &PaymentRequest{
		InvoiceID: uuid.New(),
		Amount:    decimal.NewFromFloat(100.00),
		Currency:  "USD",
		Method:    PaymentMethodPayPal,
	}

	result, err := processor.ProcessPayment(nil, req)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, PaymentStatusCompleted, result.Status)
}

func TestPayPalProcessorProcessRefund(t *testing.T) {
	processor := NewPayPalProcessor("client_123", "secret_123", "sandbox")

	req := &RefundRequest{
		PaymentID: uuid.New(),
		Amount:    decimal.NewFromFloat(50.00),
		Reason:    "Customer request",
	}

	result, err := processor.ProcessRefund(nil, req)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, PaymentStatusRefunded, result.Status)
}

func TestNewProcessorRegistry(t *testing.T) {
	registry := NewProcessorRegistry()

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.processors)
}

func TestProcessorRegistryRegister(t *testing.T) {
	registry := NewProcessorRegistry()

	var factory PaymentProcessorFactory = func(provider string, config interface{}) (PaymentProcessor, error) {
		return &StripeProcessor{apiKey: "test"}, nil
	}

	registry.Register("stripe", factory)

	assert.Contains(t, registry.processors, "stripe")
}

func TestProcessorRegistryGetProcessor(t *testing.T) {
	registry := NewProcessorRegistry()

	var factory PaymentProcessorFactory = func(provider string, config interface{}) (PaymentProcessor, error) {
		return &StripeProcessor{apiKey: "test"}, nil
	}
	registry.Register("stripe", factory)

	processor, err := registry.GetProcessor("stripe", nil)

	require.NoError(t, err)
	require.NotNil(t, processor)
}

func TestProcessorRegistryGetProcessorNotFound(t *testing.T) {
	registry := NewProcessorRegistry()

	_, err := registry.GetProcessor("unknown", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "payment processor not found")
}

func TestPaymentError(t *testing.T) {
	err := &PaymentError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
	}

	assert.Equal(t, "TEST_ERROR", err.Code)
	assert.Equal(t, "Test error message", err.Error())
}

func TestProcessorNotFoundError(t *testing.T) {
	err := &ProcessorNotFoundError{Name: "test"}

	assert.Equal(t, "test", err.Name)
	assert.Contains(t, err.Error(), "test")
}

func TestPaymentResult(t *testing.T) {
	result := &PaymentResult{
		Success:       true,
		PaymentID:     "pay_123",
		TransactionID: "txn_123",
		ProviderID:    "pi_123",
		Status:        PaymentStatusCompleted,
	}

	assert.True(t, result.Success)
	assert.Equal(t, "pay_123", result.PaymentID)
	assert.Equal(t, PaymentStatusCompleted, result.Status)
}

func TestRefundResult(t *testing.T) {
	result := &RefundResult{
		Success:       true,
		RefundID:      "ref_123",
		TransactionID: "txn_123",
		Status:        PaymentStatusRefunded,
	}

	assert.True(t, result.Success)
	assert.Equal(t, "ref_123", result.RefundID)
}
