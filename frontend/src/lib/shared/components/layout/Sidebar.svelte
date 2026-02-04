<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { getContext, onMount, onDestroy } from 'svelte';
  import { type Writable } from 'svelte/store';
  import type { Core, PluginInstance, PluginManifestEntry } from '$lib/core';
  import UserProfileDropdown from './UserProfileDropdown.svelte';
  import NotificationCenter from './NotificationCenter.svelte';

  export let collapsed = false;
  export let pluginManifests: PluginManifestEntry[] = [];

  const core = getContext<Core>('core');
  const pluginsLoaded = getContext<Writable<boolean>>('pluginsLoaded');

  interface NavItem {
    id: string;
    label: string;
    path: string;
    icon: string;
    status: string;
    priority: number;
  }

  interface Category {
    name: string;
    priority: number;
    items: NavItem[];
  }

  let navItems: NavItem[] = [];
  let categories: Category[] = [];
  let currentPath = '/dashboard';
  let unsubscribe: (() => void) | null = null;

  page.subscribe(p => {
    currentPath = p.url.pathname;
  });

  function getIcon(name: string): string {
    const icons: Record<string, string> = {
      'dashboard': '🏠',
      'menu': '📦',
      'modules': '📦',
      'clients': '👥',
      'users': '👤',
      'products': '📦',
      'inventory': '📦',
      'warehouse': '🏭',
      'orders': '🛒',
      'invoices': '📄',
      'payments': '💳',
      'documents': '📁',
      'settings': '⚙️'
    };
    return icons[name] || '📌';
  }

  function getCategoryForModule(id: string): { name: string; priority: number } {
    if (id === 'dashboard' || id === 'menu') {
      return { name: 'Core', priority: 1 };
    }
    if (['clients', 'users', 'products'].includes(id)) {
      return { name: 'Management', priority: 2 };
    }
    if (['inventory', 'warehouse', 'orders', 'invoices', 'payments'].includes(id)) {
      return { name: 'Operations', priority: 3 };
    }
    if (['settings', 'documents'].includes(id)) {
      return { name: 'Settings', priority: 4 };
    }
    return { name: 'Other', priority: 5 };
  }

  function buildNavFromPlugins(plugins: PluginInstance[]) {
    navItems = plugins
      .filter(p => p.status === 'enabled' || p.status === 'loaded')
      .map(p => {
        const category = getCategoryForModule(p.manifest.id);
        return {
          id: p.manifest.id,
          label: p.manifest.name,
          path: p.routes?.basePath || `/${p.manifest.id}`,
          icon: getIcon(p.manifest.id),
          status: p.status,
          priority: p.manifest.priority || category.priority * 10
        };
      })
      .sort((a, b) => a.priority - b.priority);

    const categoryMap = new Map<string, Category>();
    
    for (const item of navItems) {
      const catInfo = getCategoryForModule(item.id);
      
      if (!categoryMap.has(catInfo.name)) {
        categoryMap.set(catInfo.name, {
          name: catInfo.name,
          priority: catInfo.priority,
          items: []
        });
      }
      
      categoryMap.get(catInfo.name)!.items.push(item);
    }
    
    categories = Array.from(categoryMap.values())
      .sort((a, b) => a.priority - b.priority);
  }

  function buildNavFromManifests(manifests: PluginManifestEntry[]) {
    navItems = manifests
      .filter(m => m.enabled !== false)
      .map(m => {
        const category = getCategoryForModule(m.id);
        const name = m.id.charAt(0).toUpperCase() + m.id.slice(1);
        return {
          id: m.id,
          label: name,
          path: `/${m.id}`,
          icon: getIcon(m.id),
          status: 'loaded',
          priority: category.priority * 10
        };
      })
      .sort((a, b) => a.priority - b.priority);

    const categoryMap = new Map<string, Category>();
    
    for (const item of navItems) {
      const catInfo = getCategoryForModule(item.id);
      
      if (!categoryMap.has(catInfo.name)) {
        categoryMap.set(catInfo.name, {
          name: catInfo.name,
          priority: catInfo.priority,
          items: []
        });
      }
      
      categoryMap.get(catInfo.name)!.items.push(item);
    }
    
    categories = Array.from(categoryMap.values())
      .sort((a, b) => a.priority - b.priority);
  }

  function updateNavItems() {
    if (!core) return;
    
    try {
      const plugins = core.registry.getAll();
      if (plugins && plugins.length > 0) {
        buildNavFromPlugins(plugins);
      }
    } catch (e) {
      console.warn('Registry not ready yet');
    }
  }

  function isActive(path: string): boolean {
    if (path === '/dashboard') {
      return currentPath === '/dashboard' || currentPath === '/';
    }
    return currentPath.startsWith(path);
  }

  function navigate(path: string) {
    goto(path);
  }

  onMount(() => {
    unsubscribe = core.registry.getStore().subscribe(() => {
      updateNavItems();
    });
  });

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });

  $: if (pluginManifests && pluginManifests.length > 0 && navItems.length === 0) {
    buildNavFromManifests(pluginManifests);
  }

  $: if ($pluginsLoaded && navItems.length === 0) {
    updateNavItems();
  }
