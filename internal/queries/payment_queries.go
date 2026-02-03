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

type PaymentQueryHandler struct {
	readModelStore *repository.ReadModelStore
	cache          *repository.Cache
	logger         *logger.Logger
	tracer         trace.Tracer
}

func NewPaymentQueryHandler(
	readModelStore *repository.ReadModelStore,
	cache *repository.Cache,
	log *logger.Logger,
) *PaymentQueryHandler {
	return &PaymentQueryHandler{
		readModelStore: readModelStore,
		cache:          cache,
		logger:         log,
		tracer:         otel.Tracer("payment-query-handler"),
	}
}

type GetPaymentByIDQuery struct {
	PaymentID string
	TenantID  string
}

type ListPaymentsQuery struct {
	TenantID  string
	ClientID  string
	InvoiceID string
	Page      int
	PageSize  int
	Status    string
	Method    string
	StartDate time.Time
	EndDate   time.Time
	SortBy    string
	SortOrder string
}

type GetPaymentsByInvoiceQuery struct {
	InvoiceID string
	TenantID  string
	Page      int
	PageSize  int
}

type GetPaymentStatsQuery struct {
	TenantID  string
	StartDate time.Time
	EndDate   time.Time
}

type ListPaymentsResult struct {
	Payments   []events.PaymentSummary `json:"payments"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"pageSize"`
	TotalPages int                     `json:"totalPages"`
}

type PaymentStats struct {
	TenantID       string    `json:"tenantId"`
	TotalPayments  int64     `json:"totalPayments"`
	PendingCount   int64     `json:"pendingCount"`
	CompletedCount int64     `json:"completedCount"`
	FailedCount    int64     `json:"failedCount"`
	RefundedCount  int64     `json:"refundedCount"`
	CancelledCount int64     `json:"cancelledCount"`
	TotalAmount    string    `json:"totalAmount"`
	TotalRefunded  string    `json:"totalRefunded"`
	AverageAmount  string    `json:"averageAmount"`
	PeriodStart    time.Time `json:"periodStart,omitempty"`
	PeriodEnd      time.Time `json:"periodEnd,omitempty"`
}

func (h *PaymentQueryHandler) GetPaymentByID(ctx context.Context, query *GetPaymentByIDQuery) (*events.PaymentSummary, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_payment_by_id",
		trace.WithAttributes(
			attribute.String("payment_id", query.PaymentID),
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("payment:summary:%s", query.PaymentID)
	if cached, err := h.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		return nil, nil
	}

	filter := map[string]interface{}{
		"_id":      query.PaymentID,
		"tenantId": query.TenantID,
	}

	result, err := h.readModelStore.FindOne(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if result == nil {
		return nil, nil
	}

	payment, ok := result.(events.PaymentSummary)
	if !ok {
		return nil, fmt.Errorf("invalid payment data")
	}

	h.cache.Set(ctx, cacheKey, payment, 5*time.Minute)

	return &payment, nil
}

func (h *PaymentQueryHandler) ListPayments(ctx context.Context, query *ListPaymentsQuery) (*ListPaymentsResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.list_payments",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
			attribute.Int("page", query.Page),
			attribute.Int("page_size", query.PageSize),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("payment:list:%s:%s:%s:%d:%d:%s:%s",
		query.TenantID, query.ClientID, query.InvoiceID, query.Page, query.PageSize, query.Status, query.Method)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var result ListPaymentsResult
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

	if query.InvoiceID != "" {
		filter["invoiceId"] = query.InvoiceID
	}

	if query.Status != "" {
		filter["status"] = query.Status
	}

	if query.Method != "" {
		filter["method"] = query.Method
	}

	if !query.StartDate.IsZero() && !query.EndDate.IsZero() {
		filter["createdAt"] = map[string]interface{}{
			"$gte": query.StartDate,
			"$lte": query.EndDate,
		}
	}

	total, err := h.readModelStore.Count(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to count payments: %w", err)
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	skip := (query.Page - 1) * pageSize

	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "createdAt"
	}

	findOpts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(map[string]int{sortBy: getSortOrder(query.SortOrder)})

	results, err := h.readModelStore.Find(ctx, filter, findOpts)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	payments := make([]events.PaymentSummary, 0, len(results))
	for _, r := range results {
		if payment, ok := r.(events.PaymentSummary); ok {
			payments = append(payments, payment)
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	result := &ListPaymentsResult{
		Payments:   payments,
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

func (h *PaymentQueryHandler) GetPaymentsByInvoice(ctx context.Context, query *GetPaymentsByInvoiceQuery) (*ListPaymentsResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_payments_by_invoice",
		trace.WithAttributes(
			attribute.String("invoice_id", query.InvoiceID),
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"tenantId":  query.TenantID,
		"invoiceId": query.InvoiceID,
	}

	total, err := h.readModelStore.Count(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to count payments by invoice: %w", err)
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	skip := (query.Page - 1) * pageSize

	findOpts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(map[string]int{"createdAt": -1})

	results, err := h.readModelStore.Find(ctx, filter, findOpts)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get payments by invoice: %w", err)
	}

	payments := make([]events.PaymentSummary, 0, len(results))
	for _, r := range results {
		if payment, ok := r.(events.PaymentSummary); ok {
			payments = append(payments, payment)
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &ListPaymentsResult{
		Payments:   payments,
		Total:      total,
		Page:       query.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (h *PaymentQueryHandler) GetPaymentStats(ctx context.Context, query *GetPaymentStatsQuery) (*PaymentStats, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_payment_stats",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("payment:stats:%s", query.TenantID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var stats PaymentStats
		if err := json.Unmarshal(cached, &stats); err == nil {
			return &stats, nil
		}
	}

	filter := map[string]interface{}{
		"tenantId": query.TenantID,
	}

	if !query.StartDate.IsZero() && !query.EndDate.IsZero() {
		filter["createdAt"] = map[string]interface{}{
			"$gte": query.StartDate,
			"$lte": query.EndDate,
		}
	}

	totalCount, err := h.readModelStore.Count(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get payment stats: %w", err)
	}

	pendingFilter := map[string]interface{}{"tenantId": query.TenantID, "status": "pending"}
	pendingCount, _ := h.readModelStore.Count(ctx, pendingFilter)

	completedFilter := map[string]interface{}{"tenantId": query.TenantID, "status": "completed"}
	completedCount, _ := h.readModelStore.Count(ctx, completedFilter)

	failedFilter := map[string]interface{}{"tenantId": query.TenantID, "status": "failed"}
	failedCount, _ := h.readModelStore.Count(ctx, failedFilter)

	refundedFilter := map[string]interface{}{"tenantId": query.TenantID, "status": "refunded"}
	refundedCount, _ := h.readModelStore.Count(ctx, refundedFilter)

	cancelledFilter := map[string]interface{}{"tenantId": query.TenantID, "status": "cancelled"}
	cancelledCount, _ := h.readModelStore.Count(ctx, cancelledFilter)

	stats := &PaymentStats{
		TenantID:       query.TenantID,
		TotalPayments:  totalCount,
		PendingCount:   pendingCount,
		CompletedCount: completedCount,
		FailedCount:    failedCount,
		RefundedCount:  refundedCount,
		CancelledCount: cancelledCount,
		TotalAmount:    "0",
		TotalRefunded:  "0",
		AverageAmount:  "0",
		PeriodStart:    query.StartDate,
		PeriodEnd:      query.EndDate,
	}

	if data, err := json.Marshal(stats); err == nil {
		h.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return stats, nil
}
