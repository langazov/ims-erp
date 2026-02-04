<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import Tabs from '$lib/shared/components/layout/Tabs.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import { ArrowLeft, Package, MapPin, History, TrendingUp, TrendingDown } from 'lucide-svelte';

  const itemId = $page.params.id;

  interface InventoryItem {
    id: string;
    productId: string;
    productName: string;
    sku: string;
    warehouseId: string;
    warehouseName: string;
    locationId: string;
    locationCode: string;
    quantity: number;
    reservedQuantity: number;
    availableQuantity: number;
    reorderPoint: number;
    reorderQuantity: number;
    unitCost: number;
    lastMovementAt?: string;
    status: 'in_stock' | 'low_stock' | 'out_of_stock';
  }

  interface Transaction {
    id: string;
    type: 'receipt' | 'issue' | 'transfer' | 'adjustment';
    quantity: number;
    reference: string;
    createdAt: string;
    createdBy: string;
  }

  let item: InventoryItem | null = null;
  let transactions: Transaction[] = [];
  let loading = true;
  let error: string | null = null;
  let activeTab = 'overview';
  let showReserveModal = false;
  let reserveQuantity = 1;
  let reserving = false;

  const transactionColumns = [
    { key: 'type', label: 'Type', sortable: true },
    { key: 'quantity', label: 'Quantity', sortable: true },
    { key: 'reference', label: 'Reference', sortable: true },
    { key: 'createdAt', label: 'Date', sortable: true },
  ];

  async function loadItem() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      item = {
        id: itemId,
        productId: 'prod-001',
        productName: 'Wireless Bluetooth Headphones',
        sku: 'WBH-001-BLK',
        warehouseId: 'wh-001',
        warehouseName: 'Main Distribution Center',
        locationId: 'loc-001',
        locationCode: 'A-01-05',
        quantity: 150,
        reservedQuantity: 25,
        availableQuantity: 125,
        reorderPoint: 50,
        reorderQuantity: 100,
        unitCost: 45.99,
        lastMovementAt: '2024-01-15T14:30:00Z',
        status: 'in_stock'
      };

      transactions = [
        {
          id: 'txn-001',
          type: 'receipt',
          quantity: 100,
          reference: 'PO-2024-001',
          createdAt: '2024-01-15T10:00:00Z',
          createdBy: 'John Doe'
        },
        {
          id: 'txn-002',
          type: 'issue',
          quantity: -25,
          reference: 'SO-2024-045',
          createdAt: '2024-01-14T16:00:00Z',
          createdBy: 'Jane Smith'
        },
        {
          id: 'txn-003',
          type: 'adjustment',
          quantity: -2,
          reference: 'ADJ-2024-001',
          createdAt: '2024-01-13T11:00:00Z',
          createdBy: 'Mike Johnson'
        }
      ];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load item';
    } finally {
      loading = false;
    }
  }

  function getStatusVariant(status: string): 'green' | 'yellow' | 'red' {
    switch (status) {
      case 'in_stock':
        return 'green';
      case 'low_stock':
        return 'yellow';
      case 'out_of_stock':
        return 'red';
      default:
        return 'green';
    }
  }

  function getStatusLabel(status: string): string {
    return status.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase());
  }

  function getTransactionIcon(type: string) {
    switch (type) {
      case 'receipt':
        return TrendingUp;
      case 'issue':
        return TrendingDown;
      default:
        return History;
    }
  }

  function formatDate(dateStr?: string): string {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function formatCurrency(value: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(value);
  }

  function handleBack() {
    goto('/inventory/items');
  }

  function handleReserve() {
    showReserveModal = true;
  }

  async function confirmReserve() {
    if (!item || reserveQuantity <= 0) return;
    
    reserving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      item.reservedQuantity += reserveQuantity;
      item.availableQuantity -= reserveQuantity;
      showReserveModal = false;
      reserveQuantity = 1;
    } catch (err) {
      error = 'Failed to reserve stock';
    } finally {
      reserving = false;
    }
  }

  onMount(() => {
    loadItem();
  });
</script>

