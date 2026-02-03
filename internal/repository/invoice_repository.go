package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MongoInvoiceRepository implements the commands.InvoiceRepository interface
type MongoInvoiceRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger
	tracer     trace.Tracer
}

// NewMongoInvoiceRepository creates a new MongoInvoiceRepository
func NewMongoInvoiceRepository(db *MongoDB, logger *logger.Logger) *MongoInvoiceRepository {
	return &MongoInvoiceRepository{
		collection: db.Collection("invoices"),
		logger:     logger,
		tracer:     otel.Tracer("invoice-repository"),
	}
}

// Create inserts a new invoice into the database
func (r *MongoInvoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) error {
	ctx, span := r.tracer.Start(ctx, "mongo.invoice.create",
		trace.WithAttributes(
			attribute.String("invoice_id", invoice.ID.String()),
			attribute.String("tenant_id", invoice.TenantID.String()),
		),
	)
	defer span.End()

	// Ensure CreatedAt and UpdatedAt are set
	now := time.Now().UTC()
	if invoice.CreatedAt.IsZero() {
		invoice.CreatedAt = now
	}
	if invoice.UpdatedAt.IsZero() {
		invoice.UpdatedAt = now
	}

	_, err := r.collection.InsertOne(ctx, invoice)
	if err != nil {
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to create invoice",
			"invoice_id", invoice.ID,
			"error", err,
		)
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	r.logger.New(ctx).Info("Invoice created",
		"invoice_id", invoice.ID,
		"invoice_number", invoice.InvoiceNumber,
	)

	return nil
}

// Update updates an existing invoice in the database
func (r *MongoInvoiceRepository) Update(ctx context.Context, invoice *domain.Invoice) error {
	ctx, span := r.tracer.Start(ctx, "mongo.invoice.update",
		trace.WithAttributes(
			attribute.String("invoice_id", invoice.ID.String()),
			attribute.Int64("version", invoice.Version),
		),
	)
	defer span.End()

	// Update the UpdatedAt timestamp
	invoice.UpdatedAt = time.Now().UTC()

	filter := bson.M{
		"_id":     invoice.ID,
		"version": invoice.Version,
	}

	update := bson.M{
		"$set": invoice,
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to update invoice",
			"invoice_id", invoice.ID,
			"error", err,
		)
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	if result.MatchedCount == 0 {
		err := fmt.Errorf("invoice not found or version mismatch: %s", invoice.ID)
		span.RecordError(err)
		return err
	}

	invoice.Version++

	r.logger.New(ctx).Info("Invoice updated",
		"invoice_id", invoice.ID,
		"version", invoice.Version,
	)

	return nil
}

// FindByID retrieves an invoice by its ID
func (r *MongoInvoiceRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	ctx, span := r.tracer.Start(ctx, "mongo.invoice.find_by_id",
		trace.WithAttributes(attribute.String("invoice_id", id.String())),
	)
	defer span.End()

	filter := bson.M{"_id": id}

	var invoice domain.Invoice
	err := r.collection.FindOne(ctx, filter).Decode(&invoice)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetAttributes(attribute.String("result", "not_found"))
			return nil, fmt.Errorf("invoice not found: %s", id)
		}
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to find invoice by ID",
			"invoice_id", id,
			"error", err,
		)
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	span.SetAttributes(attribute.String("result", "found"))
	return &invoice, nil
}

// FindByInvoiceNumber retrieves an invoice by its invoice number within a tenant
func (r *MongoInvoiceRepository) FindByInvoiceNumber(ctx context.Context, tenantID uuid.UUID, invoiceNumber string) (*domain.Invoice, error) {
	ctx, span := r.tracer.Start(ctx, "mongo.invoice.find_by_number",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.String("invoice_number", invoiceNumber),
		),
	)
	defer span.End()

	filter := bson.M{
		"tenantId":      tenantID,
		"invoiceNumber": invoiceNumber,
	}

	var invoice domain.Invoice
	err := r.collection.FindOne(ctx, filter).Decode(&invoice)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetAttributes(attribute.String("result", "not_found"))
			return nil, fmt.Errorf("invoice not found: %s", invoiceNumber)
		}
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to find invoice by number",
			"tenant_id", tenantID,
			"invoice_number", invoiceNumber,
			"error", err,
		)
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	span.SetAttributes(attribute.String("result", "found"))
	return &invoice, nil
}

// FindByClientID retrieves invoices for a specific client with pagination
func (r *MongoInvoiceRepository) FindByClientID(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*domain.Invoice, error) {
	ctx, span := r.tracer.Start(ctx, "mongo.invoice.find_by_client",
		trace.WithAttributes(
			attribute.String("client_id", clientID.String()),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		),
	)
	defer span.End()

	filter := bson.M{"clientId": clientID}

	opts := options.Find().
		SetSort(bson.M{"createdAt": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to find invoices by client",
			"client_id", clientID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to find invoices: %w", err)
	}
	defer cursor.Close(ctx)

	var invoices []*domain.Invoice
	if err := cursor.All(ctx, &invoices); err != nil {
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to decode invoices",
			"client_id", clientID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to decode invoices: %w", err)
	}

	span.SetAttributes(attribute.Int("count", len(invoices)))
	return invoices, nil
}
