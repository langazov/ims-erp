package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/ims-erp/system/internal/events"
)

// InvoicePaymentEventBroadcaster broadcasts invoice and payment events via WebSocket
type InvoicePaymentEventBroadcaster struct {
	hub     *Hub
	bridge  *EventBridge
	events  <-chan *events.EventEnvelope
	filters map[string]EventFilter
	mu      sync.RWMutex
}

// EventFilter defines filtering criteria for events
type EventFilter struct {
	TenantID         string
	EventTypes       []string
	SubscriptionType string
}

// NewInvoicePaymentEventBroadcaster creates a new event broadcaster
func NewInvoicePaymentEventBroadcaster(hub *Hub) *InvoicePaymentEventBroadcaster {
	return &InvoicePaymentEventBroadcaster{
		hub:     hub,
		filters: make(map[string]EventFilter),
	}
}

// Start begins listening for events and broadcasting to WebSocket clients
func (b *InvoicePaymentEventBroadcaster) Start(ctx context.Context) {
	for {
		select {
		case event := <-b.events:
			if event == nil {
				continue
			}
			b.handleEvent(event)
		case <-ctx.Done():
			return
		}
	}
}

// SetEventChannel sets the channel for receiving events
func (b *InvoicePaymentEventBroadcaster) SetEventChannel(events <-chan *events.EventEnvelope) {
	b.events = events
}

// RegisterFilter registers a filter for a specific client
func (b *InvoicePaymentEventBroadcaster) RegisterFilter(clientID string, filter EventFilter) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.filters[clientID] = filter
}

// UnregisterFilter removes a filter for a client
func (b *InvoicePaymentEventBroadcaster) UnregisterFilter(clientID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.filters, clientID)
}

// handleEvent processes and broadcasts an event
func (b *InvoicePaymentEventBroadcaster) handleEvent(event *events.EventEnvelope) {
	if event == nil {
		return
	}

	// Check if this is an invoice or payment event
	if !b.isInvoicePaymentEvent(event.Type) {
		return
	}

	// Apply filters
	b.mu.RLock()
	defer b.mu.RUnlock()

	for clientID, filter := range b.filters {
		if b.shouldBroadcast(event, filter) {
			b.broadcastToClient(clientID, event)
		}
	}
}

// isInvoicePaymentEvent checks if the event type is invoice or payment related
func (b *InvoicePaymentEventBroadcaster) isInvoicePaymentEvent(eventType string) bool {
	invoiceEvents := []string{
		"invoice.created",
		"invoice.updated",
		"invoice.paid",
		"invoice.voided",
		"invoice.sent",
		"invoice.overdue",
	}

	paymentEvents := []string{
		"payment.received",
		"payment.failed",
		"payment.refunded",
		"payment.processing",
	}

	for _, t := range invoiceEvents {
		if eventType == t {
			return true
		}
	}

	for _, t := range paymentEvents {
		if eventType == t {
			return true
		}
	}

	return false
}

