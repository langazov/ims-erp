# SvelteKit Modular Plugin System Architecture

## Overview

A production-ready, type-safe plugin architecture for SvelteKit applications featuring dynamic module loading, a centralized message bus, and runtime route injection.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SvelteKit Application                          │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                           CORE SYSTEM                                │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │   │
│  │  │ Plugin       │  │ Message      │  │ Route                    │  │   │
│  │  │ Registry     │◄─┤ Bus          │◄─┤ Manager                  │  │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────────────────────────┘  │   │
│  │         │                 │                                         │   │
│  │         ▼                 ▼                                         │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │   │
│  │  │ Plugin       │  │ State        │  │ API                      │  │   │
│  │  │ Loader       │  │ Manager      │  │ Gateway                  │  │   │
│  │  └──────────────┘  └──────────────┘  └──────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        PLUGIN LAYER                                  │   │
│  │                                                                       │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌───────────┐   │   │
│  │  │ Dashboard   │  │ Analytics   │  │ Settings    │  │ Custom    │   │   │
│  │  │ Plugin      │  │ Plugin      │  │ Plugin      │  │ Plugin... │   │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └───────────┘   │   │
│  │                                                                       │   │
│  │  Each plugin exports:                                                │   │
│  │  • manifest (metadata, dependencies, permissions)                    │   │
│  │  • api (exposed functions for other plugins)                        │   │
│  │  • routes (SvelteKit route components)                              │   │
│  │  • messages (message handlers & schemas)                            │   │
│  │  • stores (reactive state)                                          │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
src/
├── lib/
│   ├── core/                          # Core plugin system
│   │   ├── index.ts                   # Public API exports
│   │   ├── types.ts                   # Type definitions
│   │   ├── plugin-registry.ts         # Plugin registration & lifecycle
│   │   ├── plugin-loader.ts           # Dynamic import & validation
│   │   ├── message-bus.ts             # Inter-plugin messaging
│   │   ├── route-manager.ts           # Dynamic route injection
│   │   ├── api-gateway.ts             # Plugin API exposure
│   │   ├── state-manager.ts           # Shared state management
│   │   └── permissions.ts             # Plugin sandboxing
│   │
│   ├── plugins/                       # Built-in plugins
│   │   ├── dashboard/
│   │   │   ├── index.ts               # Plugin entry point
│   │   │   ├── manifest.ts            # Plugin metadata
│   │   │   ├── api.ts                 # Exposed API
│   │   │   ├── messages.ts            # Message handlers
│   │   │   ├── stores.ts              # Plugin state
│   │   │   └── routes/
│   │   │       ├── +page.svelte
│   │   │       └── +layout.svelte
│   │   │
│   │   └── [plugin-name]/             # Other plugins follow same structure
│   │
│   └── shared/                        # Shared utilities
│       ├── components/
│       ├── utils/
│       └── schemas/
│
├── routes/
│   ├── +layout.svelte                 # Root layout with plugin context
│   ├── +layout.server.ts              # Plugin manifest loading
│   ├── +page.svelte                   # Main app shell
│   ├── api/
│   │   └── plugins/
│   │       ├── manifest/+server.ts    # GET plugin manifests
│   │       ├── [plugin]/+server.ts    # Plugin-specific API proxy
│   │       └── messages/+server.ts    # WebSocket/SSE for messaging
│   │
│   └── [[...catchall]]/               # Dynamic plugin routes
│       └── +page.svelte
│
├── plugins/                           # External plugins directory
│   └── [external-plugin]/
│
└── app.d.ts                           # Global type augmentation
```

---

## Core Type Definitions

```typescript
// src/lib/core/types.ts

import type { ComponentType, SvelteComponent } from 'svelte';
import type { Readable, Writable } from 'svelte/store';

// ============================================================================
// Plugin Manifest
// ============================================================================

export interface PluginManifest {
  /** Unique plugin identifier (kebab-case) */
  id: string;
  
  /** Display name */
  name: string;
  
  /** Semantic version */
  version: string;
  
  /** Plugin description */
  description: string;
  
  /** Plugin author */
  author?: string;
  
  /** Plugin homepage/docs URL */
  homepage?: string;
  
  /** Plugin icon (URL or component) */
  icon?: string | ComponentType<SvelteComponent>;
  
  /** Plugin dependencies */
  dependencies?: PluginDependency[];
  
  /** Required permissions */
  permissions?: PluginPermission[];
  
  /** Plugin lifecycle hooks */
  lifecycle?: PluginLifecycle;
  
  /** Plugin configuration schema */
  configSchema?: ConfigSchema;
  
  /** Plugin priority (load order) */
  priority?: number;
  
  /** Whether plugin is enabled by default */
  enabled?: boolean;
}

export interface PluginDependency {
  pluginId: string;
  version: string;
  optional?: boolean;
}

export type PluginPermission = 
  | 'storage:read'
  | 'storage:write'
  | 'network:fetch'
  | 'user:read'
  | 'user:write'
  | 'plugins:communicate'
  | 'routes:register'
  | 'api:expose'
  | string; // Custom permissions

export interface PluginLifecycle {
  /** Called before plugin is loaded */
  onBeforeLoad?: () => Promise<void> | void;
  
  /** Called after plugin is loaded */
  onLoad?: () => Promise<void> | void;
  
  /** Called when plugin is enabled */
  onEnable?: () => Promise<void> | void;
  
  /** Called when plugin is disabled */
  onDisable?: () => Promise<void> | void;
  
  /** Called before plugin is unloaded */
  onUnload?: () => Promise<void> | void;
  
  /** Called on plugin error */
  onError?: (error: Error) => void;
}

export interface ConfigSchema {
  type: 'object';
  properties: Record<string, ConfigProperty>;
  required?: string[];
}

export interface ConfigProperty {
  type: 'string' | 'number' | 'boolean' | 'array' | 'object';
  title?: string;
  description?: string;
  default?: unknown;
  enum?: unknown[];
}

// ============================================================================
// Plugin API
// ============================================================================

export interface PluginAPI {
  /** API version */
  version: string;
  
  /** Exposed methods */
  methods: Record<string, PluginMethod>;
  
  /** API documentation */
  docs?: APIDocumentation;
}

export interface PluginMethod<TInput = unknown, TOutput = unknown> {
  /** Method handler */
  handler: (input: TInput, context: MethodContext) => Promise<TOutput> | TOutput;
  
  /** Input schema for validation */
  inputSchema?: JSONSchema;
  
  /** Output schema for validation */
  outputSchema?: JSONSchema;
  
  /** Method description */
  description?: string;
  
  /** Required permissions to call this method */
  permissions?: PluginPermission[];
}

export interface MethodContext {
  /** Calling plugin ID (if from another plugin) */
  callerId?: string;
  
  /** Current user context */
  user?: UserContext;
  
  /** Abort signal for cancellation */
  signal?: AbortSignal;
}

export interface APIDocumentation {
  title: string;
  description: string;
  methods: Record<string, MethodDocumentation>;
}

export interface MethodDocumentation {
  summary: string;
  description?: string;
  examples?: APIExample[];
}

export interface APIExample {
  title: string;
  input: unknown;
  output: unknown;
}

