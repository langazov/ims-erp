<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Textarea from '$lib/shared/components/forms/Textarea.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import ConfirmModal from '$lib/shared/components/layout/ConfirmModal.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import {
    getPaymentById,
    refundPayment,
    cancelPayment,
    type Payment,
    type PaymentStatus,
    type PaymentMethod
  } from '$lib/shared/api/payments';
  import { toast } from '$lib/shared/stores/toast';

  let payment: Payment | null = null;
  let loading = true;
  let error: string | null = null;

  let refundModalOpen = false;
  let refundAmount = '';
  let refundReason = '';
  let refundLoading = false;
  let refundError: string | null = null;

  let cancelModalOpen = false;
  let cancelling = false;

  $: paymentId = $page.params.id;

  function formatCurrency(s: string): string {
    const num = parseFloat(s);
    if (isNaN(num)) return s;
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(num);
  }

  function formatDate(s: string | null): string {
    if (!s) return '—';
    return new Date(s).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function getStatusVariant(status: PaymentStatus): 'green' | 'yellow' | 'red' | 'gray' | 'blue' {
    switch (status) {
      case 'completed': return 'green';
      case 'pending': return 'yellow';
      case 'processing': return 'blue';
      case 'failed': return 'red';
      case 'refunded': return 'gray';
      case 'cancelled': return 'gray';
      default: return 'gray';
    }
  }

  function formatMethodLabel(method: PaymentMethod): string {
    const labels: Record<PaymentMethod, string> = {
      card: 'Card',
      bank_transfer: 'Bank Transfer',
      cash: 'Cash',
      check: 'Check',
      stripe: 'Stripe',
      paypal: 'PayPal'
    };
    return labels[method] ?? method;
  }

  async function loadPayment() {
    loading = true;
    error = null;
    try {
      payment = await getPaymentById(paymentId);
      refundAmount = payment.amount;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load payment';
    } finally {
      loading = false;
    }
  }

  function openRefundModal() {
    if (payment) refundAmount = payment.amount;
    refundReason = '';
    refundError = null;
    refundModalOpen = true;
  }

  async function handleRefund() {
    if (!payment) return;
    const amt = parseFloat(refundAmount);
    if (isNaN(amt) || amt <= 0) {
      refundError = 'Please enter a valid refund amount';
      return;
    }
    if (!refundReason.trim()) {
      refundError = 'Please provide a reason for the refund';
      return;
    }
    refundLoading = true;
    refundError = null;
    try {
      await refundPayment(payment.id, { amount: amt, reason: refundReason.trim() });
      toast.success('Payment refunded successfully');
      refundModalOpen = false;
      await loadPayment();
    } catch (e) {
      refundError = e instanceof Error ? e.message : 'Failed to process refund';
    } finally {
      refundLoading = false;
    }
  }

  async function handleCancel() {
    if (!payment) return;
    cancelling = true;
    try {
      await cancelPayment(payment.id);
      toast.success('Payment cancelled successfully');
      cancelModalOpen = false;
      await loadPayment();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to cancel payment');
    } finally {
      cancelling = false;
    }
  }

  onMount(loadPayment);
</script>

<div class="page-container">
  {#if loading}
    <div class="spinner-wrap">
      <Spinner size="lg" />
    </div>
  {:else if error}
    <div class="error-wrap">
      <Alert variant="error">{error}</Alert>
      <Button variant="secondary" on:click={() => goto('/payments')}>← Back to Payments</Button>
    </div>
  {:else if payment}
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <Button variant="ghost" size="sm" on:click={() => goto('/payments')}>← Back</Button>
        <div class="header-info">
          <div class="header-top">
            <h1 class="payment-number">{payment.paymentNumber}</h1>
            <Badge variant={getStatusVariant(payment.status)} dot>
              {payment.status.charAt(0).toUpperCase() + payment.status.slice(1)}
            </Badge>
          </div>
          <p class="client-name">{payment.clientName}</p>
        </div>
      </div>
      <div class="amount-display">{formatCurrency(payment.amount)}</div>
    </div>

    <div class="content-grid">
      <div class="main-col">
        <!-- Payment Details -->
        <Card>
          <h2 class="section-title">Payment Details</h2>
          <div class="details-grid">
            <div class="detail-item">
              <span class="detail-label">Payment Method</span>
              <span class="detail-value">{formatMethodLabel(payment.method)}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Currency</span>
              <span class="detail-value">{payment.currency}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Gateway Reference</span>
              <span class="detail-value">{payment.gatewayReference ?? 'N/A'}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Paid At</span>
              <span class="detail-value">{formatDate(payment.paidAt)}</span>
            </div>
            {#if payment.notes}
              <div class="detail-item full-width">
                <span class="detail-label">Notes</span>
                <span class="detail-value">{payment.notes}</span>
              </div>
            {/if}
          </div>
        </Card>

        <!-- Invoice -->
        <Card>
          <h2 class="section-title">Related Invoice</h2>
          <div class="invoice-link-row">
            <span class="detail-label">Invoice</span>
            <a href="/invoices/{payment.invoiceId}" class="invoice-link">
              {payment.invoiceNumber} →
            </a>
          </div>
        </Card>

        <!-- Timestamps -->
        <Card>
          <h2 class="section-title">Timestamps</h2>
          <div class="details-grid">
            <div class="detail-item">
              <span class="detail-label">Created</span>
              <span class="detail-value">{formatDate(payment.createdAt)}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Last Updated</span>
              <span class="detail-value">{formatDate(payment.updatedAt)}</span>
            </div>
          </div>
        </Card>
      </div>

      <!-- Actions Panel -->
      <div class="actions-col">
        <Card>
          <h2 class="section-title">Actions</h2>
          <div class="actions-list">
            {#if payment.status === 'completed'}
              <Button variant="secondary" on:click={openRefundModal}>
                Refund Payment
              </Button>
            {/if}
            {#if payment.status === 'pending'}
              <Button variant="danger" on:click={() => (cancelModalOpen = true)}>
                Cancel Payment
              </Button>
            {/if}
            <Button variant="ghost" on:click={() => goto('/payments')}>
              Back to Payments
            </Button>
          </div>
        </Card>
      </div>
    </div>
  {/if}
</div>

<!-- Refund Modal -->
<Modal bind:open={refundModalOpen} title="Refund Payment" size="md">
  <div class="modal-body">
    {#if refundError}
      <Alert variant="error" dismissible on:dismiss={() => (refundError = null)}>{refundError}</Alert>
    {/if}
    <Input
      id="refund-amount"
      label="Refund Amount"
      type="number"
      placeholder="0.00"
      bind:value={refundAmount}
    />
    <Textarea
      id="refund-reason"
      label="Reason for Refund"
      placeholder="Enter reason for refund..."
      bind:value={refundReason}
      rows={3}
    />
  </div>
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={refundLoading}>Cancel</Button>
    <Button variant="primary" on:click={handleRefund} loading={refundLoading}>
      Process Refund
    </Button>
  </svelte:fragment>
</Modal>

<!-- Cancel Modal -->
<ConfirmModal
  bind:open={cancelModalOpen}
  title="Cancel Payment"
  message="Are you sure you want to cancel this payment? This action cannot be undone."
  confirmLabel="Yes, Cancel Payment"
  cancelLabel="Keep Payment"
  variant="danger"
  loading={cancelling}
  on:confirm={handleCancel}
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
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .header-left {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
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

  .payment-number {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-gray-900, #111827);
    margin: 0;
  }

  .client-name {
    font-size: 0.875rem;
    color: var(--color-gray-500, #6b7280);
    margin: 0;
  }

  .amount-display {
    font-size: 2rem;
    font-weight: 800;
    color: var(--color-gray-900, #111827);
  }

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
  }

  .section-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900, #111827);
    margin: 0 0 1rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--color-gray-100, #f3f4f6);
  }

  .details-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .detail-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .detail-item.full-width {
    grid-column: 1 / -1;
  }

  .detail-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-gray-500, #6b7280);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .detail-value {
    font-size: 0.875rem;
    color: var(--color-gray-900, #111827);
    font-weight: 500;
  }

  .invoice-link-row {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .invoice-link {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-primary-600, #2563eb);
    text-decoration: none;
  }

  .invoice-link:hover {
    text-decoration: underline;
  }

  .actions-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .modal-body {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
</style>
