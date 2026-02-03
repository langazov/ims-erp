package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ims-erp/system/internal/domain"
	eventpkg "github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/pkg/errors"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/shopspring/decimal"
)

// WebhookHandler processes incoming payment webhooks from external providers
type WebhookHandler struct {
	paymentRepo PaymentRepository
	invoiceRepo InvoiceRepository
	publisher   Publisher
	logger      *logger.Logger

	stripeWebhookSecret string
	paypalWebhookID     string
}

// StripeEvent represents a Stripe webhook event
type StripeEvent struct {
	ID              string              `json:"id"`
	Type            string              `json:"type"`
	APIVersion      string              `json:"api_version"`
	Created         int64               `json:"created"`
	Data            StripeEventData     `json:"data"`
	Livemode        bool                `json:"livemode"`
	PendingWebhooks int                 `json:"pending_webhooks"`
	Request         *StripeEventRequest `json:"request"`
}

// StripeEventData contains the event data
type StripeEventData struct {
	Object map[string]interface{} `json:"object"`
}

// StripeEventRequest contains request information
type StripeEventRequest struct {
	ID             string `json:"id"`
	IDempotencyKey string `json:"idempotency_key"`
}

// PayPalEvent represents a PayPal webhook event
type PayPalEvent struct {
	ID              string                 `json:"id"`
	EventType       string                 `json:"event_type"`
	ResourceType    string                 `json:"resource_type"`
	ResourceVersion string                 `json:"resource_version"`
	Summary         string                 `json:"summary"`
	Resource        map[string]interface{} `json:"resource"`
	Links           []PayPalLink           `json:"links"`
	CreateTime      string                 `json:"create_time"`
}

// PayPalLink represents a HATEOAS link
type PayPalLink struct {
	Href   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

// WebhookResult contains the result of processing a webhook
type WebhookResult struct {
	Success   bool
	EventID   string
	EventType string
	PaymentID string
	Error     error
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	paymentRepo PaymentRepository,
	invoiceRepo InvoiceRepository,
	publisher Publisher,
	log *logger.Logger,
	stripeSecret string,
	paypalID string,
) *WebhookHandler {
	return &WebhookHandler{
		paymentRepo:         paymentRepo,
		invoiceRepo:         invoiceRepo,
		publisher:           publisher,
		logger:              log,
		stripeWebhookSecret: stripeSecret,
		paypalWebhookID:     paypalID,
	}
}

// HandleStripeWebhook processes incoming Stripe webhook events
func (h *WebhookHandler) HandleStripeWebhook(ctx context.Context, payload []byte, signature string) (*WebhookResult, error) {
	log := h.logger.New(ctx)

	// Verify webhook signature
	if err := h.verifyStripeSignature(payload, signature); err != nil {
		log.Error("Stripe webhook signature verification failed", "error", err)
		return nil, errors.Newf(errors.CodeUnauthorized, "invalid webhook signature: %v", err)
	}

	// Parse the event
	var event StripeEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Error("Failed to parse Stripe webhook event", "error", err)
		return nil, errors.InvalidArgument("invalid webhook payload")
	}

	log.Info("Processing Stripe webhook",
		"event_id", event.ID,
		"event_type", event.Type,
	)

	result := &WebhookResult{
		EventID:   event.ID,
		EventType: event.Type,
	}

	// Process based on event type
	switch event.Type {
	case "payment_intent.succeeded":
		err := h.processStripePaymentIntentSucceeded(ctx, &event)
		if err != nil {
			result.Error = err
			return result, err
		}
		result.Success = true

	case "payment_intent.payment_failed":
		err := h.processStripePaymentIntentFailed(ctx, &event)
		if err != nil {
			result.Error = err
			return result, err
		}
		result.Success = true

	case "charge.refunded":
		err := h.processStripeChargeRefunded(ctx, &event)
		if err != nil {
			result.Error = err
			return result, err
		}
		result.Success = true

	default:
		log.Info("Unhandled Stripe event type", "event_type", event.Type)
		result.Success = true
	}

	return result, nil
}