export interface JSONSchema {
  type: string;
  properties?: Record<string, JSONSchema>;
  items?: JSONSchema;
  required?: string[];
  [key: string]: unknown;
}

// ============================================================================
// Plugin Routes
// ============================================================================

export interface PluginRoutes {
  /** Base path for plugin routes */
  basePath: string;
  
  /** Route definitions */
  routes: RouteDefinition[];
  
  /** Navigation items */
  navigation?: NavigationItem[];
}

export interface RouteDefinition {
  /** Route path (relative to basePath) */
  path: string;
  
  /** Route component */
  component: ComponentType<SvelteComponent>;
  
  /** Layout component */
  layout?: ComponentType<SvelteComponent>;
  
  /** Route metadata */
  meta?: RouteMeta;
  
  /** Child routes */
  children?: RouteDefinition[];
  
  /** Route guards */
  guards?: RouteGuard[];
}

export interface RouteMeta {
  title?: string;
  description?: string;
  icon?: string;
  order?: number;
  hidden?: boolean;
  permissions?: PluginPermission[];
}

export interface RouteGuard {
  canActivate: (context: RouteContext) => boolean | Promise<boolean>;
  redirectTo?: string;
}

export interface RouteContext {
  params: Record<string, string>;
  searchParams: URLSearchParams;
  user?: UserContext;
}

export interface NavigationItem {
  id: string;
  label: string;
  path: string;
  icon?: string | ComponentType<SvelteComponent>;
  order?: number;
  children?: NavigationItem[];
  badge?: string | number;
}

// ============================================================================
// Plugin Messaging
// ============================================================================

export interface PluginMessages {
  /** Message handlers */
  handlers: Record<string, MessageHandler>;
  
  /** Message schemas */
  schemas: Record<string, MessageSchema>;
  
  /** Subscriptions to other plugin messages */
  subscriptions?: MessageSubscription[];
}

export interface MessageSchema {
  /** Message type identifier */
  type: string;
  
  /** Payload schema */
  payload: JSONSchema;
  
  /** Response schema (for request/response pattern) */
  response?: JSONSchema;
  
  /** Message description */
  description?: string;
  
  /** Whether this message expects a response */
  expectsResponse?: boolean;
}

export interface MessageHandler<TPayload = unknown, TResponse = void> {
  /** Handler function */
  handle: (message: Message<TPayload>) => Promise<TResponse> | TResponse;
  
  /** Filter function for selective handling */
  filter?: (message: Message<TPayload>) => boolean;
  
  /** Priority for handler execution order */
  priority?: number;
}

export interface Message<TPayload = unknown> {
  /** Unique message ID */
  id: string;
  
  /** Message type */
  type: string;
  
  /** Source plugin ID */
  source: string;
  
  /** Target plugin ID (optional, for direct messages) */
  target?: string;
  
  /** Message payload */
  payload: TPayload;
  
  /** Message timestamp */
  timestamp: number;
  
  /** Correlation ID for request/response */
  correlationId?: string;
  
  /** Message metadata */
  meta?: Record<string, unknown>;
}

export interface MessageSubscription {
  /** Message type pattern (supports wildcards) */
  type: string;
  
  /** Source plugin filter */
  source?: string | string[];
  
  /** Handler function */
  handler: MessageHandler;
}

// ============================================================================
// Plugin Stores
// ============================================================================

export interface PluginStores {
  /** Exposed readable stores */
  readable?: Record<string, Readable<unknown>>;
  
  /** Exposed writable stores */
  writable?: Record<string, Writable<unknown>>;
  
  /** Store schemas for type safety */
  schemas?: Record<string, JSONSchema>;
}

// ============================================================================
// Plugin Definition (Complete Plugin Export)
// ============================================================================

export interface PluginDefinition {
  manifest: PluginManifest;
  api?: PluginAPI;
  routes?: PluginRoutes;
  messages?: PluginMessages;
  stores?: PluginStores;
  
  /** Plugin setup function */
  setup?: (context: PluginContext) => Promise<void> | void;
  
  /** Plugin teardown function */
  teardown?: () => Promise<void> | void;
}

export interface PluginContext {
  /** Core API for interacting with the system */
  core: CoreAPI;
  
  /** Plugin's own config */
  config: Record<string, unknown>;
  
  /** Logger scoped to plugin */
  logger: Logger;
  
  /** Storage scoped to plugin */
  storage: PluginStorage;
}

// ============================================================================
// Core API (Exposed to Plugins)
// ============================================================================

export interface CoreAPI {
  /** Message bus for inter-plugin communication */
  messages: MessageBusAPI;
  
  /** Plugin registry for querying other plugins */
  plugins: PluginRegistryAPI;
  
  /** Route manager for dynamic routes */
  routes: RouteManagerAPI;
  
  /** State manager for shared state */
  state: StateManagerAPI;
  
  /** Event emitter for lifecycle events */
  events: EventEmitterAPI;
}

export interface MessageBusAPI {
  /** Send a message */
  send<T>(type: string, payload: T, options?: SendOptions): void;
  
  /** Send a message and wait for response */
  request<TReq, TRes>(type: string, payload: TReq, options?: RequestOptions): Promise<TRes>;
  
  /** Subscribe to messages */
  subscribe(type: string, handler: MessageHandler, options?: SubscribeOptions): Unsubscribe;
  
  /** Subscribe to messages once */
  once<T>(type: string, options?: SubscribeOptions): Promise<Message<T>>;
}

export interface PluginRegistryAPI {
  /** Get plugin by ID */
  get(id: string): PluginInstance | undefined;
  
  /** Get all loaded plugins */
  getAll(): PluginInstance[];
  
  /** Check if plugin is loaded */
  has(id: string): boolean;
  
  /** Call another plugin's API method */
  call<T>(pluginId: string, method: string, input?: unknown): Promise<T>;
  
  /** Subscribe to plugin lifecycle events */
  onPluginLoad(handler: (plugin: PluginInstance) => void): Unsubscribe;
  onPluginUnload(handler: (pluginId: string) => void): Unsubscribe;
}

export interface RouteManagerAPI {
  /** Register routes */
  register(routes: RouteDefinition[]): void;
  
  /** Unregister routes */
  unregister(paths: string[]): void;
  
  /** Navigate programmatically */
  navigate(path: string, options?: NavigateOptions): Promise<void>;
  
  /** Get current route */
  getCurrentRoute(): RouteInfo;
}

export interface StateManagerAPI {
  /** Get shared state */
  get<T>(key: string): T | undefined;
  
  /** Set shared state */
  set<T>(key: string, value: T): void;
  
  /** Subscribe to state changes */
  subscribe<T>(key: string, handler: (value: T) => void): Unsubscribe;
  
  /** Get state store */
  getStore<T>(key: string): Readable<T>;
}

export interface EventEmitterAPI {
  on(event: string, handler: EventHandler): Unsubscribe;
  off(event: string, handler: EventHandler): void;
  emit(event: string, data?: unknown): void;
}

// ============================================================================
// Utility Types
// ============================================================================

export type Unsubscribe = () => void;
export type EventHandler = (data: unknown) => void;

