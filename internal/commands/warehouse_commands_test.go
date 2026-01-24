package commands

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockPublisher struct {
	events []*events.EventEnvelope
}

func (m *MockPublisher) PublishEvent(ctx context.Context, event *events.EventEnvelope) error {
	m.events = append(m.events, event)
	return nil
}

type MockWarehouseRepository struct {
	warehouses map[uuid.UUID]*domain.Warehouse
}

func NewMockWarehouseRepository() *MockWarehouseRepository {
	return &MockWarehouseRepository{
		warehouses: make(map[uuid.UUID]*domain.Warehouse),
	}
}

func (r *MockWarehouseRepository) Create(ctx context.Context, warehouse *domain.Warehouse) error {
	r.warehouses[warehouse.ID] = warehouse
	return nil
}

func (r *MockWarehouseRepository) Update(ctx context.Context, warehouse *domain.Warehouse) error {
	r.warehouses[warehouse.ID] = warehouse
	return nil
}

func (r *MockWarehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.warehouses, id)
	return nil
}

func (r *MockWarehouseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	if w, ok := r.warehouses[id]; ok {
		return w, nil
	}
	return nil, mongo.ErrNoDocuments
}

func (r *MockWarehouseRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.Warehouse, error) {
	for _, w := range r.warehouses {
		if w.TenantID == tenantID && w.Code == code {
			return w, nil
		}
	}
	return nil, mongo.ErrNoDocuments
}

func (r *MockWarehouseRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Warehouse, error) {
	var result []*domain.Warehouse
	for _, w := range r.warehouses {
		if w.TenantID == tenantID {
			result = append(result, w)
		}
	}
	return result, nil
}

func (r *MockWarehouseRepository) FindActive(ctx context.Context, tenantID uuid.UUID) ([]*domain.Warehouse, error) {
	var result []*domain.Warehouse
	for _, w := range r.warehouses {
		if w.TenantID == tenantID && w.IsActive {
			result = append(result, w)
		}
	}
	return result, nil
}

type MockLocationRepository struct {
	locations map[uuid.UUID]*domain.WarehouseLocation
}

func NewMockLocationRepository() *MockLocationRepository {
	return &MockLocationRepository{
		locations: make(map[uuid.UUID]*domain.WarehouseLocation),
	}
}

func (r *MockLocationRepository) Create(ctx context.Context, location *domain.WarehouseLocation) error {
	r.locations[location.ID] = location
	return nil
}

func (r *MockLocationRepository) Update(ctx context.Context, location *domain.WarehouseLocation) error {
	r.locations[location.ID] = location
	return nil
}

func (r *MockLocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.locations, id)
	return nil
}

func (r *MockLocationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.WarehouseLocation, error) {
	if l, ok := r.locations[id]; ok {
		return l, nil
	}
	return nil, mongo.ErrNoDocuments
}

func (r *MockLocationRepository) FindByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*domain.WarehouseLocation, error) {
	var result []*domain.WarehouseLocation
	for _, l := range r.locations {
		if l.WarehouseID == warehouseID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (r *MockLocationRepository) FindByPath(ctx context.Context, warehouseID uuid.UUID, zone, aisle, rack, bin string) (*domain.WarehouseLocation, error) {
	for _, l := range r.locations {
		if l.WarehouseID == warehouseID && l.Zone == zone && l.Aisle == aisle && l.Rack == rack && l.Bin == bin {
			return l, nil
		}
	}
	return nil, mongo.ErrNoDocuments
}

func (r *MockLocationRepository) FindByBarcode(ctx context.Context, barcode string) (*domain.WarehouseLocation, error) {
	return nil, mongo.ErrNoDocuments
}

func (r *MockLocationRepository) FindAvailable(ctx context.Context, warehouseID uuid.UUID, quantity int) ([]*domain.WarehouseLocation, error) {
	return nil, mongo.ErrNoDocuments
}

type MockOperationRepository struct {
	operations map[uuid.UUID]*domain.WarehouseOperation
}

func NewMockOperationRepository() *MockOperationRepository {
	return &MockOperationRepository{
		operations: make(map[uuid.UUID]*domain.WarehouseOperation),
	}
}

func (r *MockOperationRepository) Create(ctx context.Context, operation *domain.WarehouseOperation) error {
	r.operations[operation.ID] = operation
	return nil
}

func (r *MockOperationRepository) Update(ctx context.Context, operation *domain.WarehouseOperation) error {
	r.operations[operation.ID] = operation
	return nil
}

func (r *MockOperationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.operations, id)
	return nil
}

func (r *MockOperationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.WarehouseOperation, error) {
	if o, ok := r.operations[id]; ok {
		return o, nil
	}
	return nil, mongo.ErrNoDocuments
}

func (r *MockOperationRepository) FindByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*domain.WarehouseOperation, error) {
	return nil, nil
}

