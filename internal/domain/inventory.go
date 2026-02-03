package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InventoryStatus string

const (
	InventoryStatusAvailable  InventoryStatus = "available"
	InventoryStatusReserved   InventoryStatus = "reserved"
	InventoryStatusAllocated  InventoryStatus = "allocated"
	InventoryStatusInTransit  InventoryStatus = "in_transit"
	InventoryStatusQuarantine InventoryStatus = "quarantine"
	InventoryStatusDamaged    InventoryStatus = "damaged"
	InventoryStatusExpired    InventoryStatus = "expired"
	InventoryStatusReturned   InventoryStatus = "returned"
)

type MovementType string

const (
	MovementTypeReceipt            MovementType = "receipt"
	MovementTypeShipment           MovementType = "shipment"
	MovementTypeTransferIn         MovementType = "transfer_in"
	MovementTypeTransferOut        MovementType = "transfer_out"
	MovementTypeAdjustment         MovementType = "adjustment"
	MovementTypeReservation        MovementType = "reservation"
	MovementTypeReservationRelease MovementType = "reservation_release"
	MovementTypeAllocation         MovementType = "allocation"
	MovementTypeDeallocation       MovementType = "deallocation"
	MovementTypeReturn             MovementType = "return"
	MovementTypeWriteOff           MovementType = "write_off"
	MovementTypeCycleCount         MovementType = "cycle_count"
	MovementTypeDamaged            MovementType = "damaged"
	MovementTypeExpired            MovementType = "expired"
)

type WarehouseType string

const (
	WarehouseTypeMain         WarehouseType = "main"
	WarehouseTypeSatellite    WarehouseType = "satellite"
	WarehouseTypeReturns      WarehouseType = "returns"
	WarehouseTypeDistribution WarehouseType = "distribution"
	WarehouseTypeFulfillment  WarehouseType = "fulfillment"
	WarehouseTypeCrossDock    WarehouseType = "cross_dock"
)

func (t WarehouseType) IsValid() bool {
	switch t {
	case WarehouseTypeMain, WarehouseTypeSatellite, WarehouseTypeReturns,
		WarehouseTypeDistribution, WarehouseTypeFulfillment, WarehouseTypeCrossDock:
		return true
	}
	return false
}

type WarehouseStatus string

const (
	WarehouseStatusActive   WarehouseStatus = "active"
	WarehouseStatusInactive WarehouseStatus = "inactive"
	WarehouseStatusClosed   WarehouseStatus = "closed"
	WarehouseStatusPending  WarehouseStatus = "pending"
)

func (s WarehouseStatus) IsValid() bool {
	switch s {
	case WarehouseStatusActive, WarehouseStatusInactive, WarehouseStatusClosed, WarehouseStatusPending:
		return true
	}
	return false
}

type LocationStatus string

const (
	LocationStatusActive      LocationStatus = "active"
	LocationStatusInactive    LocationStatus = "inactive"
	LocationStatusReserved    LocationStatus = "reserved"
	LocationStatusMaintenance LocationStatus = "maintenance"
)

func (s LocationStatus) IsValid() bool {
	switch s {
	case LocationStatusActive, LocationStatusInactive, LocationStatusReserved, LocationStatusMaintenance:
		return true
	}
	return false
}

type OperationType string

const (
	OperationTypeReceipt    OperationType = "receipt"
	OperationTypePick       OperationType = "pick"
	OperationTypePack       OperationType = "pack"
	OperationTypeShip       OperationType = "ship"
	OperationTypeTransfer   OperationType = "transfer"
	OperationTypeAdjustment OperationType = "adjustment"
	OperationTypeCycleCount OperationType = "cycle_count"
)

func (t OperationType) IsValid() bool {
	switch t {
	case OperationTypeReceipt, OperationTypePick, OperationTypePack, OperationTypeShip,
		OperationTypeTransfer, OperationTypeAdjustment, OperationTypeCycleCount:
		return true
	}
	return false
}