// HandlePayPalWebhook processes incoming PayPal webhook events
func (h *WebhookHandler) HandlePayPalWebhook(ctx context.Context, payload []byte, headers map[string]string) (*WebhookResult, error) {
	log := h.logger.New(ctx)

	// Verify webhook
	if err := h.verifyPayPalWebhook(payload, headers); err != nil {
		log.Error("PayPal webhook verification failed", "error", err)
		return nil, errors.Newf(errors.CodeUnauthorized, "invalid webhook verification: %v", err)
	}

	// Parse the event
	var event PayPalEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Error("Failed to parse PayPal webhook event", "error", err)
		return nil, errors.InvalidArgument("invalid webhook payload")
	}

	log.Info("Processing PayPal webhook",
		"event_id", event.ID,
		"event_type", event.EventType,
	)

	result := &WebhookResult{
		EventID:   event.ID,
		EventType: event.EventType,
	}

	// Process based on event type
	switch event.EventType {
	case "PAYMENT.CAPTURE.COMPLETED":
		err := h.processPayPalCaptureCompleted(ctx, &event)
		if err != nil {
			result.Error = err
			return result, err
		}
		result.Success = true

	case "PAYMENT.CAPTURE.DENIED":
		err := h.processPayPalCaptureDenied(ctx, &event)
		if err != nil {
			result.Error = err
			return result, err
		}
		result.Success = true

	case "CUSTOMER.DISPUTE.CREATED":
		err := h.processPayPalDisputeCreated(ctx, &event)
		if err != nil {
			result.Error = err
			return result, err
		}
		result.Success = true

	default:
		log.Info("Unhandled PayPal event type", "event_type", event.EventType)
		result.Success = true
	}

	return result, nil
}

// verifyStripeSignature verifies the Stripe webhook signature
func (h *WebhookHandler) verifyStripeSignature(payload []byte, signature string) error {
	// TODO: Implement actual Stripe signature verification
	// This should use stripe-go library or implement the signature verification algorithm
	// Reference: https://stripe.com/docs/webhooks/signatures

	if h.stripeWebhookSecret == "" {
		return errors.New(errors.CodeInternalError, "stripe webhook secret not configured")
	}

	// Stub implementation - always returns nil for now
	// In production, implement proper signature verification
	return nil
}

// verifyPayPalWebhook verifies the PayPal webhook
func (h *WebhookHandler) verifyPayPalWebhook(payload []byte, headers map[string]string) error {
	// TODO: Implement actual PayPal webhook verification
	// This should verify the transmission ID, timestamp, and certificate ID
	// Reference: https://developer.paypal.com/api/rest/webhooks/

	if h.paypalWebhookID == "" {
		return errors.New(errors.CodeInternalError, "paypal webhook ID not configured")
	}

	// Stub implementation - always returns nil for now
	// In production, implement proper webhook verification
	return nil
}

