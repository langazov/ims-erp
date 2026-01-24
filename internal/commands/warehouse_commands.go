package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/events"
)

type CreateWarehouse struct {
	Name           string
	Code           string
	Type           string
	Address        domain.Address
	Capacity       int
	ManagerID      *uuid.UUID
	ContactEmail   string
	ContactPhone   string
	OperatingHours string
}

type UpdateWarehouse struct {
	ID             uuid.UUID
	Name           *string
	Address        *domain.Address
	Capacity       *int
	ManagerID      *uuid.UUID
	ContactEmail   *string
	ContactPhone   *string
	OperatingHours *string
}

type ActivateWarehouse struct {
	ID uuid.UUID
}

type DeactivateWarehouse struct {
	ID uuid.UUID
}

type CreateLocation struct {
	WarehouseID uuid.UUID
	Name        string
	Code        string
	Type        string
	Zone        string
	Aisle       string
	Rack        string
	Bin         string
	Capacity    int
}

type UpdateLocation struct {
	ID       uuid.UUID
	Name     *string
	Capacity *int
	IsActive *bool
}

type CreateWarehouseOperation struct {
	WarehouseID   uuid.UUID
	Type          string
	ReferenceType string
	ReferenceID   uuid.UUID
	Priority      int
	Items         []OperationItemInput
	Notes         string
}

type OperationItemInput struct {
	ProductID  uuid.UUID
	VariantID  *uuid.UUID
	LocationID uuid.UUID
	Quantity   int
}

type StartWarehouseOperation struct {
	ID uuid.UUID
}

type CompleteWarehouseOperation struct {
	ID             uuid.UUID
	CompletedItems []CompletedItemInput
}

type CompletedItemInput struct {
	ItemID   uuid.UUID
	Quantity int
}

type CancelWarehouseOperation struct {
	ID     uuid.UUID
	Reason string
}

type WarehouseCommandHandler struct {
	warehouseRepo domain.WarehouseRepository
	locationRepo  domain.LocationRepository
	operationRepo domain.OperationRepository
	publisher     events.Publisher
}

func NewWarehouseCommandHandler(
	warehouseRepo domain.WarehouseRepository,
	locationRepo domain.LocationRepository,
	operationRepo domain.OperationRepository,
	publisher events.Publisher,
) *WarehouseCommandHandler {
	return &WarehouseCommandHandler{
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
		operationRepo: operationRepo,
		publisher:     publisher,
	}
}

func (h *WarehouseCommandHandler) HandleCreateWarehouse(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input CreateWarehouse
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	warehouseType := domain.WarehouseType(input.Type)
	if !warehouseType.IsValid() {
		return nil, domain.ErrInvalidWarehouseType
	}

	warehouse := domain.NewWarehouse(tenantID, input.Name, input.Code, warehouseType)
	warehouse.Address = input.Address
	warehouse.Capacity = input.Capacity
	warehouse.ManagerID = input.ManagerID
	warehouse.ContactEmail = input.ContactEmail
	warehouse.ContactPhone = input.ContactPhone
	warehouse.OperatingHours = input.OperatingHours

	if err := h.warehouseRepo.Create(ctx, warehouse); err != nil {
		return nil, fmt.Errorf("failed to create warehouse: %w", err)
	}

	evt := events.NewWarehouseCreatedEvent(warehouse, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    warehouse,
		Events:  []interface{}{evt},
	}, nil
}

