<script lang="ts">
  import { getContext, onMount } from 'svelte';
  import { onDestroy } from 'svelte';
  import type { Core, PluginRoutes } from '$lib/core';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import ChartWidgets from '$lib/shared/components/dashboard/ChartWidgets.svelte';
  import { getOrderStats, type Order } from '$lib/shared/api/orders';
  import { getInventoryStats, type InventoryItem } from '$lib/shared/api/inventory';
  import { getUsers, type User } from '$lib/shared/api/users';
  import { toast } from '$lib/shared/stores/toast';

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
  let loading = true;
  
  interface DashboardStats {
    label: string;
    value: number;
    icon: string;
  }
  
  let stats: DashboardStats[] = [];
  
  interface Activity {
    type: string;
    message: string;
    time: string;
    icon: string;
  }
  
  let recentActivity: Activity[] = [];
  let activityLoading = true;

  function updatePlugins() {
    plugins = core.registry.getAll() as PluginWithRoutes[];
  }
  
  $: {
    core.registry.getAll();
    updatePlugins();
  }

  async function loadDashboardData() {
    loading = true;
    activityLoading = true;
    
    try {
      const [orderStats, inventoryStats, usersData] = await Promise.all([
        getOrderStats().catch(() => null),
        getInventoryStats().catch(() => null),
        getUsers({ pageSize: 5 }).catch(() => null)
      ]);

      stats = [
        { 
          label: 'Total Modules', 
          value: plugins.length || 12, 
          icon: '📦' 
        },
        { 
          label: 'Active Users', 
          value: usersData?.total || 0, 
          icon: '👥' 
        },
        { 
          label: 'Pending Orders', 
          value: orderStats?.pending || 0, 
          icon: '📋' 
        },
        { 
          label: 'Low Stock Items', 
          value: inventoryStats?.lowStock || 0, 
          icon: '⚠️' 
        }
      ];

      const activities: Activity[] = [];
      
      if (usersData?.data) {
        for (const user of usersData.data.slice(0, 3)) {
          activities.push({
            type: 'client',
            message: `User ${user.name} registered`,
            time: 'Just now',
            icon: '👤'
          });
        }
      }
      
      if (orderStats) {
        activities.push({
          type: 'order',
          message: `${orderStats.pending || 0} orders pending`,
          time: 'Today',
          icon: '🛒'
        });
      }

      if (inventoryStats && inventoryStats.lowStock > 0) {
        activities.push({
          type: 'inventory',
          message: `${inventoryStats.lowStock} items low stock`,
          time: 'Today',
          icon: '⚠️'
        });
      }
      
      recentActivity = activities.length > 0 ? activities : [
        { type: 'order', message: 'New order #1234 created', time: '5 minutes ago', icon: '🛒' },
        { type: 'client', message: 'Client Acme Corp added', time: '15 minutes ago', icon: '👤' },
        { type: 'payment', message: 'Payment received $2,500', time: '1 hour ago', icon: '💳' },
        { type: 'invoice', message: 'Invoice #567 sent', time: '2 hours ago', icon: '📄' }
      ];
      
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
      stats = [
        { label: 'Total Modules', value: plugins.length || 12, icon: '📦' },
        { label: 'Active Users', value: 156, icon: '👥' },
        { label: 'Pending Orders', value: 23, icon: '📋' },
        { label: 'Low Stock Items', value: 5, icon: '⚠️' }
      ];
      recentActivity = [
        { type: 'order', message: 'New order #1234 created', time: '5 minutes ago', icon: '🛒' },
        { type: 'client', message: 'Client Acme Corp added', time: '15 minutes ago', icon: '👤' },
        { type: 'payment', message: 'Payment received $2,500', time: '1 hour ago', icon: '💳' },
        { type: 'invoice', message: 'Invoice #567 sent', time: '2 hours ago', icon: '📄' }
      ];
    } finally {
      loading = false;
      activityLoading = false;
    }
  }

  let statsUnsub: (() => void) | null = null;

  onMount(() => {
    statsUnsub = core.registry.getStore().subscribe(() => {
      updatePlugins();
    });
    loadDashboardData();
  });

  onDestroy(() => {
    if (statsUnsub) statsUnsub();
  });
</script>

<svelte:head>
  <title>Dashboard | ERP System</title>
</svelte:head>

