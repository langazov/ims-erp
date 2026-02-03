package plugin

import (
	"context"
	"net/http"
	"time"

	"github.com/ims-erp/system/internal/commands"
	"github.com/ims-erp/system/internal/events"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// Plugin is the interface that all plugins must implement
type Plugin interface {
	// Metadata
	Name() string
	Version() string
	Description() string
	Author() string

	// Lifecycle
	Initialize(ctx context.Context, sdk PluginSDK) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	HealthCheck(ctx context.Context) HealthStatus

	// Event handling
	CanHandleEvent(eventType string) bool
	HandleEvent(ctx context.Context, event events.EventEnvelope) error

	// Command handling
	CanHandleCommand(commandType string) bool
	HandleCommand(ctx context.Context, cmd commands.CommandEnvelope) (interface{}, error)

	// HTTP handling
	GetRoutes() []Route
}

// PluginSDK provides plugins with access to system resources
type PluginSDK interface {
	// Event publishing
	PublishEvent(ctx context.Context, event events.EventEnvelope) error

	// Command publishing
	PublishCommand(ctx context.Context, cmd commands.CommandEnvelope) error

	// Request-reply pattern
	RequestReply(ctx context.Context, subject string, data interface{}, timeout time.Duration) (interface{}, error)

	// Database access
	GetCollection(name string) *mongo.Collection
	GetDatabase() *mongo.Database

	// Cache access
	GetCache() redis.Cmdable

	// Logging
	Logger() Logger

	// Metrics
	Metrics() MetricsCollector

	// Configuration
	GetConfig(key string) interface{}
	GetConfigString(key string) string
	GetConfigInt(key string) int
	GetConfigBool(key string) bool
	GetSecret(key string) (string, error)

	// Plugin info
	GetPluginID() string
	GetTenantID() string
}

// Logger provides logging capabilities to plugins
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// MetricsCollector provides metrics capabilities to plugins
type MetricsCollector interface {
	Counter(name string, value float64, labels map[string]string)
	Gauge(name string, value float64, labels map[string]string)
	Histogram(name string, value float64, labels map[string]string)
	Timer(name string, duration time.Duration, labels map[string]string)
}

// HealthStatus represents the health status of a plugin
type HealthStatus struct {
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	CheckedAt time.Time `json:"checkedAt"`
}

const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
)

// Route defines an HTTP route exposed by a plugin
type Route struct {
	Path       string
	Method     string
	Handler    http.HandlerFunc
	Middleware []Middleware
}

// Middleware is a function that wraps an HTTP handler
type Middleware func(http.HandlerFunc) http.HandlerFunc

// EventHandler is a specialized plugin that handles events
type EventHandler interface {
	Plugin
	EventTypes() []string
}

// CommandHandler is a specialized plugin that handles commands
type CommandHandler interface {
	Plugin
	CommandTypes() []string
}

// APIExtension is a specialized plugin that extends the API
type APIExtension interface {
	Plugin
	BasePath() string
}

// ScheduledTask is a specialized plugin that runs on a schedule
type ScheduledTask interface {
	Plugin
	Schedule() string
	TimeZone() string
	Run(ctx context.Context) error
}

// PluginConfig holds configuration for a plugin
type PluginConfig struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Version  string                 `json:"version"`
	TenantID string                 `json:"tenantId"`
	Enabled  bool                   `json:"enabled"`
	Settings map[string]interface{} `json:"settings"`
	Secrets  map[string]string      `json:"secrets"`
}

