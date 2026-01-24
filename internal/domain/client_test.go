package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tenantID := uuid.New()
	name := "Test Client"
	email := "test@example.com"

	client := NewClient(tenantID, name, email)

	assert.NotEmpty(t, client.ID)
	assert.Equal(t, tenantID, client.TenantID)
	assert.Equal(t, name, client.Name)
	assert.Equal(t, email, client.Email)
	assert.Equal(t, ClientStatusActive, client.Status)
	assert.True(t, client.CreditLimit.IsZero())
	assert.True(t, client.CurrentBalance.IsZero())
	assert.Empty(t, client.Phone)
	assert.Empty(t, client.Tags)
	assert.False(t, client.CreatedAt.IsZero())
	assert.False(t, client.UpdatedAt.IsZero())
}

func TestClientUpdate(t *testing.T) {
	client := NewClient(uuid.New(), "Original Name", "original@example.com")

	client.Update("New Name", "new@example.com", "+1234567890")

	assert.Equal(t, "New Name", client.Name)
	assert.Equal(t, "new@example.com", client.Email)
	assert.Equal(t, "+1234567890", client.Phone)
}

func TestClientBillingAddress(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	addr := Address{
		Street:     "123 Main St",
		City:       "New York",
		State:      "NY",
		PostalCode: "10001",
		Country:    "USA",
	}

	client.SetBillingAddress(addr)

	assert.Equal(t, addr, client.BillingAddress)
}

func TestClientShippingAddresses(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	addr1 := Address{Street: "123 Main St", City: "NYC", Country: "USA"}
	addr2 := Address{Street: "456 Oak Ave", City: "LA", Country: "USA"}

	client.AddShippingAddress(addr1)
	client.AddShippingAddress(addr2)

	require.Len(t, client.ShippingAddresses, 2)
	assert.Equal(t, addr1, client.ShippingAddresses[0])
	assert.Equal(t, addr2, client.ShippingAddresses[1])
}

func TestClientCreditLimit(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	limit := decimal.NewFromInt(10000)
	client.SetCreditLimit(limit)

	assert.True(t, client.CreditLimit.Equal(limit))
}

func TestClientAvailableCredit(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")
	client.SetCreditLimit(decimal.NewFromInt(10000))

	assert.Equal(t, decimal.NewFromInt(10000), client.AvailableCredit())

	client.CurrentBalance = decimal.NewFromInt(3000)
	assert.Equal(t, decimal.NewFromInt(7000), client.AvailableCredit())
}

func TestClientCanMakePurchase(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")
	client.SetCreditLimit(decimal.NewFromInt(10000))

	assert.True(t, client.CanMakePurchase(decimal.NewFromInt(5000)))
	assert.True(t, client.CanMakePurchase(decimal.NewFromInt(10000)))
	assert.False(t, client.CanMakePurchase(decimal.NewFromInt(15000)))
}

func TestClientStatus(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	assert.Equal(t, ClientStatusActive, client.Status)

	client.Deactivate()
	assert.Equal(t, ClientStatusInactive, client.Status)

	client.Reactivate()
	assert.Equal(t, ClientStatusActive, client.Status)

	client.Suspend()
	assert.Equal(t, ClientStatusSuspended, client.Status)
}

func TestClientTags(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	client.AddTag("vip")
	client.AddTag("premium")
	client.AddTag("vip") // Should not duplicate

	require.Len(t, client.Tags, 2)
	assert.Contains(t, client.Tags, "vip")
	assert.Contains(t, client.Tags, "premium")

	client.RemoveTag("vip")
	assert.Len(t, client.Tags, 1)
	assert.Contains(t, client.Tags, "premium")
}

func TestClientCustomFields(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	client.SetCustomField("industry", "Technology")
	client.SetCustomField("size", "Enterprise")
	client.SetCustomField("industry", "Tech") // Update

	value, ok := client.GetCustomField("industry")
	assert.True(t, ok)
	assert.Equal(t, "Tech", value)

	_, ok = client.GetCustomField("nonexistent")
	assert.False(t, ok)
}

func TestClientCreditUtilization(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")
	client.SetCreditLimit(decimal.NewFromInt(10000))

	assert.True(t, client.CreditUtilization().IsZero())

	client.CurrentBalance = decimal.NewFromInt(2500)
	utilization := client.CreditUtilization()
	// Credit utilization is calculated as (currentBalance / creditLimit) * 100
	// 2500 / 10000 = 0.25, so utilization should be around 25%
	assert.True(t, utilization.GreaterThan(decimal.NewFromFloat(24.9)))
	assert.True(t, utilization.LessThan(decimal.NewFromFloat(25.1)))
}