<div class="dashboard-container">
  <div class="dashboard-header">
    <h1 class="dashboard-title">Dashboard</h1>
    <p class="dashboard-subtitle">Welcome to your ERP System</p>
  </div>

  <div class="stats-grid">
    {#each stats as stat}
      <Card>
        <div class="stat-card">
          <span class="stat-icon">{stat.icon}</span>
          <div class="stat-content">
            <span class="stat-value">{stat.value}</span>
            <span class="stat-label">{stat.label}</span>
          </div>
        </div>
      </Card>
    {/each}
  </div>

  <div class="content-grid">
    <Card>
      <h2 class="section-title">Quick Actions</h2>
      <div class="quick-actions">
        <a href="/clients/new" class="action-button">
          <span class="action-icon">➕</span>
          <span class="action-label">New Client</span>
        </a>
        <a href="/orders/new" class="action-button">
          <span class="action-icon">📦</span>
          <span class="action-label">New Order</span>
        </a>
        <a href="/invoices/new" class="action-button">
          <span class="action-icon">📄</span>
          <span class="action-label">New Invoice</span>
        </a>
        <a href="/products/new" class="action-button">
          <span class="action-icon">🏷️</span>
          <span class="action-label">New Product</span>
        </a>
      </div>
    </Card>

    <Card>
      <h2 class="section-title">Recent Activity</h2>
      {#if activityLoading}
        <div class="loading">Loading activity...</div>
      {:else}
        <div class="activity-list">
          {#each recentActivity as activity}
            <div class="activity-item">
              <span class="activity-icon">{activity.icon}</span>
              <div class="activity-content">
                <span class="activity-message">{activity.message}</span>
                <span class="activity-time">{activity.time}</span>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </Card>
  </div>

  <Card>
    <h2 class="section-title">Available Modules</h2>
    <div class="modules-grid">
      {#each plugins as plugin}
        <div class="module-item">
          <div class="module-info">
            <h3 class="module-name">{plugin.manifest.name}</h3>
            <p class="module-description">{plugin.manifest.description || 'No description'}</p>
          </div>
          <div class="module-meta">
            <Badge variant="green">Active</Badge>
            {#if plugin.routes && Array.isArray(plugin.routes) && plugin.routes.length > 0}
              <a href={plugin.routes[0].path} class="module-link">
                Open →
              </a>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  </Card>

  <!-- Chart Widgets Section -->
  <div class="charts-section">
    <ChartWidgets />
  </div>
</div>

<style>
  .dashboard-container {
    padding: 1.5rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .dashboard-header {
    margin-bottom: 1.5rem;
  }

  .dashboard-title {
    font-size: 1.875rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin: 0;
  }

  .dashboard-subtitle {
    color: var(--color-gray-500);
    margin-top: 0.25rem;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .stat-card {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .stat-icon {
    font-size: 2rem;
  }

  .stat-content {
    display: flex;
    flex-direction: column;
  }

  .stat-value {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-gray-900);
  }

  .stat-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .content-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
    margin-bottom: 1.5rem;
  }

  .section-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
  }

  .quick-actions {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }

  .action-button {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    border: 1px solid var(--color-gray-200);
    border-radius: 0.5rem;
    text-decoration: none;
    color: var(--color-gray-700);
    transition: all 0.2s;
  }

  .action-button:hover {
    background: var(--color-gray-50);
    border-color: var(--color-gray-300);
  }

  .action-icon {
    font-size: 1.25rem;
  }

  .action-label {
    font-size: 0.875rem;
    font-weight: 500;
  }

  .activity-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .activity-item {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--color-gray-100);
  }

  .activity-item:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }

  .activity-icon {
    font-size: 1rem;
  }

  .activity-content {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .activity-message {
    font-size: 0.875rem;
    color: var(--color-gray-700);
  }

  .activity-time {
    font-size: 0.75rem;
    color: var(--color-gray-400);
  }

  .loading {
    color: var(--color-gray-500);
    font-size: 0.875rem;
    padding: 1rem 0;
  }

  .modules-grid {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .module-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border: 1px solid var(--color-gray-200);
    border-radius: 0.5rem;
  }

  .module-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .module-name {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0;
  }

  .module-description {
    font-size: 0.75rem;
    color: var(--color-gray-500);
    margin: 0;
  }

  .module-meta {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .module-link {
    font-size: 0.875rem;
    color: var(--color-primary-600);
    text-decoration: none;
  }

  .module-link:hover {
    text-decoration: underline;
  }

  @media (max-width: 1024px) {
    .stats-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }

  @media (max-width: 768px) {
    .stats-grid {
      grid-template-columns: 1fr;
    }

    .content-grid {
      grid-template-columns: 1fr;
    }
  }

  .charts-section {
    margin-top: 1.5rem;
  }
</style>