export interface SendOptions {
  target?: string;
  meta?: Record<string, unknown>;
}

export interface RequestOptions extends SendOptions {
  timeout?: number;
}

export interface SubscribeOptions {
  source?: string | string[];
  priority?: number;
}

export interface NavigateOptions {
  replace?: boolean;
  state?: unknown;
}

export interface RouteInfo {
  path: string;
  params: Record<string, string>;
  query: Record<string, string>;
  plugin?: string;
}

export interface UserContext {
  id: string;
  roles: string[];
  permissions: string[];
}

export interface Logger {
  debug(message: string, ...args: unknown[]): void;
  info(message: string, ...args: unknown[]): void;
  warn(message: string, ...args: unknown[]): void;
  error(message: string, ...args: unknown[]): void;
}

export interface PluginStorage {
  get<T>(key: string): Promise<T | undefined>;
  set<T>(key: string, value: T): Promise<void>;
  delete(key: string): Promise<void>;
  clear(): Promise<void>;
}

export interface PluginInstance {
  manifest: PluginManifest;
  status: PluginStatus;
  api?: PluginAPI;
  routes?: PluginRoutes;
  stores?: PluginStores;
}

export type PluginStatus = 
  | 'loading'
  | 'loaded'
  | 'enabled'
  | 'disabled'
  | 'error'
  | 'unloading';
```

---

## Core Implementation

### Plugin Registry

```typescript
// src/lib/core/plugin-registry.ts

import { writable, derived, get, type Readable, type Writable } from 'svelte/store';
import type {
  PluginDefinition,
  PluginInstance,
  PluginManifest,
  PluginStatus,
  PluginContext,
  CoreAPI,
  Unsubscribe
} from './types';
import { createMessageBus } from './message-bus';
import { createRouteManager } from './route-manager';
import { createStateManager } from './state-manager';
import { createLogger } from './logger';
import { createPluginStorage } from './storage';
import { validateManifest, validateDependencies } from './validators';

export interface PluginRegistry {
  /** Register a plugin definition */
  register(definition: PluginDefinition): Promise<void>;
  
  /** Unregister a plugin */
  unregister(id: string): Promise<void>;
  
  /** Enable a plugin */
  enable(id: string): Promise<void>;
  
  /** Disable a plugin */
  disable(id: string): Promise<void>;
  
  /** Get plugin instance */
  get(id: string): PluginInstance | undefined;
  
  /** Get all plugins */
  getAll(): PluginInstance[];
  
  /** Check if plugin exists */
  has(id: string): boolean;
  
  /** Get plugins store (reactive) */
  getStore(): Readable<Map<string, PluginInstance>>;
  
  /** Subscribe to plugin events */
  onLoad(handler: (plugin: PluginInstance) => void): Unsubscribe;
  onUnload(handler: (pluginId: string) => void): Unsubscribe;
  onError(handler: (pluginId: string, error: Error) => void): Unsubscribe;
}

export function createPluginRegistry(coreAPI: CoreAPI): PluginRegistry {
  const plugins: Writable<Map<string, PluginInstance>> = writable(new Map());
  const definitions: Map<string, PluginDefinition> = new Map();
  const contexts: Map<string, PluginContext> = new Map();
  
  // Event handlers
  const loadHandlers: Set<(plugin: PluginInstance) => void> = new Set();
  const unloadHandlers: Set<(pluginId: string) => void> = new Set();
  const errorHandlers: Set<(pluginId: string, error: Error) => void> = new Set();

  function updateStatus(id: string, status: PluginStatus): void {
    plugins.update(map => {
      const plugin = map.get(id);
      if (plugin) {
        map.set(id, { ...plugin, status });
      }
      return new Map(map);
    });
  }

  function emitError(id: string, error: Error): void {
    errorHandlers.forEach(handler => handler(id, error));
    const definition = definitions.get(id);
    definition?.manifest.lifecycle?.onError?.(error);
  }

  async function createPluginContext(definition: PluginDefinition): Promise<PluginContext> {
    const { manifest } = definition;
    
    return {
      core: coreAPI,
      config: {}, // Load from storage or defaults
      logger: createLogger(manifest.id),
      storage: createPluginStorage(manifest.id)
    };
  }

  return {
    async register(definition: PluginDefinition): Promise<void> {
      const { manifest } = definition;
      
      // Validate manifest
      const validation = validateManifest(manifest);
      if (!validation.valid) {
        throw new Error(`Invalid manifest for ${manifest.id}: ${validation.errors.join(', ')}`);
      }

      // Check if already registered
      if (definitions.has(manifest.id)) {
        throw new Error(`Plugin ${manifest.id} is already registered`);
      }

      // Validate dependencies
      const currentPlugins = get(plugins);
      const depValidation = validateDependencies(manifest, currentPlugins);
      if (!depValidation.valid) {
        throw new Error(`Missing dependencies for ${manifest.id}: ${depValidation.missing.join(', ')}`);
      }

      try {
        updateStatus(manifest.id, 'loading');
        
        // Execute lifecycle hook
        await manifest.lifecycle?.onBeforeLoad?.();

        // Store definition
        definitions.set(manifest.id, definition);

        // Create plugin instance
        const instance: PluginInstance = {
          manifest,
          status: 'loading',
          api: definition.api,
          routes: definition.routes,
          stores: definition.stores
        };

        // Create context
        const context = await createPluginContext(definition);
        contexts.set(manifest.id, context);

        // Run setup
        await definition.setup?.(context);

        // Register routes
        if (definition.routes) {
          coreAPI.routes.register(definition.routes.routes);
        }

        // Register message handlers
        if (definition.messages) {
          for (const [type, handler] of Object.entries(definition.messages.handlers)) {
            coreAPI.messages.subscribe(
              `${manifest.id}:${type}`,
              handler,
              { source: '*' }
            );
          }
        }

        // Update instance status
        instance.status = manifest.enabled !== false ? 'enabled' : 'disabled';
        
        plugins.update(map => {
          map.set(manifest.id, instance);
          return new Map(map);
        });

        // Execute lifecycle hook
        await manifest.lifecycle?.onLoad?.();

        // Emit load event
        loadHandlers.forEach(handler => handler(instance));

      } catch (error) {
        updateStatus(manifest.id, 'error');
        emitError(manifest.id, error as Error);
        throw error;
      }
    },

    async unregister(id: string): Promise<void> {
      const definition = definitions.get(id);
      if (!definition) {
        throw new Error(`Plugin ${id} is not registered`);
      }

      try {
        updateStatus(id, 'unloading');

        // Execute lifecycle hook
        await definition.manifest.lifecycle?.onUnload?.();

        // Run teardown
        await definition.teardown?.();

        // Unregister routes
        if (definition.routes) {
          const paths = definition.routes.routes.map(r => 
            `${definition.routes!.basePath}${r.path}`
          );
          coreAPI.routes.unregister(paths);
        }

        // Clean up
        definitions.delete(id);
        contexts.delete(id);
        
        plugins.update(map => {
          map.delete(id);
          return new Map(map);
        });

        // Emit unload event
        unloadHandlers.forEach(handler => handler(id));

      } catch (error) {
        emitError(id, error as Error);
        throw error;
      }
    },

    async enable(id: string): Promise<void> {
      const definition = definitions.get(id);
      if (!definition) {
        throw new Error(`Plugin ${id} is not registered`);
      }

      await definition.manifest.lifecycle?.onEnable?.();
      updateStatus(id, 'enabled');
    },

    async disable(id: string): Promise<void> {
      const definition = definitions.get(id);
      if (!definition) {
        throw new Error(`Plugin ${id} is not registered`);
      }

      await definition.manifest.lifecycle?.onDisable?.();
      updateStatus(id, 'disabled');
    },

    get(id: string): PluginInstance | undefined {
      return get(plugins).get(id);
    },

    getAll(): PluginInstance[] {
      return Array.from(get(plugins).values());
    },

    has(id: string): boolean {
      return get(plugins).has(id);
    },

    getStore(): Readable<Map<string, PluginInstance>> {
      return { subscribe: plugins.subscribe };
    },

    onLoad(handler): Unsubscribe {
      loadHandlers.add(handler);
      return () => loadHandlers.delete(handler);
    },

    onUnload(handler): Unsubscribe {
      unloadHandlers.add(handler);
      return () => unloadHandlers.delete(handler);
    },

    onError(handler): Unsubscribe {
      errorHandlers.add(handler);
      return () => errorHandlers.delete(handler);
    }
  };
}
```

### Message Bus

```typescript
// src/lib/core/message-bus.ts