// processStripePaymentIntentSucceeded handles payment_intent.succeeded events
func (h *WebhookHandler) processStripePaymentIntentSucceeded(ctx context.Context, event *StripeEvent) error {
	log := h.logger.New(ctx)

	object := event.Data.Object
	paymentIntentID, ok := object["id"].(string)
	if !ok {
		return errors.InvalidArgument("missing payment intent ID")
	}

	// Find payment by provider ID
	payment, err := h.paymentRepo.FindByProviderID(ctx, paymentIntentID)
	if err != nil {
		log.Error("Payment not found for Stripe payment intent",
			"payment_intent_id", paymentIntentID,
			"error", err,
		)
		return errors.Newf(errors.CodeNotFound, "payment not found for payment intent: %s", paymentIntentID)
	}

	// Extract charge information
	charges, ok := object["charges"].(map[string]interface{})
	if !ok {
		return errors.InternalError("invalid charges data in payment intent")
	}

	data, ok := charges["data"].([]interface{})
	if !ok || len(data) == 0 {
		return errors.InternalError("no charge data found in payment intent")
	}

	charge, ok := data[0].(map[string]interface{})
	if !ok {
		return errors.InternalError("invalid charge data structure")
	}

	transactionID, _ := charge["id"].(string)
	amountCaptured := decimal.Zero
	if amount, ok := charge["amount_captured"].(float64); ok {
		amountCaptured = decimal.NewFromFloat(amount).Div(decimal.NewFromInt(100))
	}

	// Mark payment as completed
	processedAt := time.Now().UTC()
	if created, ok := object["created"].(float64); ok {
		processedAt = time.Unix(int64(created), 0).UTC()
	}

	payment.MarkAsCompleted(processedAt)
	if transactionID != "" {
		payment.TransactionID = transactionID
	}

	// Update metadata
	if payment.Metadata == nil {
		payment.Metadata = make(map[string]string)
	}
	payment.Metadata["stripe_payment_intent_id"] = paymentIntentID
	payment.Metadata["webhook_event_id"] = event.ID

	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		log.Error("Failed to update payment status", "error", err)
		return errors.InternalError("failed to update payment status")
	}

	// Update invoice if applicable
	if err := h.updateInvoiceForPayment(ctx, payment, amountCaptured); err != nil {
		log.Error("Failed to update invoice for payment", "error", err)
		// Don't return error here - payment was successful, invoice update failure is logged
	}

	// Emit payment processed event
	eventData := map[string]interface{}{
		"invoiceId":      payment.InvoiceID.String(),
		"amount":         payment.Amount.String(),
		"transactionId":  payment.TransactionID,
		"providerId":     payment.ProviderID,
		"processedAt":    processedAt,
		"method":         string(payment.Method),
		"webhookEventId": event.ID,
		"stripeChargeId": transactionID,
	}

	ev := eventpkg.NewEvent(
		payment.ID.String(),
		"payment",
		"payment.processed",
		payment.TenantID.String(),
		"system", // Webhook events are system-generated
		eventData,
	)
	ev.WithMetadata("source", "stripe_webhook")
	ev.WithMetadata("webhook_event_id", event.ID)

	if err := h.publisher.PublishEvent(ctx, ev); err != nil {
		log.Error("Failed to publish payment processed event", "error", err)
	}

	log.Info("Payment completed via Stripe webhook",
		"payment_id", payment.ID,
		"payment_intent_id", paymentIntentID,
		"transaction_id", transactionID,
		"amount", payment.Amount.String(),
	)

	return nil
}

// processStripePaymentIntentFailed handles payment_intent.payment_failed events
func (h *WebhookHandler) processStripePaymentIntentFailed(ctx context.Context, event *StripeEvent) error {
	log := h.logger.New(ctx)

	object := event.Data.Object
	paymentIntentID, ok := object["id"].(string)
	if !ok {
		return errors.InvalidArgument("missing payment intent ID")
	}

	// Find payment by provider ID
	payment, err := h.paymentRepo.FindByProviderID(ctx, paymentIntentID)
	if err != nil {
		log.Error("Payment not found for failed Stripe payment intent",
			"payment_intent_id", paymentIntentID,
			"error", err,
		)
		return errors.Newf(errors.CodeNotFound, "payment not found for payment intent: %s", paymentIntentID)
	}

	// Extract error information
	lastPaymentError, ok := object["last_payment_error"].(map[string]interface{})
	if !ok {
		lastPaymentError = make(map[string]interface{})
	}

	failureCode, _ := lastPaymentError["code"].(string)
	failureMessage, _ := lastPaymentError["message"].(string)
	declineCode, _ := lastPaymentError["decline_code"].(string)

	if failureCode == "" {
		failureCode = "payment_failed"
	}
	if failureMessage == "" {
		failureMessage = "Payment processing failed"
	}

	// Mark payment as failed
	payment.MarkAsFailed(failureCode, failureMessage)

	// Update metadata
	if payment.Metadata == nil {
		payment.Metadata = make(map[string]string)
	}
	payment.Metadata["stripe_payment_intent_id"] = paymentIntentID
	payment.Metadata["webhook_event_id"] = event.ID
	if declineCode != "" {
		payment.Metadata["stripe_decline_code"] = declineCode
	}

	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		log.Error("Failed to update payment failure status", "error", err)
		return errors.InternalError("failed to update payment status")
	}

	// Emit payment failed event
	eventData := map[string]interface{}{
		"invoiceId":      payment.InvoiceID.String(),
		"amount":         payment.Amount.String(),
		"failureCode":    payment.FailureCode,
		"failureMessage": payment.FailureMessage,
		"webhookEventId": event.ID,
	}
	if declineCode != "" {
		eventData["stripeDeclineCode"] = declineCode
	}

	ev := eventpkg.NewEvent(
		payment.ID.String(),
		"payment",
		"payment.failed",
		payment.TenantID.String(),
		"system",
		eventData,
	)
	ev.WithMetadata("source", "stripe_webhook")
	ev.WithMetadata("webhook_event_id", event.ID)

	if err := h.publisher.PublishEvent(ctx, ev); err != nil {
		log.Error("Failed to publish payment failed event", "error", err)
	}

	log.Info("Payment failed via Stripe webhook",
		"payment_id", payment.ID,
		"payment_intent_id", paymentIntentID,
		"failure_code", failureCode,
		"failure_message", failureMessage,
	)

	return nil
}

