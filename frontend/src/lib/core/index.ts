// src/lib/core/index.ts

import { createMessageBus, type MessageBus } from './message-bus';
import { createPluginRegistry, type PluginRegistry } from './plugin-registry';
import { createPluginLoader, type PluginLoader, type PluginManifestEntry } from './plugin-loader';
import { createRouteManager, type RouteManager } from './route-manager';
import { createStateManager, type StateManager } from './state-manager';
import { createEventEmitter, type EventEmitter } from './state-manager';
import type { CoreAPI, PluginInstance, PluginDefinition } from './types';

export interface Core {
  /** Core API exposed to plugins */
  api: CoreAPI;
  
  /** Plugin registry for managing plugins */
  registry: PluginRegistry;
  
  /** Plugin loader for loading plugins from various sources */
  loader: PluginLoader;
  
  /** Message bus for inter-plugin communication */
  messages: MessageBus;
  
  /** Route manager for dynamic routing */
  routes: RouteManager;
  
  /** State manager for shared state */
  state: StateManager;
  
  /** Event emitter for lifecycle events */
  events: EventEmitter;
  
  /** Initialize the core with plugin manifests */
  initialize(manifests: PluginManifestEntry[]): Promise<void>;
  
  /** Destroy the core and cleanup */
  destroy(): Promise<void>;
}

export function createCore(): Core {
  const messages = createMessageBus();
  const routes = createRouteManager();
  const state = createStateManager();
  const events = createEventEmitter();

  // Create CoreAPI that will be exposed to plugins
  const api: CoreAPI = {
    messages: messages,
    plugins: null!, // Will be set after registry creation
    routes: routes,
    state: state,
    events: events
  };

  const registry = createPluginRegistry(api);
  const loader = createPluginLoader();

  // Wire up plugins API
  api.plugins = {
    get: registry.get,
    getAll: registry.getAll,
    has: registry.has,
    
    async call<T>(pluginId: string, method: string, input?: unknown): Promise<T> {
      const plugin = registry.get(pluginId);
      if (!plugin) {
        throw new Error(`Plugin not found: ${pluginId}`);
      }
      if (!plugin.api?.methods[method]) {
        throw new Error(`Method not found: ${pluginId}.${method}`);
      }
      return plugin.api.methods[method].handler(input, {}) as T;
    },
    
    onPluginLoad: registry.onLoad,
    onPluginUnload: registry.onUnload
  };

  const core: Core = {
    api,
    registry,
    loader,
    messages,
    routes,
    state,
    events,

    async initialize(manifests: PluginManifestEntry[]): Promise<void> {
      console.log('[Core] Initializing with', manifests.length, 'plugin manifests');
      
      // Emit initialization event
      events.emit('core:initializing');
      
      // Load all plugins
      const plugins = await loader.loadFromManifests(manifests);
      
      // Register all plugins
      for (const plugin of plugins) {
        try {
          await registry.register(plugin);
          console.log(`[Core] Registered plugin: ${plugin.manifest.id}`);
        } catch (error) {
          console.error(`[Core] Failed to register plugin ${plugin.manifest.id}:`, error);
        }
      }
      
      // Emit ready event
      events.emit('core:ready', {
        plugins: registry.getAll().map(p => p.manifest.id)
      });
      
      console.log('[Core] Initialization complete');
    },

    async destroy(): Promise<void> {
      console.log('[Core] Destroying...');
      
      events.emit('core:destroying');
      
      // Unregister all plugins
      const plugins = registry.getAll();
      for (const plugin of plugins) {
        try {
          await registry.unregister(plugin.manifest.id);
        } catch (error) {
          console.error(`[Core] Failed to unregister plugin ${plugin.manifest.id}:`, error);
        }
      }
      
      // Clear state
      state.clear();
      messages.clearHistory();
      events.removeAllListeners();
      
      events.emit('core:destroyed');
      console.log('[Core] Destroyed');
    }
  };

  return core;
}

// Re-export types
export * from './types';
export type { PluginManifestEntry, PluginSource } from './plugin-loader';
export { createBuiltinEntries, createPluginEntry } from './plugin-loader';
export type { MessageBus } from './message-bus';
export type { PluginRegistry } from './plugin-registry';
export type { PluginLoader } from './plugin-loader';
export type { RouteManager, RegisteredRoute } from './route-manager';
export type { StateManager } from './state-manager';
export type { EventEmitter } from './state-manager';

// Singleton instance (optional, for simpler usage)
let coreInstance: Core | null = null;

export function getCore(): Core {
  if (!coreInstance) {
    coreInstance = createCore();
  }
  return coreInstance;
}

export function resetCore(): void {
  if (coreInstance) {
    coreInstance.destroy();
    coreInstance = null;
  }
}
