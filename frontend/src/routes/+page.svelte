<script lang="ts">
  import { getContext } from 'svelte';
  import type { Core, PluginRoutes } from '$lib/core';

  interface PluginWithRoutes {
    manifest: {
      name: string;
      description?: string;
      version?: string;
    };
    routes: PluginRoutes;
  }

  const core = getContext<Core>('core');
  
  let plugins: PluginWithRoutes[] = [];
  
  // Subscribe to plugins
  function updatePlugins() {
    plugins = core.registry.getAll() as PluginWithRoutes[];
  }
  
  // Use reactive statement for Svelte 4 compatibility
  $: {
    core.registry.getAll();
    updatePlugins();
  }
</script>

<svelte:head>
  <title>ERP System</title>
</svelte:head>

<div class="p-6">
  <div class="mb-8">
    <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">Welcome to ERP System</h1>
    <p class="text-gray-600 dark:text-gray-400">
      A modular enterprise resource planning system built with SvelteKit.
    </p>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
    {#each plugins as plugin}
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
          {plugin.manifest.name}
        </h2>
        <p class="text-gray-600 dark:text-gray-400 text-sm mb-4">
          {plugin.manifest.description || 'No description'}
        </p>
        <div class="flex items-center justify-between">
          <span class="text-xs text-gray-500 dark:text-gray-400">
            Version {plugin.manifest.version || '1.0.0'}
          </span>
          {#if plugin.routes && Array.isArray(plugin.routes) && plugin.routes.length > 0}
            <a
              href={plugin.routes[0].path}
              class="text-primary-600 hover:text-primary-700 text-sm font-medium"
            >
              Open →
            </a>
          {/if}
        </div>
      </div>
    {/each}
  </div>

  {#if plugins.length === 0}
    <div class="text-center py-12">
      <div class="mx-auto w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center mb-4">
        <svg
          class="w-8 h-8 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
          />
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No plugins loaded</h3>
      <p class="text-gray-500 dark:text-gray-400">
        Plugins will appear here once they are loaded.
      </p>
    </div>
  {/if}
</div>
