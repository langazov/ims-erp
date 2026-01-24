package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
)

type PaymentMethod string

const (
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodDebitCard    PaymentMethod = "debit_card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodPayPal       PaymentMethod = "paypal"
	PaymentMethodStripe       PaymentMethod = "stripe"
	PaymentMethodCheck        PaymentMethod = "check"
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodWire         PaymentMethod = "wire"
	PaymentMethodCrypto       PaymentMethod = "crypto"
)

type Payment struct {
	ID             uuid.UUID         `json:"id" bson:"_id"`
	TenantID       uuid.UUID         `json:"tenantId" bson:"tenantId"`
	InvoiceID      uuid.UUID         `json:"invoiceId" bson:"invoiceId"`
	ClientID       uuid.UUID         `json:"clientId" bson:"clientId"`
	Amount         decimal.Decimal   `json:"amount" bson:"amount"`
	Currency       string            `json:"currency" bson:"currency"`
	Status         PaymentStatus     `json:"status" bson:"status"`
	Method         PaymentMethod     `json:"method" bson:"method"`
	Provider       string            `json:"provider" bson:"provider"`
	ProviderID     string            `json:"providerId" bson:"providerId"`
	TransactionID  string            `json:"transactionId" bson:"transactionId"`
	Reference      string            `json:"reference" bson:"reference"`
	Description    string            `json:"description" bson:"description"`
	Metadata       map[string]string `json:"metadata" bson:"metadata"`
	FailureCode    string            `json:"failureCode" bson:"failureCode"`
	FailureMessage string            `json:"failureMessage" bson:"failureMessage"`
	ProcessedAt    *time.Time        `json:"processedAt" bson:"processedAt"`
	CreatedAt      time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt" bson:"updatedAt"`
}

