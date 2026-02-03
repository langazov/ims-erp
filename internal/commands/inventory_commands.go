package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/events"
	"github.com/shopspring/decimal"
)

// Command types for inventory operations
type ReserveStock struct {
	ProductID     uuid.UUID
	VariantID     *uuid.UUID
	WarehouseID   uuid.UUID
	ReferenceType string
	ReferenceID   uuid.UUID
	Quantity      int
	ExpiresAt     *string
}

type ReleaseReservation struct {
	ReservationID uuid.UUID
	Reason        string
}

type CommitReservation struct {
	ReservationID uuid.UUID
}

type ReceiveInventory struct {
	ProductID     uuid.UUID
	VariantID     *uuid.UUID
	WarehouseID   uuid.UUID
	LocationID    uuid.UUID
	Quantity      int
	UnitCost      string
	LotNumber     string
	SerialNumber  string
	ReferenceType string
	ReferenceID   uuid.UUID
}

type ShipInventory struct {
	ProductID     uuid.UUID
	WarehouseID   uuid.UUID
	LocationID    uuid.UUID
	Quantity      int
	ReferenceType string
	ReferenceID   uuid.UUID
}

type TransferInventory struct {
	ProductID       uuid.UUID
	VariantID       *uuid.UUID
	FromWarehouseID uuid.UUID
	FromLocationID  uuid.UUID
	ToWarehouseID   uuid.UUID
	ToLocationID    uuid.UUID
	Quantity        int
	ReferenceType   string
	ReferenceID     uuid.UUID
}

type AdjustInventory struct {
	ProductID      uuid.UUID
	VariantID      *uuid.UUID
	WarehouseID    uuid.UUID
	LocationID     *uuid.UUID
	AdjustmentType string
	Quantity       int
	Reason         string
	ReferenceType  string
	ReferenceID    uuid.UUID
}

type CycleCountInventory struct {
	ProductID   uuid.UUID
	VariantID   *uuid.UUID
	WarehouseID uuid.UUID
	LocationID  uuid.UUID
	CountedQty  int
	Notes       string
}

// InventoryCommandHandler handles inventory-related commands
type InventoryCommandHandler struct {
	inventoryRepo   domain.InventoryRepository
	warehouseRepo   domain.WarehouseRepository
	locationRepo    domain.LocationRepository
	reservationRepo domain.ReservationRepository
	transactionRepo domain.TransactionRepository
	publisher       events.Publisher
}

func NewInventoryCommandHandler(
	inventoryRepo domain.InventoryRepository,
	warehouseRepo domain.WarehouseRepository,
	locationRepo domain.LocationRepository,
	reservationRepo domain.ReservationRepository,
	transactionRepo domain.TransactionRepository,
	publisher events.Publisher,
) *InventoryCommandHandler {
	return &InventoryCommandHandler{
		inventoryRepo:   inventoryRepo,
		warehouseRepo:   warehouseRepo,
		locationRepo:    locationRepo,
		reservationRepo: reservationRepo,
		transactionRepo: transactionRepo,
		publisher:       publisher,
	}
}

func (h *InventoryCommandHandler) HandleReserveStock(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input ReserveStock
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Find inventory item to check availability
	item, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, input.ProductID, input.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	if item.TenantID != tenantID {
		return nil, fmt.Errorf("inventory belongs to different tenant")
	}

	// Reserve stock on the item
	if err := item.Reserve(input.Quantity); err != nil {
		return nil, err
	}

	// Create reservation record
	reservation := domain.NewStockReservation(
		tenantID,
		input.ProductID,
		input.WarehouseID,
		input.ReferenceID,
		input.ReferenceType,
		input.Quantity,
	)

	if input.ExpiresAt != nil {
		// Parse expiration date if provided
		// For simplicity, we'll set it to nil for now
		reservation.ExpiresAt = nil
	}

	if err := h.reservationRepo.Create(ctx, reservation); err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Update inventory
	if err := h.inventoryRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Publish event
	evt := events.NewStockReservedEvent(reservation, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    reservation,
		Events:  []interface{}{evt},
	}, nil
}

func (h *InventoryCommandHandler) HandleReleaseReservation(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input ReleaseReservation
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	reservation, err := h.reservationRepo.FindByID(ctx, input.ReservationID)
	if err != nil {
		return nil, fmt.Errorf("reservation not found: %w", err)
	}

	if reservation.Status != "active" {
		return nil, fmt.Errorf("reservation is not active")
	}

	// Find inventory item
	item, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, reservation.ProductID, reservation.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	// Release reservation on inventory
	item.ReleaseReservation(reservation.Quantity)

	// Release reservation record
	reservation.Release()

	if err := h.reservationRepo.Update(ctx, reservation); err != nil {
		return nil, fmt.Errorf("failed to update reservation: %w", err)
	}

	if err := h.inventoryRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Publish event
	evt := events.NewReservationReleasedEvent(reservation, cmd.UserID, input.Reason)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    reservation,
		Events:  []interface{}{evt},
	}, nil
}

