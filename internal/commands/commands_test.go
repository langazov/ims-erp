package commands_test

import (
	"context"
	"testing"

	"github.com/ims-erp/system/internal/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandEnvelopeCreation(t *testing.T) {
	cmd := commands.NewCommand(
		"client.create",
		"tenant-123",
		"client-456",
		"user-789",
		map[string]interface{}{
			"name":  "Test Client",
			"email": "test@example.com",
		},
	)

	assert.NotEmpty(t, cmd.ID)
	assert.Equal(t, "client.create", cmd.Type)
	assert.Equal(t, "tenant-123", cmd.TenantID)
	assert.Equal(t, "client-456", cmd.TargetID)
	assert.Equal(t, "user-789", cmd.UserID)
	assert.NotEmpty(t, cmd.CorrelationID)
	assert.False(t, cmd.Timestamp.IsZero())
	assert.Equal(t, "Test Client", cmd.Data["name"])
	assert.Equal(t, "test@example.com", cmd.Data["email"])
}

func TestCommandEnvelopeWithCorrelationID(t *testing.T) {
	cmd := commands.NewCommand("test", "tenant", "", "user", nil)
	cmd.WithCorrelationID("custom-correlation-id")

	assert.Equal(t, "custom-correlation-id", cmd.CorrelationID)
}

func TestCommandEnvelopeWithExpectedVersion(t *testing.T) {
	cmd := commands.NewCommand("test", "tenant", "", "user", nil)
	cmd.WithExpectedVersion(5)

	assert.Equal(t, int64(5), cmd.ExpectedVersion)
}

func TestCommandEnvelopeWithMetadata(t *testing.T) {
	cmd := commands.NewCommand("test", "tenant", "", "user", nil)
	cmd.WithMetadata("key1", "value1")
	cmd.WithMetadata("key2", "value2")

	assert.Equal(t, "value1", cmd.Metadata["key1"])
	assert.Equal(t, "value2", cmd.Metadata["key2"])
}

func TestCommandEnvelopeSubject(t *testing.T) {
	cmd := commands.NewCommand("client.create", "tenant", "", "user", nil)

	assert.Equal(t, "cmd.client.create", cmd.Subject())
}

func TestCommandEnvelopeJSON(t *testing.T) {
	cmd := commands.NewCommand(
		"test.type",
		"tenant",
		"target",
		"user",
		map[string]interface{}{"key": "value"},
	)

	data, err := cmd.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	decoded, err := commands.CommandFromJSON(data)
	require.NoError(t, err)
	assert.Equal(t, cmd.ID, decoded.ID)
	assert.Equal(t, cmd.Type, decoded.Type)
	assert.Equal(t, cmd.TenantID, decoded.TenantID)
}

func TestCommandHandlerRegistry(t *testing.T) {
	registry := commands.NewCommandHandlerRegistry()

	handler := func(ctx context.Context, cmd *commands.CommandEnvelope) (interface{}, error) {
		return "handled", nil
	}

	registry.Register("test.command", handler)

	result, ok := registry.GetHandler("test.command")
	assert.True(t, ok)
	assert.NotNil(t, result)

	_, ok = registry.GetHandler("unknown.command")
	assert.False(t, ok)
}

func TestCommandHandlerRegistryHandle(t *testing.T) {
	registry := commands.NewCommandHandlerRegistry()

	registry.Register("test.command", func(ctx context.Context, cmd *commands.CommandEnvelope) (interface{}, error) {
		return cmd.Data["name"], nil
	})

	result, err := registry.Handle(context.Background(), &commands.CommandEnvelope{
		Type: "test.command",
		Data: map[string]interface{}{"name": "John"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "John", result)
}

func TestCommandHandlerRegistryNotFound(t *testing.T) {
	registry := commands.NewCommandHandlerRegistry()

	result, err := registry.Handle(context.Background(), &commands.CommandEnvelope{
		Type: "unknown",
		Data: nil,
	})

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestCommandResult(t *testing.T) {
	result := &commands.CommandResult{
		Success: true,
		Data:    map[string]string{"key": "value"},
		Version: 5,
	}

	assert.True(t, result.Success)
	assert.NotNil(t, result.Data)
	assert.Equal(t, int64(5), result.Version)
}

func TestCommandEnvelopeWithExpectedVersionZero(t *testing.T) {
	cmd := commands.NewCommand("test", "tenant", "", "user", nil)
	cmd.WithExpectedVersion(0)

	assert.Equal(t, int64(0), cmd.ExpectedVersion)
}

func TestCommandEnvelopeMultipleDataFields(t *testing.T) {
	data := map[string]interface{}{
		"name":   "Test",
		"email":  "test@example.com",
		"phone":  "+1234567890",
		"age":    25,
		"active": true,
	}

	cmd := commands.NewCommand("client.create", "tenant", "", "user", data)

	assert.Equal(t, "Test", cmd.Data["name"])
	assert.Equal(t, "test@example.com", cmd.Data["email"])
	assert.Equal(t, "+1234567890", cmd.Data["phone"])
	assert.Equal(t, 25, cmd.Data["age"])
	assert.Equal(t, true, cmd.Data["active"])
}