// shouldBroadcast checks if an event should be broadcast based on filters
func (b *InvoicePaymentEventBroadcaster) shouldBroadcast(event *events.EventEnvelope, filter EventFilter) bool {
	// Check tenant filter
	if filter.TenantID != "" && filter.TenantID != event.TenantID {
		return false
	}

	// Check event type filter
	if len(filter.EventTypes) > 0 {
		found := false
		for _, t := range filter.EventTypes {
			if t == event.Type {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// broadcastToClient sends an event to a specific client
func (b *InvoicePaymentEventBroadcaster) broadcastToClient(clientID string, event *events.EventEnvelope) {
	msg := &Message{
		Type:     event.Type,
		Payload:  mustMarshal(event.Data),
		TenantID: event.TenantID,
	}

	b.hub.Broadcast(msg)
}

// BroadcastInvoiceCreated broadcasts an invoice created event
func (b *InvoicePaymentEventBroadcaster) BroadcastInvoiceCreated(invoice *events.InvoiceSummary) error {
	data := map[string]interface{}{
		"id":            invoice.ID,
		"tenantId":      invoice.TenantID,
		"clientId":      invoice.ClientID,
		"invoiceNumber": invoice.InvoiceNumber,
		"status":        invoice.Status,
		"total":         invoice.Total,
		"amountDue":     invoice.AmountDue,
		"createdAt":     invoice.CreatedAt,
	}

	event := events.NewEvent(
		invoice.ID,
		"invoice",
		"invoice.created",
		invoice.TenantID,
		"system",
		data,
	)

	msg := &Message{
		Type:     event.Type,
		Payload:  mustMarshal(event.Data),
		TenantID: event.TenantID,
	}

	b.hub.Broadcast(msg)
	return nil
}

// BroadcastInvoicePaid broadcasts an invoice paid event
func (b *InvoicePaymentEventBroadcaster) BroadcastInvoicePaid(invoice *events.InvoiceSummary, payment *events.PaymentSummary) error {
	data := map[string]interface{}{
		"invoice": map[string]interface{}{
			"id":            invoice.ID,
			"invoiceNumber": invoice.InvoiceNumber,
			"status":        invoice.Status,
			"amountPaid":    invoice.AmountPaid,
			"amountDue":     invoice.AmountDue,
		},
		"payment": map[string]interface{}{
			"id":     payment.ID,
			"amount": payment.Amount,
			"method": payment.Method,
		},
	}

	event := events.NewEvent(
		invoice.ID,
		"invoice",
		"invoice.paid",
		invoice.TenantID,
		"system",
		data,
	)

	msg := &Message{
		Type:     event.Type,
		Payload:  mustMarshal(event.Data),
		TenantID: event.TenantID,
	}

	b.hub.Broadcast(msg)
	return nil
}

// BroadcastPaymentReceived broadcasts a payment received event
func (b *InvoicePaymentEventBroadcaster) BroadcastPaymentReceived(payment *events.PaymentSummary) error {
	data := map[string]interface{}{
		"id":        payment.ID,
		"tenantId":  payment.TenantID,
		"invoiceId": payment.InvoiceID,
		"amount":    payment.Amount,
		"currency":  payment.Currency,
		"method":    payment.Method,
		"status":    payment.Status,
		"createdAt": payment.CreatedAt,
	}

	event := events.NewEvent(
		payment.ID,
		"payment",
		"payment.received",
		payment.TenantID,
		"system",
		data,
	)

	msg := &Message{
		Type:     event.Type,
		Payload:  mustMarshal(event.Data),
		TenantID: event.TenantID,
	}

	b.hub.Broadcast(msg)
	return nil
}

// BroadcastPaymentFailed broadcasts a payment failed event
func (b *InvoicePaymentEventBroadcaster) BroadcastPaymentFailed(payment *events.PaymentSummary, reason string) error {
	data := map[string]interface{}{
		"id":        payment.ID,
		"tenantId":  payment.TenantID,
		"invoiceId": payment.InvoiceID,
		"amount":    payment.Amount,
		"reason":    reason,
		"failedAt":  time.Now().UTC(),
	}

	event := events.NewEvent(
		payment.ID,
		"payment",
		"payment.failed",
		payment.TenantID,
		"system",
		data,
	)

	msg := &Message{
		Type:     event.Type,
		Payload:  mustMarshal(event.Data),
		TenantID: event.TenantID,
	}

	b.hub.Broadcast(msg)
	return nil
}

// BroadcastInvoiceUpdated broadcasts an invoice updated event
func (b *InvoicePaymentEventBroadcaster) BroadcastInvoiceUpdated(invoice *events.InvoiceSummary, changes map[string]interface{}) error {
	data := map[string]interface{}{
		"id":            invoice.ID,
		"tenantId":      invoice.TenantID,
		"invoiceNumber": invoice.InvoiceNumber,
		"status":        invoice.Status,
		"changes":       changes,
		"updatedAt":     time.Now().UTC(),
	}

	event := events.NewEvent(
		invoice.ID,
		"invoice",
		"invoice.updated",
		invoice.TenantID,
		"system",
		data,
	)

	msg := &Message{
		Type:     event.Type,
		Payload:  mustMarshal(event.Data),
		TenantID: event.TenantID,
	}

	b.hub.Broadcast(msg)
	return nil
}

// BroadcastPaymentRefunded broadcasts a payment refunded event
func (b *InvoicePaymentEventBroadcaster) BroadcastPaymentRefunded(payment *events.PaymentSummary, amount string) error {
	data := map[string]interface{}{
		"id":             payment.ID,
		"tenantId":       payment.TenantID,
		"invoiceId":      payment.InvoiceID,
		"refundedAmount": amount,
		"refundedAt":     time.Now().UTC(),
	}

	event := events.NewEvent(
		payment.ID,
		"payment",
		"payment.refunded",
		payment.TenantID,
		"system",
		data,
	)

	msg := &Message{
		Type:     event.Type,
		Payload:  mustMarshal(event.Data),
		TenantID: event.TenantID,
	}

	b.hub.Broadcast(msg)
	return nil
}

// GetStats returns broadcaster statistics
func (b *InvoicePaymentEventBroadcaster) GetStats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return map[string]interface{}{
		"activeFilters": len(b.filters),
		"hubClients":    len(b.hub.clients),
	}
}

// InvoicePaymentEventHandler handles invoice and payment events from NATS
type InvoicePaymentEventHandler struct {
	broadcaster *InvoicePaymentEventBroadcaster
	handler     events.EventHandler
}

// NewInvoicePaymentEventHandler creates a new event handler
func NewInvoicePaymentEventHandler(broadcaster *InvoicePaymentEventBroadcaster) *InvoicePaymentEventHandler {
	return &InvoicePaymentEventHandler{
		broadcaster: broadcaster,
	}
}

// Start begins listening for NATS events
func (h *InvoicePaymentEventHandler) Start(ctx context.Context) error {
	// Event handling is done through the event channel
	return nil
}

// HandleEvent processes events
func (h *InvoicePaymentEventHandler) HandleEvent(ctx context.Context, event *events.EventEnvelope) error {
	switch event.Type {
	case "invoice.created":
		var invoice events.InvoiceSummary
		data, _ := json.Marshal(event.Data)
		if err := json.Unmarshal(data, &invoice); err == nil {
			return h.broadcaster.BroadcastInvoiceCreated(&invoice)
		}
	case "invoice.paid":
		var data struct {
			Invoice events.InvoiceSummary `json:"invoice"`
			Payment events.PaymentSummary `json:"payment"`
		}
		eventData, _ := json.Marshal(event.Data)
		if err := json.Unmarshal(eventData, &data); err == nil {
			return h.broadcaster.BroadcastInvoicePaid(&data.Invoice, &data.Payment)
		}
	case "invoice.updated":
		var data struct {
			Invoice events.InvoiceSummary  `json:"invoice"`
			Changes map[string]interface{} `json:"changes"`
		}
		eventData, _ := json.Marshal(event.Data)
		if err := json.Unmarshal(eventData, &data); err == nil {
			return h.broadcaster.BroadcastInvoiceUpdated(&data.Invoice, data.Changes)
		}
	case "payment.received":
		var payment events.PaymentSummary
		data, _ := json.Marshal(event.Data)
		if err := json.Unmarshal(data, &payment); err == nil {
			return h.broadcaster.BroadcastPaymentReceived(&payment)
		}
	case "payment.failed":
		var data struct {
			Payment events.PaymentSummary `json:"payment"`
			Reason  string                `json:"reason"`
		}
		eventData, _ := json.Marshal(event.Data)
		if err := json.Unmarshal(eventData, &data); err == nil {
			return h.broadcaster.BroadcastPaymentFailed(&data.Payment, data.Reason)
		}
	case "payment.refunded":
		var data struct {
			Payment        events.PaymentSummary `json:"payment"`
			RefundedAmount string                `json:"refundedAmount"`
		}
		eventData, _ := json.Marshal(event.Data)
		if err := json.Unmarshal(eventData, &data); err == nil {
			return h.broadcaster.BroadcastPaymentRefunded(&data.Payment, data.RefundedAmount)
		}
	}

	return nil
}
