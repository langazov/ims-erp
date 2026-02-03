package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type InvoiceQueryHandler struct {
	readModelStore *repository.ReadModelStore
	cache          *repository.Cache
	logger         *logger.Logger
	tracer         trace.Tracer
}

func NewInvoiceQueryHandler(
	readModelStore *repository.ReadModelStore,
	cache *repository.Cache,
	log *logger.Logger,
) *InvoiceQueryHandler {
	return &InvoiceQueryHandler{
		readModelStore: readModelStore,
		cache:          cache,
		logger:         log,
		tracer:         otel.Tracer("invoice-query-handler"),
	}
}

type GetInvoiceByIDQuery struct {
	InvoiceID string
	TenantID  string
}

type ListInvoicesQuery struct {
	TenantID  string
	ClientID  string
	Page      int
	PageSize  int
	Search    string
	Status    string
	Type      string
	StartDate time.Time
	EndDate   time.Time
	SortBy    string
	SortOrder string
}

type SearchInvoicesQuery struct {
	TenantID string
	Term     string
	Limit    int
}

type GetOverdueInvoicesQuery struct {
	TenantID string
	Page     int
	PageSize int
}

type GetInvoiceStatsQuery struct {
	TenantID  string
	StartDate time.Time
	EndDate   time.Time
}

type ListInvoicesResult struct {
	Invoices   []events.InvoiceSummary `json:"invoices"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"pageSize"`
	TotalPages int                     `json:"totalPages"`
}

type InvoiceStats struct {
	TenantID         string    `json:"tenantId"`
	TotalInvoices    int64     `json:"totalInvoices"`
	PendingCount     int64     `json:"pendingCount"`
	SentCount        int64     `json:"sentCount"`
	PaidCount        int64     `json:"paidCount"`
	OverdueCount     int64     `json:"overdueCount"`
	CancelledCount   int64     `json:"cancelledCount"`
	TotalAmount      string    `json:"totalAmount"`
	TotalPaid        string    `json:"totalPaid"`
	TotalOutstanding string    `json:"totalOutstanding"`
	PeriodStart      time.Time `json:"periodStart,omitempty"`
	PeriodEnd        time.Time `json:"periodEnd,omitempty"`
}