// processStripeChargeRefunded handles charge.refunded events
func (h *WebhookHandler) processStripeChargeRefunded(ctx context.Context, event *StripeEvent) error {
	log := h.logger.New(ctx)

	object := event.Data.Object
	chargeID, ok := object["id"].(string)
	if !ok {
		return errors.InvalidArgument("missing charge ID")
	}

	// Find payment by transaction ID (charge ID)
	// Note: This might need adjustment based on how transaction IDs are stored
	payment, err := h.paymentRepo.FindByProviderID(ctx, chargeID)
	if err != nil {
		log.Error("Payment not found for refunded charge",
			"charge_id", chargeID,
			"error", err,
		)
		return errors.Newf(errors.CodeNotFound, "payment not found for charge: %s", chargeID)
	}

	// Check if payment is in a refundable state
	if payment.Status != domain.PaymentStatusCompleted {
		log.Warn("Cannot refund payment - not in completed state",
			"payment_id", payment.ID,
			"status", payment.Status,
		)
		return errors.Newf(errors.CodeInvalidArgument, "cannot refund payment with status: %s", payment.Status)
	}

	// Extract refund information
	refunds, ok := object["refunds"].(map[string]interface{})
	if !ok {
		return errors.InternalError("invalid refunds data in charge")
	}

	data, ok := refunds["data"].([]interface{})
	if !ok || len(data) == 0 {
		return errors.InternalError("no refund data found in charge")
	}

	refund, ok := data[0].(map[string]interface{})
	if !ok {
		return errors.InternalError("invalid refund data structure")
	}

	refundID, _ := refund["id"].(string)
	refundAmount := decimal.Zero
	if amount, ok := refund["amount"].(float64); ok {
		refundAmount = decimal.NewFromFloat(amount).Div(decimal.NewFromInt(100))
	}
	reason, _ := refund["reason"].(string)
	if reason == "" {
		reason = "Stripe refund"
	}

	// Mark payment as refunded
	payment.MarkAsRefunded()

	// Update metadata
	if payment.Metadata == nil {
		payment.Metadata = make(map[string]string)
	}
	payment.Metadata["stripe_refund_id"] = refundID
	payment.Metadata["stripe_charge_id"] = chargeID
	payment.Metadata["webhook_event_id"] = event.ID
	payment.Metadata["refund_reason"] = reason
	payment.Metadata["refund_amount"] = refundAmount.String()

	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		log.Error("Failed to update payment refund status", "error", err)
		return errors.InternalError("failed to update payment refund status")
	}

	// Emit payment refunded event
	ev := eventpkg.NewEvent(
		payment.ID.String(),
		"payment",
		"payment.refunded",
		payment.TenantID.String(),
		"system",
		map[string]interface{}{
			"invoiceId":      payment.InvoiceID.String(),
			"originalAmount": payment.Amount.String(),
			"refundAmount":   refundAmount.String(),
			"reason":         reason,
			"refundId":       refundID,
			"transactionId":  chargeID,
			"webhookEventId": event.ID,
		},
	)
	ev.WithMetadata("source", "stripe_webhook")
	ev.WithMetadata("webhook_event_id", event.ID)

	if err := h.publisher.PublishEvent(ctx, ev); err != nil {
		log.Error("Failed to publish payment refunded event", "error", err)
	}

	log.Info("Payment refunded via Stripe webhook",
		"payment_id", payment.ID,
		"charge_id", chargeID,
		"refund_id", refundID,
		"refund_amount", refundAmount.String(),
	)

	return nil
}

