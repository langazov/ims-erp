package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewWarehouse(t *testing.T) {
	tenantID := uuid.New()
	name := "Main Warehouse"
	code := "WH-001"
	wType := WarehouseTypeMain

	warehouse := NewWarehouse(tenantID, name, code, wType)

	assert.NotEmpty(t, warehouse.ID)
	assert.Equal(t, tenantID, warehouse.TenantID)
	assert.Equal(t, code, warehouse.Code)
	assert.Equal(t, name, warehouse.Name)
	assert.Equal(t, wType, warehouse.Type)
	assert.True(t, warehouse.IsActive)
	assert.False(t, warehouse.IsPrimary)
	assert.Equal(t, 0, warehouse.Capacity)
}

func TestWarehouseActivateDeactivate(t *testing.T) {
	warehouse := NewWarehouse(uuid.New(), "Main", "WH-001", WarehouseTypeMain)
	assert.True(t, warehouse.IsActive)

	warehouse.Deactivate()
	assert.False(t, warehouse.IsActive)

	warehouse.Activate()
	assert.True(t, warehouse.IsActive)
}

func TestWarehouseSetAddress(t *testing.T) {
	warehouse := NewWarehouse(uuid.New(), "Main", "WH-001", WarehouseTypeMain)
	addr := Address{
		Street:     "123 Main St",
		City:       "New York",
		State:      "NY",
		PostalCode: "10001",
		Country:    "USA",
	}

	warehouse.SetAddress(addr)

	assert.Equal(t, addr, warehouse.Address)
}

func TestWarehouseLocationStruct(t *testing.T) {
	location := &WarehouseLocation{
		ID:           uuid.New(),
		TenantID:     uuid.New(),
		WarehouseID:  uuid.New(),
		Code:         "A-01-01-01",
		Zone:         "A",
		Aisle:        "01",
		Rack:         "01",
		Bin:          "01",
		Capacity:     100,
		CurrentStock: 50,
		IsActive:     true,
	}

	assert.NotEmpty(t, location.ID)
	assert.Equal(t, "A-01-01-01", location.Code)
	assert.Equal(t, 100, location.Capacity)
	assert.Equal(t, 50, location.CurrentStock)
	assert.True(t, location.IsActive)
}

func TestNewWarehouseOperation(t *testing.T) {
	tenantID := uuid.New()
	warehouseID := uuid.New()
	createdBy := uuid.New()
	opType := OperationTypeReceipt
	referenceType := "purchase_order"
	referenceID := uuid.New()

	operation, err := NewWarehouseOperation(tenantID, warehouseID, createdBy, opType, referenceType, referenceID)

	assert.NoError(t, err)
	assert.NotEmpty(t, operation.ID)
	assert.Equal(t, tenantID, operation.TenantID)
	assert.Equal(t, warehouseID, operation.WarehouseID)
	assert.Equal(t, opType, operation.Type)
	assert.Equal(t, "pending", operation.Status)
	assert.Equal(t, 5, operation.Priority)
}

func TestNewWarehouseOperationInvalidType(t *testing.T) {
	operation, err := NewWarehouseOperation(uuid.New(), uuid.New(), uuid.New(), OperationType("invalid"), "type", uuid.New())

	assert.Error(t, err)
	assert.Nil(t, operation)
}

func TestWarehouseOperationWorkflow(t *testing.T) {
	operation, _ := NewWarehouseOperation(uuid.New(), uuid.New(), uuid.New(), OperationTypePick, "order", uuid.New())
	userID := uuid.New()

	operation.AssignTo(userID)
	assert.NotNil(t, operation.AssignedTo)

	operation.Start()
	assert.Equal(t, "in_progress", operation.Status)
	assert.NotNil(t, operation.StartedAt)

	itemID := uuid.New()
	operation.AddItem(OperationItem{
		ID:         itemID,
		ProductID:  uuid.New(),
		LocationID: uuid.New(),
		Quantity:   10,
		Status:     "pending",
	})

	err := operation.CompleteItem(itemID, 5)
	assert.NoError(t, err)
	assert.Equal(t, "partial", operation.Items[0].Status)

	err = operation.CompleteItem(itemID, 5)
	assert.NoError(t, err)
	assert.True(t, operation.IsComplete())

	operation.Complete()
	assert.Equal(t, "completed", operation.Status)
	assert.NotNil(t, operation.CompletedAt)
}

func TestWarehouseOperationCancel(t *testing.T) {
	operation, _ := NewWarehouseOperation(uuid.New(), uuid.New(), uuid.New(), OperationTypeReceipt, "po", uuid.New())

	operation.Cancel("Duplicate order")

	assert.Equal(t, "cancelled", operation.Status)
	assert.Equal(t, "Duplicate order", operation.Notes)
}

