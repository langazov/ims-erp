package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	eventpkg "github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/errors"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/shopspring/decimal"
)

type ClientCommandHandler struct {
	eventStore   *repository.EventStore
	publisher    eventpkg.Publisher
	logger       *logger.Logger
	tenantConfig TenantConfig
}

type TenantConfig struct {
	AutoGenerateCode   bool
	CodePrefix         string
	DefaultCreditLimit decimal.Decimal
	RequireEmail       bool
}

type Publisher interface {
	PublishEvent(ctx context.Context, event *eventpkg.EventEnvelope) error
}

func NewClientCommandHandler(
	eventStore *repository.EventStore,
	publisher Publisher,
	log *logger.Logger,
	tenantConfig TenantConfig,
) *ClientCommandHandler {
	return &ClientCommandHandler{
		eventStore:   eventStore,
		publisher:    publisher,
		logger:       log,
		tenantConfig: tenantConfig,
	}
}

type CreateClientCmd struct {
	Name              string
	Email             string
	Phone             string
	CreditLimit       decimal.Decimal
	BillingAddress    domain.Address
	ShippingAddresses []domain.Address
	Tags              []string
	CustomFields      map[string]interface{}
}

func (h *ClientCommandHandler) HandleCreateClient(ctx context.Context, cmd *CommandEnvelope) (*domain.Client, error) {
	data := cmd.Data
	name, _ := data["name"].(string)
	email, _ := data["email"].(string)

	if name == "" {
		return nil, errors.InvalidArgument("name is required")
	}
	if email == "" && h.tenantConfig.RequireEmail {
		return nil, errors.InvalidArgument("email is required")
	}

	tenantID, err := uuid.Parse(cmd.TenantID)
	if err != nil {
		return nil, errors.InvalidArgument("invalid tenant ID")
	}

	client := domain.NewClient(
		tenantID,
		name,
		email,
	)

	if phone, ok := data["phone"].(string); ok {
		client.Phone = phone
	}

	if creditLimit, ok := data["creditLimit"].(string); ok {
		if limit, err := decimal.NewFromString(creditLimit); err == nil {
			client.CreditLimit = limit
		}
	} else {
		client.CreditLimit = h.tenantConfig.DefaultCreditLimit
	}

	if billingAddrData, ok := data["billingAddress"].(map[string]interface{}); ok {
		client.BillingAddress = parseAddress(billingAddrData)
	}

	if shippingAddrs, ok := data["shippingAddresses"].([]interface{}); ok {
		for _, addr := range shippingAddrs {
			if addrMap, ok := addr.(map[string]interface{}); ok {
				client.ShippingAddresses = append(client.ShippingAddresses, parseAddress(addrMap))
			}
		}
	}

	if tags, ok := data["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				client.AddTag(tagStr)
			}
		}
	}

	if customFields, ok := data["customFields"].(map[string]interface{}); ok {
		for k, v := range customFields {
			client.SetCustomField(k, v)
		}
	}

	event := eventpkg.NewEvent(
		client.ID.String(),
		"Client",
		"ClientCreated",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"name":              client.Name,
			"email":             client.Email,
			"phone":             client.Phone,
			"creditLimit":       client.CreditLimit.String(),
			"billingAddress":    client.BillingAddress,
			"shippingAddresses": client.ShippingAddresses,
			"tags":              client.Tags,
		},
	).WithCorrelationID(cmd.CorrelationID)

	storedEvent := repository.StoredEvent{
		ID:            event.ID,
		AggregateID:   client.ID.String(),
		AggregateType: "Client",
		EventType:     "ClientCreated",
		EventData:     event.Data,
		Version:       1,
		Timestamp:     event.Timestamp,
		Metadata: repository.EventMetadata{
			TenantID:      cmd.TenantID,
			UserID:        cmd.UserID,
			CorrelationID: cmd.CorrelationID,
			CausationID:   cmd.ID,
			Timestamp:     event.Timestamp,
		},
	}

	if err := h.eventStore.Save(ctx, []repository.StoredEvent{storedEvent}); err != nil {
		return nil, errors.Wrap(err, errors.CodeInternalError, "failed to save event")
	}

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.Error("Failed to publish ClientCreated event", "error", err)
	}

	return client, nil
}

