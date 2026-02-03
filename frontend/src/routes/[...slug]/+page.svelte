<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import type { PageData } from './$types';

  export let data: PageData;

  let matchedRoute: any = null;
  let params: Record<string, string> = {};
  let loading = true;
  let notFound = false;
  let PluginComponent: any = null;
  let clientLoaded = false;

  $: currentPath = $page.url.pathname;

  function getRoutesForPlugin(pluginId: string): any {
    const routes: Record<string, any> = {
      dashboard: {
        basePath: '/dashboard',
        routes: [
          { path: '/', meta: { title: 'Dashboard' } }
        ]
      },
      clients: {
        basePath: '/clients',
        routes: [
          { path: '/', meta: { title: 'Clients' } },
          { path: '/new', meta: { title: 'New Client' } },
          { path: '/:id', meta: { title: 'Client Details' } },
          { path: '/:id/edit', meta: { title: 'Edit Client' } }
        ]
      },
      users: {
        basePath: '/users',
        routes: [
          { path: '/', meta: { title: 'Users' } },
          { path: '/new', meta: { title: 'New User' } },
          { path: '/:id', meta: { title: 'User Details' } },
          { path: '/:id/edit', meta: { title: 'Edit User' } }
        ]
      },
      products: {
        basePath: '/products',
        routes: [
          { path: '/', meta: { title: 'Products' } },
          { path: '/new', meta: { title: 'New Product' } },
          { path: '/:id', meta: { title: 'Product Details' } }
        ]
      },
      inventory: {
        basePath: '/inventory',
        routes: [
          { path: '/', meta: { title: 'Inventory' } },
          { path: '/:id', meta: { title: 'Inventory Item' } }
        ]
      },
      warehouse: {
        basePath: '/warehouse',
        routes: [
          { path: '/', meta: { title: 'Warehouse' } },
          { path: '/:id', meta: { title: 'Warehouse Details' } }
        ]
      },
      orders: {
        basePath: '/orders',
        routes: [
          { path: '/', meta: { title: 'Orders' } },
          { path: '/new', meta: { title: 'New Order' } },
          { path: '/:id', meta: { title: 'Order Details' } }
        ]
      },
      invoices: {
        basePath: '/invoices',
        routes: [
          { path: '/', meta: { title: 'Invoices' } },
          { path: '/new', meta: { title: 'New Invoice' } },
          { path: '/:id', meta: { title: 'Invoice Details' } }
        ]
      },
      payments: {
        basePath: '/payments',
        routes: [
          { path: '/', meta: { title: 'Payments' } },
          { path: '/new', meta: { title: 'New Payment' } },
          { path: '/:id', meta: { title: 'Payment Details' } }
        ]
      },
      documents: {
        basePath: '/documents',
        routes: [
          { path: '/', meta: { title: 'Documents' } },
          { path: '/:id', meta: { title: 'Document Details' } }
        ]
      },
      settings: {
        basePath: '/settings',
        routes: [
          { path: '/', meta: { title: 'Settings' } }
        ]
      },
      menu: {
        basePath: '/menu',
        routes: [
          { path: '/', meta: { title: 'Menu' } }
        ]
      }
    };
    
    return routes[pluginId] || { basePath: `/${pluginId}`, routes: [{ path: '/', meta: { title: pluginId } }] };
  }

  function findMatchingRoute(path: string): { route: any; pluginId: string; routePath: string; params: Record<string, string> } | null {
    if (!data?.pluginManifests) return null;
    
    for (const manifest of data.pluginManifests) {
      const pluginRoutes = getRoutesForPlugin(manifest.id);
      const basePath = pluginRoutes.basePath;

      for (const route of pluginRoutes.routes) {
        const fullPath = route.path.startsWith('/')
          ? `${basePath}${route.path}`
          : `${basePath}/${route.path}`;

        const result = matchPath(fullPath, path);
        if (result) {
          return { route, pluginId: manifest.id, routePath: route.path, params: result };
        }
      }
    }

    return null;
  }

  function matchPath(pattern: string, path: string): Record<string, string> | null {
    const normalizedPattern = pattern.replace(/\/+$/, '');
    const normalizedPath = path.replace(/\/+$/, '');
    
    const patternParts = normalizedPattern.split('/').filter(Boolean);
    const pathParts = normalizedPath.split('/').filter(Boolean);

    if (patternParts.length !== pathParts.length) {
      return null;
    }

    const params: Record<string, string> = {};

    for (let i = 0; i < patternParts.length; i++) {
      const patternPart = patternParts[i];
      const pathPart = pathParts[i];

      if (patternPart.startsWith('[') && patternPart.endsWith(']')) {
        const paramName = patternPart.slice(1, -1);
        params[paramName] = pathPart;
      } else if (patternPart.startsWith(':')) {
        const paramName = patternPart.slice(1);
        params[paramName] = pathPart;
      } else if (patternPart !== pathPart) {
        return null;
      }
    }

    return params;
  }

  async function loadPluginComponent(pluginId: string, routePath: string): Promise<any> {
    const normalizedRoutePath = routePath === '/' ? '' : routePath.replace(/\/$/, '');
    const componentPath = `/src/lib/plugins/${pluginId}/routes${normalizedRoutePath}/+page.svelte`;

    try {
      const module = await import(/* @vite-ignore */ componentPath);
      return module.default;
    } catch (e) {
      console.error(`Failed to load component for ${pluginId}${routePath}:`, e);
      return null;
    }
  }

  async function updateRoute() {
    const path = (currentPath || '/') as string;

    if (path === '/') {
      loading = false;
      return;
    }

    if (path === '/dashboard') {
      if (typeof window !== 'undefined') {
        window.location.href = '/';
      }
      loading = false;
      return;
    }

    const result = findMatchingRoute(path);

    if (result) {
      matchedRoute = result.route;
      params = result.params;
      notFound = false;

      if (clientLoaded) {
        const Component = await loadPluginComponent(result.pluginId, result.routePath);
        if (Component) {
          PluginComponent = Component;
        } else {
          notFound = true;
        }
      }
    } else {
      matchedRoute = null;
      params = {};
      notFound = true;
      PluginComponent = null;
    }

    loading = false;
  }

  onMount(() => {
    clientLoaded = true;
    if (data?.pluginManifests) {
      updateRoute();
    }
  });

  $: if (browser && data?.pluginManifests && currentPath && !PluginComponent) {
    updateRoute();
  }