func (r *MockOperationRepository) FindByStatus(ctx context.Context, warehouseID uuid.UUID, status string) ([]*domain.WarehouseOperation, error) {
	return nil, nil
}

func (r *MockOperationRepository) FindPending(ctx context.Context, warehouseID uuid.UUID) ([]*domain.WarehouseOperation, error) {
	return nil, nil
}

func (r *MockOperationRepository) FindByReference(ctx context.Context, referenceType string, referenceID uuid.UUID) ([]*domain.WarehouseOperation, error) {
	return nil, nil
}

func TestWarehouseCommandHandler_CreateWarehouse(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	cmd := NewCommand("createWarehouse", tenantID.String(), "", userID.String(), map[string]interface{}{
		"name":     "Main Warehouse",
		"code":     "WH-001",
		"type":     "main",
		"capacity": 10000,
	})

	result, err := handler.HandleCreateWarehouse(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	warehouse := result.Data.(*domain.Warehouse)
	assert.Equal(t, "Main Warehouse", warehouse.Name)
	assert.Equal(t, "WH-001", warehouse.Code)
	assert.Equal(t, domain.WarehouseTypeMain, warehouse.Type)
	assert.True(t, warehouse.IsActive)

	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "warehouse.created", publisher.events[0].Type)
}

func TestWarehouseCommandHandler_CreateWarehouseInvalidType(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	cmd := NewCommand("createWarehouse", tenantID.String(), "", userID.String(), map[string]interface{}{
		"name": "Main Warehouse",
		"code": "WH-001",
		"type": "invalid_type",
	})

	result, err := handler.HandleCreateWarehouse(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Invalid warehouse type")
}

func TestWarehouseCommandHandler_ActivateWarehouse(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	warehouse := domain.NewWarehouse(tenantID, "Test WH", "WH-001", domain.WarehouseTypeMain)
	warehouse.Deactivate()
	warehouseRepo.Create(context.Background(), warehouse)

	cmd := NewCommand("activateWarehouse", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id": warehouse.ID.String(),
	})

	result, err := handler.HandleActivateWarehouse(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	updatedWarehouse, _ := warehouseRepo.FindByID(context.Background(), warehouse.ID)
	assert.True(t, updatedWarehouse.IsActive)

	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "warehouse.activated", publisher.events[0].Type)
}

func TestWarehouseCommandHandler_DeactivateWarehouseWithLocations(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	warehouse := domain.NewWarehouse(tenantID, "Test WH", "WH-001", domain.WarehouseTypeMain)
	warehouseRepo.Create(context.Background(), warehouse)

	location := &domain.WarehouseLocation{
		ID:          uuid.New(),
		TenantID:    tenantID,
		WarehouseID: warehouse.ID,
		Zone:        "A",
		Aisle:       "01",
		Rack:        "01",
		Bin:         "01",
		IsActive:    true,
	}
	locationRepo.Create(context.Background(), location)

	cmd := NewCommand("deactivateWarehouse", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id": warehouse.ID.String(),
	})

	result, err := handler.HandleDeactivateWarehouse(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Cannot deactivate warehouse with active locations")
}

