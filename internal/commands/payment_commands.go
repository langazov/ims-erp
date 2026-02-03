package commands

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	eventpkg "github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/errors"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/shopspring/decimal"
)

type PaymentCommandHandler struct {
	paymentRepo PaymentRepository
	invoiceRepo InvoiceRepository
	eventStore  *repository.EventStore
	publisher   Publisher
	logger      *logger.Logger
	processors  *domain.ProcessorRegistry
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	Update(ctx context.Context, payment *domain.Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	FindByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]*domain.Payment, error)
	FindByProviderID(ctx context.Context, providerID string) (*domain.Payment, error)
}

func NewPaymentCommandHandler(
	paymentRepo PaymentRepository,
	invoiceRepo InvoiceRepository,
	eventStore *repository.EventStore,
	publisher Publisher,
	log *logger.Logger,
	processors *domain.ProcessorRegistry,
) *PaymentCommandHandler {
	return &PaymentCommandHandler{
		paymentRepo: paymentRepo,
		invoiceRepo: invoiceRepo,
		eventStore:  eventStore,
		publisher:   publisher,
		logger:      log,
		processors:  processors,
	}
}

func (h *PaymentCommandHandler) HandleCreatePayment(ctx context.Context, cmd *CommandEnvelope) (*domain.Payment, error) {
	data := cmd.Data

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	invoiceID, err := uuid.Parse(getString(data, "invoiceId"))
	if err != nil {
		return nil, errors.InvalidArgument("invalid invoice ID")
	}

	clientID, err := uuid.Parse(getString(data, "clientId"))
	if err != nil {
		return nil, errors.InvalidArgument("invalid client ID")
	}

	amount := decimal.Zero
	if amountStr, ok := data["amount"].(string); ok {
		amount, _ = decimal.NewFromString(amountStr)
	}

	if amount.IsZero() || amount.IsNegative() {
		return nil, errors.InvalidArgument("payment amount must be greater than zero")
	}

	currency := getString(data, "currency")
	if currency == "" {
		currency = "USD"
	}

	method := domain.PaymentMethod(getString(data, "method"))
	if method == "" {
		method = domain.PaymentMethodCreditCard
	}

	payment := domain.NewPayment(
		tenantID,
		invoiceID,
		clientID,
		amount,
		currency,
		method,
	)

	if provider, ok := data["provider"].(string); ok {
		payment.Provider = provider
	}

	if reference, ok := data["reference"].(string); ok {
		payment.SetReference(reference)
	}

	if description, ok := data["description"].(string); ok {
		payment.Description = description
	}

	if err := h.paymentRepo.Create(ctx, payment); err != nil {
		h.logger.New(ctx).Error("Failed to create payment", "error", err)
		return nil, errors.InternalError("failed to create payment")
	}

	event := eventpkg.NewEvent(
		payment.ID.String(),
		"payment",
		"payment.created",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"invoiceId":   payment.InvoiceID.String(),
			"clientId":    payment.ClientID.String(),
			"amount":      payment.Amount.String(),
			"currency":    payment.Currency,
			"method":      string(payment.Method),
			"provider":    payment.Provider,
			"reference":   payment.Reference,
			"status":      string(payment.Status),
			"description": payment.Description,
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish payment created event", "error", err)
	}

	h.logger.New(ctx).Info("Payment created",
		"payment_id", payment.ID,
		"invoice_id", payment.InvoiceID,
		"amount", payment.Amount.String(),
		"method", payment.Method,
	)

	return payment, nil
}