// processPayPalCaptureCompleted handles PAYMENT.CAPTURE.COMPLETED events
func (h *WebhookHandler) processPayPalCaptureCompleted(ctx context.Context, event *PayPalEvent) error {
	log := h.logger.New(ctx)

	resource := event.Resource
	captureID, ok := resource["id"].(string)
	if !ok {
		return errors.InvalidArgument("missing capture ID")
	}

	// Find payment by provider ID (using capture ID)
	payment, err := h.paymentRepo.FindByProviderID(ctx, captureID)
	if err != nil {
		log.Error("Payment not found for PayPal capture",
			"capture_id", captureID,
			"error", err,
		)
		return errors.Newf(errors.CodeNotFound, "payment not found for capture: %s", captureID)
	}

	// Extract payment information
	amount := decimal.Zero
	if amt, ok := resource["amount"].(map[string]interface{}); ok {
		if value, ok := amt["value"].(string); ok {
			amount, _ = decimal.NewFromString(value)
		}
	}

	transactionID := captureID
	if customID, ok := resource["custom_id"].(string); ok && customID != "" {
		transactionID = customID
	}

	// Mark payment as completed
	processedAt := time.Now().UTC()
	if createTime, ok := resource["create_time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, createTime); err == nil {
			processedAt = parsed
		}
	}

	payment.MarkAsCompleted(processedAt)
	payment.TransactionID = transactionID

	// Update metadata
	if payment.Metadata == nil {
		payment.Metadata = make(map[string]string)
	}
	payment.Metadata["paypal_capture_id"] = captureID
	payment.Metadata["webhook_event_id"] = event.ID

	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		log.Error("Failed to update payment status", "error", err)
		return errors.InternalError("failed to update payment status")
	}

	// Update invoice if applicable
	if err := h.updateInvoiceForPayment(ctx, payment, amount); err != nil {
		log.Error("Failed to update invoice for payment", "error", err)
		// Don't return error here - payment was successful
	}

	// Emit payment processed event
	ev := eventpkg.NewEvent(
		payment.ID.String(),
		"payment",
		"payment.processed",
		payment.TenantID.String(),
		"system",
		map[string]interface{}{
			"invoiceId":       payment.InvoiceID.String(),
			"amount":          payment.Amount.String(),
			"transactionId":   payment.TransactionID,
			"providerId":      payment.ProviderID,
			"processedAt":     processedAt,
			"method":          string(payment.Method),
			"webhookEventId":  event.ID,
			"paypalCaptureId": captureID,
		},
	)
	ev.WithMetadata("source", "paypal_webhook")
	ev.WithMetadata("webhook_event_id", event.ID)

	if err := h.publisher.PublishEvent(ctx, ev); err != nil {
		log.Error("Failed to publish payment processed event", "error", err)
	}

	log.Info("Payment completed via PayPal webhook",
		"payment_id", payment.ID,
		"capture_id", captureID,
		"transaction_id", transactionID,
		"amount", payment.Amount.String(),
	)

	return nil
}

