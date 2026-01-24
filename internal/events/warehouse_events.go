package events

import (
	"time"

	"github.com/ims-erp/system/internal/domain"
)

type WarehouseCreatedEvent struct {
	EventEnvelope
}

func NewWarehouseCreatedEvent(warehouse *domain.Warehouse, userID string) *WarehouseCreatedEvent {
	event := NewEvent(
		warehouse.ID.String(),
		"Warehouse",
		"warehouse.created",
		warehouse.TenantID.String(),
		userID,
		map[string]interface{}{
			"name":         warehouse.Name,
			"code":         warehouse.Code,
			"type":         string(warehouse.Type),
			"capacity":     warehouse.Capacity,
			"managerId":    warehouse.ManagerID,
			"contactEmail": warehouse.ContactEmail,
			"contactPhone": warehouse.ContactPhone,
			"isPrimary":    warehouse.IsPrimary,
		},
	)
	return &WarehouseCreatedEvent{*event}
}

type WarehouseUpdatedEvent struct {
	EventEnvelope
}

func NewWarehouseUpdatedEvent(warehouse *domain.Warehouse, userID string) *WarehouseUpdatedEvent {
	event := NewEvent(
		warehouse.ID.String(),
		"Warehouse",
		"warehouse.updated",
		warehouse.TenantID.String(),
		userID,
		map[string]interface{}{
			"name":           warehouse.Name,
			"capacity":       warehouse.Capacity,
			"managerId":      warehouse.ManagerID,
			"contactEmail":   warehouse.ContactEmail,
			"contactPhone":   warehouse.ContactPhone,
			"operatingHours": warehouse.OperatingHours,
		},
	)
	return &WarehouseUpdatedEvent{*event}
}

type WarehouseActivatedEvent struct {
	EventEnvelope
}

func NewWarehouseActivatedEvent(warehouse *domain.Warehouse, userID string) *WarehouseActivatedEvent {
	event := NewEvent(
		warehouse.ID.String(),
		"Warehouse",
		"warehouse.activated",
		warehouse.TenantID.String(),
		userID,
		map[string]interface{}{},
	)
	return &WarehouseActivatedEvent{*event}
}

type WarehouseDeactivatedEvent struct {
	EventEnvelope
}

func NewWarehouseDeactivatedEvent(warehouse *domain.Warehouse, userID string) *WarehouseDeactivatedEvent {
	event := NewEvent(
		warehouse.ID.String(),
		"Warehouse",
		"warehouse.deactivated",
		warehouse.TenantID.String(),
		userID,
		map[string]interface{}{},
	)
	return &WarehouseDeactivatedEvent{*event}
}

type LocationCreatedEvent struct {
	EventEnvelope
}

func NewLocationCreatedEvent(location *domain.WarehouseLocation, userID string) *LocationCreatedEvent {
	event := NewEvent(
		location.ID.String(),
		"Location",
		"location.created",
		location.TenantID.String(),
		userID,
		map[string]interface{}{
			"warehouseId": location.WarehouseID,
			"name":        location.Name,
			"code":        location.Code,
			"type":        location.Type,
			"zone":        location.Zone,
			"aisle":       location.Aisle,
			"rack":        location.Rack,
			"bin":         location.Bin,
			"capacity":    location.Capacity,
		},
	)
	return &LocationCreatedEvent{*event}
}

type LocationUpdatedEvent struct {
	EventEnvelope
}

func NewLocationUpdatedEvent(location *domain.WarehouseLocation, userID string) *LocationUpdatedEvent {
	event := NewEvent(
		location.ID.String(),
		"Location",
		"location.updated",
		location.TenantID.String(),
		userID,
		map[string]interface{}{
			"name":     location.Name,
			"capacity": location.Capacity,
			"isActive": location.IsActive,
		},
	)
	return &LocationUpdatedEvent{*event}
}

type WarehouseOperationCreatedEvent struct {
	EventEnvelope
}

