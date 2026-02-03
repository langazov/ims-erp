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

type WarehouseQueryHandler struct {
	warehouseRepo domain.WarehouseRepository
	locationRepo  domain.LocationRepository
	operationRepo domain.OperationRepository
	cache         *repository.Cache
	logger        *logger.Logger
	tracer        trace.Tracer
}

func NewWarehouseQueryHandler(
	warehouseRepo domain.WarehouseRepository,
	locationRepo domain.LocationRepository,
	operationRepo domain.OperationRepository,
	cache *repository.Cache,
	log *logger.Logger,
) *WarehouseQueryHandler {
	return &WarehouseQueryHandler{
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
		operationRepo: operationRepo,
		cache:         cache,
		logger:        log,
		tracer:        otel.Tracer("warehouse-query-handler"),
	}
}

// Query types
type GetWarehouseByIDQuery struct {
	WarehouseID string
	TenantID    string
}

type ListWarehousesQuery struct {
	TenantID  string
	Page      int
	PageSize  int
	Status    string
	Type      string
	Search    string
	SortBy    string
	SortOrder string
}

type GetWarehouseLocationsQuery struct {
	WarehouseID string
	TenantID    string
	IsActive    *bool
}

type GetLocationByIDQuery struct {
	LocationID string
	TenantID   string
}

type GetWarehouseOperationsQuery struct {
	WarehouseID string
	TenantID    string
	Status      string
	Type        string
	Page        int
	PageSize    int
}

type GetWarehouseInventoryQuery struct {
	WarehouseID string
	TenantID    string
	ProductID   string
	Page        int
	PageSize    int
}

// Result types
type WarehouseSummary struct {
	ID                 string    `json:"id" bson:"_id"`
	TenantID           string    `json:"tenantId" bson:"tenantId"`
	Name               string    `json:"name" bson:"name"`
	Code               string    `json:"code" bson:"code"`
	Type               string    `json:"type" bson:"type"`
	Status             string    `json:"status" bson:"status"`
	Capacity           int       `json:"capacity" bson:"capacity"`
	CurrentUtilization float64   `json:"currentUtilization" bson:"currentUtilization"`
	IsActive           bool      `json:"isActive" bson:"isActive"`
	IsPrimary          bool      `json:"isPrimary" bson:"isPrimary"`
	LocationCount      int       `json:"locationCount" bson:"locationCount"`
	CreatedAt          time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt" bson:"updatedAt"`
}

type WarehouseDetail struct {
	ID                 string            `json:"id" bson:"_id"`
	TenantID           string            `json:"tenantId" bson:"tenantId"`
	Name               string            `json:"name" bson:"name"`
	Code               string            `json:"code" bson:"code"`
	Type               string            `json:"type" bson:"type"`
	Address            domain.Address    `json:"address" bson:"address"`
	Status             string            `json:"status" bson:"status"`
	Capacity           int               `json:"capacity" bson:"capacity"`
	CurrentUtilization float64           `json:"currentUtilization" bson:"currentUtilization"`
	IsActive           bool              `json:"isActive" bson:"isActive"`
	IsPrimary          bool              `json:"isPrimary" bson:"isPrimary"`
	ManagerID          *string           `json:"managerId" bson:"managerId"`
	ContactEmail       string            `json:"contactEmail" bson:"contactEmail"`
	ContactPhone       string            `json:"contactPhone" bson:"contactPhone"`
	OperatingHours     string            `json:"operatingHours" bson:"operatingHours"`
	Locations          []LocationSummary `json:"locations" bson:"locations"`
	CreatedAt          time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt" bson:"updatedAt"`
}

type LocationSummary struct {
	ID           string `json:"id" bson:"_id"`
	Name         string `json:"name" bson:"name"`
	Code         string `json:"code" bson:"code"`
	Type         string `json:"type" bson:"type"`
	Zone         string `json:"zone" bson:"zone"`
	Aisle        string `json:"aisle" bson:"aisle"`
	Rack         string `json:"rack" bson:"rack"`
	Bin          string `json:"bin" bson:"bin"`
	Capacity     int    `json:"capacity" bson:"capacity"`
	CurrentStock int    `json:"currentStock" bson:"currentStock"`
	IsActive     bool   `json:"isActive" bson:"isActive"`
}

