<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Pagination from '$lib/shared/components/data/Pagination.svelte';
  import { getInventory } from '$lib/shared/api/inventory';
  import type { InventoryItem, InventoryFilter } from '$lib/shared/api/inventory';

  let items: InventoryItem[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter = '';
  let warehouseFilter = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'available', label: 'Available' },
    { value: 'reserved', label: 'Reserved' },
    { value: 'allocated', label: 'Allocated' },
    { value: 'in_transit', label: 'In Transit' },
    { value: 'quarantine', label: 'Quarantine' },
    { value: 'damaged', label: 'Damaged' }
  ];

  const warehouseOptions = [
    { value: '', label: 'All Warehouses' },
    { value: '1', label: 'Main Distribution Center' },
    { value: '2', label: 'West Coast Hub' },
    { value: '3', label: 'Retail Store - Downtown' }
  ];

  const columns = [
    { key: 'sku', label: 'SKU', sortable: true },
    { key: 'product', label: 'Product', sortable: true },
    { key: 'warehouse', label: 'Warehouse', sortable: true },
    { key: 'location', label: 'Location', sortable: true },
    { key: 'quantity', label: 'Qty', sortable: true, align: 'right' as const },
    { key: 'reserved', label: 'Reserved', sortable: true, align: 'right' as const },
    { key: 'available', label: 'Available', sortable: true, align: 'right' as const },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'expiration', label: 'Expiration', sortable: true }
  ];

  function getStatusVariant(status: string): 'green' | 'gray' | 'yellow' | 'red' | 'blue' | 'purple' | 'orange' {
    switch (status) {
      case 'available': return 'green';
      case 'reserved': return 'blue';
      case 'allocated': return 'purple';
      case 'in_transit': return 'yellow';
      case 'quarantine': return 'orange';
      case 'damaged': return 'red';
      default: return 'gray';
    }
  }

  function isExpiringSoon(dateStr: string | null): boolean {
    if (!dateStr) return false;
    const expirationDate = new Date(dateStr);
    const today = new Date();
    const thirtyDaysFromNow = new Date(today.setDate(today.getDate() + 30));
    return expirationDate <= thirtyDaysFromNow;
  }

  function isExpired(dateStr: string | null): boolean {
    if (!dateStr) return false;
    return new Date(dateStr) < new Date();
  }

  async function loadItems() {
    loading = true;
    error = null;
    try {
      const filter: InventoryFilter = {
        page: currentPage,
        pageSize,
      };
      if (searchQuery) filter.search = searchQuery;
      if (warehouseFilter) filter.warehouseId = warehouseFilter;
      const response = await getInventory(filter);
      items = response.data;
      totalItems = response.total;
      totalPages = response.totalPages;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load inventory items';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadItems();
  }

  function handleRowClick(item: InventoryItem) {
    window.location.href = `/inventory/items/${item.id}`;
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadItems();
  }

  function formatDate(dateStr: string | null): string {
    if (!dateStr) return '—';
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  onMount(() => {
    loadItems();
  });
</script>

<svelte:head>
  <title>Inventory Items | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Inventory Items</h1>
      <p class="page-description">Detailed view of all inventory items across warehouses</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={() => window.location.href = '/inventory/adjustments'}>
        New Adjustment
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  <Card>
    <div class="filters">
      <div class="filter-row">
        <div class="filter-item search-filter">
          <Input
            id="search"
            label="Search"
            type="search"
            placeholder="Search by SKU or product name..."
            bind:value={searchQuery}
            on:keydown={(e) => e.key === 'Enter' && handleSearch()}
          />
        </div>
        <div class="filter-item">
          <Select
            id="status"
            label="Status"
            options={statusOptions}
            bind:value={statusFilter}
            on:change={() => handleSearch()}
          />
        </div>
        <div class="filter-item">
          <Select
            id="warehouse"
            label="Warehouse"
            options={warehouseOptions}
            bind:value={warehouseFilter}
            on:change={() => handleSearch()}
          />
        </div>
        <div class="filter-item">
          <Button variant="secondary" on:click={handleSearch}>
            Search
          </Button>
        </div>
      </div>
    </div>

    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading inventory items...</p>
      </div>
    {:else if items.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
        <p class="text-gray-500">No inventory items found</p>
      </div>
    {:else}
      <Table columns={columns} data={items}>
        <tbody>
          {#each items as item}
            <tr on:click={() => handleRowClick(item)} class="clickable-row">
              <td class="font-medium">{item.productSku}</td>
              <td>{item.productName}</td>
              <td>{item.warehouseName}</td>
              <td>{item.locationCode}</td>
              <td class="text-right">{item.quantity}</td>
              <td class="text-right">{item.reservedQuantity}</td>
              <td class="text-right">
                <span class={item.availableQuantity < 10 ? 'text-red-600 font-medium' : ''}>
                  {item.availableQuantity}
                </span>
              </td>
              <td>
                <Badge variant={getStatusVariant(item.status)}>
                  {item.status.replace('_', ' ')}
                </Badge>
              </td>
              <td>
                {#if item.expirationDate}
                  <span 
                    class="text-sm"
                    class:text-red-600={isExpired(item.expirationDate)}
                    class:text-orange-600={!isExpired(item.expirationDate) && isExpiringSoon(item.expirationDate)}
                  >
                    {formatDate(item.expirationDate)}
                    {#if isExpired(item.expirationDate)}
                      <span class="text-xs ml-1">(Expired)</span>
                    {:else if isExpiringSoon(item.expirationDate)}
                      <span class="text-xs ml-1">(Soon)</span>
                    {/if}
                  </span>
                {:else}
                  <span class="text-gray-400">—</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </Table>

      <Pagination
        {currentPage}
        {totalPages}
        {totalItems}
        {pageSize}
        on:pageChange={(e) => handlePageChange(e.detail)}
      />
    {/if}
  </Card>
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1400px;
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

  .filters {
    margin-bottom: 1rem;
  }

  .filter-row {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .filter-item {
    flex: 1;
    min-width: 200px;
  }

  .search-filter {
    flex: 2;
    min-width: 300px;
  }

  .loading-container,
  .empty-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .clickable-row {
    cursor: pointer;
  }

  .clickable-row:hover {
    background-color: var(--color-gray-50);
  }

  @media (max-width: 768px) {
    .filter-row {
      flex-direction: column;
    }

    .filter-item,
    .search-filter {
      min-width: 100%;
    }
  }
</style>