func (h *InventoryCommandHandler) HandleCommitReservation(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input CommitReservation
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	reservation, err := h.reservationRepo.FindByID(ctx, input.ReservationID)
	if err != nil {
		return nil, fmt.Errorf("reservation not found: %w", err)
	}

	if reservation.Status != "active" {
		return nil, fmt.Errorf("reservation is not active")
	}

	// Find inventory item
	item, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, reservation.ProductID, reservation.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	// Commit reservation (deduct from inventory)
	if err := item.Ship(reservation.Quantity); err != nil {
		return nil, err
	}

	// Mark reservation as fulfilled
	reservation.Fulfill()

	if err := h.reservationRepo.Update(ctx, reservation); err != nil {
		return nil, fmt.Errorf("failed to update reservation: %w", err)
	}

	if err := h.inventoryRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Publish event
	evt := events.NewReservationCommittedEvent(reservation, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    reservation,
		Events:  []interface{}{evt},
	}, nil
}

func (h *InventoryCommandHandler) HandleReceiveInventory(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input ReceiveInventory
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	unitCost, err := decimal.NewFromString(input.UnitCost)
	if err != nil {
		return nil, fmt.Errorf("invalid unit cost: %w", err)
	}

	// Find or create inventory item
	item, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, input.ProductID, input.WarehouseID)
	if err != nil {
		// Create new inventory item
		item = domain.NewInventoryItem(
			tenantID,
			input.ProductID,
			input.WarehouseID,
			"", // SKU will be populated from product
			0,
			unitCost,
		)
		item.LocationID = input.LocationID
		item.LotNumber = input.LotNumber
		item.SerialNumber = input.SerialNumber
	}

	// Receive inventory
	item.Receive(input.Quantity, unitCost)

	// Create transaction record
	transaction := domain.NewInventoryTransaction(
		tenantID,
		input.ProductID,
		input.WarehouseID,
		userID,
		domain.MovementTypeReceipt,
		input.Quantity,
	)
	transaction.SetReference(input.ReferenceType, input.ReferenceID)

	if err := h.inventoryRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	if err := h.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Publish event
	evt := events.NewInventoryReceivedEvent(transaction, unitCost.String(), cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    item,
		Events:  []interface{}{evt},
	}, nil
}

func (h *InventoryCommandHandler) HandleShipInventory(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input ShipInventory
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Find inventory item
	item, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, input.ProductID, input.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	// Ship inventory
	if err := item.Ship(input.Quantity); err != nil {
		return nil, err
	}

	// Create transaction record
	transaction := domain.NewInventoryTransaction(
		tenantID,
		input.ProductID,
		input.WarehouseID,
		userID,
		domain.MovementTypeShipment,
		input.Quantity,
	)
	transaction.SetReference(input.ReferenceType, input.ReferenceID)

	if err := h.inventoryRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	if err := h.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Publish event
	evt := events.NewInventoryShippedEvent(transaction, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    item,
		Events:  []interface{}{evt},
	}, nil
}

func (h *InventoryCommandHandler) HandleAdjustInventory(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input AdjustInventory
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Find inventory item
	item, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, input.ProductID, input.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	previousQty := item.Quantity

	// Adjust inventory
	item.Adjust(input.Quantity, input.Reason)

	// Create adjustment record
	adjustment := &domain.InventoryAdjustment{
		ID:             uuid.New(),
		TenantID:       tenantID,
		ProductID:      input.ProductID,
		VariantID:      input.VariantID,
		WarehouseID:    input.WarehouseID,
		LocationID:     input.LocationID,
		AdjustmentType: input.AdjustmentType,
		Quantity:       input.Quantity,
		Reason:         input.Reason,
		ReferenceType:  input.ReferenceType,
		ReferenceID:    input.ReferenceID,
		PerformedBy:    userID,
	}

	if err := h.inventoryRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Publish event
	evt := events.NewInventoryAdjustedEvent(adjustment, previousQty, item.Quantity, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    item,
		Events:  []interface{}{evt},
	}, nil
}

func (h *InventoryCommandHandler) HandleCycleCountInventory(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input CycleCountInventory
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Find inventory item
	item, err := h.inventoryRepo.FindByProductAndWarehouse(ctx, input.ProductID, input.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	previousQty := item.Quantity

	// Perform cycle count
	item.Count(input.CountedQty)

	// Create adjustment record for the difference
	adjustment := &domain.InventoryAdjustment{
		ID:             uuid.New(),
		TenantID:       tenantID,
		ProductID:      input.ProductID,
		VariantID:      input.VariantID,
		WarehouseID:    input.WarehouseID,
		LocationID:     &input.LocationID,
		AdjustmentType: "cycle_count",
		Quantity:       input.CountedQty - previousQty,
		Reason:         input.Notes,
		ReferenceType:  "cycle_count",
		ReferenceID:    uuid.New(), // Generate a reference ID for the count
		PerformedBy:    userID,
	}

	if err := h.inventoryRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Publish event
	evt := events.NewInventoryAdjustedEvent(adjustment, previousQty, item.Quantity, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    item,
		Events:  []interface{}{evt},
	}, nil
}
