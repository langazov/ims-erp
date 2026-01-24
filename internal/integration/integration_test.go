package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/internal/messaging"
	"github.com/ims-erp/system/internal/queries"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func setupTestEnvironment(t *testing.T) (*repository.MongoDB, *repository.Redis, *messaging.Publisher, *messaging.Subscriber, *repository.Cache, *logger.Logger, func()) {
	cfg := &config.Config{
		MongoDB: config.MongoDBConfig{
			URI:             "mongodb://localhost:27017",
			Database:        "erp_system",
			MaxPoolSize:     10,
			MinPoolSize:     1,
			ConnectTimeout:  10 * time.Second,
			ServerSelection: 5 * time.Second,
		},
		Redis: config.RedisConfig{
			Mode:      "standalone",
			Addresses: []string{"localhost:6379"},
			PoolSize:  10,
		},
		NATS: config.NATSConfig{
			URLs:           []string{"nats://localhost:4222"},
			MaxReconnect:   10,
			ReconnectWait:  2 * time.Second,
			ConnectTimeout: 10 * time.Second,
			JetStream: config.JetStreamConfig{
				Enabled:      true,
				StreamPrefix: "test-",
			},
		},
	}

	log, err := logger.New(logger.Config{
		Level:       "debug",
		Format:      "json",
		ServiceName: "integration-test",
	})
	require.NoError(t, err)

	mongodb, err := repository.NewMongoDB(cfg.MongoDB, log)
	require.NoError(t, err)

	redis, err := repository.NewRedis(cfg.Redis, log)
	require.NoError(t, err)

	natsConfig := messaging.NATSConfig{
		URLs:           cfg.NATS.URLs,
		MaxReconnect:   cfg.NATS.MaxReconnect,
		ReconnectWait:  cfg.NATS.ReconnectWait,
		ConnectTimeout: cfg.NATS.ConnectTimeout,
		JetStream:      cfg.NATS.JetStream.Enabled,
		StreamPrefix:   cfg.NATS.JetStream.StreamPrefix,
	}

	publisher, err := messaging.NewPublisher(natsConfig, log)
	require.NoError(t, err)

	subscriber, err := messaging.NewSubscriber(natsConfig, log)
	require.NoError(t, err)

	cache := repository.NewCache(redis, "test", log)

	cleanup := func() {
		mongodb.Close(context.Background())
		redis.Close()
		publisher.Close()
		subscriber.Close()
	}

	return mongodb, redis, publisher, subscriber, cache, log, cleanup
}

func TestClientEventProjection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mongodb, _, publisher, _, cache, log, cleanup := setupTestEnvironment(t)
	defer cleanup()

	readModelStore := repository.NewReadModelStore(mongodb, "client_read", log)
	eventHandler := events.NewClientEventHandler(readModelStore, cache, log)

	ctx := context.Background()
	tenantID := "test-tenant-123"

	t.Run("should project ClientCreated event", func(t *testing.T) {
		event := &events.EventEnvelope{
			ID:            "event-1",
			Type:          "ClientCreated",
			AggregateID:   "client-1",
			AggregateType: "Client",
			TenantID:      tenantID,
			UserID:        "user-1",
			Data: map[string]interface{}{
				"name":        "Test Client",
				"email":       "test@example.com",
				"phone":       "+1234567890",
				"creditLimit": "10000",
				"status":      "active",
				"tags":        []string{"vip", "enterprise"},
				"billingAddress": map[string]interface{}{
					"street":     "123 Main St",
					"city":       "New York",
					"state":      "NY",
					"postalCode": "10001",
					"country":    "USA",
				},
			},
			Timestamp: time.Now(),
		}

		err := eventHandler.HandleClientCreated(ctx, event)
		require.NoError(t, err)

		summary, err := readModelStore.FindOne(ctx, map[string]interface{}{
			"_id":      "client-1",
			"tenantId": tenantID,
		})
		require.NoError(t, err)
		require.NotNil(t, summary)

		summaryMap, ok := summary.(events.ClientSummary)
		require.True(t, ok)
		require.Equal(t, "Test Client", summaryMap.Name)
		require.Equal(t, "test@example.com", summaryMap.Email)
	})

	t.Run("should project ClientUpdated event", func(t *testing.T) {
		event := &events.EventEnvelope{
			ID:            "event-2",
			Type:          "ClientUpdated",
			AggregateID:   "client-1",
			AggregateType: "Client",
			TenantID:      tenantID,
			UserID:        "user-1",
			Data: map[string]interface{}{
				"name":  "Updated Client Name",
				"email": "updated@example.com",
			},
			Timestamp: time.Now(),
		}

		err := eventHandler.HandleClientUpdated(ctx, event)
		require.NoError(t, err)
	})

	t.Run("should project ClientDeactivated event", func(t *testing.T) {
		event := &events.EventEnvelope{
			ID:            "event-3",
			Type:          "ClientDeactivated",
			AggregateID:   "client-1",
			AggregateType: "Client",
			TenantID:      tenantID,
			UserID:        "user-1",
			Data: map[string]interface{}{
				"reason": "Customer request",
			},
			Timestamp: time.Now(),
		}

		err := eventHandler.HandleClientDeactivated(ctx, event)
		require.NoError(t, err)
	})

	_ = publisher
}

func TestClientQueryHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mongodb, _, _, _, cache, log, cleanup := setupTestEnvironment(t)
	defer cleanup()

	readModelStore := repository.NewReadModelStore(mongodb, "client_read", log)
	queryHandler := queries.NewClientQueryHandler(readModelStore, cache, log)

	ctx := context.Background()
	tenantID := "test-tenant-456"

	t.Run("should list clients with pagination", func(t *testing.T) {
		query := &queries.ListClientsQuery{
			TenantID:  tenantID,
			Page:      1,
			PageSize:  10,
			SortBy:    "name",
			SortOrder: "asc",
		}

		result, err := queryHandler.ListClients(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("should search clients", func(t *testing.T) {
		query := &queries.SearchClientsQuery{
			TenantID: tenantID,
			Term:     "test",
			Limit:    5,
		}

		clients, err := queryHandler.SearchClients(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, clients)
	})

	t.Run("should get client credit status", func(t *testing.T) {
		query := &queries.GetClientCreditStatusQuery{
			ClientID: "non-existent-client",
			TenantID: tenantID,
		}

		status, err := queryHandler.GetClientCreditStatus(ctx, query)
		require.NoError(t, err)
		require.Nil(t, status)
	})
}

func TestInvoiceDomain(t *testing.T) {
	tenantID := uuid.New()
	clientID := uuid.New()
	createdBy := uuid.New()

	t.Run("should create invoice with lines", func(t *testing.T) {
		invoice, err := domain.NewInvoice(
			tenantID,
			clientID,
			createdBy,
			domain.InvoiceTypeStandard,
			"USD",
			domain.PaymentTermNet30,
			time.Now(),
		)
		require.NoError(t, err)
		require.NotNil(t, invoice)

		invoice.AddLine(domain.InvoiceLine{
			Description: "Consulting Services",
			Quantity:    decimal.NewFromInt(40),
			UnitPrice:   decimal.NewFromFloat(150.00),
			Discount:    decimal.Zero,
			TaxRate:     decimal.NewFromFloat(10),
		})

		require.Equal(t, domain.InvoiceStatusDraft, invoice.Status)
		require.Len(t, invoice.Lines, 1)
	})

	t.Run("should calculate due date based on payment term", func(t *testing.T) {
		invoice, err := domain.NewInvoice(
			tenantID,
			clientID,
			createdBy,
			domain.InvoiceTypeStandard,
			"USD",
			domain.PaymentTermNet30,
			time.Now(),
		)
		require.NoError(t, err)

		dueDate := invoice.CalculateDueDate()
		expectedDue := time.Now().AddDate(0, 0, 30)
		require.Equal(t, expectedDue.Year(), dueDate.Year())
		require.Equal(t, expectedDue.Month(), dueDate.Month())
		require.Equal(t, expectedDue.Day(), dueDate.Day())
	})

	t.Run("should detect overdue invoices", func(t *testing.T) {
		invoice, err := domain.NewInvoice(
			tenantID,
			clientID,
			createdBy,
			domain.InvoiceTypeStandard,
			"USD",
			domain.PaymentTermNet15,
			time.Now().AddDate(0, 0, -30),
		)
		require.NoError(t, err)

		sentDate := time.Now().AddDate(0, 0, -30)
		invoice.Send()
		invoice.SentDate = &sentDate
		dueDate := time.Now().AddDate(0, 0, -15)
		invoice.DueDate = &dueDate

		require.True(t, invoice.IsOverdue())
	})
}

func TestPaymentDomain(t *testing.T) {
	tenantID := uuid.New()
	invoiceID := uuid.New()
	clientID := uuid.New()

	t.Run("should create payment", func(t *testing.T) {
		payment := domain.NewPayment(
			tenantID,
			invoiceID,
			clientID,
			decimal.NewFromFloat(1000.00),
			"USD",
			domain.PaymentMethodCreditCard,
		)
		require.NotNil(t, payment)
		require.Equal(t, domain.PaymentStatusPending, payment.Status)
	})

	t.Run("should process payment with Stripe", func(t *testing.T) {
		processor := domain.NewStripeProcessor("sk_test_xxx", "whsec_xxx")
		require.NotNil(t, processor)

		req := &domain.PaymentRequest{
			InvoiceID:   invoiceID,
			Amount:      decimal.NewFromFloat(500.00),
			Currency:    "USD",
			Method:      domain.PaymentMethodStripe,
			Description: "Test payment",
		}

		result, err := processor.ProcessPayment(context.Background(), req)
		require.NoError(t, err)
		require.True(t, result.Success)
		require.Equal(t, domain.PaymentStatusCompleted, result.Status)
	})

	t.Run("should process refund", func(t *testing.T) {
		processor := domain.NewStripeProcessor("sk_test_xxx", "whsec_xxx")
		paymentID := uuid.New()

		req := &domain.RefundRequest{
			PaymentID:  paymentID,
			Amount:     decimal.NewFromFloat(250.00),
			Reason:     "Customer request",
			RefundType: "full",
		}

		result, err := processor.ProcessRefund(context.Background(), req)
		require.NoError(t, err)
		require.True(t, result.Success)
		require.Equal(t, domain.PaymentStatusRefunded, result.Status)
	})
}
