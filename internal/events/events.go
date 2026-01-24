package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Publisher interface {
	PublishEvent(ctx context.Context, event *EventEnvelope) error
}

type EventEnvelope struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	AggregateID   string                 `json:"aggregateId"`
	AggregateType string                 `json:"aggregateType"`
	TenantID      string                 `json:"tenantId"`
	Version       int64                  `json:"version"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlationId"`
	CausationID   string                 `json:"causationId"`
	UserID        string                 `json:"userId"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]string      `json:"metadata"`
}

func NewEvent(aggregateID, aggregateType, eventType, tenantID, userID string, data map[string]interface{}) *EventEnvelope {
	return &EventEnvelope{
		ID:            uuid.New().String(),
		Type:          eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		TenantID:      tenantID,
		Version:       1,
		Timestamp:     time.Now().UTC(),
		CorrelationID: uuid.New().String(),
		UserID:        userID,
		Data:          data,
		Metadata:      make(map[string]string),
	}
}

func (e *EventEnvelope) WithCorrelationID(correlationID string) *EventEnvelope {
	e.CorrelationID = correlationID
	return e
}

func (e *EventEnvelope) WithCausationID(causationID string) *EventEnvelope {
	e.CausationID = causationID
	return e
}

func (e *EventEnvelope) WithMetadata(key, value string) *EventEnvelope {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

func (e *EventEnvelope) IncrementVersion() {
	e.Version++
}

func (e *EventEnvelope) Subject() string {
	return "evt." + e.AggregateType + "." + e.Type
}

func (e *EventEnvelope) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func EventFromJSON(data []byte) (*EventEnvelope, error) {
	var event EventEnvelope
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

type EventHandler func(ctx context.Context, event *EventEnvelope) error

type EventHandlerRegistry struct {
	handlers map[string][]EventHandler
}

func NewEventHandlerRegistry() *EventHandlerRegistry {
	return &EventHandlerRegistry{
		handlers: make(map[string][]EventHandler),
	}
}

func (r *EventHandlerRegistry) Register(eventType string, handler EventHandler) {
	r.handlers[eventType] = append(r.handlers[eventType], handler)
}

func (r *EventHandlerRegistry) GetHandlers(eventType string) []EventHandler {
	return r.handlers[eventType]
}

func (r *EventHandlerRegistry) Handle(ctx context.Context, event *EventEnvelope) []error {
	errors := make([]error, 0)
	handlers := r.GetHandlers(event.Type)
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

type Event interface {
	EventType() string
	AggregateID() string
	TenantID() string
	UserID() string
	Timestamp() time.Time
	Data() map[string]interface{}
}

type BaseEvent struct {
	eventType   string
	aggregateID string
	tenantID    string
	userID      string
	timestamp   time.Time
	data        map[string]interface{}
}

func (e *BaseEvent) EventType() string            { return e.eventType }
func (e *BaseEvent) AggregateID() string          { return e.aggregateID }
func (e *BaseEvent) TenantID() string             { return e.tenantID }
func (e *BaseEvent) UserID() string               { return e.userID }
func (e *BaseEvent) Timestamp() time.Time         { return e.timestamp }
func (e *BaseEvent) Data() map[string]interface{} { return e.data }

func NewBaseEvent(eventType, aggregateID, tenantID, userID string) BaseEvent {
	return BaseEvent{
		eventType:   eventType,
		aggregateID: aggregateID,
		tenantID:    tenantID,
		userID:      userID,
		timestamp:   time.Now().UTC(),
		data:        make(map[string]interface{}),
	}
}

func (e *BaseEvent) With(key string, value interface{}) *BaseEvent {
	e.data[key] = value
	return e
}

func (e *BaseEvent) ToEnvelope() *EventEnvelope {
	return &EventEnvelope{
		ID:            uuid.New().String(),
		Type:          e.eventType,
		AggregateID:   e.aggregateID,
		AggregateType: "",
		TenantID:      e.tenantID,
		Version:       1,
		Timestamp:     e.timestamp,
		UserID:        e.userID,
		Data:          e.data,
	}
}