type PaymentRequest struct {
	InvoiceID   uuid.UUID         `json:"invoiceId"`
	Amount      decimal.Decimal   `json:"amount"`
	Currency    string            `json:"currency"`
	Method      PaymentMethod     `json:"method"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
}

type PaymentResult struct {
	Success       bool          `json:"success"`
	PaymentID     string        `json:"paymentId"`
	TransactionID string        `json:"transactionId"`
	ProviderID    string        `json:"providerId"`
	ErrorCode     string        `json:"errorCode"`
	ErrorMessage  string        `json:"errorMessage"`
	Status        PaymentStatus `json:"status"`
	ProcessedAt   *time.Time    `json:"processedAt"`
}

type RefundRequest struct {
	PaymentID  uuid.UUID       `json:"paymentId"`
	Amount     decimal.Decimal `json:"amount"`
	Reason     string          `json:"reason"`
	RefundType string          `json:"refundType"`
}

type RefundResult struct {
	Success       bool          `json:"success"`
	RefundID      string        `json:"refundId"`
	TransactionID string        `json:"transactionId"`
	ErrorCode     string        `json:"errorCode"`
	ErrorMessage  string        `json:"errorMessage"`
	Status        PaymentStatus `json:"status"`
}

type PaymentProcessor interface {
	ProcessPayment(ctx interface{}, req *PaymentRequest) (*PaymentResult, error)
	ProcessRefund(ctx interface{}, req *RefundRequest) (*RefundResult, error)
	GetPaymentStatus(ctx interface{}, providerID string) (*PaymentResult, error)
}

type PaymentProcessorFactory func(provider string, config interface{}) (PaymentProcessor, error)

type PaymentService struct {
	tenantID    uuid.UUID
	paymentRepo PaymentRepository
	processor   PaymentProcessor
	logger      interface{}
}

type PaymentRepository interface {
	Create(ctx interface{}, payment *Payment) error
	Update(ctx interface{}, payment *Payment) error
	FindByID(ctx interface{}, id uuid.UUID) (*Payment, error)
	FindByInvoiceID(ctx interface{}, invoiceID uuid.UUID) ([]*Payment, error)
	FindByProviderID(ctx interface{}, providerID string) (*Payment, error)
}

type PaymentGateway interface {
	Initialize(config interface{}) error
	ProcessPayment(req *PaymentRequest) (*PaymentResult, error)
	ProcessRefund(req *RefundRequest) (*RefundResult, error)
	GetPaymentStatus(providerID string) (*PaymentResult, error)
	GetSupportedMethods() []PaymentMethod
	IsMethodSupported(method PaymentMethod) bool
}

type StripeProcessor struct {
	apiKey        string
	webhookSecret string
	version       string
}

type PayPalProcessor struct {
	clientID     string
	clientSecret string
	mode         string
}

type BankTransferProcessor struct {
	bankCode      string
	accountNumber string
	routingNumber string
}

type GenericProcessor struct {
	name        string
	apiEndpoint string
	apiKey      string
}

func NewPayment(tenantID, invoiceID, clientID uuid.UUID, amount decimal.Decimal, currency string, method PaymentMethod) *Payment {
	now := time.Now().UTC()
	return &Payment{
		ID:        uuid.New(),
		TenantID:  tenantID,
		InvoiceID: invoiceID,
		ClientID:  clientID,
		Amount:    amount,
		Currency:  currency,
		Status:    PaymentStatusPending,
		Method:    method,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (p *Payment) MarkAsProcessing(providerID, transactionID string) {
	p.Status = PaymentStatusProcessing
	p.ProviderID = providerID
	p.TransactionID = transactionID
	p.UpdatedAt = time.Now().UTC()
}

func (p *Payment) MarkAsCompleted(processedAt time.Time) {
	p.Status = PaymentStatusCompleted
	p.ProcessedAt = &processedAt
	p.UpdatedAt = processedAt
}

func (p *Payment) MarkAsFailed(failureCode, failureMessage string) {
	p.Status = PaymentStatusFailed
	p.FailureCode = failureCode
	p.FailureMessage = failureMessage
	p.UpdatedAt = time.Now().UTC()
}

func (p *Payment) MarkAsRefunded() {
	p.Status = PaymentStatusRefunded
	p.UpdatedAt = time.Now().UTC()
}

func (p *Payment) SetReference(reference string) {
	p.Reference = reference
	p.UpdatedAt = time.Now().UTC()
}

func (p *Payment) SetMetadata(metadata map[string]string) {
	p.Metadata = metadata
	p.UpdatedAt = time.Now().UTC()
}

func NewStripeProcessor(apiKey, webhookSecret string) *StripeProcessor {
	return &StripeProcessor{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		version:       "2023-10-16",
	}
}

func (p *StripeProcessor) ProcessPayment(ctx interface{}, req *PaymentRequest) (*PaymentResult, error) {
	return &PaymentResult{
		Success:   true,
		PaymentID: uuid.New().String(),
		Status:    PaymentStatusCompleted,
	}, nil
}

func (p *StripeProcessor) ProcessRefund(ctx interface{}, req *RefundRequest) (*RefundResult, error) {
	return &RefundResult{
		Success:  true,
		RefundID: uuid.New().String(),
		Status:   PaymentStatusRefunded,
	}, nil
}

func (p *StripeProcessor) GetPaymentStatus(ctx interface{}, providerID string) (*PaymentResult, error) {
	return &PaymentResult{
		Success: true,
		Status:  PaymentStatusCompleted,
	}, nil
}

func NewPayPalProcessor(clientID, clientSecret, mode string) *PayPalProcessor {
	return &PayPalProcessor{
		clientID:     clientID,
		clientSecret: clientSecret,
		mode:         mode,
	}
}

func (p *PayPalProcessor) ProcessPayment(ctx interface{}, req *PaymentRequest) (*PaymentResult, error) {
	return &PaymentResult{
		Success:   true,
		PaymentID: uuid.New().String(),
		Status:    PaymentStatusCompleted,
	}, nil
}

func (p *PayPalProcessor) ProcessRefund(ctx interface{}, req *RefundRequest) (*RefundResult, error) {
	return &RefundResult{
		Success:  true,
		RefundID: uuid.New().String(),
		Status:   PaymentStatusRefunded,
	}, nil
}

func (p *PayPalProcessor) GetPaymentStatus(ctx interface{}, providerID string) (*PaymentResult, error) {
	return &PaymentResult{
		Success: true,
		Status:  PaymentStatusCompleted,
	}, nil
}

type ProcessorRegistry struct {
	processors map[string]PaymentProcessorFactory
}

func NewProcessorRegistry() *ProcessorRegistry {
	return &ProcessorRegistry{
		processors: make(map[string]PaymentProcessorFactory),
	}
}

func (r *ProcessorRegistry) Register(name string, factory PaymentProcessorFactory) {
	r.processors[name] = factory
}

func (r *ProcessorRegistry) GetProcessor(name string, config interface{}) (PaymentProcessor, error) {
	factory, ok := r.processors[name]
	if !ok {
		return nil, &ProcessorNotFoundError{Name: name}
	}
	return factory(name, config)
}

type ProcessorNotFoundError struct {
	Name string
}

func (e *ProcessorNotFoundError) Error() string {
	return "payment processor not found: " + e.Name
}

var ErrPaymentNotFound = &PaymentError{
	Code:    "PAYMENT_NOT_FOUND",
	Message: "Payment not found",
}

type PaymentError struct {
	Code    string
	Message string
}

func (e *PaymentError) Error() string {
	return e.Message
}
