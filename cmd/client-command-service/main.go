package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ims-erp/system/internal/commands"
	"github.com/ims-erp/system/internal/config"
	eventpkg "github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/internal/health"
	"github.com/ims-erp/system/internal/messaging"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/ims-erp/system/pkg/tracer"
	"github.com/shopspring/decimal"
)

func main() {
	cfg, err := config.Load("", "client-command-service")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(logger.Config{
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		ServiceName: cfg.App.Name,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	tr, err := tracer.New(tracer.Config{
		Enabled:      cfg.Tracing.Enabled,
		ServiceName:  cfg.App.Name,
		ExporterType: cfg.Tracing.ExporterType,
		Endpoint:     cfg.Tracing.Endpoint,
		SamplerType:  cfg.Tracing.SamplerType,
		SamplerRatio: cfg.Tracing.SamplerRatio,
	})
	if err != nil {
		log.Error("Failed to create tracer", "error", err)
		os.Exit(1)
	}
	defer tr.Shutdown(context.Background())

	messaging.SetupTracePropagation()

	mongodb, err := repository.NewMongoDB(cfg.MongoDB, log)
	if err != nil {
		log.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer mongodb.Close(context.Background())
	log.Info("Connected to MongoDB")

	redis, err := repository.NewRedis(cfg.Redis, log)
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redis.Close()
	log.Info("Connected to Redis")

	natsConfig := messaging.NATSConfig{
		URLs:           cfg.NATS.URLs,
		Username:       cfg.NATS.Username,
		Password:       cfg.NATS.Password,
		Token:          cfg.NATS.Token,
		MaxReconnect:   cfg.NATS.MaxReconnect,
		ReconnectWait:  cfg.NATS.ReconnectWait,
		ConnectTimeout: cfg.NATS.ConnectTimeout,
		JetStream:      cfg.NATS.JetStream.Enabled,
		StreamPrefix:   cfg.NATS.JetStream.StreamPrefix,
	}

	publisher, err := messaging.NewPublisher(natsConfig, log)
	if err != nil {
		log.Error("Failed to create NATS publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()
	log.Info("Connected to NATS")

	eventStore := repository.NewEventStore(mongodb, log)
	readModelStore := repository.NewReadModelStore(mongodb, "client_read", log)
	cache := repository.NewCache(redis, "t:"+cfg.MongoDB.Database, log)

	defaultCreditLimit := decimal.NewFromInt(10000)

	clientCmdHandler := commands.NewClientCommandHandler(
		eventStore,
		publisher,
		log,
		commands.TenantConfig{
			AutoGenerateCode:   true,
			CodePrefix:         "CLT",
			DefaultCreditLimit: defaultCreditLimit,
			RequireEmail:       true,
		},
	)

	cmdRegistry := commands.NewCommandHandlerRegistry()
	cmdRegistry.Register("client.create", func(ctx context.Context, cmd *commands.CommandEnvelope) (interface{}, error) {
		return clientCmdHandler.HandleCreateClient(ctx, cmd)
	})
	cmdRegistry.Register("client.update", func(ctx context.Context, cmd *commands.CommandEnvelope) (interface{}, error) {
		return clientCmdHandler.HandleUpdateClient(ctx, cmd)
	})
	cmdRegistry.Register("client.deactivate", func(ctx context.Context, cmd *commands.CommandEnvelope) (interface{}, error) {
		return nil, clientCmdHandler.HandleDeactivateClient(ctx, cmd)
	})
	cmdRegistry.Register("client.assign_credit_limit", func(ctx context.Context, cmd *commands.CommandEnvelope) (interface{}, error) {
		return nil, clientCmdHandler.HandleAssignCreditLimit(ctx, cmd)
	})
	cmdRegistry.Register("client.update_billing_info", func(ctx context.Context, cmd *commands.CommandEnvelope) (interface{}, error) {
		return nil, clientCmdHandler.HandleUpdateBillingInfo(ctx, cmd)
	})
	cmdRegistry.Register("client.merge", func(ctx context.Context, cmd *commands.CommandEnvelope) (interface{}, error) {
		return nil, clientCmdHandler.HandleMergeClients(ctx, cmd)
	})

	clientEventHandler := eventpkg.NewClientEventHandler(readModelStore, cache, log)

	eventHandlerRegistry := eventpkg.NewEventHandlerRegistry()
	eventHandlerRegistry.Register("ClientCreated", clientEventHandler.HandleClientCreated)
	eventHandlerRegistry.Register("ClientUpdated", clientEventHandler.HandleClientUpdated)
	eventHandlerRegistry.Register("ClientDeactivated", clientEventHandler.HandleClientDeactivated)
	eventHandlerRegistry.Register("CreditLimitAssigned", clientEventHandler.HandleCreditLimitAssigned)
	eventHandlerRegistry.Register("BillingInfoUpdated", clientEventHandler.HandleBillingInfoUpdated)
	eventHandlerRegistry.Register("ClientsMerged", clientEventHandler.HandleClientsMerged)

	healthChecker := health.NewHealthChecker(cfg, mongodb, redis, log)
	readinessChecker := health.NewReadinessChecker(log)
	livenessChecker := health.NewLivenessChecker()

	mux := http.NewServeMux()
	mux.Handle("/health", healthChecker.Handler())
	mux.Handle("/ready", readinessChecker.Handler())
	mux.Handle("/live", livenessChecker.Handler())

	mux.HandleFunc("/api/v1/commands", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var cmd commands.CommandEnvelope
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		ctx = logger.WithRequestID(ctx, generateRequestID())

		result, err := cmdRegistry.Handle(ctx, &cmd)
		if err != nil {
			log.Error("Command failed", "error", err, "command_type", cmd.Type)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      mux,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		log.Info("Starting server", "port", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server stopped")
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
