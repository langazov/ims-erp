package events

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	aggregateID := uuid.New().String()
	aggregateType := "warehouse"
	eventType := "created"
	tenantID := uuid.New().String()
	userID := uuid.New().String()
	data := map[string]interface{}{
		"name": "Test Warehouse",
	}

	event := NewEvent(aggregateID, aggregateType, eventType, tenantID, userID, data)

	assert.NotNil(t, event)
	assert.NotEmpty(t, event.ID)
	assert.Equal(t, eventType, event.Type)
	assert.Equal(t, aggregateID, event.AggregateID)
	assert.Equal(t, aggregateType, event.AggregateType)
	assert.Equal(t, tenantID, event.TenantID)
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, int64(1), event.Version)
	assert.Equal(t, "Test Warehouse", event.Data["name"])
	assert.NotZero(t, event.Timestamp)
	assert.NotEmpty(t, event.CorrelationID)
}

func TestEventEnvelope_WithCorrelationID(t *testing.T) {
	event := NewEvent(uuid.New().String(), "warehouse", "created", uuid.New().String(), uuid.New().String(), nil)
	correlationID := uuid.New().String()

	result := event.WithCorrelationID(correlationID)

	assert.Equal(t, correlationID, result.CorrelationID)
	assert.Same(t, event, result)
}

func TestEventEnvelope_WithCausationID(t *testing.T) {
	event := NewEvent(uuid.New().String(), "warehouse", "created", uuid.New().String(), uuid.New().String(), nil)
	causationID := uuid.New().String()

	result := event.WithCausationID(causationID)

	assert.Equal(t, causationID, result.CausationID)
	assert.Same(t, event, result)
}

func TestEventEnvelope_WithMetadata(t *testing.T) {
	event := NewEvent(uuid.New().String(), "warehouse", "created", uuid.New().String(), uuid.New().String(), nil)

	result := event.WithMetadata("key1", "value1").WithMetadata("key2", "value2")

	assert.Equal(t, "value1", result.Metadata["key1"])
	assert.Equal(t, "value2", result.Metadata["key2"])
	assert.Same(t, event, result)
}

func TestEventEnvelope_IncrementVersion(t *testing.T) {
	event := NewEvent(uuid.New().String(), "warehouse", "created", uuid.New().String(), uuid.New().String(), nil)
	initialVersion := event.Version

	event.IncrementVersion()

	assert.Equal(t, initialVersion+1, event.Version)
}

func TestEventEnvelope_Subject(t *testing.T) {
	event := NewEvent(uuid.New().String(), "warehouse", "created", uuid.New().String(), uuid.New().String(), nil)

	subject := event.Subject()

	assert.Equal(t, "evt.warehouse.created", subject)
}

func TestEventEnvelope_ToJSON(t *testing.T) {
	event := NewEvent(uuid.New().String(), "warehouse", "created", uuid.New().String(), uuid.New().String(), map[string]interface{}{
		"name": "Test",
	})

	data, err := event.ToJSON()

	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "warehouse")
	assert.Contains(t, string(data), "created")
}

func TestEventFromJSON(t *testing.T) {
	original := NewEvent(uuid.New().String(), "warehouse", "created", uuid.New().String(), uuid.New().String(), map[string]interface{}{
		"name": "Test Warehouse",
	})

	jsonData, _ := original.ToJSON()

	restored, err := EventFromJSON(jsonData)

	require.NoError(t, err)
	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.AggregateID, restored.AggregateID)
	assert.Equal(t, original.TenantID, restored.TenantID)
}

func TestEventHandlerRegistry_Register(t *testing.T) {
	registry := NewEventHandlerRegistry()
	eventType := "warehouse.created"

	handler := func(ctx context.Context, event *EventEnvelope) error {
		return nil
	}

	registry.Register(eventType, handler)

	handlers := registry.GetHandlers(eventType)
	assert.Len(t, handlers, 1)
}

func TestEventHandlerRegistry_Handle(t *testing.T) {
	registry := NewEventHandlerRegistry()
	eventType := "warehouse.created"
	handlerCalled := false

	handler := func(ctx context.Context, event *EventEnvelope) error {
		handlerCalled = true
		return nil
	}

	registry.Register(eventType, handler)

	event := NewEvent(uuid.New().String(), "warehouse", eventType, uuid.New().String(), uuid.New().String(), nil)
	errors := registry.Handle(context.Background(), event)

	assert.True(t, handlerCalled)
	assert.Empty(t, errors)
}

