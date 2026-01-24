package queries

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClientByIDQuery(t *testing.T) {
	query := &GetClientByIDQuery{
		ClientID: "client-123",
		TenantID: "tenant-456",
	}

	assert.Equal(t, "client-123", query.ClientID)
	assert.Equal(t, "tenant-456", query.TenantID)
}

func TestGetClientDetailQuery(t *testing.T) {
	query := &GetClientDetailQuery{
		ClientID: "client-123",
		TenantID: "tenant-456",
	}

	assert.Equal(t, "client-123", query.ClientID)
	assert.Equal(t, "tenant-456", query.TenantID)
}

func TestListClientsQuery(t *testing.T) {
	query := &ListClientsQuery{
		TenantID:  "tenant-456",
		Page:      1,
		PageSize:  20,
		Search:    "test",
		Status:    "active",
		Tags:      []string{"vip", "premium"},
		SortBy:    "name",
		SortOrder: "asc",
	}

	assert.Equal(t, "tenant-456", query.TenantID)
	assert.Equal(t, 1, query.Page)
	assert.Equal(t, 20, query.PageSize)
	assert.Equal(t, "test", query.Search)
	assert.Equal(t, "active", query.Status)
	assert.Len(t, query.Tags, 2)
	assert.Equal(t, "name", query.SortBy)
	assert.Equal(t, "asc", query.SortOrder)
}

func TestSearchClientsQuery(t *testing.T) {
	query := &SearchClientsQuery{
		TenantID: "tenant-456",
		Term:     "test search",
		Limit:    10,
	}

	assert.Equal(t, "tenant-456", query.TenantID)
	assert.Equal(t, "test search", query.Term)
	assert.Equal(t, 10, query.Limit)
}

func TestGetClientCreditStatusQuery(t *testing.T) {
	query := &GetClientCreditStatusQuery{
		ClientID: "client-123",
		TenantID: "tenant-456",
	}

	assert.Equal(t, "client-123", query.ClientID)
	assert.Equal(t, "tenant-456", query.TenantID)
}

func TestListClientsResult(t *testing.T) {
	result := &ListClientsResult{
		Total:      100,
		Page:       1,
		PageSize:   20,
		TotalPages: 5,
	}

	assert.Equal(t, int64(100), result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
	assert.Equal(t, 5, result.TotalPages)
	assert.Nil(t, result.Clients)
}

func TestGetSortOrder(t *testing.T) {
	tests := []struct {
		name     string
		order    string
		expected int
	}{
		{"ascending order", "asc", 1},
		{"descending order", "desc", -1},
		{"empty order defaults to asc", "", 1},
		{"unknown order defaults to asc", "random", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSortOrder(tt.order)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListClientsQueryDefaults(t *testing.T) {
	query := &ListClientsQuery{
		TenantID: "tenant-456",
	}

	assert.Equal(t, "tenant-456", query.TenantID)
	assert.Equal(t, 0, query.Page)
	assert.Equal(t, 0, query.PageSize)
	assert.Equal(t, "", query.Search)
	assert.Equal(t, "", query.Status)
	assert.Nil(t, query.Tags)
	assert.Equal(t, "", query.SortBy)
	assert.Equal(t, "", query.SortOrder)
}

func TestSearchClientsQueryDefaults(t *testing.T) {
	query := &SearchClientsQuery{
		TenantID: "tenant-456",
	}

	assert.Equal(t, "tenant-456", query.TenantID)
	assert.Equal(t, "", query.Term)
	assert.Equal(t, 0, query.Limit)
}
