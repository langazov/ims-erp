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

  interface OrderItem {
    id: string;
    productId: string;
    productName: string;
    quantity: number;
    unitPrice: number;
    total: number;
  }

  interface Order {
    id: string;
    number: string;
    clientId: string;
    clientName: string;
    orderDate: string;
    status: 'pending' | 'processing' | 'shipped' | 'delivered' | 'cancelled' | 'returned';
    total: number;
    items: OrderItem[];
    shippingAddress: string;
    trackingNumber: string;
  }

  let orders: Order[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter: string = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;
  let showCreateModal = false;
  let showViewModal = false;
  let selectedOrder: Order | null = null;
  let deleteOrderId: string | null = null;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'pending', label: 'Pending' },
    { value: 'processing', label: 'Processing' },
    { value: 'shipped', label: 'Shipped' },
    { value: 'delivered', label: 'Delivered' },
    { value: 'cancelled', label: 'Cancelled' },
    { value: 'returned', label: 'Returned' }
  ];

  const columns = [
    { key: 'number', label: 'Order #', sortable: true },
    { key: 'client', label: 'Client', sortable: true },
    { key: 'orderDate', label: 'Order Date', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'items', label: 'Items', sortable: false },
    { key: 'total', label: 'Total', sortable: true },
    { key: 'tracking', label: 'Tracking', sortable: false },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  function getStatusVariant(status: string): 'green' | 'gray' | 'yellow' | 'red' | 'blue' | 'purple' {
    switch (status) {
      case 'delivered': return 'green';
      case 'shipped': return 'blue';
      case 'processing': return 'yellow';
      case 'pending': return 'gray';
      case 'cancelled': return 'red';
      case 'returned': return 'purple';
      default: return 'gray';
    }
  }

  function formatCurrency(value: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(value);
  }

  function formatDate(date: string): string {
    return new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  async function loadOrders() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      orders = [
        {
          id: '1',
          number: 'ORD-2024-001',
          clientId: '1',
          clientName: 'Acme Corporation',
          orderDate: '2024-01-15',
          status: 'delivered',
          total: 1250.00,
          items: [
            { id: '1', productId: '101', productName: 'Office Chair', quantity: 5, unitPrice: 150.00, total: 750.00 },
            { id: '2', productId: '102', productName: 'Desk Lamp', quantity: 10, unitPrice: 50.00, total: 500.00 }
          ],
          shippingAddress: '123 Business Ave, Suite 100, New York, NY 10001',
          trackingNumber: '1Z999AA10123456784'
        },
        {
          id: '2',
          number: 'ORD-2024-002',
          clientId: '2',
          clientName: 'TechStart Inc',
          orderDate: '2024-01-20',
          status: 'shipped',
          total: 3400.00,
          items: [
            { id: '3', productId: '103', productName: 'Laptop Stand', quantity: 20, unitPrice: 120.00, total: 2400.00 },
            { id: '4', productId: '104', productName: 'Wireless Mouse', quantity: 20, unitPrice: 50.00, total: 1000.00 }
          ],
          shippingAddress: '456 Tech Blvd, Floor 3, San Francisco, CA 94105',
          trackingNumber: '1Z888BB20234567890'
        },
        {
          id: '3',
          number: 'ORD-2024-003',
          clientId: '3',
          clientName: 'Global Solutions LLC',
          orderDate: '2024-01-22',
          status: 'processing',
          total: 2750.00,
          items: [
            { id: '5', productId: '105', productName: 'Conference Phone', quantity: 5, unitPrice: 400.00, total: 2000.00 },
            { id: '6', productId: '106', productName: 'Webcam HD', quantity: 5, unitPrice: 150.00, total: 750.00 }
          ],
          shippingAddress: '789 Global Way, Building B, Chicago, IL 60601',
          trackingNumber: ''
        },
        {
          id: '4',
          number: 'ORD-2024-004',
          clientId: '1',
          clientName: 'Acme Corporation',
          orderDate: '2024-01-25',
          status: 'pending',
          total: 890.00,
          items: [
            { id: '7', productId: '107', productName: 'Monitor 27"', quantity: 2, unitPrice: 350.00, total: 700.00 },
            { id: '8', productId: '108', productName: 'Keyboard Mechanical', quantity: 3, unitPrice: 63.33, total: 190.00 }
          ],
          shippingAddress: '123 Business Ave, Suite 100, New York, NY 10001',
          trackingNumber: ''
        },
        {
          id: '5',
          number: 'ORD-2024-005',
          clientId: '4',
          clientName: 'Digital Ventures Co',
          orderDate: '2024-01-10',
          status: 'cancelled',
          total: 1500.00,
          items: [
            { id: '9', productId: '109', productName: 'Projector 4K', quantity: 1, unitPrice: 1200.00, total: 1200.00 },
            { id: '10', productId: '110', productName: 'Screen 100"', quantity: 1, unitPrice: 300.00, total: 300.00 }
          ],
          shippingAddress: '321 Digital Dr, Austin, TX 78701',
          trackingNumber: ''
        }
      ];
      
      totalItems = orders.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load orders';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadOrders();
  }

  function handleView(order: Order, event: Event) {
    event.stopPropagation();
    selectedOrder = order;
    showViewModal = true;
  }

  function handleEdit(order: Order, event: Event) {
    event.stopPropagation();
    window.location.href = `/orders/${order.id}/edit`;
  }

  async function handleDelete(order: Order, event: Event) {
    event.stopPropagation();
    deleteOrderId = order.id;
  }

  async function confirmDelete() {
    if (!deleteOrderId) return;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      orders = orders.filter(o => o.id !== deleteOrderId);
      totalItems = orders.length;
      deleteOrderId = null;
    } catch (err) {
      error = 'Failed to delete order';
    }
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadOrders();
  }

  onMount(() => {
    loadOrders();
  });