import { nanoid } from 'nanoid';
import type {
  Message,
  MessageHandler,
  MessageBusAPI,
  SendOptions,
  RequestOptions,
  SubscribeOptions,
  Unsubscribe
} from './types';

interface Subscription {
  type: string;
  handler: MessageHandler;
  options: SubscribeOptions;
  pluginId?: string;
}

export interface MessageBus extends MessageBusAPI {
  /** Create a scoped message bus for a plugin */
  createScoped(pluginId: string): MessageBusAPI;
  
  /** Get message history */
  getHistory(type?: string, limit?: number): Message[];
  
  /** Clear message history */
  clearHistory(): void;
}

export function createMessageBus(): MessageBus {
  const subscriptions: Map<string, Set<Subscription>> = new Map();
  const pendingRequests: Map<string, {
    resolve: (value: unknown) => void;
    reject: (error: Error) => void;
    timeout: ReturnType<typeof setTimeout>;
  }> = new Map();
  const messageHistory: Message[] = [];
  const MAX_HISTORY = 1000;

  function matchType(pattern: string, type: string): boolean {
    if (pattern === '*') return true;
    if (pattern === type) return true;
    
    // Support wildcard patterns like 'plugin:*' or '*.event'
    const regexPattern = pattern
      .replace(/\./g, '\\.')
      .replace(/\*/g, '.*');
    return new RegExp(`^${regexPattern}$`).test(type);
  }

  function getSubscriptions(type: string): Subscription[] {
    const result: Subscription[] = [];
    
    for (const [pattern, subs] of subscriptions) {
      if (matchType(pattern, type)) {
        result.push(...subs);
      }
    }
    
    // Sort by priority (higher priority first)
    return result.sort((a, b) => (b.options.priority ?? 0) - (a.options.priority ?? 0));
  }

  function createMessage<T>(
    type: string,
    payload: T,
    source: string,
    options?: SendOptions
  ): Message<T> {
    return {
      id: nanoid(),
      type,
      source,
      target: options?.target,
      payload,
      timestamp: Date.now(),
      meta: options?.meta
    };
  }

  function addToHistory(message: Message): void {
    messageHistory.push(message);
    if (messageHistory.length > MAX_HISTORY) {
      messageHistory.shift();
    }
  }

  async function dispatch(message: Message): Promise<unknown> {
    const subs = getSubscriptions(message.type);
    let result: unknown;

    for (const sub of subs) {
      // Check source filter
      if (sub.options.source) {
        const sources = Array.isArray(sub.options.source) 
          ? sub.options.source 
          : [sub.options.source];
        if (!sources.includes(message.source) && !sources.includes('*')) {
          continue;
        }
      }

      // Check target filter
      if (message.target && sub.pluginId && message.target !== sub.pluginId) {
        continue;
      }

      // Apply filter if exists
      if (sub.handler.filter && !sub.handler.filter(message)) {
        continue;
      }

      try {
        result = await sub.handler.handle(message);
      } catch (error) {
        console.error(`Error in message handler for ${message.type}:`, error);
      }
    }

    return result;
  }

  const bus: MessageBus = {
    send<T>(type: string, payload: T, options?: SendOptions): void {
      const message = createMessage(type, payload, 'core', options);
      addToHistory(message);
      dispatch(message);
    },

    async request<TReq, TRes>(
      type: string,
      payload: TReq,
      options?: RequestOptions
    ): Promise<TRes> {
      const correlationId = nanoid();
      const timeout = options?.timeout ?? 30000;

      const message = createMessage(type, payload, 'core', options);
      message.correlationId = correlationId;

      return new Promise<TRes>((resolve, reject) => {
        const timeoutHandle = setTimeout(() => {
          pendingRequests.delete(correlationId);
          reject(new Error(`Request timeout for ${type}`));
        }, timeout);

        pendingRequests.set(correlationId, {
          resolve: resolve as (value: unknown) => void,
          reject,
          timeout: timeoutHandle
        });

        addToHistory(message);
        dispatch(message).then(result => {
          const pending = pendingRequests.get(correlationId);
          if (pending) {
            clearTimeout(pending.timeout);
            pendingRequests.delete(correlationId);
            pending.resolve(result);
          }
        });
      });
    },

    subscribe(
      type: string,
      handler: MessageHandler,
      options: SubscribeOptions = {}
    ): Unsubscribe {
      const subscription: Subscription = { type, handler, options };

      if (!subscriptions.has(type)) {
        subscriptions.set(type, new Set());
      }
      subscriptions.get(type)!.add(subscription);

      return () => {
        subscriptions.get(type)?.delete(subscription);
        if (subscriptions.get(type)?.size === 0) {
          subscriptions.delete(type);
        }
      };
    },

    async once<T>(type: string, options?: SubscribeOptions): Promise<Message<T>> {
      return new Promise((resolve) => {
        const unsubscribe = this.subscribe(
          type,
          {
            handle: (message) => {
              unsubscribe();
              resolve(message as Message<T>);
            }
          },
          options
        );
      });
    },

    createScoped(pluginId: string): MessageBusAPI {
      return {
        send: <T>(type: string, payload: T, options?: SendOptions) => {
          const message = createMessage(type, payload, pluginId, options);
          addToHistory(message);
          dispatch(message);
        },

        request: async <TReq, TRes>(
          type: string,
          payload: TReq,
          options?: RequestOptions
        ): Promise<TRes> => {
          const correlationId = nanoid();
          const timeout = options?.timeout ?? 30000;

          const message = createMessage(type, payload, pluginId, options);
          message.correlationId = correlationId;

          return new Promise<TRes>((resolve, reject) => {
            const timeoutHandle = setTimeout(() => {
              pendingRequests.delete(correlationId);
              reject(new Error(`Request timeout for ${type}`));
            }, timeout);

            pendingRequests.set(correlationId, {
              resolve: resolve as (value: unknown) => void,
              reject,
              timeout: timeoutHandle
            });

            addToHistory(message);
            dispatch(message).then(result => {
              const pending = pendingRequests.get(correlationId);
              if (pending) {
                clearTimeout(pending.timeout);
                pendingRequests.delete(correlationId);
                pending.resolve(result);
              }
            });
          });
        },

        subscribe: (
          type: string,
          handler: MessageHandler,
          options: SubscribeOptions = {}
        ): Unsubscribe => {
          const subscription: Subscription = { 
            type, 
            handler, 
            options,
            pluginId 
          };

          if (!subscriptions.has(type)) {
            subscriptions.set(type, new Set());
          }
          subscriptions.get(type)!.add(subscription);

          return () => {
            subscriptions.get(type)?.delete(subscription);
          };
        },

        once: async <T>(type: string, options?: SubscribeOptions): Promise<Message<T>> => {
          return new Promise((resolve) => {
            const unsubscribe = bus.createScoped(pluginId).subscribe(
              type,
              {
                handle: (message) => {
                  unsubscribe();
                  resolve(message as Message<T>);
                }
              },
              options
            );
          });
        }
      };
    },

    getHistory(type?: string, limit = 100): Message[] {
      let history = [...messageHistory];
      
      if (type) {
        history = history.filter(m => matchType(type, m.type));
      }
      
      return history.slice(-limit);
    },

    clearHistory(): void {
      messageHistory.length = 0;
    }
  };

  return bus;
}
```

### Plugin Loader

```typescript
// src/lib/core/plugin-loader.ts