func TestEventHandlerRegistry_HandleMultipleHandlers(t *testing.T) {
	registry := NewEventHandlerRegistry()
	eventType := "warehouse.created"
	handler1Called := false
	handler2Called := false

	handler1 := func(ctx context.Context, event *EventEnvelope) error {
		handler1Called = true
		return nil
	}
	handler2 := func(ctx context.Context, event *EventEnvelope) error {
		handler2Called = true
		return nil
	}

	registry.Register(eventType, handler1)
	registry.Register(eventType, handler2)

	event := NewEvent(uuid.New().String(), "warehouse", eventType, uuid.New().String(), uuid.New().String(), nil)
	registry.Handle(context.Background(), event)

	assert.True(t, handler1Called)
	assert.True(t, handler2Called)
}

func TestEventHandlerRegistry_HandleErrors(t *testing.T) {
	registry := NewEventHandlerRegistry()
	eventType := "warehouse.created"
	handlerError := assert.AnError

	handler := func(ctx context.Context, event *EventEnvelope) error {
		return handlerError
	}

	registry.Register(eventType, handler)

	event := NewEvent(uuid.New().String(), "warehouse", eventType, uuid.New().String(), uuid.New().String(), nil)
	errors := registry.Handle(context.Background(), event)

	assert.Len(t, errors, 1)
	assert.Equal(t, handlerError, errors[0])
}

func TestNewBaseEvent(t *testing.T) {
	eventType := "created"
	aggregateID := uuid.New().String()
	tenantID := uuid.New().String()
	userID := uuid.New().String()

	event := NewBaseEvent(eventType, aggregateID, tenantID, userID)

	assert.Equal(t, eventType, event.EventType())
	assert.Equal(t, aggregateID, event.AggregateID())
	assert.Equal(t, tenantID, event.TenantID())
	assert.Equal(t, userID, event.UserID())
	assert.NotZero(t, event.Timestamp())
	assert.NotNil(t, event.Data())
}

func TestBaseEvent_With(t *testing.T) {
	event := NewBaseEvent("created", uuid.New().String(), uuid.New().String(), uuid.New().String())

	result := event.With("name", "Test").With("capacity", 1000)

	assert.Equal(t, "Test", result.Data()["name"])
	assert.Equal(t, 1000, result.Data()["capacity"])
}

func TestBaseEvent_ToEnvelope(t *testing.T) {
	event := NewBaseEvent("created", uuid.New().String(), uuid.New().String(), uuid.New().String())
	event.With("name", "Test")

	envelope := event.ToEnvelope()

	assert.Equal(t, "created", envelope.Type)
	assert.Equal(t, event.AggregateID(), envelope.AggregateID)
	assert.Equal(t, event.TenantID(), envelope.TenantID)
	assert.Equal(t, event.UserID(), envelope.UserID)
	assert.Equal(t, "Test", envelope.Data["name"])
}

func TestNewWarehouseCreatedEvent(t *testing.T) {
	tenantID := uuid.New()
	warehouseID := uuid.New()
	userID := uuid.New()

	warehouse := &domain.Warehouse{
		ID:           warehouseID,
		TenantID:     tenantID,
		Name:         "Main Warehouse",
		Code:         "WH-001",
		Type:         domain.WarehouseTypeMain,
		Capacity:     10000,
		ContactEmail: "test@example.com",
		ContactPhone: "+1234567890",
		IsPrimary:    true,
	}

	event := NewWarehouseCreatedEvent(warehouse, userID.String())

	assert.Equal(t, "warehouse.created", event.Type)
	assert.Equal(t, warehouseID.String(), event.AggregateID)
	assert.Equal(t, "Warehouse", event.AggregateType)
	assert.Equal(t, tenantID.String(), event.TenantID)
	assert.Equal(t, userID.String(), event.UserID)
	assert.Equal(t, "Main Warehouse", event.Data["name"])
	assert.Equal(t, "WH-001", event.Data["code"])
	assert.Equal(t, "main", event.Data["type"])
	assert.Equal(t, 10000, event.Data["capacity"])
}