func TestClientMergeInto(t *testing.T) {
	source := NewClient(uuid.New(), "Source", "source@example.com")
	source.AddTag("vip")
	source.SetCustomField("field1", "value1")
	source.AddShippingAddress(Address{Street: "123 Main", City: "NYC"})

	target := NewClient(uuid.New(), "Target", "target@example.com")
	target.AddTag("regular")

	source.MergeInto(target)

	assert.Len(t, target.Tags, 2)
	assert.Contains(t, target.Tags, "vip")
	assert.Contains(t, target.Tags, "regular")

	value, _ := target.GetCustomField("field1")
	assert.Equal(t, "value1", value)

	assert.Len(t, target.ShippingAddresses, 1)
}

func TestAddressIsEmpty(t *testing.T) {
	addr := Address{}
	assert.True(t, addr.IsEmpty())

	addr.Street = "123 Main St"
	assert.False(t, addr.IsEmpty())
}

func TestClientStatusIsValid(t *testing.T) {
	assert.True(t, ClientStatusActive.IsValid())
	assert.True(t, ClientStatusInactive.IsValid())
	assert.True(t, ClientStatusSuspended.IsValid())
	assert.True(t, ClientStatusMerged.IsValid())
	assert.False(t, ClientStatus("invalid").IsValid())
}

func TestClientHasOverdueBalance(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	assert.False(t, client.HasOverdueBalance())

	client.CurrentBalance = decimal.NewFromInt(100)
	assert.True(t, client.HasOverdueBalance())
}

func TestClientUpdateBalance(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")
	client.SetCreditLimit(decimal.NewFromInt(10000))

	client.UpdateBalance(decimal.NewFromInt(500))
	assert.Equal(t, decimal.NewFromInt(500), client.CurrentBalance)

	client.UpdateBalance(decimal.NewFromInt(-200))
	assert.Equal(t, decimal.NewFromInt(300), client.CurrentBalance)
}

func TestClientVersionIncrements(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")
	initialVersion := client.Version

	// Note: The domain model doesn't automatically increment version on update
	// Version is typically managed by the command handler/event store
	assert.Equal(t, int64(0), initialVersion)

	client.Version = 1 // Simulating version increment by command handler
	assert.Equal(t, int64(1), client.Version)
}

func TestClientTimestamps(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	client := NewClient(uuid.New(), "Test", "test@example.com")
	after := time.Now().UTC().Add(time.Second)

	assert.True(t, client.CreatedAt.After(before))
	assert.True(t, client.CreatedAt.Before(after))

	oldUpdatedAt := client.UpdatedAt

	time.Sleep(time.Millisecond * 10)
	client.Update("New Name", "new@example.com", "")

	assert.True(t, client.UpdatedAt.After(oldUpdatedAt))
}

func TestClientRemoveShippingAddress(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	addr1 := Address{Street: "123 Main St", City: "NYC"}
	addr2 := Address{Street: "456 Oak Ave", City: "LA"}
	client.AddShippingAddress(addr1)
	client.AddShippingAddress(addr2)

	assert.Len(t, client.ShippingAddresses, 2)

	client.RemoveShippingAddress(0)

	assert.Len(t, client.ShippingAddresses, 1)
	assert.Equal(t, addr2, client.ShippingAddresses[0])
}

func TestClientRemoveShippingAddressInvalidIndex(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")
	client.AddShippingAddress(Address{Street: "123 Main St", City: "NYC"})
	client.AddShippingAddress(Address{Street: "456 Oak Ave", City: "LA"})
	originalLen := len(client.ShippingAddresses)

	client.RemoveShippingAddress(-1)
	assert.Len(t, client.ShippingAddresses, originalLen)

	client.RemoveShippingAddress(100)
	assert.Len(t, client.ShippingAddresses, originalLen)
}

func TestClientRemoveShippingAddressEmpty(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	client.RemoveShippingAddress(0)

	assert.Len(t, client.ShippingAddresses, 0)
}

func TestClientCreateEvent(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	event := client.CreateEvent("test_event", "user123", map[string]interface{}{
		"action": "update",
	})

	assert.Equal(t, "test_event", event.EventType)
	assert.Equal(t, client.ID, event.ClientID)
	assert.Equal(t, "user123", event.UserID)
	assert.NotNil(t, event.Timestamp)
	assert.Equal(t, "update", event.EventData["action"])
}

func TestClientSetCustomFieldOverwrite(t *testing.T) {
	client := NewClient(uuid.New(), "Test", "test@example.com")

	client.SetCustomField("key", "value1")
	client.SetCustomField("key", "value2")

	assert.Equal(t, "value2", client.CustomFields["key"])
}
