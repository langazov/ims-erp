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
  import Textarea from '$lib/shared/components/forms/Textarea.svelte';

  interface Payment {
    id: string;
    paymentNumber: string;
    invoiceId: string;
    invoiceNumber: string;
    clientId: string;
    clientName: string;
    amount: string;
    method: 'credit_card' | 'bank_transfer' | 'cash' | 'check' | 'paypal' | 'stripe';
    status: 'pending' | 'completed' | 'failed' | 'refunded';
    transactionId?: string;
    reference?: string;
    paidAt: string;
    notes?: string;
    createdAt: string;
    updatedAt: string;
  }

  const paymentId = $page.params.id;

  let payment: Payment | null = null;
  let loading = true;
  let error: string | null = null;
  let showRefundModal = false;
  let showPrintModal = false;
  let processing = false;
  let actionError: string | null = null;
  let refundReason = '';

  // Mock related payments - in real implementation, this would come from API
  let relatedPayments: Payment[] = [];

  async function loadPayment() {
    loading = true;
    error = null;
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // Mock payment data
      payment = {
        id: paymentId,
        paymentNumber: 'PAY-2024-001',
        invoiceId: 'inv-123',
        invoiceNumber: 'INV-2024-001',
        clientId: 'client-001',
        clientName: 'Acme Corporation',
        amount: '5400.00',
        method: 'bank_transfer',
        status: 'completed',
        transactionId: 'TXN-789456123',
        reference: 'Wire Transfer - Jan 2024',
        paidAt: '2024-01-20T10:30:00Z',
        notes: 'Payment received via wire transfer. Customer confirmed payment on 2024-01-18.',
        createdAt: '2024-01-20T10:30:00Z',
        updatedAt: '2024-01-20T10:30:00Z'
      };

      // Mock related payments from same client
      relatedPayments = [
        {
          id: 'pay-002',
          paymentNumber: 'PAY-2024-002',
          invoiceId: 'inv-124',
          invoiceNumber: 'INV-2024-005',
          clientId: 'client-001',
          clientName: 'Acme Corporation',
          amount: '3200.00',
          method: 'credit_card',
          status: 'completed',
          transactionId: 'TXN-789456124',
          paidAt: '2024-02-15T14:20:00Z',
          createdAt: '2024-02-15T14:20:00Z',
          updatedAt: '2024-02-15T14:20:00Z'
        },
        {
          id: 'pay-003',
          paymentNumber: 'PAY-2024-003',
          invoiceId: 'inv-125',
          invoiceNumber: 'INV-2024-008',
          clientId: 'client-001',
          clientName: 'Acme Corporation',
          amount: '1500.00',
          method: 'check',
          status: 'pending',
          reference: 'Check #4521',
          paidAt: '2024-03-01T09:00:00Z',
          createdAt: '2024-03-01T09:00:00Z',
          updatedAt: '2024-03-01T09:00:00Z'
        }
      ];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load payment';
    } finally {
      loading = false;
    }
  }

  function getStatusVariant(status: Payment['status']): 'green' | 'gray' | 'yellow' | 'red' | 'blue' {
    switch (status) {
      case 'completed': return 'green';
      case 'pending': return 'yellow';
      case 'failed': return 'red';
      case 'refunded': return 'blue';
      default: return 'gray';
    }
  }

  function getMethodIcon(method: Payment['method']): string {
    const icons: Record<string, string> = {
      credit_card: '💳',
      bank_transfer: '🏦',
      cash: '💵',
      check: '📝',
      paypal: '💰',
      stripe: '💳'
    };
    return icons[method] || '💳';
  }

  function getMethodLabel(method: Payment['method']): string {
    const labels: Record<string, string> = {
      credit_card: 'Credit Card',
      bank_transfer: 'Bank Transfer',
      cash: 'Cash',
      check: 'Check',
      paypal: 'PayPal',
      stripe: 'Stripe'
    };
    return labels[method] || method;
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

  function handleRefund() {
    showRefundModal = true;
  }

  async function confirmRefund() {
    if (!refundReason.trim()) {
      actionError = 'Please provide a reason for the refund';
      return;
    }

    processing = true;
    actionError = null;
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      if (payment) {
        payment = { ...payment, status: 'refunded' };
      }
      
      showRefundModal = false;
      refundReason = '';
    } catch (err) {
      actionError = err instanceof Error ? err.message : 'Failed to process refund';
    } finally {
      processing = false;
    }
  }

  function handleDownloadReceipt() {
    // TODO: Implement receipt download
    alert('Receipt download functionality will be implemented here');
  }

  function handlePrint() {
    showPrintModal = true;
  }

  function confirmPrint() {
    window.print();
    showPrintModal = false;
  }

  function navigateToInvoice() {
    if (payment) {
      goto(`/invoices/${payment.invoiceId}`);
    }
  }

  function navigateToClient() {
    if (payment) {
      goto(`/clients/${payment.clientId}`);
    }
  }

  function navigateToPayment(id: string) {
    goto(`/payments/${id}`);
  }

  function canRefund(status: Payment['status']): boolean {
    return status === 'completed';
  }

  // Mock activity timeline - in real implementation, this would come from API
  const activities = [
    { type: 'created', date: '2024-01-20T10:30:00Z', description: 'Payment recorded', user: 'John Doe' },
    { type: 'processed', date: '2024-01-20T10:35:00Z', description: 'Payment processed successfully', user: 'System' },
    { type: 'notified', date: '2024-01-20T10:36:00Z', description: 'Invoice marked as paid', user: 'System' }
  ];

  onMount(() => {
    loadPayment();
  });