import type { PluginDefinition, PluginManifest } from './types';

export interface PluginLoader {
  /** Load plugins from manifest list */
  loadFromManifests(manifests: PluginManifestEntry[]): Promise<PluginDefinition[]>;
  
  /** Load a single plugin by ID */
  loadPlugin(id: string, source: PluginSource): Promise<PluginDefinition>;
  
  /** Discover plugins from directory */
  discoverPlugins(basePath: string): Promise<PluginManifestEntry[]>;
}

export interface PluginManifestEntry {
  id: string;
  source: PluginSource;
  enabled?: boolean;
}

export type PluginSource = 
  | { type: 'builtin'; path: string }
  | { type: 'external'; url: string }
  | { type: 'local'; path: string };

export function createPluginLoader(): PluginLoader {
  const loadedPlugins: Map<string, PluginDefinition> = new Map();

  async function loadBuiltinPlugin(path: string): Promise<PluginDefinition> {
    // Dynamic import for built-in plugins
    const module = await import(/* @vite-ignore */ `../plugins/${path}/index.ts`);
    return module.default as PluginDefinition;
  }

  async function loadExternalPlugin(url: string): Promise<PluginDefinition> {
    // Fetch and load external plugin
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`Failed to fetch plugin from ${url}`);
    }
    
    const code = await response.text();
    
    // Create a sandboxed module
    const blob = new Blob([code], { type: 'application/javascript' });
    const objectUrl = URL.createObjectURL(blob);
    
    try {
      const module = await import(/* @vite-ignore */ objectUrl);
      return module.default as PluginDefinition;
    } finally {
      URL.revokeObjectURL(objectUrl);
    }
  }

  async function loadLocalPlugin(path: string): Promise<PluginDefinition> {
    // Load from local file system (for development)
    const module = await import(/* @vite-ignore */ path);
    return module.default as PluginDefinition;
  }

  return {
    async loadFromManifests(manifests: PluginManifestEntry[]): Promise<PluginDefinition[]> {
      const results: PluginDefinition[] = [];
      
      // Sort by dependencies (topological sort)
      const sorted = topologicalSort(manifests);
      
      for (const entry of sorted) {
        if (entry.enabled === false) continue;
        
        try {
          const plugin = await this.loadPlugin(entry.id, entry.source);
          results.push(plugin);
          loadedPlugins.set(entry.id, plugin);
        } catch (error) {
          console.error(`Failed to load plugin ${entry.id}:`, error);
        }
      }
      
      return results;
    },

    async loadPlugin(id: string, source: PluginSource): Promise<PluginDefinition> {
      // Check cache
      if (loadedPlugins.has(id)) {
        return loadedPlugins.get(id)!;
      }

      let plugin: PluginDefinition;

      switch (source.type) {
        case 'builtin':
          plugin = await loadBuiltinPlugin(source.path);
          break;
        case 'external':
          plugin = await loadExternalPlugin(source.url);
          break;
        case 'local':
          plugin = await loadLocalPlugin(source.path);
          break;
        default:
          throw new Error(`Unknown plugin source type`);
      }

      // Validate plugin structure
      if (!plugin.manifest?.id) {
        throw new Error(`Plugin missing required manifest.id`);
      }

      loadedPlugins.set(id, plugin);
      return plugin;
    },

    async discoverPlugins(basePath: string): Promise<PluginManifestEntry[]> {
      // In production, this would scan a directory
      // For SvelteKit, we use a manifest file or API endpoint
      const response = await fetch(`${basePath}/manifest.json`);
      if (!response.ok) {
        return [];
      }
      
      return response.json();
    }
  };
}

function topologicalSort(manifests: PluginManifestEntry[]): PluginManifestEntry[] {
  // Simple topological sort based on dependencies
  // In production, implement proper dependency resolution
  return manifests.sort((a, b) => {
    // Plugins without dependencies load first
    return 0;
  });
}
```

### Route Manager

```typescript
// src/lib/core/route-manager.ts

import { writable, get, type Readable } from 'svelte/store';
import { goto } from '$app/navigation';
import { page } from '$app/stores';
import type {
  RouteDefinition,
  RouteManagerAPI,
  NavigateOptions,
  RouteInfo,
  NavigationItem,
  Unsubscribe
} from './types';

export interface RouteManager extends RouteManagerAPI {
  /** Get all registered routes */
  getAllRoutes(): RouteDefinition[];
  
  /** Get navigation items */
  getNavigation(): NavigationItem[];
  
  /** Get routes store */
  getStore(): Readable<RouteDefinition[]>;
  
  /** Match a path to a route */
  matchRoute(path: string): { route: RouteDefinition; params: Record<string, string> } | null;
}

interface RegisteredRoute extends RouteDefinition {
  pluginId: string;
  fullPath: string;
}

