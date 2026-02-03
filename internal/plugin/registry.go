package plugin

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ims-erp/system/internal/events"
)

// Registry manages plugin registration and lifecycle
type Registry struct {
	mu            sync.RWMutex
	plugins       map[string]Plugin
	statuses      map[string]*PluginStatus
	manifests     map[string]*PluginManifest
	eventRouter   *EventRouter
	commandRouter *CommandRouter
	httpRouter    *HTTPRouter
	enabled       map[string]bool
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins:       make(map[string]Plugin),
		statuses:      make(map[string]*PluginStatus),
		manifests:     make(map[string]*PluginManifest),
		eventRouter:   NewEventRouter(),
		commandRouter: NewCommandRouter(),
		httpRouter:    NewHTTPRouter(),
		enabled:       make(map[string]bool),
	}
}

// Register registers a plugin with the registry
func (r *Registry) Register(plugin Plugin, manifest *PluginManifest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := generatePluginID(plugin.Name())

	// Check if already registered
	if _, exists := r.plugins[id]; exists {
		return fmt.Errorf("plugin %s already registered", plugin.Name())
	}

	// Store plugin
	r.plugins[id] = plugin
	r.manifests[id] = manifest
	r.enabled[id] = true

	// Initialize status
	r.statuses[id] = &PluginStatus{
		ID:      id,
		Name:    plugin.Name(),
		Version: plugin.Version(),
		State:   PluginStateRegistered,
		Health:  HealthStatus{Status: HealthStatusHealthy, CheckedAt: time.Now().UTC()},
	}

	// Register with routers
	r.eventRouter.Register(plugin)
	r.commandRouter.Register(plugin)

	if apiExt, ok := plugin.(APIExtension); ok {
		r.httpRouter.Register(apiExt)
	}

	return nil
}

// Unregister removes a plugin from the registry
func (r *Registry) Unregister(pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	plugin, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// Stop the plugin if running
	if r.statuses[pluginID].State == PluginStateRunning {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := plugin.Stop(ctx); err != nil {
			// Log error but continue
		}

		r.statuses[pluginID].State = PluginStateStopped
		r.statuses[pluginID].StoppedAt = &[]time.Time{time.Now().UTC()}[0]
	}

	// Unregister from routers
	r.eventRouter.Unregister(pluginID)
	r.commandRouter.Unregister(pluginID)
	r.httpRouter.Unregister(pluginID)

	// Remove from maps
	delete(r.plugins, pluginID)
	delete(r.statuses, pluginID)
	delete(r.manifests, pluginID)
	delete(r.enabled, pluginID)

	return nil
}

// Get retrieves a plugin by ID
func (r *Registry) Get(pluginID string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[pluginID]
	return plugin, exists
}

// GetByName retrieves a plugin by name
func (r *Registry) GetByName(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, manifest := range r.manifests {
		if manifest.Name == name {
			return r.plugins[id], true
		}
	}

	return nil, false
}

// GetStatus retrieves the status of a plugin
func (r *Registry) GetStatus(pluginID string) (*PluginStatus, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status, exists := r.statuses[pluginID]
	return status, exists
}

// GetAllStatuses retrieves all plugin statuses
func (r *Registry) GetAllStatuses() []*PluginStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	statuses := make([]*PluginStatus, 0, len(r.statuses))
	for _, status := range r.statuses {
		statuses = append(statuses, status)
	}

	return statuses
}

// List returns all registered plugins
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// IsEnabled checks if a plugin is enabled
func (r *Registry) IsEnabled(pluginID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.enabled[pluginID]
}

// Enable enables a plugin
func (r *Registry) Enable(pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[pluginID]; !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	r.enabled[pluginID] = true
	return nil
}

// Disable disables a plugin
func (r *Registry) Disable(pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[pluginID]; !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	r.enabled[pluginID] = false
	return nil
}

// Start initializes and starts all registered plugins
func (r *Registry) Start(ctx context.Context, sdkFactory SDKFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, plugin := range r.plugins {
		if !r.enabled[id] {
			continue
		}

		r.statuses[id].State = PluginStateStarting

		// Create SDK for plugin
		sdk := sdkFactory.Create(id, r.manifests[id])

		// Initialize
		if err := plugin.Initialize(ctx, sdk); err != nil {
			r.statuses[id].State = PluginStateFailed
			r.statuses[id].Error = fmt.Sprintf("initialize failed: %v", err)
			continue
		}

		r.statuses[id].State = PluginStateInitialized

		// Start
		if err := plugin.Start(ctx); err != nil {
			r.statuses[id].State = PluginStateFailed
			r.statuses[id].Error = fmt.Sprintf("start failed: %v", err)
			continue
		}

		now := time.Now().UTC()
		r.statuses[id].State = PluginStateRunning
		r.statuses[id].StartedAt = &now
	}

	return nil
}

// Stop stops all running plugins
func (r *Registry) Stop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, plugin := range r.plugins {
		if r.statuses[id].State != PluginStateRunning {
			continue
		}

		r.statuses[id].State = PluginStateStopping

		if err := plugin.Stop(ctx); err != nil {
			r.statuses[id].Error = fmt.Sprintf("stop failed: %v", err)
		}

		now := time.Now().UTC()
		r.statuses[id].State = PluginStateStopped
		r.statuses[id].StoppedAt = &now
	}

	return nil
}

// HealthCheck runs health checks on all running plugins
func (r *Registry) HealthCheck(ctx context.Context) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, plugin := range r.plugins {
		if r.statuses[id].State != PluginStateRunning {
			continue
		}

		health := plugin.HealthCheck(ctx)
		r.statuses[id].Health = health
	}
}

// HandleEvent routes an event to appropriate plugins
func (r *Registry) HandleEvent(ctx context.Context, event events.EventEnvelope) []error {
	return r.eventRouter.Handle(ctx, event)
}

// HandleCommand routes a command to appropriate plugins
func (r *Registry) HandleCommand(ctx context.Context, cmd interface{}) (interface{}, error) {
	return r.commandRouter.Handle(ctx, cmd)
}

// GetHTTPHandler returns the HTTP handler for plugin routes
func (r *Registry) GetHTTPHandler() http.Handler {
	return r.httpRouter
}

// SDKFactory creates PluginSDK instances
type SDKFactory interface {
	Create(pluginID string, manifest *PluginManifest) PluginSDK
}
