package events

import (
	"context"
	"time"

	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type PaymentEventHandler struct {
	readModelStore *repository.ReadModelStore
	cache          *repository.Cache
	logger         *logger.Logger
	tracer         trace.Tracer
}

func NewPaymentEventHandler(
	readModelStore *repository.ReadModelStore,
	cache *repository.Cache,
	log *logger.Logger,
) *PaymentEventHandler {
	return &PaymentEventHandler{
		readModelStore: readModelStore,
		cache:          cache,
		logger:         log,
		tracer:         otel.Tracer("payment-event-handler"),
	}
}

func (h *PaymentEventHandler) HandlePaymentCreated(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_payment_created",
		trace.WithAttributes(
			attribute.String("payment_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	paymentSummary := PaymentSummary{
		ID:          event.AggregateID,
		TenantID:    event.TenantID,
		InvoiceID:   getString(event.Data, "invoiceId"),
		ClientID:    getString(event.Data, "clientId"),
		Amount:      getString(event.Data, "amount"),
		Currency:    getString(event.Data, "currency"),
		Status:      string(domain.PaymentStatusPending),
		Method:      getString(event.Data, "method"),
		Provider:    getString(event.Data, "provider"),
		Reference:   getString(event.Data, "reference"),
		Description: getString(event.Data, "description"),
		CreatedAt:   event.Timestamp,
		UpdatedAt:   event.Timestamp,
	}

	if err := h.readModelStore.Save(ctx, paymentSummary); err != nil {
		span.RecordError(err)
		return err
	}

	paymentDetail := PaymentDetail{
		ID:          event.AggregateID,
		TenantID:    event.TenantID,
		InvoiceID:   getString(event.Data, "invoiceId"),
		ClientID:    getString(event.Data, "clientId"),
		Amount:      getString(event.Data, "amount"),
		Currency:    getString(event.Data, "currency"),
		Status:      string(domain.PaymentStatusPending),
		Method:      getString(event.Data, "method"),
		Provider:    getString(event.Data, "provider"),
		Reference:   getString(event.Data, "reference"),
		Description: getString(event.Data, "description"),
		Metadata:    getMap(event.Data, "metadata"),
		ActivityLog: []PaymentActivity{
			{
				Action:    "created",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
			},
		},
		CreatedAt: event.Timestamp,
		UpdatedAt: event.Timestamp,
	}

	if err := h.readModelStore.Save(ctx, paymentDetail); err != nil {
		span.RecordError(err)
		return err
	}

	h.logger.New(ctx).Info("Payment created in read model",
		"payment_id", event.AggregateID,
		"tenant_id", event.TenantID,
		"invoice_id", paymentSummary.InvoiceID,
		"amount", paymentSummary.Amount,
	)

	return nil
}

func (h *PaymentEventHandler) HandlePaymentProcessed(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_payment_processed",
		trace.WithAttributes(
			attribute.String("payment_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	processedAt := getTime(event.Data, "processedAt")
	if processedAt.IsZero() {
		processedAt = event.Timestamp
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":        string(domain.PaymentStatusCompleted),
			"transactionId": getString(event.Data, "transactionId"),
			"providerId":    getString(event.Data, "providerId"),
			"processedAt":   processedAt,
			"updatedAt":     event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": PaymentActivity{
				Action:    "processed",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
				Details:   "Payment processed successfully",
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "payment:detail:"+event.AggregateID)
	h.cache.Delete(ctx, "payment:summary:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "payment:list:*")

	h.logger.New(ctx).Info("Payment processed in read model",
		"payment_id", event.AggregateID,
		"tenant_id", event.TenantID,
		"transaction_id", getString(event.Data, "transactionId"),
	)

	return nil
}

func (h *PaymentEventHandler) HandlePaymentFailed(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_payment_failed",
		trace.WithAttributes(
			attribute.String("payment_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	failureCode := getString(event.Data, "failureCode")
	failureMessage := getString(event.Data, "failureMessage")

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":         string(domain.PaymentStatusFailed),
			"failureCode":    failureCode,
			"failureMessage": failureMessage,
			"updatedAt":      event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": PaymentActivity{
				Action:    "failed",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
				Details:   failureMessage,
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "payment:detail:"+event.AggregateID)
	h.cache.Delete(ctx, "payment:summary:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "payment:list:*")

	h.logger.New(ctx).Error("Payment failed in read model",
		"payment_id", event.AggregateID,
		"tenant_id", event.TenantID,
		"failure_code", failureCode,
		"failure_message", failureMessage,
	)

	return nil
}

func (h *PaymentEventHandler) HandlePaymentRefunded(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_payment_refunded",
		trace.WithAttributes(
			attribute.String("payment_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	reason := getString(event.Data, "reason")
	refundAmount := getString(event.Data, "refundAmount")

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":    string(domain.PaymentStatusRefunded),
			"refundId":  getString(event.Data, "refundId"),
			"updatedAt": event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": PaymentActivity{
				Action:    "refunded",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
				Details:   "Refund amount: " + refundAmount + ", Reason: " + reason,
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "payment:detail:"+event.AggregateID)
	h.cache.Delete(ctx, "payment:summary:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "payment:list:*")

	h.logger.New(ctx).Info("Payment refunded in read model",
		"payment_id", event.AggregateID,
		"tenant_id", event.TenantID,
		"refund_amount", refundAmount,
		"reason", reason,
	)

	return nil
}

func (h *PaymentEventHandler) HandlePaymentCancelled(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_payment_cancelled",
		trace.WithAttributes(
			attribute.String("payment_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	previousStatus := getString(event.Data, "previousStatus")

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":    string(domain.PaymentStatusCancelled),
			"updatedAt": event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": PaymentActivity{
				Action:    "cancelled",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
				Details:   "Cancelled from status: " + previousStatus,
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "payment:detail:"+event.AggregateID)
	h.cache.Delete(ctx, "payment:summary:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "payment:list:*")

	h.logger.New(ctx).Info("Payment cancelled in read model",
		"payment_id", event.AggregateID,
		"tenant_id", event.TenantID,
		"previous_status", previousStatus,
	)

	return nil
}

type PaymentSummary struct {
	ID          string    `bson:"_id" json:"id"`
	TenantID    string    `bson:"tenantId" json:"tenantId"`
	InvoiceID   string    `bson:"invoiceId" json:"invoiceId"`
	ClientID    string    `bson:"clientId" json:"clientId"`
	Amount      string    `bson:"amount" json:"amount"`
	Currency    string    `bson:"currency" json:"currency"`
	Status      string    `bson:"status" json:"status"`
	Method      string    `bson:"method" json:"method"`
	Provider    string    `bson:"provider" json:"provider,omitempty"`
	Reference   string    `bson:"reference" json:"reference,omitempty"`
	Description string    `bson:"description" json:"description,omitempty"`
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}

type PaymentDetail struct {
	ID             string                 `bson:"_id" json:"id"`
	TenantID       string                 `bson:"tenantId" json:"tenantId"`
	InvoiceID      string                 `bson:"invoiceId" json:"invoiceId"`
	ClientID       string                 `bson:"clientId" json:"clientId"`
	Amount         string                 `bson:"amount" json:"amount"`
	Currency       string                 `bson:"currency" json:"currency"`
	Status         string                 `bson:"status" json:"status"`
	Method         string                 `bson:"method" json:"method"`
	Provider       string                 `bson:"provider" json:"provider,omitempty"`
	ProviderID     string                 `bson:"providerId" json:"providerId,omitempty"`
	TransactionID  string                 `bson:"transactionId" json:"transactionId,omitempty"`
	Reference      string                 `bson:"reference" json:"reference,omitempty"`
	Description    string                 `bson:"description" json:"description,omitempty"`
	Metadata       map[string]interface{} `bson:"metadata" json:"metadata,omitempty"`
	FailureCode    string                 `bson:"failureCode" json:"failureCode,omitempty"`
	FailureMessage string                 `bson:"failureMessage" json:"failureMessage,omitempty"`
	RefundID       string                 `bson:"refundId" json:"refundId,omitempty"`
	ProcessedAt    *time.Time             `bson:"processedAt" json:"processedAt,omitempty"`
	ActivityLog    []PaymentActivity      `bson:"activityLog" json:"activityLog,omitempty"`
	CreatedAt      time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time              `bson:"updatedAt" json:"updatedAt"`
}

type PaymentActivity struct {
	Action    string    `bson:"action" json:"action"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	UserID    string    `bson:"userId" json:"userId,omitempty"`
	Details   string    `bson:"details,omitempty" json:"details,omitempty"`
}
