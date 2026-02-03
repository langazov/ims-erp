package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/events"
)

// SagaStatus represents the current status of a saga
type SagaStatus string

const (
	SagaStatusPending      SagaStatus = "pending"
	SagaStatusInProgress   SagaStatus = "in_progress"
	SagaStatusCompleted    SagaStatus = "completed"
	SagaStatusFailed       SagaStatus = "failed"
	SagaStatusCompensating SagaStatus = "compensating"
	SagaStatusCompensated  SagaStatus = "compensated"
)

// StepStatus represents the status of an individual saga step
type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusCompleted  StepStatus = "completed"
	StepStatusFailed     StepStatus = "failed"
	StepStatusSkipped    StepStatus = "skipped"
)

// SagaStep represents a single step in the saga
type SagaStep struct {
	Name        string
	Status      StepStatus
	StartedAt   *time.Time
	CompletedAt *time.Time
	Error       error
	Compensated bool
}

// OrderFulfillmentSaga orchestrates the order fulfillment process
type OrderFulfillmentSaga struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	OrderID      uuid.UUID
	CustomerID   uuid.UUID
	Items        []FulfillmentItem
	Status       SagaStatus
	Steps        []SagaStep
	WarehouseID  uuid.UUID
	OperationID  *uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
	ErrorMessage string
}

// FulfillmentItem represents an order item being fulfilled
type FulfillmentItem struct {
	ProductID        uuid.UUID
	VariantID        *uuid.UUID
	Quantity         int
	ReservedQuantity int
	PickedQuantity   int
	Status           string
	Reservations     []uuid.UUID
}