func TestWarehouseCommandHandler_CreateLocation(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	warehouse := domain.NewWarehouse(tenantID, "Test WH", "WH-001", domain.WarehouseTypeMain)
	warehouse.ID = warehouseID
	warehouseRepo.Create(context.Background(), warehouse)

	cmd := NewCommand("createLocation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"warehouseId": warehouseID.String(),
		"name":        "Bin A1",
		"code":        "A-01-01-01",
		"type":        "bin",
		"zone":        "A",
		"aisle":       "01",
		"rack":        "01",
		"bin":         "01",
		"capacity":    100,
	})

	result, err := handler.HandleCreateLocation(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	location := result.Data.(*domain.WarehouseLocation)
	assert.Equal(t, "Bin A1", location.Name)
	assert.Equal(t, "A-01-01-01", location.Code)
	assert.Equal(t, warehouseID, location.WarehouseID)
	assert.True(t, location.IsActive)

	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "location.created", publisher.events[0].Type)
}

func TestWarehouseCommandHandler_CreateLocationMissingPath(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	warehouse := domain.NewWarehouse(tenantID, "Test WH", "WH-001", domain.WarehouseTypeMain)
	warehouse.ID = warehouseID
	warehouseRepo.Create(context.Background(), warehouse)

	cmd := NewCommand("createLocation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"warehouseId": warehouseID.String(),
		"name":        "Bin A1",
		"code":        "A-01-01-01",
		"type":        "bin",
		"zone":        "A",
		"aisle":       "",
		"rack":        "01",
		"bin":         "01",
		"capacity":    100,
	})

	result, err := handler.HandleCreateLocation(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Zone, aisle, rack, and bin are required")
}

func TestWarehouseCommandHandler_CreateWarehouseOperation(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()
	referenceID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	warehouse := domain.NewWarehouse(tenantID, "Test WH", "WH-001", domain.WarehouseTypeMain)
	warehouse.ID = warehouseID
	warehouseRepo.Create(context.Background(), warehouse)

	cmd := NewCommand("createWarehouseOperation", tenantID.String(), warehouseID.String(), userID.String(), map[string]interface{}{
		"type":          "receipt",
		"referenceType": "purchase_order",
		"referenceId":   referenceID.String(),
		"priority":      5,
		"items": []map[string]interface{}{
			{
				"productId":  uuid.New().String(),
				"locationId": uuid.New().String(),
				"quantity":   100,
			},
		},
		"notes": "Receiving shipment",
	})

	result, err := handler.HandleCreateWarehouseOperation(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	operation := result.Data.(*domain.WarehouseOperation)
	assert.Equal(t, domain.OperationTypeReceipt, operation.Type)
	assert.Equal(t, "pending", operation.Status)
	assert.Equal(t, 5, operation.Priority)
	assert.Len(t, operation.Items, 1)

	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "warehouse.operation.created", publisher.events[0].Type)
}

func TestWarehouseCommandHandler_StartWarehouseOperation(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()
	operationID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	operation, _ := domain.NewWarehouseOperation(
		tenantID, warehouseID, userID,
		domain.OperationTypeReceipt, "purchase_order", uuid.New(),
	)
	operation.ID = operationID
	operationRepo.Create(context.Background(), operation)

	cmd := NewCommand("startWarehouseOperation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id": operationID.String(),
	})

	result, err := handler.HandleStartWarehouseOperation(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	updatedOp, _ := operationRepo.FindByID(context.Background(), operationID)
	assert.Equal(t, "in_progress", updatedOp.Status)
	assert.NotNil(t, updatedOp.StartedAt)

	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "warehouse.operation.started", publisher.events[0].Type)
}

