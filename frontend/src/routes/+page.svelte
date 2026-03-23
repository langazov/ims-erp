<script lang="ts">
  import { onMount } from 'svelte';
  import StatCard from '$lib/shared/components/display/StatCard.svelte';
  import Skeleton from '$lib/shared/components/display/Skeleton.svelte';
  import ChartWidgets from '$lib/shared/components/dashboard/ChartWidgets.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import { getDashboardKPIs } from '$lib/shared/api/analytics';
  import { toast } from '$lib/shared/stores/toast';

  interface Activity {
    type: string;
    message: string;
    time: string;
    icon: string;
  }

  let kpisLoading = $state(true);

  let revenueToday = $state('$0');
  let activeOrders = $state(0);
  let overdueInvoices = $state(0);
  let lowStockItems = $state(0);
  let revenueGrowthPct = $state<number | null>(null);

  const recentActivity: Activity[] = [
    { type: 'order',   message: 'New order #1234 created',   time: '5 minutes ago',  icon: '🛒' },
    { type: 'client',  message: 'Client Acme Corp added',    time: '15 minutes ago', icon: '👤' },
    { type: 'payment', message: 'Payment received $2,500',   time: '1 hour ago',     icon: '💳' },
    { type: 'invoice', message: 'Invoice #567 sent',         time: '2 hours ago',    icon: '📄' },
    { type: 'stock',   message: '3 items restocked',         time: '3 hours ago',    icon: '📦' },
  ];

  const quickActions = [
    { label: 'New Client',  href: '/clients/new',  icon: '👤' },
    { label: 'New Order',   href: '/orders/new',   icon: '🛒' },
    { label: 'New Invoice', href: '/invoices/new', icon: '📄' },
    { label: 'New Product', href: '/products/new', icon: '🏷️' },
  ];

  onMount(async () => {
    try {
      const kpis = await getDashboardKPIs();
      revenueToday     = kpis.revenueToday;
      activeOrders     = kpis.activeOrders;
      overdueInvoices  = kpis.overdueInvoices;
      lowStockItems    = kpis.lowStockItems;
      revenueGrowthPct = kpis.revenueGrowthPct ?? null;
    } catch {
      toast.error('Failed to load dashboard data');
      revenueToday     = '$0';
      activeOrders     = 0;
      overdueInvoices  = 0;
      lowStockItems    = 0;
    } finally {
      kpisLoading = false;
    }
  });
</script>

<svelte:head>
  <title>Dashboard | ERP System</title>
</svelte:head>

<div class="p-6 max-w-screen-xl mx-auto space-y-6">
  <!-- Header -->
  <div>
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
    <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Welcome back — here's what's happening today.</p>
  </div>

  <!-- KPI Stat Cards -->
  <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
    {#if kpisLoading}
      {#each Array(4) as _}
        <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
          <Skeleton class="h-4 w-1/2 mb-4" />
          <Skeleton class="h-8 w-1/3 mb-2" />
          <Skeleton class="h-3 w-2/3" />
        </div>
      {/each}
    {:else}
      <StatCard
        title="Revenue Today"
        value={revenueToday}
        change={revenueGrowthPct}
        changeLabel="vs yesterday"
        icon="💰"
        iconBg="bg-emerald-100"
        iconColor="text-emerald-600"
        href="/invoices"
      />
      <StatCard
        title="Active Orders"
        value={activeOrders}
        icon="🛒"
        iconBg="bg-blue-100"
        iconColor="text-blue-600"
        href="/orders"
      />
      <StatCard
        title="Overdue Invoices"
        value={overdueInvoices}
        icon="⚠️"
        iconBg="bg-amber-100"
        iconColor="text-amber-600"
        href="/invoices?filter=overdue"
      />
      <StatCard
        title="Low Stock Items"
        value={lowStockItems}
        icon="📦"
        iconBg="bg-red-100"
        iconColor="text-red-600"
        href="/inventory?filter=low-stock"
      />
    {/if}
  </div>

  <!-- Quick Actions + Recent Activity -->
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- Quick Actions -->
    <Card>
      <h2 class="text-sm font-semibold text-gray-900 dark:text-white mb-4">Quick Actions</h2>
      <div class="grid grid-cols-2 gap-3">
        {#each quickActions as action}
          <a
            href={action.href}
            class="flex items-center gap-3 p-3 rounded-lg border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 hover:border-gray-300 dark:hover:border-gray-600 transition-colors"
          >
            <span class="text-xl">{action.icon}</span>
            <span class="text-sm font-medium">{action.label}</span>
          </a>
        {/each}
      </div>
    </Card>

    <!-- Recent Activity -->
    <Card>
      <h2 class="text-sm font-semibold text-gray-900 dark:text-white mb-4">Recent Activity</h2>
      <ul class="space-y-3">
        {#each recentActivity as activity, i}
          <li class="flex items-start gap-3 {i < recentActivity.length - 1 ? 'pb-3 border-b border-gray-100 dark:border-gray-700' : ''}">
            <span class="mt-0.5 text-base">{activity.icon}</span>
            <div class="min-w-0 flex-1">
              <p class="text-sm text-gray-700 dark:text-gray-300 truncate">{activity.message}</p>
              <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">{activity.time}</p>
            </div>
          </li>
        {/each}
      </ul>
    </Card>
  </div>

  <!-- Charts -->
  <ChartWidgets />
</div>