func TestNewWarehouseActivatedEvent(t *testing.T) {
	tenantID := uuid.New()
	warehouseID := uuid.New()
	userID := uuid.New()

	warehouse := &domain.Warehouse{
		ID:       warehouseID,
		TenantID: tenantID,
		Name:     "Test WH",
		Code:     "WH-002",
		Type:     domain.WarehouseTypeDistribution,
		IsActive: true,
	}

	event := NewWarehouseActivatedEvent(warehouse, userID.String())

	assert.Equal(t, "warehouse.activated", event.Type)
	assert.Equal(t, warehouseID.String(), event.AggregateID)
	assert.Equal(t, "Warehouse", event.AggregateType)
}

func TestNewWarehouseDeactivatedEvent(t *testing.T) {
	tenantID := uuid.New()
	warehouseID := uuid.New()
	userID := uuid.New()

	warehouse := &domain.Warehouse{
		ID:       warehouseID,
		TenantID: tenantID,
		Name:     "Test WH",
		Code:     "WH-002",
		Type:     domain.WarehouseTypeDistribution,
		IsActive: false,
	}

	event := NewWarehouseDeactivatedEvent(warehouse, userID.String())

	assert.Equal(t, "warehouse.deactivated", event.Type)
	assert.Equal(t, warehouseID.String(), event.AggregateID)
}

func TestNewLocationCreatedEvent(t *testing.T) {
	tenantID := uuid.New()
	warehouseID := uuid.New()
	locationID := uuid.New()
	userID := uuid.New()

	location := &domain.WarehouseLocation{
		ID:          locationID,
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		Name:        "Bin A1",
		Code:        "A-01-01-01",
		Type:        "bin",
		Zone:        "A",
		Aisle:       "01",
		Rack:        "01",
		Bin:         "01",
		Capacity:    100,
		IsActive:    true,
	}

	event := NewLocationCreatedEvent(location, userID.String())

	assert.Equal(t, "location.created", event.Type)
	assert.Equal(t, locationID.String(), event.AggregateID)
	assert.Equal(t, "Location", event.AggregateType)
	assert.Equal(t, tenantID.String(), event.TenantID)
	assert.Equal(t, "Bin A1", event.Data["name"])
	assert.Equal(t, "A-01-01-01", event.Data["code"])
	assert.Equal(t, "A", event.Data["zone"])
}