</script>

<svelte:head>
  {#if matchedRoute?.meta?.title}
    <title>{matchedRoute.meta.title} | ERP System</title>
  {:else if notFound}
    <title>Page Not Found | ERP System</title>
  {/if}
</svelte:head>

{#if loading}
  <div class="loading-container">
    <div class="loading-content">
      <div class="spinner"></div>
      <p>Loading...</p>
    </div>
  </div>
{:else if notFound}
  <div class="not-found-container">
    <div class="not-found-content">
      <h1>404</h1>
      <p>Page not found</p>
      <p class="path">{currentPath}</p>
      <a href="/dashboard" class="back-link">Back to Dashboard</a>
    </div>
  </div>
{:else if matchedRoute && PluginComponent}
  <svelte:component this={PluginComponent} {params} />
{:else if matchedRoute}
  <div class="loading-container">
    <div class="loading-content">
      <div class="spinner"></div>
      <p>Loading {matchedRoute.meta?.title || 'page'}...</p>
    </div>
  </div>
{:else}
  <div class="not-found-container">
    <div class="not-found-content">
      <h1>404</h1>
      <p>Page not found</p>
      <p class="path">{currentPath}</p>
      <a href="/dashboard" class="back-link">Back to Dashboard</a>
    </div>
  </div>
{/if}

<style>
  .loading-container,
  .not-found-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
  }

  .loading-content,
  .not-found-content {
    text-align: center;
  }

  .spinner {
    width: 48px;
    height: 48px;
    border: 4px solid rgba(59, 130, 246, 0.2);
    border-top-color: #3b82f6;
    border-radius: 50%;
    margin: 0 auto 16px;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  h1 {
    font-size: 6rem;
    font-weight: 700;
    color: var(--color-gray-200, #e5e7eb);
    margin: 0;
    line-height: 1;
  }

  p {
    font-size: 1.25rem;
    color: var(--color-gray-600, #4b5563);
    margin: 1rem 0 0;
  }

  .path {
    font-size: 0.875rem;
    color: var(--color-gray-400, #9ca3af);
    font-family: monospace;
  }

  .back-link {
    display: inline-block;
    margin-top: 1.5rem;
    padding: 0.75rem 1.5rem;
    background: #3b82f6;
    color: white;
    border-radius: 0.5rem;
    text-decoration: none;
    font-weight: 500;
    transition: background 0.2s;
  }

  .back-link:hover {
    background: #2563eb;
  }
</style>