type InventoryItem struct {
	ID             uuid.UUID       `json:"id" bson:"_id"`
	TenantID       uuid.UUID       `json:"tenantId" bson:"tenantId"`
	ProductID      uuid.UUID       `json:"productId" bson:"productId"`
	VariantID      *uuid.UUID      `json:"variantId" bson:"variantId"`
	SKU            string          `json:"sku" bson:"sku"`
	WarehouseID    uuid.UUID       `json:"warehouseId" bson:"warehouseId"`
	LocationID     uuid.UUID       `json:"locationId" bson:"locationId"`
	BinID          uuid.UUID       `json:"binId" bson:"binId"`
	LotNumber      string          `json:"lotNumber" bson:"lotNumber"`
	SerialNumber   string          `json:"serialNumber" bson:"serialNumber"`
	BatchNumber    string          `json:"batchNumber" bson:"batchNumber"`
	ExpirationDate *time.Time      `json:"expirationDate" bson:"expirationDate"`
	Quantity       int             `json:"quantity" bson:"quantity"`
	ReservedQty    int             `json:"reservedQty" bson:"reservedQty"`
	AvailableQty   int             `json:"availableQty" bson:"availableQty"`
	AllocatedQty   int             `json:"allocatedQty" bson:"allocatedQty"`
	Status         InventoryStatus `json:"status" bson:"status"`
	UnitCost       decimal.Decimal `json:"unitCost" bson:"unitCost"`
	TotalValue     decimal.Decimal `json:"totalValue" bson:"totalValue"`
	LastCountedAt  *time.Time      `json:"lastCountedAt" bson:"lastCountedAt"`
	CreatedAt      time.Time       `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt" bson:"updatedAt"`
	Version        int64           `json:"-" bson:"version"`
}

type InventoryTransaction struct {
	ID             uuid.UUID    `json:"id" bson:"_id"`
	TenantID       uuid.UUID    `json:"tenantId" bson:"tenantId"`
	ProductID      uuid.UUID    `json:"productId" bson:"productId"`
	VariantID      *uuid.UUID   `json:"variantId" bson:"variantId"`
	WarehouseID    uuid.UUID    `json:"warehouseId" bson:"warehouseId"`
	FromLocationID *uuid.UUID   `json:"fromLocationId" bson:"fromLocationId"`
	ToLocationID   *uuid.UUID   `json:"toLocationId" bson:"toLocationId"`
	MovementType   MovementType `json:"movementType" bson:"movementType"`
	Quantity       int          `json:"quantity" bson:"quantity"`
	ReferenceType  string       `json:"referenceType" bson:"referenceType"`
	ReferenceID    uuid.UUID    `json:"referenceId" bson:"referenceId"`
	LotNumber      string       `json:"lotNumber" bson:"lotNumber"`
	SerialNumber   string       `json:"serialNumber" bson:"serialNumber"`
	Reason         string       `json:"reason" bson:"reason"`
	Notes          string       `json:"notes" bson:"notes"`
	PerformedBy    uuid.UUID    `json:"performedBy" bson:"performedBy"`
	CreatedAt      time.Time    `json:"createdAt" bson:"createdAt"`
}

type Warehouse struct {
	ID                 uuid.UUID     `json:"id" bson:"_id"`
	TenantID           uuid.UUID     `json:"tenantId" bson:"tenantId"`
	Name               string        `json:"name" bson:"name"`
	Code               string        `json:"code" bson:"code"`
	Type               WarehouseType `json:"type" bson:"type"`
	Address            Address       `json:"address" bson:"address"`
	IsActive           bool          `json:"isActive" bson:"isActive"`
	IsPrimary          bool          `json:"isPrimary" bson:"isPrimary"`
	Capacity           int           `json:"capacity" bson:"capacity"`
	CurrentUtilization float64       `json:"currentUtilization" bson:"currentUtilization"`
	ManagerID          *uuid.UUID    `json:"managerId" bson:"managerId"`
	ContactEmail       string        `json:"contactEmail" bson:"contactEmail"`
	ContactPhone       string        `json:"contactPhone" bson:"contactPhone"`
	OperatingHours     string        `json:"operatingHours" bson:"operatingHours"`
	CreatedAt          time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time     `json:"updatedAt" bson:"updatedAt"`
}