export function createRouteManager(): RouteManager {
  const routes = writable<RegisteredRoute[]>([]);
  const navigation = writable<NavigationItem[]>([]);

  function parsePath(path: string): { segments: string[]; params: string[] } {
    const segments = path.split('/').filter(Boolean);
    const params = segments
      .filter(s => s.startsWith(':') || s.startsWith('['))
      .map(s => s.replace(/[:[\]]/g, ''));
    return { segments, params };
  }

  function matchPath(
    pattern: string,
    path: string
  ): Record<string, string> | null {
    const patternParts = pattern.split('/').filter(Boolean);
    const pathParts = path.split('/').filter(Boolean);

    if (patternParts.length !== pathParts.length) {
      // Check for catch-all
      const lastPattern = patternParts[patternParts.length - 1];
      if (!lastPattern?.startsWith('...') && !lastPattern?.includes('...')) {
        return null;
      }
    }

    const params: Record<string, string> = {};

    for (let i = 0; i < patternParts.length; i++) {
      const pattern = patternParts[i];
      const part = pathParts[i];

      if (pattern.startsWith(':')) {
        // Dynamic segment
        params[pattern.slice(1)] = part;
      } else if (pattern.startsWith('[') && pattern.endsWith(']')) {
        // SvelteKit style dynamic segment
        const paramName = pattern.slice(1, -1).replace('...', '');
        if (pattern.includes('...')) {
          // Catch-all
          params[paramName] = pathParts.slice(i).join('/');
          break;
        }
        params[paramName] = part;
      } else if (pattern !== part) {
        return null;
      }
    }

    return params;
  }

  return {
    register(newRoutes: RouteDefinition[]): void {
      routes.update(current => {
        const updated = [...current];
        
        for (const route of newRoutes) {
          const existing = updated.findIndex(r => r.fullPath === route.path);
          if (existing >= 0) {
            updated[existing] = route as RegisteredRoute;
          } else {
            updated.push(route as RegisteredRoute);
          }
        }
        
        return updated;
      });
    },

    unregister(paths: string[]): void {
      routes.update(current => 
        current.filter(r => !paths.includes(r.fullPath))
      );
    },

    async navigate(path: string, options?: NavigateOptions): Promise<void> {
      await goto(path, {
        replaceState: options?.replace,
        state: options?.state
      });
    },

    getCurrentRoute(): RouteInfo {
      const $page = get(page);
      return {
        path: $page.url.pathname,
        params: $page.params,
        query: Object.fromEntries($page.url.searchParams)
      };
    },

    getAllRoutes(): RouteDefinition[] {
      return get(routes);
    },

    getNavigation(): NavigationItem[] {
      return get(navigation);
    },

    getStore(): Readable<RouteDefinition[]> {
      return { subscribe: routes.subscribe };
    },

    matchRoute(path: string) {
      const allRoutes = get(routes);
      
      for (const route of allRoutes) {
        const params = matchPath(route.fullPath, path);
        if (params) {
          return { route, params };
        }
      }
      
      return null;
    }
  };
}
```

---

## Example Plugin Implementation

```typescript
// src/lib/plugins/dashboard/index.ts

import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';
import { api } from './api';
import { messages } from './messages';
import { stores } from './stores';
import { routes } from './routes';

const dashboardPlugin: PluginDefinition = {
  manifest,
  api,
  messages,
  stores,
  routes,

  async setup(context) {
    const { core, logger, storage } = context;
    
    logger.info('Dashboard plugin initializing...');
    
    // Load saved preferences
    const prefs = await storage.get('preferences');
    if (prefs) {
      stores.writable?.preferences?.set(prefs);
    }
    
    // Subscribe to core events
    core.events.on('user:login', (user) => {
      logger.info('User logged in:', user);
    });
    
    // Subscribe to messages from other plugins
    core.messages.subscribe('analytics:data-updated', {
      handle: (message) => {
        logger.debug('Analytics data updated:', message.payload);
        // Update dashboard widgets
      }
    });
    
    logger.info('Dashboard plugin initialized');
  },

  async teardown() {
    console.log('Dashboard plugin cleaning up...');
  }
};

export default dashboardPlugin;
```

```typescript
// src/lib/plugins/dashboard/manifest.ts

import type { PluginManifest } from '$lib/core/types';
import DashboardIcon from './components/DashboardIcon.svelte';

export const manifest: PluginManifest = {
  id: 'dashboard',
  name: 'Dashboard',
  version: '1.0.0',
  description: 'Main dashboard with customizable widgets',
  author: 'Your Team',
  icon: DashboardIcon,
  
  dependencies: [
    { pluginId: 'core-ui', version: '^1.0.0' }
  ],
  
  permissions: [
    'storage:read',
    'storage:write',
    'plugins:communicate',
    'routes:register'
  ],
  
  lifecycle: {
    onEnable: () => console.log('Dashboard enabled'),
    onDisable: () => console.log('Dashboard disabled')
  },
  
  configSchema: {
    type: 'object',
    properties: {
      defaultLayout: {
        type: 'string',
        title: 'Default Layout',
        description: 'The default dashboard layout',
        enum: ['grid', 'list', 'compact'],
        default: 'grid'
      },
      refreshInterval: {
        type: 'number',
        title: 'Refresh Interval',
        description: 'Data refresh interval in seconds',
        default: 30
      }
    }
  },
  
  priority: 100,
  enabled: true
};
```

```typescript
// src/lib/plugins/dashboard/api.ts

import type { PluginAPI } from '$lib/core/types';

export const api: PluginAPI = {
  version: '1.0.0',
  
  methods: {
    getWidgets: {
      handler: async (input, context) => {
        // Return available widgets
        return [
          { id: 'stats', type: 'statistics', title: 'Statistics' },
          { id: 'chart', type: 'chart', title: 'Activity Chart' },
          { id: 'recent', type: 'list', title: 'Recent Items' }
        ];
      },
      description: 'Get available dashboard widgets',
      outputSchema: {
        type: 'array',
        items: {
          type: 'object',
          properties: {
            id: { type: 'string' },
            type: { type: 'string' },
            title: { type: 'string' }
          }
        }
      }
    },
    
    getLayout: {
      handler: async (input: { userId?: string }, context) => {
        // Return user's dashboard layout
        return {
          columns: 3,
          widgets: [
            { id: 'stats', position: { x: 0, y: 0, w: 2, h: 1 } },
            { id: 'chart', position: { x: 2, y: 0, w: 1, h: 2 } },
            { id: 'recent', position: { x: 0, y: 1, w: 2, h: 1 } }
          ]
        };
      },
      inputSchema: {
        type: 'object',
        properties: {
          userId: { type: 'string' }
        }
      },
      description: 'Get dashboard layout for a user'
    },
    
    saveLayout: {
      handler: async (input: { layout: unknown }, context) => {
        // Save layout to storage
        return { success: true };
      },
      permissions: ['storage:write'],
      description: 'Save dashboard layout'
    }
  },
  
  docs: {
    title: 'Dashboard API',
    description: 'API for managing dashboard widgets and layouts',
    methods: {
      getWidgets: {
        summary: 'Retrieve available widgets',
        examples: [{
          title: 'Get all widgets',
          input: {},
          output: [{ id: 'stats', type: 'statistics', title: 'Statistics' }]
        }]
      }
    }
  }
};
```

```typescript
// src/lib/plugins/dashboard/messages.ts

import type { PluginMessages } from '$lib/core/types';

