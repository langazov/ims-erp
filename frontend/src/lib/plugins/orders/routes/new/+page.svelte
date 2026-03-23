<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Textarea from '$lib/shared/components/forms/Textarea.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import { getClients, type Client } from '$lib/shared/api/clients';
  import { createOrder } from '$lib/shared/api/orders';

  interface LineItemDraft {
    productName: string;
    quantityStr: string;
    unitPriceStr: string;
  }

  let clients: Client[] = [];
  let clientsLoading = true;

  let selectedClientId = '';
  let shippingStreet = '';
  let shippingCity = '';
  let shippingState = '';
  let shippingPostalCode = '';
  let shippingCountry = '';
  let notes = '';

  let lineItems: LineItemDraft[] = [{ productName: '', quantityStr: '1', unitPriceStr: '0' }];

  let submitting = false;
  let error: string | null = null;

  $: clientOptions = [
    { value: '', label: 'Select a client...' },
    ...clients.map((c) => ({ value: c.id, label: c.name }))
  ];

  $: subtotal = lineItems.reduce((sum, item) => {
    const qty = parseFloat(item.quantityStr) || 0;
    const price = parseFloat(item.unitPriceStr) || 0;
    return sum + qty * price;
  }, 0);

  function formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
  }

  function addLineItem() {
    lineItems = [...lineItems, { productName: '', quantityStr: '1', unitPriceStr: '0' }];
  }

  function removeLineItem(index: number) {
    lineItems = lineItems.filter((_, i) => i !== index);
  }

  function validate(): string | null {
    if (!selectedClientId) return 'Please select a client';
    if (lineItems.length === 0) return 'Please add at least one item';
    for (let i = 0; i < lineItems.length; i++) {
      const item = lineItems[i];
      const qty = parseFloat(item.quantityStr);
      const price = parseFloat(item.unitPriceStr);
      if (!item.productName.trim()) return `Item ${i + 1}: product name is required`;
      if (isNaN(qty) || qty <= 0) return `Item ${i + 1}: quantity must be greater than 0`;
      if (isNaN(price) || price <= 0) return `Item ${i + 1}: unit price must be greater than 0`;
    }
    return null;
  }

  async function handleSubmit() {
    const validationError = validate();
    if (validationError) {
      error = validationError;
      return;
    }

    submitting = true;
    error = null;
    try {
      const items = lineItems.map((item) => {
        const qty = parseFloat(item.quantityStr);
        const price = parseFloat(item.unitPriceStr);
        return {
          productName: item.productName.trim(),
          quantity: qty,
          unitPrice: price.toFixed(2),
          total: (qty * price).toFixed(2)
        };
      });

      const shippingAddress =
        shippingStreet || shippingCity
          ? {
              street: shippingStreet,
              city: shippingCity,
              state: shippingState,
              postalCode: shippingPostalCode,
              country: shippingCountry
            }
          : undefined;

      const newOrder = await createOrder({
        clientId: selectedClientId,
        items,
        shippingAddress,
        notes: notes.trim() || undefined
      });

      goto(`/orders/${newOrder.id}`);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create order';
    } finally {
      submitting = false;
    }
  }

  onMount(async () => {
    try {
      const response = await getClients({ pageSize: 500 });
      clients = response.data;
    } catch {
      // non-fatal
    } finally {
      clientsLoading = false;
    }
  });
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <Button variant="ghost" size="sm" on:click={() => goto('/orders')}>← Back to Orders</Button>
      <h1 class="page-title">New Order</h1>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => (error = null)}>{error}</Alert>
  {/if}

  <div class="form-grid">
    <div class="main-col">
      <!-- Client Selection -->
      <Card>
        <h2 class="section-title">Customer</h2>
        {#if clientsLoading}
          <div class="inline-spinner"><Spinner size="lg" /></div>
        {:else}
          <Select
            id="client"
            label="Client *"
            options={clientOptions}
            bind:value={selectedClientId}
          />
        {/if}
      </Card>

      <!-- Line Items -->
      <Card>
        <div class="section-header">
          <h2 class="section-title">Order Items</h2>
          <Button variant="secondary" size="sm" on:click={addLineItem}>+ Add Item</Button>
        </div>

        <div class="line-items">
          <div class="line-items-header">
            <span class="col-product">Product</span>
            <span class="col-qty">Qty</span>
            <span class="col-price">Unit Price</span>
            <span class="col-total">Total</span>
            <span class="col-action"></span>
          </div>

          {#each lineItems as item, i (i)}
            <div class="line-item-row">
              <div class="col-product">
                <Input
                  id="product-{i}"
                  label=""
                  type="text"
                  placeholder="Product name"
                  bind:value={item.productName}
                />
              </div>
              <div class="col-qty">
                <Input
                  id="qty-{i}"
                  label=""
                  type="number"
                  placeholder="1"
                  bind:value={item.quantityStr}
                />
              </div>
              <div class="col-price">
                <Input
                  id="price-{i}"
                  label=""
                  type="number"
                  placeholder="0.00"
                  bind:value={item.unitPriceStr}
                />
              </div>
              <div class="col-total line-total">
                {formatCurrency((parseFloat(item.quantityStr) || 0) * (parseFloat(item.unitPriceStr) || 0))}
              </div>
              <div class="col-action">
                {#if lineItems.length > 1}
                  <button
                    class="remove-btn"
                    type="button"
                    on:click={() => removeLineItem(i)}
                    aria-label="Remove item"
                  >
                    ✕
                  </button>
                {/if}
              </div>
            </div>
          {/each}

          <div class="subtotal-row">
            <span class="subtotal-label">Subtotal</span>
            <span class="subtotal-value">{formatCurrency(subtotal)}</span>
          </div>
        </div>
      </Card>

      <!-- Notes -->
      <Card>
        <h2 class="section-title">Order Notes</h2>
        <Textarea
          id="notes"
          label=""
          placeholder="Add any special instructions or notes..."
          bind:value={notes}
          rows={3}
        />
      </Card>
    </div>

    <div class="side-col">
      <!-- Shipping Address -->
      <Card>
        <h2 class="section-title">Shipping Address</h2>
        <div class="address-fields">
          <Input id="street" label="Street" type="text" placeholder="123 Main St" bind:value={shippingStreet} />
          <div class="two-col">
            <Input id="city" label="City" type="text" placeholder="City" bind:value={shippingCity} />
            <Input id="state" label="State" type="text" placeholder="State" bind:value={shippingState} />
          </div>
          <div class="two-col">
            <Input id="postal" label="Postal Code" type="text" placeholder="12345" bind:value={shippingPostalCode} />
            <Input id="country" label="Country" type="text" placeholder="US" bind:value={shippingCountry} />
          </div>
        </div>
      </Card>

      <!-- Summary & Submit -->
      <Card>
        <h2 class="section-title">Order Summary</h2>
        <div class="summary-rows">
          <div class="summary-row">
            <span>Items</span>
            <span>{lineItems.length}</span>
          </div>
          <div class="summary-row total-row">
            <span>Subtotal</span>
            <span>{formatCurrency(subtotal)}</span>
          </div>
        </div>
        <div class="form-actions">
          <Button
            variant="primary"
            on:click={handleSubmit}
            loading={submitting}
            disabled={submitting}
          >
            Create Order
          </Button>
          <Button variant="ghost" on:click={() => goto('/orders')} disabled={submitting}>
            Cancel
          </Button>
        </div>
      </Card>
    </div>
  </div>
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1100px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .page-header {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .page-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-gray-900, #111827);
    margin: 0.25rem 0 0;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 300px;
    gap: 1.5rem;
    align-items: start;
  }

  @media (max-width: 768px) {
    .form-grid {
      grid-template-columns: 1fr;
    }
  }

  .main-col, .side-col {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .side-col {
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

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--color-gray-100, #f3f4f6);
  }

  .section-header .section-title {
    margin: 0;
    padding: 0;
    border: none;
  }

  .inline-spinner {
    display: flex;
    justify-content: center;
    padding: 1rem;
  }

  .line-items {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .line-items-header {
    display: grid;
    grid-template-columns: 1fr 80px 100px 100px 36px;
    gap: 0.5rem;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--color-gray-500, #6b7280);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding: 0 0 0.25rem;
  }

  .line-item-row {
    display: grid;
    grid-template-columns: 1fr 80px 100px 100px 36px;
    gap: 0.5rem;
    align-items: center;
  }

  .col-product { flex: 1; }
  .col-qty { width: 80px; }
  .col-price { width: 100px; }

  .line-total {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-800, #1f2937);
    display: flex;
    align-items: center;
    justify-content: flex-end;
  }

  .col-action {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .remove-btn {
    width: 28px;
    height: 28px;
    border: none;
    background: none;
    cursor: pointer;
    color: var(--color-gray-400, #9ca3af);
    font-size: 0.875rem;
    border-radius: 0.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.15s, background-color 0.15s;
  }

  .remove-btn:hover {
    color: #dc2626;
    background-color: #fee2e2;
  }

  .subtotal-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 0 0;
    margin-top: 0.5rem;
    border-top: 1px solid var(--color-gray-100, #f3f4f6);
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-900, #111827);
  }

  .address-fields {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .two-col {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
  }

  .summary-rows {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .summary-row {
    display: flex;
    justify-content: space-between;
    font-size: 0.875rem;
    color: var(--color-gray-600, #4b5563);
  }

  .total-row {
    font-weight: 700;
    font-size: 1rem;
    color: var(--color-gray-900, #111827);
    padding-top: 0.5rem;
    border-top: 1px solid var(--color-gray-100, #f3f4f6);
  }

  .form-actions {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
</style>