type WarehouseLocation struct {
	ID           uuid.UUID `json:"id" bson:"_id"`
	TenantID     uuid.UUID `json:"tenantId" bson:"tenantId"`
	WarehouseID  uuid.UUID `json:"warehouseId" bson:"warehouseId"`
	Name         string    `json:"name" bson:"name"`
	Code         string    `json:"code" bson:"code"`
	Type         string    `json:"type" bson:"type"`
	Zone         string    `json:"zone" bson:"zone"`
	Aisle        string    `json:"aisle" bson:"aisle"`
	Rack         string    `json:"rack" bson:"rack"`
	Bin          string    `json:"bin" bson:"bin"`
	Capacity     int       `json:"capacity" bson:"capacity"`
	CurrentStock int       `json:"currentStock" bson:"currentStock"`
	IsActive     bool      `json:"isActive" bson:"isActive"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
}

type StockReservation struct {
	ID            uuid.UUID  `json:"id" bson:"_id"`
	TenantID      uuid.UUID  `json:"tenantId" bson:"tenantId"`
	ProductID     uuid.UUID  `json:"productId" bson:"productId"`
	VariantID     *uuid.UUID `json:"variantId" bson:"variantId"`
	WarehouseID   uuid.UUID  `json:"warehouseId" bson:"warehouseId"`
	ReferenceType string     `json:"referenceType" bson:"referenceType"`
	ReferenceID   uuid.UUID  `json:"referenceId" bson:"referenceId"`
	Quantity      int        `json:"quantity" bson:"quantity"`
	Status        string     `json:"status" bson:"status"`
	ExpiresAt     *time.Time `json:"expiresAt" bson:"expiresAt"`
	CreatedAt     time.Time  `json:"createdAt" bson:"createdAt"`
	ReleasedAt    *time.Time `json:"releasedAt" bson:"releasedAt"`
}

type InventoryAdjustment struct {
	ID             uuid.UUID  `json:"id" bson:"_id"`
	TenantID       uuid.UUID  `json:"tenantId" bson:"tenantId"`
	ProductID      uuid.UUID  `json:"productId" bson:"productId"`
	VariantID      *uuid.UUID `json:"variantId" bson:"variantId"`
	WarehouseID    uuid.UUID  `json:"warehouseId" bson:"warehouseId"`
	LocationID     *uuid.UUID `json:"locationId" bson:"locationId"`
	AdjustmentType string     `json:"adjustmentType" bson:"adjustmentType"`
	Quantity       int        `json:"quantity" bson:"quantity"`
	Reason         string     `json:"reason" bson:"reason"`
	ReferenceType  string     `json:"referenceType" bson:"referenceType"`
	ReferenceID    uuid.UUID  `json:"referenceId" bson:"referenceId"`
	PerformedBy    uuid.UUID  `json:"performedBy" bson:"performedBy"`
	CreatedAt      time.Time  `json:"createdAt" bson:"createdAt"`
}

type InventoryLevel struct {
	ProductID         uuid.UUID  `json:"productId" bson:"productId"`
	VariantID         *uuid.UUID `json:"variantId" bson:"variantId"`
	SKU               string     `json:"sku" bson:"sku"`
	TotalOnHand       int        `json:"totalOnHand" bson:"totalOnHand"`
	TotalReserved     int        `json:"totalReserved" bson:"totalReserved"`
	TotalAvailable    int        `json:"totalAvailable" bson:"totalAvailable"`
	TotalAllocated    int        `json:"totalAllocated" bson:"totalAllocated"`
	WarehouseCount    int        `json:"warehouseCount" bson:"warehouseCount"`
	BelowReorderPoint bool       `json:"belowReorderPoint" bson:"belowReorderPoint"`
	OutOfStock        bool       `json:"outOfStock" bson:"outOfStock"`
}

func NewInventoryItem(
	tenantID, productID, warehouseID uuid.UUID,
	sku string,
	quantity int,
	unitCost decimal.Decimal,
) *InventoryItem {
	now := time.Now().UTC()
	id := uuid.New()

	return &InventoryItem{
		ID:           id,
		TenantID:     tenantID,
		ProductID:    productID,
		WarehouseID:  warehouseID,
		SKU:          sku,
		Quantity:     quantity,
		ReservedQty:  0,
		AvailableQty: quantity,
		AllocatedQty: 0,
		Status:       InventoryStatusAvailable,
		UnitCost:     unitCost,
		TotalValue:   unitCost.Mul(decimal.NewFromInt(int64(quantity))),
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      0,
	}
}

func (i *InventoryItem) Reserve(quantity int) error {
	if quantity > i.AvailableQty {
		return ErrInsufficientInventory
	}
	i.ReservedQty += quantity
	i.AvailableQty -= quantity
	i.UpdatedAt = time.Now().UTC()
	return nil
}

func (i *InventoryItem) ReleaseReservation(quantity int) {
	i.ReservedQty -= quantity
	i.AvailableQty += quantity
	i.UpdatedAt = time.Now().UTC()
}

func (i *InventoryItem) Allocate(quantity int) error {
	if quantity > i.AvailableQty {
		return ErrInsufficientInventory
	}
	i.AllocatedQty += quantity
	i.AvailableQty -= quantity
	i.UpdatedAt = time.Now().UTC()
	return nil
}

func (i *InventoryItem) Deallocate(quantity int) {
	i.AllocatedQty -= quantity
	i.AvailableQty += quantity
	i.UpdatedAt = time.Now().UTC()
}

func (i *InventoryItem) Receive(quantity int, unitCost decimal.Decimal) {
	i.Quantity += quantity
	i.AvailableQty += quantity
	i.UnitCost = unitCost
	i.TotalValue = unitCost.Mul(decimal.NewFromInt(int64(i.Quantity)))
	i.UpdatedAt = time.Now().UTC()
	i.Status = InventoryStatusAvailable
}

func (i *InventoryItem) Ship(quantity int) error {
	if quantity > i.Quantity {
		return ErrInsufficientInventory
	}
	i.Quantity -= quantity
	i.AvailableQty -= quantity
	i.TotalValue = i.UnitCost.Mul(decimal.NewFromInt(int64(i.Quantity)))
	i.UpdatedAt = time.Now().UTC()
	return nil
}

func (i *InventoryItem) Adjust(adjustment int, reason string) {
	i.Quantity += adjustment
	i.AvailableQty += adjustment
	i.TotalValue = i.UnitCost.Mul(decimal.NewFromInt(int64(i.Quantity)))
	i.UpdatedAt = time.Now().UTC()
	_ = reason
}

func (i *InventoryItem) Count(countedQty int) {
	difference := countedQty - i.Quantity
	i.Quantity = countedQty
	i.AvailableQty = i.AvailableQty + difference
	i.LastCountedAt = new(time.Time)
	*i.LastCountedAt = time.Now().UTC()
	i.UpdatedAt = time.Now().UTC()
}

func NewWarehouse(
	tenantID uuid.UUID,
	name, code string,
	warehouseType WarehouseType,
) *Warehouse {
	now := time.Now().UTC()

	return &Warehouse{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      name,
		Code:      code,
		Type:      warehouseType,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (w *Warehouse) SetAddress(addr Address) {
	w.Address = addr
	w.UpdatedAt = time.Now().UTC()
}

func (w *Warehouse) Deactivate() {
	w.IsActive = false
	w.UpdatedAt = time.Now().UTC()
}

func (w *Warehouse) Activate() {
	w.IsActive = true
	w.UpdatedAt = time.Now().UTC()
}

func NewStockReservation(
	tenantID, productID, warehouseID, referenceID uuid.UUID,
	referenceType string,
	quantity int,
) *StockReservation {
	now := time.Now().UTC()

	return &StockReservation{
		ID:            uuid.New(),
		TenantID:      tenantID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
		Quantity:      quantity,
		Status:        "active",
		ExpiresAt:     nil,
		CreatedAt:     now,
	}
}

func (r *StockReservation) Expire() {
	r.Status = "expired"
	r.ReleasedAt = new(time.Time)
	*r.ReleasedAt = time.Now().UTC()
}

func (r *StockReservation) Release() {
	r.Status = "released"
	r.ReleasedAt = new(time.Time)
	*r.ReleasedAt = time.Now().UTC()
}

func (r *StockReservation) Fulfill() {
	r.Status = "fulfilled"
}

func NewInventoryTransaction(
	tenantID, productID, warehouseID, performedBy uuid.UUID,
	movementType MovementType,
	quantity int,
) *InventoryTransaction {
	return &InventoryTransaction{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ProductID:    productID,
		WarehouseID:  warehouseID,
		MovementType: movementType,
		Quantity:     quantity,
		PerformedBy:  performedBy,
		CreatedAt:    time.Now().UTC(),
	}
}

func (t *InventoryTransaction) SetReference(refType string, refID uuid.UUID) {
	t.ReferenceType = refType
	t.ReferenceID = refID
}

func (t *InventoryTransaction) SetTransfer(fromID, toID uuid.UUID) {
	t.FromLocationID = &fromID
	t.ToLocationID = &toID
}

func (t *InventoryTransaction) SetLotInfo(lot, serial string) {
	t.LotNumber = lot
	t.SerialNumber = serial
}

var ErrInsufficientInventory = &InventoryError{
	Code:    "INSUFFICIENT_INVENTORY",
	Message: "Insufficient inventory available",
}

var ErrNegativeInventory = &InventoryError{
	Code:    "NEGATIVE_INVENTORY",
	Message: "Inventory count would go negative",
}

type InventoryError struct {
	Code    string
	Message string
}

func (e *InventoryError) Error() string {
	return e.Message
}

func (e *InventoryError) Is(target error) bool {
	_, ok := target.(*InventoryError)
	return ok
}

type WarehouseError struct {
	Code    string
	Message string
}

func (e *WarehouseError) Error() string {
	return e.Message
}

func (e *WarehouseError) Is(target error) bool {
	_, ok := target.(*WarehouseError)
	return ok
}

var (
	ErrWarehouseCodeRequired                  = &WarehouseError{Code: "WAREHOUSE_CODE_REQUIRED", Message: "Warehouse code is required"}
	ErrWarehouseNameRequired                  = &WarehouseError{Code: "WAREHOUSE_NAME_REQUIRED", Message: "Warehouse name is required"}
	ErrInvalidWarehouseType                   = &WarehouseError{Code: "INVALID_WAREHOUSE_TYPE", Message: "Invalid warehouse type"}
	ErrLocationCodeRequired                   = &WarehouseError{Code: "LOCATION_CODE_REQUIRED", Message: "Location code is required"}
	ErrLocationPathRequired                   = &WarehouseError{Code: "LOCATION_PATH_REQUIRED", Message: "Zone, aisle, rack, and bin are required"}
	ErrInvalidLocationStatus                  = &WarehouseError{Code: "INVALID_LOCATION_STATUS", Message: "Invalid location status"}
	ErrInvalidOperationType                   = &WarehouseError{Code: "INVALID_OPERATION_TYPE", Message: "Invalid operation type"}
	ErrCapacityCannotBeLessThanUsed           = &WarehouseError{Code: "CAPACITY_CANNOT_BE_LESS", Message: "Capacity cannot be less than used capacity"}
	ErrCapacityCannotBeLessThanStock          = &WarehouseError{Code: "CAPACITY_CANNOT_BE_LESS_STOCK", Message: "Capacity cannot be less than current stock"}
	ErrCapacityExceeded                       = &WarehouseError{Code: "CAPACITY_EXCEEDED", Message: "Warehouse capacity exceeded"}
	ErrLocationCapacityExceeded               = &WarehouseError{Code: "LOCATION_CAPACITY_EXCEEDED", Message: "Location capacity exceeded"}
	ErrCannotActivateClosedWarehouse          = &WarehouseError{Code: "CANNOT_ACTIVATE_CLOSED", Message: "Cannot activate a closed warehouse"}
	ErrCannotDeactivateWarehouseWithLocations = &WarehouseError{Code: "CANNOT_DEACTIVATE_WITH_LOCATIONS", Message: "Cannot deactivate warehouse with active locations"}
	ErrCannotCloseWarehouseWithLocations      = &WarehouseError{Code: "CANNOT_CLOSE_WITH_LOCATIONS", Message: "Cannot close warehouse with active locations"}
	ErrNoLocationsToRemove                    = &WarehouseError{Code: "NO_LOCATIONS_TO_REMOVE", Message: "No locations to remove"}
	ErrCannotPickReservedStock                = &WarehouseError{Code: "CANNOT_PICK_RESERVED", Message: "Cannot pick more than available stock"}
	ErrCannotReleaseMoreThanReserved          = &WarehouseError{Code: "CANNOT_RELEASE_MORE", Message: "Cannot release more than reserved"}
	ErrCannotDeactivateLocationWithStock      = &WarehouseError{Code: "CANNOT_DEACTIVATE_WITH_STOCK", Message: "Cannot deactivate location with stock"}
	ErrOperationItemNotFound                  = &WarehouseError{Code: "OPERATION_ITEM_NOT_FOUND", Message: "Operation item not found"}
)

type WarehouseOperation struct {
	ID            uuid.UUID       `json:"id" bson:"_id"`
	TenantID      uuid.UUID       `json:"tenantId" bson:"tenantId"`
	WarehouseID   uuid.UUID       `json:"warehouseId" bson:"warehouseId"`
	Type          OperationType   `json:"type" bson:"type"`
	ReferenceType string          `json:"referenceType" bson:"referenceType"`
	ReferenceID   uuid.UUID       `json:"referenceId" bson:"referenceId"`
	Status        string          `json:"status" bson:"status"`
	Priority      int             `json:"priority" bson:"priority"`
	AssignedTo    *uuid.UUID      `json:"assignedTo" bson:"assignedTo"`
	Items         []OperationItem `json:"items" bson:"items"`
	Notes         string          `json:"notes" bson:"notes"`
	StartedAt     *time.Time      `json:"startedAt" bson:"startedAt"`
	CompletedAt   *time.Time      `json:"completedAt" bson:"completedAt"`
	CreatedBy     uuid.UUID       `json:"createdBy" bson:"createdBy"`
	CreatedAt     time.Time       `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt" bson:"updatedAt"`
}

