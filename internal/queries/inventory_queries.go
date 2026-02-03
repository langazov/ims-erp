package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type InventoryQueryHandler struct {
	inventoryRepo   domain.InventoryRepository
	reservationRepo domain.ReservationRepository
	transactionRepo domain.TransactionRepository
	warehouseRepo   domain.WarehouseRepository
	cache           *repository.Cache
	logger          *logger.Logger
	tracer          trace.Tracer
}

func NewInventoryQueryHandler(
	inventoryRepo domain.InventoryRepository,
	reservationRepo domain.ReservationRepository,
	transactionRepo domain.TransactionRepository,
	warehouseRepo domain.WarehouseRepository,
	cache *repository.Cache,
	log *logger.Logger,
) *InventoryQueryHandler {
	return &InventoryQueryHandler{
		inventoryRepo:   inventoryRepo,
		reservationRepo: reservationRepo,
		transactionRepo: transactionRepo,
		warehouseRepo:   warehouseRepo,
		cache:           cache,
		logger:          log,
		tracer:          otel.Tracer("inventory-query-handler"),
	}
}

// Query types
type GetInventoryByProductQuery struct {
	ProductID   string
	TenantID    string
	WarehouseID string
}

type ListInventoryQuery struct {
	TenantID    string
	WarehouseID string
	ProductID   string
	Status      string
	Page        int
	PageSize    int
	SortBy      string
	SortOrder   string
}

type GetLowStockQuery struct {
	TenantID string
	Page     int
	PageSize int
}

type GetInventoryTransactionsQuery struct {
	ProductID    string
	WarehouseID  string
	TenantID     string
	MovementType string
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	PageSize     int
}

type GetReservationsQuery struct {
	ProductID     string
	WarehouseID   string
	TenantID      string
	Status        string
	ReferenceType string
	ReferenceID   string
	Page          int
	PageSize      int
}

// Result types
type InventoryItemSummary struct {
	ID            string     `json:"id" bson:"_id"`
	TenantID      string     `json:"tenantId" bson:"tenantId"`
	ProductID     string     `json:"productId" bson:"productId"`
	SKU           string     `json:"sku" bson:"sku"`
	WarehouseID   string     `json:"warehouseId" bson:"warehouseId"`
	LocationID    string     `json:"locationId" bson:"locationId"`
	Quantity      int        `json:"quantity" bson:"quantity"`
	ReservedQty   int        `json:"reservedQty" bson:"reservedQty"`
	AvailableQty  int        `json:"availableQty" bson:"availableQty"`
	AllocatedQty  int        `json:"allocatedQty" bson:"allocatedQty"`
	Status        string     `json:"status" bson:"status"`
	UnitCost      string     `json:"unitCost" bson:"unitCost"`
	TotalValue    string     `json:"totalValue" bson:"totalValue"`
	LastCountedAt *time.Time `json:"lastCountedAt" bson:"lastCountedAt"`
	CreatedAt     time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt" bson:"updatedAt"`
}

type InventoryItemDetail struct {
	ID             string     `json:"id" bson:"_id"`
	TenantID       string     `json:"tenantId" bson:"tenantId"`
	ProductID      string     `json:"productId" bson:"productId"`
	VariantID      *string    `json:"variantId" bson:"variantId"`
	SKU            string     `json:"sku" bson:"sku"`
	WarehouseID    string     `json:"warehouseId" bson:"warehouseId"`
	LocationID     string     `json:"locationId" bson:"locationId"`
	BinID          string     `json:"binId" bson:"binId"`
	LotNumber      string     `json:"lotNumber" bson:"lotNumber"`
	SerialNumber   string     `json:"serialNumber" bson:"serialNumber"`
	BatchNumber    string     `json:"batchNumber" bson:"batchNumber"`
	ExpirationDate *time.Time `json:"expirationDate" bson:"expirationDate"`
	Quantity       int        `json:"quantity" bson:"quantity"`
	ReservedQty    int        `json:"reservedQty" bson:"reservedQty"`
	AvailableQty   int        `json:"availableQty" bson:"availableQty"`
	AllocatedQty   int        `json:"allocatedQty" bson:"allocatedQty"`
	Status         string     `json:"status" bson:"status"`
	UnitCost       string     `json:"unitCost" bson:"unitCost"`
	TotalValue     string     `json:"totalValue" bson:"totalValue"`
	LastCountedAt  *time.Time `json:"lastCountedAt" bson:"lastCountedAt"`
	CreatedAt      time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt" bson:"updatedAt"`
}

