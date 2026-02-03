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

type InvoiceCommandHandler struct {
	invoiceRepo    InvoiceRepository
	eventStore     *repository.EventStore
	publisher      Publisher
	logger         *logger.Logger
	invoiceCounter InvoiceCounter
}

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *domain.Invoice) error
	Update(ctx context.Context, invoice *domain.Invoice) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error)
	FindByInvoiceNumber(ctx context.Context, tenantID uuid.UUID, invoiceNumber string) (*domain.Invoice, error)
	FindByClientID(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*domain.Invoice, error)
}

type InvoiceCounter interface {
	GetNextInvoiceNumber(ctx context.Context, tenantID uuid.UUID, year int) (string, error)
}

func NewInvoiceCommandHandler(
	invoiceRepo InvoiceRepository,
	eventStore *repository.EventStore,
	publisher Publisher,
	log *logger.Logger,
	invoiceCounter InvoiceCounter,
) *InvoiceCommandHandler {
	return &InvoiceCommandHandler{
		invoiceRepo:    invoiceRepo,
		eventStore:     eventStore,
		publisher:      publisher,
		logger:         log,
		invoiceCounter: invoiceCounter,
	}
}

func (h *InvoiceCommandHandler) HandleCreateInvoice(ctx context.Context, cmd *CommandEnvelope) (*domain.Invoice, error) {
	data := cmd.Data

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	clientID, err := uuid.Parse(getString(data, "clientId"))
	if err != nil {
		return nil, errors.InvalidArgument("invalid client ID")
	}

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid user ID")
	}

	invoiceType := domain.InvoiceType(getString(data, "type"))
	if invoiceType == "" {
		invoiceType = domain.InvoiceTypeStandard
	}

	currency := getString(data, "currency")
	if currency == "" {
		currency = "USD"
	}

	paymentTerm := domain.PaymentTerm(getString(data, "paymentTerm"))
	if paymentTerm == "" {
		paymentTerm = domain.PaymentTermNet30
	}

	issueDate := time.Now().UTC()
	if dateStr, ok := data["issueDate"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, dateStr); err == nil {
			issueDate = parsed
		}
	}

	invoice, err := domain.NewInvoice(
		tenantID,
		clientID,
		userID,
		invoiceType,
		currency,
		paymentTerm,
		issueDate,
	)
	if err != nil {
		return nil, err
	}

	if dueDateStr, ok := data["dueDate"].(string); ok {
		if dueDate, err := time.Parse(time.RFC3339, dueDateStr); err == nil {
			invoice.SetDueDate(dueDate)
		}
	} else {
		dueDate := invoice.CalculateDueDate()
		invoice.SetDueDate(dueDate)
	}

	if notes, ok := data["notes"].(string); ok {
		invoice.SetNotes(notes)
	}

	if terms, ok := data["terms"].(string); ok {
		invoice.SetTerms(terms)
	}

	year := issueDate.Year()
	invoiceNumber, err := h.invoiceCounter.GetNextInvoiceNumber(ctx, tenantID, year)
	if err != nil {
		h.logger.New(ctx).Error("Failed to generate invoice number", "error", err)
		return nil, errors.InternalError("failed to generate invoice number")
	}
	invoice.SetInvoiceNumber(invoiceNumber)

	if err := h.invoiceRepo.Create(ctx, invoice); err != nil {
		h.logger.New(ctx).Error("Failed to create invoice", "error", err)
		return nil, errors.InternalError("failed to create invoice")
	}

	event := eventpkg.NewEvent(
		invoice.ID.String(),
		"invoice",
		"invoice.created",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"invoiceNumber": invoice.InvoiceNumber,
			"clientId":      invoice.ClientID.String(),
			"type":          string(invoice.Type),
			"currency":      invoice.Currency,
			"subtotal":      invoice.Subtotal.String(),
			"taxTotal":      invoice.TaxTotal.String(),
			"total":         invoice.Total.String(),
			"paymentTerm":   string(invoice.PaymentTerm),
			"status":        string(invoice.Status),
			"notes":         invoice.Notes,
			"terms":         invoice.Terms,
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish invoice created event", "error", err)
	}

	h.logger.New(ctx).Info("Invoice created",
		"invoice_id", invoice.ID,
		"invoice_number", invoice.InvoiceNumber,
		"tenant_id", cmd.TenantID,
		"client_id", invoice.ClientID,
	)

	return invoice, nil
}