// processPayPalCaptureDenied handles PAYMENT.CAPTURE.DENIED events
func (h *WebhookHandler) processPayPalCaptureDenied(ctx context.Context, event *PayPalEvent) error {
	log := h.logger.New(ctx)

	resource := event.Resource
	captureID, ok := resource["id"].(string)
	if !ok {
		return errors.InvalidArgument("missing capture ID")
	}

	// Find payment by provider ID
	payment, err := h.paymentRepo.FindByProviderID(ctx, captureID)
	if err != nil {
		log.Error("Payment not found for denied PayPal capture",
			"capture_id", captureID,
			"error", err,
		)
		return errors.Newf(errors.CodeNotFound, "payment not found for capture: %s", captureID)
	}

	// Extract failure reason if available
	failureCode := "capture_denied"
	failureMessage := "Payment capture was denied"

	if status, ok := resource["status"].(string); ok && status != "" {
		failureCode = status
	}
	if statusDetails, ok := resource["status_details"].(map[string]interface{}); ok {
		if reason, ok := statusDetails["reason"].(string); ok {
			failureMessage = reason
		}
	}

	// Mark payment as failed
	payment.MarkAsFailed(failureCode, failureMessage)

	// Update metadata
	if payment.Metadata == nil {
		payment.Metadata = make(map[string]string)
	}
	payment.Metadata["paypal_capture_id"] = captureID
	payment.Metadata["webhook_event_id"] = event.ID

	if err := h.paymentRepo.Update(ctx, payment); err != nil {
		log.Error("Failed to update payment failure status", "error", err)
		return errors.InternalError("failed to update payment status")
	}

	// Emit payment failed event
	ev := eventpkg.NewEvent(
		payment.ID.String(),
		"payment",
		"payment.failed",
		payment.TenantID.String(),
		"system",
		map[string]interface{}{
			"invoiceId":       payment.InvoiceID.String(),
			"amount":          payment.Amount.String(),
			"failureCode":     payment.FailureCode,
			"failureMessage":  payment.FailureMessage,
			"webhookEventId":  event.ID,
			"paypalCaptureId": captureID,
		},
	)
	ev.WithMetadata("source", "paypal_webhook")
	ev.WithMetadata("webhook_event_id", event.ID)

	if err := h.publisher.PublishEvent(ctx, ev); err != nil {
		log.Error("Failed to publish payment failed event", "error", err)
	}

	log.Info("Payment denied via PayPal webhook",
		"payment_id", payment.ID,
		"capture_id", captureID,
		"failure_code", failureCode,
		"failure_message", failureMessage,
	)

	return nil
}

// processPayPalDisputeCreated handles CUSTOMER.DISPUTE.CREATED events
func (h *WebhookHandler) processPayPalDisputeCreated(ctx context.Context, event *PayPalEvent) error {
	log := h.logger.New(ctx)

	resource := event.Resource
	disputeID, ok := resource["dispute_id"].(string)
	if !ok {
		disputeID, _ = resource["id"].(string)
	}

	// Extract transaction IDs from the dispute
	transactionIDs := []string{}
	if transactions, ok := resource["transactions"].([]interface{}); ok {
		for _, tx := range transactions {
			if txMap, ok := tx.(map[string]interface{}); ok {
				if txID, ok := txMap["transaction_id"].(string); ok {
					transactionIDs = append(transactionIDs, txID)
				}
			}
		}
	}

	if len(transactionIDs) == 0 {
		log.Warn("No transaction IDs found in dispute", "dispute_id", disputeID)
		return nil
	}

	reason, _ := resource["reason"].(string)
	status, _ := resource["status"].(string)
	amount := decimal.Zero
	if disputedAmount, ok := resource["disputed_amount"].(map[string]interface{}); ok {
		if value, ok := disputedAmount["value"].(string); ok {
			amount, _ = decimal.NewFromString(value)
		}
	}

	// Find and flag payments associated with the disputed transactions
	for _, txID := range transactionIDs {
		payment, err := h.paymentRepo.FindByProviderID(ctx, txID)
		if err != nil {
			log.Warn("Payment not found for disputed transaction",
				"transaction_id", txID,
				"dispute_id", disputeID,
			)
			continue
		}

		// Update payment metadata to indicate dispute
		if payment.Metadata == nil {
			payment.Metadata = make(map[string]string)
		}
		payment.Metadata["paypal_dispute_id"] = disputeID
		payment.Metadata["dispute_status"] = status
		payment.Metadata["dispute_reason"] = reason
		payment.UpdatedAt = time.Now().UTC()

		if err := h.paymentRepo.Update(ctx, payment); err != nil {
			log.Error("Failed to update payment with dispute info",
				"payment_id", payment.ID,
				"error", err,
			)
			continue
		}

		// Emit dispute created event
		ev := eventpkg.NewEvent(
			payment.ID.String(),
			"payment",
			"payment.dispute.created",
			payment.TenantID.String(),
			"system",
			map[string]interface{}{
				"invoiceId":      payment.InvoiceID.String(),
				"amount":         payment.Amount.String(),
				"disputeId":      disputeID,
				"disputeStatus":  status,
				"disputeReason":  reason,
				"disputedAmount": amount.String(),
				"transactionId":  txID,
				"webhookEventId": event.ID,
			},
		)
		ev.WithMetadata("source", "paypal_webhook")
		ev.WithMetadata("webhook_event_id", event.ID)
		ev.WithMetadata("dispute_id", disputeID)

		if err := h.publisher.PublishEvent(ctx, ev); err != nil {
			log.Error("Failed to publish dispute event", "error", err)
		}

		log.Info("Payment dispute flagged",
			"payment_id", payment.ID,
			"dispute_id", disputeID,
			"reason", reason,
		)
	}

	return nil
}

