<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import ConfirmModal from '$lib/shared/components/layout/ConfirmModal.svelte';
  import {
    getInvoiceById,
    deleteInvoice,
    markInvoiceAsPaid,
    updateInvoice,
    type Invoice
  } from '$lib/shared/api/invoices';
  import { toast } from '$lib/shared/stores/toast';

  const invoiceId = $page.params.id;

  let invoice: Invoice | null = null;
  let loading = true;
  let error: string | null = null;

  // Modals
  let showPaymentModal = false;
  let showVoidModal = false;
  let processingPayment = false;
  let processingVoid = false;
  let processingSend = false;

  // Payment form
  let paymentAmount = '';
  let paymentMethod = 'bank_transfer';
  let paymentDate = new Date().toISOString().split('T')[0];
  let paymentNote = '';

  const paymentMethodOptions = [
    { value: 'bank_transfer', label: 'Bank Transfer' },
    { value: 'credit_card', label: 'Credit Card' },
    { value: 'cash', label: 'Cash' },
    { value: 'cheque', label: 'Cheque' },
    { value: 'other', label: 'Other' }
  ];

  function getStatusVariant(status: string): 'green' | 'gray' | 'yellow' | 'red' | 'blue' {
    switch (status) {
      case 'paid':      return 'green';
      case 'sent':      return 'blue';
      case 'draft':     return 'gray';
      case 'overdue':   return 'red';
      case 'cancelled': return 'yellow';
      default:          return 'gray';
    }
  }

  function formatCurrency(value: string | number): string {
    const num = typeof value === 'string' ? parseFloat(value) : value;
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(
      isNaN(num) ? 0 : num
    );
  }

  function formatDate(date: string | null | undefined): string {
    if (!date) return '—';
    return new Date(date).toLocaleDateString('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric'
    });
  }

  // Status timeline steps
  type Step = { key: string; label: string };
  const timelineSteps: Step[] = [
    { key: 'draft', label: 'Draft' },
    { key: 'sent',  label: 'Sent' },
    { key: 'paid',  label: 'Paid' }
  ];

  function stepState(stepKey: string): 'completed' | 'active' | 'upcoming' {
    if (!invoice) return 'upcoming';
    const status = invoice.status;
    if (status === 'cancelled') return 'upcoming';
    const order: Record<string, number> = { draft: 0, sent: 1, overdue: 1, paid: 2 };
    const current = order[status] ?? 0;
    const step = order[stepKey] ?? 0;
    if (step < current) return 'completed';
    if (step === current) return 'active';
    return 'upcoming';
  }

  async function loadInvoice() {
    loading = true;
    error = null;
    try {
      invoice = await getInvoiceById(invoiceId);
      paymentAmount = invoice.total;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load invoice';
    } finally {
      loading = false;
    }
  }

  async function handleSend() {
    if (!invoice) return;
    processingSend = true;
    try {
      invoice = await updateInvoice(invoice.id, { status: 'sent' });
      toast.success('Invoice marked as sent');
    } catch (err) {
      toast.error('Failed to send invoice');
    } finally {
      processingSend = false;
    }
  }

  async function handleRecordPayment() {
    if (!invoice) return;
    processingPayment = true;
    try {
      invoice = await markInvoiceAsPaid(invoice.id);
      toast.success('Payment recorded — invoice marked as paid');
      showPaymentModal = false;
    } catch (err) {
      toast.error('Failed to record payment');
    } finally {
      processingPayment = false;
    }
  }

  async function handleVoid() {
    if (!invoice) return;
    processingVoid = true;
    try {
      invoice = await updateInvoice(invoice.id, { status: 'cancelled' });
      toast.success('Invoice voided');
      showVoidModal = false;
    } catch (err) {
      toast.error('Failed to void invoice');
    } finally {
      processingVoid = false;
    }
  }

  function handleDownloadPdf() {
    toast.info('PDF download is not yet available');
  }

  onMount(loadInvoice);
</script>

<svelte:head>
  <title>
    {invoice ? `Invoice ${invoice.invoiceNumber}` : 'Invoice Detail'} | ERP System
  </title>
</svelte:head>