// PluginMetadata holds metadata about a plugin
type PluginMetadata struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	EntryPoint   string                 `json:"entryPoint"`
	Permissions  []Permission           `json:"permissions"`
	Dependencies []Dependency           `json:"dependencies"`
	Schedule     string                 `json:"schedule,omitempty"`
	Routes       []RouteDefinition      `json:"routes,omitempty"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
	Secrets      map[string]string      `json:"secrets,omitempty"`
	TenantID     string                 `json:"tenantId,omitempty"`
}

// PluginManifest is an alias for PluginMetadata
type PluginManifest = PluginMetadata

// Permission defines what a plugin is allowed to do
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// Dependency defines a plugin dependency
type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// RouteDefinition defines a route in the plugin manifest
type RouteDefinition struct {
	Path       string   `json:"path"`
	Method     string   `json:"method"`
	Handler    string   `json:"handler"`
	Middleware []string `json:"middleware,omitempty"`
}

// PluginStatus represents the runtime status of a plugin
type PluginStatus struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Version   string       `json:"version"`
	State     PluginState  `json:"state"`
	Health    HealthStatus `json:"health"`
	StartedAt *time.Time   `json:"startedAt,omitempty"`
	StoppedAt *time.Time   `json:"stoppedAt,omitempty"`
	Error     string       `json:"error,omitempty"`
}

// PluginState represents the lifecycle state of a plugin
type PluginState string

const (
	PluginStateRegistered  PluginState = "registered"
	PluginStateInitialized PluginState = "initialized"
	PluginStateStarting    PluginState = "starting"
	PluginStateRunning     PluginState = "running"
	PluginStateStopping    PluginState = "stopping"
	PluginStateStopped     PluginState = "stopped"
	PluginStateFailed      PluginState = "failed"
)

// BasePlugin provides a base implementation of the Plugin interface
type BasePlugin struct {
	id           string
	name         string
	version      string
	description  string
	author       string
	sdk          PluginSDK
	eventTypes   []string
	commandTypes []string
	routes       []Route
}

func NewBasePlugin(name, version, description, author string) *BasePlugin {
	return &BasePlugin{
		id:           generatePluginID(name),
		name:         name,
		version:      version,
		description:  description,
		author:       author,
		eventTypes:   []string{},
		commandTypes: []string{},
		routes:       []Route{},
	}
}

func (p *BasePlugin) Name() string {
	return p.name
}

func (p *BasePlugin) Version() string {
	return p.version
}

func (p *BasePlugin) Description() string {
	return p.description
}

func (p *BasePlugin) Author() string {
	return p.author
}

func (p *BasePlugin) Initialize(ctx context.Context, sdk PluginSDK) error {
	p.sdk = sdk
	return nil
}

func (p *BasePlugin) Start(ctx context.Context) error {
	return nil
}

func (p *BasePlugin) Stop(ctx context.Context) error {
	return nil
}

func (p *BasePlugin) HealthCheck(ctx context.Context) HealthStatus {
	return HealthStatus{
		Status:    HealthStatusHealthy,
		CheckedAt: time.Now().UTC(),
	}
}

func (p *BasePlugin) CanHandleEvent(eventType string) bool {
	for _, et := range p.eventTypes {
		if et == eventType {
			return true
		}
	}
	return false
}

func (p *BasePlugin) HandleEvent(ctx context.Context, event events.EventEnvelope) error {
	return nil
}

func (p *BasePlugin) CanHandleCommand(commandType string) bool {
	for _, ct := range p.commandTypes {
		if ct == commandType {
			return true
		}
	}
	return false
}

func (p *BasePlugin) HandleCommand(ctx context.Context, cmd commands.CommandEnvelope) (interface{}, error) {
	return nil, nil
}

func (p *BasePlugin) GetRoutes() []Route {
	return p.routes
}

func (p *BasePlugin) SetEventTypes(types []string) {
	p.eventTypes = types
}

func (p *BasePlugin) SetCommandTypes(types []string) {
	p.commandTypes = types
}

func (p *BasePlugin) AddRoute(route Route) {
	p.routes = append(p.routes, route)
}

func (p *BasePlugin) GetSDK() PluginSDK {
	return p.sdk
}

func generatePluginID(name string) string {
	return "plugin-" + name + "-" + time.Now().Format("20060102150405")
}