// updateInvoiceForPayment updates the invoice status when a payment is received
func (h *WebhookHandler) updateInvoiceForPayment(ctx context.Context, payment *domain.Payment, amount decimal.Decimal) error {
	log := h.logger.New(ctx)

	invoice, err := h.invoiceRepo.FindByID(ctx, payment.InvoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Verify invoice belongs to same tenant
	if invoice.TenantID != payment.TenantID {
		return errors.Newf(errors.CodeForbidden, "invoice does not belong to payment tenant")
	}

	// Apply payment to invoice
	if err := invoice.ApplyPayment(amount); err != nil {
		return fmt.Errorf("failed to apply payment to invoice: %w", err)
	}

	if err := h.invoiceRepo.Update(ctx, invoice); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	log.Info("Invoice updated for payment",
		"invoice_id", invoice.ID,
		"payment_id", payment.ID,
		"amount_paid", amount.String(),
		"amount_due", invoice.AmountDue.String(),
		"status", invoice.Status,
	)

	return nil
}

// ParseStripeWebhook parses the Stripe webhook from HTTP request
func (h *WebhookHandler) ParseStripeWebhook(r *http.Request) ([]byte, string, error) {
	// Read request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	// Get Stripe signature from header
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		return nil, "", errors.New(errors.CodeUnauthorized, "missing Stripe-Signature header")
	}

	return payload, signature, nil
}

// ParsePayPalWebhook parses the PayPal webhook from HTTP request
func (h *WebhookHandler) ParsePayPalWebhook(r *http.Request) ([]byte, map[string]string, error) {
	// Read request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	// Extract relevant headers for verification
	headers := map[string]string{
		"Content-Type":             r.Header.Get("Content-Type"),
		"PayPal-Transmission-Id":   r.Header.Get("PayPal-Transmission-Id"),
		"PayPal-Cert-Url":          r.Header.Get("PayPal-Cert-Url"),
		"PayPal-Auth-Algo":         r.Header.Get("PayPal-Auth-Algo"),
		"PayPal-Transmission-Time": r.Header.Get("PayPal-Transmission-Time"),
		"PayPal-Transmission-Sig":  r.Header.Get("PayPal-Transmission-Sig"),
	}

	return payload, headers, nil
}

// Helper method to get string value from map
func getWebhookString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

// Helper method to get decimal value from map
func getWebhookDecimal(data map[string]interface{}, key string) decimal.Decimal {
	if val, ok := data[key].(float64); ok {
		return decimal.NewFromFloat(val)
	}
	if val, ok := data[key].(string); ok {
		d, _ := decimal.NewFromString(val)
		return d
	}
	return decimal.Zero
}