</script>

<svelte:head>
  <title>{payment ? `Payment ${payment.paymentNumber}` : 'Payment Details'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading payment details...</p>
    </div>
  {:else if error}
    <div class="error-container">
      <Alert variant="error">{error}</Alert>
      <Button variant="secondary" on:click={() => goto('/payments')}>
        Back to Payments
      </Button>
    </div>
  {:else if payment}
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <h1 class="page-title">Payment {payment.paymentNumber}</h1>
          <Badge variant={getStatusVariant(payment.status)} size="md">
            {payment.status}
          </Badge>
        </div>
        <p class="page-description">
          Processed on {formatDate(payment.paidAt)}
        </p>
      </div>
      <div class="header-actions">
        {#if canRefund(payment.status)}
          <Button variant="secondary" on:click={handleRefund}>
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
            </svg>
            Refund
          </Button>
        {/if}
        <Button variant="secondary" on:click={handleDownloadReceipt}>
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          Download Receipt
        </Button>
        <Button variant="secondary" on:click={handlePrint}>
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
          </svg>
          Print
        </Button>
      </div>
    </div>

    <div class="content-grid">
      <!-- Main Content -->
      <div class="main-content">
        <!-- Payment Amount Card -->
        <Card>
          <div class="amount-display">
            <span class="amount-label">Payment Amount</span>
            <span class="amount-value">{formatCurrency(payment.amount)}</span>
            <div class="amount-method">
              <span class="method-icon">{getMethodIcon(payment.method)}</span>
              <span class="method-label">{getMethodLabel(payment.method)}</span>
            </div>
          </div>
        </Card>

        <!-- Invoice Information -->
        <Card>
          <h2 class="card-title">Invoice Information</h2>
          <div class="info-section">
            <div class="info-row">
              <span class="info-label">Invoice Number</span>
              <button class="info-value link" on:click={navigateToInvoice}>
                {payment.invoiceNumber}
              </button>
            </div>
            <div class="info-row">
              <span class="info-label">Invoice ID</span>
              <span class="info-value">{payment.invoiceId}</span>
            </div>
          </div>
        </Card>

        <!-- Client Information -->
        <Card>
          <h2 class="card-title">Client Information</h2>
          <div class="info-section">
            <div class="info-row">
              <span class="info-label">Client Name</span>
              <button class="info-value link" on:click={navigateToClient}>
                {payment.clientName}
              </button>
            </div>
            <div class="info-row">
              <span class="info-label">Client ID</span>
              <span class="info-value">{payment.clientId}</span>
            </div>
          </div>
        </Card>

        <!-- Payment Details -->
        <Card>
          <h2 class="card-title">Payment Details</h2>
          <div class="info-section">
            <div class="info-row">
              <span class="info-label">Payment Method</span>
              <span class="info-value">
                <span class="method-badge">
                  <span class="method-icon-small">{getMethodIcon(payment.method)}</span>
                  {getMethodLabel(payment.method)}
                </span>
              </span>
            </div>
            {#if payment.transactionId}
              <div class="info-row">
                <span class="info-label">Transaction ID</span>
                <span class="info-value mono">{payment.transactionId}</span>
              </div>
            {/if}
            {#if payment.reference}
              <div class="info-row">
                <span class="info-label">Reference</span>
                <span class="info-value">{payment.reference}</span>
              </div>
            {/if}
            <div class="info-row">
              <span class="info-label">Payment Date</span>
              <span class="info-value">{formatDateTime(payment.paidAt)}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Created</span>
              <span class="info-value">{formatDateTime(payment.createdAt)}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Last Updated</span>
              <span class="info-value">{formatDateTime(payment.updatedAt)}</span>
            </div>
          </div>
          {#if payment.notes}
            <div class="notes-section">
              <h3 class="notes-title">Notes</h3>
              <p class="notes-content">{payment.notes}</p>
            </div>
          {/if}
        </Card>

        <!-- Related Payments -->
        {#if relatedPayments.length > 0}
          <Card>
            <h2 class="card-title">Other Payments from {payment.clientName}</h2>
            <div class="related-payments">
              {#each relatedPayments as relatedPayment}
                <button 
                  class="related-payment-item"
                  on:click={() => navigateToPayment(relatedPayment.id)}
                >
                  <div class="related-payment-info">
                    <span class="related-payment-number">{relatedPayment.paymentNumber}</span>
                    <span class="related-payment-invoice">{relatedPayment.invoiceNumber}</span>
                  </div>
                  <div class="related-payment-details">
                    <span class="related-payment-amount">{formatCurrency(relatedPayment.amount)}</span>
                    <Badge variant={getStatusVariant(relatedPayment.status)} size="sm">
                      {relatedPayment.status}
                    </Badge>
                  </div>
                </button>
              {/each}
            </div>
          </Card>
        {/if}
      </div>

      <!-- Sidebar -->
      <div class="sidebar">
        <!-- Status Card -->
        <Card>
          <h3 class="sidebar-title">Status</h3>
          <div class="status-display">
            <Badge variant={getStatusVariant(payment.status)} size="md">
              {payment.status}
            </Badge>
          </div>
          <div class="status-info">
            <div class="info-row-small">
              <span class="label">Payment Date</span>
              <span class="value">{formatDateTime(payment.paidAt)}</span>
            </div>
            <div class="info-row-small">
              <span class="label">Created</span>
              <span class="value">{formatDateTime(payment.createdAt)}</span>
            </div>
            <div class="info-row-small">
              <span class="label">Last Updated</span>
              <span class="value">{formatDateTime(payment.updatedAt)}</span>
            </div>
          </div>
        </Card>

        <!-- Amount Summary -->
        <Card>
          <h3 class="sidebar-title">Amount</h3>
          <div class="amount-summary">
            <span class="summary-amount">{formatCurrency(payment.amount)}</span>
            <span class="summary-method">{getMethodLabel(payment.method)}</span>
          </div>
        </Card>

        <!-- Activity Timeline -->
        <Card>
          <h3 class="sidebar-title">Activity</h3>
          <div class="timeline">
            {#each activities as activity}
              <div class="timeline-item">
                <div class="timeline-dot" class:created={activity.type === 'created'} class:processed={activity.type === 'processed'}></div>
                <div class="timeline-content">
                  <p class="timeline-description">{activity.description}</p>
                  <p class="timeline-user">by {activity.user}</p>
                  <p class="timeline-date">{formatDateTime(activity.date)}</p>
                </div>
              </div>
            {/each}
          </div>
        </Card>
      </div>
    </div>
  {:else}
    <Alert variant="error">Payment not found</Alert>
  {/if}
</div>

<!-- Refund Modal -->
<Modal
  bind:open={showRefundModal}
  title="Process Refund"
  size="md"
>
  <div class="refund-modal-content">
    <p class="refund-info">
      You are about to refund payment <strong>{payment?.paymentNumber}</strong> for 
      <strong>{payment ? formatCurrency(payment.amount) : ''}</strong>.
    </p>
    
    <div class="refund-form">
      <Textarea
        id="refund-reason"
        label="Refund Reason"
        placeholder="Enter the reason for this refund..."
        bind:value={refundReason}
        required
        rows={4}
      />
    </div>
    
    {#if actionError}
      <Alert variant="error" class="mt-4">{actionError}</Alert>
    {/if}
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showRefundModal = false; refundReason = ''; actionError = null; }} disabled={processing}>
      Cancel
    </Button>
    <Button variant="danger" on:click={confirmRefund} loading={processing}>
      {processing ? 'Processing...' : 'Process Refund'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Print Modal -->
<Modal
  bind:open={showPrintModal}
  title="Print Payment Receipt"
  size="sm"
>
  <p>Print receipt for payment <strong>{payment?.paymentNumber}</strong>?</p>
  <p class="text-sm text-gray-500 mt-2">This will open your browser's print dialog.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showPrintModal = false; }}>
      Cancel
    </Button>
    <Button variant="primary" on:click={confirmPrint}>
      Print Receipt
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

  .card-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
  }

  .amount-display {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2rem;
    text-align: center;
  }

  .amount-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 0.5rem;
  }

  .amount-value {
    font-size: 3rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin-bottom: 1rem;
  }

  .amount-method {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    background-color: var(--color-gray-100);
    border-radius: 9999px;
  }

  .method-icon {
    font-size: 1.25rem;
  }

  .method-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-700);
  }

  .info-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 0;
    border-bottom: 1px solid var(--color-gray-100);
  }

  .info-row:last-child {
    border-bottom: none;
  }

  .info-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .info-value {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-900);
  }

  .info-value.link {
    color: var(--color-primary-600);
    background: none;
    border: none;
    cursor: pointer;
    text-decoration: underline;
    padding: 0;
  }

  .info-value.link:hover {
    color: var(--color-primary-700);
  }

  .info-value.mono {
    font-family: monospace;
    font-size: 0.8125rem;
    background-color: var(--color-gray-100);
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
  }

  .method-badge {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.375rem 0.75rem;
    background-color: var(--color-gray-100);
    border-radius: 0.375rem;
    font-size: 0.875rem;
  }

  .method-icon-small {
    font-size: 1rem;
  }

  .notes-section {
    margin-top: 1.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--color-gray-200);
  }

  .notes-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-700);
    margin: 0 0 0.5rem 0;
  }

  .notes-content {
    font-size: 0.875rem;
    color: var(--color-gray-600);
    line-height: 1.6;
    margin: 0;
    white-space: pre-wrap;
  }

  .related-payments {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .related-payment-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
    border: 1px solid var(--color-gray-200);
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: left;
    width: 100%;
  }

  .related-payment-item:hover {
    background-color: var(--color-gray-100);
    border-color: var(--color-gray-300);
  }

  .related-payment-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .related-payment-number {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .related-payment-invoice {
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  .related-payment-details {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.25rem;
  }

  .related-payment-amount {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .status-display {
    margin-bottom: 1rem;
  }

  .status-info {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .info-row-small {
    display: flex;
    justify-content: space-between;
    font-size: 0.875rem;
  }

  .info-row-small .label {
    color: var(--color-gray-500);
  }

  .info-row-small .value {
    color: var(--color-gray-900);
    font-weight: 500;
  }

  .amount-summary {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    padding: 1rem;
  }

  .summary-amount {
    font-size: 2rem;
    font-weight: 700;
    color: var(--color-gray-900);
  }

  .summary-method {
    font-size: 0.875rem;
    color: var(--color-gray-500);
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

  .timeline-dot.processed {
    background-color: var(--color-green-500);
  }

  .timeline-content {
    flex: 1;
  }

  .timeline-description {
    font-size: 0.875rem;
    color: var(--color-gray-900);
    margin: 0 0 0.125rem 0;
  }

  .timeline-user {
    font-size: 0.75rem;
    color: var(--color-gray-500);
    margin: 0 0 0.125rem 0;
  }

  .timeline-date {
    font-size: 0.75rem;
    color: var(--color-gray-400);
    margin: 0;
  }

  .refund-modal-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .refund-info {
    color: var(--color-gray-700);
    margin: 0;
  }

  .refund-form {
    margin-top: 0.5rem;
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

    .amount-value {
      font-size: 2.5rem;
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

    .amount-value {
      font-size: 2rem;
    }

    .info-row {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.25rem;
    }
  }

  @media print {
    .page-header .header-actions,
    .sidebar {
      display: none;
    }

    .content-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