func (h *ClientCommandHandler) HandleUpdateClient(ctx context.Context, cmd *CommandEnvelope) (*domain.Client, error) {
	data := cmd.Data
	clientID, _ := data["clientId"].(string)

	if clientID == "" {
		return nil, errors.InvalidArgument("clientId is required")
	}

	events, err := h.eventStore.Load(ctx, clientID)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternalError, "failed to load events")
	}

	if len(events) == 0 {
		return nil, errors.NotFound("client not found: %s", clientID)
	}

	client := &domain.Client{}
	for _, e := range events {
		client.ID = uuid.Must(uuid.Parse(e.AggregateID))
		client.TenantID = uuid.Must(uuid.Parse(e.Metadata.TenantID))
		client.Version = e.Version

		switch e.EventType {
		case "ClientCreated":
			client.Name = getString(e.EventData, "name")
			client.Email = getString(e.EventData, "email")
			client.Phone = getString(e.EventData, "phone")
			client.CreditLimit = getDecimal(e.EventData, "creditLimit")
			client.CreatedAt = e.Timestamp
			client.Status = domain.ClientStatusActive
		case "ClientUpdated":
			if name, ok := e.EventData["name"].(string); ok {
				client.Name = name
			}
			if email, ok := e.EventData["email"].(string); ok {
				client.Email = email
			}
			if phone, ok := e.EventData["phone"].(string); ok {
				client.Phone = phone
			}
			client.UpdatedAt = e.Timestamp
		}
	}

	if cmd.ExpectedVersion > 0 && client.Version != cmd.ExpectedVersion {
		return nil, errors.Conflict("client version mismatch: expected %d, got %d", cmd.ExpectedVersion, client.Version)
	}

	if name, ok := data["name"].(string); ok {
		client.Name = name
	}
	if email, ok := data["email"].(string); ok {
		client.Email = email
	}
	if phone, ok := data["phone"].(string); ok {
		client.Phone = phone
	}

	client.Version++
	client.UpdatedAt = events[0].Timestamp

	event := eventpkg.NewEvent(
		clientID,
		"Client",
		"ClientUpdated",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"name":    client.Name,
			"email":   client.Email,
			"phone":   client.Phone,
			"changes": data,
		},
	).WithCorrelationID(cmd.CorrelationID)

	storedEvent := repository.StoredEvent{
		ID:            event.ID,
		AggregateID:   clientID,
		AggregateType: "Client",
		EventType:     "ClientUpdated",
		EventData:     event.Data,
		Version:       client.Version,
		Timestamp:     event.Timestamp,
		Metadata: repository.EventMetadata{
			TenantID:      cmd.TenantID,
			UserID:        cmd.UserID,
			CorrelationID: cmd.CorrelationID,
			CausationID:   cmd.ID,
			Timestamp:     event.Timestamp,
		},
	}

	if err := h.eventStore.Save(ctx, []repository.StoredEvent{storedEvent}); err != nil {
		return nil, errors.Wrap(err, errors.CodeInternalError, "failed to save event")
	}

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.Error("Failed to publish ClientUpdated event", "error", err)
	}

	return client, nil
}