<svelte:head>
  <title>Inventory Item | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading item details...</p>
    </div>
  {:else if item}
    <div class="page-header">
      <div class="header-content">
        <Button variant="ghost" on:click={handleBack} class="back-button">
          <ArrowLeft class="w-4 h-4 mr-2" />
          Back to Items
        </Button>
        <div class="header-title">
          <h1 class="page-title">{item.productName}</h1>
          <Badge variant={getStatusVariant(item.status)} size="md">
            {getStatusLabel(item.status)}
          </Badge>
        </div>
        <p class="page-description">
          SKU: {item.sku} • Location: {item.locationCode}
        </p>
      </div>
      <div class="header-actions">
        <Button variant="secondary" on:click={handleReserve}>
          Reserve Stock
        </Button>
      </div>
    </div>

    {#if error}
      <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
        {error}
      </Alert>
    {/if}

    <Tabs bind:activeTab tabs={[
      { id: 'overview', label: 'Overview' },
      { id: 'transactions', label: 'Transactions' },
      { id: 'reservations', label: 'Reservations' }
    ]}>
      {#if activeTab === 'overview'}
        <div class="overview-grid">
          <Card class="stats-card">
            <div class="stat-item">
              <div class="stat-icon blue">
                <Package class="w-6 h-6" />
              </div>
              <div class="stat-content">
                <span class="stat-label">Total Quantity</span>
                <span class="stat-value">{item.quantity}</span>
              </div>
            </div>
          </Card>

          <Card class="stats-card">
            <div class="stat-item">
              <div class="stat-icon green">
                <TrendingUp class="w-6 h-6" />
              </div>
              <div class="stat-content">
                <span class="stat-label">Available</span>
                <span class="stat-value">{item.availableQuantity}</span>
              </div>
            </div>
          </Card>

          <Card class="stats-card">
            <div class="stat-item">
              <div class="stat-icon yellow">
                <TrendingDown class="w-6 h-6" />
              </div>
              <div class="stat-content">
                <span class="stat-label">Reserved</span>
                <span class="stat-value">{item.reservedQuantity}</span>
              </div>
            </div>
          </Card>

          <Card class="stats-card">
            <div class="stat-item">
              <div class="stat-icon purple">
                <MapPin class="w-6 h-6" />
              </div>
              <div class="stat-content">
                <span class="stat-label">Unit Cost</span>
                <span class="stat-value">{formatCurrency(item.unitCost)}</span>
              </div>
            </div>
          </Card>
        </div>

        <div class="details-grid">
          <Card>
            <h2 class="section-title">Stock Levels</h2>
            <div class="stock-levels">
              <div class="stock-bar-container">
                <div class="stock-bar-labels">
                  <span>Available</span>
                  <span>{Math.round((item.availableQuantity / item.quantity) * 100)}%</span>
                </div>
                <div class="stock-bar">
                  <div 
                    class="stock-fill available"
                    style="width: {(item.availableQuantity / item.quantity) * 100}%"
                  ></div>
                  <div 
                    class="stock-fill reserved"
                    style="width: {(item.reservedQuantity / item.quantity) * 100}%"
                  ></div>
                </div>
                <div class="stock-legend">
                  <div class="legend-item">
                    <div class="legend-dot available"></div>
                    <span>Available ({item.availableQuantity})</span>
                  </div>
                  <div class="legend-item">
                    <div class="legend-dot reserved"></div>
                    <span>Reserved ({item.reservedQuantity})</span>
                  </div>
                </div>
              </div>

              <div class="reorder-info">
                <div class="reorder-item">
                  <span class="reorder-label">Reorder Point</span>
                  <span class="reorder-value">{item.reorderPoint} units</span>
                </div>
                <div class="reorder-item">
                  <span class="reorder-label">Reorder Quantity</span>
                  <span class="reorder-value">{item.reorderQuantity} units</span>
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <h2 class="section-title">Location</h2>
            <div class="location-info">
              <div class="location-item">
                <span class="location-label">Warehouse</span>
                <span class="location-value">{item.warehouseName}</span>
              </div>
              <div class="location-item">
                <span class="location-label">Location Code</span>
                <span class="location-value">{item.locationCode}</span>
              </div>
              <div class="location-item">
                <span class="location-label">Last Movement</span>
                <span class="location-value">{formatDate(item.lastMovementAt)}</span>
              </div>
            </div>
          </Card>
        </div>
      {:else if activeTab === 'transactions'}
        <Card>
          <Table columns={transactionColumns} data={transactions}>
            <svelte:fragment slot="cell" let:column let:row>
              {#if column.key === 'type'}
                <div class="flex items-center gap-2">
                  <svelte:component this={getTransactionIcon(row.type)} class="w-4 h-4 text-gray-500" />
                  <span class="capitalize">{row.type}</span>
                </div>
              {:else if column.key === 'quantity'}
                <span class={row.quantity > 0 ? 'text-green-600' : 'text-red-600'}>
                  {row.quantity > 0 ? '+' : ''}{row.quantity}
                </span>
              {:else if column.key === 'createdAt'}
                {formatDate(row.createdAt)}
              {:else}
                {row[column.key]}
              {/if}
            </svelte:fragment>
          </Table>
        </Card>
      {:else if activeTab === 'reservations'}
        <Card>
          <div class="empty-state">
            <p>Reservation management coming soon.</p>
          </div>
        </Card>
      {/if}
    </Tabs>
  {:else}
    <Alert variant="error">Item not found</Alert>
  {/if}
</div>

<Modal
  bind:open={showReserveModal}
  title="Reserve Stock"
  size="sm"
>
  <div class="reserve-form">
    <p class="mb-4">
      Available to reserve: <strong>{item?.availableQuantity}</strong> units
    </p>
    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
      Quantity to Reserve
    </label>
    <input
      type="number"
      bind:value={reserveQuantity}
      min="1"
      max={item?.availableQuantity}
      class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-700 dark:text-white"
    />
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showReserveModal = false; }} disabled={reserving}>
      Cancel
    </Button>
    <Button
      variant="primary"
      on:click={confirmReserve}
      loading={reserving}
      disabled={reserveQuantity <= 0 || reserveQuantity > (item?.availableQuantity || 0)}
    >
      {reserving ? 'Reserving...' : 'Reserve'}
    </Button>
  </svelte:fragment>
</Modal>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .loading-container {
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
  }

  .header-content {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .back-button {
    align-self: flex-start;
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .page-title {
    font-size: 1.875rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin: 0;
  }

  .page-description {
    color: var(--color-gray-500);
    margin: 0;
  }

  .header-actions {
    display: flex;
    gap: 0.5rem;
  }

  .overview-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  :global(.stats-card) {
    padding: 1.25rem;
  }

  .stat-item {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .stat-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 3rem;
    height: 3rem;
    border-radius: 0.75rem;
  }

  .stat-icon.blue {
    background-color: var(--color-blue-100);
    color: var(--color-blue-600);
  }

  .stat-icon.green {
    background-color: var(--color-green-100);
    color: var(--color-green-600);
  }

  .stat-icon.yellow {
    background-color: var(--color-yellow-100);
    color: var(--color-yellow-600);
  }

  .stat-icon.purple {
    background-color: var(--color-purple-100);
    color: var(--color-purple-600);
  }

  .stat-content {
    display: flex;
    flex-direction: column;
  }

  .stat-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .stat-value {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-gray-900);
  }

  .details-grid {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 1rem;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin-bottom: 1rem;
  }

  .stock-levels {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .stock-bar-container {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .stock-bar-labels {
    display: flex;
    justify-content: space-between;
    font-size: 0.875rem;
    color: var(--color-gray-600);
  }

  .stock-bar {
    height: 1.5rem;
    background-color: var(--color-gray-200);
    border-radius: 9999px;
    overflow: hidden;
    display: flex;
  }

  .stock-fill {
    height: 100%;
    transition: width 0.3s ease;
  }

  .stock-fill.available {
    background-color: var(--color-green-500);
  }

  .stock-fill.reserved {
    background-color: var(--color-yellow-500);
  }

  .stock-legend {
    display: flex;
    gap: 1rem;
    margin-top: 0.5rem;
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: var(--color-gray-600);
  }

  .legend-dot {
    width: 0.75rem;
    height: 0.75rem;
    border-radius: 9999px;
  }

  .legend-dot.available {
    background-color: var(--color-green-500);
  }

  .legend-dot.reserved {
    background-color: var(--color-yellow-500);
  }

  .reorder-info {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--color-gray-200);
  }

  .reorder-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .reorder-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .reorder-value {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .location-info {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .location-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .location-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .location-value {
    font-size: 1rem;
    font-weight: 500;
    color: var(--color-gray-900);
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: var(--color-gray-500);
  }

  .reserve-form {
    padding: 1rem 0;
  }

  @media (max-width: 768px) {
    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .overview-grid {
      grid-template-columns: repeat(2, 1fr);
    }

    .details-grid {
      grid-template-columns: 1fr;
    }

    .reorder-info {
      grid-template-columns: 1fr;
    }
  }
</style>