</script>

<svelte:head>
  <title>Orders | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Orders</h1>
      <p class="page-description">Manage customer orders and fulfillment</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={() => showCreateModal = true}>
        Create Order
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
            placeholder="Search orders..."
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
          <Button variant="secondary" on:click={handleSearch}>
            Search
          </Button>
        </div>
      </div>
    </div>

    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading orders...</p>
      </div>
    {:else if orders.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
        </svg>
        <p class="text-gray-500 mb-4">No orders found</p>
        <Button variant="primary" on:click={() => showCreateModal = true}>
          Create Your First Order
        </Button>
      </div>
    {:else}
      <Table {columns}>
        <tbody>
          {#each orders as order}
            <tr class="clickable-row">
              <td class="font-medium">{order.number}</td>
              <td>{order.clientName}</td>
              <td>{formatDate(order.orderDate)}</td>
              <td>
                <Badge variant={getStatusVariant(order.status)}>
                  {order.status}
                </Badge>
              </td>
              <td>{order.items.length} item{order.items.length !== 1 ? 's' : ''}</td>
              <td class="font-medium">{formatCurrency(order.total)}</td>
              <td>
                {#if order.trackingNumber}
                  <span class="tracking-number">{order.trackingNumber}</span>
                {:else}
                  <span class="text-gray-400">-</span>
                {/if}
              </td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={(e) => handleView(order, e)}>
                    View
                  </Button>
                  <Button variant="ghost" size="sm" on:click={(e) => handleEdit(order, e)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={(e) => handleDelete(order, e)}>
                    Delete
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

<Modal
  bind:open={showCreateModal}
  title="Create Order"
  size="lg"
>
  <p class="text-gray-600">Order creation form will be implemented here.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close}>Cancel</Button>
    <Button variant="primary" on:click={() => {
      showCreateModal = false;
    }}>Create Order</Button>
  </svelte:fragment>
</Modal>

{#if selectedOrder}
  <Modal
    bind:open={showViewModal}
    title={`Order ${selectedOrder.number}`}
    size="lg"
  >
    <div class="order-details">
      <div class="detail-section">
        <h3 class="section-title">Order Information</h3>
        <div class="detail-grid">
          <div class="detail-item">
            <span class="detail-label">Client</span>
            <span class="detail-value">{selectedOrder.clientName}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Order Date</span>
            <span class="detail-value">{formatDate(selectedOrder.orderDate)}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Status</span>
            <span class="detail-value">
              <Badge variant={getStatusVariant(selectedOrder.status)}>
                {selectedOrder.status}
              </Badge>
            </span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Total</span>
            <span class="detail-value font-medium">{formatCurrency(selectedOrder.total)}</span>
          </div>
        </div>
      </div>

      <div class="detail-section">
        <h3 class="section-title">Shipping Address</h3>
        <p class="shipping-address">{selectedOrder.shippingAddress}</p>
        {#if selectedOrder.trackingNumber}
          <div class="tracking-info">
            <span class="detail-label">Tracking Number:</span>
            <span class="tracking-number">{selectedOrder.trackingNumber}</span>
          </div>
        {/if}
      </div>

      <div class="detail-section">
        <h3 class="section-title">Order Items</h3>
        <table class="items-table">
          <thead>
            <tr>
              <th>Product</th>
              <th>Quantity</th>
              <th>Unit Price</th>
              <th>Total</th>
            </tr>
          </thead>
          <tbody>
            {#each selectedOrder.items as item}
              <tr>
                <td>{item.productName}</td>
                <td>{item.quantity}</td>
                <td>{formatCurrency(item.unitPrice)}</td>
                <td class="font-medium">{formatCurrency(item.total)}</td>
              </tr>
            {/each}
          </tbody>
          <tfoot>
            <tr>
              <td colspan="3" class="text-right font-medium">Total:</td>
              <td class="font-bold">{formatCurrency(selectedOrder.total)}</td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
    
    <svelte:fragment slot="footer" let:close>
      <Button variant="secondary" on:click={close}>Close</Button>
      <Button variant="primary" on:click={() => {
        showViewModal = false;
        window.location.href = `/orders/${selectedOrder?.id}/edit`;
      }}>Edit Order</Button>
    </svelte:fragment>
  </Modal>
{/if}

{#if deleteOrderId}
  <Modal
    open={true}
    title="Delete Order"
    size="sm"
  >
    <p>Are you sure you want to delete this order? This action cannot be undone.</p>
    
    <svelte:fragment slot="footer" let:close>
      <Button variant="secondary" on:click={() => { close(); deleteOrderId = null; }}>Cancel</Button>
      <Button variant="danger" on:click={confirmDelete}>Delete</Button>
    </svelte:fragment>
  </Modal>
{/if}

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

  .tracking-number {
    font-family: monospace;
    font-size: 0.875rem;
    color: var(--color-blue-600);
  }

  .order-details {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .detail-section {
    border-bottom: 1px solid var(--color-gray-200);
    padding-bottom: 1.5rem;
  }

  .detail-section:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }

  .section-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
  }

  .detail-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .detail-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .detail-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .detail-value {
    font-size: 0.875rem;
    color: var(--color-gray-900);
  }

  .shipping-address {
    font-size: 0.875rem;
    color: var(--color-gray-700);
    line-height: 1.5;
    margin: 0 0 0.75rem 0;
  }

  .tracking-info {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
  }

  .items-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .items-table th,
  .items-table td {
    padding: 0.75rem;
    text-align: left;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .items-table th {
    font-weight: 600;
    color: var(--color-gray-700);
    background-color: var(--color-gray-50);
  }

  .items-table tfoot td {
    border-top: 2px solid var(--color-gray-300);
    border-bottom: none;
    padding-top: 1rem;
  }

  .text-right {
    text-align: right;
  }
</style>