// NewOrderFulfillmentSaga creates a new order fulfillment saga
func NewOrderFulfillmentSaga(tenantID, orderID, customerID, warehouseID uuid.UUID, items []FulfillmentItem) *OrderFulfillmentSaga {
	now := time.Now().UTC()
	return &OrderFulfillmentSaga{
		ID:          uuid.New(),
		TenantID:    tenantID,
		OrderID:     orderID,
		CustomerID:  customerID,
		WarehouseID: warehouseID,
		Items:       items,
		Status:      SagaStatusPending,
		Steps: []SagaStep{
			{Name: "validate_stock", Status: StepStatusPending},
			{Name: "reserve_stock", Status: StepStatusPending},
			{Name: "create_picking_task", Status: StepStatusPending},
			{Name: "wait_for_picking", Status: StepStatusPending},
			{Name: "pack_items", Status: StepStatusPending},
			{Name: "ship_order", Status: StepStatusPending},
			{Name: "commit_reservations", Status: StepStatusPending},
			{Name: "send_notification", Status: StepStatusPending},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// OrderFulfillmentSagaHandler handles the execution of order fulfillment sagas
type OrderFulfillmentSagaHandler struct {
	inventoryRepo   domain.InventoryRepository
	reservationRepo domain.ReservationRepository
	operationRepo   domain.OperationRepository
	warehouseRepo   domain.WarehouseRepository
	publisher       events.Publisher
}

// NewOrderFulfillmentSagaHandler creates a new saga handler
func NewOrderFulfillmentSagaHandler(
	inventoryRepo domain.InventoryRepository,
	reservationRepo domain.ReservationRepository,
	operationRepo domain.OperationRepository,
	warehouseRepo domain.WarehouseRepository,
	publisher events.Publisher,
) *OrderFulfillmentSagaHandler {
	return &OrderFulfillmentSagaHandler{
		inventoryRepo:   inventoryRepo,
		reservationRepo: reservationRepo,
		operationRepo:   operationRepo,
		warehouseRepo:   warehouseRepo,
		publisher:       publisher,
	}
}

// Execute runs the order fulfillment saga
func (h *OrderFulfillmentSagaHandler) Execute(ctx context.Context, saga *OrderFulfillmentSaga, userID string) error {
	saga.Status = SagaStatusInProgress
	saga.UpdatedAt = time.Now().UTC()

	// Step 1: Validate Stock
	if err := h.validateStock(ctx, saga); err != nil {
		saga.Status = SagaStatusFailed
		saga.ErrorMessage = err.Error()
		saga.UpdatedAt = time.Now().UTC()
		return err
	}
	saga.completeStep("validate_stock")

	// Step 2: Reserve Stock
	if err := h.reserveStock(ctx, saga, userID); err != nil {
		saga.Status = SagaStatusFailed
		saga.ErrorMessage = err.Error()
		saga.UpdatedAt = time.Now().UTC()
		// Attempt compensation
		h.compensate(ctx, saga, userID)
		return err
	}
	saga.completeStep("reserve_stock")

	// Step 3: Create Picking Task
	if err := h.createPickingTask(ctx, saga, userID); err != nil {
		saga.Status = SagaStatusFailed
		saga.ErrorMessage = err.Error()
		saga.UpdatedAt = time.Now().UTC()
		h.compensate(ctx, saga, userID)
		return err
	}
	saga.completeStep("create_picking_task")

	// For now, we consider the saga successfully initiated up to this point
	// The remaining steps (wait_for_picking, pack, ship) are typically async
	saga.Status = SagaStatusInProgress
	saga.UpdatedAt = time.Now().UTC()

	return nil
}

// validateStock checks if all items have sufficient inventory
func (h *OrderFulfillmentSagaHandler) validateStock(ctx context.Context, saga *OrderFulfillmentSaga) error {
	for i, item := range saga.Items {
		inventory, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, item.ProductID, saga.WarehouseID)
		if err != nil {
			return fmt.Errorf("inventory not found for product %s: %w", item.ProductID, err)
		}

		if inventory.AvailableQty < item.Quantity {
			return fmt.Errorf("insufficient inventory for product %s: need %d, have %d",
				item.ProductID, item.Quantity, inventory.AvailableQty)
		}

		saga.Items[i].Status = "validated"
	}

	return nil
}

// reserveStock creates reservations for all order items
func (h *OrderFulfillmentSagaHandler) reserveStock(ctx context.Context, saga *OrderFulfillmentSaga, userID string) error {
	for i, item := range saga.Items {
		// Create reservation
		reservation := domain.NewStockReservation(
			saga.TenantID,
			item.ProductID,
			saga.WarehouseID,
			saga.OrderID,
			"order",
			item.Quantity,
		)

		if err := h.reservationRepo.Create(ctx, reservation); err != nil {
			return fmt.Errorf("failed to create reservation for product %s: %w", item.ProductID, err)
		}

		// Reserve inventory
		inventory, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, item.ProductID, saga.WarehouseID)
		if err != nil {
			return fmt.Errorf("inventory not found for product %s: %w", item.ProductID, err)
		}

		if err := inventory.Reserve(item.Quantity); err != nil {
			return fmt.Errorf("failed to reserve inventory for product %s: %w", item.ProductID, err)
		}

		if err := h.inventoryRepo.Update(ctx, inventory); err != nil {
			return fmt.Errorf("failed to update inventory for product %s: %w", item.ProductID, err)
		}

		// Store reservation ID
		saga.Items[i].Reservations = []uuid.UUID{reservation.ID}
		saga.Items[i].ReservedQuantity = item.Quantity
		saga.Items[i].Status = "reserved"

		// Publish event
		evt := events.NewStockReservedEvent(reservation, userID)
		if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
			// Log but don't fail
			_ = err
		}
	}

	return nil
}

// createPickingTask creates a warehouse operation for picking
func (h *OrderFulfillmentSagaHandler) createPickingTask(ctx context.Context, saga *OrderFulfillmentSaga, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Create picking operation
	operation, err := domain.NewWarehouseOperation(
		saga.TenantID,
		saga.WarehouseID,
		userUUID,
		domain.OperationTypePick,
		"order",
		saga.OrderID,
	)
	if err != nil {
		return fmt.Errorf("failed to create picking operation: %w", err)
	}

	// Add items to operation
	for _, item := range saga.Items {
		// Find location for item (in a real implementation, this would use location selection logic)
		locations, err := h.warehouseRepo.FindByTenant(ctx, saga.TenantID)
		if err != nil || len(locations) == 0 {
			return fmt.Errorf("no locations available for picking")
		}

		// For simplicity, use a placeholder location ID
		locationID := uuid.New()

		opItem := domain.OperationItem{
			ID:         uuid.New(),
			ProductID:  item.ProductID,
			VariantID:  item.VariantID,
			LocationID: locationID,
			Quantity:   item.Quantity,
			Status:     "pending",
		}
		operation.AddItem(opItem)
	}

	// Set high priority for the picking task
	operation.SetPriority(7)

	if err := h.operationRepo.Create(ctx, operation); err != nil {
		return fmt.Errorf("failed to create picking task: %w", err)
	}

	saga.OperationID = &operation.ID

	// Publish event
	evt := events.NewWarehouseOperationCreatedEvent(operation, userID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		_ = err
	}

	return nil
}

