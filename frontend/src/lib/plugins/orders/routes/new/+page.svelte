<script lang="ts">
  import { onMount } from 'svelte';
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
  import { createOrder, type OrderItem, type Address } from '$lib/shared/api/orders';
  import { getClients, type Client } from '$lib/shared/api/clients';
  import { getProducts, type Product } from '$lib/shared/api/products';

  // Form state
  let selectedClientId = '';
  let orderNumber = '';
  let orderDate = new Date().toISOString().split('T')[0];
  let notes = '';
  let taxRate = '8';
  let saveAsDraft = true;

  // Shipping address
  let shippingAddress: Address = {
    street: '',
    city: '',
    state: '',
    postalCode: '',
    country: ''
  };
  let useClientAddress = false;

  // Line items
  let lineItems: Array<{
    id: string;
    productId: string;
    productName: string;
    quantity: string;
    unitPrice: string;
  }> = [];

  // Data loading
  let clients: Client[] = [];
  let products: Product[] = [];
  let loadingClients = true;
  let loadingProducts = true;
  let submitting = false;
  let error: string | null = null;
  let errors: Record<string, string> = {};

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

  $: productOptions = products.map(product => ({
    value: product.id,
    label: `${product.name} - ${formatCurrency(parseFloat(product.price))}`
  }));

  $: selectedClient = clients.find(c => c.id === selectedClientId);

  onMount(async () => {
    await Promise.all([loadClients(), loadProducts()]);
    generateOrderNumber();
  });

  async function loadClients() {
    loadingClients = true;
    try {
      const response = await getClients();
      clients = response.data;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load clients';
    } finally {
      loadingClients = false;
    }
  }

  async function loadProducts() {
    loadingProducts = true;
    try {
      const response = await getProducts();
      products = response.data;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load products';
    } finally {
      loadingProducts = false;
    }
  }

  function generateOrderNumber() {
    const date = new Date();
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const random = Math.floor(Math.random() * 1000).toString().padStart(3, '0');
    orderNumber = `ORD-${year}-${month}-${day}-${random}`;
  }

  function addLineItem() {
    lineItems = [
      ...lineItems,
      { id: crypto.randomUUID(), productId: '', productName: '', quantity: '1', unitPrice: '' }
    ];
  }

  function removeLineItem(index: number) {
    if (lineItems.length > 0) {
      lineItems = lineItems.filter((_, i) => i !== index);
    }
  }

  function updateLineItemProduct(index: number, productId: string) {
    const product = products.find(p => p.id === productId);
    lineItems = lineItems.map((item, i) => {
      if (i === index) {
        return {
          ...item,
          productId,
          productName: product?.name || '',
          unitPrice: product?.price || ''
        };
      }
      return item;
    });
  }

  function updateLineItemQuantity(index: number, quantity: string) {
    lineItems = lineItems.map((item, i) => {
      if (i === index) {
        return { ...item, quantity };
      }
      return item;
    });
  }

  function getLineTotal(item: typeof lineItems[0]): number {
    const price = parseFloat(item.unitPrice) || 0;
    const qty = parseInt(item.quantity) || 0;
    return qty * price;
  }

  function handleUseClientAddress() {
    useClientAddress = !useClientAddress;
    if (useClientAddress && selectedClient?.billingAddress) {
      shippingAddress = { ...selectedClient.billingAddress };
    } else if (!useClientAddress) {
      shippingAddress = {
        street: '',
        city: '',
        state: '',
        postalCode: '',
        country: ''
      };
    }
  }

  function validateForm(): boolean {
    errors = {};

    if (!selectedClientId) {
      errors.clientId = 'Please select a client';
    }

    if (!orderNumber.trim()) {
      errors.orderNumber = 'Order number is required';
    }

    if (!orderDate) {
      errors.orderDate = 'Order date is required';
    }

    // Validate shipping address
    if (!shippingAddress.street.trim()) {
      errors.shippingStreet = 'Street is required';
    }
    if (!shippingAddress.city.trim()) {
      errors.shippingCity = 'City is required';
    }
    if (!shippingAddress.state.trim()) {
      errors.shippingState = 'State is required';
    }
    if (!shippingAddress.postalCode.trim()) {
      errors.shippingPostalCode = 'Postal code is required';
    }
    if (!shippingAddress.country.trim()) {
      errors.shippingCountry = 'Country is required';
    }

    // Validate line items
    if (lineItems.length === 0) {
      errors.lineItems = 'At least one line item is required';
    } else {
      let hasLineItemErrors = false;
      lineItems.forEach((item, index) => {
        if (!item.productId) {
          errors[`item_${index}_product`] = 'Product is required';
          hasLineItemErrors = true;
        }
        if (!item.quantity || parseInt(item.quantity) <= 0) {
          errors[`item_${index}_quantity`] = 'Quantity must be greater than 0';
          hasLineItemErrors = true;
        }
      });

      if (hasLineItemErrors) {
        errors.lineItems = 'Please fix line item errors';
      }
    }

    return Object.keys(errors).length === 0;
  }

  async function handleSubmit() {
    if (!validateForm()) {
      return;
    }

    const items: OrderItem[] = lineItems.map(item => ({
      productId: item.productId,
      productName: item.productName,
      quantity: parseInt(item.quantity) || 1,
      unitPrice: item.unitPrice,
      total: getLineTotal(item).toFixed(2)
    }));

    const data = {
      clientId: selectedClientId,
      items,
      shippingAddress,
      notes: notes || undefined
    };

    submitting = true;
    error = null;

    try {
      const order = await createOrder(data);
      goto(`/orders/${order.id}`);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to create order';
      submitting = false;
    }
  }

  function handleCancel() {
    goto('/orders');
  }

  function formatCurrency(value: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(value);
  }