func (h *InvoiceQueryHandler) GetInvoiceByID(ctx context.Context, query *GetInvoiceByIDQuery) (*events.InvoiceSummary, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_invoice_by_id",
		trace.WithAttributes(
			attribute.String("invoice_id", query.InvoiceID),
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("invoice:summary:%s", query.InvoiceID)
	if cached, err := h.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		return nil, nil
	}

	filter := map[string]interface{}{
		"_id":      query.InvoiceID,
		"tenantId": query.TenantID,
	}

	result, err := h.readModelStore.FindOne(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	if result == nil {
		return nil, nil
	}

	invoice, ok := result.(events.InvoiceSummary)
	if !ok {
		return nil, fmt.Errorf("invalid invoice data")
	}

	h.cache.Set(ctx, cacheKey, invoice, 5*time.Minute)

	return &invoice, nil
}

func (h *InvoiceQueryHandler) ListInvoices(ctx context.Context, query *ListInvoicesQuery) (*ListInvoicesResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.list_invoices",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
			attribute.Int("page", query.Page),
			attribute.Int("page_size", query.PageSize),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("invoice:list:%s:%s:%d:%d:%s:%s:%s",
		query.TenantID, query.ClientID, query.Page, query.PageSize, query.Search, query.Status, query.Type)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var result ListInvoicesResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	filter := map[string]interface{}{
		"tenantId": query.TenantID,
	}

	if query.ClientID != "" {
		filter["clientId"] = query.ClientID
	}

	if query.Search != "" {
		filter["$or"] = []map[string]interface{}{
			{"invoiceNumber": map[string]interface{}{"$regex": query.Search, "$options": "i"}},
			{"clientName": map[string]interface{}{"$regex": query.Search, "$options": "i"}},
		}
	}

	if query.Status != "" {
		filter["status"] = query.Status
	}

	if query.Type != "" {
		filter["type"] = query.Type
	}

	if !query.StartDate.IsZero() && !query.EndDate.IsZero() {
		filter["issueDate"] = map[string]interface{}{
			"$gte": query.StartDate,
			"$lte": query.EndDate,
		}
	}

	total, err := h.readModelStore.Count(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to count invoices: %w", err)
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	skip := (query.Page - 1) * pageSize

	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "issueDate"
	}

	findOpts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(map[string]int{sortBy: getSortOrder(query.SortOrder)})

	results, err := h.readModelStore.Find(ctx, filter, findOpts)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}

	invoices := make([]events.InvoiceSummary, 0, len(results))
	for _, r := range results {
		if invoice, ok := r.(events.InvoiceSummary); ok {
			invoices = append(invoices, invoice)
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	result := &ListInvoicesResult{
		Invoices:   invoices,
		Total:      total,
		Page:       query.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	if data, err := json.Marshal(result); err == nil {
		h.cache.Set(ctx, cacheKey, data, 2*time.Minute)
	}

	return result, nil
}

func (h *InvoiceQueryHandler) SearchInvoices(ctx context.Context, query *SearchInvoicesQuery) ([]events.InvoiceSummary, error) {
	ctx, span := h.tracer.Start(ctx, "query.search_invoices",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
			attribute.String("term", query.Term),
		),
	)
	defer span.End()

	limit := query.Limit
	if limit <= 0 {
		limit = 10
	}

	filter := map[string]interface{}{
		"tenantId": query.TenantID,
		"$or": []map[string]interface{}{
			{"invoiceNumber": map[string]interface{}{"$regex": query.Term, "$options": "i"}},
			{"clientName": map[string]interface{}{"$regex": query.Term, "$options": "i"}},
			{"notes": map[string]interface{}{"$regex": query.Term, "$options": "i"}},
		},
	}

	findOpts := options.Find().
		SetLimit(int64(limit)).
		SetSort(map[string]int{"issueDate": -1})

	results, err := h.readModelStore.Find(ctx, filter, findOpts)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to search invoices: %w", err)
	}

	invoices := make([]events.InvoiceSummary, 0, len(results))
	for _, r := range results {
		if invoice, ok := r.(events.InvoiceSummary); ok {
			invoices = append(invoices, invoice)
		}
	}

	return invoices, nil
}

func (h *InvoiceQueryHandler) GetOverdueInvoices(ctx context.Context, query *GetOverdueInvoicesQuery) (*ListInvoicesResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_overdue_invoices",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	now := time.Now().UTC()

	filter := map[string]interface{}{
		"tenantId": query.TenantID,
		"status":   map[string]interface{}{"$in": []string{"pending", "sent", "partial"}},
		"dueDate":  map[string]interface{}{"$lt": now},
	}

	total, err := h.readModelStore.Count(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to count overdue invoices: %w", err)
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	skip := (query.Page - 1) * pageSize

	findOpts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(map[string]int{"dueDate": 1})

	results, err := h.readModelStore.Find(ctx, filter, findOpts)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}

	invoices := make([]events.InvoiceSummary, 0, len(results))
	for _, r := range results {
		if invoice, ok := r.(events.InvoiceSummary); ok {
			invoices = append(invoices, invoice)
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &ListInvoicesResult{
		Invoices:   invoices,
		Total:      total,
		Page:       query.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (h *InvoiceQueryHandler) GetInvoiceStats(ctx context.Context, query *GetInvoiceStatsQuery) (*InvoiceStats, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_invoice_stats",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("invoice:stats:%s", query.TenantID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var stats InvoiceStats
		if err := json.Unmarshal(cached, &stats); err == nil {
			return &stats, nil
		}
	}

	filter := map[string]interface{}{
		"tenantId": query.TenantID,
	}

	if !query.StartDate.IsZero() && !query.EndDate.IsZero() {
		filter["issueDate"] = map[string]interface{}{
			"$gte": query.StartDate,
			"$lte": query.EndDate,
		}
	}

	totalCount, err := h.readModelStore.Count(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get invoice stats: %w", err)
	}

	pendingFilter := map[string]interface{}{"tenantId": query.TenantID, "status": "pending"}
	pendingCount, _ := h.readModelStore.Count(ctx, pendingFilter)

	sentFilter := map[string]interface{}{"tenantId": query.TenantID, "status": "sent"}
	sentCount, _ := h.readModelStore.Count(ctx, sentFilter)

	paidFilter := map[string]interface{}{"tenantId": query.TenantID, "status": "paid"}
	paidCount, _ := h.readModelStore.Count(ctx, paidFilter)

	now := time.Now().UTC()
	overdueFilter := map[string]interface{}{
		"tenantId": query.TenantID,
		"status":   map[string]interface{}{"$in": []string{"pending", "sent", "partial"}},
		"dueDate":  map[string]interface{}{"$lt": now},
	}
	overdueCount, _ := h.readModelStore.Count(ctx, overdueFilter)

	cancelledFilter := map[string]interface{}{"tenantId": query.TenantID, "status": "cancelled"}
	cancelledCount, _ := h.readModelStore.Count(ctx, cancelledFilter)

	stats := &InvoiceStats{
		TenantID:         query.TenantID,
		TotalInvoices:    totalCount,
		PendingCount:     pendingCount,
		SentCount:        sentCount,
		PaidCount:        paidCount,
		OverdueCount:     overdueCount,
		CancelledCount:   cancelledCount,
		TotalAmount:      "0",
		TotalPaid:        "0",
		TotalOutstanding: "0",
		PeriodStart:      query.StartDate,
		PeriodEnd:        query.EndDate,
	}

	if data, err := json.Marshal(stats); err == nil {
		h.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return stats, nil
}
