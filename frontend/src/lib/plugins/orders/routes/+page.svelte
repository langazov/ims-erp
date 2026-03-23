<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Pagination from '$lib/shared/components/data/Pagination.svelte';
  import ConfirmModal from '$lib/shared/components/layout/ConfirmModal.svelte';
  import EmptyState from '$lib/shared/components/display/EmptyState.svelte';
  import {
    getOrders,
    updateOrderStatus,
    type Order,
    type OrderStatus,
    type OrderFilter
  } from '$lib/shared/api/orders';
  import { toast } from '$lib/shared/stores/toast';

  let orders: Order[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter: OrderStatus | '' = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;

  let cancelTarget: Order | null = null;
  let cancelModalOpen = false;
  let cancelling = false;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'pending', label: 'Pending' },
    { value: 'confirmed', label: 'Confirmed' },
    { value: 'processing', label: 'Processing' },
    { value: 'shipped', label: 'Shipped' },
    { value: 'delivered', label: 'Delivered' },
    { value: 'cancelled', label: 'Cancelled' }
  ];

  const columns = [
    { key: 'orderNumber', label: 'Order #', sortable: true },
    { key: 'clientName', label: 'Client', sortable: true },
    { key: 'createdAt', label: 'Date', sortable: true },
    { key: 'items', label: 'Items', sortable: false },
    { key: 'total', label: 'Total', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  function formatCurrency(s: string | number): string {
    const num = typeof s === 'string' ? parseFloat(s) : s;
    if (isNaN(num)) return String(s);
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(num);
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function getStatusVariant(status: OrderStatus): 'green' | 'yellow' | 'blue' | 'purple' | 'gray' {
    switch (status) {
      case 'pending': return 'yellow';
      case 'confirmed': return 'blue';
      case 'processing': return 'blue';
      case 'shipped': return 'purple';
      case 'delivered': return 'green';
      case 'cancelled': return 'gray';
      default: return 'gray';
    }
  }

  async function loadOrders() {
    loading = true;
    error = null;
    try {
      const filter: OrderFilter = {
        page: currentPage,
        pageSize
      };
      if (statusFilter) filter.status = statusFilter;
      if (searchQuery.trim()) filter.search = searchQuery.trim();

      const response = await getOrders(filter);
      orders = response.data;
      totalPages = response.totalPages;
      totalItems = response.total;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load orders';
    } finally {
      loading = false;
    }
  }

  function handleFilterChange() {
    currentPage = 1;
    loadOrders();
  }

  function openCancelModal(order: Order) {
    cancelTarget = order;
    cancelModalOpen = true;
  }

  async function handleCancelConfirm() {
    if (!cancelTarget) return;
    cancelling = true;
    try {
      await updateOrderStatus(cancelTarget.id, 'cancelled');
      toast.success(`Order ${cancelTarget.orderNumber} cancelled`);
      cancelModalOpen = false;
      cancelTarget = null;
      await loadOrders();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to cancel order');
    } finally {
      cancelling = false;
    }
  }

  onMount(loadOrders);
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Orders</h1>
      <p class="page-subtitle">Manage customer orders and fulfillment</p>
    </div>
    <Button variant="primary" on:click={() => goto('/orders/new')}>New Order</Button>
  </div>

  <Card>
    <div class="filters-row">
      <div class="search-wrap">
        <Input
          id="search"
          label=""
          type="text"
          placeholder="Search orders..."
          bind:value={searchQuery}
        />
      </div>
      <div class="filter-wrap">
        <Select
          id="status-filter"
          label=""
          options={statusOptions}
          bind:value={statusFilter}
          on:change={handleFilterChange}
        />
      </div>
      <Button variant="secondary" on:click={handleFilterChange}>Search</Button>
    </div>

    {#if error}
      <Alert variant="error" dismissible on:dismiss={() => (error = null)}>{error}</Alert>
    {/if}

    {#if loading}
      <div class="spinner-wrap">
        <Spinner size="lg" />
      </div>
    {:else if orders.length === 0}
      <EmptyState
        title="No orders found"
        description="No orders match your current filters. Create a new order to get started."
        actionLabel="New Order"
        icon="📦"
        on:action={() => goto('/orders/new')}
      />
    {:else}
      <Table {columns}>
        <tbody>
          {#each orders as order (order.id)}
            <tr
              class="table-row"
              on:click={() => goto(`/orders/${order.id}`)}
              role="button"
              tabindex="0"
              on:keydown={(e) => e.key === 'Enter' && goto(`/orders/${order.id}`)}
            >
              <td class="td td-mono">{order.orderNumber}</td>
              <td class="td">{order.clientName}</td>
              <td class="td">{formatDate(order.createdAt)}</td>
              <td class="td">{order.items.length} item{order.items.length !== 1 ? 's' : ''}</td>
              <td class="td font-medium">{formatCurrency(order.total)}</td>
              <td class="td">
                <Badge variant={getStatusVariant(order.status)} dot>
                  {order.status.charAt(0).toUpperCase() + order.status.slice(1)}
                </Badge>
              </td>
              <td class="td actions-cell" on:click|stopPropagation>
                <Button
                  variant="ghost"
                  size="sm"
                  on:click={() => goto(`/orders/${order.id}`)}
                >
                  View
                </Button>
                {#if order.status !== 'cancelled' && order.status !== 'delivered'}
                  <Button
                    variant="ghost"
                    size="sm"
                    on:click={() => openCancelModal(order)}
                  >
                    Cancel
                  </Button>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </Table>

      {#if totalPages > 1}
        <div class="pagination-wrap">
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            totalItems={totalItems}
            pageSize={pageSize}
            on:pageChange={(e) => {
              currentPage = e.detail;
              loadOrders();
            }}
          />
        </div>
      {/if}
    {/if}
  </Card>
</div>

<ConfirmModal
  bind:open={cancelModalOpen}
  title="Cancel Order"
  message="Are you sure you want to cancel order {cancelTarget?.orderNumber}? This action cannot be undone."
  confirmLabel="Cancel Order"
  cancelLabel="Keep Order"
  variant="danger"
  loading={cancelling}
  on:confirm={handleCancelConfirm}
  on:cancel={() => (cancelModalOpen = false)}
/>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1400px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .page-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-gray-900, #111827);
    margin: 0;
  }

  .page-subtitle {
    font-size: 0.875rem;
    color: var(--color-gray-500, #6b7280);
    margin: 0.25rem 0 0;
  }

  .filters-row {
    display: flex;
    gap: 0.75rem;
    align-items: flex-end;
    margin-bottom: 1rem;
    flex-wrap: wrap;
  }

  .search-wrap {
    flex: 1;
    min-width: 200px;
  }

  .filter-wrap {
    width: 160px;
  }

  .spinner-wrap {
    display: flex;
    justify-content: center;
    padding: 3rem 0;
  }

  .table-row {
    cursor: pointer;
    transition: background-color 0.15s;
  }

  .table-row:hover {
    background-color: var(--color-gray-50, #f9fafb);
  }

  .td {
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    color: var(--color-gray-700, #374151);
    border-bottom: 1px solid var(--color-gray-100, #f3f4f6);
    vertical-align: middle;
  }

  .td-mono {
    font-family: monospace;
    font-size: 0.8125rem;
    color: var(--color-gray-600, #4b5563);
  }

  .font-medium {
    font-weight: 600;
  }

  .actions-cell {
    display: flex;
    gap: 0.25rem;
    align-items: center;
  }

  .pagination-wrap {
    margin-top: 1rem;
    display: flex;
    justify-content: center;
  }
</style>
