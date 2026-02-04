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
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import Pagination from '$lib/shared/components/data/Pagination.svelte';

  interface InventoryItem {
    id: string;
    productId: string;
    productName: string;
    productSku: string;
    warehouseId: string;
    warehouseName: string;
    quantity: number;
    reservedQuantity: number;
    availableQuantity: number;
    status: 'available' | 'reserved' | 'allocated' | 'in_transit' | 'quarantine' | 'damaged';
    lotNumber: string | null;
    expirationDate: string | null;
    lastUpdated: string;
  }

  let items: InventoryItem[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter: string = '';
  let warehouseFilter: string = '';
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
    { key: 'quantity', label: 'Qty', sortable: true },
    { key: 'reserved', label: 'Reserved', sortable: true },
    { key: 'available', label: 'Available', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
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

  async function loadInventory() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      items = [
        {
          id: '1',
          productId: '1',
          productName: 'Wireless Bluetooth Headphones',
          productSku: 'PROD-001',
          warehouseId: '1',
          warehouseName: 'Main Distribution Center',
          quantity: 150,
          reservedQuantity: 25,
          availableQuantity: 125,
          status: 'available',
          lotNumber: 'LOT-2024-001',
          expirationDate: null,
          lastUpdated: '2024-01-15T10:30:00Z'
        },
        {
          id: '2',
          productId: '2',
          productName: 'Cotton T-Shirt',
          productSku: 'PROD-002',
          warehouseId: '1',
          warehouseName: 'Main Distribution Center',
          quantity: 500,
          reservedQuantity: 100,
          availableQuantity: 400,
          status: 'available',
          lotNumber: null,
          expirationDate: null,
          lastUpdated: '2024-01-14T16:45:00Z'
        },
        {
          id: '3',
          productId: '3',
          productName: 'Organic Coffee Beans',
          productSku: 'PROD-003',
          warehouseId: '2',
          warehouseName: 'West Coast Hub',
          quantity: 50,
          reservedQuantity: 42,
          availableQuantity: 8,
          status: 'reserved',
          lotNumber: 'LOT-2024-003',
          expirationDate: '2024-12-31',
          lastUpdated: '2024-01-10T09:15:00Z'
        }
      ];
      
      totalItems = items.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load inventory';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadInventory();
  }

  function handleRowClick(item: InventoryItem) {
    window.location.href = `/inventory/items/${item.id}`;
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadInventory();
  }

  onMount(() => {
    loadInventory();
  });
</script>

<svelte:head>
  <title>Inventory | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Inventory</h1>
      <p class="page-description">Track stock levels and manage inventory across warehouses</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={() => window.location.href = '/inventory/adjustments'}>
        New Adjustment
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null}>
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
        <p>Loading inventory...</p>
      </div>
    {:else if items.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
        <p class="text-gray-500">No inventory items found</p>
      </div>
    {:else}
      <Table {columns}>
        <tbody>
          {#each items as item}
            <tr on:click={() => handleRowClick(item)} class="clickable-row">
              <td class="font-medium">{item.productSku}</td>
              <td>{item.productName}</td>
              <td>{item.warehouseName}</td>
              <td class="font-medium">{item.quantity}</td>
              <td>{item.reservedQuantity}</td>
              <td class={item.availableQuantity < 10 ? 'text-red-600 font-medium' : ''}>
                {item.availableQuantity}
              </td>
              <td>
                <Badge variant={getStatusVariant(item.status)}>
                  {item.status.replace('_', ' ')}
                </Badge>
              </td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm">
                    View
                  </Button>
                </div>
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

  .actions-cell {
    display: flex;
    gap: 0.5rem;
  }
</style>