type OperationItem struct {
	ID           uuid.UUID  `json:"id" bson:"_id"`
	ProductID    uuid.UUID  `json:"productId" bson:"productId"`
	VariantID    *uuid.UUID `json:"variantId" bson:"variantId"`
	LocationID   uuid.UUID  `json:"locationId" bson:"locationId"`
	Quantity     int        `json:"quantity" bson:"quantity"`
	QuantityDone int        `json:"quantityDone" bson:"quantityDone"`
	Status       string     `json:"status" bson:"status"`
}

func NewWarehouseOperation(
	tenantID, warehouseID, createdBy uuid.UUID,
	opType OperationType,
	referenceType string,
	referenceID uuid.UUID,
) (*WarehouseOperation, error) {
	if !opType.IsValid() {
		return nil, ErrInvalidOperationType
	}

	now := time.Now().UTC()
	return &WarehouseOperation{
		ID:            uuid.New(),
		TenantID:      tenantID,
		WarehouseID:   warehouseID,
		Type:          opType,
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
		Status:        "pending",
		Priority:      5,
		Items:         []OperationItem{},
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (o *WarehouseOperation) AddItem(item OperationItem) {
	o.Items = append(o.Items, item)
	o.UpdatedAt = time.Now().UTC()
}

func (o *WarehouseOperation) AssignTo(userID uuid.UUID) {
	o.AssignedTo = &userID
	o.UpdatedAt = time.Now().UTC()
}

func (o *WarehouseOperation) Start() {
	now := time.Now().UTC()
	o.StartedAt = &now
	o.Status = "in_progress"
	o.UpdatedAt = now
}

func (o *WarehouseOperation) CompleteItem(itemID uuid.UUID, quantity int) error {
	for i := range o.Items {
		if o.Items[i].ID == itemID {
			o.Items[i].QuantityDone += quantity
			if o.Items[i].QuantityDone >= o.Items[i].Quantity {
				o.Items[i].Status = "completed"
			} else {
				o.Items[i].Status = "partial"
			}
			o.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return ErrOperationItemNotFound
}

func (o *WarehouseOperation) IsComplete() bool {
	for _, item := range o.Items {
		if item.Status != "completed" {
			return false
		}
	}
	return true
}

func (o *WarehouseOperation) Complete() {
	now := time.Now().UTC()
	o.CompletedAt = &now
	o.Status = "completed"
	o.UpdatedAt = now
}

func (o *WarehouseOperation) Cancel(reason string) {
	o.Status = "cancelled"
	o.Notes = reason
	o.UpdatedAt = time.Now().UTC()
}

func (o *WarehouseOperation) SetPriority(priority int) {
	if priority < 1 {
		priority = 1
	} else if priority > 10 {
		priority = 10
	}
	o.Priority = priority
	o.UpdatedAt = time.Now().UTC()
}

type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *Warehouse) error
	Update(ctx context.Context, warehouse *Warehouse) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Warehouse, error)
	FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Warehouse, error)
	FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]*Warehouse, error)
	FindActive(ctx context.Context, tenantID uuid.UUID) ([]*Warehouse, error)
}