func (h *ClientCommandHandler) HandleDeactivateClient(ctx context.Context, cmd *CommandEnvelope) error {
	data := cmd.Data
	clientID, _ := data["clientId"].(string)

	if clientID == "" {
		return errors.InvalidArgument("clientId is required")
	}

	reason, _ := data["reason"].(string)

	events, err := h.eventStore.Load(ctx, clientID)
	if err != nil {
		return errors.Wrap(err, errors.CodeInternalError, "failed to load events")
	}

	if len(events) == 0 {
		return errors.NotFound("client not found: %s", clientID)
	}

	var currentVersion int64
	for _, e := range events {
		currentVersion = e.Version
	}

	event := eventpkg.NewEvent(
		clientID,
		"Client",
		"ClientDeactivated",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"reason": reason,
		},
	).WithCorrelationID(cmd.CorrelationID)

	storedEvent := repository.StoredEvent{
		ID:            event.ID,
		AggregateID:   clientID,
		AggregateType: "Client",
		EventType:     "ClientDeactivated",
		EventData:     event.Data,
		Version:       currentVersion + 1,
		Timestamp:     event.Timestamp,
		Metadata: repository.EventMetadata{
			TenantID:      cmd.TenantID,
			UserID:        cmd.UserID,
			CorrelationID: cmd.CorrelationID,
			CausationID:   cmd.ID,
			Timestamp:     event.Timestamp,
		},
	}

	if err := h.eventStore.Save(ctx, []repository.StoredEvent{storedEvent}); err != nil {
		return errors.Wrap(err, errors.CodeInternalError, "failed to save event")
	}

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.Error("Failed to publish ClientDeactivated event", "error", err)
	}

	return nil
}

func (h *ClientCommandHandler) HandleAssignCreditLimit(ctx context.Context, cmd *CommandEnvelope) error {
	data := cmd.Data
	clientID, _ := data["clientId"].(string)

	if clientID == "" {
		return errors.InvalidArgument("clientId is required")
	}

	creditLimitStr, _ := data["creditLimit"].(string)
	creditLimit, err := decimal.NewFromString(creditLimitStr)
	if err != nil {
		return errors.InvalidArgument("invalid creditLimit format")
	}

	reason, _ := data["reason"].(string)

	events, err := h.eventStore.Load(ctx, clientID)
	if err != nil {
		return errors.Wrap(err, errors.CodeInternalError, "failed to load events")
	}

	if len(events) == 0 {
		return errors.NotFound("client not found: %s", clientID)
	}

	var oldLimit decimal.Decimal
	for _, e := range events {
		if e.EventType == "ClientCreated" {
			oldLimit = getDecimal(e.EventData, "creditLimit")
		}
		if e.EventType == "CreditLimitAssigned" {
			oldLimit = getDecimal(e.EventData, "newLimit")
		}
	}

	event := eventpkg.NewEvent(
		clientID,
		"Client",
		"CreditLimitAssigned",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"oldLimit": oldLimit.String(),
			"newLimit": creditLimit.String(),
			"reason":   reason,
		},
	).WithCorrelationID(cmd.CorrelationID)

	storedEvent := repository.StoredEvent{
		ID:            event.ID,
		AggregateID:   clientID,
		AggregateType: "Client",
		EventType:     "CreditLimitAssigned",
		EventData:     event.Data,
		Version:       int64(len(events) + 1),
		Timestamp:     event.Timestamp,
		Metadata: repository.EventMetadata{
			TenantID:      cmd.TenantID,
			UserID:        cmd.UserID,
			CorrelationID: cmd.CorrelationID,
			CausationID:   cmd.ID,
			Timestamp:     event.Timestamp,
		},
	}

	if err := h.eventStore.Save(ctx, []repository.StoredEvent{storedEvent}); err != nil {
		return errors.Wrap(err, errors.CodeInternalError, "failed to save event")
	}

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.Error("Failed to publish CreditLimitAssigned event", "error", err)
	}

	return nil
}

