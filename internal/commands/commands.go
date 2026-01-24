package commands

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CommandEnvelope struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	TenantID        string                 `json:"tenantId"`
	TargetID        string                 `json:"targetId,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
	CorrelationID   string                 `json:"correlationId"`
	UserID          string                 `json:"userId"`
	ExpectedVersion int64                  `json:"expectedVersion,omitempty"`
	Data            map[string]interface{} `json:"data"`
	Metadata        map[string]string      `json:"metadata"`
}

func NewCommand(commandType, tenantID, targetID, userID string, data map[string]interface{}) *CommandEnvelope {
	return &CommandEnvelope{
		ID:            uuid.New().String(),
		Type:          commandType,
		TenantID:      tenantID,
		TargetID:      targetID,
		Timestamp:     time.Now().UTC(),
		CorrelationID: uuid.New().String(),
		UserID:        userID,
		Data:          data,
		Metadata:      make(map[string]string),
	}
}

func (c *CommandEnvelope) WithCorrelationID(correlationID string) *CommandEnvelope {
	c.CorrelationID = correlationID
	return c
}

func (c *CommandEnvelope) WithExpectedVersion(version int64) *CommandEnvelope {
	c.ExpectedVersion = version
	return c
}

func (c *CommandEnvelope) WithMetadata(key, value string) *CommandEnvelope {
	if c.Metadata == nil {
		c.Metadata = make(map[string]string)
	}
	c.Metadata[key] = value
	return c
}

func (c *CommandEnvelope) Subject() string {
	return "cmd." + c.Type
}

func (c *CommandEnvelope) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

func CommandFromJSON(data []byte) (*CommandEnvelope, error) {
	var cmd CommandEnvelope
	if err := json.Unmarshal(data, &cmd); err != nil {
		return nil, err
	}
	return &cmd, nil
}

type CommandHandler func(ctx context.Context, cmd *CommandEnvelope) (interface{}, error)

type CommandHandlerRegistry struct {
	handlers map[string]CommandHandler
}

func NewCommandHandlerRegistry() *CommandHandlerRegistry {
	return &CommandHandlerRegistry{
		handlers: make(map[string]CommandHandler),
	}
}

func (r *CommandHandlerRegistry) Register(commandType string, handler CommandHandler) {
	r.handlers[commandType] = handler
}

func (r *CommandHandlerRegistry) GetHandler(commandType string) (CommandHandler, bool) {
	handler, ok := r.handlers[commandType]
	return handler, ok
}

func (r *CommandHandlerRegistry) Handle(ctx context.Context, cmd *CommandEnvelope) (interface{}, error) {
	handler, ok := r.GetHandler(cmd.Type)
	if !ok {
		return nil, nil
	}
	return handler(ctx, cmd)
}

type CommandResult struct {
	Success bool
	Data    interface{}
	Error   error
	Version int64
	Events  []interface{}
}