func (h *PaymentCommandHandler) HandleProcessPayment(ctx context.Context, cmd *CommandEnvelope) (*domain.Payment, error) {
	paymentID, err := uuid.Parse(cmd.TargetID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid payment ID")
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, errors.NotFound("payment not found")
	}

	if payment.TenantID != tenantID {
		return nil, errors.Newf(errors.CodeForbidden, "payment does not belong to tenant")
	}

	if payment.Status != domain.PaymentStatusPending {
		return nil, errors.InvalidArgument("payment is not in pending status")
	}

	processor, err := h.processors.GetProcessor(payment.Provider, nil)
	if err != nil {
		h.logger.New(ctx).Error("Payment processor not found", "provider", payment.Provider, "error", err)
		return nil, errors.InvalidArgument("payment processor not available")
	}

	payment.MarkAsProcessing("provider_"+payment.ID.String(), "tx_"+payment.ID.String())

	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		h.logger.New(ctx).Error("Failed to update payment status to processing", "error", err)
		return nil, errors.InternalError("failed to process payment")
	}

	req := &domain.PaymentRequest{
		InvoiceID:   payment.InvoiceID,
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		Method:      payment.Method,
		Description: payment.Description,
		Metadata:    payment.Metadata,
	}

	result, err := processor.ProcessPayment(ctx, req)

	if err != nil || !result.Success {
		failureCode := "PROCESSING_ERROR"
		failureMessage := "Payment processing failed"
		if err != nil {
			failureMessage = err.Error()
		}
		if result != nil && result.ErrorCode != "" {
			failureCode = result.ErrorCode
			failureMessage = result.ErrorMessage
		}

		payment.MarkAsFailed(failureCode, failureMessage)

		if updateErr := h.paymentRepo.Update(ctx, payment); updateErr != nil {
			h.logger.New(ctx).Error("Failed to update payment failure status", "error", updateErr)
		}

		event := eventpkg.NewEvent(
			payment.ID.String(),
			"payment",
			"payment.failed",
			cmd.TenantID,
			cmd.UserID,
			map[string]interface{}{
				"invoiceId":      payment.InvoiceID.String(),
				"amount":         payment.Amount.String(),
				"failureCode":    payment.FailureCode,
				"failureMessage": payment.FailureMessage,
			},
		)
		event.WithCorrelationID(cmd.CorrelationID)
		h.publisher.PublishEvent(ctx, event)

		h.logger.New(ctx).Error("Payment processing failed",
			"payment_id", payment.ID,
			"failure_code", failureCode,
			"failure_message", failureMessage,
		)

		return payment, errors.Newf(errors.CodeInternalError, "payment processing failed: %s", failureMessage)
	}

	processedAt := time.Now().UTC()
	if result.ProcessedAt != nil {
		processedAt = *result.ProcessedAt
	}

	payment.MarkAsCompleted(processedAt)
	if result.TransactionID != "" {
		payment.TransactionID = result.TransactionID
	}
	if result.ProviderID != "" {
		payment.ProviderID = result.ProviderID
	}

	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		h.logger.New(ctx).Error("Failed to update payment completion status", "error", err)
		return nil, errors.InternalError("failed to complete payment")
	}

	invoice, err := h.invoiceRepo.FindByID(ctx, payment.InvoiceID)
	if err == nil && invoice != nil {
		invoice.MarkAsPaid(payment.Amount)
		if updateErr := h.invoiceRepo.Update(ctx, invoice); updateErr != nil {
			h.logger.New(ctx).Error("Failed to update invoice payment status", "error", updateErr)
		}
	}

	event := eventpkg.NewEvent(
		payment.ID.String(),
		"payment",
		"payment.processed",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"invoiceId":     payment.InvoiceID.String(),
			"amount":        payment.Amount.String(),
			"transactionId": payment.TransactionID,
			"providerId":    payment.ProviderID,
			"processedAt":   processedAt,
			"method":        string(payment.Method),
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish payment processed event", "error", err)
	}

	h.logger.New(ctx).Info("Payment processed successfully",
		"payment_id", payment.ID,
		"invoice_id", payment.InvoiceID,
		"amount", payment.Amount.String(),
		"transaction_id", payment.TransactionID,
	)

	return payment, nil
}