// compensate reverses the saga steps on failure
func (h *OrderFulfillmentSagaHandler) compensate(ctx context.Context, saga *OrderFulfillmentSaga, userID string) {
	saga.Status = SagaStatusCompensating
	saga.UpdatedAt = time.Now().UTC()

	// Release all reservations
	for _, item := range saga.Items {
		for _, reservationID := range item.Reservations {
			reservation, err := h.reservationRepo.FindByID(ctx, reservationID)
			if err != nil {
				continue
			}

			if reservation.Status == "active" {
				// Release inventory
				inventory, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, reservation.ProductID, reservation.WarehouseID)
				if err == nil {
					inventory.ReleaseReservation(reservation.Quantity)
					h.inventoryRepo.Update(ctx, inventory)
				}

				// Release reservation
				reservation.Release()
				h.reservationRepo.Update(ctx, reservation)

				// Publish event
				evt := events.NewReservationReleasedEvent(reservation, userID, "saga_compensation")
				h.publisher.PublishEvent(ctx, &evt.EventEnvelope)
			}
		}
	}

	// Cancel picking operation if created
	if saga.OperationID != nil {
		operation, err := h.operationRepo.FindByID(ctx, *saga.OperationID)
		if err == nil && operation.Status != "completed" {
			operation.Cancel("saga_compensation")
			h.operationRepo.Update(ctx, operation)

			evt := events.NewWarehouseOperationCancelledEvent(operation, userID)
			h.publisher.PublishEvent(ctx, &evt.EventEnvelope)
		}
	}

	saga.Status = SagaStatusCompensated
	saga.UpdatedAt = time.Now().UTC()
}

// CompletePicking completes the picking step of the saga
func (h *OrderFulfillmentSagaHandler) CompletePicking(ctx context.Context, saga *OrderFulfillmentSaga, userID string) error {
	saga.completeStep("wait_for_picking")

	// Step 5: Pack Items
	if err := h.packItems(ctx, saga, userID); err != nil {
		return err
	}
	saga.completeStep("pack_items")

	// Step 6: Ship Order
	if err := h.shipOrder(ctx, saga, userID); err != nil {
		return err
	}
	saga.completeStep("ship_order")

	// Step 7: Commit Reservations
	if err := h.commitReservations(ctx, saga, userID); err != nil {
		return err
	}
	saga.completeStep("commit_reservations")

	// Step 8: Send Notification
	saga.completeStep("send_notification")

	// Complete saga
	saga.Status = SagaStatusCompleted
	now := time.Now().UTC()
	saga.CompletedAt = &now
	saga.UpdatedAt = now

	return nil
}

// packItems creates a packing operation
func (h *OrderFulfillmentSagaHandler) packItems(ctx context.Context, saga *OrderFulfillmentSaga, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	operation, err := domain.NewWarehouseOperation(
		saga.TenantID,
		saga.WarehouseID,
		userUUID,
		domain.OperationTypePack,
		"order",
		saga.OrderID,
	)
	if err != nil {
		return fmt.Errorf("failed to create packing operation: %w", err)
	}

	// Add items
	for _, item := range saga.Items {
		opItem := domain.OperationItem{
			ID:        uuid.New(),
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			Status:    "pending",
		}
		operation.AddItem(opItem)
	}

	if err := h.operationRepo.Create(ctx, operation); err != nil {
		return fmt.Errorf("failed to create packing task: %w", err)
	}

	// Complete immediately (in reality, this would wait for packing)
	operation.Start()
	operation.Complete()
	h.operationRepo.Update(ctx, operation)

	evt := events.NewWarehouseOperationCompletedEvent(operation, userID)
	h.publisher.PublishEvent(ctx, &evt.EventEnvelope)

	return nil
}