func TestNewWarehouseOperationCreatedEvent(t *testing.T) {
	tenantID := uuid.New()
	warehouseID := uuid.New()
	operationID := uuid.New()
	userID := uuid.New()
	productID := uuid.New()
	referenceID := uuid.New()
	now := time.Now()

	operation := &domain.WarehouseOperation{
		ID:            operationID,
		TenantID:      tenantID,
		WarehouseID:   warehouseID,
		Type:          domain.OperationTypeReceipt,
		Status:        "pending",
		Priority:      5,
		ReferenceType: "purchase_order",
		ReferenceID:   referenceID,
		Items:         []domain.OperationItem{},
		Notes:         "Receiving shipment",
		StartedAt:     nil,
		CompletedAt:   nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	operation.Items = append(operation.Items, domain.OperationItem{
		ID:         uuid.New(),
		ProductID:  productID,
		LocationID: uuid.New(),
		Quantity:   100,
		Status:     "pending",
	})

	event := NewWarehouseOperationCreatedEvent(operation, userID.String())

	assert.Equal(t, "warehouse.operation.created", event.Type)
	assert.Equal(t, operationID.String(), event.AggregateID)
	assert.Equal(t, warehouseID, event.Data["warehouseId"])
	assert.Equal(t, string(domain.OperationTypeReceipt), event.Data["type"])
	assert.Equal(t, "purchase_order", event.Data["referenceType"])
	assert.Equal(t, referenceID, event.Data["referenceId"])
	assert.Equal(t, 5, event.Data["priority"])
	assert.Equal(t, 1, event.Data["itemCount"])
}

func TestNewWarehouseOperationStartedEvent(t *testing.T) {
	tenantID := uuid.New()
	warehouseID := uuid.New()
	operationID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	operation := &domain.WarehouseOperation{
		ID:          operationID,
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		Type:        domain.OperationTypeReceipt,
		Status:      "in_progress",
		StartedAt:   &now,
	}

	event := NewWarehouseOperationStartedEvent(operation, userID.String())

	assert.Equal(t, "warehouse.operation.started", event.Type)
	assert.Equal(t, operationID.String(), event.AggregateID)
}

func TestNewWarehouseOperationCompletedEvent(t *testing.T) {
	tenantID := uuid.New()
	warehouseID := uuid.New()
	operationID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	operation := &domain.WarehouseOperation{
		ID:          operationID,
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		Type:        domain.OperationTypePick,
		Status:      "completed",
		StartedAt:   &now,
		CompletedAt: &now,
		Items:       []domain.OperationItem{},
	}

	event := NewWarehouseOperationCompletedEvent(operation, userID.String())

	assert.Equal(t, "warehouse.operation.completed", event.Type)
	assert.Equal(t, operationID.String(), event.AggregateID)
}

func TestNewWarehouseOperationCancelledEvent(t *testing.T) {
	tenantID := uuid.New()
	warehouseID := uuid.New()
	operationID := uuid.New()
	userID := uuid.New()

	operation := &domain.WarehouseOperation{
		ID:          operationID,
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		Type:        domain.OperationTypeReceipt,
		Status:      "cancelled",
		Notes:       "Duplicate order",
	}

	event := NewWarehouseOperationCancelledEvent(operation, userID.String())

	assert.Equal(t, "warehouse.operation.cancelled", event.Type)
	assert.Equal(t, operationID.String(), event.AggregateID)
	assert.Equal(t, "Duplicate order", event.Data["reason"])
}

func TestNewStockReservedEvent(t *testing.T) {
	tenantID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()
	reservationID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	reservation := &domain.StockReservation{
		ID:            reservationID,
		TenantID:      tenantID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		Quantity:      100,
		ReferenceType: "order",
		ReferenceID:   uuid.New(),
		ExpiresAt:     &expiresAt,
		Status:        "pending",
		CreatedAt:     now,
	}

	event := NewStockReservedEvent(reservation, userID.String())

	assert.Equal(t, "inventory.stock.reserved", event.Type)
	assert.Equal(t, reservationID.String(), event.AggregateID)
	assert.Equal(t, 100, event.Data["quantity"])
	assert.Equal(t, "order", event.Data["referenceType"])
}

func TestNewInventoryAdjustedEvent(t *testing.T) {
	tenantID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()
	locationID := uuid.New()
	adjustmentID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	adjustment := &domain.InventoryAdjustment{
		ID:             adjustmentID,
		TenantID:       tenantID,
		ProductID:      productID,
		WarehouseID:    warehouseID,
		LocationID:     &locationID,
		AdjustmentType: "count_adjustment",
		Reason:         "Physical count",
		Quantity:       10,
		ReferenceType:  "adjustment",
		ReferenceID:    uuid.New(),
		PerformedBy:    userID,
		CreatedAt:      now,
	}

	event := NewInventoryAdjustedEvent(adjustment, 90, 100, userID.String())

	assert.Equal(t, "inventory.adjusted", event.Type)
	assert.Equal(t, adjustmentID.String(), event.AggregateID)
	assert.Equal(t, 90, event.Data["previousQuantity"])
	assert.Equal(t, 100, event.Data["newQuantity"])
}

func TestNewInventoryReceivedEvent(t *testing.T) {
	tenantID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()
	locationID := uuid.New()
	transactionID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	transaction := &domain.InventoryTransaction{
		ID:            transactionID,
		TenantID:      tenantID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		ToLocationID:  &locationID,
		Quantity:      50,
		ReferenceType: "purchase_order",
		ReferenceID:   uuid.New(),
		CreatedAt:     now,
	}

	event := NewInventoryReceivedEvent(transaction, "10.50", userID.String())

	assert.Equal(t, "inventory.received", event.Type)
	assert.Equal(t, transactionID.String(), event.AggregateID)
	assert.Equal(t, 50, event.Data["quantity"])
	assert.Equal(t, "10.50", event.Data["unitCost"])
}

func TestNewInventoryShippedEvent(t *testing.T) {
	tenantID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()
	locationID := uuid.New()
	transactionID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	transaction := &domain.InventoryTransaction{
		ID:             transactionID,
		TenantID:       tenantID,
		ProductID:      productID,
		WarehouseID:    warehouseID,
		FromLocationID: &locationID,
		Quantity:       25,
		ReferenceType:  "sales_order",
		ReferenceID:    uuid.New(),
		CreatedAt:      now,
	}

	event := NewInventoryShippedEvent(transaction, userID.String())

	assert.Equal(t, "inventory.shipped", event.Type)
	assert.Equal(t, transactionID.String(), event.AggregateID)
	assert.Equal(t, 25, event.Data["quantity"])
	assert.Equal(t, "sales_order", event.Data["referenceType"])
}