export const messages: PluginMessages = {
  handlers: {
    'widget:refresh': {
      handle: async (message) => {
        const { widgetId } = message.payload as { widgetId: string };
        console.log(`Refreshing widget: ${widgetId}`);
        // Refresh widget data
        return { refreshed: true, widgetId };
      },
      priority: 10
    },
    
    'layout:changed': {
      handle: async (message) => {
        const { layout } = message.payload as { layout: unknown };
        console.log('Layout changed:', layout);
        // Handle layout change
      }
    }
  },
  
  schemas: {
    'widget:refresh': {
      type: 'dashboard:widget:refresh',
      payload: {
        type: 'object',
        properties: {
          widgetId: { type: 'string' }
        },
        required: ['widgetId']
      },
      response: {
        type: 'object',
        properties: {
          refreshed: { type: 'boolean' },
          widgetId: { type: 'string' }
        }
      },
      expectsResponse: true,
      description: 'Request a widget to refresh its data'
    },
    
    'layout:changed': {
      type: 'dashboard:layout:changed',
      payload: {
        type: 'object',
        properties: {
          layout: { type: 'object' }
        }
      },
      description: 'Notification that the dashboard layout has changed'
    },
    
    'widget:added': {
      type: 'dashboard:widget:added',
      payload: {
        type: 'object',
        properties: {
          widgetId: { type: 'string' },
          widgetType: { type: 'string' },
          position: { type: 'object' }
        }
      },
      description: 'Notification that a widget was added to the dashboard'
    }
  },
  
  subscriptions: [
    {
      type: 'analytics:*',
      handler: {
        handle: (message) => {
          console.log('Analytics event received:', message.type);
        }
      }
    },
    {
      type: 'user:preferences:changed',
      source: 'settings',
      handler: {
        handle: (message) => {
          console.log('User preferences changed');
        }
      }
    }
  ]
};
```

```typescript
// src/lib/plugins/dashboard/stores.ts

import { writable, derived } from 'svelte/store';
import type { PluginStores } from '$lib/core/types';

// Internal stores
const widgets = writable<Widget[]>([]);
const layout = writable<Layout | null>(null);
const preferences = writable<DashboardPreferences>({
  theme: 'auto',
  refreshInterval: 30,
  compactMode: false
});
const isLoading = writable(false);

// Derived stores
const activeWidgets = derived(
  [widgets, layout],
  ([$widgets, $layout]) => {
    if (!$layout) return [];
    return $layout.widgets
      .map(w => $widgets.find(widget => widget.id === w.id))
      .filter(Boolean);
  }
);

// Types
interface Widget {
  id: string;
  type: string;
  title: string;
  config?: Record<string, unknown>;
}

interface Layout {
  columns: number;
  widgets: Array<{
    id: string;
    position: { x: number; y: number; w: number; h: number };
  }>;
}

interface DashboardPreferences {
  theme: 'light' | 'dark' | 'auto';
  refreshInterval: number;
  compactMode: boolean;
}

export const stores: PluginStores = {
  readable: {
    activeWidgets,
    isLoading: { subscribe: isLoading.subscribe }
  },
  
  writable: {
    widgets,
    layout,
    preferences
  },
  
  schemas: {
    widgets: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          id: { type: 'string' },
          type: { type: 'string' },
          title: { type: 'string' }
        }
      }
    },
    preferences: {
      type: 'object',
      properties: {
        theme: { type: 'string', enum: ['light', 'dark', 'auto'] },
        refreshInterval: { type: 'number' },
        compactMode: { type: 'boolean' }
      }
    }
  }
};

// Export individual stores for internal use
export { widgets, layout, preferences, isLoading, activeWidgets };
```

```typescript
// src/lib/plugins/dashboard/routes/index.ts

import type { PluginRoutes } from '$lib/core/types';
import DashboardPage from './+page.svelte';
import DashboardLayout from './+layout.svelte';
import WidgetSettings from './widgets/+page.svelte';
import DashboardIcon from '../components/DashboardIcon.svelte';

export const routes: PluginRoutes = {
  basePath: '/dashboard',
  
  routes: [
    {
      path: '/',
      component: DashboardPage,
      layout: DashboardLayout,
      meta: {
        title: 'Dashboard',
        description: 'Main dashboard view',
        icon: 'dashboard',
        order: 1
      }
    },
    {
      path: '/widgets',
      component: WidgetSettings,
      meta: {
        title: 'Widget Settings',
        description: 'Manage dashboard widgets',
        order: 2,
        permissions: ['storage:write']
      },
      guards: [
        {
          canActivate: (context) => {
            // Check if user has permission
            return context.user?.permissions.includes('dashboard:manage') ?? false;
          },
          redirectTo: '/dashboard'
        }
      ]
    }
  ],
  
  navigation: [
    {
      id: 'dashboard-main',
      label: 'Dashboard',
      path: '/dashboard',
      icon: DashboardIcon,
      order: 1,
      children: [
        {
          id: 'dashboard-widgets',
          label: 'Widgets',
          path: '/dashboard/widgets',
          order: 1
        }
      ]
    }
  ]
};
```

---

## SvelteKit Integration

### Root Layout

```svelte
<!-- src/routes/+layout.svelte -->
<script lang="ts">
  import { onMount, setContext } from 'svelte';
  import { writable } from 'svelte/store';
  import { createCore } from '$lib/core';
  import { PluginProvider } from '$lib/core/components';
  import type { LayoutData } from './$types';
  
  export let data: LayoutData;
  
  const core = createCore();
  const pluginsLoaded = writable(false);
  
  setContext('core', core);
  setContext('pluginsLoaded', pluginsLoaded);
  
  onMount(async () => {
    // Load plugins from server-provided manifests
    for (const manifest of data.pluginManifests) {
      try {
        const plugin = await core.loader.loadPlugin(manifest.id, manifest.source);
        await core.registry.register(plugin);
      } catch (error) {
        console.error(`Failed to load plugin ${manifest.id}:`, error);
      }
    }
    
    $pluginsLoaded = true;
  });
</script>

<PluginProvider {core}>
  {#if $pluginsLoaded}
    <slot />
  {:else}
    <div class="loading">
      <span>Loading plugins...</span>
    </div>
  {/if}
</PluginProvider>

<style>
  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
  }
</style>
```

### Server Layout (Plugin Discovery)

```typescript
// src/routes/+layout.server.ts

import type { LayoutServerLoad } from './$types';
import type { PluginManifestEntry } from '$lib/core/plugin-loader';

export const load: LayoutServerLoad = async ({ fetch }) => {
  // Load plugin manifests from API or file system
  const manifests: PluginManifestEntry[] = [
    // Built-in plugins
    { id: 'dashboard', source: { type: 'builtin', path: 'dashboard' }, enabled: true },
    { id: 'settings', source: { type: 'builtin', path: 'settings' }, enabled: true },
    { id: 'analytics', source: { type: 'builtin', path: 'analytics' }, enabled: true },
  ];
  
  // Optionally load from external sources
  try {
    const response = await fetch('/api/plugins/manifest');
    if (response.ok) {
      const externalManifests = await response.json();
      manifests.push(...externalManifests);
    }
  } catch (error) {
    console.error('Failed to load external plugin manifests:', error);
  }
  
  return {
    pluginManifests: manifests
  };
};
```

### Dynamic Route Handler

```svelte
<!-- src/routes/[[...catchall]]/+page.svelte -->
<script lang="ts">
  import { page } from '$app/stores';
  import { getContext } from 'svelte';
  import type { CoreAPI } from '$lib/core/types';
  
  const core = getContext<CoreAPI>('core');
  
  $: currentPath = $page.url.pathname;
  $: matchedRoute = core.routes.matchRoute(currentPath);