type ListInventoryResult struct {
	Items      []InventoryItemSummary `json:"items"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"pageSize"`
	TotalPages int                    `json:"totalPages"`
}

type TransactionSummary struct {
	ID            string    `json:"id" bson:"_id"`
	TenantID      string    `json:"tenantId" bson:"tenantId"`
	ProductID     string    `json:"productId" bson:"productId"`
	WarehouseID   string    `json:"warehouseId" bson:"warehouseId"`
	MovementType  string    `json:"movementType" bson:"movementType"`
	Quantity      int       `json:"quantity" bson:"quantity"`
	ReferenceType string    `json:"referenceType" bson:"referenceType"`
	ReferenceID   string    `json:"referenceId" bson:"referenceId"`
	LotNumber     string    `json:"lotNumber" bson:"lotNumber"`
	SerialNumber  string    `json:"serialNumber" bson:"serialNumber"`
	Reason        string    `json:"reason" bson:"reason"`
	PerformedBy   string    `json:"performedBy" bson:"performedBy"`
	CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
}

type ListTransactionsResult struct {
	Transactions []TransactionSummary `json:"transactions"`
	Total        int64                `json:"total"`
	Page         int                  `json:"page"`
	PageSize     int                  `json:"pageSize"`
	TotalPages   int                  `json:"totalPages"`
}

type ReservationSummary struct {
	ID            string     `json:"id" bson:"_id"`
	TenantID      string     `json:"tenantId" bson:"tenantId"`
	ProductID     string     `json:"productId" bson:"productId"`
	VariantID     *string    `json:"variantId" bson:"variantId"`
	WarehouseID   string     `json:"warehouseId" bson:"warehouseId"`
	ReferenceType string     `json:"referenceType" bson:"referenceType"`
	ReferenceID   string     `json:"referenceId" bson:"referenceId"`
	Quantity      int        `json:"quantity" bson:"quantity"`
	Status        string     `json:"status" bson:"status"`
	ExpiresAt     *time.Time `json:"expiresAt" bson:"expiresAt"`
	CreatedAt     time.Time  `json:"createdAt" bson:"createdAt"`
	ReleasedAt    *time.Time `json:"releasedAt" bson:"releasedAt"`
}

type ListReservationsResult struct {
	Reservations []ReservationSummary `json:"reservations"`
	Total        int64                `json:"total"`
	Page         int                  `json:"page"`
	PageSize     int                  `json:"pageSize"`
	TotalPages   int                  `json:"totalPages"`
}

type StockLevel struct {
	ProductID         string `json:"productId" bson:"productId"`
	SKU               string `json:"sku" bson:"sku"`
	ProductName       string `json:"productName" bson:"productName"`
	WarehouseID       string `json:"warehouseId" bson:"warehouseId"`
	WarehouseName     string `json:"warehouseName" bson:"warehouseName"`
	QuantityOnHand    int    `json:"quantityOnHand" bson:"quantityOnHand"`
	QuantityReserved  int    `json:"quantityReserved" bson:"quantityReserved"`
	QuantityAvailable int    `json:"quantityAvailable" bson:"quantityAvailable"`
	ReorderPoint      int    `json:"reorderPoint" bson:"reorderPoint"`
	ReorderQuantity   int    `json:"reorderQuantity" bson:"reorderQuantity"`
	IsLowStock        bool   `json:"isLowStock" bson:"isLowStock"`
	IsOutOfStock      bool   `json:"isOutOfStock" bson:"isOutOfStock"`
}

type LowStockReport struct {
	TenantID      string       `json:"tenantId" bson:"tenantId"`
	GeneratedAt   time.Time    `json:"generatedAt" bson:"generatedAt"`
	Items         []StockLevel `json:"items" bson:"items"`
	TotalItems    int          `json:"totalItems" bson:"totalItems"`
	CriticalCount int          `json:"criticalCount" bson:"criticalCount"`
	WarningCount  int          `json:"warningCount" bson:"warningCount"`
}