func (h *WarehouseCommandHandler) HandleUpdateWarehouse(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input UpdateWarehouse
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	warehouse, err := h.warehouseRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("warehouse not found: %w", err)
	}

	if warehouse.TenantID != tenantID {
		return nil, fmt.Errorf("warehouse belongs to different tenant")
	}

	if input.Name != nil {
		warehouse.Name = *input.Name
	}
	if input.Address != nil {
		warehouse.Address = *input.Address
	}
	if input.Capacity != nil {
		if float64(*input.Capacity) < warehouse.CurrentUtilization {
			return nil, domain.ErrCapacityCannotBeLessThanUsed
		}
		warehouse.Capacity = *input.Capacity
	}
	if input.ManagerID != nil {
		warehouse.ManagerID = input.ManagerID
	}
	if input.ContactEmail != nil {
		warehouse.ContactEmail = *input.ContactEmail
	}
	if input.ContactPhone != nil {
		warehouse.ContactPhone = *input.ContactPhone
	}
	if input.OperatingHours != nil {
		warehouse.OperatingHours = *input.OperatingHours
	}

	if err := h.warehouseRepo.Update(ctx, warehouse); err != nil {
		return nil, fmt.Errorf("failed to update warehouse: %w", err)
	}

	evt := events.NewWarehouseUpdatedEvent(warehouse, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    warehouse,
		Events:  []interface{}{evt},
	}, nil
}

func (h *WarehouseCommandHandler) HandleActivateWarehouse(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input ActivateWarehouse
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	warehouse, err := h.warehouseRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("warehouse not found: %w", err)
	}

	if warehouse.TenantID != tenantID {
		return nil, fmt.Errorf("warehouse belongs to different tenant")
	}

	if !warehouse.IsActive && warehouse.Type == domain.WarehouseType("closed") {
		return nil, domain.ErrCannotActivateClosedWarehouse
	}

	warehouse.Activate()

	if err := h.warehouseRepo.Update(ctx, warehouse); err != nil {
		return nil, fmt.Errorf("failed to activate warehouse: %w", err)
	}

	evt := events.NewWarehouseActivatedEvent(warehouse, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    warehouse,
		Events:  []interface{}{evt},
	}, nil
}

func (h *WarehouseCommandHandler) HandleDeactivateWarehouse(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input DeactivateWarehouse
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	warehouse, err := h.warehouseRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("warehouse not found: %w", err)
	}

	if warehouse.TenantID != tenantID {
		return nil, fmt.Errorf("warehouse belongs to different tenant")
	}

	locations, err := h.locationRepo.FindByWarehouse(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check locations: %w", err)
	}

	activeLocations := 0
	for _, loc := range locations {
		if loc.IsActive {
			activeLocations++
		}
	}

	if activeLocations > 0 {
		return nil, domain.ErrCannotDeactivateWarehouseWithLocations
	}

	warehouse.Deactivate()

	if err := h.warehouseRepo.Update(ctx, warehouse); err != nil {
		return nil, fmt.Errorf("failed to deactivate warehouse: %w", err)
	}

	evt := events.NewWarehouseDeactivatedEvent(warehouse, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    warehouse,
		Events:  []interface{}{evt},
	}, nil
}

