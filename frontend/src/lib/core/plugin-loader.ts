// src/lib/core/plugin-loader.ts

import type { PluginDefinition } from './types';

export interface PluginLoader {
  loadFromManifests(manifests: PluginManifestEntry[]): Promise<PluginDefinition[]>;
  loadPlugin(id: string, source: PluginSource): Promise<PluginDefinition>;
  discoverPlugins(basePath: string): Promise<PluginManifestEntry[]>;
  getLoaded(): Map<string, PluginDefinition>;
}

export interface PluginManifestEntry {
  id: string;
  source: PluginSource;
  enabled?: boolean;
}

export type PluginSource = 
  | { type: 'builtin'; path: string }
  | { type: 'external'; url: string }
  | { type: 'local'; path: string }
  | { type: 'dynamic'; loader: () => Promise<PluginDefinition> };

export function createPluginLoader(): PluginLoader {
  const loadedPlugins: Map<string, PluginDefinition> = new Map();
  const pluginModules: Map<string, () => Promise<{ default: PluginDefinition }>> = new Map();

  // Register built-in plugin loaders
  function registerBuiltinPlugins(plugins: Record<string, () => Promise<{ default: PluginDefinition }>>) {
    for (const [path, loader] of Object.entries(plugins)) {
      pluginModules.set(path, loader);
    }
  }

  async function loadBuiltinPlugin(path: string): Promise<PluginDefinition> {
    const loader = pluginModules.get(path);
    if (!loader) {
      // Try dynamic import for development
      try {
        const module = await import(/* @vite-ignore */ `/src/lib/plugins/${path}/index.ts`);
        return module.default as PluginDefinition;
      } catch (e) {
        throw new Error(`Built-in plugin not found: ${path}`);
      }
    }
    const module = await loader();
    return module.default;
  }

  async function loadExternalPlugin(url: string): Promise<PluginDefinition> {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`Failed to fetch plugin from ${url}: ${response.statusText}`);
    }
    
    const code = await response.text();
    
    // Create sandboxed execution
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
    const module = await import(/* @vite-ignore */ path);
    return module.default as PluginDefinition;
  }

  async function loadDynamicPlugin(loader: () => Promise<PluginDefinition>): Promise<PluginDefinition> {
    return await loader();
  }

  function validatePlugin(plugin: PluginDefinition): void {
    if (!plugin.manifest?.id) {
      throw new Error('Plugin missing required manifest.id');
    }
    if (!plugin.manifest?.name) {
      throw new Error('Plugin missing required manifest.name');
    }
    if (!plugin.manifest?.version) {
      throw new Error('Plugin missing required manifest.version');
    }
  }

  function topologicalSort(manifests: PluginManifestEntry[]): PluginManifestEntry[] {
    const graph = new Map<string, Set<string>>();
    const inDegree = new Map<string, number>();
    
    // Initialize
    for (const manifest of manifests) {
      graph.set(manifest.id, new Set());
      inDegree.set(manifest.id, 0);
    }
    
    // Build dependency graph (would need to load manifests first for full implementation)
    // For now, return as-is sorted by enabled status
    return [...manifests].sort((a, b) => {
      if (a.enabled === false && b.enabled !== false) return 1;
      if (a.enabled !== false && b.enabled === false) return -1;
      return 0;
    });
  }

  return {
    async loadFromManifests(manifests: PluginManifestEntry[]): Promise<PluginDefinition[]> {
      const results: PluginDefinition[] = [];
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
        case 'dynamic':
          plugin = await loadDynamicPlugin(source.loader);
          break;
        default:
          throw new Error(`Unknown plugin source type`);
      }

      validatePlugin(plugin);
      loadedPlugins.set(id, plugin);
      
      return plugin;
    },

    async discoverPlugins(basePath: string): Promise<PluginManifestEntry[]> {
      try {
        const response = await fetch(`${basePath}/manifest.json`);
        if (!response.ok) {
          console.warn(`No plugin manifest found at ${basePath}/manifest.json`);
          return [];
        }
        return response.json();
      } catch (error) {
        console.error('Failed to discover plugins:', error);
        return [];
      }
    },

    getLoaded(): Map<string, PluginDefinition> {
      return new Map(loadedPlugins);
    }
  };
}

// Helper to create a plugin manifest entry
export function createPluginEntry(
  id: string,
  source: PluginSource,
  enabled = true
): PluginManifestEntry {
  return { id, source, enabled };
}

// Helper for built-in plugins using Vite's import.meta.glob
export function createBuiltinEntries(
  modules: Record<string, () => Promise<{ default: PluginDefinition }>>
): PluginManifestEntry[] {
  return Object.keys(modules).map(path => {
    // Extract plugin id from path like './dashboard/index.ts'
    const match = path.match(/\.\/([^/]+)\//);
    const id = match ? match[1] : path;
    
    return {
      id,
      source: { 
        type: 'dynamic' as const, 
        loader: async () => {
          const module = await modules[path]();
          return module.default;
        }
      },
      enabled: true
    };
  });
}