func (h *InvoiceCommandHandler) HandleAddLineItem(ctx context.Context, cmd *CommandEnvelope) (*domain.Invoice, error) {
	data := cmd.Data

	invoiceID, err := uuid.Parse(cmd.TargetID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid invoice ID")
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	invoice, err := h.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, errors.NotFound("invoice not found")
	}

	if invoice.TenantID != tenantID {
		return nil, errors.Newf(errors.CodeForbidden, "invoice does not belong to tenant")
	}

	if invoice.Status != domain.InvoiceStatusDraft {
		return nil, errors.InvalidArgument("can only add lines to draft invoices")
	}

	description := getString(data, "description")
	if description == "" {
		return nil, errors.InvalidArgument("description is required")
	}

	quantity := decimal.Zero
	if qtyStr, ok := data["quantity"].(string); ok {
		quantity, _ = decimal.NewFromString(qtyStr)
	}
	if quantity.IsZero() {
		return nil, errors.InvalidArgument("quantity must be greater than zero")
	}

	unitPrice := decimal.Zero
	if priceStr, ok := data["unitPrice"].(string); ok {
		unitPrice, _ = decimal.NewFromString(priceStr)
	}

	discount := decimal.Zero
	if discountStr, ok := data["discount"].(string); ok {
		discount, _ = decimal.NewFromString(discountStr)
	}

	taxRate := decimal.Zero
	if rateStr, ok := data["taxRate"].(string); ok {
		taxRate, _ = decimal.NewFromString(rateStr)
	}

	line := domain.InvoiceLine{
		Description: description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		Discount:    discount,
		TaxRate:     taxRate,
	}

	if productIDStr, ok := data["productId"].(string); ok {
		if productID, err := uuid.Parse(productIDStr); err == nil {
			line.ProductID = &productID
		}
	}

	if sortOrder, ok := data["sortOrder"].(float64); ok {
		line.SortOrder = int(sortOrder)
	}

	invoice.AddLine(line)

	if err := h.invoiceRepo.Update(ctx, invoice); err != nil {
		h.logger.New(ctx).Error("Failed to update invoice with line item", "error", err)
		return nil, errors.InternalError("failed to add line item")
	}

	event := eventpkg.NewEvent(
		invoice.ID.String(),
		"invoice",
		"invoice.line_added",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"lineId":      invoice.Lines[len(invoice.Lines)-1].ID.String(),
			"description": description,
			"quantity":    quantity.String(),
			"unitPrice":   unitPrice.String(),
			"total":       invoice.Total.String(),
			"subtotal":    invoice.Subtotal.String(),
			"taxTotal":    invoice.TaxTotal.String(),
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish line added event", "error", err)
	}

	h.logger.New(ctx).Info("Invoice line added",
		"invoice_id", invoice.ID,
		"line_id", invoice.Lines[len(invoice.Lines)-1].ID,
		"description", description,
	)

	return invoice, nil
}

func (h *InvoiceCommandHandler) HandleRemoveLineItem(ctx context.Context, cmd *CommandEnvelope) (*domain.Invoice, error) {
	invoiceID, err := uuid.Parse(cmd.TargetID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid invoice ID")
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	lineID, err := uuid.Parse(getString(cmd.Data, "lineId"))
	if err != nil {
		return nil, errors.InvalidArgument("invalid line ID")
	}

	invoice, err := h.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, errors.NotFound("invoice not found")
	}

	if invoice.TenantID != tenantID {
		return nil, errors.Newf(errors.CodeForbidden, "invoice does not belong to tenant")
	}

	if invoice.Status != domain.InvoiceStatusDraft {
		return nil, errors.InvalidArgument("can only remove lines from draft invoices")
	}

	invoice.RemoveLine(lineID)

	if err := h.invoiceRepo.Update(ctx, invoice); err != nil {
		h.logger.New(ctx).Error("Failed to remove line item", "error", err)
		return nil, errors.InternalError("failed to remove line item")
	}

	event := eventpkg.NewEvent(
		invoice.ID.String(),
		"invoice",
		"invoice.line_removed",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"lineId":   lineID.String(),
			"total":    invoice.Total.String(),
			"subtotal": invoice.Subtotal.String(),
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish line removed event", "error", err)
	}

	h.logger.New(ctx).Info("Invoice line removed",
		"invoice_id", invoice.ID,
		"line_id", lineID,
	)

	return invoice, nil
}