// GetInventoryByProduct retrieves inventory for a specific product
func (h *InventoryQueryHandler) GetInventoryByProduct(ctx context.Context, query *GetInventoryByProductQuery) (*InventoryItemDetail, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_inventory_by_product",
		trace.WithAttributes(
			attribute.String("product_id", query.ProductID),
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("inventory:detail:%s:%s", query.ProductID, query.WarehouseID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var item InventoryItemDetail
		if err := json.Unmarshal(cached, &item); err == nil {
			return &item, nil
		}
	}

	// Return placeholder
	item := &InventoryItemDetail{
		ProductID: query.ProductID,
		TenantID:  query.TenantID,
	}

	if data, err := json.Marshal(item); err == nil {
		h.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return item, nil
}

// ListInventory retrieves a paginated list of inventory items
func (h *InventoryQueryHandler) ListInventory(ctx context.Context, query *ListInventoryQuery) (*ListInventoryResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.list_inventory",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
			attribute.String("warehouse_id", query.WarehouseID),
			attribute.Int("page", query.Page),
			attribute.Int("page_size", query.PageSize),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("inventory:list:%s:%s:%s:%d:%d",
		query.TenantID, query.WarehouseID, query.Status, query.Page, query.PageSize)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var result ListInventoryResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}

	// Return empty result for now
	result := &ListInventoryResult{
		Items:      []InventoryItemSummary{},
		Total:      0,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 0,
	}

	if data, err := json.Marshal(result); err == nil {
		h.cache.Set(ctx, cacheKey, data, 2*time.Minute)
	}

	return result, nil
}

// GetLowStock retrieves inventory items below reorder point
func (h *InventoryQueryHandler) GetLowStock(ctx context.Context, query *GetLowStockQuery) (*LowStockReport, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_low_stock",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("inventory:lowstock:%s", query.TenantID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var report LowStockReport
		if err := json.Unmarshal(cached, &report); err == nil {
			return &report, nil
		}
	}

	// Return empty report
	report := &LowStockReport{
		TenantID:      query.TenantID,
		GeneratedAt:   time.Now().UTC(),
		Items:         []StockLevel{},
		TotalItems:    0,
		CriticalCount: 0,
		WarningCount:  0,
	}

	if data, err := json.Marshal(report); err == nil {
		h.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return report, nil
}

// GetInventoryTransactions retrieves transaction history
func (h *InventoryQueryHandler) GetInventoryTransactions(ctx context.Context, query *GetInventoryTransactionsQuery) (*ListTransactionsResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_inventory_transactions",
		trace.WithAttributes(
			attribute.String("product_id", query.ProductID),
			attribute.String("movement_type", query.MovementType),
		),
	)
	defer span.End()

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}

	// Return empty result
	result := &ListTransactionsResult{
		Transactions: []TransactionSummary{},
		Total:        0,
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   0,
	}

	return result, nil
}

// GetReservations retrieves stock reservations
func (h *InventoryQueryHandler) GetReservations(ctx context.Context, query *GetReservationsQuery) (*ListReservationsResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_reservations",
		trace.WithAttributes(
			attribute.String("product_id", query.ProductID),
			attribute.String("status", query.Status),
		),
	)
	defer span.End()

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}

	// Return empty result
	result := &ListReservationsResult{
		Reservations: []ReservationSummary{},
		Total:        0,
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   0,
	}

	return result, nil
}

// GetStockLevel retrieves current stock level for a product
func (h *InventoryQueryHandler) GetStockLevel(ctx context.Context, productID, warehouseID, tenantID string) (*StockLevel, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_stock_level",
		trace.WithAttributes(
			attribute.String("product_id", productID),
			attribute.String("warehouse_id", warehouseID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("inventory:stocklevel:%s:%s", productID, warehouseID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var level StockLevel
		if err := json.Unmarshal(cached, &level); err == nil {
			return &level, nil
		}
	}

	// Return placeholder
	level := &StockLevel{
		ProductID:   productID,
		WarehouseID: warehouseID,
	}

	if data, err := json.Marshal(level); err == nil {
		h.cache.Set(ctx, cacheKey, data, 1*time.Minute)
	}

	return level, nil
}

// GetGlobalInventory retrieves inventory aggregated across all warehouses
func (h *InventoryQueryHandler) GetGlobalInventory(ctx context.Context, tenantID string, productID string) (*domain.InventoryLevel, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_global_inventory",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID),
			attribute.String("product_id", productID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("inventory:global:%s:%s", tenantID, productID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var level domain.InventoryLevel
		if err := json.Unmarshal(cached, &level); err == nil {
			return &level, nil
		}
	}

	// Return empty level
	level := &domain.InventoryLevel{
		ProductID:   parseUUID(productID),
		SKU:         "",
		TotalOnHand: 0,
	}

	if data, err := json.Marshal(level); err == nil {
		h.cache.Set(ctx, cacheKey, data, 2*time.Minute)
	}

	return level, nil
}

func parseUUID(s string) [16]byte {
	// Simple placeholder - in real implementation, use proper UUID parsing
	var b [16]byte
	copy(b[:], s)
	return b
}