func (h *PaymentCommandHandler) HandleRefundPayment(ctx context.Context, cmd *CommandEnvelope) (*domain.Payment, error) {
	paymentID, err := uuid.Parse(cmd.TargetID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid payment ID")
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, errors.NotFound("payment not found")
	}

	if payment.TenantID != tenantID {
		return nil, errors.Newf(errors.CodeForbidden, "payment does not belong to tenant")
	}

	if payment.Status != domain.PaymentStatusCompleted {
		return nil, errors.InvalidArgument("can only refund completed payments")
	}

	amount := payment.Amount
	if amountStr, ok := cmd.Data["amount"].(string); ok {
		refundAmount, _ := decimal.NewFromString(amountStr)
		if !refundAmount.IsZero() && refundAmount.IsPositive() && refundAmount.LessThanOrEqual(payment.Amount) {
			amount = refundAmount
		}
	}

	reason := getString(cmd.Data, "reason")
	if reason == "" {
		reason = "Customer requested refund"
	}

	processor, err := h.processors.GetProcessor(payment.Provider, nil)
	if err != nil {
		h.logger.New(ctx).Error("Payment processor not found for refund", "provider", payment.Provider, "error", err)
		return nil, errors.InvalidArgument("payment processor not available")
	}

	refundReq := &domain.RefundRequest{
		PaymentID:  paymentID,
		Amount:     amount,
		Reason:     reason,
		RefundType: "full",
	}

	if amount.LessThan(payment.Amount) {
		refundReq.RefundType = "partial"
	}

	result, err := processor.ProcessRefund(ctx, refundReq)

	if err != nil || !result.Success {
		h.logger.New(ctx).Error("Refund processing failed",
			"payment_id", payment.ID,
			"error", err,
		)
		return nil, errors.Newf(errors.CodeInternalError, "refund processing failed: %v", err)
	}

	payment.MarkAsRefunded()

	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		h.logger.New(ctx).Error("Failed to update payment refund status", "error", err)
		return nil, errors.InternalError("failed to process refund")
	}

	event := eventpkg.NewEvent(
		payment.ID.String(),
		"payment",
		"payment.refunded",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"invoiceId":      payment.InvoiceID.String(),
			"originalAmount": payment.Amount.String(),
			"refundAmount":   amount.String(),
			"reason":         reason,
			"refundId":       result.RefundID,
			"transactionId":  result.TransactionID,
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish payment refunded event", "error", err)
	}

	h.logger.New(ctx).Info("Payment refunded",
		"payment_id", payment.ID,
		"invoice_id", payment.InvoiceID,
		"amount", amount.String(),
		"reason", reason,
	)

	return payment, nil
}

func (h *PaymentCommandHandler) HandleCancelPayment(ctx context.Context, cmd *CommandEnvelope) (*domain.Payment, error) {
	paymentID, err := uuid.Parse(cmd.TargetID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid payment ID")
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, errors.NotFound("payment not found")
	}

	if payment.TenantID != tenantID {
		return nil, errors.Newf(errors.CodeForbidden, "payment does not belong to tenant")
	}

	if payment.Status == domain.PaymentStatusCompleted {
		return nil, errors.InvalidArgument("cannot cancel completed payments, use refund instead")
	}

	if payment.Status == domain.PaymentStatusRefunded {
		return nil, errors.InvalidArgument("payment is already refunded")
	}

	if payment.Status == domain.PaymentStatusCancelled {
		return nil, errors.InvalidArgument("payment is already cancelled")
	}

	payment.Status = domain.PaymentStatusCancelled
	payment.UpdatedAt = time.Now().UTC()

	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		h.logger.New(ctx).Error("Failed to cancel payment", "error", err)
		return nil, errors.InternalError("failed to cancel payment")
	}

	previousStatus := string(domain.PaymentStatusPending)
	if payment.Status == domain.PaymentStatusProcessing {
		previousStatus = string(domain.PaymentStatusProcessing)
	}

	event := eventpkg.NewEvent(
		payment.ID.String(),
		"payment",
		"payment.cancelled",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"invoiceId":      payment.InvoiceID.String(),
			"amount":         payment.Amount.String(),
			"previousStatus": previousStatus,
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish payment cancelled event", "error", err)
	}

	h.logger.New(ctx).Info("Payment cancelled",
		"payment_id", payment.ID,
		"invoice_id", payment.InvoiceID,
	)

	return payment, nil
}
