<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import DatePicker from '$lib/shared/components/forms/DatePicker.svelte';
  import Textarea from '$lib/shared/components/forms/Textarea.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import { 
    getInvoiceById, 
    updateInvoice,
    type Invoice,
    type InvoiceItem,
    type InvoiceStatus 
  } from '$lib/shared/api/invoices';
  import { getClients, type Client } from '$lib/shared/api/clients';

  const invoiceId = $page.params.id;

  let invoice: Invoice | null = null;
  let loading = true;
  let error: string | null = null;
  let submitting = false;
  let errors: Record<string, string> = {};

  // Form state
  let selectedClientId = '';
  let invoiceNumber = '';
  let issueDate = '';
  let dueDate = '';
  let notes = '';
  let taxRate = '8';
  let currentStatus: InvoiceStatus = 'draft';

  // Line items
  let lineItems: Array<{
    id: string;
    description: string;
    quantity: string;
    unitPrice: string;
  }> = [];

  // Clients data
  let clients: Client[] = [];
  let loadingClients = true;

  // Computed values
  $: subtotal = lineItems.reduce((sum, item) => {
    const price = parseFloat(item.unitPrice) || 0;
    const qty = parseInt(item.quantity) || 0;
    return sum + (qty * price);
  }, 0);

  $: taxAmount = subtotal * ((parseFloat(taxRate) || 0) / 100);
  $: total = subtotal + taxAmount;

  $: clientOptions = clients.map(client => ({
    value: client.id,
    label: `${client.name} (${client.email})`
  }));

  $: isEditable = currentStatus !== 'paid' && currentStatus !== 'cancelled';

  onMount(async () => {
    await Promise.all([loadInvoice(), loadClients()]);
  });

  async function loadInvoice() {
    loading = true;
    error = null;
    
    try {
      invoice = await getInvoiceById(invoiceId);
      
      // Populate form with invoice data
      selectedClientId = invoice.clientId;
      invoiceNumber = invoice.invoiceNumber;
      issueDate = invoice.createdAt.split('T')[0];
      dueDate = invoice.dueDate;
      notes = invoice.notes;
      currentStatus = invoice.status;
      
      // Calculate tax rate from invoice data
      const subtotalValue = parseFloat(invoice.subtotal) || 0;
      const taxValue = parseFloat(invoice.tax) || 0;
      taxRate = String(subtotalValue > 0 ? (taxValue / subtotalValue) * 100 : 8);
      
      // Populate line items
      lineItems = invoice.items.map((item, index) => ({
        id: `item-${index}`,
        description: item.description,
        quantity: String(item.quantity),
        unitPrice: item.unitPrice
      }));
      
      // Ensure at least one line item
      if (lineItems.length === 0) {
        lineItems = [{ id: crypto.randomUUID(), description: '', quantity: '1', unitPrice: '' }];
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load invoice';
    } finally {
      loading = false;
    }
  }

  async function loadClients() {
    loadingClients = true;
    try {
      const response = await getClients();
      clients = response.data;
    } catch (err) {
      console.error('Failed to load clients:', err);
    } finally {
      loadingClients = false;
    }
  }

  function addLineItem() {
    lineItems = [
      ...lineItems,
      { id: crypto.randomUUID(), description: '', quantity: '1', unitPrice: '' }
    ];
  }

  function removeLineItem(index: number) {
    if (lineItems.length > 1) {
      lineItems = lineItems.filter((_, i) => i !== index);
    }
  }

  function getLineTotal(item: typeof lineItems[0]): number {
    const price = parseFloat(item.unitPrice) || 0;
    const qty = parseInt(item.quantity) || 0;
    return qty * price;
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

  function validateForm(): boolean {
    errors = {};

    if (!selectedClientId) {
      errors.clientId = 'Please select a client';
    }

    if (!dueDate) {
      errors.dueDate = 'Due date is required';
    } else if (new Date(dueDate) < new Date(issueDate)) {
      errors.dueDate = 'Due date must be after issue date';
    }

    // Validate line items
    let hasLineItemErrors = false;
    lineItems.forEach((item, index) => {
      if (!item.description.trim()) {
        errors[`item_${index}_description`] = 'Description is required';
        hasLineItemErrors = true;
      }
      if (!item.quantity || parseInt(item.quantity) <= 0) {
        errors[`item_${index}_quantity`] = 'Quantity must be greater than 0';
        hasLineItemErrors = true;
      }
      if (!item.unitPrice || parseFloat(item.unitPrice) <= 0) {
        errors[`item_${index}_price`] = 'Price must be greater than 0';
        hasLineItemErrors = true;
      }
    });

    if (hasLineItemErrors) {
      errors.lineItems = 'Please fix line item errors';
    }

    return Object.keys(errors).length === 0;
  }

  async function handleSubmit() {
    if (!validateForm()) {
      return;
    }

    const items: InvoiceItem[] = lineItems.map(item => ({
      description: item.description,
      quantity: parseInt(item.quantity) || 1,
      unitPrice: item.unitPrice,
      total: getLineTotal(item).toFixed(2)
    }));

    const data = {
      items,
      dueDate,
      notes: notes || undefined
    };

    submitting = true;
    error = null;

    try {
      await updateInvoice(invoiceId, data);
      goto(`/invoices/${invoiceId}`);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to update invoice';
      submitting = false;
    }
  }

  function handleCancel() {
    goto(`/invoices/${invoiceId}`);
  }

  function formatCurrency(value: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(value);
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }
</script>

<svelte:head>
  <title>{invoice ? `Edit Invoice ${invoice.invoiceNumber}` : 'Edit Invoice'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading invoice...</p>
    </div>
  {:else if error && !invoice}
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
          <h1 class="page-title">Edit Invoice</h1>
          <span class="invoice-number">{invoice.invoiceNumber}</span>
          <Badge variant={getStatusVariant(currentStatus)} size="md">
            {currentStatus}
          </Badge>
        </div>
        <p class="page-description">
          Last updated: {formatDate(invoice.updatedAt)}
        </p>
      </div>
      <div class="header-actions">
        <Button variant="secondary" on:click={handleCancel} disabled={submitting}>
          Cancel
        </Button>
      </div>
    </div>

    {#if !isEditable}
      <Alert variant="warning" class="mb-4">
        This invoice cannot be edited because it has been {currentStatus}. 
        Only draft, sent, or overdue invoices can be modified.
      </Alert>
    {/if}

    {#if error}
      <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
        {error}
      </Alert>
    {/if}

    <form on:submit|preventDefault={handleSubmit}>
      <div class="form-grid">
        <!-- Client Selection -->
        <Card class="form-card">
          <h2 class="section-title">Client Information</h2>
          <div class="form-group">
            <Select
              id="client"
              label="Client"
              options={clientOptions}
              bind:value={selectedClientId}
              placeholder="Select a client"
              required
              error={errors.clientId}
              disabled={!isEditable}
            />
            {#if selectedClientId}
              {@const selectedClient = clients.find(c => c.id === selectedClientId)}
              {#if selectedClient}
                <div class="client-preview">
                  <Badge variant="blue" size="sm">{selectedClient.code}</Badge>
                  <span class="client-email">{selectedClient.email}</span>
                  {#if selectedClient.phone}
                    <span class="client-phone">{selectedClient.phone}</span>
                  {/if}
                </div>
              {/if}
            {/if}
          </div>
        </Card>

        <!-- Invoice Details -->
        <Card class="form-card">
          <h2 class="section-title">Invoice Details</h2>
          <div class="form-row">
            <div class="form-group">
              <Input
                id="invoiceNumber"
                label="Invoice Number"
                type="text"
                bind:value={invoiceNumber}
                required
                readonly
                disabled
              />
            </div>
            <div class="form-group">
              <DatePicker
                id="issueDate"
                label="Issue Date"
                bind:value={issueDate}
                required
                disabled
              />
            </div>
            <div class="form-group">
              <DatePicker
                id="dueDate"
                label="Due Date"
                bind:value={dueDate}
                required
                error={errors.dueDate}
                min={issueDate}
                disabled={!isEditable}
              />
            </div>
          </div>
        </Card>

        <!-- Line Items -->
        <Card class="form-card full-width">
          <div class="line-items-header">
            <h2 class="section-title">Line Items</h2>
            {#if isEditable}
              <Button variant="secondary" size="sm" on:click={addLineItem} type="button">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Add Item
              </Button>
            {/if}
          </div>

          {#if errors.lineItems}
            <Alert variant="error" class="mb-4">{errors.lineItems}</Alert>
          {/if}

          <div class="line-items-table">
            <table>
              <thead>
                <tr>
                  <th class="description-col">Description</th>
                  <th class="quantity-col">Quantity</th>
                  <th class="price-col">Unit Price</th>
                  <th class="total-col">Total</th>
                  {#if isEditable}
                    <th class="actions-col"></th>
                  {/if}
                </tr>
              </thead>
              <tbody>
                {#each lineItems as item, index (item.id)}
                  <tr>
                    <td>
                      <Input
                        id="item-desc-{index}"
                        label=""
                        type="text"
                        placeholder="Item description"
                        bind:value={item.description}
                        error={errors[`item_${index}_description`]}
                        disabled={!isEditable}
                      />
                    </td>
                    <td>
                      <Input
                        id="item-qty-{index}"
                        label=""
                        type="number"
                        bind:value={item.quantity}
                        min="1"
                        step="1"
                        error={errors[`item_${index}_quantity`]}
                        disabled={!isEditable}
                      />
                    </td>
                    <td>
                      <Input
                        id="item-price-{index}"
                        label=""
                        type="number"
                        placeholder="0.00"
                        bind:value={item.unitPrice}
                        min="0"
                        step="0.01"
                        error={errors[`item_${index}_price`]}
                        disabled={!isEditable}
                      />
                    </td>
                    <td class="line-total">
                      {formatCurrency(getLineTotal(item))}
                    </td>
                    {#if isEditable}
                      <td>
                        <Button
                          variant="ghost"
                          size="sm"
                          on:click={() => removeLineItem(index)}
                          disabled={lineItems.length === 1 || !isEditable}
                          type="button"
                        >
                          <svg class="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </Button>
                      </td>
                    {/if}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </Card>

        <!-- Totals -->
        <Card class="form-card totals-card">
          <h2 class="section-title">Totals</h2>
          <div class="totals-section">
            <div class="totals-row">
              <span class="label">Subtotal</span>
              <span class="value">{formatCurrency(subtotal)}</span>
            </div>
            <div class="totals-row tax-row">
              <div class="tax-input">
                <label for="taxRate">Tax Rate (%)</label>
                <Input
                  id="taxRate"
                  label=""
                  type="number"
                  bind:value={taxRate}
                  min="0"
                  max="100"
                  step="0.01"
                  disabled={!isEditable}
                />
              </div>
              <span class="value">{formatCurrency(taxAmount)}</span>
            </div>
            <div class="totals-row total-row">
              <span class="label">Total</span>
              <span class="value total">{formatCurrency(total)}</span>
            </div>
          </div>
        </Card>

        <!-- Notes -->
        <Card class="form-card">
          <h2 class="section-title">Additional Information</h2>
          <Textarea
            id="notes"
            label="Notes"
            bind:value={notes}
            placeholder="Add any additional notes or terms..."
            rows={4}
            disabled={!isEditable}
          />
        </Card>

        <!-- Actions -->
        {#if isEditable}
          <Card class="form-card full-width actions-card">
            <div class="actions-row">
              <div class="action-info">
                <p class="info-text">
                  <svg class="w-4 h-4 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  Changes will update the invoice total and line items.
                </p>
              </div>
              <div class="action-buttons">
                <Button variant="secondary" on:click={handleCancel} disabled={submitting}>
                  Cancel
                </Button>
                <Button variant="primary" type="submit" loading={submitting}>
                  {submitting ? 'Saving...' : 'Save Changes'}
                </Button>
              </div>
            </div>
          </Card>
        {:else}
          <Card class="form-card full-width actions-card readonly">
            <div class="readonly-notice">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
              <span>This invoice is read-only and cannot be modified.</span>
              <Button variant="secondary" size="sm" on:click={handleCancel}>
                Back to Invoice
              </Button>
            </div>
          </Card>
        {/if}
      </div>
    </form>
  {:else}
    <Alert variant="error">Invoice not found</Alert>
  {/if}
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1200px;
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

  .invoice-number {
    font-size: 1.25rem;
    color: var(--color-gray-500);
    font-weight: 500;
  }

  .page-description {
    color: var(--color-gray-500);
    margin-top: 0.25rem;
  }

  .header-actions {
    display: flex;
    gap: 0.5rem;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  :global(.form-card) {
    padding: 1.5rem;
  }

  :global(.form-card.full-width) {
    grid-column: 1 / -1;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .form-group {
    margin-bottom: 1rem;
  }

  .form-group:last-child {
    margin-bottom: 0;
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
  }

  .client-preview {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.75rem;
    padding: 0.75rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
  }

  .client-email {
    font-size: 0.875rem;
    color: var(--color-gray-600);
  }

  .client-phone {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .line-items-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .line-items-header .section-title {
    margin: 0;
    padding: 0;
    border: none;
  }

  .line-items-table {
    overflow-x: auto;
  }

  .line-items-table table {
    width: 100%;
    border-collapse: collapse;
  }

  .line-items-table th {
    text-align: left;
    padding: 0.75rem;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--color-gray-500);
    border-bottom: 1px solid var(--color-gray-200);
  }

  .line-items-table td {
    padding: 0.5rem;
    vertical-align: top;
  }

  .description-col {
    width: 40%;
  }

  .quantity-col,
  .price-col {
    width: 20%;
  }

  .total-col {
    width: 15%;
  }

  .actions-col {
    width: 5%;
  }

  .line-total {
    font-weight: 600;
    color: var(--color-gray-900);
    padding-top: 0.75rem;
  }

  .totals-card {
    grid-column: 2;
  }

  .totals-section {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .totals-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0;
  }

  .totals-row:not(:last-child) {
    border-bottom: 1px solid var(--color-gray-100);
  }

  .totals-row .label {
    color: var(--color-gray-600);
  }

  .totals-row .value {
    font-weight: 500;
    color: var(--color-gray-900);
  }

  .totals-row .value.total {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-primary-600);
  }

  .tax-row {
    align-items: flex-start;
  }

  .tax-input {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .tax-input label {
    font-size: 0.875rem;
    color: var(--color-gray-600);
  }

  .actions-card :global(.section-title) {
    display: none;
  }

  .actions-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .action-info {
    display: flex;
    align-items: center;
  }

  .info-text {
    color: var(--color-gray-500);
    font-size: 0.875rem;
    margin: 0;
  }

  .action-buttons {
    display: flex;
    gap: 0.75rem;
  }

  .readonly-notice {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    padding: 1rem;
    color: var(--color-gray-600);
    flex-wrap: wrap;
  }

  .readonly-notice svg {
    color: var(--color-yellow-500);
  }

  @media (max-width: 1024px) {
    .form-grid {
      grid-template-columns: 1fr;
    }

    :global(.form-card.full-width) {
      grid-column: 1;
    }

    .totals-card {
      grid-column: 1;
    }

    .form-row {
      grid-template-columns: 1fr;
    }

    .actions-row {
      flex-direction: column;
      align-items: stretch;
    }

    .action-buttons {
      justify-content: flex-end;
    }
  }

  @media (max-width: 640px) {
    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .header-actions {
      width: 100%;
      justify-content: flex-end;
    }

    .header-title {
      flex-direction: column;
      align-items: flex-start;
    }

    .line-items-table {
      font-size: 0.875rem;
    }

    .line-items-table th,
    .line-items-table td {
      padding: 0.25rem;
    }

    .action-buttons {
      flex-direction: column;
    }

    .action-buttons :global(button) {
      width: 100%;
    }

    .readonly-notice {
      flex-direction: column;
      text-align: center;
    }
  }
</style>