func TestWarehouseCommandHandler_CompleteWarehouseOperation(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()
	operationID := uuid.New()
	itemID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	operation, _ := domain.NewWarehouseOperation(
		tenantID, warehouseID, userID,
		domain.OperationTypePick, "order", uuid.New(),
	)
	operation.ID = operationID
	operation.Start()
	operation.AddItem(domain.OperationItem{
		ID:         itemID,
		ProductID:  uuid.New(),
		LocationID: uuid.New(),
		Quantity:   10,
		Status:     "pending",
	})
	operationRepo.Create(context.Background(), operation)

	cmd := NewCommand("completeWarehouseOperation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id": operationID.String(),
		"completedItems": []map[string]interface{}{
			{
				"itemId":   itemID.String(),
				"quantity": 10,
			},
		},
	})

	result, err := handler.HandleCompleteWarehouseOperation(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	updatedOp, _ := operationRepo.FindByID(context.Background(), operationID)
	assert.Equal(t, "completed", updatedOp.Status)
	assert.NotNil(t, updatedOp.CompletedAt)
	assert.True(t, updatedOp.IsComplete())

	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "warehouse.operation.completed", publisher.events[0].Type)
}

func TestWarehouseCommandHandler_CancelWarehouseOperation(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()
	operationID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	operation, _ := domain.NewWarehouseOperation(
		tenantID, warehouseID, userID,
		domain.OperationTypeReceipt, "purchase_order", uuid.New(),
	)
	operation.ID = operationID
	operationRepo.Create(context.Background(), operation)

	cmd := NewCommand("cancelWarehouseOperation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id":     operationID.String(),
		"reason": "Duplicate order",
	})

	result, err := handler.HandleCancelWarehouseOperation(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	updatedOp, _ := operationRepo.FindByID(context.Background(), operationID)
	assert.Equal(t, "cancelled", updatedOp.Status)
	assert.Equal(t, "Duplicate order", updatedOp.Notes)

	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "warehouse.operation.cancelled", publisher.events[0].Type)
}

func TestWarehouseCommandHandler_CancelAlreadyCompletedOperation(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()
	operationID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	operation, _ := domain.NewWarehouseOperation(
		tenantID, warehouseID, userID,
		domain.OperationTypeReceipt, "purchase_order", uuid.New(),
	)
	operation.ID = operationID
	operation.Complete()
	operationRepo.Create(context.Background(), operation)

	cmd := NewCommand("cancelWarehouseOperation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id":     operationID.String(),
		"reason": "Too late to cancel",
	})

	result, err := handler.HandleCancelWarehouseOperation(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot cancel completed")
}

func TestWarehouseCommandHandler_UpdateWarehouse(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	warehouse := domain.NewWarehouse(tenantID, "Test WH", "WH-001", domain.WarehouseTypeMain)
	warehouse.ID = warehouseID
	warehouseRepo.Create(context.Background(), warehouse)

	newName := "Updated Warehouse"
	newCapacity := 20000
	cmd := NewCommand("updateWarehouse", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id":       warehouseID.String(),
		"name":     newName,
		"capacity": newCapacity,
	})

	result, err := handler.HandleUpdateWarehouse(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	updated, _ := warehouseRepo.FindByID(context.Background(), warehouseID)
	assert.Equal(t, newName, updated.Name)
	assert.Equal(t, newCapacity, updated.Capacity)

	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "warehouse.updated", publisher.events[0].Type)
}

func TestWarehouseCommandHandler_UpdateWarehouseCapacityLessThanUsed(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	warehouse := domain.NewWarehouse(tenantID, "Test WH", "WH-001", domain.WarehouseTypeMain)
	warehouse.ID = warehouseID
	warehouse.CurrentUtilization = 10000
	warehouse.Capacity = 20000
	warehouseRepo.Create(context.Background(), warehouse)

	cmd := NewCommand("updateWarehouse", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id":       warehouseID.String(),
		"capacity": 5000,
	})

	result, err := handler.HandleUpdateWarehouse(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Capacity cannot be less than used capacity")
}