func (h *ClientCommandHandler) HandleUpdateBillingInfo(ctx context.Context, cmd *CommandEnvelope) error {
	data := cmd.Data
	clientID, _ := data["clientId"].(string)

	if clientID == "" {
		return errors.InvalidArgument("clientId is required")
	}

	addrData, ok := data["billingAddress"].(map[string]interface{})
	if !ok {
		return errors.InvalidArgument("billingAddress is required")
	}

	addr := parseAddress(addrData)

	event := eventpkg.NewEvent(
		clientID,
		"Client",
		"BillingInfoUpdated",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"billingAddress": addr,
		},
	).WithCorrelationID(cmd.CorrelationID)

	events, err := h.eventStore.Load(ctx, clientID)
	if err != nil {
		return errors.Wrap(err, errors.CodeInternalError, "failed to load events")
	}

	if len(events) == 0 {
		return errors.NotFound("client not found: %s", clientID)
	}

	storedEvent := repository.StoredEvent{
		ID:            event.ID,
		AggregateID:   clientID,
		AggregateType: "Client",
		EventType:     "BillingInfoUpdated",
		EventData:     event.Data,
		Version:       int64(len(events) + 1),
		Timestamp:     event.Timestamp,
		Metadata: repository.EventMetadata{
			TenantID:      cmd.TenantID,
			UserID:        cmd.UserID,
			CorrelationID: cmd.CorrelationID,
			CausationID:   cmd.ID,
			Timestamp:     event.Timestamp,
		},
	}

	if err := h.eventStore.Save(ctx, []repository.StoredEvent{storedEvent}); err != nil {
		return errors.Wrap(err, errors.CodeInternalError, "failed to save event")
	}

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.Error("Failed to publish BillingInfoUpdated event", "error", err)
	}

	return nil
}

func (h *ClientCommandHandler) HandleMergeClients(ctx context.Context, cmd *CommandEnvelope) error {
	data := cmd.Data
	sourceID, _ := data["sourceClientId"].(string)
	targetID, _ := data["targetClientId"].(string)

	if sourceID == "" || targetID == "" {
		return errors.InvalidArgument("sourceClientId and targetClientId are required")
	}

	if sourceID == targetID {
		return errors.InvalidArgument("source and target client must be different")
	}

	event := eventpkg.NewEvent(
		targetID,
		"Client",
		"ClientsMerged",
		cmd.TenantID,
		cmd.UserID,
		map[string]interface{}{
			"sourceClientId": sourceID,
			"targetClientId": targetID,
		},
	).WithCorrelationID(cmd.CorrelationID)

	targetEvents, err := h.eventStore.Load(ctx, targetID)
	if err != nil {
		return errors.Wrap(err, errors.CodeInternalError, "failed to load target events")
	}

	if len(targetEvents) == 0 {
		return errors.NotFound("target client not found: %s", targetID)
	}

	storedEvent := repository.StoredEvent{
		ID:            event.ID,
		AggregateID:   targetID,
		AggregateType: "Client",
		EventType:     "ClientsMerged",
		EventData:     event.Data,
		Version:       int64(len(targetEvents) + 1),
		Timestamp:     event.Timestamp,
		Metadata: repository.EventMetadata{
			TenantID:      cmd.TenantID,
			UserID:        cmd.UserID,
			CorrelationID: cmd.CorrelationID,
			CausationID:   cmd.ID,
			Timestamp:     event.Timestamp,
		},
	}

	if err := h.eventStore.Save(ctx, []repository.StoredEvent{storedEvent}); err != nil {
		return errors.Wrap(err, errors.CodeInternalError, "failed to save event")
	}

	if err := h.publisher.PublishEvent(ctx, event); err != nil {
		h.logger.Error("Failed to publish ClientsMerged event", "error", err)
	}

	return nil
}

func parseAddress(data map[string]interface{}) domain.Address {
	return domain.Address{
		Street:     getString(data, "street"),
		City:       getString(data, "city"),
		State:      getString(data, "state"),
		PostalCode: getString(data, "postalCode"),
		Country:    getString(data, "country"),
	}
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

func getDecimal(data map[string]interface{}, key string) decimal.Decimal {
	if v, ok := data[key].(string); ok {
		if d, err := decimal.NewFromString(v); err == nil {
			return d
		}
	}
	return decimal.Zero
}