<div class="page-container">
  <!-- Back link -->
  <button class="back-link" on:click={() => goto('/invoices')}>
    ← Back to Invoices
  </button>

  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading invoice…</p>
    </div>
  {:else if error}
    <Alert variant="error">{error}</Alert>
    <div class="mt-4">
      <Button variant="secondary" on:click={() => goto('/invoices')}>Back to Invoices</Button>
    </div>
  {:else if invoice}
    <!-- Header card -->
    <Card>
      <div class="invoice-header">
        <div class="invoice-meta">
          <h1 class="invoice-number">{invoice.invoiceNumber}</h1>
          <div class="invoice-client">{invoice.clientName}</div>
          <div class="invoice-dates">
            <span>Issued: {formatDate(invoice.createdAt)}</span>
            <span class="date-sep">·</span>
            <span class:overdue-text={invoice.status === 'overdue'}>
              Due: {formatDate(invoice.dueDate)}
            </span>
            {#if invoice.paidDate}
              <span class="date-sep">·</span>
              <span class="paid-text">Paid: {formatDate(invoice.paidDate)}</span>
            {/if}
          </div>
        </div>

        <div class="invoice-right">
          <div class="status-badge-wrap">
            <Badge variant={getStatusVariant(invoice.status)} size="md">
              {invoice.status.toUpperCase()}
            </Badge>
          </div>
          <div class="invoice-total-display">{formatCurrency(invoice.total)}</div>

          <!-- Action buttons -->
          <div class="action-buttons">
            {#if invoice.status === 'draft'}
              <Button variant="primary" size="sm" loading={processingSend} on:click={handleSend}>
                Send Invoice
              </Button>
              <Button variant="secondary" size="sm" on:click={() => goto(`/invoices/${invoice?.id}/edit`)}>
                Edit
              </Button>
            {/if}
            {#if invoice.status === 'sent' || invoice.status === 'overdue'}
              <Button variant="primary" size="sm" on:click={() => (showPaymentModal = true)}>
                Record Payment
              </Button>
            {/if}
            <Button variant="secondary" size="sm" on:click={handleDownloadPdf}>
              Download PDF
            </Button>
            {#if invoice.status !== 'cancelled' && invoice.status !== 'paid'}
              <Button variant="danger" size="sm" on:click={() => (showVoidModal = true)}>
                Void
              </Button>
            {/if}
          </div>
        </div>
      </div>

      <!-- Status timeline -->
      {#if invoice.status !== 'cancelled'}
        <div class="timeline-wrapper">
          {#each timelineSteps as step, i}
            <div class="timeline-step" data-state={stepState(step.key)}>
              <div class="step-dot" data-state={stepState(step.key)}></div>
              <div class="step-label">{step.label}</div>
            </div>
            {#if i < timelineSteps.length - 1}
              <div
                class="step-connector"
                class:step-connector-done={stepState(timelineSteps[i + 1].key) !== 'upcoming'}
              ></div>
            {/if}
          {/each}
        </div>
      {:else}
        <div class="cancelled-banner">
          <Badge variant="yellow">CANCELLED / VOID</Badge>
        </div>
      {/if}
    </Card>

    <!-- Line items -->
    <Card>
      <h2 class="section-title">Line Items</h2>
      <div class="line-items-wrapper">
        <table class="line-items-table">
          <thead>
            <tr>
              <th>Description</th>
              <th class="text-right">Qty</th>
              <th class="text-right">Unit Price</th>
              <th class="text-right">Total</th>
            </tr>
          </thead>
          <tbody>
            {#each invoice.items as item}
              <tr>
                <td>{item.description}</td>
                <td class="text-right">{item.quantity}</td>
                <td class="text-right">{formatCurrency(item.unitPrice)}</td>
                <td class="text-right font-medium">{formatCurrency(item.total)}</td>
              </tr>
            {/each}
          </tbody>
          <tfoot>
            <tr class="subtotal-row">
              <td colspan="3" class="text-right">Subtotal</td>
              <td class="text-right">{formatCurrency(invoice.subtotal)}</td>
            </tr>
            <tr class="tax-row">
              <td colspan="3" class="text-right">Tax</td>
              <td class="text-right">{formatCurrency(invoice.tax)}</td>
            </tr>
            <tr class="total-row">
              <td colspan="3" class="text-right font-bold">Total</td>
              <td class="text-right font-bold grand-total">{formatCurrency(invoice.total)}</td>
            </tr>
          </tfoot>
        </table>
      </div>
    </Card>

    <!-- Payment history -->
    <Card>
      <h2 class="section-title">Payment History</h2>
      {#if invoice.paidDate}
        <div class="payment-entry">
          <div class="payment-icon">✓</div>
          <div class="payment-info">
            <div class="payment-amount">{formatCurrency(invoice.total)}</div>
            <div class="payment-date">Paid on {formatDate(invoice.paidDate)}</div>
          </div>
          <Badge variant="green">Paid in Full</Badge>
        </div>
      {:else}
        <p class="no-payments">No payments recorded yet.</p>
        {#if invoice.status === 'sent' || invoice.status === 'overdue'}
          <Button variant="secondary" size="sm" on:click={() => (showPaymentModal = true)}>
            Record a Payment
          </Button>
        {/if}
      {/if}
    </Card>

    <!-- Notes -->
    {#if invoice.notes}
      <Card>
        <h2 class="section-title">Notes</h2>
        <p class="notes-text">{invoice.notes}</p>
      </Card>
    {/if}
  {/if}
</div>

<!-- Record Payment modal -->
<Modal bind:open={showPaymentModal} title="Record Payment" size="sm">
  <div class="payment-form">
    <Input
      id="paymentAmount"
      label="Amount"
      type="number"
      bind:value={paymentAmount}
      helpText="Full invoice amount is pre-filled"
    />
    <Select
      id="paymentMethod"
      label="Payment Method"
      options={paymentMethodOptions}
      bind:value={paymentMethod}
    />
    <Input
      id="paymentDate"
      label="Payment Date"
      type="date"
      bind:value={paymentDate}
    />
    <Input
      id="paymentNote"
      label="Note (optional)"
      type="text"
      placeholder="Reference number, memo…"
      bind:value={paymentNote}
    />
  </div>
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={processingPayment}>Cancel</Button>
    <Button variant="primary" loading={processingPayment} on:click={handleRecordPayment}>
      Record Payment
    </Button>
  </svelte:fragment>
</Modal>

<!-- Void confirmation -->
<ConfirmModal
  bind:open={showVoidModal}
  title="Void Invoice"
  message={`Are you sure you want to void invoice ${invoice?.invoiceNumber}? This will cancel it and cannot be undone.`}
  confirmLabel="Void Invoice"
  variant="danger"
  loading={processingVoid}
  on:confirm={handleVoid}
  on:cancel={() => (showVoidModal = false)}
/>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 860px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .back-link {
    background: none;
    border: none;
    cursor: pointer;
    color: #6b7280;
    font-size: 0.875rem;
    padding: 0;
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    transition: color 0.15s;
  }

  .back-link:hover { color: #111827; }

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 3rem;
    gap: 1rem;
    color: #6b7280;
  }

  /* Header */
  .invoice-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1.5rem;
    flex-wrap: wrap;
  }

  .invoice-number {
    font-size: 1.5rem;
    font-weight: 700;
    color: #111827;
    font-family: monospace;
    margin: 0 0 0.25rem 0;
  }

  .invoice-client {
    font-size: 1.125rem;
    color: #374151;
    margin-bottom: 0.5rem;
  }

  .invoice-dates {
    font-size: 0.875rem;
    color: #6b7280;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .date-sep { color: #d1d5db; }
  .overdue-text { color: #ef4444; font-weight: 600; }
  .paid-text { color: #10b981; }

  .invoice-right {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.75rem;
  }

  .invoice-total-display {
    font-size: 2rem;
    font-weight: 800;
    color: #111827;
  }

  .action-buttons {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  /* Timeline */
  .timeline-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 1.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid #e5e7eb;
    gap: 0;
  }

  .timeline-step {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.375rem;
  }

  .step-dot {
    width: 1.5rem;
    height: 1.5rem;
    border-radius: 50%;
    border: 2px solid #d1d5db;
    background: #fff;
    position: relative;
    transition: all 0.2s;
  }

  .step-dot[data-state="completed"] {
    background: #10b981;
    border-color: #10b981;
  }

  .step-dot[data-state="active"] {
    background: #6366f1;
    border-color: #6366f1;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
  }

  .step-dot[data-state="completed"]::after {
    content: '✓';
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 0.7rem;
    font-weight: 700;
  }

  .step-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: #9ca3af;
  }

  .timeline-step[data-state="active"] .step-label { color: #6366f1; font-weight: 700; }
  .timeline-step[data-state="completed"] .step-label { color: #10b981; }

  .step-connector {
    flex: 1;
    height: 2px;
    background: #e5e7eb;
    min-width: 3rem;
    max-width: 6rem;
    margin-bottom: 1.25rem;
    transition: background-color 0.2s;
  }

  .step-connector-done { background: #10b981; }

  .cancelled-banner {
    margin-top: 1.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid #e5e7eb;
    display: flex;
    justify-content: center;
  }

  /* Line items table */
  .section-title {
    font-size: 1rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 1rem 0;
  }

  .line-items-wrapper { overflow-x: auto; }

  .line-items-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .line-items-table th {
    padding: 0.5rem 0.75rem;
    text-align: left;
    font-weight: 600;
    color: #6b7280;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    border-bottom: 2px solid #e5e7eb;
  }

  .line-items-table td {
    padding: 0.625rem 0.75rem;
    color: #374151;
    border-bottom: 1px solid #f3f4f6;
  }

  .text-right { text-align: right; }
  .font-medium { font-weight: 500; }
  .font-bold { font-weight: 700; }

  .subtotal-row td,
  .tax-row td {
    padding: 0.4rem 0.75rem;
    color: #6b7280;
    font-size: 0.875rem;
    border-bottom: none;
  }

  .total-row td {
    padding: 0.625rem 0.75rem;
    border-top: 2px solid #e5e7eb;
    border-bottom: none;
  }

  .grand-total {
    font-size: 1.1rem;
    color: #111827;
  }

  /* Payment history */
  .payment-entry {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.75rem;
    background: #f0fdf4;
    border: 1px solid #bbf7d0;
    border-radius: 0.5rem;
  }

  .payment-icon {
    width: 2rem;
    height: 2rem;
    background: #10b981;
    color: #fff;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    flex-shrink: 0;
  }

  .payment-info { flex: 1; }
  .payment-amount { font-weight: 600; color: #111827; }
  .payment-date { font-size: 0.8125rem; color: #6b7280; }

  .no-payments {
    color: #9ca3af;
    font-size: 0.875rem;
    margin-bottom: 0.75rem;
  }

  .notes-text {
    color: #374151;
    font-size: 0.875rem;
    line-height: 1.6;
    white-space: pre-wrap;
    margin: 0;
  }

  /* Payment form in modal */
  .payment-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
</style>
