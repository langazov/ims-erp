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
    getInvoiceById, 
    deleteInvoice, 
    markInvoiceAsPaid,
    updateInvoice,
    type Invoice,
    type InvoiceStatus 
  } from '$lib/shared/api/invoices';

  const invoiceId = $page.params.id;

  let invoice: Invoice | null = null;
  let loading = true;
  let error: string | null = null;
  let showDeleteModal = false;
  let showMarkPaidModal = false;
  let showSendModal = false;
  let processing = false;
  let actionError: string | null = null;

  async function loadInvoice() {
    loading = true;
    error = null;
    
    try {
      invoice = await getInvoiceById(invoiceId);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load invoice';
    } finally {
      loading = false;
    }
  }

  function getStatusVariant(status: InvoiceStatus): 'green' | 'gray' | 'yellow' | 'red' | 'blue' {
    switch (status) {
      case 'paid': return 'green';
      case 'sent': return 'blue';
      case 'draft': return 'gray';
      case 'overdue': return 'red';
      case 'cancelled': return 'yellow';
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

  function isOverdue(invoice: Invoice): boolean {
    if (invoice.status === 'paid' || invoice.status === 'cancelled') return false;
    return new Date(invoice.dueDate) < new Date();
  }

  function handleEdit() {
    goto(`/invoices/${invoiceId}/edit`);
  }

  function handleDelete() {
    showDeleteModal = true;
  }

  async function confirmDelete() {
    processing = true;
    actionError = null;
    try {
      await deleteInvoice(invoiceId);
      goto('/invoices');
    } catch (err) {
      actionError = err instanceof Error ? err.message : 'Failed to delete invoice';
      processing = false;
    }
  }

  function handleSend() {
    showSendModal = true;
  }

  async function confirmSend() {
    processing = true;
    actionError = null;
    try {
      await updateInvoice(invoiceId, { status: 'sent' });
      await loadInvoice();
      showSendModal = false;
    } catch (err) {
      actionError = err instanceof Error ? err.message : 'Failed to send invoice';
    } finally {
      processing = false;
    }
  }

  function handleMarkAsPaid() {
    showMarkPaidModal = true;
  }

  async function confirmMarkAsPaid() {
    processing = true;
    actionError = null;
    try {
      await markInvoiceAsPaid(invoiceId);
      await loadInvoice();
      showMarkPaidModal = false;
    } catch (err) {
      actionError = err instanceof Error ? err.message : 'Failed to mark invoice as paid';
    } finally {
      processing = false;
    }
  }

  function handleDownloadPDF() {
    // TODO: Implement PDF download
    alert('PDF download functionality will be implemented here');
  }

  function canEdit(status: InvoiceStatus): boolean {
    return status === 'draft' || status === 'sent';
  }

  function canSend(status: InvoiceStatus): boolean {
    return status === 'draft';
  }

  function canMarkAsPaid(status: InvoiceStatus): boolean {
    return status === 'sent' || status === 'overdue';
  }

  function canDelete(status: InvoiceStatus): boolean {
    return status !== 'paid';
  }

  // Mock activity timeline - in real implementation, this would come from API
  const activities = [
    { type: 'created', date: '2024-01-15T10:30:00Z', description: 'Invoice created' },
    { type: 'updated', date: '2024-01-15T11:00:00Z', description: 'Line items updated' },
    { type: 'sent', date: '2024-01-15T14:00:00Z', description: 'Invoice sent to client' }
  ];

  onMount(() => {
    loadInvoice();
  });
</script>

<svelte:head>
  <title>{invoice ? `Invoice ${invoice.invoiceNumber}` : 'Invoice Details'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading invoice details...</p>
    </div>
  {:else if error}
    <div class="error-container">
      <Alert variant="error">{error}</Alert>
      <Button variant="secondary" on:click={() => goto('/invoices')}>
        Back to Invoices
      </Button>
    </div>
  {:else if invoice}
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <h1 class="page-title">Invoice {invoice.invoiceNumber}</h1>
          <Badge variant={getStatusVariant(invoice.status)} size="md">
            {invoice.status}
          </Badge>
          {#if isOverdue(invoice)}
            <Badge variant="red" size="md">Overdue</Badge>
          {/if}
        </div>
        <p class="page-description">
          Created on {formatDate(invoice.createdAt)}
        </p>
      </div>
      <div class="header-actions">
        {#if canEdit(invoice.status)}
          <Button variant="secondary" on:click={handleEdit}>
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            Edit
          </Button>
        {/if}
        {#if canSend(invoice.status)}
          <Button variant="primary" on:click={handleSend}>
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
            Send
          </Button>
        {/if}
        {#if canMarkAsPaid(invoice.status)}
          <Button variant="primary" on:click={handleMarkAsPaid}>
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Mark as Paid
          </Button>
        {/if}
        <Button variant="secondary" on:click={handleDownloadPDF}>
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          PDF
        </Button>
        {#if canDelete(invoice.status)}
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
      <!-- Invoice Summary -->
      <div class="main-content">
        <Card>
          <div class="invoice-header">
            <div class="invoice-from">
              <h3 class="company-name">Your Company</h3>
              <p class="company-details">
                123 Business Street<br />
                City, State 12345<br />
                contact@company.com
              </p>
            </div>
            <div class="invoice-meta">
              <div class="meta-row">
                <span class="meta-label">Invoice Number:</span>
                <span class="meta-value">{invoice.invoiceNumber}</span>
              </div>
              <div class="meta-row">
                <span class="meta-label">Issue Date:</span>
                <span class="meta-value">{formatDate(invoice.createdAt)}</span>
              </div>
              <div class="meta-row">
                <span class="meta-label">Due Date:</span>
                <span class="meta-value" class:overdue={isOverdue(invoice)}>
                  {formatDate(invoice.dueDate)}
                  {#if isOverdue(invoice)}
                    <span class="overdue-badge">(Overdue)</span>
                  {/if}
                </span>
              </div>
              {#if invoice.paidDate}
                <div class="meta-row">
                  <span class="meta-label">Paid Date:</span>
                  <span class="meta-value paid">{formatDate(invoice.paidDate)}</span>
                </div>
              {/if}
            </div>
          </div>

          <div class="invoice-to">
            <h4 class="section-label">Bill To:</h4>
            <p class="client-name">{invoice.clientName}</p>
            <p class="client-id">Client ID: {invoice.clientId}</p>
          </div>

          <div class="line-items">
            <table class="items-table">
              <thead>
                <tr>
                  <th class="description">Description</th>
                  <th class="quantity">Quantity</th>
                  <th class="price">Unit Price</th>
                  <th class="total">Total</th>
                </tr>
              </thead>
              <tbody>
                {#each invoice.items as item}
                  <tr>
                    <td class="description">{item.description}</td>
                    <td class="quantity">{item.quantity}</td>
                    <td class="price">{formatCurrency(item.unitPrice)}</td>
                    <td class="total">{formatCurrency(item.total)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>

          <div class="invoice-totals">
            <div class="total-row">
              <span class="label">Subtotal</span>
              <span class="value">{formatCurrency(invoice.subtotal)}</span>
            </div>
            <div class="total-row">
              <span class="label">Tax</span>
              <span class="value">{formatCurrency(invoice.tax)}</span>
            </div>
            <div class="total-row grand-total">
              <span class="label">Total</span>
              <span class="value">{formatCurrency(invoice.total)}</span>
            </div>
          </div>

          {#if invoice.notes}
            <div class="invoice-notes">
              <h4 class="section-label">Notes:</h4>
              <p class="notes-content">{invoice.notes}</p>
            </div>
          {/if}
        </Card>

        <!-- Payment History -->
        <Card class="payment-history">
          <h2 class="card-title">Payment History</h2>
          {#if invoice.status === 'paid'}
            <div class="payment-record">
              <div class="payment-icon paid">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div class="payment-details">
                <span class="payment-amount">{formatCurrency(invoice.total)}</span>
                <span class="payment-date">Paid on {invoice.paidDate ? formatDate(invoice.paidDate) : 'N/A'}</span>
                <Badge variant="green" size="sm">Paid in Full</Badge>
              </div>
            </div>
          {:else}
            <div class="no-payments">
              <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
              <p class="text-gray-500">No payments recorded yet</p>
            </div>
          {/if}
        </Card>
      </div>

      <!-- Sidebar -->
      <div class="sidebar">
        <!-- Status Card -->
        <Card>
          <h3 class="sidebar-title">Status</h3>
          <div class="status-display">
            <Badge variant={getStatusVariant(invoice.status)} size="md">
              {invoice.status}
            </Badge>
            {#if isOverdue(invoice)}
              <Badge variant="red" size="sm">Overdue</Badge>
            {/if}
          </div>
          <div class="status-info">
            <div class="info-row">
              <span class="label">Created</span>
              <span class="value">{formatDateTime(invoice.createdAt)}</span>
            </div>
            <div class="info-row">
              <span class="label">Last Updated</span>
              <span class="value">{formatDateTime(invoice.updatedAt)}</span>
            </div>
            {#if invoice.paidDate}
              <div class="info-row">
                <span class="label">Paid On</span>
                <span class="value paid">{formatDateTime(invoice.paidDate)}</span>
              </div>
            {/if}
          </div>
        </Card>

        <!-- Amount Summary -->
        <Card>
          <h3 class="sidebar-title">Amount Summary</h3>
          <div class="amount-breakdown">
            <div class="amount-row">
              <span class="label">Subtotal</span>
              <span class="value">{formatCurrency(invoice.subtotal)}</span>
            </div>
            <div class="amount-row">
              <span class="label">Tax</span>
              <span class="value">{formatCurrency(invoice.tax)}</span>
            </div>
            <div class="amount-row total">
              <span class="label">Total</span>
              <span class="value">{formatCurrency(invoice.total)}</span>
            </div>
          </div>
        </Card>

        <!-- Activity Timeline -->
        <Card>
          <h3 class="sidebar-title">Activity</h3>
          <div class="timeline">
            {#each activities as activity}
              <div class="timeline-item">
                <div class="timeline-dot" class:created={activity.type === 'created'} class:sent={activity.type === 'sent'}></div>
                <div class="timeline-content">
                  <p class="timeline-description">{activity.description}</p>
                  <p class="timeline-date">{formatDateTime(activity.date)}</p>
                </div>
              </div>
            {/each}
          </div>
        </Card>
      </div>
    </div>
  {:else}
    <Alert variant="error">Invoice not found</Alert>
  {/if}
</div>

<!-- Delete Modal -->
<Modal
  bind:open={showDeleteModal}
  title="Delete Invoice"
  size="sm"
>
  <p>Are you sure you want to delete invoice <strong>{invoice?.invoiceNumber}</strong>? This action cannot be undone.</p>
  
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

<!-- Send Modal -->
<Modal
  bind:open={showSendModal}
  title="Send Invoice"
  size="sm"
>
  <p>Send invoice <strong>{invoice?.invoiceNumber}</strong> to the client?</p>
  <p class="text-sm text-gray-500 mt-2">This will mark the invoice as sent and notify the client.</p>
  
  {#if actionError}
    <Alert variant="error" class="mt-4">{actionError}</Alert>
  {/if}
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showSendModal = false; }} disabled={processing}>
      Cancel
    </Button>
    <Button variant="primary" on:click={confirmSend} loading={processing}>
      {processing ? 'Sending...' : 'Send Invoice'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Mark as Paid Modal -->
<Modal
  bind:open={showMarkPaidModal}
  title="Mark as Paid"
  size="sm"
>
  <p>Mark invoice <strong>{invoice?.invoiceNumber}</strong> as paid?</p>
  <p class="text-sm text-gray-500 mt-2">This will record a payment of {invoice ? formatCurrency(invoice.total) : ''}.</p>
  
  {#if actionError}
    <Alert variant="error" class="mt-4">{actionError}</Alert>
  {/if}
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showMarkPaidModal = false; }} disabled={processing}>
      Cancel
    </Button>
    <Button variant="primary" on:click={confirmMarkAsPaid} loading={processing}>
      {processing ? 'Processing...' : 'Mark as Paid'}
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

  .invoice-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 2rem;
    padding-bottom: 2rem;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .company-name {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin: 0 0 0.5rem 0;
  }

  .company-details {
    color: var(--color-gray-600);
    font-size: 0.875rem;
    line-height: 1.6;
    margin: 0;
  }

  .invoice-meta {
    text-align: right;
  }

  .meta-row {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    margin-bottom: 0.25rem;
  }

  .meta-label {
    color: var(--color-gray-500);
    font-size: 0.875rem;
  }

  .meta-value {
    color: var(--color-gray-900);
    font-weight: 500;
    font-size: 0.875rem;
  }

  .meta-value.overdue {
    color: var(--color-red-600);
  }

  .meta-value.paid {
    color: var(--color-green-600);
  }

  .overdue-badge {
    font-size: 0.75rem;
    margin-left: 0.25rem;
  }

  .invoice-to {
    margin-bottom: 2rem;
  }

  .section-label {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-500);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: 0 0 0.5rem 0;
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

  .line-items {
    margin-bottom: 2rem;
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

  .items-table .description {
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

  .invoice-totals {
    margin-left: auto;
    width: 300px;
    padding-top: 1rem;
    border-top: 2px solid var(--color-gray-200);
  }

  .total-row {
    display: flex;
    justify-content: space-between;
    padding: 0.5rem 0;
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
    border-top: 1px solid var(--color-gray-200);
  }

  .total-row.grand-total .label,
  .total-row.grand-total .value {
    font-size: 1.25rem;
    font-weight: 700;
  }

  .total-row.grand-total .value {
    color: var(--color-primary-600);
  }

  .invoice-notes {
    margin-top: 2rem;
    padding-top: 2rem;
    border-top: 1px solid var(--color-gray-200);
  }

  .notes-content {
    color: var(--color-gray-700);
    line-height: 1.6;
    margin: 0;
    white-space: pre-wrap;
  }

  .card-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
  }

  :global(.payment-history) {
    padding: 1.5rem;
  }

  .payment-record {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background-color: var(--color-green-50);
    border-radius: 0.5rem;
    border: 1px solid var(--color-green-200);
  }

  .payment-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 3rem;
    height: 3rem;
    border-radius: 50%;
    background-color: var(--color-green-100);
    color: var(--color-green-600);
  }

  .payment-details {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .payment-amount {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-green-700);
  }

  .payment-date {
    font-size: 0.875rem;
    color: var(--color-gray-600);
  }

  .no-payments {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2rem;
    color: var(--color-gray-500);
  }

  .status-display {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
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

  .info-row .value.paid {
    color: var(--color-green-600);
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

  .timeline-dot.sent {
    background-color: var(--color-green-500);
  }

  .timeline-content {
    flex: 1;
  }

  .timeline-description {
    font-size: 0.875rem;
    color: var(--color-gray-900);
    margin: 0 0 0.25rem 0;
  }

  .timeline-date {
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

    .invoice-header {
      flex-direction: column;
      gap: 1.5rem;
    }

    .invoice-meta {
      text-align: left;
    }

    .meta-row {
      justify-content: flex-start;
    }

    .invoice-totals {
      width: 100%;
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

    .items-table {
      font-size: 0.875rem;
    }

    .items-table th,
    .items-table td {
      padding: 0.5rem 0.25rem;
    }

    .items-table .description {
      width: 40%;
    }
  }
</style>