func NewWarehouseOperationCreatedEvent(operation *domain.WarehouseOperation, userID string) *WarehouseOperationCreatedEvent {
	event := NewEvent(
		operation.ID.String(),
		"WarehouseOperation",
		"warehouse.operation.created",
		operation.TenantID.String(),
		userID,
		map[string]interface{}{
			"warehouseId":   operation.WarehouseID,
			"type":          string(operation.Type),
			"referenceType": operation.ReferenceType,
			"referenceId":   operation.ReferenceID,
			"priority":      operation.Priority,
			"itemCount":     len(operation.Items),
		},
	)
	return &WarehouseOperationCreatedEvent{*event}
}

type WarehouseOperationStartedEvent struct {
	EventEnvelope
}

func NewWarehouseOperationStartedEvent(operation *domain.WarehouseOperation, userID string) *WarehouseOperationStartedEvent {
	event := NewEvent(
		operation.ID.String(),
		"WarehouseOperation",
		"warehouse.operation.started",
		operation.TenantID.String(),
		userID,
		map[string]interface{}{
			"warehouseId": operation.WarehouseID,
			"type":        string(operation.Type),
		},
	)
	return &WarehouseOperationStartedEvent{*event}
}

type WarehouseOperationCompletedEvent struct {
	EventEnvelope
}

func NewWarehouseOperationCompletedEvent(operation *domain.WarehouseOperation, userID string) *WarehouseOperationCompletedEvent {
	var duration float64
	if operation.StartedAt != nil {
		duration = time.Since(*operation.StartedAt).Seconds()
	}

	completedItems := 0
	for _, item := range operation.Items {
		if item.Status == "completed" {
			completedItems++
		}
	}

	event := NewEvent(
		operation.ID.String(),
		"WarehouseOperation",
		"warehouse.operation.completed",
		operation.TenantID.String(),
		userID,
		map[string]interface{}{
			"warehouseId":     operation.WarehouseID,
			"type":            string(operation.Type),
			"itemsCompleted":  completedItems,
			"totalItems":      len(operation.Items),
			"durationSeconds": duration,
		},
	)
	return &WarehouseOperationCompletedEvent{*event}
}

type WarehouseOperationCancelledEvent struct {
	EventEnvelope
}

func NewWarehouseOperationCancelledEvent(operation *domain.WarehouseOperation, userID string) *WarehouseOperationCancelledEvent {
	event := NewEvent(
		operation.ID.String(),
		"WarehouseOperation",
		"warehouse.operation.cancelled",
		operation.TenantID.String(),
		userID,
		map[string]interface{}{
			"warehouseId": operation.WarehouseID,
			"type":        string(operation.Type),
			"reason":      operation.Notes,
		},
	)
	return &WarehouseOperationCancelledEvent{*event}
}

type StockReservedEvent struct {
	EventEnvelope
}

func NewStockReservedEvent(reservation *domain.StockReservation, userID string) *StockReservedEvent {
	event := NewEvent(
		reservation.ID.String(),
		"StockReservation",
		"inventory.stock.reserved",
		reservation.TenantID.String(),
		userID,
		map[string]interface{}{
			"productId":     reservation.ProductID,
			"warehouseId":   reservation.WarehouseID,
			"quantity":      reservation.Quantity,
			"referenceType": reservation.ReferenceType,
			"referenceId":   reservation.ReferenceID,
			"expiresAt":     reservation.ExpiresAt,
		},
	)
	return &StockReservedEvent{*event}
}

type ReservationReleasedEvent struct {
	EventEnvelope
}

func NewReservationReleasedEvent(reservation *domain.StockReservation, userID string, reason string) *ReservationReleasedEvent {
	event := NewEvent(
		reservation.ID.String(),
		"StockReservation",
		"inventory.stock.reservation_released",
		reservation.TenantID.String(),
		userID,
		map[string]interface{}{
			"productId":   reservation.ProductID,
			"warehouseId": reservation.WarehouseID,
			"quantity":    reservation.Quantity,
			"reason":      reason,
		},
	)
	return &ReservationReleasedEvent{*event}
}