</script>

{#if matchedRoute}
  <svelte:component 
    this={matchedRoute.route.component} 
    params={matchedRoute.params}
  />
{:else}
  <div class="not-found">
    <h1>404 - Page Not Found</h1>
    <p>The requested page does not exist.</p>
  </div>
{/if}
```

### Plugin Manifest API

```typescript
// src/routes/api/plugins/manifest/+server.ts

import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import type { PluginManifestEntry } from '$lib/core/plugin-loader';

export const GET: RequestHandler = async ({ url }) => {
  // In production, this could read from a database, file system, or external service
  const manifests: PluginManifestEntry[] = [];
  
  // Example: Load from environment-configured external plugins
  const externalPlugins = process.env.EXTERNAL_PLUGINS?.split(',') ?? [];
  
  for (const pluginUrl of externalPlugins) {
    try {
      const response = await fetch(`${pluginUrl}/manifest.json`);
      if (response.ok) {
        const manifest = await response.json();
        manifests.push({
          id: manifest.id,
          source: { type: 'external', url: `${pluginUrl}/index.js` },
          enabled: true
        });
      }
    } catch (error) {
      console.error(`Failed to load manifest from ${pluginUrl}:`, error);
    }
  }
  
  return json(manifests);
};
```

---

## Message Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         MESSAGE BUS ARCHITECTURE                            │
└─────────────────────────────────────────────────────────────────────────────┘

  Plugin A                    Message Bus                    Plugin B
  ────────                    ───────────                    ────────
      │                            │                            │
      │  1. send('user:action',    │                            │
      │     { action: 'click' })   │                            │
      │ ─────────────────────────► │                            │
      │                            │                            │
      │                            │  2. Match subscriptions    │
      │                            │     type: 'user:*'         │
      │                            │ ─────────────────────────► │
      │                            │                            │
      │                            │  3. Handler executes       │
      │                            │ ◄───────────────────────── │
      │                            │                            │
      │                            │                            │
      │  4. request('data:fetch',  │                            │
      │     { id: 123 })           │                            │
      │ ─────────────────────────► │                            │
      │                            │                            │
      │                            │  5. Route to target        │
      │                            │ ─────────────────────────► │
      │                            │                            │
      │                            │  6. Process & respond      │
      │                            │ ◄───────────────────────── │
      │                            │                            │
      │  7. Receive response       │                            │
      │ ◄───────────────────────── │                            │
      │                            │                            │


┌─────────────────────────────────────────────────────────────────────────────┐
│                         MESSAGE TYPES                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  BROADCAST (Fire & Forget)          REQUEST/RESPONSE                         │
│  ─────────────────────────          ──────────────────                       │
│                                                                              │
│  core.messages.send(                core.messages.request(                   │
│    'notification:show',               'user:get',                            │
│    { text: 'Hello' }                  { id: 123 },                           │
│  );                                   { timeout: 5000 }                      │
│                                     );                                       │
│  // No response expected            // Returns Promise<User>                 │
│                                                                              │
│  TARGETED MESSAGE                   WILDCARD SUBSCRIPTION                    │
│  ────────────────────               ──────────────────────                   │
│                                                                              │
│  core.messages.send(                core.messages.subscribe(                 │
│    'settings:update',                 'analytics:*',                         │
│    { theme: 'dark' },                 { handle: (msg) => {                   │
│    { target: 'settings' }               // Handle all analytics events       │
│  );                                   }}                                     │
│                                     );                                       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Configuration & Utilities

```typescript
// src/lib/core/index.ts - Main export

import { createMessageBus } from './message-bus';
import { createPluginRegistry } from './plugin-registry';
import { createPluginLoader } from './plugin-loader';
import { createRouteManager } from './route-manager';
import { createStateManager } from './state-manager';
import { createEventEmitter } from './event-emitter';
import type { CoreAPI } from './types';

export interface Core {
  api: CoreAPI;
  registry: ReturnType<typeof createPluginRegistry>;
  loader: ReturnType<typeof createPluginLoader>;
}

export function createCore(): Core {
  const messageBus = createMessageBus();
  const routeManager = createRouteManager();
  const stateManager = createStateManager();
  const eventEmitter = createEventEmitter();

  const api: CoreAPI = {
    messages: messageBus,
    plugins: null!, // Set after registry creation
    routes: routeManager,
    state: stateManager,
    events: eventEmitter
  };

  const registry = createPluginRegistry(api);
  const loader = createPluginLoader();

  // Wire up plugins API
  api.plugins = {
    get: registry.get,
    getAll: registry.getAll,
    has: registry.has,
    call: async (pluginId, method, input) => {
      const plugin = registry.get(pluginId);
      if (!plugin?.api?.methods[method]) {
        throw new Error(`Method ${method} not found on plugin ${pluginId}`);
      }
      return plugin.api.methods[method].handler(input, {});
    },
    onPluginLoad: registry.onLoad,
    onPluginUnload: registry.onUnload
  };

  return {
    api,
    registry,
    loader
  };
}

// Re-export types
export * from './types';
```

---

## Development Roadmap

### Phase 1: Core Foundation (Week 1-2)
- [ ] Implement type definitions
- [ ] Build message bus with pub/sub
- [ ] Create plugin registry
- [ ] Implement plugin loader
- [ ] Basic route manager

### Phase 2: Plugin Infrastructure (Week 3-4)
- [ ] Plugin validation & sandboxing
- [ ] Dependency resolution
- [ ] Permission system
- [ ] Storage abstraction
- [ ] Logging system

### Phase 3: SvelteKit Integration (Week 5-6)
- [ ] Dynamic route injection
- [ ] Layout composition
- [ ] Server-side plugin discovery
- [ ] API endpoints for plugins
- [ ] Hot module replacement support

### Phase 4: Developer Experience (Week 7-8)
- [ ] Plugin CLI scaffolding
- [ ] Documentation generator
- [ ] Debug tools & devtools panel
- [ ] Testing utilities
- [ ] Example plugins

### Phase 5: Production Hardening (Week 9-10)
- [ ] Error boundaries
- [ ] Performance optimization
- [ ] Security audit
- [ ] Monitoring & metrics
- [ ] Production deployment guide

---

## Best Practices

### Plugin Development Guidelines

1. **Single Responsibility**: Each plugin should focus on one domain
2. **Explicit Dependencies**: Declare all dependencies in manifest
3. **Graceful Degradation**: Handle missing optional dependencies
4. **Message Contracts**: Define schemas for all messages
5. **Type Safety**: Export TypeScript types for your API
6. **Documentation**: Include API documentation in your plugin
7. **Testing**: Write tests for handlers and API methods
8. **Error Handling**: Never throw unhandled errors in handlers

### Message Naming Convention

```
<domain>:<entity>:<action>

Examples:
- user:profile:updated
- dashboard:widget:added
- analytics:event:tracked
- settings:theme:changed
```

### Route Naming Convention

```
/<plugin-name>/<feature>/<action>

Examples:
- /dashboard/widgets/configure
- /settings/profile/edit
- /analytics/reports/weekly
```