</script>

<aside class="sidebar" class:collapsed>
  <div class="sidebar-header">
    <div class="logo" on:click={() => navigate('/dashboard')} on:keypress={() => navigate('/dashboard')} role="button" tabindex="0">
      <span class="logo-icon">🏢</span>
      {#if !collapsed}
        <span class="logo-text">IMS ERP</span>
      {/if}
    </div>
  </div>

  <nav class="sidebar-nav">
    {#each categories as category}
      {#if !collapsed}
        <div class="category-label">{category.name}</div>
      {/if}
      <ul class="nav-list">
        {#each category.items as item}
          <li class="nav-item">
            <button
              class="nav-button"
              class:active={isActive(item.path)}
              on:click={() => navigate(item.path)}
              title={collapsed ? item.label : undefined}
            >
              <span class="nav-icon">{item.icon}</span>
              {#if !collapsed}
                <span class="nav-label">{item.label}</span>
              {/if}
            </button>
          </li>
        {/each}
      </ul>
    {/each}

    {#if navItems.length === 0}
      <div class="empty-state">
        {#if collapsed}
          <span class="empty-icon">📭</span>
        {:else}
          <p class="empty-text">No modules loaded</p>
        {/if}
      </div>
    {/if}
  </nav>

  <div class="sidebar-footer">
    <div class="footer-sections">
      <!-- Notification Center -->
      <NotificationCenter {collapsed} />
      
      <!-- User Profile Dropdown -->
      <UserProfileDropdown {collapsed} />
    </div>

    <button
      class="collapse-button"
      on:click={() => collapsed = !collapsed}
      title={collapsed ? 'Expand' : 'Collapse'}
    >
      <span class="collapse-icon">{collapsed ? '→' : '←'}</span>
      {#if !collapsed}
        <span class="collapse-text">Collapse</span>
      {/if}
    </button>
  </div>
</aside>

<style>
  .sidebar {
    width: 260px;
    height: 100vh;
    background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
    display: flex;
    flex-direction: column;
    position: fixed;
    left: 0;
    top: 0;
    z-index: 100;
    transition: width 0.3s ease;
  }

  .sidebar.collapsed {
    width: 72px;
  }

  .sidebar-header {
    padding: 1.25rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    cursor: pointer;
  }

  .logo:hover .logo-text {
    color: #60a5fa;
  }

  .logo-icon {
    font-size: 1.5rem;
  }

  .logo-text {
    font-size: 1.25rem;
    font-weight: 700;
    color: white;
    white-space: nowrap;
    transition: color 0.2s;
  }

  .sidebar-nav {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem 0;
  }

  .category-label {
    padding: 0.5rem 1.25rem;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: rgba(255, 255, 255, 0.4);
  }

  .nav-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .nav-item {
    margin: 0;
    padding: 0 0.75rem;
  }

  .nav-button {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.625rem 0.75rem;
    border: none;
    background: transparent;
    color: rgba(255, 255, 255, 0.7);
    cursor: pointer;
    border-radius: 0.375rem;
    transition: all 0.2s ease;
    text-align: left;
  }

  .nav-button:hover {
    background: rgba(255, 255, 255, 0.1);
    color: white;
  }

  .nav-button.active {
    background: rgba(59, 130, 246, 0.2);
    color: #60a5fa;
  }

  .nav-icon {
    font-size: 1.125rem;
    width: 24px;
    text-align: center;
    flex-shrink: 0;
  }

  .nav-label {
    font-size: 0.8125rem;
    font-weight: 500;
    white-space: nowrap;
  }

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem 1rem;
    text-align: center;
  }

  .empty-icon {
    font-size: 1.5rem;
    opacity: 0.5;
  }

  .empty-text {
    font-size: 0.8125rem;
    color: rgba(255, 255, 255, 0.5);
    margin: 0;
  }

  .sidebar-footer {
    padding: 1rem;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .footer-sections {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .collapse-button {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.5rem;
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: transparent;
    color: rgba(255, 255, 255, 0.7);
    cursor: pointer;
    border-radius: 0.375rem;
    transition: all 0.2s ease;
  }

  .collapse-button:hover {
    background: rgba(255, 255, 255, 0.1);
    color: white;
    border-color: rgba(255, 255, 255, 0.3);
  }

  .collapse-icon {
    font-size: 0.75rem;
  }

  .collapse-text {
    font-size: 0.75rem;
    white-space: nowrap;
  }

  .sidebar.collapsed .nav-button {
    justify-content: center;
    padding: 0.625rem;
  }

  .sidebar.collapsed .collapse-button {
    justify-content: center;
    padding: 0.5rem;
  }

  @media (max-width: 768px) {
    .sidebar {
      width: 72px;
    }

    .sidebar .logo-text,
    .sidebar .nav-label,
    .sidebar .collapse-text,
    .sidebar .category-label {
      display: none;
    }
  }
</style>
