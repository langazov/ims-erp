// src/lib/core/plugin-registry.ts

import { writable, get, type Readable, type Writable } from 'svelte/store';
import type {
  PluginDefinition,
  PluginInstance,
  PluginManifest,
  PluginStatus,
  PluginContext,
  CoreAPI,
  Unsubscribe,
  Logger,
  PluginStorage
} from './types';

export interface PluginRegistry {
  register(definition: PluginDefinition): Promise<void>;
  unregister(id: string): Promise<void>;
  enable(id: string): Promise<void>;
  disable(id: string): Promise<void>;
  get(id: string): PluginInstance | undefined;
  getAll(): PluginInstance[];
  has(id: string): boolean;
  getStore(): Readable<Map<string, PluginInstance>>;
  onLoad(handler: (plugin: PluginInstance) => void): Unsubscribe;
  onUnload(handler: (pluginId: string) => void): Unsubscribe;
  onError(handler: (pluginId: string, error: Error) => void): Unsubscribe;
}

function createLogger(pluginId: string): Logger {
  const prefix = `[${pluginId}]`;
  return {
    debug: (msg, ...args) => console.debug(prefix, msg, ...args),
    info: (msg, ...args) => console.info(prefix, msg, ...args),
    warn: (msg, ...args) => console.warn(prefix, msg, ...args),
    error: (msg, ...args) => console.error(prefix, msg, ...args)
  };
}

function createPluginStorage(pluginId: string): PluginStorage {
  const storageKey = `plugin:${pluginId}:`;
  
  return {
    async get<T>(key: string): Promise<T | undefined> {
      try {
        const value = localStorage.getItem(storageKey + key);
        return value ? JSON.parse(value) : undefined;
      } catch {
        return undefined;
      }
    },
    async set<T>(key: string, value: T): Promise<void> {
      localStorage.setItem(storageKey + key, JSON.stringify(value));
    },
    async delete(key: string): Promise<void> {
      localStorage.removeItem(storageKey + key);
    },
    async clear(): Promise<void> {
      const keys = Object.keys(localStorage).filter(k => k.startsWith(storageKey));
      keys.forEach(k => localStorage.removeItem(k));
    }
  };
}

function validateManifest(manifest: PluginManifest): { valid: boolean; errors: string[] } {
  const errors: string[] = [];
  
  if (!manifest.id) errors.push('Missing required field: id');
  if (!manifest.name) errors.push('Missing required field: name');
  if (!manifest.version) errors.push('Missing required field: version');
  
  if (manifest.id && !/^[a-z0-9-]+$/.test(manifest.id)) {
    errors.push('Plugin id must be kebab-case (lowercase letters, numbers, hyphens)');
  }
  
  if (manifest.version && !/^\d+\.\d+\.\d+/.test(manifest.version)) {
    errors.push('Version must be semver format (e.g., 1.0.0)');
  }
  
  return { valid: errors.length === 0, errors };
}

function validateDependencies(
  manifest: PluginManifest,
  loadedPlugins: Map<string, PluginInstance>
): { valid: boolean; missing: string[] } {
  const missing: string[] = [];
  
  for (const dep of manifest.dependencies ?? []) {
    if (!dep.optional && !loadedPlugins.has(dep.pluginId)) {
      missing.push(`${dep.pluginId}@${dep.version}`);
    }
  }
  
  return { valid: missing.length === 0, missing };
}

export function createPluginRegistry(coreAPI: CoreAPI): PluginRegistry {
  const plugins: Writable<Map<string, PluginInstance>> = writable(new Map());
  const definitions: Map<string, PluginDefinition> = new Map();
  const contexts: Map<string, PluginContext> = new Map();
  
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
      config: {},
      logger: createLogger(manifest.id),
      storage: createPluginStorage(manifest.id)
    };
  }

  return {
    async register(definition: PluginDefinition): Promise<void> {
      const { manifest } = definition;
      
      const validation = validateManifest(manifest);
      if (!validation.valid) {
        throw new Error(`Invalid manifest for ${manifest.id}: ${validation.errors.join(', ')}`);
      }

      if (definitions.has(manifest.id)) {
        throw new Error(`Plugin ${manifest.id} is already registered`);
      }

      const currentPlugins = get(plugins);
      const depValidation = validateDependencies(manifest, currentPlugins);
      if (!depValidation.valid) {
        throw new Error(`Missing dependencies for ${manifest.id}: ${depValidation.missing.join(', ')}`);
      }

      try {
        // Create initial instance
        const instance: PluginInstance = {
          manifest,
          status: 'loading',
          api: definition.api,
          routes: definition.routes,
          stores: definition.stores
        };

        plugins.update(map => {
          map.set(manifest.id, instance);
          return new Map(map);
        });

        await manifest.lifecycle?.onBeforeLoad?.();

        definitions.set(manifest.id, definition);

        const context = await createPluginContext(definition);
        contexts.set(manifest.id, context);

        await definition.setup?.(context);

        if (definition.routes) {
          coreAPI.routes.register(definition.routes.routes);
        }

        if (definition.messages) {
          for (const [type, handler] of Object.entries(definition.messages.handlers)) {
            coreAPI.messages.subscribe(
              `${manifest.id}:${type}`,
              handler,
              { source: '*' }
            );
          }
        }

        instance.status = manifest.enabled !== false ? 'enabled' : 'disabled';
        
        plugins.update(map => {
          map.set(manifest.id, instance);
          return new Map(map);
        });

        await manifest.lifecycle?.onLoad?.();

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

        await definition.manifest.lifecycle?.onUnload?.();
        await definition.teardown?.();

        if (definition.routes) {
          const paths = definition.routes.routes.map(r => 
            `${definition.routes!.basePath}${r.path}`
          );
          coreAPI.routes.unregister(paths);
        }

        definitions.delete(id);
        contexts.delete(id);
        
        plugins.update(map => {
          map.delete(id);
          return new Map(map);
        });

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