</script>

<svelte:head>
  <title>New Order | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">New Order</h1>
      <p class="page-description">Create a new order for your client</p>
    </div>
    <div class="header-actions">
      <Button variant="secondary" on:click={handleCancel} disabled={submitting}>
        Cancel
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  {#if loadingClients || loadingProducts}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading...</p>
    </div>
  {:else}
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
            />
            {#if selectedClient}
              <div class="client-preview">
                <Badge variant="blue" size="sm">{selectedClient.code}</Badge>
                <span class="client-email">{selectedClient.email}</span>
                {#if selectedClient.phone}
                  <span class="client-phone">{selectedClient.phone}</span>
                {/if}
              </div>
            {/if}
          </div>
        </Card>

        <!-- Order Details -->
        <Card class="form-card">
          <h2 class="section-title">Order Details</h2>
          <div class="form-row">
            <div class="form-group">
              <Input
                id="orderNumber"
                label="Order Number"
                type="text"
                bind:value={orderNumber}
                required
                error={errors.orderNumber}
                readonly
              />
            </div>
            <div class="form-group">
              <DatePicker
                id="orderDate"
                label="Order Date"
                bind:value={orderDate}
                required
                error={errors.orderDate}
              />
            </div>
          </div>
        </Card>

        <!-- Line Items -->
        <Card class="form-card full-width">
          <div class="line-items-header">
            <h2 class="section-title">Line Items</h2>
            <Button variant="secondary" size="sm" on:click={addLineItem} type="button">
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              Add Item
            </Button>
          </div>

          {#if errors.lineItems}
            <Alert variant="error" class="mb-4">{errors.lineItems}</Alert>
          {/if}

          {#if lineItems.length === 0}
            <div class="empty-line-items">
              <p class="text-gray-500">No items added yet. Click "Add Item" to add products to this order.</p>
            </div>
          {:else}
            <div class="line-items-table">
              <table>
                <thead>
                  <tr>
                    <th class="product-col">Product</th>
                    <th class="quantity-col">Quantity</th>
                    <th class="price-col">Unit Price</th>
                    <th class="total-col">Total</th>
                    <th class="actions-col"></th>
                  </tr>
                </thead>
                <tbody>
                  {#each lineItems as item, index (item.id)}
                    <tr>
                      <td>
                        <Select
                          id="item-product-{index}"
                          label=""
                          options={productOptions}
                          bind:value={item.productId}
                          placeholder="Select a product"
                          error={errors[`item_${index}_product`]}
                          on:change={(e) => updateLineItemProduct(index, e.detail)}
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
                        />
                      </td>
                      <td class="price-cell">
                        <Input
                          id="item-price-{index}"
                          label=""
                          type="number"
                          placeholder="0.00"
                          bind:value={item.unitPrice}
                          min="0"
                          step="0.01"
                          readonly
                        />
                      </td>
                      <td class="line-total">
                        {formatCurrency(getLineTotal(item))}
                      </td>
                      <td>
                        <Button
                          variant="ghost"
                          size="sm"
                          on:click={() => removeLineItem(index)}
                          type="button"
                        >
                          <svg class="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </Button>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        </Card>

        <!-- Shipping Address -->
        <Card class="form-card full-width">
          <div class="shipping-header">
            <h2 class="section-title">Shipping Address</h2>
            {#if selectedClient?.billingAddress}
              <label class="checkbox-option">
                <input
                  type="checkbox"
                  bind:checked={useClientAddress}
                  on:change={handleUseClientAddress}
                />
                <span class="checkbox-label">Use client's billing address</span>
              </label>
            {/if}
          </div>
          <div class="address-grid">
            <div class="form-group full-width">
              <Input
                id="shippingStreet"
                label="Street Address"
                type="text"
                bind:value={shippingAddress.street}
                placeholder="123 Main Street"
                required
                error={errors.shippingStreet}
              />
            </div>
            <div class="form-group">
              <Input
                id="shippingCity"
                label="City"
                type="text"
                bind:value={shippingAddress.city}
                placeholder="New York"
                required
                error={errors.shippingCity}
              />
            </div>
            <div class="form-group">
              <Input
                id="shippingState"
                label="State/Province"
                type="text"
                bind:value={shippingAddress.state}
                placeholder="NY"
                required
                error={errors.shippingState}
              />
            </div>
            <div class="form-group">
              <Input
                id="shippingPostalCode"
                label="Postal Code"
                type="text"
                bind:value={shippingAddress.postalCode}
                placeholder="10001"
                required
                error={errors.shippingPostalCode}
              />
            </div>
            <div class="form-group">
              <Input
                id="shippingCountry"
                label="Country"
                type="text"
                bind:value={shippingAddress.country}
                placeholder="United States"
                required
                error={errors.shippingCountry}
              />
            </div>
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
            placeholder="Add any additional notes or special instructions..."
            rows={4}
          />
        </Card>

        <!-- Actions -->
        <Card class="form-card full-width actions-card">
          <div class="actions-row">
            <div class="save-options">
              <label class="radio-option">
                <input
                  type="radio"
                  name="saveOption"
                  value="draft"
                  bind:group={saveAsDraft}
                  checked={saveAsDraft}
                />
                <span class="radio-label">
                  <span class="radio-title">Save as Draft</span>
                  <span class="radio-description">Order will be saved as pending</span>
                </span>
              </label>
              <label class="radio-option">
                <input
                  type="radio"
                  name="saveOption"
                  value="confirmed"
                  bind:group={saveAsDraft}
                  checked={!saveAsDraft}
                />
                <span class="radio-label">
                  <span class="radio-title">Confirm Immediately</span>
                  <span class="radio-description">Order will be confirmed and processed</span>
                </span>
              </label>
            </div>
            <div class="action-buttons">
              <Button variant="secondary" on:click={handleCancel} disabled={submitting}>
                Cancel
              </Button>
              <Button variant="primary" type="submit" loading={submitting}>
                {submitting ? 'Creating...' : (saveAsDraft ? 'Save as Draft' : 'Create & Confirm')}
              </Button>
            </div>
          </div>
        </Card>
      </div>
    </form>
  {/if}
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

  .form-group.full-width {
    grid-column: 1 / -1;
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
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

  .empty-line-items {
    padding: 2rem;
    text-align: center;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
    border: 2px dashed var(--color-gray-200);
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

  .product-col {
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

  .price-cell :global(input) {
    background-color: var(--color-gray-100);
  }

  .line-total {
    font-weight: 600;
    color: var(--color-gray-900);
    padding-top: 0.75rem;
  }

  .shipping-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .shipping-header .section-title {
    margin: 0;
    padding: 0;
    border: none;
  }

  .checkbox-option {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-size: 0.875rem;
    color: var(--color-gray-700);
  }

  .checkbox-option input[type="checkbox"] {
    width: 1rem;
    height: 1rem;
    accent-color: var(--color-primary-600);
  }

  .address-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
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

  .save-options {
    display: flex;
    gap: 1.5rem;
  }

  .radio-option {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    cursor: pointer;
  }

  .radio-option input[type="radio"] {
    margin-top: 0.25rem;
    width: 1rem;
    height: 1rem;
    accent-color: var(--color-primary-600);
  }

  .radio-label {
    display: flex;
    flex-direction: column;
  }

  .radio-title {
    font-weight: 500;
    color: var(--color-gray-900);
  }

  .radio-description {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .action-buttons {
    display: flex;
    gap: 0.75rem;
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

    .address-grid {
      grid-template-columns: 1fr;
    }

    .actions-row {
      flex-direction: column;
      align-items: stretch;
    }

    .save-options {
      flex-direction: column;
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

    .shipping-header {
      flex-direction: column;
      gap: 0.5rem;
      align-items: flex-start;
    }
  }
</style>