func TestWarehouseOperationSetPriority(t *testing.T) {
	operation, _ := NewWarehouseOperation(uuid.New(), uuid.New(), uuid.New(), OperationTypeReceipt, "po", uuid.New())

	assert.Equal(t, 5, operation.Priority)

	operation.SetPriority(8)
	assert.Equal(t, 8, operation.Priority)

	operation.SetPriority(0)
	assert.Equal(t, 1, operation.Priority)

	operation.SetPriority(15)
	assert.Equal(t, 10, operation.Priority)
}

func TestWarehouseTypeIsValid(t *testing.T) {
	assert.True(t, WarehouseTypeMain.IsValid())
	assert.True(t, WarehouseTypeSatellite.IsValid())
	assert.True(t, WarehouseTypeReturns.IsValid())
	assert.True(t, WarehouseTypeDistribution.IsValid())
	assert.True(t, WarehouseTypeFulfillment.IsValid())
	assert.True(t, WarehouseTypeCrossDock.IsValid())
	assert.False(t, WarehouseType("invalid").IsValid())
}

func TestWarehouseStatusIsValid(t *testing.T) {
	assert.True(t, WarehouseStatusActive.IsValid())
	assert.True(t, WarehouseStatusInactive.IsValid())
	assert.True(t, WarehouseStatusClosed.IsValid())
	assert.True(t, WarehouseStatusPending.IsValid())
	assert.False(t, WarehouseStatus("invalid").IsValid())
}

func TestLocationStatusIsValid(t *testing.T) {
	assert.True(t, LocationStatusActive.IsValid())
	assert.True(t, LocationStatusInactive.IsValid())
	assert.True(t, LocationStatusReserved.IsValid())
	assert.True(t, LocationStatusMaintenance.IsValid())
	assert.False(t, LocationStatus("invalid").IsValid())
}

func TestOperationTypeIsValid(t *testing.T) {
	assert.True(t, OperationTypeReceipt.IsValid())
	assert.True(t, OperationTypePick.IsValid())
	assert.True(t, OperationTypePack.IsValid())
	assert.True(t, OperationTypeShip.IsValid())
	assert.True(t, OperationTypeTransfer.IsValid())
	assert.True(t, OperationTypeAdjustment.IsValid())
	assert.True(t, OperationTypeCycleCount.IsValid())
	assert.False(t, OperationType("invalid").IsValid())
}

func TestWarehouseError(t *testing.T) {
	err := &WarehouseError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
	}

	assert.Equal(t, "TEST_ERROR", err.Code)
	assert.Equal(t, "Test error message", err.Error())
}

func TestMovementTypeConstants(t *testing.T) {
	assert.Equal(t, MovementType("receipt"), MovementTypeReceipt)
	assert.Equal(t, MovementType("shipment"), MovementTypeShipment)
	assert.Equal(t, MovementType("transfer_in"), MovementTypeTransferIn)
	assert.Equal(t, MovementType("transfer_out"), MovementTypeTransferOut)
	assert.Equal(t, MovementType("adjustment"), MovementTypeAdjustment)
}

func TestInventoryStatusConstants(t *testing.T) {
	assert.Equal(t, InventoryStatus("available"), InventoryStatusAvailable)
	assert.Equal(t, InventoryStatus("reserved"), InventoryStatusReserved)
	assert.Equal(t, InventoryStatus("allocated"), InventoryStatusAllocated)
	assert.Equal(t, InventoryStatus("in_transit"), InventoryStatusInTransit)
}

func TestStockReservationLifecycle(t *testing.T) {
	reservation := NewStockReservation(
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		"order", 10,
	)

	assert.Equal(t, "active", reservation.Status)
	assert.Equal(t, 10, reservation.Quantity)

	reservation.Release()
	assert.Equal(t, "released", reservation.Status)
}

func TestInventoryTransaction(t *testing.T) {
	tx := NewInventoryTransaction(
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		MovementTypeReceipt, 100,
	)

	assert.NotEmpty(t, tx.ID)
	assert.Equal(t, MovementTypeReceipt, tx.MovementType)
	assert.Equal(t, 100, tx.Quantity)

	tx.SetReference("purchase_order", uuid.New())
	assert.Equal(t, "purchase_order", tx.ReferenceType)
}

func TestNewInventoryItem(t *testing.T) {
	item := NewInventoryItem(
		uuid.New(), uuid.New(), uuid.New(),
		"SKU-001", 50,
		decimal.NewFromFloat(10.00),
	)

	assert.NotEmpty(t, item.ID)
	assert.Equal(t, "SKU-001", item.SKU)
	assert.Equal(t, 50, item.Quantity)
	assert.Equal(t, 50, item.AvailableQty)
	assert.Equal(t, InventoryStatusAvailable, item.Status)
}

func TestInventoryItemReserve(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))

	err := item.Reserve(20)
	assert.NoError(t, err)
	assert.Equal(t, 20, item.ReservedQty)
	assert.Equal(t, 30, item.AvailableQty)
}