func (h *InvoiceCommandHandler) HandleFinalizeInvoice(ctx context.Context, cmd *CommandEnvelope) (*domain.Invoice, error) {
	invoiceID, err := uuid.Parse(cmd.TargetID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid invoice ID")
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	invoice, err := h.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, errors.NotFound("invoice not found")
	}

	if invoice.TenantID != tenantID {
		return nil, errors.Newf(errors.CodeForbidden, "invoice does not belong to tenant")
	}

	if invoice.Status != domain.InvoiceStatusDraft {
		return nil, errors.InvalidArgument("only draft invoices can be finalized")
	}

	if len(invoice.Lines) == 0 {
		return nil, errors.InvalidArgument("cannot finalize invoice with no line items")
	}

	invoice.SetStatus(domain.InvoiceStatusPending)

	if err := h.invoiceRepo.Update(ctx, invoice); err != nil {
		h.logger.New(ctx).Error("Failed to finalize invoice", "error", err)
		return nil, errors.InternalError("failed to finalize invoice")
	}

	event := eventpkg.NewEvent(
		invoice.ID.String(),
		"invoice",
		"invoice.finalized",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"invoiceNumber": invoice.InvoiceNumber,
			"total":         invoice.Total.String(),
			"amountDue":     invoice.AmountDue.String(),
			"dueDate":       invoice.DueDate,
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish finalized event", "error", err)
	}

	h.logger.New(ctx).Info("Invoice finalized",
		"invoice_id", invoice.ID,
		"invoice_number", invoice.InvoiceNumber,
	)

	return invoice, nil
}

func (h *InvoiceCommandHandler) HandleSendInvoice(ctx context.Context, cmd *CommandEnvelope) (*domain.Invoice, error) {
	invoiceID, err := uuid.Parse(cmd.TargetID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid invoice ID")
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	invoice, err := h.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, errors.NotFound("invoice not found")
	}

	if invoice.TenantID != tenantID {
		return nil, errors.Newf(errors.CodeForbidden, "invoice does not belong to tenant")
	}

	if invoice.Status != domain.InvoiceStatusDraft && invoice.Status != domain.InvoiceStatusPending {
		return nil, errors.InvalidArgument("only draft or pending invoices can be sent")
	}

	if invoice.Status == domain.InvoiceStatusDraft {
		invoice.SetStatus(domain.InvoiceStatusPending)
	}

	invoice.Send()

	if err := h.invoiceRepo.Update(ctx, invoice); err != nil {
		h.logger.New(ctx).Error("Failed to send invoice", "error", err)
		return nil, errors.InternalError("failed to send invoice")
	}

	event := eventpkg.NewEvent(
		invoice.ID.String(),
		"invoice",
		"invoice.sent",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"invoiceNumber": invoice.InvoiceNumber,
			"sentDate":      invoice.SentDate,
			"total":         invoice.Total.String(),
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish sent event", "error", err)
	}

	h.logger.New(ctx).Info("Invoice sent",
		"invoice_id", invoice.ID,
		"invoice_number", invoice.InvoiceNumber,
	)

	return invoice, nil
}

