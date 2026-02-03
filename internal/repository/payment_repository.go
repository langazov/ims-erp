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

// MongoPaymentRepository implements the commands.PaymentRepository interface
type MongoPaymentRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger
	tracer     trace.Tracer
}

// NewMongoPaymentRepository creates a new MongoPaymentRepository
func NewMongoPaymentRepository(db *MongoDB, logger *logger.Logger) *MongoPaymentRepository {
	return &MongoPaymentRepository{
		collection: db.Collection("payments"),
		logger:     logger,
		tracer:     otel.Tracer("payment-repository"),
	}
}

// Create inserts a new payment into the database
func (r *MongoPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	ctx, span := r.tracer.Start(ctx, "mongo.payment.create",
		trace.WithAttributes(
			attribute.String("payment_id", payment.ID.String()),
			attribute.String("invoice_id", payment.InvoiceID.String()),
			attribute.String("tenant_id", payment.TenantID.String()),
		),
	)
	defer span.End()

	// Ensure CreatedAt and UpdatedAt are set
	now := time.Now().UTC()
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = now
	}
	if payment.UpdatedAt.IsZero() {
		payment.UpdatedAt = now
	}

	_, err := r.collection.InsertOne(ctx, payment)
	if err != nil {
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to create payment",
			"payment_id", payment.ID,
			"error", err,
		)
		return fmt.Errorf("failed to create payment: %w", err)
	}

	r.logger.New(ctx).Info("Payment created",
		"payment_id", payment.ID,
		"invoice_id", payment.InvoiceID,
		"amount", payment.Amount.String(),
	)

	return nil
}

// Update updates an existing payment in the database
func (r *MongoPaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	ctx, span := r.tracer.Start(ctx, "mongo.payment.update",
		trace.WithAttributes(
			attribute.String("payment_id", payment.ID.String()),
			attribute.String("status", string(payment.Status)),
		),
	)
	defer span.End()

	// Update the UpdatedAt timestamp
	payment.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": payment.ID}
	update := bson.M{"$set": payment}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to update payment",
			"payment_id", payment.ID,
			"error", err,
		)
		return fmt.Errorf("failed to update payment: %w", err)
	}

	if result.MatchedCount == 0 {
		err := fmt.Errorf("payment not found: %s", payment.ID)
		span.RecordError(err)
		return err
	}

	r.logger.New(ctx).Info("Payment updated",
		"payment_id", payment.ID,
		"status", payment.Status,
	)

	return nil
}

// FindByID retrieves a payment by its ID
func (r *MongoPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	ctx, span := r.tracer.Start(ctx, "mongo.payment.find_by_id",
		trace.WithAttributes(attribute.String("payment_id", id.String())),
	)
	defer span.End()

	filter := bson.M{"_id": id}

	var payment domain.Payment
	err := r.collection.FindOne(ctx, filter).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetAttributes(attribute.String("result", "not_found"))
			return nil, fmt.Errorf("payment not found: %s", id)
		}
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to find payment by ID",
			"payment_id", id,
			"error", err,
		)
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	span.SetAttributes(attribute.String("result", "found"))
	return &payment, nil
}

// FindByInvoiceID retrieves all payments for a specific invoice
func (r *MongoPaymentRepository) FindByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]*domain.Payment, error) {
	ctx, span := r.tracer.Start(ctx, "mongo.payment.find_by_invoice",
		trace.WithAttributes(attribute.String("invoice_id", invoiceID.String())),
	)
	defer span.End()

	filter := bson.M{"invoiceId": invoiceID}

	opts := options.Find().SetSort(bson.M{"createdAt": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to find payments by invoice",
			"invoice_id", invoiceID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*domain.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to decode payments",
			"invoice_id", invoiceID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to decode payments: %w", err)
	}

	span.SetAttributes(attribute.Int("count", len(payments)))
	return payments, nil
}

// FindByProviderID retrieves a payment by its provider ID
func (r *MongoPaymentRepository) FindByProviderID(ctx context.Context, providerID string) (*domain.Payment, error) {
	ctx, span := r.tracer.Start(ctx, "mongo.payment.find_by_provider_id",
		trace.WithAttributes(attribute.String("provider_id", providerID)),
	)
	defer span.End()

	filter := bson.M{"providerId": providerID}

	var payment domain.Payment
	err := r.collection.FindOne(ctx, filter).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetAttributes(attribute.String("result", "not_found"))
			return nil, fmt.Errorf("payment not found for provider ID: %s", providerID)
		}
		span.RecordError(err)
		r.logger.New(ctx).Error("Failed to find payment by provider ID",
			"provider_id", providerID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	span.SetAttributes(attribute.String("result", "found"))
	return &payment, nil
}