// shipOrder creates a shipping operation
func (h *OrderFulfillmentSagaHandler) shipOrder(ctx context.Context, saga *OrderFulfillmentSaga, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	operation, err := domain.NewWarehouseOperation(
		saga.TenantID,
		saga.WarehouseID,
		userUUID,
		domain.OperationTypeShip,
		"order",
		saga.OrderID,
	)
	if err != nil {
		return fmt.Errorf("failed to create shipping operation: %w", err)
	}

	// Add items
	for _, item := range saga.Items {
		opItem := domain.OperationItem{
			ID:        uuid.New(),
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			Status:    "pending",
		}
		operation.AddItem(opItem)
	}

	if err := h.operationRepo.Create(ctx, operation); err != nil {
		return fmt.Errorf("failed to create shipping task: %w", err)
	}

	// Complete immediately
	operation.Start()
	operation.Complete()
	h.operationRepo.Update(ctx, operation)

	// Create shipping transactions
	for _, item := range saga.Items {
		transaction := domain.NewInventoryTransaction(
			saga.TenantID,
			item.ProductID,
			saga.WarehouseID,
			userUUID,
			domain.MovementTypeShipment,
			item.Quantity,
		)
		transaction.SetReference("order", saga.OrderID)
		// Transaction would be saved to repository

		// Update inventory
		inventory, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, item.ProductID, saga.WarehouseID)
		if err == nil {
			inventory.Ship(item.Quantity)
			h.inventoryRepo.Update(ctx, inventory)
		}

		evt := events.NewInventoryShippedEvent(transaction, userID)
		h.publisher.PublishEvent(ctx, &evt.EventEnvelope)
	}

	evt := events.NewWarehouseOperationCompletedEvent(operation, userID)
	h.publisher.PublishEvent(ctx, &evt.EventEnvelope)

	return nil
}

// commitReservations commits all reservations (deducts from inventory)
func (h *OrderFulfillmentSagaHandler) commitReservations(ctx context.Context, saga *OrderFulfillmentSaga, userID string) error {
	for _, item := range saga.Items {
		for _, reservationID := range item.Reservations {
			reservation, err := h.reservationRepo.FindByID(ctx, reservationID)
			if err != nil {
				continue
			}

			if reservation.Status == "active" {
				reservation.Fulfill()
				h.reservationRepo.Update(ctx, reservation)

				evt := events.NewReservationCommittedEvent(reservation, userID)
				h.publisher.PublishEvent(ctx, &evt.EventEnvelope)
			}
		}
	}

	return nil
}

// Helper methods

func (s *OrderFulfillmentSaga) completeStep(name string) {
	for i := range s.Steps {
		if s.Steps[i].Name == name {
			now := time.Now().UTC()
			s.Steps[i].Status = StepStatusCompleted
			s.Steps[i].CompletedAt = &now
			if s.Steps[i].StartedAt == nil {
				s.Steps[i].StartedAt = &now
			}
			break
		}
	}
}

func (s *OrderFulfillmentSaga) failStep(name string, err error) {
	for i := range s.Steps {
		if s.Steps[i].Name == name {
			s.Steps[i].Status = StepStatusFailed
			s.Steps[i].Error = err
			break
		}
	}
}

func (s *OrderFulfillmentSaga) getStep(name string) *SagaStep {
	for i := range s.Steps {
		if s.Steps[i].Name == name {
			return &s.Steps[i]
		}
	}
	return nil
}

// IsComplete returns true if all steps are completed
func (s *OrderFulfillmentSaga) IsComplete() bool {
	for _, step := range s.Steps {
		if step.Status != StepStatusCompleted && step.Status != StepStatusSkipped {
			return false
		}
	}
	return true
}

// GetCurrentStep returns the currently active step
func (s *OrderFulfillmentSaga) GetCurrentStep() *SagaStep {
	for i := range s.Steps {
		if s.Steps[i].Status == StepStatusInProgress {
			return &s.Steps[i]
		}
		if s.Steps[i].Status == StepStatusPending {
			return &s.Steps[i]
		}
	}
	return nil
}
