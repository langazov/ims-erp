<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import ConfirmModal from '$lib/shared/components/layout/ConfirmModal.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import {
    getOrderById,
    updateOrderStatus,
    type Order,
    type OrderStatus
  } from '$lib/shared/api/orders';
  import { toast } from '$lib/shared/stores/toast';

  let order: Order | null = null;
  let loading = true;
  let error: string | null = null;
  let actionLoading = false;

  let cancelModalOpen = false;
  let cancelling = false;

  let shipModalOpen = false;
  let trackingNumber = '';
  let shipping = false;

  $: orderId = $page.params.id;

  const statusSteps: OrderStatus[] = ['pending', 'confirmed', 'processing', 'shipped', 'delivered'];

  $: currentStepIndex = order ? statusSteps.indexOf(order.status) : -1;

  function getStepState(stepStatus: OrderStatus): 'complete' | 'current' | 'future' {
    if (!order) return 'future';
    const stepIndex = statusSteps.indexOf(stepStatus);
    if (stepIndex < currentStepIndex) return 'complete';
    if (stepIndex === currentStepIndex) return 'current';
    return 'future';
  }

  function formatCurrency(s: string | number): string {
    const num = typeof s === 'string' ? parseFloat(s) : s;
    if (isNaN(num)) return String(s);
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(num);
  }

  function formatDate(s: string): string {
    return new Date(s).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
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

  function capitalizeFirst(s: string): string {
    return s.charAt(0).toUpperCase() + s.slice(1);
  }

  async function loadOrder() {
    loading = true;
    error = null;
    try {
      order = await getOrderById(orderId);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load order';
    } finally {
      loading = false;
    }
  }

  async function performStatusUpdate(status: OrderStatus, successMsg: string) {
    if (!order) return;
    actionLoading = true;
    try {
      await updateOrderStatus(order.id, status);
      toast.success(successMsg);
      await loadOrder();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to update order status');
    } finally {
      actionLoading = false;
    }
  }

  async function handleConfirmOrder() {
    await performStatusUpdate('confirmed', 'Order confirmed');
  }

  async function handleMarkProcessing() {
    await performStatusUpdate('processing', 'Order marked as processing');
  }

  async function handleMarkDelivered() {
    await performStatusUpdate('delivered', 'Order marked as delivered');
  }

  async function handleShipSubmit() {
    if (!order) return;
    shipping = true;
    try {
      await updateOrderStatus(order.id, 'shipped');
      toast.success('Order marked as shipped');
      shipModalOpen = false;
      trackingNumber = '';
      await loadOrder();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to mark order as shipped');
    } finally {
      shipping = false;
    }
  }

  async function handleCancelConfirm() {
    if (!order) return;
    cancelling = true;
    try {
      await updateOrderStatus(order.id, 'cancelled');
      toast.success('Order cancelled');
      cancelModalOpen = false;
      await loadOrder();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to cancel order');
    } finally {
      cancelling = false;
    }
  }

  onMount(loadOrder);
</script>

<div class="page-container">
  {#if loading}
    <div class="spinner-wrap">
      <Spinner size="lg" />
    </div>
  {:else if error}
    <div class="error-wrap">
      <Alert variant="error">{error}</Alert>
      <Button variant="secondary" on:click={() => goto('/orders')}>← Back to Orders</Button>
    </div>
  {:else if order}
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <Button variant="ghost" size="sm" on:click={() => goto('/orders')}>← Back to Orders</Button>
        <div class="header-info">
          <div class="header-top">
            <h1 class="order-number">{order.orderNumber}</h1>
            <Badge variant={getStatusVariant(order.status)} dot>
              {capitalizeFirst(order.status)}
            </Badge>
          </div>
          <p class="order-meta">{order.clientName} · {formatDate(order.createdAt)}</p>
        </div>
      </div>
    </div>

    <!-- Status Timeline -->
    {#if order.status !== 'cancelled'}
      <Card>
        <div class="timeline">
          {#each statusSteps as step, i}
            {@const state = getStepState(step)}
            <div class="timeline-step">
              <div class="step-indicator-wrap">
                <div class="step-circle {state}">
                  {#if state === 'complete'}
                    <svg viewBox="0 0 16 16" fill="currentColor" width="12" height="12">
                      <path d="M13.854 3.646a.5.5 0 0 1 0 .708l-7 7a.5.5 0 0 1-.708 0l-3.5-3.5a.5.5 0 1 1 .708-.708L6.5 10.293l6.646-6.647a.5.5 0 0 1 .708 0z"/>
                    </svg>
                  {:else if state === 'current'}
                    <div class="step-dot"></div>
                  {/if}
                </div>
                {#if i < statusSteps.length - 1}
                  <div class="step-line {state === 'complete' ? 'active' : ''}"></div>
                {/if}
              </div>
              <span class="step-label {state}">{capitalizeFirst(step)}</span>
            </div>
          {/each}
        </div>
      </Card>
    {:else}
      <Card>
        <div class="cancelled-indicator">
          <span class="cancelled-icon">✕</span>
          <span class="cancelled-text">This order has been cancelled</span>
        </div>
      </Card>
    {/if}

    <div class="content-grid">
      <div class="main-col">
        <!-- Line Items -->
        <Card>
          <h2 class="section-title">Order Items</h2>
          <table class="items-table">
            <thead>
              <tr>
                <th>Product</th>
                <th class="text-right">Qty</th>
                <th class="text-right">Unit Price</th>
                <th class="text-right">Total</th>
              </tr>
            </thead>
            <tbody>
              {#each order.items as item}
                <tr>
                  <td>{item.productName}</td>
                  <td class="text-right">{item.quantity}</td>
                  <td class="text-right">{formatCurrency(item.unitPrice)}</td>
                  <td class="text-right font-medium">{formatCurrency(item.total)}</td>
                </tr>
              {/each}
            </tbody>
            <tfoot>
              <tr class="subtotal-row">
                <td colspan="3" class="text-right label-cell">Subtotal</td>
                <td class="text-right">{formatCurrency(order.subtotal)}</td>
              </tr>
              <tr class="tax-row">
                <td colspan="3" class="text-right label-cell">Tax</td>
                <td class="text-right">{formatCurrency(order.tax)}</td>
              </tr>
              <tr class="total-row">
                <td colspan="3" class="text-right label-cell">Total</td>
                <td class="text-right total-value">{formatCurrency(order.total)}</td>
              </tr>
            </tfoot>
          </table>
        </Card>

        <!-- Shipping Address -->
        {#if order.shippingAddress}
          <Card>
            <h2 class="section-title">Shipping Address</h2>
            <address class="address-block">
              {#if order.shippingAddress.street}<div>{order.shippingAddress.street}</div>{/if}
              {#if order.shippingAddress.city || order.shippingAddress.state}
                <div>
                  {[order.shippingAddress.city, order.shippingAddress.state].filter(Boolean).join(', ')}
                  {order.shippingAddress.postalCode || ''}
                </div>
              {/if}
              {#if order.shippingAddress.country}<div>{order.shippingAddress.country}</div>{/if}
            </address>
          </Card>
        {/if}

        <!-- Notes -->
        {#if order.notes}
          <Card>
            <h2 class="section-title">Notes</h2>
            <p class="notes-text">{order.notes}</p>
          </Card>
        {/if}
      </div>

      <!-- Actions Panel -->
      <div class="actions-col">
        <Card>
          <h2 class="section-title">Order Actions</h2>
          <div class="actions-list">
            {#if order.status === 'pending'}
              <Button
                variant="primary"
                on:click={handleConfirmOrder}
                loading={actionLoading}
                disabled={actionLoading}
              >
                Confirm Order
              </Button>
            {/if}
            {#if order.status === 'confirmed'}
              <Button
                variant="primary"
                on:click={handleMarkProcessing}
                loading={actionLoading}
                disabled={actionLoading}
              >
                Mark Processing
              </Button>
            {/if}
            {#if order.status === 'processing'}
              <Button
                variant="primary"
                on:click={() => (shipModalOpen = true)}
                disabled={actionLoading}
              >
                Mark Shipped
              </Button>
            {/if}
            {#if order.status === 'shipped'}
              <Button
                variant="primary"
                on:click={handleMarkDelivered}
                loading={actionLoading}
                disabled={actionLoading}
              >
                Mark Delivered
              </Button>
            {/if}
            {#if order.status !== 'delivered' && order.status !== 'cancelled'}
              <Button
                variant="danger"
                on:click={() => (cancelModalOpen = true)}
                disabled={actionLoading}
              >
                Cancel Order
              </Button>
            {/if}
          </div>
        </Card>

        <!-- Order Summary Card -->
        <Card>
          <h2 class="section-title">Summary</h2>
          <div class="summary-rows">
            <div class="summary-row">
              <span class="summary-label">Order #</span>
              <span class="summary-value mono">{order.orderNumber}</span>
            </div>
            <div class="summary-row">
              <span class="summary-label">Client</span>
              <span class="summary-value">{order.clientName}</span>
            </div>
            <div class="summary-row">
              <span class="summary-label">Total</span>
              <span class="summary-value font-bold">{formatCurrency(order.total)}</span>
            </div>
            <div class="summary-row">
              <span class="summary-label">Created</span>
              <span class="summary-value">{formatDate(order.createdAt)}</span>
            </div>
          </div>
        </Card>
      </div>
    </div>
  {/if}
</div>

<!-- Ship Modal -->
<Modal bind:open={shipModalOpen} title="Mark as Shipped" size="sm">
  <div class="modal-body">
    <p class="modal-text">Enter a tracking number for this shipment (optional).</p>
    <Input
      id="tracking"
      label="Tracking Number"
      type="text"
      placeholder="e.g. 1Z999AA1012345678"
      bind:value={trackingNumber}
    />
  </div>
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={shipping}>Cancel</Button>
    <Button variant="primary" on:click={handleShipSubmit} loading={shipping}>
      Mark Shipped
    </Button>
  </svelte:fragment>
</Modal>

<!-- Cancel Modal -->
<ConfirmModal
  bind:open={cancelModalOpen}
  title="Cancel Order"
  message="Are you sure you want to cancel this order? This action cannot be undone."
  confirmLabel="Yes, Cancel Order"
  cancelLabel="Keep Order"
  variant="danger"
  loading={cancelling}
  on:confirm={handleCancelConfirm}
  on:cancel={() => (cancelModalOpen = false)}
/>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1100px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .spinner-wrap {
    display: flex;
    justify-content: center;
    padding: 4rem 0;
  }

  .error-wrap {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-width: 500px;
  }

  .page-header {
    display: flex;
    align-items: flex-start;
  }

  .header-left {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .header-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .header-top {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .order-number {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-gray-900, #111827);
    margin: 0;
  }

  .order-meta {
    font-size: 0.875rem;
    color: var(--color-gray-500, #6b7280);
    margin: 0;
  }

  /* Timeline */
  .timeline {
    display: flex;
    align-items: flex-start;
    gap: 0;
    overflow-x: auto;
    padding: 0.5rem 0;
  }

  .timeline-step {
    display: flex;
    flex-direction: column;
    align-items: center;
    flex: 1;
    min-width: 80px;
  }

  .step-indicator-wrap {
    display: flex;
    align-items: center;
    width: 100%;
    position: relative;
  }

  .step-circle {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    border: 2px solid;
    transition: all 0.2s;
    z-index: 1;
  }

  .step-circle.complete {
    background: #059669;
    border-color: #059669;
    color: white;
  }

  .step-circle.current {
    background: white;
    border-color: #2563eb;
    color: #2563eb;
  }

  .step-circle.future {
    background: white;
    border-color: var(--color-gray-200, #e5e7eb);
    color: var(--color-gray-300, #d1d5db);
  }

  .step-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #2563eb;
  }

  .step-line {
    flex: 1;
    height: 2px;
    background: var(--color-gray-200, #e5e7eb);
  }

  .step-line.active {
    background: #059669;
  }

  .step-label {
    font-size: 0.75rem;
    font-weight: 500;
    margin-top: 0.5rem;
    text-align: center;
  }

  .step-label.complete { color: #059669; }
  .step-label.current { color: #2563eb; font-weight: 700; }
  .step-label.future { color: var(--color-gray-400, #9ca3af); }

  .cancelled-indicator {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    padding: 0.75rem;
  }

  .cancelled-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: #fee2e2;
    color: #dc2626;
    font-size: 0.875rem;
    font-weight: 700;
  }

  .cancelled-text {
    font-size: 0.875rem;
    font-weight: 500;
    color: #dc2626;
  }

  /* Content grid */
  .content-grid {
    display: grid;
    grid-template-columns: 1fr 280px;
    gap: 1.5rem;
    align-items: start;
  }

  @media (max-width: 768px) {
    .content-grid {
      grid-template-columns: 1fr;
    }
  }

  .main-col {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .actions-col {
    position: sticky;
    top: 1rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .section-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900, #111827);
    margin: 0 0 1rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--color-gray-100, #f3f4f6);
  }

  /* Items table */
  .items-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .items-table thead tr th {
    padding: 0.5rem 0.75rem;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--color-gray-500, #6b7280);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid var(--color-gray-100, #f3f4f6);
  }

  .items-table tbody tr td {
    padding: 0.75rem;
    color: var(--color-gray-700, #374151);
    border-bottom: 1px solid var(--color-gray-50, #f9fafb);
    vertical-align: middle;
  }

  .items-table tfoot tr td {
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
    color: var(--color-gray-600, #4b5563);
  }

  .subtotal-row td { padding-top: 0.75rem; border-top: 1px solid var(--color-gray-100, #f3f4f6); }
  .tax-row td { color: var(--color-gray-500, #6b7280); }

  .total-row td {
    padding-top: 0.5rem;
    border-top: 2px solid var(--color-gray-200, #e5e7eb);
    font-weight: 700;
  }

  .total-value {
    font-size: 1rem;
    color: var(--color-gray-900, #111827);
  }

  .text-right { text-align: right; }
  .label-cell { color: var(--color-gray-500, #6b7280); }
  .font-medium { font-weight: 600; }

  .address-block {
    font-style: normal;
    font-size: 0.875rem;
    color: var(--color-gray-700, #374151);
    line-height: 1.75;
  }

  .notes-text {
    font-size: 0.875rem;
    color: var(--color-gray-700, #374151);
    white-space: pre-wrap;
    margin: 0;
  }

  /* Actions */
  .actions-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  /* Summary */
  .summary-rows {
    display: flex;
    flex-direction: column;
    gap: 0.625rem;
  }

  .summary-row {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: 0.5rem;
  }

  .summary-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-gray-500, #6b7280);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .summary-value {
    font-size: 0.875rem;
    color: var(--color-gray-800, #1f2937);
    text-align: right;
  }

  .summary-value.mono {
    font-family: monospace;
    font-size: 0.8125rem;
  }

  .summary-value.font-bold {
    font-weight: 700;
  }

  /* Modal */
  .modal-body {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .modal-text {
    font-size: 0.875rem;
    color: var(--color-gray-600, #4b5563);
    margin: 0;
  }
</style>
