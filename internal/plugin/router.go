package plugin

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ims-erp/system/internal/commands"
	"github.com/ims-erp/system/internal/events"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// EventRouter routes events to plugins that can handle them
type EventRouter struct {
	mu       sync.RWMutex
	handlers map[string][]string // event type -> plugin IDs
}

// NewEventRouter creates a new event router
func NewEventRouter() *EventRouter {
	return &EventRouter{
		handlers: make(map[string][]string),
	}
}

// Register registers a plugin's event handlers
func (r *EventRouter) Register(plugin Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if handler, ok := plugin.(EventHandler); ok {
		for _, eventType := range handler.EventTypes() {
			r.handlers[eventType] = append(r.handlers[eventType], plugin.Name())
		}
	}
}

// Unregister removes a plugin's event handlers
func (r *EventRouter) Unregister(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for eventType, pluginIDs := range r.handlers {
		filtered := make([]string, 0, len(pluginIDs))
		for _, id := range pluginIDs {
			if id != pluginID {
				filtered = append(filtered, id)
			}
		}
		r.handlers[eventType] = filtered
	}
}

// Handle routes an event to all registered handlers
func (r *EventRouter) Handle(ctx context.Context, event events.EventEnvelope) []error {
	r.mu.RLock()
	pluginIDs := r.handlers[event.Type]
	r.mu.RUnlock()

	var errors []error
	for _, pluginID := range pluginIDs {
		// In actual implementation, this would call the plugin's HandleEvent
		// For now, just collect any errors
		_ = pluginID
	}

	return errors
}

// CommandRouter routes commands to plugins that can handle them
type CommandRouter struct {
	mu       sync.RWMutex
	handlers map[string]string // command type -> plugin ID
}

// NewCommandRouter creates a new command router
func NewCommandRouter() *CommandRouter {
	return &CommandRouter{
		handlers: make(map[string]string),
	}
}

// Register registers a plugin's command handlers
func (r *CommandRouter) Register(plugin Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if handler, ok := plugin.(CommandHandler); ok {
		for _, commandType := range handler.CommandTypes() {
			r.handlers[commandType] = plugin.Name()
		}
	}
}

// Unregister removes a plugin's command handlers
func (r *CommandRouter) Unregister(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for cmdType, id := range r.handlers {
		if id == pluginID {
			delete(r.handlers, cmdType)
		}
	}
}

// Handle routes a command to the appropriate handler
func (r *CommandRouter) Handle(ctx context.Context, cmd interface{}) (interface{}, error) {
	// In actual implementation, this would:
	// 1. Extract command type from cmd
	// 2. Find the appropriate plugin
	// 3. Call the plugin's HandleCommand method
	return nil, nil
}

// HTTPRouter routes HTTP requests to plugins
type HTTPRouter struct {
	mu     sync.RWMutex
	routes map[string]Route // path -> route
}

// NewHTTPRouter creates a new HTTP router
func NewHTTPRouter() *HTTPRouter {
	return &HTTPRouter{
		routes: make(map[string]Route),
	}
}

// Register registers a plugin's HTTP routes
func (r *HTTPRouter) Register(plugin APIExtension) {
	r.mu.Lock()
	defer r.mu.Unlock()

	basePath := plugin.BasePath()
	for _, route := range plugin.GetRoutes() {
		fullPath := basePath + route.Path
		r.routes[fullPath] = route
	}
}

// Unregister removes a plugin's HTTP routes
func (r *HTTPRouter) Unregister(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// In actual implementation, we'd need to track which routes belong to which plugin
	// For now, this is a placeholder
}

// ServeHTTP implements the http.Handler interface
func (r *HTTPRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	route, exists := r.routes[req.URL.Path]
	r.mu.RUnlock()

	if !exists {
		http.NotFound(w, req)
		return
	}

	// Check method
	if route.Method != "" && route.Method != req.Method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Apply middleware
	handler := route.Handler
	for i := len(route.Middleware) - 1; i >= 0; i-- {
		handler = route.Middleware[i](handler)
	}

	handler(w, req)
}

