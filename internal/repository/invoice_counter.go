package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// InvoiceCounterDocument represents the counter document in MongoDB
type InvoiceCounterDocument struct {
	TenantID  string    `bson:"_id"` // Composite key: "tenantID-year"
	Tenant    uuid.UUID `bson:"tenantId"`
	Year      int       `bson:"year"`
	Sequence  int64     `bson:"sequence"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

// MongoInvoiceCounter implements the commands.InvoiceCounter interface
type MongoInvoiceCounter struct {
	collection *mongo.Collection
	logger     *logger.Logger
	tracer     trace.Tracer
}

// NewMongoInvoiceCounter creates a new MongoInvoiceCounter
func NewMongoInvoiceCounter(db *MongoDB, logger *logger.Logger) *MongoInvoiceCounter {
	return &MongoInvoiceCounter{
		collection: db.Collection("invoice_counters"),
		logger:     logger,
		tracer:     otel.Tracer("invoice-counter"),
	}
}

// GetNextInvoiceNumber atomically increments the counter and returns a formatted invoice number
// Format: "INV-{year}-{sequence:06d}"
func (c *MongoInvoiceCounter) GetNextInvoiceNumber(ctx context.Context, tenantID uuid.UUID, year int) (string, error) {
	ctx, span := c.tracer.Start(ctx, "mongo.invoice_counter.get_next",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.Int("year", year),
		),
	)
	defer span.End()

	// Create composite key for tenant and year sharding
	compositeKey := fmt.Sprintf("%s-%d", tenantID.String(), year)

	filter := bson.M{"_id": compositeKey}

	// Atomically increment the sequence counter
	update := bson.M{
		"$inc": bson.M{"sequence": 1},
		"$set": bson.M{
			"tenantId":  tenantID,
			"year":      year,
			"updatedAt": time.Now().UTC(),
		},
	}

	// Return the document after update to get the new sequence number
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result InvoiceCounterDocument
	err := c.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		span.RecordError(err)
		c.logger.New(ctx).Error("Failed to get next invoice number",
			"tenant_id", tenantID,
			"year", year,
			"error", err,
		)
		return "", fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Format the invoice number: INV-{year}-{sequence:06d}
	invoiceNumber := fmt.Sprintf("INV-%d-%06d", year, result.Sequence)

	span.SetAttributes(attribute.String("invoice_number", invoiceNumber))
	c.logger.New(ctx).Info("Generated invoice number",
		"tenant_id", tenantID,
		"year", year,
		"sequence", result.Sequence,
		"invoice_number", invoiceNumber,
	)

	return invoiceNumber, nil
}

// GetCurrentSequence returns the current sequence number for a tenant and year (for testing/monitoring)
func (c *MongoInvoiceCounter) GetCurrentSequence(ctx context.Context, tenantID uuid.UUID, year int) (int64, error) {
	ctx, span := c.tracer.Start(ctx, "mongo.invoice_counter.get_current",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.Int("year", year),
		),
	)
	defer span.End()

	compositeKey := fmt.Sprintf("%s-%d", tenantID.String(), year)
	filter := bson.M{"_id": compositeKey}

	var result InvoiceCounterDocument
	err := c.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No counter exists yet, return 0
			return 0, nil
		}
		span.RecordError(err)
		return 0, fmt.Errorf("failed to get current sequence: %w", err)
	}

	return result.Sequence, nil
}

// ResetCounter resets the counter for a tenant and year (use with caution, mainly for testing)
func (c *MongoInvoiceCounter) ResetCounter(ctx context.Context, tenantID uuid.UUID, year int) error {
	ctx, span := c.tracer.Start(ctx, "mongo.invoice_counter.reset",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.Int("year", year),
		),
	)
	defer span.End()

	compositeKey := fmt.Sprintf("%s-%d", tenantID.String(), year)
	filter := bson.M{"_id": compositeKey}

	update := bson.M{
		"$set": bson.M{
			"sequence":  0,
			"tenantId":  tenantID,
			"year":      year,
			"updatedAt": time.Now().UTC(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := c.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		span.RecordError(err)
		c.logger.New(ctx).Error("Failed to reset invoice counter",
			"tenant_id", tenantID,
			"year", year,
			"error", err,
		)
		return fmt.Errorf("failed to reset counter: %w", err)
	}

	c.logger.New(ctx).Info("Invoice counter reset",
		"tenant_id", tenantID,
		"year", year,
	)

	return nil
}
