<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Textarea from '$lib/shared/components/forms/Textarea.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import { createInvoice } from '$lib/shared/api/invoices';
  import { getClients, type Client } from '$lib/shared/api/clients';
  import { toast } from '$lib/shared/stores/toast';

  // ── Client selector ────────────────────────────────────────────────────────
  let clients: Client[] = [];
  let clientOptions: { value: string; label: string }[] = [];
  let loadingClients = false;

  // ── Form state ─────────────────────────────────────────────────────────────
  let clientId = '';
  let invoiceType = 'standard';
  let issueDate = new Date().toISOString().split('T')[0];
  let paymentTerm = 'net30';
  let dueDate = '';
  let currency = 'USD';
  let notes = '';
  let terms = '';
  let submitting = false;
  let errors: Record<string, string> = {};

  // ── Line items ─────────────────────────────────────────────────────────────
  interface LineItem {
    id: string;
    description: string;
    quantity: number;
    unitPrice: number;
    discount: number;
    taxRate: number;
  }

  let lineItems: LineItem[] = [newLineItem()];

  function newLineItem(): LineItem {
    return {
      id: crypto.randomUUID(),
      description: '',
      quantity: 1,
      unitPrice: 0,
      discount: 0,
      taxRate: 0
    };
  }

  function addLineItem() {
    lineItems = [...lineItems, newLineItem()];
  }

  function removeLineItem(id: string) {
    if (lineItems.length === 1) return;
    lineItems = lineItems.filter((l) => l.id !== id);
  }

  function lineTotal(item: LineItem): number {
    const gross = item.quantity * item.unitPrice;
    const afterDiscount = gross * (1 - item.discount / 100);
    return afterDiscount * (1 + item.taxRate / 100);
  }

  $: subtotal = lineItems.reduce((s, i) => s + i.quantity * i.unitPrice, 0);
  $: discountTotal = lineItems.reduce(
    (s, i) => s + i.quantity * i.unitPrice * (i.discount / 100),
    0
  );
  $: taxTotal = lineItems.reduce((s, i) => {
    const afterDiscount = i.quantity * i.unitPrice * (1 - i.discount / 100);
    return s + afterDiscount * (i.taxRate / 100);
  }, 0);
  $: grandTotal = subtotal - discountTotal + taxTotal;

  // ── Payment terms → due date ───────────────────────────────────────────────
  const termDays: Record<string, number> = {
    immediate: 0,
    net15: 15,
    net30: 30,
    net45: 45,
    net60: 60
  };

  $: {
    if (issueDate && paymentTerm) {
      const d = new Date(issueDate);
      d.setDate(d.getDate() + (termDays[paymentTerm] ?? 30));
      dueDate = d.toISOString().split('T')[0];
    }
  }

  const invoiceTypeOptions = [
    { value: 'standard', label: 'Standard Invoice' },
    { value: 'credit_note', label: 'Credit Note' },
    { value: 'debit_note', label: 'Debit Note' },
    { value: 'recurring', label: 'Recurring Invoice' }
  ];

  const paymentTermOptions = [
    { value: 'immediate', label: 'Due on Receipt' },
    { value: 'net15', label: 'Net 15' },
    { value: 'net30', label: 'Net 30' },
    { value: 'net45', label: 'Net 45' },
    { value: 'net60', label: 'Net 60' }
  ];

  const currencyOptions = [
    { value: 'USD', label: 'USD — US Dollar' },
    { value: 'EUR', label: 'EUR — Euro' },
    { value: 'GBP', label: 'GBP — British Pound' }
  ];

  function formatCurrency(value: number): string {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(value);
  }

  // ── Validation ─────────────────────────────────────────────────────────────
  function validate(): boolean {
    errors = {};
    if (!clientId) errors.clientId = 'Client is required';
    const validItems = lineItems.filter((i) => i.description.trim() && i.quantity > 0);
    if (validItems.length === 0)
      errors.lineItems = 'At least one line item with a description and quantity is required';
    return Object.keys(errors).length === 0;
  }

  // ── Submit ─────────────────────────────────────────────────────────────────
  async function handleSubmit() {
    if (!validate()) return;
    submitting = true;
    try {
      const items = lineItems
        .filter((i) => i.description.trim() && i.quantity > 0)
        .map((i) => ({
          description: i.description,
          quantity: i.quantity,
          unitPrice: String(i.unitPrice),
          total: String(lineTotal(i).toFixed(2))
        }));

      const invoice = await createInvoice({ clientId, items, dueDate, notes });
      toast.success('Invoice created successfully');
      goto(`/invoices/${invoice.id}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create invoice');
    } finally {
      submitting = false;
    }
  }

  onMount(async () => {
    loadingClients = true;
    try {
      const res = await getClients({ pageSize: 200 });
      clients = res.data;
      clientOptions = clients.map((c) => ({ value: c.id, label: c.name }));
    } catch {
      toast.error('Could not load clients');
    } finally {
      loadingClients = false;
    }
  });
</script>

<svelte:head>
  <title>New Invoice | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">New Invoice</h1>
      <p class="page-description">Create a new invoice for a client</p>
    </div>
    <div class="header-actions">
      <Button variant="secondary" on:click={() => goto('/invoices')}>Cancel</Button>
      <Button variant="primary" loading={submitting} on:click={handleSubmit}>
        Create Invoice
      </Button>
    </div>
  </div>

  {#if errors.lineItems}
    <Alert variant="error">{errors.lineItems}</Alert>
  {/if}

  <div class="form-grid">
    <!-- Left column -->
    <div class="form-left">
      <!-- Client & meta -->
      <Card>
        <h2 class="section-title">Invoice Details</h2>
        <div class="fields-grid">
          <div class="field-full">
            {#if loadingClients}
              <div class="flex items-center gap-2 text-sm text-gray-500">
                <Spinner size="sm" /> Loading clients…
              </div>
            {:else}
              <Select
                id="clientId"
                label="Client"
                options={clientOptions}
                bind:value={clientId}
                placeholder="Select a client"
                required
                error={errors.clientId}
              />
            {/if}
          </div>

          <Select
            id="invoiceType"
            label="Invoice Type"
            options={invoiceTypeOptions}
            bind:value={invoiceType}
          />

          <Select
            id="currency"
            label="Currency"
            options={currencyOptions}
            bind:value={currency}
          />

          <Input
            id="issueDate"
            label="Issue Date"
            type="date"
            bind:value={issueDate}
            required
          />

          <Select
            id="paymentTerm"
            label="Payment Terms"
            options={paymentTermOptions}
            bind:value={paymentTerm}
          />

          <Input
            id="dueDate"
            label="Due Date"
            type="date"
            bind:value={dueDate}
            helpText="Auto-calculated from payment terms"
          />
        </div>
      </Card>

      <!-- Line items -->
      <Card>
        <div class="section-header">
          <h2 class="section-title">Line Items</h2>
          <Button variant="secondary" size="sm" on:click={addLineItem}>+ Add Item</Button>
        </div>

        <div class="line-items-table-wrapper">
          <table class="line-items-table">
            <thead>
              <tr>
                <th class="col-desc">Description</th>
                <th class="col-qty">Qty</th>
                <th class="col-price">Unit Price</th>
                <th class="col-disc">Disc %</th>
                <th class="col-tax">Tax %</th>
                <th class="col-total">Total</th>
                <th class="col-action"></th>
              </tr>
            </thead>
            <tbody>
              {#each lineItems as item (item.id)}
                <tr class="line-item-row">
                  <td class="col-desc">
                    <input
                      class="cell-input"
                      type="text"
                      placeholder="Item description"
                      bind:value={item.description}
                    />
                  </td>
                  <td class="col-qty">
                    <input
                      class="cell-input text-right"
                      type="number"
                      min="0"
                      step="1"
                      bind:value={item.quantity}
                    />
                  </td>
                  <td class="col-price">
                    <input
                      class="cell-input text-right"
                      type="number"
                      min="0"
                      step="0.01"
                      bind:value={item.unitPrice}
                    />
                  </td>
                  <td class="col-disc">
                    <input
                      class="cell-input text-right"
                      type="number"
                      min="0"
                      max="100"
                      step="0.1"
                      bind:value={item.discount}
                    />
                  </td>
                  <td class="col-tax">
                    <input
                      class="cell-input text-right"
                      type="number"
                      min="0"
                      max="100"
                      step="0.1"
                      bind:value={item.taxRate}
                    />
                  </td>
                  <td class="col-total text-right font-medium">
                    {formatCurrency(lineTotal(item))}
                  </td>
                  <td class="col-action">
                    <button
                      type="button"
                      class="remove-btn"
                      disabled={lineItems.length === 1}
                      on:click={() => removeLineItem(item.id)}
                      aria-label="Remove line item"
                    >
                      ×
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        <!-- Summary -->
        <div class="totals-section">
          <div class="totals-grid">
            <span class="totals-label">Subtotal</span>
            <span class="totals-value">{formatCurrency(subtotal)}</span>

            {#if discountTotal > 0}
              <span class="totals-label text-green-600">Discount</span>
              <span class="totals-value text-green-600">−{formatCurrency(discountTotal)}</span>
            {/if}

            <span class="totals-label">Tax</span>
            <span class="totals-value">{formatCurrency(taxTotal)}</span>

            <span class="totals-label grand-label">Total</span>
            <span class="totals-value grand-value">{formatCurrency(grandTotal)}</span>
          </div>
        </div>
      </Card>
    </div>

    <!-- Right column: notes & terms -->
    <div class="form-right">
      <Card>
        <h2 class="section-title">Notes &amp; Terms</h2>
        <div class="fields-stack">
          <Textarea
            id="notes"
            label="Notes"
            placeholder="Additional notes visible to the client…"
            bind:value={notes}
            rows={4}
          />
          <Textarea
            id="terms"
            label="Terms &amp; Conditions"
            placeholder="Payment terms, late fees, etc…"
            bind:value={terms}
            rows={4}
          />
        </div>
      </Card>

      <Card>
        <h2 class="section-title">Summary</h2>
        <div class="summary-list">
          <div class="summary-row">
            <span>Subtotal</span><span>{formatCurrency(subtotal)}</span>
          </div>
          {#if discountTotal > 0}
            <div class="summary-row text-green-700">
              <span>Discount</span><span>−{formatCurrency(discountTotal)}</span>
            </div>
          {/if}
          <div class="summary-row">
            <span>Tax</span><span>{formatCurrency(taxTotal)}</span>
          </div>
          <div class="summary-row summary-total-row">
            <span>Total ({currency})</span><span>{formatCurrency(grandTotal)}</span>
          </div>
        </div>
      </Card>
    </div>
  </div>

  <div class="bottom-actions">
    <Button variant="secondary" on:click={() => goto('/invoices')}>Cancel</Button>
    <Button variant="primary" loading={submitting} on:click={handleSubmit}>
      Create Invoice
    </Button>
  </div>
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
  }

  .page-title {
    font-size: 1.75rem;
    font-weight: 700;
    color: #111827;
    margin: 0;
  }

  .page-description {
    color: #6b7280;
    margin-top: 0.25rem;
    margin-bottom: 0;
  }

  .header-actions {
    display: flex;
    gap: 0.75rem;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 320px;
    gap: 1.5rem;
    align-items: start;
  }

  @media (max-width: 900px) {
    .form-grid {
      grid-template-columns: 1fr;
    }
  }

  .form-left,
  .form-right {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .section-title {
    font-size: 1rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 1rem 0;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .section-header .section-title {
    margin: 0;
  }

  .fields-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .field-full {
    grid-column: 1 / -1;
  }

  .fields-stack {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  /* Line items table */
  .line-items-table-wrapper {
    overflow-x: auto;
  }

  .line-items-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .line-items-table thead tr {
    border-bottom: 2px solid #e5e7eb;
  }

  .line-items-table th {
    padding: 0.5rem 0.5rem;
    text-align: left;
    font-weight: 600;
    color: #6b7280;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .line-item-row {
    border-bottom: 1px solid #f3f4f6;
  }

  .line-item-row td {
    padding: 0.375rem 0.25rem;
    vertical-align: middle;
  }

  .col-desc  { width: 35%; }
  .col-qty   { width: 8%; }
  .col-price { width: 13%; }
  .col-disc  { width: 10%; }
  .col-tax   { width: 10%; }
  .col-total { width: 14%; padding-right: 0.5rem; }
  .col-action{ width: 5%; text-align: center; }

  .cell-input {
    width: 100%;
    border: 1px solid #e5e7eb;
    border-radius: 0.375rem;
    padding: 0.375rem 0.5rem;
    font-size: 0.875rem;
    color: #111827;
    background: #fff;
    transition: border-color 0.15s;
    box-sizing: border-box;
  }

  .cell-input:focus {
    outline: none;
    border-color: #6366f1;
    box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
  }

  .text-right { text-align: right; }

  .remove-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: #9ca3af;
    font-size: 1.25rem;
    line-height: 1;
    padding: 0.25rem;
    border-radius: 0.25rem;
    transition: color 0.1s, background-color 0.1s;
  }

  .remove-btn:hover:not(:disabled) {
    color: #ef4444;
    background-color: #fee2e2;
  }

  .remove-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  /* Totals */
  .totals-section {
    display: flex;
    justify-content: flex-end;
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e5e7eb;
  }

  .totals-grid {
    display: grid;
    grid-template-columns: auto auto;
    gap: 0.35rem 1.5rem;
    text-align: right;
    font-size: 0.875rem;
  }

  .totals-label {
    color: #6b7280;
    text-align: left;
  }

  .totals-value {
    font-weight: 500;
    color: #111827;
  }

  .grand-label {
    font-weight: 700;
    color: #111827;
    font-size: 1rem;
    padding-top: 0.5rem;
    border-top: 2px solid #e5e7eb;
  }

  .grand-value {
    font-weight: 700;
    color: #111827;
    font-size: 1.1rem;
    padding-top: 0.5rem;
    border-top: 2px solid #e5e7eb;
  }

  /* Side summary card */
  .summary-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .summary-row {
    display: flex;
    justify-content: space-between;
    font-size: 0.875rem;
    color: #374151;
  }

  .summary-total-row {
    font-weight: 700;
    font-size: 1rem;
    color: #111827;
    padding-top: 0.5rem;
    border-top: 2px solid #e5e7eb;
    margin-top: 0.25rem;
  }

  .bottom-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    margin-top: 1.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid #e5e7eb;
  }
</style>
