package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexManager manages MongoDB indexes for optimal query performance
type IndexManager struct {
	database *mongo.Database
}

// NewIndexManager creates a new index manager
func NewIndexManager(db *mongo.Database) *IndexManager {
	return &IndexManager{database: db}
}

// IndexDefinition defines an index to create
type IndexDefinition struct {
	Collection string
	Keys       bson.D
	Options    *options.IndexOptions
}

// CreateIndexes creates all optimized indexes for the ERP system
func (m *IndexManager) CreateIndexes(ctx context.Context) error {
	indexes := m.getIndexDefinitions()

	for _, idx := range indexes {
		collection := m.database.Collection(idx.Collection)

		model := mongo.IndexModel{
			Keys:    idx.Keys,
			Options: idx.Options,
		}

		_, err := collection.Indexes().CreateOne(ctx, model)
		if err != nil {
			return fmt.Errorf("failed to create index on %s: %w", idx.Collection, err)
		}
	}

	return nil
}

// getIndexDefinitions returns all index definitions for the ERP system
func (m *IndexManager) getIndexDefinitions() []IndexDefinition {
	return []IndexDefinition{
		// Client indexes
		{
			Collection: "clients",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "_id", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_id"),
		},
		{
			Collection: "clients",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "status", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_status"),
		},
		{
			Collection: "clients",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "email", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_email").SetUnique(true),
		},
		{
			Collection: "clients",
			Keys:       bson.D{{Key: "tags", Value: 1}},
			Options:    options.Index().SetName("idx_tags"),
		},

		// Invoice indexes
		{
			Collection: "invoices",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "clientId", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_client"),
		},
		{
			Collection: "invoices",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "status", Value: 1}, {Key: "issuedAt", Value: -1}},
			Options:    options.Index().SetName("idx_tenant_status_date"),
		},
		{
			Collection: "invoices",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "invoiceNumber", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_invoice_num").SetUnique(true),
		},
		{
			Collection: "invoices",
			Keys:       bson.D{{Key: "dueDate", Value: 1}},
			Options:    options.Index().SetName("idx_due_date"),
		},

		// Payment indexes
		{
			Collection: "payments",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "invoiceId", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_invoice"),
		},
		{
			Collection: "payments",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "status", Value: 1}, {Key: "createdAt", Value: -1}},
			Options:    options.Index().SetName("idx_tenant_status_created"),
		},

		// Warehouse indexes
		{
			Collection: "warehouses",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "code", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_code").SetUnique(true),
		},
		{
			Collection: "warehouses",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "isActive", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_active"),
		},

		// Inventory indexes
		{
			Collection: "inventory",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "productId", Value: 1}, {Key: "warehouseId", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_product_warehouse").SetUnique(true),
		},
		{
			Collection: "inventory",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "status", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_status"),
		},
		{
			Collection: "inventory",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "quantity", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_quantity"),
		},

		// Document indexes
		{
			Collection: "documents",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "type", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_type"),
		},
		{
			Collection: "documents",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "processingStatus", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_processing_status"),
		},
		{
			Collection: "documents",
			Keys:       bson.D{{Key: "checksum", Value: 1}},
			Options:    options.Index().SetName("idx_checksum"),
		},
		{
			Collection: "documents",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options:    options.Index().SetName("idx_tenant_created"),
		},

		// Events indexes
		{
			Collection: "events",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "eventType", Value: 1}, {Key: "timestamp", Value: -1}},
			Options:    options.Index().SetName("idx_tenant_event_timestamp"),
		},
		{
			Collection: "events",
			Keys:       bson.D{{Key: "aggregateId", Value: 1}},
			Options:    options.Index().SetName("idx_aggregate_id"),
		},

		// User indexes
		{
			Collection: "users",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "email", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_email").SetUnique(true),
		},
		{
			Collection: "users",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "status", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_status"),
		},

		// Product indexes
		{
			Collection: "products",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "sku", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_sku").SetUnique(true),
		},
		{
			Collection: "products",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "status", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_product_status"),
		},
		{
			Collection: "products",
			Keys:       bson.D{{Key: "tenantId", Value: 1}, {Key: "category", Value: 1}},
			Options:    options.Index().SetName("idx_tenant_category"),
		},
	}
}

// DropAllIndexes drops all non-_id indexes (useful for reindexing)
func (m *IndexManager) DropAllIndexes(ctx context.Context) error {
	collections := []string{
		"clients", "invoices", "payments", "warehouses", "inventory",
		"documents", "events", "users", "products",
	}

	for _, collName := range collections {
		collection := m.database.Collection(collName)
		_, err := collection.Indexes().DropAll(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop indexes from %s: %w", collName, err)
		}
	}

	return nil
}

// AnalyzeQueries analyzes slow queries and suggests indexes
func (m *IndexManager) AnalyzeQueries(ctx context.Context) ([]QueryAnalysis, error) {
	// In a real implementation, this would query MongoDB's profiler
	// For now, return empty
	return []QueryAnalysis{}, nil
}

// QueryAnalysis represents analysis of a query
type QueryAnalysis struct {
	Collection string
	Query      string
	Duration   int64
	Suggestion string
}