func (h *WarehouseCommandHandler) HandleCreateLocation(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input CreateLocation
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	warehouse, err := h.warehouseRepo.FindByID(ctx, input.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("warehouse not found: %w", err)
	}

	if warehouse.TenantID != tenantID {
		return nil, fmt.Errorf("warehouse belongs to different tenant")
	}

	if input.Zone == "" || input.Aisle == "" || input.Rack == "" || input.Bin == "" {
		return nil, domain.ErrLocationPathRequired
	}

	existingPath, err := h.locationRepo.FindByPath(ctx, input.WarehouseID, input.Zone, input.Aisle, input.Rack, input.Bin)
	if err == nil && existingPath != nil {
		return nil, fmt.Errorf("location path already exists")
	}

	now := time.Now().UTC()
	location := &domain.WarehouseLocation{
		ID:           uuid.New(),
		TenantID:     tenantID,
		WarehouseID:  input.WarehouseID,
		Name:         input.Name,
		Code:         input.Code,
		Type:         input.Type,
		Zone:         input.Zone,
		Aisle:        input.Aisle,
		Rack:         input.Rack,
		Bin:          input.Bin,
		Capacity:     input.Capacity,
		CurrentStock: 0,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.locationRepo.Create(ctx, location); err != nil {
		return nil, fmt.Errorf("failed to create location: %w", err)
	}

	evt := events.NewLocationCreatedEvent(location, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    location,
		Events:  []interface{}{evt},
	}, nil
}

func (h *WarehouseCommandHandler) HandleCreateWarehouseOperation(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input CreateWarehouseOperation
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

	operationType := domain.OperationType(input.Type)
	if !operationType.IsValid() {
		return nil, domain.ErrInvalidOperationType
	}

	operation, err := domain.NewWarehouseOperation(
		tenantID, input.WarehouseID, userID,
		operationType, input.ReferenceType, input.ReferenceID,
	)
	if err != nil {
		return nil, err
	}

	operation.Priority = input.Priority
	operation.Notes = input.Notes

	for _, itemInput := range input.Items {
		item := domain.OperationItem{
			ID:         uuid.New(),
			ProductID:  itemInput.ProductID,
			VariantID:  itemInput.VariantID,
			LocationID: itemInput.LocationID,
			Quantity:   itemInput.Quantity,
			Status:     "pending",
		}
		operation.AddItem(item)
	}

	if err := h.operationRepo.Create(ctx, operation); err != nil {
		return nil, fmt.Errorf("failed to create operation: %w", err)
	}

	evt := events.NewWarehouseOperationCreatedEvent(operation, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    operation,
		Events:  []interface{}{evt},
	}, nil
}

func (h *WarehouseCommandHandler) HandleStartWarehouseOperation(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input StartWarehouseOperation
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	operation, err := h.operationRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("operation not found: %w", err)
	}

	if operation.Status != "pending" {
		return nil, fmt.Errorf("operation already started")
	}

	operation.Start()

	if err := h.operationRepo.Update(ctx, operation); err != nil {
		return nil, fmt.Errorf("failed to start operation: %w", err)
	}

	evt := events.NewWarehouseOperationStartedEvent(operation, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    operation,
		Events:  []interface{}{evt},
	}, nil
}

func (h *WarehouseCommandHandler) HandleCompleteWarehouseOperation(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input CompleteWarehouseOperation
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	operation, err := h.operationRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("operation not found: %w", err)
	}

	for _, completed := range input.CompletedItems {
		if err := operation.CompleteItem(completed.ItemID, completed.Quantity); err != nil {
			return nil, err
		}
	}

	if operation.IsComplete() {
		operation.Complete()
	}

	if err := h.operationRepo.Update(ctx, operation); err != nil {
		return nil, fmt.Errorf("failed to complete operation: %w", err)
	}

	evt := events.NewWarehouseOperationCompletedEvent(operation, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    operation,
		Events:  []interface{}{evt},
	}, nil
}

func (h *WarehouseCommandHandler) HandleCancelWarehouseOperation(ctx context.Context, cmd *CommandEnvelope) (*CommandResult, error) {
	var input CancelWarehouseOperation
	if err := parseCommandData(cmd, &input); err != nil {
		return nil, fmt.Errorf("failed to parse command data: %w", err)
	}

	operation, err := h.operationRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("operation not found: %w", err)
	}

	if operation.Status == "completed" || operation.Status == "cancelled" {
		return nil, fmt.Errorf("cannot cancel completed or already cancelled operation")
	}

	operation.Cancel(input.Reason)

	if err := h.operationRepo.Update(ctx, operation); err != nil {
		return nil, fmt.Errorf("failed to cancel operation: %w", err)
	}

	evt := events.NewWarehouseOperationCancelledEvent(operation, cmd.UserID)
	if err := h.publisher.PublishEvent(ctx, &evt.EventEnvelope); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &CommandResult{
		Success: true,
		Data:    operation,
		Events:  []interface{}{evt},
	}, nil
}

func parseCommandData(cmd *CommandEnvelope, v interface{}) error {
	data, err := json.Marshal(cmd.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
