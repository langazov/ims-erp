package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   config.MongoDBConfig
	logger   *logger.Logger
}

func NewMongoDB(cfg config.MongoDBConfig, log *logger.Logger) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime).
		SetServerSelectionTimeout(cfg.ServerSelection)

	if cfg.Username != "" && cfg.Password != "" {
		creds := options.Credential{
			AuthSource: cfg.AuthDatabase,
		}
		clientOpts.SetAuth(creds)
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &MongoDB{
		client:   client,
		database: client.Database(cfg.Database),
		config:   cfg,
		logger:   log,
	}, nil
}

func (m *MongoDB) Client() *mongo.Client {
	return m.client
}

func (m *MongoDB) Database() *mongo.Database {
	return m.database
}

func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func (m *MongoDB) Health(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}

type EventStore struct {
	collection *mongo.Collection
	logger     *logger.Logger
	tracer     trace.Tracer
}

func NewEventStore(db *MongoDB, logger *logger.Logger) *EventStore {
	return &EventStore{
		collection: db.Collection("events"),
		logger:     logger,
		tracer:     otel.Tracer("event-store"),
	}
}

type StoredEvent struct {
	ID            string                 `bson:"_id"`
	AggregateID   string                 `bson:"aggregateId"`
	AggregateType string                 `bson:"aggregateType"`
	EventType     string                 `bson:"eventType"`
	EventData     map[string]interface{} `bson:"eventData"`
	Metadata      EventMetadata          `bson:"metadata"`
	Version       int64                  `bson:"version"`
	Timestamp     time.Time              `bson:"timestamp"`
}

type EventMetadata struct {
	TenantID      string    `bson:"tenantId"`
	UserID        string    `bson:"userId"`
	CorrelationID string    `bson:"correlationId"`
	CausationID   string    `bson:"causationId"`
	Timestamp     time.Time `bson:"timestamp"`
}

func (es *EventStore) Save(ctx context.Context, events []StoredEvent) error {
	ctx, span := es.tracer.Start(ctx, "mongo.save_events")
	defer span.End()

	if len(events) == 0 {
		return nil
	}

	docs := make([]interface{}, len(events))
	for i, e := range events {
		docs[i] = e
		span.AddEvent(fmt.Sprintf("event_%d", i), trace.WithAttributes(
			attribute.String("event_type", e.EventType),
			attribute.String("aggregate_id", e.AggregateID),
		))
	}

	_, err := es.collection.InsertMany(ctx, docs)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to save events: %w", err)
	}

	return nil
}

func (es *EventStore) Load(ctx context.Context, aggregateID string) ([]StoredEvent, error) {
	ctx, span := es.tracer.Start(ctx, "mongo.load_events",
		trace.WithAttributes(attribute.String("aggregate_id", aggregateID)),
	)
	defer span.End()

	filter := map[string]interface{}{
		"aggregateId": aggregateID,
	}

	opts := options.Find().SetSort(map[string]int{"version": 1})
	cursor, err := es.collection.Find(ctx, filter, opts)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to load events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []StoredEvent
	if err := cursor.All(ctx, &events); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	span.SetAttributes(attribute.Int("event_count", len(events)))
	return events, nil
}

