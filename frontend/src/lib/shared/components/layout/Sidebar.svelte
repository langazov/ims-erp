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
    iconSvg: string;
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

  // Lucide-style SVG icons (inline for zero-dep)
  function getIconSvg(name: string): string {
    const icons: Record<string, string> = {
      dashboard: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>`,
      clients: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"/>`,
      users: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>`,
      products: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"/>`,
      inventory: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/>`,
      warehouse: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>`,
      orders: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z"/>`,
      invoices: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2z"/>`,
      payments: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"/>`,
      documents: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>`,
      settings: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>`,
      menu: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M4 6h16M4 12h16M4 18h16"/>`,
      analytics: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>`,
    };
    const paths = icons[name] || `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M5 12h14M12 5l7 7-7 7"/>`;
    return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" class="nav-icon-svg">${paths}</svg>`;
  }

  function getCategoryForModule(id: string): { name: string; priority: number } {
    if (id === 'dashboard') return { name: '', priority: 1 };
    if (['clients', 'users', 'products'].includes(id)) return { name: 'Management', priority: 2 };
    if (['inventory', 'warehouse', 'orders', 'invoices', 'payments'].includes(id)) return { name: 'Operations', priority: 3 };
    if (['analytics'].includes(id)) return { name: 'Analytics', priority: 4 };
    if (['settings', 'documents'].includes(id)) return { name: 'System', priority: 5 };
    return { name: 'Other', priority: 6 };
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
          iconSvg: getIconSvg(p.manifest.id),
          status: p.status,
          priority: p.manifest.priority || category.priority * 10
        };
      })
      .sort((a, b) => a.priority - b.priority);
    buildCategories();
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
          iconSvg: getIconSvg(m.id),
          status: 'loaded',
          priority: category.priority * 10
        };
      })
      .sort((a, b) => a.priority - b.priority);
    buildCategories();
  }

  function buildCategories() {
    const categoryMap = new Map<string, Category>();
    for (const item of navItems) {
      const catInfo = getCategoryForModule(item.id);
      if (!categoryMap.has(catInfo.name)) {
        categoryMap.set(catInfo.name, { name: catInfo.name, priority: catInfo.priority, items: [] });
      }
      categoryMap.get(catInfo.name)!.items.push(item);
    }
    categories = Array.from(categoryMap.values()).sort((a, b) => a.priority - b.priority);
  }

  function updateNavItems() {
    if (!core) return;
    try {
      const plugins = core.registry.getAll();
      if (plugins && plugins.length > 0) buildNavFromPlugins(plugins);
    } catch (e) {
      console.warn('Registry not ready yet');
    }
  }

  function isActive(path: string): boolean {
    if (path === '/dashboard') return currentPath === '/dashboard' || currentPath === '/';
    return currentPath.startsWith(path);
  }

  onMount(() => {
    unsubscribe = core.registry.getStore().subscribe(() => updateNavItems());
  });

  onDestroy(() => { if (unsubscribe) unsubscribe(); });

  $: if (pluginManifests && pluginManifests.length > 0 && navItems.length === 0) {
    buildNavFromManifests(pluginManifests);
  }
  $: if ($pluginsLoaded && navItems.length === 0) updateNavItems();
</script>