func (h *InvoiceCommandHandler) HandleVoidInvoice(ctx context.Context, cmd *CommandEnvelope) (*domain.Invoice, error) {
	invoiceID, err := uuid.Parse(cmd.TargetID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid invoice ID")
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	invoice, err := h.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, errors.NotFound("invoice not found")
	}

	if invoice.TenantID != tenantID {
		return nil, errors.Newf(errors.CodeForbidden, "invoice does not belong to tenant")
	}

	if invoice.Status == domain.InvoiceStatusPaid {
		return nil, errors.InvalidArgument("cannot void a paid invoice")
	}

	if invoice.Status == domain.InvoiceStatusRefunded {
		return nil, errors.InvalidArgument("cannot void a refunded invoice")
	}

	if invoice.Status == domain.InvoiceStatusCancelled {
		return nil, errors.InvalidArgument("invoice is already cancelled")
	}

	reason := getString(cmd.Data, "reason")
	if reason == "" {
		reason = "Voided by user"
	}

	previousStatus := string(domain.InvoiceStatusDraft)
	if invoice.Status == domain.InvoiceStatusSent {
		previousStatus = string(domain.InvoiceStatusSent)
	}

	invoice.Cancel(reason)

	if err := h.invoiceRepo.Update(ctx, invoice); err != nil {
		h.logger.New(ctx).Error("Failed to void invoice", "error", err)
		return nil, errors.InternalError("failed to void invoice")
	}

	event := eventpkg.NewEvent(
		invoice.ID.String(),
		"invoice",
		"invoice.voided",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"invoiceNumber":  invoice.InvoiceNumber,
			"reason":         reason,
			"previousStatus": previousStatus,
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish voided event", "error", err)
	}

	h.logger.New(ctx).Info("Invoice voided",
		"invoice_id", invoice.ID,
		"invoice_number", invoice.InvoiceNumber,
		"reason", reason,
	)

	return invoice, nil
}

func (h *InvoiceCommandHandler) HandleRecordPayment(ctx context.Context, cmd *CommandEnvelope) (*domain.Invoice, error) {
	invoiceID, err := uuid.Parse(cmd.TargetID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid invoice ID")
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	invoice, err := h.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, errors.NotFound("invoice not found")
	}

	if invoice.TenantID != tenantID {
		return nil, errors.Newf(errors.CodeForbidden, "invoice does not belong to tenant")
	}

	if invoice.Status == domain.InvoiceStatusPaid {
		return nil, errors.InvalidArgument("invoice is already fully paid")
	}

	if invoice.Status == domain.InvoiceStatusCancelled {
		return nil, errors.InvalidArgument("cannot record payment for cancelled invoice")
	}

	amount := decimal.Zero
	if amountStr, ok := cmd.Data["amount"].(string); ok {
		amount, _ = decimal.NewFromString(amountStr)
	}

	if amount.IsZero() || amount.IsNegative() {
		return nil, errors.InvalidArgument("payment amount must be greater than zero")
	}

	if amount.GreaterThan(invoice.AmountDue) {
		return nil, errors.Newf(errors.CodeInvalidArgument, "payment amount exceeds amount due: %s", invoice.AmountDue.String())
	}

	if err := invoice.ApplyPayment(amount); err != nil {
		return nil, err
	}

	if err := h.invoiceRepo.Update(ctx, invoice); err != nil {
		h.logger.New(ctx).Error("Failed to record payment", "error", err)
		return nil, errors.InternalError("failed to record payment")
	}

	paymentMethod := getString(cmd.Data, "paymentMethod")
	reference := getString(cmd.Data, "reference")

	event := eventpkg.NewEvent(
		invoice.ID.String(),
		"invoice",
		"invoice.payment_recorded",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"invoiceNumber": invoice.InvoiceNumber,
			"amount":        amount.String(),
			"amountPaid":    invoice.AmountPaid.String(),
			"amountDue":     invoice.AmountDue.String(),
			"paymentMethod": paymentMethod,
			"reference":     reference,
			"status":        string(invoice.Status),
		},
	)
	event.WithCorrelationID(cmd.CorrelationID)

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.New(ctx).Error("Failed to publish payment recorded event", "error", err)
	}

	h.logger.New(ctx).Info("Payment recorded",
		"invoice_id", invoice.ID,
		"invoice_number", invoice.InvoiceNumber,
		"amount", amount.String(),
		"amount_paid", invoice.AmountPaid.String(),
		"amount_due", invoice.AmountDue.String(),
	)

	return invoice, nil
}
