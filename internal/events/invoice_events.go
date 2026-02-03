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

type InvoiceEventHandler struct {
	readModelStore *repository.ReadModelStore
	cache          *repository.Cache
	logger         *logger.Logger
	tracer         trace.Tracer
}

func NewInvoiceEventHandler(
	readModelStore *repository.ReadModelStore,
	cache *repository.Cache,
	log *logger.Logger,
) *InvoiceEventHandler {
	return &InvoiceEventHandler{
		readModelStore: readModelStore,
		cache:          cache,
		logger:         log,
		tracer:         otel.Tracer("invoice-event-handler"),
	}
}

func (h *InvoiceEventHandler) HandleInvoiceCreated(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_invoice_created",
		trace.WithAttributes(
			attribute.String("invoice_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	dueDate := getTime(event.Data, "dueDate")

	summary := InvoiceSummary{
		ID:            event.AggregateID,
		TenantID:      event.TenantID,
		InvoiceNumber: getString(event.Data, "invoiceNumber"),
		ClientID:      getString(event.Data, "clientId"),
		Type:          getString(event.Data, "type"),
		Status:        getString(event.Data, "status"),
		Currency:      getString(event.Data, "currency"),
		Subtotal:      getString(event.Data, "subtotal"),
		TaxTotal:      getString(event.Data, "taxTotal"),
		DiscountTotal: "0",
		Total:         getString(event.Data, "total"),
		AmountPaid:    "0",
		AmountDue:     getString(event.Data, "total"),
		PaymentTerm:   getString(event.Data, "paymentTerm"),
		DueDate:       dueDate,
		IssueDate:     event.Timestamp,
		LineCount:     0,
		Notes:         getString(event.Data, "notes"),
		CreatedAt:     event.Timestamp,
		UpdatedAt:     event.Timestamp,
	}

	if err := h.readModelStore.Save(ctx, summary); err != nil {
		span.RecordError(err)
		return err
	}

	detail := InvoiceDetail{
		ID:            event.AggregateID,
		TenantID:      event.TenantID,
		InvoiceNumber: getString(event.Data, "invoiceNumber"),
		ClientID:      getString(event.Data, "clientId"),
		Type:          getString(event.Data, "type"),
		Status:        getString(event.Data, "status"),
		Currency:      getString(event.Data, "currency"),
		Subtotal:      getString(event.Data, "subtotal"),
		TaxTotal:      getString(event.Data, "taxTotal"),
		Total:         getString(event.Data, "total"),
		AmountPaid:    "0",
		AmountDue:     getString(event.Data, "total"),
		PaymentTerm:   getString(event.Data, "paymentTerm"),
		DueDate:       dueDate,
		IssueDate:     event.Timestamp,
		Lines:         []InvoiceLineSummary{},
		Notes:         getString(event.Data, "notes"),
		Terms:         getString(event.Data, "terms"),
		ActivityLog: []InvoiceActivity{
			{
				Action:    "created",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
			},
		},
		CreatedAt: event.Timestamp,
		UpdatedAt: event.Timestamp,
	}

	if err := h.readModelStore.Save(ctx, detail); err != nil {
		span.RecordError(err)
		return err
	}

	h.logger.New(ctx).Info("Invoice created in read model",
		"invoice_id", event.AggregateID,
		"invoice_number", summary.InvoiceNumber,
		"tenant_id", event.TenantID,
	)

	return nil
}

func (h *InvoiceEventHandler) HandleLineItemAdded(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_line_item_added",
		trace.WithAttributes(
			attribute.String("invoice_id", event.AggregateID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	lineID := getString(event.Data, "lineId")

	lineSummary := InvoiceLineSummary{
		ID:          lineID,
		Description: getString(event.Data, "description"),
		Quantity:    getString(event.Data, "quantity"),
		UnitPrice:   getString(event.Data, "unitPrice"),
		Total:       "0",
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"subtotal":  getString(event.Data, "subtotal"),
			"taxTotal":  getString(event.Data, "taxTotal"),
			"total":     getString(event.Data, "total"),
			"amountDue": getString(event.Data, "total"),
			"updatedAt": event.Timestamp,
		},
		"$push": map[string]interface{}{
			"lines": lineSummary,
		},
		"$inc": map[string]interface{}{
			"lineCount": 1,
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "invoice:detail:"+event.AggregateID)
	h.cache.Delete(ctx, "invoice:summary:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "invoice:list:*")

	h.logger.New(ctx).Debug("Line item added to invoice read model",
		"invoice_id", event.AggregateID,
		"line_id", lineID,
	)

	return nil
}

func (h *InvoiceEventHandler) HandleLineItemRemoved(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_line_item_removed",
		trace.WithAttributes(
			attribute.String("invoice_id", event.AggregateID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	lineID := getString(event.Data, "lineId")

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"subtotal":  getString(event.Data, "subtotal"),
			"total":     getString(event.Data, "total"),
			"amountDue": getString(event.Data, "total"),
			"updatedAt": event.Timestamp,
		},
		"$pull": map[string]interface{}{
			"lines": map[string]interface{}{
				"id": lineID,
			},
		},
		"$inc": map[string]interface{}{
			"lineCount": -1,
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "invoice:detail:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "invoice:list:*")

	h.logger.New(ctx).Debug("Line item removed from invoice read model",
		"invoice_id", event.AggregateID,
		"line_id", lineID,
	)

	return nil
}

func (h *InvoiceEventHandler) HandleInvoiceFinalized(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_invoice_finalized",
		trace.WithAttributes(
			attribute.String("invoice_id", event.AggregateID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":    string(domain.InvoiceStatusPending),
			"total":     getString(event.Data, "total"),
			"amountDue": getString(event.Data, "amountDue"),
			"dueDate":   getTime(event.Data, "dueDate"),
			"updatedAt": event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": InvoiceActivity{
				Action:    "finalized",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "invoice:detail:"+event.AggregateID)
	h.cache.Delete(ctx, "invoice:summary:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "invoice:list:*")

	h.logger.New(ctx).Info("Invoice finalized in read model",
		"invoice_id", event.AggregateID,
	)

	return nil
}

func (h *InvoiceEventHandler) HandleInvoiceSent(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_invoice_sent",
		trace.WithAttributes(
			attribute.String("invoice_id", event.AggregateID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":    string(domain.InvoiceStatusSent),
			"sentDate":  getTime(event.Data, "sentDate"),
			"updatedAt": event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": InvoiceActivity{
				Action:    "sent",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "invoice:detail:"+event.AggregateID)
	h.cache.Delete(ctx, "invoice:summary:"+event.AggregateID)

	h.logger.New(ctx).Info("Invoice sent in read model",
		"invoice_id", event.AggregateID,
	)

	return nil
}

func (h *InvoiceEventHandler) HandleInvoiceVoided(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_invoice_voided",
		trace.WithAttributes(
			attribute.String("invoice_id", event.AggregateID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	reason := getString(event.Data, "reason")

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":    string(domain.InvoiceStatusCancelled),
			"updatedAt": event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": InvoiceActivity{
				Action:    "voided",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
				Details:   reason,
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "invoice:detail:"+event.AggregateID)
	h.cache.Delete(ctx, "invoice:summary:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "invoice:list:*")

	h.logger.New(ctx).Info("Invoice voided in read model",
		"invoice_id", event.AggregateID,
		"reason", reason,
	)

	return nil
}

func (h *InvoiceEventHandler) HandlePaymentRecorded(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_payment_recorded",
		trace.WithAttributes(
			attribute.String("invoice_id", event.AggregateID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	status := getString(event.Data, "status")
	paidDate := time.Time{}
	if status == string(domain.InvoiceStatusPaid) {
		paidDate = event.Timestamp
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"amountPaid": getString(event.Data, "amountPaid"),
			"amountDue":  getString(event.Data, "amountDue"),
			"status":     status,
			"updatedAt":  event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": InvoiceActivity{
				Action:    "payment_recorded",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
				Details:   "Payment of " + getString(event.Data, "amount") + " recorded",
			},
		},
	}

	if !paidDate.IsZero() {
		update["$set"].(map[string]interface{})["paidDate"] = paidDate
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "invoice:detail:"+event.AggregateID)
	h.cache.Delete(ctx, "invoice:summary:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "invoice:list:*")

	h.logger.New(ctx).Info("Payment recorded in invoice read model",
		"invoice_id", event.AggregateID,
		"amount_paid", getString(event.Data, "amountPaid"),
		"amount_due", getString(event.Data, "amountDue"),
	)

	return nil
}

type InvoiceSummary struct {
	ID            string    `bson:"_id" json:"id"`
	TenantID      string    `bson:"tenantId" json:"tenantId"`
	InvoiceNumber string    `bson:"invoiceNumber" json:"invoiceNumber"`
	ClientID      string    `bson:"clientId" json:"clientId"`
	ClientName    string    `bson:"clientName" json:"clientName,omitempty"`
	Type          string    `bson:"type" json:"type"`
	Status        string    `bson:"status" json:"status"`
	Currency      string    `bson:"currency" json:"currency"`
	Subtotal      string    `bson:"subtotal" json:"subtotal"`
	TaxTotal      string    `bson:"taxTotal" json:"taxTotal"`
	DiscountTotal string    `bson:"discountTotal" json:"discountTotal"`
	Total         string    `bson:"total" json:"total"`
	AmountPaid    string    `bson:"amountPaid" json:"amountPaid"`
	AmountDue     string    `bson:"amountDue" json:"amountDue"`
	PaymentTerm   string    `bson:"paymentTerm" json:"paymentTerm"`
	DueDate       time.Time `bson:"dueDate" json:"dueDate,omitempty"`
	IssueDate     time.Time `bson:"issueDate" json:"issueDate"`
	SentDate      time.Time `bson:"sentDate" json:"sentDate,omitempty"`
	PaidDate      time.Time `bson:"paidDate" json:"paidDate,omitempty"`
	LineCount     int       `bson:"lineCount" json:"lineCount"`
	Notes         string    `bson:"notes" json:"notes,omitempty"`
	CreatedAt     time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt" json:"updatedAt"`
}

type InvoiceDetail struct {
	ID            string               `bson:"_id" json:"id"`
	TenantID      string               `bson:"tenantId" json:"tenantId"`
	InvoiceNumber string               `bson:"invoiceNumber" json:"invoiceNumber"`
	ClientID      string               `bson:"clientId" json:"clientId"`
	ClientName    string               `bson:"clientName" json:"clientName,omitempty"`
	Type          string               `bson:"type" json:"type"`
	Status        string               `bson:"status" json:"status"`
	Currency      string               `bson:"currency" json:"currency"`
	Subtotal      string               `bson:"subtotal" json:"subtotal"`
	TaxTotal      string               `bson:"taxTotal" json:"taxTotal"`
	DiscountTotal string               `bson:"discountTotal" json:"discountTotal"`
	Total         string               `bson:"total" json:"total"`
	AmountPaid    string               `bson:"amountPaid" json:"amountPaid"`
	AmountDue     string               `bson:"amountDue" json:"amountDue"`
	PaymentTerm   string               `bson:"paymentTerm" json:"paymentTerm"`
	DueDate       time.Time            `bson:"dueDate" json:"dueDate,omitempty"`
	IssueDate     time.Time            `bson:"issueDate" json:"issueDate"`
	SentDate      time.Time            `bson:"sentDate" json:"sentDate,omitempty"`
	PaidDate      time.Time            `bson:"paidDate" json:"paidDate,omitempty"`
	Lines         []InvoiceLineSummary `bson:"lines" json:"lines"`
	Notes         string               `bson:"notes" json:"notes,omitempty"`
	Terms         string               `bson:"terms" json:"terms,omitempty"`
	ActivityLog   []InvoiceActivity    `bson:"activityLog" json:"activityLog,omitempty"`
	CreatedAt     time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time            `bson:"updatedAt" json:"updatedAt"`
}

type InvoiceLineSummary struct {
	ID          string `bson:"id" json:"id"`
	Description string `bson:"description" json:"description"`
	Quantity    string `bson:"quantity" json:"quantity"`
	UnitPrice   string `bson:"unitPrice" json:"unitPrice"`
	Discount    string `bson:"discount" json:"discount,omitempty"`
	TaxRate     string `bson:"taxRate" json:"taxRate,omitempty"`
	TaxAmount   string `bson:"taxAmount" json:"taxAmount,omitempty"`
	Total       string `bson:"total" json:"total"`
	ProductID   string `bson:"productId" json:"productId,omitempty"`
	SortOrder   int    `bson:"sortOrder" json:"sortOrder"`
}

type InvoiceActivity struct {
	Action    string    `bson:"action" json:"action"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	UserID    string    `bson:"userId" json:"userId,omitempty"`
	Details   string    `bson:"details,omitempty" json:"details,omitempty"`
}