func TestWarehouseCommandHandler_DeactivateWarehouse(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	warehouse := domain.NewWarehouse(tenantID, "Test WH", "WH-001", domain.WarehouseTypeMain)
	warehouse.ID = warehouseID
	warehouseRepo.Create(context.Background(), warehouse)

	cmd := NewCommand("deactivateWarehouse", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id": warehouseID.String(),
	})

	result, err := handler.HandleDeactivateWarehouse(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	updated, _ := warehouseRepo.FindByID(context.Background(), warehouseID)
	assert.False(t, updated.IsActive)

	assert.Len(t, publisher.events, 1)
	assert.Equal(t, "warehouse.deactivated", publisher.events[0].Type)
}

func TestWarehouseCommandHandler_ActivateWarehouseNotFound(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	cmd := NewCommand("activateWarehouse", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id": warehouseID.String(),
	})

	result, err := handler.HandleActivateWarehouse(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "warehouse not found")
}

func TestWarehouseCommandHandler_DeactivateWarehouseNotFound(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	cmd := NewCommand("deactivateWarehouse", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id": warehouseID.String(),
	})

	result, err := handler.HandleDeactivateWarehouse(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "warehouse not found")
}

func TestWarehouseCommandHandler_CreateWarehouseOperationInvalidType(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()
	referenceID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	warehouse := domain.NewWarehouse(tenantID, "Test WH", "WH-001", domain.WarehouseTypeMain)
	warehouse.ID = warehouseID
	warehouseRepo.Create(context.Background(), warehouse)

	cmd := NewCommand("createWarehouseOperation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"warehouseId":   warehouseID.String(),
		"type":          "invalid_type",
		"referenceType": "purchase_order",
		"referenceId":   referenceID.String(),
		"priority":      5,
		"items":         []map[string]interface{}{},
		"notes":         "Invalid operation",
	})

	result, err := handler.HandleCreateWarehouseOperation(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Invalid operation type")
}

func TestWarehouseCommandHandler_StartWarehouseOperationNotFound(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	operationID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	cmd := NewCommand("startWarehouseOperation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id": operationID.String(),
	})

	result, err := handler.HandleStartWarehouseOperation(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "operation not found")
}

func TestWarehouseCommandHandler_StartAlreadyStartedOperation(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	warehouseID := uuid.New()
	operationID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	operation, _ := domain.NewWarehouseOperation(
		tenantID, warehouseID, userID,
		domain.OperationTypeReceipt, "purchase_order", uuid.New(),
	)
	operation.ID = operationID
	operation.Start()
	operationRepo.Create(context.Background(), operation)

	cmd := NewCommand("startWarehouseOperation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id": operationID.String(),
	})

	result, err := handler.HandleStartWarehouseOperation(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already started")
}

func TestWarehouseCommandHandler_CompleteWarehouseOperationNotFound(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	operationID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	cmd := NewCommand("completeWarehouseOperation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id": operationID.String(),
	})

	result, err := handler.HandleCompleteWarehouseOperation(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "operation not found")
}

func TestWarehouseCommandHandler_CancelWarehouseOperationNotFound(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	operationID := uuid.New()

	warehouseRepo := NewMockWarehouseRepository()
	locationRepo := NewMockLocationRepository()
	operationRepo := NewMockOperationRepository()
	publisher := &MockPublisher{}

	handler := NewWarehouseCommandHandler(warehouseRepo, locationRepo, operationRepo, publisher)

	cmd := NewCommand("cancelWarehouseOperation", tenantID.String(), "", userID.String(), map[string]interface{}{
		"id":     operationID.String(),
		"reason": "Test cancel",
	})

	result, err := handler.HandleCancelWarehouseOperation(context.Background(), cmd)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "operation not found")
}
