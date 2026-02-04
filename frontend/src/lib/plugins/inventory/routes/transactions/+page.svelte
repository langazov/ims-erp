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

  interface InventoryTransaction {
    id: string;
    date: string;
    type: 'receipt' | 'shipment' | 'transfer' | 'adjustment' | 'return';
    reference: string;
    productName: string;
    productSku: string;
    warehouseName: string;
    quantity: number;
    user: string;
    notes: string;
  }

  let transactions: InventoryTransaction[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let typeFilter = '';
  let warehouseFilter = '';
  let startDate = '';
  let endDate = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;

  const typeOptions = [
    { value: '', label: 'All Types' },
    { value: 'receipt', label: 'Receipt' },
    { value: 'shipment', label: 'Shipment' },
    { value: 'transfer', label: 'Transfer' },
    { value: 'adjustment', label: 'Adjustment' },
    { value: 'return', label: 'Return' }
  ];

  const warehouseOptions = [
    { value: '', label: 'All Warehouses' },
    { value: '1', label: 'Main Distribution Center' },
    { value: '2', label: 'West Coast Hub' },
    { value: '3', label: 'Retail Store - Downtown' }
  ];

  const columns = [
    { key: 'date', label: 'Date', sortable: true },
    { key: 'type', label: 'Type', sortable: true },
    { key: 'reference', label: 'Reference', sortable: true },
    { key: 'product', label: 'Product', sortable: true },
    { key: 'warehouse', label: 'Warehouse', sortable: true },
    { key: 'quantity', label: 'Qty', sortable: true, align: 'right' as const },
    { key: 'user', label: 'User', sortable: true }
  ];

  function getTypeVariant(type: string): 'green' | 'red' | 'blue' | 'yellow' | 'purple' | 'gray' {
    switch (type) {
      case 'receipt': return 'green';
      case 'shipment': return 'red';
      case 'transfer': return 'blue';
      case 'adjustment': return 'yellow';
      case 'return': return 'purple';
      default: return 'gray';
    }
  }

  async function loadTransactions() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      transactions = [
        {
          id: '1',
          date: '2024-01-16T10:30:00Z',
          type: 'receipt',
          reference: 'PO-2024-001',
          productName: 'Wireless Bluetooth Headphones',
          productSku: 'PROD-001',
          warehouseName: 'Main Distribution Center',
          quantity: 100,
          user: 'John Smith',
          notes: 'Regular stock receipt'
        },
        {
          id: '2',
          date: '2024-01-16T09:15:00Z',
          type: 'shipment',
          reference: 'SO-2024-045',
          productName: 'Cotton T-Shirt',
          productSku: 'PROD-002',
          warehouseName: 'Main Distribution Center',
          quantity: -50,
          user: 'Sarah Johnson',
          notes: 'Customer order fulfillment'
        },
        {
          id: '3',
          date: '2024-01-15T14:20:00Z',
          type: 'transfer',
          reference: 'TF-2024-012',
          productName: 'Organic Coffee Beans',
          productSku: 'PROD-003',
          warehouseName: 'West Coast Hub',
          quantity: 25,
          user: 'Mike Davis',
          notes: 'Transfer from Main DC'
        },
        {
          id: '4',
          date: '2024-01-15T11:00:00Z',
          type: 'adjustment',
          reference: 'ADJ-2024-003',
          productName: 'Premium Chocolate Bar',
          productSku: 'PROD-005',
          warehouseName: 'Main Distribution Center',
          quantity: -5,
          user: 'Lisa Wilson',
          notes: 'Inventory count adjustment - damaged goods'
        },
        {
          id: '5',
          date: '2024-01-14T16:45:00Z',
          type: 'return',
          reference: 'RET-2024-008',
          productName: 'Wireless Bluetooth Headphones',
          productSku: 'PROD-001',
          warehouseName: 'Main Distribution Center',
          quantity: 3,
          user: 'John Smith',
          notes: 'Customer return - defective'
        },
        {
          id: '6',
          date: '2024-01-14T10:00:00Z',
          type: 'receipt',
          reference: 'PO-2024-002',
          productName: 'Fresh Milk 1L',
          productSku: 'PROD-004',
          warehouseName: 'Retail Store - Downtown',
          quantity: 48,
          user: 'Emma Brown',
          notes: 'Daily dairy delivery'
        },
        {
          id: '7',
          date: '2024-01-13T13:30:00Z',
          type: 'shipment',
          reference: 'SO-2024-044',
          productName: 'Organic Coffee Beans',
          productSku: 'PROD-003',
          warehouseName: 'West Coast Hub',
          quantity: -12,
          user: 'Sarah Johnson',
          notes: 'Wholesale order'
        },
        {
          id: '8',
          date: '2024-01-12T09:00:00Z',
          type: 'adjustment',
          reference: 'ADJ-2024-002',
          productName: 'Cotton T-Shirt',
          productSku: 'PROD-002',
          warehouseName: 'Main Distribution Center',
          quantity: 10,
          user: 'Mike Davis',
          notes: 'Found during cycle count'
        }
      ];
      
      totalItems = transactions.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load transactions';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadTransactions();
  }

  function handleRowClick(transaction: InventoryTransaction) {
    console.log('View transaction:', transaction.id);
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadTransactions();
  }

  function formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function clearFilters() {
    searchQuery = '';
    typeFilter = '';
    warehouseFilter = '';
    startDate = '';
    endDate = '';
    handleSearch();
  }

  onMount(() => {
    loadTransactions();
  });
