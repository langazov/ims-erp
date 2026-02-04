<!-- src/routes/+layout.svelte -->
<script lang="ts">
  import { onMount, setContext, onDestroy } from 'svelte';
  import { writable, type Writable } from 'svelte/store';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { createCore, type Core, type PluginManifestEntry } from '$lib/core';
  import type { LayoutData } from './$types';
  import Sidebar from '$lib/shared/components/layout/Sidebar.svelte';
  import GlobalSearch from '$lib/shared/components/search/GlobalSearch.svelte';
  import { auth } from '$lib/shared/stores/auth';
  import '../app.css';
  
  export let data: LayoutData;
  
  let core: Core | null = null;
  const pluginsLoaded: Writable<boolean> = writable(false);
  const loadingError: Writable<Error | null> = writable(null);
  const loadedPluginIds: Writable<string[]> = writable([]);
  
  let sidebarCollapsed = false;
  let authInitialized = false;
  let showGlobalSearch = false;
  
  $: path = String($page.url.pathname);
  $: isAuthPage = path === '/login' || path.startsWith('/login/') || 
                  path === '/register' || path.startsWith('/register/') ||
                  path === '/forgot-password' || path.startsWith('/forgot-password/');
  
  onMount(() => {
    if (isAuthPage) {
      authInitialized = true;
      return;
    }
    
    core = createCore();
    
    setContext('core', core);
    setContext('coreApi', core.api);
    setContext('pluginsLoaded', pluginsLoaded);
    setContext('messages', core.messages);
    setContext('routes', core.routes);
    setContext('state', core.state);
    setContext('events', core.events);
    
    const authState = auth.init();
    authInitialized = true;

    if (!authState.isAuthenticated) {
      goto('/login');
      return;
    }

    initializePlugins();
  });

  async function initializePlugins() {
    if (!core) return;
    
    try {
      await core.initialize(data.pluginManifests);
      
      const plugins = core.registry.getAll();
      loadedPluginIds.set(plugins.map(p => p.manifest.id));
      
      pluginsLoaded.set(true);
      
      console.log('[App] Plugins loaded:', $loadedPluginIds);
    } catch (error) {
      console.error('[App] Failed to initialize plugins:', error);
      loadingError.set(error as Error);
    }
  }
  
  $: if ($pluginsLoaded && core) {
    core.registry.onLoad((plugin) => {
      loadedPluginIds.update(ids => [...ids, plugin.manifest.id]);
    });
    
    core.registry.onUnload((pluginId) => {
      loadedPluginIds.update(ids => ids.filter(id => id !== pluginId));
    });
  }

  onDestroy(async () => {
    if (core) {
      await core.destroy();
    }
  });

  // Global keyboard shortcut for search
  function handleKeydown(event: KeyboardEvent) {
    // Cmd/Ctrl + K to open search
    if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
      event.preventDefault();
      showGlobalSearch = true;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<svelte:head>
  <title>SvelteKit Plugin System</title>
  <meta name="description" content="Modular SvelteKit application with dynamic plugin loading" />
</svelte:head>

{#if isAuthPage}
  <slot />
{:else if $loadingError}
  <div class="error-container">
    <div class="error-content">
      <h1>Failed to Load Application</h1>
      <p>{$loadingError.message}</p>
      <button onclick={() => window.location.reload()}>
        Retry
      </button>
    </div>
  </div>
{:else if $pluginsLoaded}
  <div class="app-layout">
    <Sidebar bind:collapsed={sidebarCollapsed} pluginManifests={data.pluginManifests} />
    <main class="main-content" class:sidebar-collapsed={sidebarCollapsed}>
      <!-- Search Trigger Button -->
      <button 
        class="fixed bottom-6 right-6 z-40 w-14 h-14 bg-primary-600 hover:bg-primary-700 text-white rounded-full shadow-lg flex items-center justify-center transition-all hover:scale-110"
        onclick={() => showGlobalSearch = true}
        title="Search (Ctrl+K)"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </button>
      
      <slot />
    </main>
  </div>
  
  <!-- Global Search Modal -->
  <GlobalSearch bind:isOpen={showGlobalSearch} />
{:else if authInitialized}
  <div class="loading-container">
    <div class="loading-content">
      <div class="spinner"></div>
      <p>Loading plugins...</p>
    </div>
  </div>
{:else}
  <div class="loading-container">
    <div class="loading-content">
      <div class="spinner"></div>
      <p>Loading...</p>
    </div>
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
  
  .app-layout {
    display: flex;
    min-height: 100vh;
  }
  
  .main-content {
    flex: 1;
    margin-left: 260px;
    transition: margin-left 0.3s ease;
    background: #f8fafc;
    min-height: 100vh;
  }
  
  .main-content.sidebar-collapsed {
    margin-left: 72px;
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