type LocationRepository interface {
	Create(ctx context.Context, location *WarehouseLocation) error
	Update(ctx context.Context, location *WarehouseLocation) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*WarehouseLocation, error)
	FindByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*WarehouseLocation, error)
	FindByPath(ctx context.Context, warehouseID uuid.UUID, zone, aisle, rack, bin string) (*WarehouseLocation, error)
	FindByBarcode(ctx context.Context, barcode string) (*WarehouseLocation, error)
	FindAvailable(ctx context.Context, warehouseID uuid.UUID, quantity int) ([]*WarehouseLocation, error)
}

type OperationRepository interface {
	Create(ctx context.Context, operation *WarehouseOperation) error
	Update(ctx context.Context, operation *WarehouseOperation) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*WarehouseOperation, error)
	FindByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*WarehouseOperation, error)
	FindByStatus(ctx context.Context, warehouseID uuid.UUID, status string) ([]*WarehouseOperation, error)
	FindPending(ctx context.Context, warehouseID uuid.UUID) ([]*WarehouseOperation, error)
	FindByReference(ctx context.Context, referenceType string, referenceID uuid.UUID) ([]*WarehouseOperation, error)
}

type InventoryRepository interface {
	Create(ctx context.Context, item *InventoryItem) error
	Update(ctx context.Context, item *InventoryItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*InventoryItem, error)
	FindByProductAndWarehouse(ctx context.Context, productID, warehouseID uuid.UUID) (*InventoryItem, error)
	FindByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*InventoryItem, error)
	FindByProduct(ctx context.Context, productID uuid.UUID) ([]*InventoryItem, error)
	FindLowStock(ctx context.Context, tenantID uuid.UUID) ([]*InventoryItem, error)
}

