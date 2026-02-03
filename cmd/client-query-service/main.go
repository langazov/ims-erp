package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/internal/health"
	"github.com/ims-erp/system/internal/messaging"
	"github.com/ims-erp/system/internal/queries"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/ims-erp/system/pkg/tracer"
	"github.com/nats-io/nats.go"
)

var allowedOrigins = []string{
	"http://localhost:5173",
	"http://localhost:5178",
	"http://localhost:5174",
	"http://localhost:5175",
	"http://localhost:5176",
	"http://localhost:5177",
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Debug logging
		fmt.Printf("[CORS] Method: %s, Path: %s, Origin: %s\n", r.Method, r.URL.Path, r.Header.Get("Origin"))

		origin := r.Header.Get("Origin")

		isAllowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			fmt.Printf("[CORS] Returning 204 for OPTIONS\n")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func optionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func main() {
	cfg, err := config.Load("", "client-query-service")
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

	subscriber, err := messaging.NewSubscriber(natsConfig, log)
	if err != nil {
		log.Error("Failed to create NATS subscriber", "error", err)
		os.Exit(1)
	}
	defer subscriber.Close()
	log.Info("Connected to NATS")

	readModelStore := repository.NewReadModelStore(mongodb, "client_read", log)
	cache := repository.NewCache(redis, "t:"+cfg.MongoDB.Database, log)

	clientQueryHandler := queries.NewClientQueryHandler(readModelStore, cache, log)

	eventHandler := events.NewClientEventHandler(readModelStore, cache, log)

	eventHandlerRegistry := events.NewEventHandlerRegistry()
	eventHandlerRegistry.Register("ClientCreated", eventHandler.HandleClientCreated)
	eventHandlerRegistry.Register("ClientUpdated", eventHandler.HandleClientUpdated)
	eventHandlerRegistry.Register("ClientDeactivated", eventHandler.HandleClientDeactivated)
	eventHandlerRegistry.Register("CreditLimitAssigned", eventHandler.HandleCreditLimitAssigned)
	eventHandlerRegistry.Register("BillingInfoUpdated", eventHandler.HandleBillingInfoUpdated)
	eventHandlerRegistry.Register("ClientsMerged", eventHandler.HandleClientsMerged)

	go func() {
		subjects := []string{
			natsConfig.StreamPrefix + "client.>",
		}
		for _, subject := range subjects {
			if err := subscriber.Subscribe(subject, createEventHandler(eventHandlerRegistry, log)); err != nil {
				log.Error("Failed to subscribe", "error", err, "subject", subject)
			}
		}
	}()

	healthChecker := health.NewHealthChecker(cfg, mongodb, redis, log)
	readinessChecker := health.NewReadinessChecker(log)
	livenessChecker := health.NewLivenessChecker()

	mux := http.NewServeMux()
	mux.Handle("/health", healthChecker.Handler())
	mux.Handle("/ready", readinessChecker.Handler())
	mux.Handle("/live", livenessChecker.Handler())

	mux.HandleFunc("/api/v1/clients", handleListClients(clientQueryHandler, log))
	mux.HandleFunc("/api/v1/clients/search", handleSearchClients(clientQueryHandler, log))
	mux.HandleFunc("/api/v1/clients/id/", handleGetClient(clientQueryHandler, log))
	mux.HandleFunc("/api/v1/clients/detail/", handleGetClientDetail(clientQueryHandler, log))
	mux.HandleFunc("/api/v1/clients/credit/", handleGetClientCreditStatus(clientQueryHandler, log))

	handler := corsMiddleware(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      handler,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		log.Info("Starting client-query-service", "port", cfg.App.Port)
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

func createEventHandler(registry *events.EventHandlerRegistry, log *logger.Logger) func(msg *nats.Msg) {
	return func(msg *nats.Msg) {
		var event events.EventEnvelope
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Error("Failed to unmarshal event", "error", err)
			return
		}

		if err := registry.Handle(context.Background(), &event); err != nil {
			log.Error("Failed to handle event", "error", err, "event_type", event.Type)
		}
	}
}

func handleListClients(handler *queries.ClientQueryHandler, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tenantID := r.URL.Query().Get("tenantId")
		if tenantID == "" {
			http.Error(w, "tenantId is required", http.StatusBadRequest)
			return
		}

		page := parseInt(r.URL.Query().Get("page"), 1)
		pageSize := parseInt(r.URL.Query().Get("pageSize"), 20)
		search := r.URL.Query().Get("search")
		status := r.URL.Query().Get("status")

		query := &queries.ListClientsQuery{
			TenantID:  tenantID,
			Page:      page,
			PageSize:  pageSize,
			Search:    search,
			Status:    status,
			SortBy:    "name",
			SortOrder: "asc",
		}

		result, err := handler.ListClients(r.Context(), query)
		if err != nil {
			log.Error("Failed to list clients", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func handleSearchClients(handler *queries.ClientQueryHandler, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tenantID := r.URL.Query().Get("tenantId")
		if tenantID == "" {
			http.Error(w, "tenantId is required", http.StatusBadRequest)
			return
		}

		term := r.URL.Query().Get("q")
		if term == "" {
			http.Error(w, "search term is required", http.StatusBadRequest)
			return
		}

		limit := parseInt(r.URL.Query().Get("limit"), 10)

		query := &queries.SearchClientsQuery{
			TenantID: tenantID,
			Term:     term,
			Limit:    limit,
		}

		clients, err := handler.SearchClients(r.Context(), query)
		if err != nil {
			log.Error("Failed to search clients", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clients)
	}
}

func handleGetClient(handler *queries.ClientQueryHandler, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tenantID := r.URL.Query().Get("tenantId")
		clientID := r.URL.Query().Get("clientId")
		if tenantID == "" || clientID == "" {
			http.Error(w, "tenantId and clientId are required", http.StatusBadRequest)
			return
		}

		query := &queries.GetClientByIDQuery{
			ClientID: clientID,
			TenantID: tenantID,
		}

		client, err := handler.GetClientByID(r.Context(), query)
		if err != nil {
			log.Error("Failed to get client", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if client == nil {
			http.Error(w, "client not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(client)
	}
}

func handleGetClientDetail(handler *queries.ClientQueryHandler, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tenantID := r.URL.Query().Get("tenantId")
		clientID := r.URL.Query().Get("clientId")
		if tenantID == "" || clientID == "" {
			http.Error(w, "tenantId and clientId are required", http.StatusBadRequest)
			return
		}

		query := &queries.GetClientDetailQuery{
			ClientID: clientID,
			TenantID: tenantID,
		}

		client, err := handler.GetClientDetail(r.Context(), query)
		if err != nil {
			log.Error("Failed to get client detail", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if client == nil {
			http.Error(w, "client not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(client)
	}
}

func handleGetClientCreditStatus(handler *queries.ClientQueryHandler, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tenantID := r.URL.Query().Get("tenantId")
		clientID := r.URL.Query().Get("clientId")
		if tenantID == "" || clientID == "" {
			http.Error(w, "tenantId and clientId are required", http.StatusBadRequest)
			return
		}

		query := &queries.GetClientCreditStatusQuery{
			ClientID: clientID,
			TenantID: tenantID,
		}

		status, err := handler.GetClientCreditStatus(r.Context(), query)
		if err != nil {
			log.Error("Failed to get client credit status", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if status == nil {
			http.Error(w, "client not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
