<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';

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

  let moduleId = '';
  let module: ModuleInfo | null = null;
  let loading = true;

  const allModules: ModuleInfo[] = [
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

  function getStatusVariant(status: string): 'green' | 'gray' | 'red' {
    switch (status) {
      case 'enabled': return 'green';
      case 'disabled': return 'gray';
      case 'error': return 'red';
      default: return 'gray';
    }
  }

  function navigateToModule() {
    if (module?.route) {
      goto(module.route);
    }
  }

  onMount(() => {
    moduleId = $page.params.id;
    module = allModules.find(m => m.id === moduleId) || null;
    loading = false;
  });
</script>

<svelte:head>
  <title>{module ? module.name : 'Module'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <p>Loading module...</p>
    </div>
  {:else if !module}
    <div class="error-container">
      <p>Module not found</p>
      <Button variant="secondary" on:click={() => goto('/modules')}>
        Back to Modules
      </Button>
    </div>
  {:else}
    <div class="page-header">
      <div class="header-content">
        <div class="header-nav">
          <Button variant="ghost" size="sm" on:click={() => goto('/modules')}>
            ← Back to Modules
          </Button>
        </div>
        <h1 class="page-title">{module.name}</h1>
        <div class="header-meta">
          <Badge variant={getStatusVariant(module.status)}>
            {module.status}
          </Badge>
          <span class="meta-item">v{module.version}</span>
          <span class="meta-item">Category: {module.category}</span>
        </div>
      </div>
      <div class="header-actions">
        <Button variant="primary" on:click={navigateToModule}>
          Open Module
        </Button>
      </div>
    </div>

    <div class="content-grid">
      <Card>
        <h2 class="card-title">Module Information</h2>
        <dl class="info-list">
          <div class="info-item">
            <dt>Module ID</dt>
            <dd>{module.id}</dd>
          </div>
          <div class="info-item">
            <dt>Name</dt>
            <dd>{module.name}</dd>
          </div>
          <div class="info-item">
            <dt>Version</dt>
            <dd>{module.version}</dd>
          </div>
          <div class="info-item">
            <dt>Category</dt>
            <dd>{module.category}</dd>
          </div>
          <div class="info-item">
            <dt>Status</dt>
            <dd>
              <Badge variant={getStatusVariant(module.status)}>
                {module.status}
              </Badge>
            </dd>
          </div>
          <div class="info-item">
            <dt>Priority</dt>
            <dd>{module.priority}</dd>
          </div>
          {#if module.author}
            <div class="info-item">
              <dt>Author</dt>
              <dd>{module.author}</dd>
            </div>
          {/if}
        </dl>
      </Card>

      <Card>
        <h2 class="card-title">Description</h2>
        <p class="description">{module.description}</p>
      </Card>

      <Card>
        <h2 class="card-title">Route</h2>
        <div class="route-info">
          <code class="route-code">{module.route || '/' + module.id}</code>
          <Button variant="secondary" size="sm" on:click={navigateToModule}>
            Navigate
          </Button>
        </div>
      </Card>

      <Card>
        <h2 class="card-title">Actions</h2>
        <div class="actions-list">
          <Button variant="secondary" fullWidth on:click={navigateToModule}>
            Open Module
          </Button>
          <Button variant="secondary" fullWidth>
            Configure
          </Button>
          <Button variant="secondary" fullWidth>
            View Documentation
          </Button>
          <Button variant="danger" fullWidth>
            Disable Module
          </Button>
        </div>
      </Card>
    </div>
  {/if}
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1000px;
    margin: 0 auto;
  }

  .loading-container,
  .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .header-nav {
    margin-bottom: 0.5rem;
  }

  .page-title {
    font-size: 1.875rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin: 0;
  }

  .header-meta {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-top: 0.5rem;
    flex-wrap: wrap;
  }

  .meta-item {
    color: var(--color-gray-500);
    font-size: 0.875rem;
  }

  .header-actions {
    display: flex;
    gap: 0.75rem;
  }

  .content-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
  }

  .card-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
  }

  .info-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .info-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--color-gray-100);
  }

  .info-item:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }

  .info-item dt {
    color: var(--color-gray-500);
    font-size: 0.875rem;
  }

  .info-item dd {
    color: var(--color-gray-900);
    font-weight: 500;
    margin: 0;
  }

  .description {
    color: var(--color-gray-700);
    line-height: 1.6;
    margin: 0;
  }

  .route-info {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .route-code {
    background: var(--color-gray-100);
    padding: 0.5rem 0.75rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    color: var(--color-gray-800);
  }

  .actions-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  @media (max-width: 768px) {
    .content-grid {
      grid-template-columns: 1fr;
    }

    .page-header {
      flex-direction: column;
    }

    .header-actions {
      width: 100%;
      justify-content: flex-start;
    }
  }
</style>