type LocationDetail struct {
	ID           string    `json:"id" bson:"_id"`
	WarehouseID  string    `json:"warehouseId" bson:"warehouseId"`
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

type OperationSummary struct {
	ID            string     `json:"id" bson:"_id"`
	Type          string     `json:"type" bson:"type"`
	Status        string     `json:"status" bson:"status"`
	Priority      int        `json:"priority" bson:"priority"`
	ReferenceType string     `json:"referenceType" bson:"referenceType"`
	ReferenceID   string     `json:"referenceId" bson:"referenceId"`
	ItemCount     int        `json:"itemCount" bson:"itemCount"`
	CreatedAt     time.Time  `json:"createdAt" bson:"createdAt"`
	StartedAt     *time.Time `json:"startedAt" bson:"startedAt"`
	CompletedAt   *time.Time `json:"completedAt" bson:"completedAt"`
}

type ListWarehousesResult struct {
	Warehouses []WarehouseSummary `json:"warehouses"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
	TotalPages int                `json:"totalPages"`
}

type ListOperationsResult struct {
	Operations []OperationSummary `json:"operations"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
	TotalPages int                `json:"totalPages"`
}

type WarehouseOverview struct {
	WarehouseID        string  `json:"warehouseId" bson:"warehouseId"`
	WarehouseName      string  `json:"warehouseName" bson:"warehouseName"`
	Status             string  `json:"status" bson:"status"`
	TotalLocations     int     `json:"totalLocations" bson:"totalLocations"`
	ActiveLocations    int     `json:"activeLocations" bson:"activeLocations"`
	TotalCapacity      int     `json:"totalCapacity" bson:"totalCapacity"`
	UsedCapacity       int     `json:"usedCapacity" bson:"usedCapacity"`
	UtilizationPercent float64 `json:"utilizationPercent" bson:"utilizationPercent"`
}

// GetWarehouseByID retrieves a single warehouse by ID
func (h *WarehouseQueryHandler) GetWarehouseByID(ctx context.Context, query *GetWarehouseByIDQuery) (*WarehouseDetail, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_warehouse_by_id",
		trace.WithAttributes(
			attribute.String("warehouse_id", query.WarehouseID),
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("warehouse:detail:%s", query.WarehouseID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var warehouse WarehouseDetail
		if err := json.Unmarshal(cached, &warehouse); err == nil {
			return &warehouse, nil
		}
	}

	// This would typically query a read model store
	// For now, return a placeholder
	h.logger.Info("Fetching warehouse", "warehouse_id", query.WarehouseID)

	// Cache the result
	placeholder := &WarehouseDetail{
		ID:       query.WarehouseID,
		TenantID: query.TenantID,
	}

	if data, err := json.Marshal(placeholder); err == nil {
		h.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return placeholder, nil
}

// ListWarehouses retrieves a paginated list of warehouses
func (h *WarehouseQueryHandler) ListWarehouses(ctx context.Context, query *ListWarehousesQuery) (*ListWarehousesResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.list_warehouses",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
			attribute.Int("page", query.Page),
			attribute.Int("page_size", query.PageSize),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("warehouse:list:%s:%d:%d:%s:%s",
		query.TenantID, query.Page, query.PageSize, query.Status, query.Type)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var result ListWarehousesResult
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
	result := &ListWarehousesResult{
		Warehouses: []WarehouseSummary{},
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

// GetWarehouseLocations retrieves all locations for a warehouse
func (h *WarehouseQueryHandler) GetWarehouseLocations(ctx context.Context, query *GetWarehouseLocationsQuery) ([]LocationSummary, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_warehouse_locations",
		trace.WithAttributes(
			attribute.String("warehouse_id", query.WarehouseID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("warehouse:locations:%s", query.WarehouseID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var locations []LocationSummary
		if err := json.Unmarshal(cached, &locations); err == nil {
			return locations, nil
		}
	}

	// Return empty slice for now
	locations := []LocationSummary{}

	if data, err := json.Marshal(locations); err == nil {
		h.cache.Set(ctx, cacheKey, data, 3*time.Minute)
	}

	return locations, nil
}

// GetLocationByID retrieves a single location by ID
func (h *WarehouseQueryHandler) GetLocationByID(ctx context.Context, query *GetLocationByIDQuery) (*LocationDetail, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_location_by_id",
		trace.WithAttributes(
			attribute.String("location_id", query.LocationID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("location:detail:%s", query.LocationID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var location LocationDetail
		if err := json.Unmarshal(cached, &location); err == nil {
			return &location, nil
		}
	}

	// Return placeholder
	location := &LocationDetail{
		ID: query.LocationID,
	}

	if data, err := json.Marshal(location); err == nil {
		h.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return location, nil
}

// GetWarehouseOperations retrieves operations for a warehouse
func (h *WarehouseQueryHandler) GetWarehouseOperations(ctx context.Context, query *GetWarehouseOperationsQuery) (*ListOperationsResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_warehouse_operations",
		trace.WithAttributes(
			attribute.String("warehouse_id", query.WarehouseID),
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

	// Return empty result for now
	result := &ListOperationsResult{
		Operations: []OperationSummary{},
		Total:      0,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 0,
	}

	return result, nil
}

// GetWarehouseOverview retrieves an overview of warehouse statistics
func (h *WarehouseQueryHandler) GetWarehouseOverview(ctx context.Context, warehouseID, tenantID string) (*WarehouseOverview, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_warehouse_overview",
		trace.WithAttributes(
			attribute.String("warehouse_id", warehouseID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("warehouse:overview:%s", warehouseID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var overview WarehouseOverview
		if err := json.Unmarshal(cached, &overview); err == nil {
			return &overview, nil
		}
	}

	// Return placeholder
	overview := &WarehouseOverview{
		WarehouseID:     warehouseID,
		TotalLocations:  0,
		ActiveLocations: 0,
	}

	if data, err := json.Marshal(overview); err == nil {
		h.cache.Set(ctx, cacheKey, data, 1*time.Minute)
	}

	return overview, nil
}