func (es *EventStore) LoadByType(ctx context.Context, aggregateType string, tenantID string, from time.Time) ([]StoredEvent, error) {
	ctx, span := es.tracer.Start(ctx, "mongo.load_events_by_type",
		trace.WithAttributes(
			attribute.String("aggregate_type", aggregateType),
			attribute.String("tenant_id", tenantID),
		),
	)
	defer span.End()

	filter := map[string]interface{}{
		"aggregateType":     aggregateType,
		"metadata.tenantId": tenantID,
		"timestamp": map[string]interface{}{
			"$gte": from,
		},
	}

	cursor, err := es.collection.Find(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to load events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []StoredEvent
	if err := cursor.All(ctx, &events); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

func (es *EventStore) GetLatestVersion(ctx context.Context, aggregateID string) (int64, error) {
	ctx, span := es.tracer.Start(ctx, "mongo.get_latest_version",
		trace.WithAttributes(attribute.String("aggregate_id", aggregateID)),
	)
	defer span.End()

	filter := map[string]interface{}{
		"aggregateId": aggregateID,
	}

	opts := options.FindOne().SetSort(map[string]int{"version": -1})
	var event StoredEvent
	if err := es.collection.FindOne(ctx, filter, opts).Decode(&event); err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		span.RecordError(err)
		return 0, fmt.Errorf("failed to get latest version: %w", err)
	}

	return event.Version, nil
}

type ReadModelStore struct {
	collection *mongo.Collection
	logger     *logger.Logger
	tracer     trace.Tracer
}

func NewReadModelStore(db *MongoDB, collectionName string, logger *logger.Logger) *ReadModelStore {
	return &ReadModelStore{
		collection: db.Collection(collectionName),
		logger:     logger,
		tracer:     otel.Tracer("read-model-store"),
	}
}

func (s *ReadModelStore) Save(ctx context.Context, model interface{}) error {
	ctx, span := s.tracer.Start(ctx, "mongo.save_read_model")
	defer span.End()

	_, err := s.collection.InsertOne(ctx, model)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to save read model: %w", err)
	}

	return nil
}

func (s *ReadModelStore) Update(ctx context.Context, filter interface{}, update interface{}) error {
	ctx, span := s.tracer.Start(ctx, "mongo.update_read_model")
	defer span.End()

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update read model: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

func (s *ReadModelStore) Upsert(ctx context.Context, filter interface{}, update interface{}) error {
	ctx, span := s.tracer.Start(ctx, "mongo.upsert_read_model")
	defer span.End()

	_, err := s.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to upsert read model: %w", err)
	}

	return nil
}

func (s *ReadModelStore) FindOne(ctx context.Context, filter interface{}) (interface{}, error) {
	ctx, span := s.tracer.Start(ctx, "mongo.find_one_read_model")
	defer span.End()

	var result interface{}
	err := s.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find read model: %w", err)
	}

	return result, nil
}

func (s *ReadModelStore) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]interface{}, error) {
	ctx, span := s.tracer.Start(ctx, "mongo.find_read_models")
	defer span.End()

	cursor, err := s.collection.Find(ctx, filter, opts...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find read models: %w", err)
	}
	defer cursor.Close(ctx)

	var results []interface{}
	if err := cursor.All(ctx, &results); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode read models: %w", err)
	}

	return results, nil
}

func (s *ReadModelStore) Delete(ctx context.Context, filter interface{}) error {
	ctx, span := s.tracer.Start(ctx, "mongo.delete_read_model")
	defer span.End()

	_, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete read model: %w", err)
	}

	return nil
}

func (s *ReadModelStore) Count(ctx context.Context, filter interface{}) (int64, error) {
	ctx, span := s.tracer.Start(ctx, "mongo.count_read_models")
	defer span.End()

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to count read models: %w", err)
	}

	return count, nil
}

type AggregateStore struct {
	eventStore     *EventStore
	readModelStore *ReadModelStore
	logger         *logger.Logger
}

func NewAggregateStore(db *MongoDB, eventCollection, readModelCollection string, logger *logger.Logger) *AggregateStore {
	return &AggregateStore{
		eventStore:     NewEventStore(db, logger),
		readModelStore: NewReadModelStore(db, readModelCollection, logger),
		logger:         logger,
	}
}

func (s *AggregateStore) LoadEvents(ctx context.Context, aggregateID string) ([]StoredEvent, error) {
	return s.eventStore.Load(ctx, aggregateID)
}

func (s *AggregateStore) SaveEvents(ctx context.Context, events []StoredEvent) error {
	return s.eventStore.Save(ctx, events)
}

func (s *AggregateStore) SaveReadModel(ctx context.Context, model interface{}) error {
	return s.readModelStore.Save(ctx, model)
}

func (s *AggregateStore) UpdateReadModel(ctx context.Context, filter, update interface{}) error {
	return s.readModelStore.Update(ctx, filter, update)
}

func (s *AggregateStore) FindReadModel(ctx context.Context, filter interface{}) (interface{}, error) {
	return s.readModelStore.FindOne(ctx, filter)
}
