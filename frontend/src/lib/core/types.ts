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
  | string;

export interface PluginLifecycle {
  onBeforeLoad?: () => Promise<void> | void;
  onLoad?: () => Promise<void> | void;
  onEnable?: () => Promise<void> | void;
  onDisable?: () => Promise<void> | void;
  onUnload?: () => Promise<void> | void;
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
  version: string;
  methods: Record<string, PluginMethod>;
  docs?: APIDocumentation;
}

export interface PluginMethod<TInput = unknown, TOutput = unknown> {
  handler: (input: TInput, context: MethodContext) => Promise<TOutput> | TOutput;
  inputSchema?: JSONSchema;
  outputSchema?: JSONSchema;
  description?: string;
  permissions?: PluginPermission[];
}

export interface MethodContext {
  callerId?: string;
  user?: UserContext;
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
  basePath: string;
  routes: RouteDefinition[];
  navigation?: NavigationItem[];
}

export interface RouteDefinition {
  path: string;
  component: ComponentType<SvelteComponent>;
  layout?: ComponentType<SvelteComponent>;
  meta?: RouteMeta;
  children?: RouteDefinition[];
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
  handlers: Record<string, MessageHandler>;
  schemas: Record<string, MessageSchema>;
  subscriptions?: MessageSubscription[];
}

export interface MessageSchema {
  type: string;
  payload: JSONSchema;
  response?: JSONSchema;
  description?: string;
  expectsResponse?: boolean;
}

export interface MessageHandler<TPayload = unknown, TResponse = void> {
  handle: (message: Message<TPayload>) => Promise<TResponse> | TResponse;
  filter?: (message: Message<TPayload>) => boolean;
  priority?: number;
}

export interface Message<TPayload = unknown> {
  id: string;
  type: string;
  source: string;
  target?: string;
  payload: TPayload;
  timestamp: number;
  correlationId?: string;
  meta?: Record<string, unknown>;
}

export interface MessageSubscription {
  type: string;
  source?: string | string[];
  handler: MessageHandler;
}

// ============================================================================
// Plugin Stores
// ============================================================================

export interface PluginStores {
  readable?: Record<string, Readable<unknown>>;
  writable?: Record<string, Writable<unknown>>;
  schemas?: Record<string, JSONSchema>;
}

// ============================================================================
// Plugin Definition
// ============================================================================

export interface PluginDefinition {
  manifest: PluginManifest;
  api?: PluginAPI;
  routes?: PluginRoutes;
  messages?: PluginMessages;
  stores?: PluginStores;
  setup?: (context: PluginContext) => Promise<void> | void;
  teardown?: () => Promise<void> | void;
}

export interface PluginContext {
  core: CoreAPI;
  config: Record<string, unknown>;
  logger: Logger;
  storage: PluginStorage;
}

// ============================================================================
// Core API
// ============================================================================

export interface CoreAPI {
  messages: MessageBusAPI;
  plugins: PluginRegistryAPI;
  routes: RouteManagerAPI;
  state: StateManagerAPI;
  events: EventEmitterAPI;
}

export interface MessageBusAPI {
  send<T>(type: string, payload: T, options?: SendOptions): void;
  request<TReq, TRes>(type: string, payload: TReq, options?: RequestOptions): Promise<TRes>;
  subscribe(type: string, handler: MessageHandler, options?: SubscribeOptions): Unsubscribe;
  once<T>(type: string, options?: SubscribeOptions): Promise<Message<T>>;
}

export interface PluginRegistryAPI {
  get(id: string): PluginInstance | undefined;
  getAll(): PluginInstance[];
  has(id: string): boolean;
  call<T>(pluginId: string, method: string, input?: unknown): Promise<T>;
  onPluginLoad(handler: (plugin: PluginInstance) => void): Unsubscribe;
  onPluginUnload(handler: (pluginId: string) => void): Unsubscribe;
}

export interface RouteManagerAPI {
  register(routes: RouteDefinition[]): void;
  unregister(paths: string[]): void;
  navigate(path: string, options?: NavigateOptions): Promise<void>;
  getCurrentRoute(): RouteInfo;
}

export interface StateManagerAPI {
  get<T>(key: string): T | undefined;
  set<T>(key: string, value: T): void;
  subscribe<T>(key: string, handler: (value: T) => void): Unsubscribe;
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