</script>

<svelte:head>
  <title>Inventory Transactions | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Inventory Transactions</h1>
      <p class="page-description">View transaction history and stock movements</p>
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
            placeholder="Search by reference or product..."
            bind:value={searchQuery}
            on:keydown={(e) => e.key === 'Enter' && handleSearch()}
          />
        </div>
        <div class="filter-item">
          <Select
            id="type"
            label="Type"
            options={typeOptions}
            bind:value={typeFilter}
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
      </div>
      
      <div class="filter-row date-filters">
        <div class="filter-item">
          <Input
            id="startDate"
            label="Start Date"
            type="date"
            bind:value={startDate}
          />
        </div>
        <div class="filter-item">
          <Input
            id="endDate"
            label="End Date"
            type="date"
            bind:value={endDate}
          />
        </div>
        <div class="filter-item filter-actions">
          <Button variant="secondary" on:click={handleSearch}>
            Search
          </Button>
          <Button variant="ghost" on:click={clearFilters}>
            Clear
          </Button>
        </div>
      </div>
    </div>

    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading transactions...</p>
      </div>
    {:else if transactions.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
        </svg>
        <p class="text-gray-500">No transactions found</p>
      </div>
    {:else}
      <Table columns={columns} data={transactions}>
        <tbody>
          {#each transactions as transaction}
            <tr on:click={() => handleRowClick(transaction)} class="clickable-row">
              <td>{formatDate(transaction.date)}</td>
              <td>
                <Badge variant={getTypeVariant(transaction.type)}>
                  {transaction.type}
                </Badge>
              </td>
              <td class="font-medium">{transaction.reference}</td>
              <td>
                <div class="product-cell">
                  <span class="product-name">{transaction.productName}</span>
                  <span class="product-sku">{transaction.productSku}</span>
                </div>
              </td>
              <td>{transaction.warehouseName}</td>
              <td class="text-right">
                <span class={transaction.quantity > 0 ? 'text-green-600' : 'text-red-600'}>
                  {transaction.quantity > 0 ? '+' : ''}{transaction.quantity}
                </span>
              </td>
              <td>{transaction.user}</td>
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
    margin-bottom: 0.75rem;
  }

  .filter-row:last-child {
    margin-bottom: 0;
  }

  .filter-item {
    flex: 1;
    min-width: 200px;
  }

  .search-filter {
    flex: 2;
    min-width: 300px;
  }

  .date-filters .filter-item {
    min-width: 150px;
  }

  .filter-actions {
    display: flex;
    align-items: flex-end;
    gap: 0.5rem;
    padding-bottom: 0.25rem;
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

  .product-cell {
    display: flex;
    flex-direction: column;
  }

  .product-name {
    font-weight: 500;
  }

  .product-sku {
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  @media (max-width: 768px) {
    .filter-row {
      flex-direction: column;
    }

    .filter-item,
    .search-filter {
      min-width: 100%;
    }

    .filter-actions {
      padding-bottom: 0;
    }
  }
</style>
