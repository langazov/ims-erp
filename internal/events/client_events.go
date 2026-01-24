package events

import (
	"context"
	"time"

	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ClientEventHandler struct {
	readModelStore *repository.ReadModelStore
	cache          *repository.Cache
	logger         *logger.Logger
	tracer         trace.Tracer
}

func NewClientEventHandler(
	readModelStore *repository.ReadModelStore,
	cache *repository.Cache,
	log *logger.Logger,
) *ClientEventHandler {
	return &ClientEventHandler{
		readModelStore: readModelStore,
		cache:          cache,
		logger:         log,
		tracer:         otel.Tracer("client-event-handler"),
	}
}

func (h *ClientEventHandler) HandleClientCreated(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_client_created",
		trace.WithAttributes(
			attribute.String("client_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	clientSummary := ClientSummary{
		ID:             event.AggregateID,
		TenantID:       event.TenantID,
		Name:           getString(event.Data, "name"),
		Email:          getString(event.Data, "email"),
		Phone:          getString(event.Data, "phone"),
		Status:         string(domain.ClientStatusActive),
		CreditLimit:    getDecimal(event.Data, "creditLimit"),
		CurrentBalance: "0",
		Tags:           getStringSlice(event.Data, "tags"),
		CreatedAt:      event.Timestamp,
		UpdatedAt:      event.Timestamp,
	}

	if err := h.readModelStore.Save(ctx, clientSummary); err != nil {
		span.RecordError(err)
		return err
	}

	clientDetail := ClientDetail{
		ID:                event.AggregateID,
		TenantID:          event.TenantID,
		Name:              getString(event.Data, "name"),
		Email:             getString(event.Data, "email"),
		Phone:             getString(event.Data, "phone"),
		Status:            string(domain.ClientStatusActive),
		CreditLimit:       getDecimal(event.Data, "creditLimit"),
		CurrentBalance:    "0",
		BillingAddress:    getAddress(event.Data, "billingAddress"),
		ShippingAddresses: getAddressSlice(event.Data, "shippingAddresses"),
		Tags:              getStringSlice(event.Data, "tags"),
		CustomFields:      getMap(event.Data, "customFields"),
		ActivityLog: []ClientActivity{
			{
				Action:    "created",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
			},
		},
		CreatedAt: event.Timestamp,
		UpdatedAt: event.Timestamp,
	}

	if err := h.readModelStore.Save(ctx, clientDetail); err != nil {
		span.RecordError(err)
		return err
	}

	h.logger.New(ctx).Info("Client created",
		"client_id", event.AggregateID,
		"tenant_id", event.TenantID,
		"name", clientSummary.Name,
	)

	return nil
}

func (h *ClientEventHandler) HandleClientUpdated(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_client_updated",
		trace.WithAttributes(
			attribute.String("client_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"name":      getString(event.Data, "name"),
			"email":     getString(event.Data, "email"),
			"phone":     getString(event.Data, "phone"),
			"updatedAt": event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": ClientActivity{
				Action:    "updated",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
				Details:   "Client information updated",
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "client:detail:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "client:list:*")

	h.logger.New(ctx).Debug("Client updated",
		"client_id", event.AggregateID,
		"tenant_id", event.TenantID,
	)

	return nil
}

func (h *ClientEventHandler) HandleClientDeactivated(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_client_deactivated",
		trace.WithAttributes(
			attribute.String("client_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	reason := getString(event.Data, "reason")

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":    string(domain.ClientStatusInactive),
			"updatedAt": event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": ClientActivity{
				Action:    "deactivated",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
				Details:   reason,
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "client:detail:"+event.AggregateID)
	h.cache.DeletePattern(ctx, "client:list:*")

	h.logger.New(ctx).Info("Client deactivated",
		"client_id", event.AggregateID,
		"tenant_id", event.TenantID,
		"reason", reason,
	)

	return nil
}

func (h *ClientEventHandler) HandleCreditLimitAssigned(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_credit_limit_assigned",
		trace.WithAttributes(
			attribute.String("client_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"creditLimit": getString(event.Data, "newLimit"),
			"updatedAt":   event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": ClientActivity{
				Action:    "credit_limit_changed",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
				Details:   "Credit limit changed from " + getString(event.Data, "oldLimit") + " to " + getString(event.Data, "newLimit"),
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "client:detail:"+event.AggregateID)
	h.cache.Delete(ctx, "client:credit:"+event.AggregateID)

	h.logger.New(ctx).Debug("Credit limit assigned",
		"client_id", event.AggregateID,
		"old_limit", getString(event.Data, "oldLimit"),
		"new_limit", getString(event.Data, "newLimit"),
	)

	return nil
}

func (h *ClientEventHandler) HandleBillingInfoUpdated(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_billing_info_updated",
		trace.WithAttributes(
			attribute.String("client_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"_id":      event.AggregateID,
		"tenantId": event.TenantID,
	}

	addr := getAddress(event.Data, "billingAddress")

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"billingAddress": addr,
			"updatedAt":      event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": ClientActivity{
				Action:    "billing_info_updated",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "client:detail:"+event.AggregateID)

	return nil
}

func (h *ClientEventHandler) HandleClientsMerged(ctx context.Context, event *EventEnvelope) error {
	ctx, span := h.tracer.Start(ctx, "handle_clients_merged",
		trace.WithAttributes(
			attribute.String("client_id", event.AggregateID),
			attribute.String("tenant_id", event.TenantID),
			attribute.String("source_id", getString(event.Data, "sourceClientId")),
		),
	)
	defer span.End()

	sourceID := getString(event.Data, "sourceClientId")
	targetID := getString(event.Data, "targetClientId")

	filter := map[string]interface{}{
		"_id":      targetID,
		"tenantId": event.TenantID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"updatedAt": event.Timestamp,
		},
		"$push": map[string]interface{}{
			"activityLog": ClientActivity{
				Action:    "merged",
				Timestamp: event.Timestamp,
				UserID:    event.UserID,
				Details:   "Merged from client " + sourceID,
			},
		},
	}

	if err := h.readModelStore.Update(ctx, filter, update); err != nil {
		span.RecordError(err)
		return err
	}

	h.cache.Delete(ctx, "client:detail:"+targetID)
	h.cache.Delete(ctx, "client:detail:"+sourceID)
	h.cache.DeletePattern(ctx, "client:list:*")

	h.logger.New(ctx).Info("Clients merged",
		"source_id", sourceID,
		"target_id", targetID,
	)

	return nil
}

type ClientSummary struct {
	ID             string    `bson:"_id" json:"id"`
	TenantID       string    `bson:"tenantId" json:"tenantId"`
	Name           string    `bson:"name" json:"name"`
	Email          string    `bson:"email" json:"email"`
	Phone          string    `bson:"phone" json:"phone"`
	Status         string    `bson:"status" json:"status"`
	CreditLimit    string    `bson:"creditLimit" json:"creditLimit"`
	CurrentBalance string    `bson:"currentBalance" json:"currentBalance"`
	Tags           []string  `bson:"tags" json:"tags"`
	CreatedAt      time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time `bson:"updatedAt" json:"updatedAt"`
}

type ClientDetail struct {
	ID                string                 `bson:"_id" json:"id"`
	TenantID          string                 `bson:"tenantId" json:"tenantId"`
	Name              string                 `bson:"name" json:"name"`
	Email             string                 `bson:"email" json:"email"`
	Phone             string                 `bson:"phone" json:"phone"`
	Status            string                 `bson:"status" json:"status"`
	CreditLimit       string                 `bson:"creditLimit" json:"creditLimit"`
	CurrentBalance    string                 `bson:"currentBalance" json:"currentBalance"`
	BillingAddress    domain.Address         `bson:"billingAddress" json:"billingAddress"`
	ShippingAddresses []domain.Address       `bson:"shippingAddresses" json:"shippingAddresses"`
	Tags              []string               `bson:"tags" json:"tags"`
	CustomFields      map[string]interface{} `bson:"customFields" json:"customFields"`
	ActivityLog       []ClientActivity       `bson:"activityLog" json:"activityLog"`
	CreatedAt         time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time              `bson:"updatedAt" json:"updatedAt"`
}

type ClientActivity struct {
	Action    string    `bson:"action" json:"action"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	UserID    string    `bson:"userId" json:"userId"`
	Details   string    `bson:"details,omitempty" json:"details,omitempty"`
}

type ClientCreditStatus struct {
	ID              string    `bson:"_id" json:"id"`
	TenantID        string    `bson:"tenantId" json:"tenantId"`
	ClientID        string    `bson:"clientId" json:"clientId"`
	CreditLimit     string    `bson:"creditLimit" json:"creditLimit"`
	CurrentBalance  string    `bson:"currentBalance" json:"currentBalance"`
	AvailableCredit string    `bson:"availableCredit" json:"availableCredit"`
	Utilization     float64   `bson:"utilization" json:"utilization"`
	RiskLevel       string    `bson:"riskLevel" json:"riskLevel"`
	LastCheck       time.Time `bson:"lastCheck" json:"lastCheck"`
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

func getDecimal(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return "0"
}

func getStringSlice(data map[string]interface{}, key string) []string {
	if v, ok := data[key].([]interface{}); ok {
		result := make([]string, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}

func getAddress(data map[string]interface{}, key string) domain.Address {
	if v, ok := data[key].(map[string]interface{}); ok {
		return domain.Address{
			Street:     getString(v, "street"),
			City:       getString(v, "city"),
			State:      getString(v, "state"),
			PostalCode: getString(v, "postalCode"),
			Country:    getString(v, "country"),
		}
	}
	return domain.Address{}
}

func getAddressSlice(data map[string]interface{}, key string) []domain.Address {
	if v, ok := data[key].([]interface{}); ok {
		result := make([]domain.Address, len(v))
		for i, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				result[i] = getAddress(map[string]interface{}{"address": m}, "address")
			}
		}
		return result
	}
	return nil
}

func getMap(data map[string]interface{}, key string) map[string]interface{} {
	if v, ok := data[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}