type ReservationRepository interface {
	Create(ctx context.Context, reservation *StockReservation) error
	Update(ctx context.Context, reservation *StockReservation) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*StockReservation, error)
	FindByProduct(ctx context.Context, productID uuid.UUID) ([]*StockReservation, error)
	FindByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*StockReservation, error)
	FindByReference(ctx context.Context, referenceType string, referenceID uuid.UUID) ([]*StockReservation, error)
	FindActiveByProduct(ctx context.Context, productID uuid.UUID) ([]*StockReservation, error)
	FindExpired(ctx context.Context, tenantID uuid.UUID) ([]*StockReservation, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *InventoryTransaction) error
	Update(ctx context.Context, transaction *InventoryTransaction) error
	FindByID(ctx context.Context, id uuid.UUID) (*InventoryTransaction, error)
	FindByProduct(ctx context.Context, productID uuid.UUID) ([]*InventoryTransaction, error)
	FindByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*InventoryTransaction, error)
	FindByReference(ctx context.Context, referenceType string, referenceID uuid.UUID) ([]*InventoryTransaction, error)
	FindByMovementType(ctx context.Context, movementType MovementType) ([]*InventoryTransaction, error)
	FindByDateRange(ctx context.Context, start, end time.Time) ([]*InventoryTransaction, error)
}