<aside class="sidebar" class:collapsed>
  <!-- Logo / Brand -->
  <div class="sidebar-header">
    <button class="logo" on:click={() => goto('/dashboard')} title="Dashboard">
      <div class="logo-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" class="logo-svg">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
        </svg>
      </div>
      {#if !collapsed}
        <span class="logo-text">IMS ERP</span>
      {/if}
    </button>
  </div>

  <!-- Navigation -->
  <nav class="sidebar-nav">
    {#each categories as category}
      {#if !collapsed && category.name}
        <div class="category-label">{category.name}</div>
      {/if}
      <ul class="nav-list">
        {#each category.items as item}
          <li class="nav-item">
            <button
              class="nav-button"
              class:active={isActive(item.path)}
              on:click={() => goto(item.path)}
              title={item.label}
            >
              {@html item.iconSvg}
              {#if !collapsed}
                <span class="nav-label">{item.label}</span>
              {/if}
            </button>
          </li>
        {/each}
      </ul>
    {/each}
  </nav>

  <!-- Footer -->
  <div class="sidebar-footer">
    <NotificationCenter {collapsed} />
    <UserProfileDropdown {collapsed} />
    <button class="collapse-button" on:click={() => (collapsed = !collapsed)} title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}>
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" style="width:16px;height:16px">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="{collapsed ? 'M13 5l7 7-7 7M5 5l7 7-7 7' : 'M11 19l-7-7 7-7m8 14l-7-7 7-7'}"/>
      </svg>
      {#if !collapsed}<span class="collapse-text">Collapse</span>{/if}
    </button>
  </div>
</aside>

<style>
  :global(.nav-icon-svg) {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
  }
  .logo-svg {
    width: 22px;
    height: 22px;
    stroke: #60a5fa;
  }

  .sidebar {
    width: var(--sidebar-width, 260px);
    height: 100vh;
    background: linear-gradient(180deg, var(--sidebar-bg-gradient-start, #1e293b) 0%, var(--sidebar-bg, #0f172a) 100%);
    display: flex;
    flex-direction: column;
    position: fixed;
    left: 0;
    top: 0;
    z-index: 100;
    transition: width 0.25s cubic-bezier(0.4, 0, 0.2, 1);
    overflow: hidden;
  }
  .sidebar.collapsed { width: var(--sidebar-width-collapsed, 72px); }

  .sidebar-header {
    padding: 1.125rem 1rem;
    border-bottom: 1px solid var(--sidebar-border, rgba(255,255,255,0.08));
    flex-shrink: 0;
  }
  .logo {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.375rem 0.5rem;
    border-radius: 0.5rem;
    width: 100%;
    transition: background 0.15s;
  }
  .logo:hover { background: rgba(255,255,255,0.06); }
  .logo-icon {
    width: 34px;
    height: 34px;
    background: rgba(59,130,246,0.2);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .logo-text {
    font-size: 1.125rem;
    font-weight: 700;
    color: #fff;
    white-space: nowrap;
    letter-spacing: -0.01em;
  }

  .sidebar-nav {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 0.75rem 0;
    scrollbar-width: thin;
    scrollbar-color: rgba(255,255,255,0.1) transparent;
  }
  .sidebar-nav::-webkit-scrollbar { width: 4px; }
  .sidebar-nav::-webkit-scrollbar-track { background: transparent; }
  .sidebar-nav::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.1); border-radius: 2px; }

  .category-label {
    padding: 0.625rem 1.25rem 0.25rem;
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: rgba(148,163,184,0.5);
    white-space: nowrap;
  }

  .nav-list {
    list-style: none;
    margin: 0;
    padding: 0 0.625rem;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }
  .nav-item { margin: 0; }

  .nav-button {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.625rem 0.75rem;
    border: none;
    background: transparent;
    color: var(--sidebar-text, rgba(148,163,184,1));
    cursor: pointer;
    border-radius: 0.5rem;
    transition: all 0.15s ease;
    text-align: left;
    white-space: nowrap;
  }
  .nav-button:hover {
    background: rgba(255,255,255,0.07);
    color: var(--sidebar-text-hover, #fff);
  }
  .nav-button.active {
    background: var(--sidebar-active-bg, rgba(59,130,246,0.15));
    color: var(--sidebar-active-text, #60a5fa);
  }
  .nav-label {
    font-size: 0.8125rem;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .sidebar-footer {
    padding: 0.75rem;
    border-top: 1px solid var(--sidebar-border, rgba(255,255,255,0.08));
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  .collapse-button {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid rgba(255,255,255,0.12);
    background: transparent;
    color: rgba(148,163,184,0.7);
    cursor: pointer;
    border-radius: 0.5rem;
    transition: all 0.15s ease;
    font-size: 0.75rem;
    white-space: nowrap;
  }
  .collapse-button:hover {
    background: rgba(255,255,255,0.07);
    color: #fff;
    border-color: rgba(255,255,255,0.2);
  }
  .collapse-text { font-size: 0.75rem; }

  .sidebar.collapsed .nav-button {
    justify-content: center;
    padding: 0.625rem;
  }
  .sidebar.collapsed .collapse-button {
    justify-content: center;
    padding: 0.5rem;
  }
  .sidebar.collapsed .logo {
    justify-content: center;
    padding: 0.375rem;
  }
</style>
