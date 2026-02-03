<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import { menuStore } from '../stores';

  interface ModuleInfo {
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

  let searchQuery = '';
  let categories: Record<string, ModuleInfo[]> = {
    'Core': [],
    'Management': [],
    'Operations': [],
    'Settings': [],
    'Other': []
  };

  const categoryOrder = ['Core', 'Management', 'Operations', 'Settings', 'Other'];

  function getCategoryIcon(category: string): string {
    switch (category) {
      case 'Core': return 'home';
      case 'Management': return 'users';
      case 'Operations': return 'settings';
      case 'Settings': return 'cog';
      default: return 'grid';
    }
  }

  function getStatusVariant(status: string): 'green' | 'gray' | 'red' {
    switch (status) {
      case 'enabled': return 'green';
      case 'disabled': return 'gray';
      case 'error': return 'red';
      default: return 'gray';
    }
  }

  function filterModules(modules: ModuleInfo[], query: string): ModuleInfo[] {
    if (!query.trim()) return modules;
    const q = query.toLowerCase();
    return modules.filter(m => 
      m.name.toLowerCase().includes(q) || 
      m.description.toLowerCase().includes(q)
    );
  }

  function navigateToModule(module: ModuleInfo) {
    if (module.route) {
      window.location.href = module.route;
    }
  }

  function toggleCategory(category: string) {
    menuStore.toggleCategory(category);
  }

  function handleSearch() {
    menuStore.setSearchQuery(searchQuery);
  }

  function handleSearchInput(event: KeyboardEvent) {
    searchQuery = (event.target as HTMLInputElement).value;
    menuStore.setSearchQuery(searchQuery);
  }

  onMount(() => {
    const demoModules: ModuleInfo[] = [
      {
        id: 'dashboard',
        name: 'Dashboard',
        version: '1.0.0',
        description: 'Main dashboard with customizable widgets and layouts',
        author: 'IMS Team',
        status: 'enabled',
        priority: 100,
        category: 'Core',
        route: '/dashboard'
      },
      {
        id: 'menu',
        name: 'Modules',
        version: '1.0.0',
        description: 'Module menu and navigation for accessing installed plugins',
        author: 'IMS Team',
        status: 'enabled',
        priority: 1,
        category: 'Core',
        route: '/modules'
      },
      {
        id: 'clients',
        name: 'Clients',
        version: '1.0.0',
        description: 'Client management for maintaining customer relationships',
        author: 'IMS Team',
        status: 'enabled',
        priority: 10,
        category: 'Management',
        route: '/clients'
      },
      {
        id: 'users',
        name: 'Users',
        version: '1.0.0',
        description: 'User management and access control',
        author: 'IMS Team',
        status: 'enabled',
        priority: 10,
        category: 'Management',
        route: '/users'
      },
      {
        id: 'products',
        name: 'Products',
        version: '1.0.0',
        description: 'Product catalog and inventory management',
        author: 'IMS Team',
        status: 'enabled',
        priority: 10,
        category: 'Management',
        route: '/products'
      },
      {
        id: 'inventory',
        name: 'Inventory',
        version: '1.0.0',
        description: 'Stock tracking and inventory management',
        author: 'IMS Team',
        status: 'enabled',
        priority: 10,
        category: 'Operations',
        route: '/inventory'
      },
      {
        id: 'warehouse',
        name: 'Warehouse',
        version: '1.0.0',
        description: 'Warehouse management and operations',
        author: 'IMS Team',
        status: 'enabled',
        priority: 10,
        category: 'Operations',
        route: '/warehouse'
      },
      {
        id: 'orders',
        name: 'Orders',
        version: '1.0.0',
        description: 'Order management and fulfillment',
        author: 'IMS Team',
        status: 'enabled',
        priority: 10,
        category: 'Operations',
        route: '/orders'
      },
      {
        id: 'invoices',
        name: 'Invoices',
        version: '1.0.0',
        description: 'Invoice generation and management',
        author: 'IMS Team',
        status: 'enabled',
        priority: 10,
        category: 'Operations',
        route: '/invoices'
      },
      {
        id: 'payments',
        name: 'Payments',
        version: '1.0.0',
        description: 'Payment processing and reconciliation',
        author: 'IMS Team',
        status: 'enabled',
        priority: 10,
        category: 'Operations',
        route: '/payments'
      },
      {
        id: 'settings',
        name: 'Settings',
        version: '1.0.0',
        description: 'System settings and configuration',
        author: 'IMS Team',
        status: 'enabled',
        priority: 10,
        category: 'Settings',
        route: '/settings'
      },
      {
        id: 'documents',
        name: 'Documents',
        version: '1.0.0',
        description: 'Document management and storage',
        author: 'IMS Team',
        status: 'enabled',
        priority: 10,
        category: 'Settings',
        route: '/documents'
      }
    ];

    for (const module of demoModules) {
      categories[module.category] = [...categories[module.category], module];
    }

    for (const cat of categoryOrder) {
      categories[cat] = filterModules(categories[cat], searchQuery);
    }
  });
</script>

<svelte:head>
  <title>Modules | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Modules</h1>
      <p class="page-description">Access and manage installed modules</p>
    </div>
    <div class="header-actions">
      <Button variant="secondary" on:click={() => menuStore.expandAll()}>
        Expand All
      </Button>
      <Button variant="secondary" on:click={() => menuStore.collapseAll()}>
        Collapse All
      </Button>
    </div>
  </div>

  <Card>
    <div class="search-section">
      <div class="form-group">
        <label for="search" class="form-label">Search Modules</label>
        <input
          id="search"
          type="search"
          class="form-input"
          placeholder="Search by name or description..."
          bind:value={searchQuery}
          on:keydown={(e) => e.key === 'Enter' && handleSearch()}
        />
      </div>
    </div>

    <div class="modules-list">
      {#each categoryOrder as category}
        {#if categories[category] && categories[category].length > 0}
          <div class="category-section">
            <button 
              class="category-header"
              on:click={() => toggleCategory(category)}
            >
              <div class="category-title">
                <span class="category-icon">
                  {#if $menuStore.expandedCategories.includes(category)}
                    ▼
                  {:else}
                    ▶
                  {/if}
                </span>
                <span class="category-name">{category}</span>
                <Badge variant="gray">{categories[category].length}</Badge>
              </div>
            </button>

            {#if $menuStore.expandedCategories.includes(category)}
              <div class="category-modules">
                {#each categories[category] as module}
                  <div 
                    class="module-card"
                    on:click={() => navigateToModule(module)}
                    on:keypress={(e) => e.key === 'Enter' && navigateToModule(module)}
                    role="button"
                    tabindex="0"
                  >
                    <div class="module-header">
                      <div class="module-info">
                        <h3 class="module-name">{module.name}</h3>
                        <span class="module-version">v{module.version}</span>
                      </div>
                      <Badge variant={getStatusVariant(module.status)}>
                        {module.status}
                      </Badge>
                    </div>
                    <p class="module-description">{module.description}</p>
                    {#if module.author}
                      <span class="module-author">by {module.author}</span>
                    {/if}
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}
      {/each}

      {#if Object.values(categories).every(cat => cat.length === 0)}
        <div class="empty-state">
          <p>No modules found</p>
        </div>
      {/if}
    </div>
  </Card>

  <div class="stats-section">
    <Card>
      <h2 class="stats-title">System Statistics</h2>
      <div class="stats-grid">
        <div class="stat-item">
          <span class="stat-value">{Object.values(categories).flat().length}</span>
          <span class="stat-label">Total Modules</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{Object.values(categories).flat().filter(m => m.status === 'enabled').length}</span>
          <span class="stat-label">Enabled</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{categoryOrder.filter(c => categories[c] && categories[c].length > 0).length}</span>
          <span class="stat-label">Categories</span>
        </div>
      </div>
    </Card>
  </div>
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
  }

  .page-title {
    font-size: 1.875rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin: 0;
  }

  .page-description {
    color: var(--color-gray-500);
    margin-top: 0.25rem;
  }

  .header-actions {
    display: flex;
    gap: 0.75rem;
  }

  .search-section {
    margin-bottom: 1.5rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .form-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-700);
  }

  .form-input {
    width: 100%;
    padding: 0.625rem 0.875rem;
    font-size: 0.875rem;
    border: 1px solid var(--color-gray-300);
    border-radius: 0.5rem;
    background: white;
    color: var(--color-gray-900);
  }

  .form-input:focus {
    outline: none;
    border-color: var(--color-primary-500);
    box-shadow: 0 0 0 3px var(--color-primary-100);
  }

  .modules-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .category-section {
    border: 1px solid var(--color-gray-200);
    border-radius: 0.5rem;
    overflow: hidden;
  }

  .category-header {
    width: 100%;
    padding: 0.75rem 1rem;
    background: var(--color-gray-50);
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: space-between;
    text-align: left;
  }

  .category-header:hover {
    background: var(--color-gray-100);
  }

  .category-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .category-icon {
    width: 1rem;
    color: var(--color-gray-500);
  }

  .category-name {
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .category-modules {
    padding: 1rem;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem;
    background: white;
  }

  .module-card {
    border: 1px solid var(--color-gray-200);
    border-radius: 0.5rem;
    padding: 1rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .module-card:hover {
    border-color: var(--color-primary-300);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  .module-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 0.5rem;
  }

  .module-info {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
  }

  .module-name {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0;
  }

  .module-version {
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  .module-description {
    font-size: 0.875rem;
    color: var(--color-gray-600);
    margin: 0 0 0.5rem 0;
    line-height: 1.5;
  }

  .module-author {
    font-size: 0.75rem;
    color: var(--color-gray-400);
  }

  .empty-state {
    text-align: center;
    padding: 2rem;
    color: var(--color-gray-500);
  }

  .stats-section {
    margin-top: 1.5rem;
  }

  .stats-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
  }

  .stat-item {
    text-align: center;
    padding: 1rem;
  }

  .stat-value {
    display: block;
    font-size: 2rem;
    font-weight: 700;
    color: var(--color-gray-900);
  }

  .stat-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  @media (max-width: 640px) {
    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .header-actions {
      width: 100%;
      justify-content: flex-start;
    }

    .stats-grid {
      grid-template-columns: 1fr;
    }

    .category-modules {
      grid-template-columns: 1fr;
    }
  }
</style>
