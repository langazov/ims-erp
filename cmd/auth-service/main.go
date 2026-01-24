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

	"github.com/ims-erp/system/internal/auth"
	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/health"
	"github.com/ims-erp/system/internal/messaging"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/ims-erp/system/pkg/tracer"
)

type RedisClientAdapter struct {
	cache *repository.Cache
}

func (r *RedisClientAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.cache.Set(ctx, key, value, expiration)
}

func (r *RedisClientAdapter) Get(ctx context.Context, key string) (string, error) {
	return r.cache.Get(ctx, key)
}

func (r *RedisClientAdapter) Del(ctx context.Context, keys ...string) error {
	return r.cache.Delete(ctx, keys...)
}

func main() {
	cfg, err := config.Load("", "auth-service")
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

	userStore := auth.NewUserRepository(repository.NewReadModelStore(mongodb, "users", log))
	cache := repository.NewCache(redis, "t:"+cfg.MongoDB.Database, log)
	redisClient := &RedisClientAdapter{cache: cache}
	tokenService := auth.NewTokenService(&cfg.Auth, redisClient, log)
	sessionService := auth.NewSessionService(redisClient, log, cfg.Auth.SessionTTL)
	rateLimiter := repository.NewRateLimiter(redis, log)

	authService := auth.NewAuthService(
		userStore,
		tokenService,
		sessionService,
		rateLimiter,
		&cfg.Auth,
		log,
	)

	healthChecker := health.NewHealthChecker(cfg, mongodb, redis, log)
	readinessChecker := health.NewReadinessChecker(log)
	livenessChecker := health.NewLivenessChecker()

	mux := http.NewServeMux()
	mux.Handle("/health", healthChecker.Handler())
	mux.Handle("/ready", readinessChecker.Handler())
	mux.Handle("/live", livenessChecker.Handler())

	mux.HandleFunc("/api/v1/auth/register", handleRegister(authService, log))
	mux.HandleFunc("/api/v1/auth/login", handleLogin(authService, log))
	mux.HandleFunc("/api/v1/auth/logout", handleLogout(authService, log))
	mux.HandleFunc("/api/v1/auth/refresh", handleRefresh(tokenService, log))
	mux.HandleFunc("/api/v1/auth/change-password", handleChangePassword(authService, log))
	mux.HandleFunc("/api/v1/auth/me", handleMe(authService, log))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      mux,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		log.Info("Starting auth service", "port", cfg.App.Port)
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

func handleRegister(authService *auth.AuthService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req auth.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		tenantID := r.URL.Query().Get("tenantId")
		if tenantID == "" {
			http.Error(w, "tenantId is required", http.StatusBadRequest)
			return
		}

		user, err := authService.Register(r.Context(), tenantID, "", &req)
		if err != nil {
			log.Error("Registration failed", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func handleLogin(authService *auth.AuthService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req auth.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		tenantID := r.URL.Query().Get("tenantId")
		if tenantID == "" {
			http.Error(w, "tenantId is required", http.StatusBadRequest)
			return
		}

		req.IPAddress = r.RemoteAddr
		req.UserAgent = r.UserAgent()

		response, err := authService.Login(r.Context(), tenantID, "", &req)
		if err != nil {
			log.Error("Login failed", "error", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func handleLogout(authService *auth.AuthService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.URL.Query().Get("userId")
		sessionID := r.URL.Query().Get("sessionId")

		if err := authService.Logout(r.Context(), userID, sessionID); err != nil {
			log.Error("Logout failed", "error", err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleRefresh(tokenService *auth.TokenService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		tokens, err := tokenService.RefreshTokens(r.Context(), req.RefreshToken)
		if err != nil {
			log.Error("Token refresh failed", "error", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokens)
	}
}

func handleChangePassword(authService *auth.AuthService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			UserID          string `json:"userId"`
			CurrentPassword string `json:"currentPassword"`
			NewPassword     string `json:"newPassword"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := authService.ChangePassword(r.Context(), req.UserID, req.CurrentPassword, req.NewPassword); err != nil {
			log.Error("Password change failed", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleMe(authService *auth.AuthService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, "userId is required", http.StatusBadRequest)
			return
		}

		user, err := authService.GetUser(r.Context(), userID)
		if err != nil {
			log.Error("Get user failed", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}