// PluginSDKImpl implements the PluginSDK interface
type PluginSDKImpl struct {
	pluginID   string
	manifest   *PluginManifest
	eventBus   events.Publisher
	commandBus commands.CommandHandler
}

// NewPluginSDK creates a new PluginSDK implementation
func NewPluginSDK(pluginID string, manifest *PluginManifest, eventBus events.Publisher) PluginSDK {
	return &PluginSDKImpl{
		pluginID: pluginID,
		manifest: manifest,
		eventBus: eventBus,
	}
}

func (s *PluginSDKImpl) PublishEvent(ctx context.Context, event events.EventEnvelope) error {
	return s.eventBus.PublishEvent(ctx, &event)
}

func (s *PluginSDKImpl) PublishCommand(ctx context.Context, cmd commands.CommandEnvelope) error {
	// Implementation depends on command bus architecture
	return nil
}

func (s *PluginSDKImpl) RequestReply(ctx context.Context, subject string, data interface{}, timeout time.Duration) (interface{}, error) {
	// Implementation using NATS or similar
	return nil, nil
}

func (s *PluginSDKImpl) GetCollection(name string) *mongo.Collection {
	// Implementation depends on database connection
	return nil
}

func (s *PluginSDKImpl) GetDatabase() *mongo.Database {
	// Implementation depends on database connection
	return nil
}

func (s *PluginSDKImpl) GetCache() redis.Cmdable {
	// Implementation depends on cache connection
	return nil
}

func (s *PluginSDKImpl) Logger() Logger {
	// Return a logger implementation
	return &defaultLogger{pluginID: s.pluginID}
}

func (s *PluginSDKImpl) Metrics() MetricsCollector {
	// Return a metrics collector
	return &defaultMetricsCollector{pluginID: s.pluginID}
}

func (s *PluginSDKImpl) GetConfig(key string) interface{} {
	if s.manifest == nil {
		return nil
	}
	return s.manifest.Settings[key]
}

func (s *PluginSDKImpl) GetConfigString(key string) string {
	val := s.GetConfig(key)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

func (s *PluginSDKImpl) GetConfigInt(key string) int {
	val := s.GetConfig(key)
	if num, ok := val.(int); ok {
		return num
	}
	if num, ok := val.(float64); ok {
		return int(num)
	}
	return 0
}

func (s *PluginSDKImpl) GetConfigBool(key string) bool {
	val := s.GetConfig(key)
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

func (s *PluginSDKImpl) GetSecret(key string) (string, error) {
	if s.manifest == nil {
		return "", fmt.Errorf("no manifest available")
	}

	secret, exists := s.manifest.Secrets[key]
	if !exists {
		return "", fmt.Errorf("secret %s not found", key)
	}

	return secret, nil
}

func (s *PluginSDKImpl) GetPluginID() string {
	return s.pluginID
}

func (s *PluginSDKImpl) GetTenantID() string {
	if s.manifest == nil {
		return ""
	}
	return s.manifest.TenantID
}

// defaultLogger is a simple logger implementation
type defaultLogger struct {
	pluginID string
}

func (l *defaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	// Log debug message
}

func (l *defaultLogger) Info(msg string, keysAndValues ...interface{}) {
	// Log info message
}

func (l *defaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	// Log warn message
}

func (l *defaultLogger) Error(msg string, keysAndValues ...interface{}) {
	// Log error message
}

// defaultMetricsCollector is a simple metrics implementation
type defaultMetricsCollector struct {
	pluginID string
}

func (m *defaultMetricsCollector) Counter(name string, value float64, labels map[string]string) {
	// Record counter metric
}

func (m *defaultMetricsCollector) Gauge(name string, value float64, labels map[string]string) {
	// Record gauge metric
}

func (m *defaultMetricsCollector) Histogram(name string, value float64, labels map[string]string) {
	// Record histogram metric
}

func (m *defaultMetricsCollector) Timer(name string, duration time.Duration, labels map[string]string) {
	// Record timer metric
}