func TestInventoryItemReserveInsufficient(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))

	err := item.Reserve(60)
	assert.Error(t, err)
}

func TestInventoryItemReleaseReservation(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))
	item.ReservedQty = 20
	item.AvailableQty = 30

	item.ReleaseReservation(10)

	assert.Equal(t, 10, item.ReservedQty)
	assert.Equal(t, 40, item.AvailableQty)
}

func TestInventoryItemAllocate(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))

	err := item.Allocate(15)
	assert.NoError(t, err)
	assert.Equal(t, 15, item.AllocatedQty)
	assert.Equal(t, 35, item.AvailableQty)
}

func TestInventoryItemAllocateInsufficient(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))
	item.AvailableQty = 5

	err := item.Allocate(10)
	assert.Error(t, err)
}

func TestInventoryItemDeallocate(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))
	item.AllocatedQty = 20
	item.AvailableQty = 30

	item.Deallocate(10)

	assert.Equal(t, 10, item.AllocatedQty)
	assert.Equal(t, 40, item.AvailableQty)
}

func TestInventoryItemReceive(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))

	item.Receive(25, decimal.NewFromFloat(12.00))

	assert.Equal(t, 75, item.Quantity)
	assert.Equal(t, 75, item.AvailableQty)
	assert.Equal(t, decimal.NewFromFloat(12.00), item.UnitCost)
}

func TestInventoryItemShip(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))

	err := item.Ship(20)
	assert.NoError(t, err)
	assert.Equal(t, 30, item.Quantity)
	assert.Equal(t, 30, item.AvailableQty)
}

func TestInventoryItemShipInsufficient(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))

	err := item.Ship(60)
	assert.Error(t, err)
}

func TestInventoryItemAdjust(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))

	item.Adjust(10, "Cycle count adjustment")

	assert.Equal(t, 60, item.Quantity)
	assert.Equal(t, 60, item.AvailableQty)
}

func TestInventoryItemAdjustNegative(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))

	item.Adjust(-15, "Damaged items")

	assert.Equal(t, 35, item.Quantity)
	assert.Equal(t, 35, item.AvailableQty)
}

func TestInventoryItemCount(t *testing.T) {
	item := NewInventoryItem(uuid.New(), uuid.New(), uuid.New(), "SKU-001", 50, decimal.NewFromFloat(10.00))
	item.ReservedQty = 10
	item.AvailableQty = 40

	item.Count(45)

	assert.Equal(t, 45, item.Quantity)
	assert.Equal(t, 35, item.AvailableQty)
	assert.NotNil(t, item.LastCountedAt)
}

func TestStockReservationExpire(t *testing.T) {
	reservation := NewStockReservation(
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		"order", 10,
	)

	reservation.Expire()

	assert.Equal(t, "expired", reservation.Status)
	assert.NotNil(t, reservation.ReleasedAt)
}

func TestStockReservationFulfill(t *testing.T) {
	reservation := NewStockReservation(
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		"order", 10,
	)

	reservation.Fulfill()

	assert.Equal(t, "fulfilled", reservation.Status)
}

func TestInventoryTransactionSetTransfer(t *testing.T) {
	tx := NewInventoryTransaction(
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		MovementTypeTransferOut, 100,
	)

	fromID := uuid.New()
	toID := uuid.New()
	tx.SetTransfer(fromID, toID)

	assert.Equal(t, &fromID, tx.FromLocationID)
	assert.Equal(t, &toID, tx.ToLocationID)
}

func TestInventoryTransactionSetLotInfo(t *testing.T) {
	tx := NewInventoryTransaction(
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		MovementTypeReceipt, 100,
	)

	tx.SetLotInfo("LOT-123", "SN-456")

	assert.Equal(t, "LOT-123", tx.LotNumber)
	assert.Equal(t, "SN-456", tx.SerialNumber)
}

func TestWarehouseErrorIs(t *testing.T) {
	err1 := &WarehouseError{Code: "ERR1", Message: "Error 1"}
	err2 := &WarehouseError{Code: "ERR2", Message: "Error 2"}
	var target *WarehouseError

	assert.True(t, errors.Is(err1, target))
	assert.True(t, errors.Is(err1, err2))
}

func TestInventoryError(t *testing.T) {
	err := &InventoryError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
	}

	assert.Equal(t, "TEST_ERROR", err.Code)
	assert.Equal(t, "Test error message", err.Error())
}

func TestInventoryErrorIs(t *testing.T) {
	err1 := &InventoryError{Code: "ERR1", Message: "Error 1"}
	err2 := &InventoryError{Code: "ERR2", Message: "Error 2"}
	var target *InventoryError

	assert.True(t, errors.Is(err1, target))
	assert.True(t, errors.Is(err1, err2))
}
