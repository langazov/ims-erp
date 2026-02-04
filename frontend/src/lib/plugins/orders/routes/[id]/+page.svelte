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
  import { 
    getOrderById, 
    deleteOrder, 
    updateOrderStatus,
    type Order,
    type OrderStatus 
  } from '$lib/shared/api/orders';

  const orderId = $page.params.id;

  let order: Order | null = null;
  let loading = true;
  let error: string | null = null;
  let showDeleteModal = false;
  let showCancelModal = false;
  let showStatusModal = false;
  let nextStatus: OrderStatus | null = null;
  let processing = false;
  let actionError: string | null = null;

  async function loadOrder() {
    loading = true;
    error = null;
    
    try {
      order = await getOrderById(orderId);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load order';
    } finally {
      loading = false;
    }
  }

  function getStatusVariant(status: OrderStatus): 'green' | 'gray' | 'yellow' | 'red' | 'blue' | 'purple' {
    switch (status) {
      case 'delivered': return 'green';
      case 'shipped': return 'blue';
      case 'processing': return 'yellow';
      case 'confirmed': return 'purple';
      case 'pending': return 'gray';
      case 'cancelled': return 'red';
      default: return 'gray';
    }
  }

  function formatCurrency(value: string): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(parseFloat(value) || 0);
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }

  function formatDateTime(dateStr: string): string {
    return new Date(dateStr).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function formatAddress(address: { street: string; city: string; state: string; postalCode: string; country: string }): string {
    return `${address.street}, ${address.city}, ${address.state} ${address.postalCode}, ${address.country}`;
  }

  function handleEdit() {
    goto(`/orders/${orderId}/edit`);
  }

  function handleDelete() {
    showDeleteModal = true;
  }

  async function confirmDelete() {
    processing = true;
    actionError = null;
    try {
      await deleteOrder(orderId);
      goto('/orders');
    } catch (err) {
      actionError = err instanceof Error ? err.message : 'Failed to delete order';
      processing = false;
    }
  }

  function handleCancel() {
    showCancelModal = true;
  }

  async function confirmCancel() {
    processing = true;
    actionError = null;
    try {
      await updateOrderStatus(orderId, 'cancelled');
      await loadOrder();
      showCancelModal = false;
    } catch (err) {
      actionError = err instanceof Error ? err.message : 'Failed to cancel order';
    } finally {
      processing = false;
    }
  }

  function handleStatusUpdate(status: OrderStatus) {
    nextStatus = status;
    showStatusModal = true;
  }

  async function confirmStatusUpdate() {
    if (!nextStatus) return;
    
    processing = true;
    actionError = null;
    try {
      await updateOrderStatus(orderId, nextStatus);
      await loadOrder();
      showStatusModal = false;
      nextStatus = null;
    } catch (err) {
      actionError = err instanceof Error ? err.message : 'Failed to update order status';
    } finally {
      processing = false;
    }
  }

  function handleGenerateInvoice() {
    // TODO: Implement invoice generation
    alert('Invoice generation will be implemented here');
  }

  function handlePrintPickList() {
    // TODO: Implement pick list printing
    window.print();
  }

  function handleViewClient() {
    if (order) {
      goto(`/clients/${order.clientId}`);
    }
  }

  function getNextStatus(currentStatus: OrderStatus): OrderStatus | null {
    const statusFlow: Record<OrderStatus, OrderStatus | null> = {
      'pending': 'confirmed',
      'confirmed': 'processing',
      'processing': 'shipped',
      'shipped': 'delivered',
      'delivered': null,
      'cancelled': null
    };
    return statusFlow[currentStatus];
  }

  function getStatusActionLabel(status: OrderStatus): string {
    const labels: Record<OrderStatus, string> = {
      'pending': 'Confirm Order',
      'confirmed': 'Start Processing',
      'processing': 'Mark as Shipped',
      'shipped': 'Mark as Delivered',
      'delivered': 'Completed',
      'cancelled': 'Cancelled'
    };
    return labels[status];
  }

  function canEdit(status: OrderStatus): boolean {
    return status === 'pending' || status === 'confirmed';
  }

  function canCancel(status: OrderStatus): boolean {
    return status !== 'delivered' && status !== 'cancelled';
  }

  function canDelete(status: OrderStatus): boolean {
    return status === 'pending' || status === 'cancelled';
  }

  function canGenerateInvoice(status: OrderStatus): boolean {
    return status === 'confirmed' || status === 'processing' || status === 'shipped' || status === 'delivered';
  }

  // Mock activity timeline - in real implementation, this would come from API
  const activities = [
    { type: 'created', date: '2024-01-15T10:30:00Z', description: 'Order created', user: 'John Doe' },
    { type: 'updated', date: '2024-01-15T11:00:00Z', description: 'Order confirmed', user: 'Jane Smith' },
    { type: 'processing', date: '2024-01-16T09:00:00Z', description: 'Order processing started', user: 'Mike Johnson' }
  ];

  // Mock related invoices
  const relatedInvoices = [
    { id: 'INV-001', number: 'INV-2024-001', amount: '1250.00', status: 'paid', date: '2024-01-15' }
  ];

  onMount(() => {
    loadOrder();
  });
</script>

<svelte:head>
  <title>{order ? `Order ${order.orderNumber}` : 'Order Details'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading order details...</p>
    </div>
  {:else if error}
    <div class="error-container">
      <Alert variant="error">{error}</Alert>
      <Button variant="secondary" on:click={() => goto('/orders')}>
        Back to Orders
      </Button>
    </div>
  {:else if order}
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <h1 class="page-title">Order {order.orderNumber}</h1>
          <Badge variant={getStatusVariant(order.status)} size="md">
            {order.status}
          </Badge>
        </div>
        <p class="page-description">
          Created on {formatDate(order.createdAt)}
        </p>
      </div>
      <div class="header-actions">
        {#if canEdit(order.status)}
          <Button variant="secondary" on:click={handleEdit}>
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            Edit
          </Button>
        {/if}
        
        {#if getNextStatus(order.status)}
          <Button variant="primary" on:click={() => handleStatusUpdate(getNextStatus(order.status)!)}>
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {getStatusActionLabel(getNextStatus(order.status)!)}
          </Button>
        {/if}
        
        {#if canGenerateInvoice(order.status)}
          <Button variant="secondary" on:click={handleGenerateInvoice}>
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            Generate Invoice
          </Button>
        {/if}
        
        <Button variant="secondary" on:click={handlePrintPickList}>
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
          </svg>
          Print Pick List
        </Button>
        
        {#if canCancel(order.status)}
          <Button variant="danger" on:click={handleCancel}>
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            Cancel
          </Button>
        {/if}
        
        {#if canDelete(order.status)}
          <Button variant="danger" on:click={handleDelete}>
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Delete
          </Button>
        {/if}
      </div>
    </div>

    <div class="content-grid">
      <!-- Main Content -->
      <div class="main-content">
        <!-- Order Summary -->
        <Card>
          <div class="order-section">
            <h3 class="section-title">Order Information</h3>
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">Order Number</span>
                <span class="info-value">{order.orderNumber}</span>
              </div>
              <div class="info-item">
                <span class="info-label">Order Date</span>
                <span class="info-value">{formatDate(order.createdAt)}</span>
              </div>
              <div class="info-item">
                <span class="info-label">Status</span>
                <span class="info-value">
                  <Badge variant={getStatusVariant(order.status)} size="sm">
                    {order.status}
                  </Badge>
                </span>
              </div>
              <div class="info-item">
                <span class="info-label">Last Updated</span>
                <span class="info-value">{formatDateTime(order.updatedAt)}</span>
              </div>
            </div>
          </div>
        </Card>

        <!-- Client Information -->
        <Card>
          <div class="order-section">
            <div class="section-header">
              <h3 class="section-title">Client Information</h3>
              <Button variant="ghost" size="sm" on:click={handleViewClient}>
                View Client
                <svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </Button>
            </div>
            <div class="client-info">
              <p class="client-name">{order.clientName}</p>
              <p class="client-id">Client ID: {order.clientId}</p>
            </div>
          </div>
        </Card>

        <!-- Line Items -->
        <Card>
          <div class="order-section">
            <h3 class="section-title">Order Items</h3>
            <table class="items-table">
              <thead>
                <tr>
                  <th class="product">Product</th>
                  <th class="quantity">Quantity</th>
                  <th class="price">Unit Price</th>
                  <th class="total">Total</th>
                </tr>
              </thead>
              <tbody>
                {#each order.items as item}
                  <tr>
                    <td class="product">
                      <span class="product-name">{item.productName}</span>
                      <span class="product-id">SKU: {item.productId}</span>
                    </td>
                    <td class="quantity">{item.quantity}</td>
                    <td class="price">{formatCurrency(item.unitPrice)}</td>
                    <td class="total">{formatCurrency(item.total)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </Card>

        <!-- Shipping Address -->
        <Card>
          <div class="order-section">
            <h3 class="section-title">Shipping Address</h3>
            <div class="shipping-address">
              <p class="address-line">{order.shippingAddress.street}</p>
              <p class="address-line">{order.shippingAddress.city}, {order.shippingAddress.state} {order.shippingAddress.postalCode}</p>
              <p class="address-line">{order.shippingAddress.country}</p>
            </div>
          </div>
        </Card>

        <!-- Order Totals -->
        <Card>
          <div class="order-section">
            <h3 class="section-title">Order Totals</h3>
            <div class="totals-breakdown">
              <div class="total-row">
                <span class="label">Subtotal</span>
                <span class="value">{formatCurrency(order.subtotal)}</span>
              </div>
              <div class="total-row">
                <span class="label">Tax</span>
                <span class="value">{formatCurrency(order.tax)}</span>
              </div>
              <div class="total-row grand-total">
                <span class="label">Total</span>
                <span class="value">{formatCurrency(order.total)}</span>
              </div>
            </div>
          </div>
        </Card>

        <!-- Notes -->
        {#if order.notes}
          <Card>
            <div class="order-section">
              <h3 class="section-title">Notes</h3>
              <p class="notes-content">{order.notes}</p>
            </div>
          </Card>
        {/if}

        <!-- Related Invoices -->
        <Card>
          <div class="order-section">
            <h3 class="section-title">Related Invoices</h3>
            {#if relatedInvoices.length > 0}
              <div class="invoices-list">
                {#each relatedInvoices as invoice}
                  <div class="invoice-item">
                    <div class="invoice-info">
                      <span class="invoice-number">{invoice.number}</span>
                      <span class="invoice-date">{formatDate(invoice.date)}</span>
                    </div>
                    <div class="invoice-amount">
                      <span class="amount">{formatCurrency(invoice.amount)}</span>
                      <Badge variant={invoice.status === 'paid' ? 'green' : 'yellow'} size="sm">
                        {invoice.status}
                      </Badge>
                    </div>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="empty-state">
                <p class="text-gray-500">No invoices generated yet</p>
              </div>
            {/if}
          </div>
        </Card>
      </div>

      <!-- Sidebar -->
      <div class="sidebar">
        <!-- Status Card -->
        <Card>
          <h3 class="sidebar-title">Status</h3>
          <div class="status-display">
            <Badge variant={getStatusVariant(order.status)} size="lg">
              {order.status}
            </Badge>
          </div>
          
          <!-- Status Workflow -->
          <div class="status-workflow">
            {#each ['pending', 'confirmed', 'processing', 'shipped', 'delivered'] as step}
              {@const isActive = 
                (step === 'pending' && ['pending', 'confirmed', 'processing', 'shipped', 'delivered'].includes(order.status)) ||
                (step === 'confirmed' && ['confirmed', 'processing', 'shipped', 'delivered'].includes(order.status)) ||
                (step === 'processing' && ['processing', 'shipped', 'delivered'].includes(order.status)) ||
                (step === 'shipped' && ['shipped', 'delivered'].includes(order.status)) ||
                (step === 'delivered' && order.status === 'delivered')
              }
              {@const isCurrent = order.status === step}
              <div class="workflow-step" class:active={isActive} class:current={isCurrent}>
                <div class="step-dot"></div>
                <span class="step-label">{step}</span>
              </div>
            {/each}
          </div>
          
          <div class="status-info">
            <div class="info-row">
              <span class="label">Created</span>
              <span class="value">{formatDateTime(order.createdAt)}</span>
            </div>
            <div class="info-row">
              <span class="label">Last Updated</span>
              <span class="value">{formatDateTime(order.updatedAt)}</span>
            </div>
          </div>
        </Card>

        <!-- Amount Summary -->
        <Card>
          <h3 class="sidebar-title">Amount Summary</h3>
          <div class="amount-breakdown">
            <div class="amount-row">
              <span class="label">Subtotal</span>
              <span class="value">{formatCurrency(order.subtotal)}</span>
            </div>
            <div class="amount-row">
              <span class="label">Tax</span>
              <span class="value">{formatCurrency(order.tax)}</span>
            </div>
            <div class="amount-row total">
              <span class="label">Total</span>
              <span class="value">{formatCurrency(order.total)}</span>
            </div>
          </div>
        </Card>

        <!-- Activity Timeline -->
        <Card>
          <h3 class="sidebar-title">Activity</h3>
          <div class="timeline">
            {#each activities as activity}
              <div class="timeline-item">
                <div class="timeline-dot" class:created={activity.type === 'created'} class:updated={activity.type === 'updated'} class:processing={activity.type === 'processing'}></div>
                <div class="timeline-content">
                  <p class="timeline-description">{activity.description}</p>
                  <p class="timeline-meta">{activity.user} • {formatDateTime(activity.date)}</p>
                </div>
              </div>
            {/each}
          </div>
        </Card>
      </div>
    </div>
  {:else}
    <Alert variant="error">Order not found</Alert>
  {/if}
</div>

<!-- Delete Modal -->
<Modal
  bind:open={showDeleteModal}
  title="Delete Order"
  size="sm"
>
  <p>Are you sure you want to delete order <strong>{order?.orderNumber}</strong>? This action cannot be undone.</p>
  
  {#if actionError}
    <Alert variant="error" class="mt-4">{actionError}</Alert>
  {/if}
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showDeleteModal = false; }} disabled={processing}>
      Cancel
    </Button>
    <Button variant="danger" on:click={confirmDelete} loading={processing}>
      {processing ? 'Deleting...' : 'Delete'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Cancel Modal -->
<Modal
  bind:open={showCancelModal}
  title="Cancel Order"
  size="sm"
>
  <p>Are you sure you want to cancel order <strong>{order?.orderNumber}</strong>?</p>
  <p class="text-sm text-gray-500 mt-2">This will mark the order as cancelled and release any reserved inventory.</p>
  
  {#if actionError}
    <Alert variant="error" class="mt-4">{actionError}</Alert>
  {/if}
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showCancelModal = false; }} disabled={processing}>
      Keep Order
    </Button>
    <Button variant="danger" on:click={confirmCancel} loading={processing}>
      {processing ? 'Cancelling...' : 'Cancel Order'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Status Update Modal -->
<Modal
  bind:open={showStatusModal}
  title="Update Order Status"
  size="sm"
>
  <p>Update order <strong>{order?.orderNumber}</strong> status to <strong>{nextStatus}</strong>?</p>
  <p class="text-sm text-gray-500 mt-2">This action will update the order status and may trigger notifications.</p>
  
  {#if actionError}
    <Alert variant="error" class="mt-4">{actionError}</Alert>
  {/if}
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showStatusModal = false; nextStatus = null; }} disabled={processing}>
      Cancel
    </Button>
    <Button variant="primary" on:click={confirmStatusUpdate} loading={processing}>
      {processing ? 'Updating...' : 'Update Status'}
    </Button>
  </svelte:fragment>
</Modal>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1400px;
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
    flex-wrap: wrap;
    gap: 1rem;
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
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
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .content-grid {
    display: grid;
    grid-template-columns: 1fr 320px;
    gap: 1.5rem;
  }

  .main-content {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .sidebar {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .sidebar-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .order-section {
    padding: 1.5rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
  }

  .section-header .section-title {
    margin: 0;
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .info-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .info-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .info-value {
    font-size: 0.875rem;
    color: var(--color-gray-900);
    font-weight: 500;
  }

  .client-info {
    padding: 1rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
  }

  .client-name {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 0.25rem 0;
  }

  .client-id {
    font-size: 0.875rem;
    color: var(--color-gray-500);
    margin: 0;
  }

  .items-table {
    width: 100%;
    border-collapse: collapse;
  }

  .items-table th {
    text-align: left;
    padding: 0.75rem;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--color-gray-500);
    border-bottom: 2px solid var(--color-gray-200);
  }

  .items-table td {
    padding: 0.75rem;
    border-bottom: 1px solid var(--color-gray-100);
    color: var(--color-gray-900);
  }

  .items-table .product {
    width: 50%;
  }

  .items-table .quantity,
  .items-table .price,
  .items-table .total {
    width: 16.67%;
    text-align: right;
  }

  .items-table .total {
    font-weight: 600;
  }

  .product-name {
    display: block;
    font-weight: 500;
  }

  .product-id {
    display: block;
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  .shipping-address {
    padding: 1rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
  }

  .address-line {
    margin: 0;
    color: var(--color-gray-700);
    line-height: 1.6;
  }

  .totals-breakdown {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .total-row {
    display: flex;
    justify-content: space-between;
    padding: 0.5rem 0;
  }

  .total-row:not(:last-child) {
    border-bottom: 1px solid var(--color-gray-100);
  }

  .total-row .label {
    color: var(--color-gray-600);
  }

  .total-row .value {
    font-weight: 500;
    color: var(--color-gray-900);
  }

  .total-row.grand-total {
    margin-top: 0.5rem;
    padding-top: 0.5rem;
    border-top: 2px solid var(--color-gray-200);
  }

  .total-row.grand-total .label,
  .total-row.grand-total .value {
    font-size: 1.25rem;
    font-weight: 700;
  }

  .total-row.grand-total .value {
    color: var(--color-primary-600);
  }

  .notes-content {
    color: var(--color-gray-700);
    line-height: 1.6;
    margin: 0;
    white-space: pre-wrap;
  }

  .invoices-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .invoice-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
  }

  .invoice-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .invoice-number {
    font-weight: 500;
    color: var(--color-gray-900);
  }

  .invoice-date {
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  .invoice-amount {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.25rem;
  }

  .amount {
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .empty-state {
    padding: 2rem;
    text-align: center;
    color: var(--color-gray-500);
  }

  .status-display {
    display: flex;
    justify-content: center;
    margin-bottom: 1rem;
  }

  .status-workflow {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
    padding: 1rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
  }

  .workflow-step {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    opacity: 0.4;
  }

  .workflow-step.active {
    opacity: 1;
  }

  .workflow-step.current .step-dot {
    background-color: var(--color-primary-600);
    box-shadow: 0 0 0 3px var(--color-primary-100);
  }

  .step-dot {
    width: 0.75rem;
    height: 0.75rem;
    border-radius: 50%;
    background-color: var(--color-gray-400);
  }

  .workflow-step.active .step-dot {
    background-color: var(--color-green-500);
  }

  .step-label {
    font-size: 0.875rem;
    color: var(--color-gray-700);
    text-transform: capitalize;
  }

  .workflow-step.current .step-label {
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .status-info {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    font-size: 0.875rem;
  }

  .info-row .label {
    color: var(--color-gray-500);
  }

  .info-row .value {
    color: var(--color-gray-900);
    font-weight: 500;
  }

  .amount-breakdown {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .amount-row {
    display: flex;
    justify-content: space-between;
    font-size: 0.875rem;
  }

  .amount-row .label {
    color: var(--color-gray-600);
  }

  .amount-row .value {
    color: var(--color-gray-900);
    font-weight: 500;
  }

  .amount-row.total {
    margin-top: 0.5rem;
    padding-top: 0.5rem;
    border-top: 1px solid var(--color-gray-200);
  }

  .amount-row.total .label,
  .amount-row.total .value {
    font-size: 1.125rem;
    font-weight: 700;
  }

  .amount-row.total .value {
    color: var(--color-primary-600);
  }

  .timeline {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .timeline-item {
    display: flex;
    gap: 0.75rem;
  }

  .timeline-dot {
    width: 0.75rem;
    height: 0.75rem;
    border-radius: 50%;
    background-color: var(--color-gray-300);
    margin-top: 0.375rem;
    flex-shrink: 0;
  }

  .timeline-dot.created {
    background-color: var(--color-blue-500);
  }

  .timeline-dot.updated {
    background-color: var(--color-yellow-500);
  }

  .timeline-dot.processing {
    background-color: var(--color-purple-500);
  }

  .timeline-content {
    flex: 1;
  }

  .timeline-description {
    font-size: 0.875rem;
    color: var(--color-gray-900);
    margin: 0 0 0.25rem 0;
  }

  .timeline-meta {
    font-size: 0.75rem;
    color: var(--color-gray-500);
    margin: 0;
  }

  @media (max-width: 1024px) {
    .content-grid {
      grid-template-columns: 1fr;
    }

    .sidebar {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 1rem;
    }

    .info-grid {
      grid-template-columns: 1fr;
    }

    .items-table {
      font-size: 0.875rem;
    }

    .items-table th,
    .items-table td {
      padding: 0.5rem 0.25rem;
    }

    .items-table .product {
      width: 40%;
    }
  }

  @media (max-width: 640px) {
    .page-header {
      flex-direction: column;
    }

    .header-actions {
      width: 100%;
      justify-content: flex-start;
    }

    .sidebar {
      grid-template-columns: 1fr;
    }

    .status-workflow {
      flex-direction: row;
      flex-wrap: wrap;
    }

    .workflow-step {
      flex: 1;
      min-width: 100px;
    }

    .invoice-item {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }

    .invoice-amount {
      align-items: flex-start;
    }
  }

  @media print {
    .page-header,
    .sidebar {
      display: none;
    }

    .content-grid {
      grid-template-columns: 1fr;
    }

    .main-content {
      gap: 1rem;
    }

    :global(.card) {
      break-inside: avoid;
    }
  }
</style>