type ReservationCommittedEvent struct {
	EventEnvelope
}

func NewReservationCommittedEvent(reservation *domain.StockReservation, userID string) *ReservationCommittedEvent {
	event := NewEvent(
		reservation.ID.String(),
		"StockReservation",
		"inventory.stock.reservation_committed",
		reservation.TenantID.String(),
		userID,
		map[string]interface{}{
			"productId":     reservation.ProductID,
			"warehouseId":   reservation.WarehouseID,
			"quantity":      reservation.Quantity,
			"referenceType": reservation.ReferenceType,
			"referenceId":   reservation.ReferenceID,
		},
	)
	return &ReservationCommittedEvent{*event}
}

type InventoryAdjustedEvent struct {
	EventEnvelope
}

func NewInventoryAdjustedEvent(adjustment *domain.InventoryAdjustment, previousQty, newQty int, userID string) *InventoryAdjustedEvent {
	event := NewEvent(
		adjustment.ID.String(),
		"InventoryAdjustment",
		"inventory.adjusted",
		adjustment.TenantID.String(),
		userID,
		map[string]interface{}{
			"productId":        adjustment.ProductID,
			"warehouseId":      adjustment.WarehouseID,
			"locationId":       adjustment.LocationID,
			"adjustmentType":   adjustment.AdjustmentType,
			"previousQuantity": previousQty,
			"newQuantity":      newQty,
			"reason":           adjustment.Reason,
		},
	)
	return &InventoryAdjustedEvent{*event}
}

type InventoryReceivedEvent struct {
	EventEnvelope
}

func NewInventoryReceivedEvent(transaction *domain.InventoryTransaction, unitCost string, userID string) *InventoryReceivedEvent {
	event := NewEvent(
		transaction.ID.String(),
		"InventoryTransaction",
		"inventory.received",
		transaction.TenantID.String(),
		userID,
		map[string]interface{}{
			"productId":     transaction.ProductID,
			"warehouseId":   transaction.WarehouseID,
			"locationId":    transaction.ToLocationID,
			"quantity":      transaction.Quantity,
			"unitCost":      unitCost,
			"referenceType": transaction.ReferenceType,
			"referenceId":   transaction.ReferenceID,
		},
	)
	return &InventoryReceivedEvent{*event}
}

type InventoryShippedEvent struct {
	EventEnvelope
}

func NewInventoryShippedEvent(transaction *domain.InventoryTransaction, userID string) *InventoryShippedEvent {
	event := NewEvent(
		transaction.ID.String(),
		"InventoryTransaction",
		"inventory.shipped",
		transaction.TenantID.String(),
		userID,
		map[string]interface{}{
			"productId":     transaction.ProductID,
			"warehouseId":   transaction.WarehouseID,
			"locationId":    transaction.FromLocationID,
			"quantity":      transaction.Quantity,
			"referenceType": transaction.ReferenceType,
			"referenceId":   transaction.ReferenceID,
		},
	)
	return &InventoryShippedEvent{*event}
}

type InventoryTransferredEvent struct {
	EventEnvelope
}

func NewInventoryTransferredEvent(transaction *domain.InventoryTransaction, userID string) *InventoryTransferredEvent {
	event := NewEvent(
		transaction.ID.String(),
		"InventoryTransaction",
		"inventory.transferred",
		transaction.TenantID.String(),
		userID,
		map[string]interface{}{
			"productId":       transaction.ProductID,
			"fromWarehouseId": transaction.WarehouseID,
			"toWarehouseId":   transaction.ToLocationID,
			"fromLocationId":  transaction.FromLocationID,
			"toLocationId":    transaction.ToLocationID,
			"quantity":        transaction.Quantity,
			"referenceType":   transaction.ReferenceType,
			"referenceId":     transaction.ReferenceID,
		},
	)
	return &InventoryTransferredEvent{*event}
}
