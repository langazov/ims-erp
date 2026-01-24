<!-- src/routes/+layout.svelte -->
<script lang="ts">
  import { onMount, setContext } from 'svelte';
  import { writable, type Writable } from 'svelte/store';
  import { createCore, type Core, type PluginManifestEntry } from '$lib/core';
  import type { LayoutData } from './$types';
  import '../app.css';
  
  export let data: LayoutData;
  
  // Create core instance
  const core = createCore();
  
  // Loading state
  const pluginsLoaded: Writable<boolean> = writable(false);
  const loadingError: Writable<Error | null> = writable(null);
  const loadedPluginIds: Writable<string[]> = writable([]);
  
  // Set contexts for child components
  setContext('core', core);
  setContext('coreApi', core.api);
  setContext('pluginsLoaded', pluginsLoaded);
  setContext('messages', core.messages);
  setContext('routes', core.routes);
  setContext('state', core.state);
  setContext('events', core.events);
  
  onMount(async () => {
    try {
      // Initialize core with plugin manifests
      await core.initialize(data.pluginManifests);
      
      // Update loaded plugins list
      const plugins = core.registry.getAll();
      loadedPluginIds.set(plugins.map(p => p.manifest.id));
      
      pluginsLoaded.set(true);
      
      console.log('[App] Plugins loaded:', $loadedPluginIds);
    } catch (error) {
      console.error('[App] Failed to initialize plugins:', error);
      loadingError.set(error as Error);
    }
    
    // Cleanup on unmount
    return () => {
      core.destroy();
    };
  });
  
  // Subscribe to plugin lifecycle events
  $: if ($pluginsLoaded) {
    core.registry.onLoad((plugin) => {
      loadedPluginIds.update(ids => [...ids, plugin.manifest.id]);
    });
    
    core.registry.onUnload((pluginId) => {
      loadedPluginIds.update(ids => ids.filter(id => id !== pluginId));
    });
  }
</script>

<svelte:head>
  <title>SvelteKit Plugin System</title>
  <meta name="description" content="Modular SvelteKit application with dynamic plugin loading" />
</svelte:head>

{#if $loadingError}
  <div class="error-container">
    <div class="error-content">
      <h1>Failed to Load Application</h1>
      <p>{$loadingError.message}</p>
      <button on:click={() => window.location.reload()}>
        Retry
      </button>
    </div>
  </div>
{:else if !$pluginsLoaded}
  <div class="loading-container">
    <div class="loading-content">
      <div class="spinner"></div>
      <p>Loading plugins...</p>
    </div>
  </div>
{:else}
  <div class="app-container">
    <slot />
  </div>
{/if}

<style>
  :global(*, *::before, *::after) {
    box-sizing: border-box;
  }
  
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen,
      Ubuntu, Cantarell, 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
  
  .app-container {
    min-height: 100vh;
  }
  
  .loading-container,
  .error-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  }
  
  .loading-content,
  .error-content {
    text-align: center;
    color: white;
  }
  
  .spinner {
    width: 48px;
    height: 48px;
    border: 4px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    margin: 0 auto 16px;
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  
  .error-content h1 {
    margin: 0 0 8px;
    font-size: 24px;
  }
  
  .error-content p {
    margin: 0 0 16px;
    opacity: 0.9;
  }
  
  .error-content button {
    padding: 12px 24px;
    font-size: 16px;
    border: 2px solid white;
    background: transparent;
    color: white;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .error-content button:hover {
    background: white;
    color: #764ba2;
  }
</style>
