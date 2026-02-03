import { writable, derived } from 'svelte/store';
import type { PluginStores, PluginInstance, PluginManifest } from '$lib/core/types';

interface MenuState {
  modules: ModuleInfo[];
  expandedCategories: string[];
  searchQuery: string;
  loading: boolean;
  error: string | null;
}

export interface ModuleInfo {
  id: string;
  name: string;
  version: string;
  description: string;
  author?: string;
  icon?: string;
  status: 'enabled' | 'disabled' | 'error';
  priority: number;
  category: string;
  route?: string;
}

const initialState: MenuState = {
  modules: [],
  expandedCategories: [],
  searchQuery: '',
  loading: false,
  error: null
};

function createMenuStore() {
  const { subscribe, set, update } = writable<MenuState>(initialState);

  function categorizeModules(modules: ModuleInfo[]): Record<string, ModuleInfo[]> {
    const categories: Record<string, ModuleInfo[]> = {
      'Core': [],
      'Management': [],
      'Operations': [],
      'Settings': [],
      'Other': []
    };

    for (const module of modules) {
      if (module.category === 'Core' || module.id === 'dashboard' || module.id === 'menu') {
        categories['Core'].push(module);
      } else if (['clients', 'users', 'products'].includes(module.id)) {
        categories['Management'].push(module);
      } else if (['inventory', 'warehouse', 'orders', 'invoices', 'payments'].includes(module.id)) {
        categories['Operations'].push(module);
      } else if (['settings', 'documents'].includes(module.id)) {
        categories['Settings'].push(module);
      } else {
        categories['Other'].push(module);
      }
    }

    return categories;
  }

  return {
    subscribe,

    async loadModules(plugins: PluginInstance[]) {
      update(state => ({ ...state, loading: true, error: null }));

      try {
        const modules: ModuleInfo[] = plugins.map(plugin => ({
          id: plugin.manifest.id,
          name: plugin.manifest.name,
          version: plugin.manifest.version,
          description: plugin.manifest.description,
          author: plugin.manifest.author,
          icon: typeof plugin.manifest.icon === 'string' ? plugin.manifest.icon : undefined,
          status: plugin.status,
          priority: plugin.manifest.priority || 0,
          category: categorizeModules([])[''] ? 'Other' : 'Other',
          route: plugin.routes?.basePath
        }));

        update(state => ({
          ...state,
          modules,
          loading: false
        }));
      } catch (error) {
        update(state => ({
          ...state,
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to load modules'
        }));
      }
    },

    setSearchQuery(query: string) {
      update(state => ({ ...state, searchQuery: query }));
    },

    toggleCategory(category: string) {
      update(state => {
        const expanded = state.expandedCategories.includes(category)
          ? state.expandedCategories.filter(c => c !== category)
          : [...state.expandedCategories, category];
        return { ...state, expandedCategories: expanded };
      });
    },

    expandAll() {
      update(state => ({
        ...state,
        expandedCategories: Object.keys(categorizeModules(state.modules))
      }));
    },

    collapseAll() {
      update(state => ({
        ...state,
        expandedCategories: []
      }));
    },

    clearError() {
      update(state => ({ ...state, error: null }));
    },

    reset() {
      set(initialState);
    }
  };
}

export const menuStore = createMenuStore();

export const stores: PluginStores = {
  readable: {
    modules: derived(menuStore, $state => $state.modules),
    searchQuery: derived(menuStore, $state => $state.searchQuery),
    expandedCategories: derived(menuStore, $state => $state.expandedCategories),
    loading: derived(menuStore, $state => $state.loading),
    error: derived(menuStore, $state => $state.error)
  },
  writable: {}
};

export { categorizeModules };
