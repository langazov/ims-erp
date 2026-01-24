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

type ClientQueryHandler struct {
	readModelStore *repository.ReadModelStore
	cache          *repository.Cache
	logger         *logger.Logger
	tracer         trace.Tracer
}

func NewClientQueryHandler(
	readModelStore *repository.ReadModelStore,
	cache *repository.Cache,
	log *logger.Logger,
) *ClientQueryHandler {
	return &ClientQueryHandler{
		readModelStore: readModelStore,
		cache:          cache,
		logger:         log,
		tracer:         otel.Tracer("client-query-handler"),
	}
}

type GetClientByIDQuery struct {
	ClientID string
	TenantID string
}

type GetClientDetailQuery struct {
	ClientID string
	TenantID string
}

type ListClientsQuery struct {
	TenantID  string
	Page      int
	PageSize  int
	Search    string
	Status    string
	Tags      []string
	SortBy    string
	SortOrder string
}

type SearchClientsQuery struct {
	TenantID string
	Term     string
	Limit    int
}

type GetClientCreditStatusQuery struct {
	ClientID string
	TenantID string
}

type ListClientsResult struct {
	Clients    []events.ClientSummary `json:"clients"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"pageSize"`
	TotalPages int                    `json:"totalPages"`
}

func (h *ClientQueryHandler) GetClientByID(ctx context.Context, query *GetClientByIDQuery) (*events.ClientSummary, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_client_by_id",
		trace.WithAttributes(
			attribute.String("client_id", query.ClientID),
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("client:summary:%s", query.ClientID)
	if cached, err := h.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		return nil, nil
	}

	filter := map[string]interface{}{
		"_id":      query.ClientID,
		"tenantId": query.TenantID,
	}

	result, err := h.readModelStore.FindOne(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	if result == nil {
		return nil, nil
	}

	client, ok := result.(events.ClientSummary)
	if !ok {
		return nil, fmt.Errorf("invalid client data")
	}

	h.cache.Set(ctx, cacheKey, client, 5*time.Minute)

	return &client, nil
}

func (h *ClientQueryHandler) GetClientDetail(ctx context.Context, query *GetClientDetailQuery) (*events.ClientDetail, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_client_detail",
		trace.WithAttributes(
			attribute.String("client_id", query.ClientID),
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("client:detail:%s", query.ClientID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var client events.ClientDetail
		if err := json.Unmarshal(cached, &client); err == nil {
			return &client, nil
		}
	}

	filter := map[string]interface{}{
		"_id":      query.ClientID,
		"tenantId": query.TenantID,
	}

	result, err := h.readModelStore.FindOne(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get client detail: %w", err)
	}

	if result == nil {
		return nil, nil
	}

	clientDetail, ok := result.(events.ClientDetail)
	if !ok {
		return nil, fmt.Errorf("invalid client detail data")
	}

	if data, err := json.Marshal(clientDetail); err == nil {
		h.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return &clientDetail, nil
}

func (h *ClientQueryHandler) ListClients(ctx context.Context, query *ListClientsQuery) (*ListClientsResult, error) {
	ctx, span := h.tracer.Start(ctx, "query.list_clients",
		trace.WithAttributes(
			attribute.String("tenant_id", query.TenantID),
			attribute.Int("page", query.Page),
			attribute.Int("page_size", query.PageSize),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("client:list:%s:%d:%d:%s:%s:%v",
		query.TenantID, query.Page, query.PageSize, query.Search, query.Status, query.Tags)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var result ListClientsResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	filter := map[string]interface{}{
		"tenantId": query.TenantID,
	}

	if query.Search != "" {
		filter["$or"] = []map[string]interface{}{
			{"name": map[string]interface{}{"$regex": query.Search, "$options": "i"}},
			{"email": map[string]interface{}{"$regex": query.Search, "$options": "i"}},
		}
	}

	if query.Status != "" {
		filter["status"] = query.Status
	}

	if len(query.Tags) > 0 {
		filter["tags"] = map[string]interface{}{
			"$in": query.Tags,
		}
	}

	total, err := h.readModelStore.Count(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to count clients: %w", err)
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	skip := (query.Page - 1) * pageSize

	findOpts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(map[string]int{query.SortBy: getSortOrder(query.SortOrder)})

	results, err := h.readModelStore.Find(ctx, filter, findOpts)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}

	clients := make([]events.ClientSummary, 0, len(results))
	for _, r := range results {
		if client, ok := r.(events.ClientSummary); ok {
			clients = append(clients, client)
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	result := &ListClientsResult{
		Clients:    clients,
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

func (h *ClientQueryHandler) SearchClients(ctx context.Context, query *SearchClientsQuery) ([]events.ClientSummary, error) {
	ctx, span := h.tracer.Start(ctx, "query.search_clients",
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
			{"name": map[string]interface{}{"$regex": query.Term, "$options": "i"}},
			{"email": map[string]interface{}{"$regex": query.Term, "$options": "i"}},
			{"phone": map[string]interface{}{"$regex": query.Term, "$options": "i"}},
		},
	}

	findOpts := options.Find().
		SetLimit(int64(limit)).
		SetSort(map[string]int{"name": 1})

	results, err := h.readModelStore.Find(ctx, filter, findOpts)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to search clients: %w", err)
	}

	clients := make([]events.ClientSummary, 0, len(results))
	for _, r := range results {
		if client, ok := r.(events.ClientSummary); ok {
			clients = append(clients, client)
		}
	}

	return clients, nil
}

func (h *ClientQueryHandler) GetClientCreditStatus(ctx context.Context, query *GetClientCreditStatusQuery) (*events.ClientCreditStatus, error) {
	ctx, span := h.tracer.Start(ctx, "query.get_client_credit_status",
		trace.WithAttributes(
			attribute.String("client_id", query.ClientID),
			attribute.String("tenant_id", query.TenantID),
		),
	)
	defer span.End()

	cacheKey := fmt.Sprintf("client:credit:%s", query.ClientID)
	if cached, err := h.cache.GetBytes(ctx, cacheKey); err == nil && cached != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		var status events.ClientCreditStatus
		if err := json.Unmarshal(cached, &status); err == nil {
			return &status, nil
		}
	}

	client, err := h.GetClientByID(ctx, &GetClientByIDQuery{
		ClientID: query.ClientID,
		TenantID: query.TenantID,
	})
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, nil
	}

	creditStatus := events.ClientCreditStatus{
		ID:              query.ClientID,
		TenantID:        query.TenantID,
		ClientID:        query.ClientID,
		CreditLimit:     client.CreditLimit,
		CurrentBalance:  client.CurrentBalance,
		AvailableCredit: "0",
		Utilization:     0,
		RiskLevel:       "low",
		LastCheck:       time.Now().UTC(),
	}

	if data, err := json.Marshal(creditStatus); err == nil {
		h.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return &creditStatus, nil
}

func getSortOrder(order string) int {
	if order == "desc" {
		return -1
	}
	return 1
}
