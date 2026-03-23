<script lang="ts">
  import { createProduct } from '$lib/shared/api/products';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';

  let sku = '';
  let name = '';
  let description = '';
  let category = '';
  let unit = 'unit';
  let price = '';
  let cost = '';
  let trackStock = true;
  let stockQuantity = '';
  let minQty = '';
  let maxQty = '';
  let reorderPoint = '';
  let status = 'active';
  let submitting = false;
  let error = '';

  $: margin = price && cost && Number(price) > 0
    ? (((Number(price) - Number(cost)) / Number(price)) * 100).toFixed(1)
    : null;

  const statusOptions = [
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'discontinued', label: 'Discontinued' }
  ];

  const unitOptions = [
    { value: 'unit', label: 'Unit' },
    { value: 'kg', label: 'Kilogram' },
    { value: 'g', label: 'Gram' },
    { value: 'l', label: 'Liter' },
    { value: 'ml', label: 'Milliliter' },
    { value: 'box', label: 'Box' },
    { value: 'piece', label: 'Piece' }
  ];

  async function handleSubmit() {
    if (!sku.trim()) { error = 'SKU is required'; return; }
    if (!name.trim()) { error = 'Name is required'; return; }
    if (!price || isNaN(Number(price))) { error = 'Valid price is required'; return; }
    error = '';
    submitting = true;
    try {
      const product = await createProduct({
        sku: sku.trim(),
        name: name.trim(),
        description: description || undefined,
        category: category || undefined,
        price: Number(price),
        cost: cost ? Number(cost) : undefined,
        stockQuantity: trackStock && stockQuantity ? Number(stockQuantity) : undefined,
        lowStockThreshold: reorderPoint ? Number(reorderPoint) : undefined,
        unit: unit || undefined
      });
      window.location.href = `/products/${product.id}`;
    } catch (e: any) {
      error = e.message || 'Failed to create product';
    } finally {
      submitting = false;
    }
  }
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">New Product</h1>
      <p class="page-description">Add a new product to your catalog</p>
    </div>
    <Button variant="secondary" on:click={() => window.location.href = '/products'}>← Back</Button>
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  <form on:submit|preventDefault={handleSubmit}>
    <div class="form-grid">
      <Card title="Basic Information">
        <div class="field-group">
          <div class="field">
            <label for="sku">SKU <span class="required">*</span></label>
            <Input id="sku" bind:value={sku} placeholder="e.g. PROD-001" required />
          </div>
          <div class="field">
            <label for="name">Name <span class="required">*</span></label>
            <Input id="name" bind:value={name} placeholder="Product name" required />
          </div>
          <div class="field full">
            <label for="desc">Description</label>
            <textarea id="desc" bind:value={description} placeholder="Product description..." class="textarea" rows="3"></textarea>
          </div>
          <div class="field">
            <label for="category">Category</label>
            <Input id="category" bind:value={category} placeholder="e.g. Electronics" />
          </div>
          <div class="field">
            <label for="unit">Unit</label>
            <Select id="unit" options={unitOptions} bind:value={unit} />
          </div>
          <div class="field">
            <label for="status">Status</label>
            <Select id="status" options={statusOptions} bind:value={status} />
          </div>
        </div>
      </Card>

      <Card title="Pricing">
        <div class="field-group">
          <div class="field">
            <label for="price">Price <span class="required">*</span></label>
            <Input id="price" type="number" min="0" step="0.01" bind:value={price} placeholder="0.00" required />
          </div>
          <div class="field">
            <label for="cost">Cost</label>
            <Input id="cost" type="number" min="0" step="0.01" bind:value={cost} placeholder="0.00" />
          </div>
          {#if margin !== null}
            <div class="field full">
              <div class="margin-indicator">
                <span class="margin-label">Margin</span>
                <span class="margin-value">{margin}%</span>
              </div>
            </div>
          {/if}
        </div>
      </Card>

      <Card title="Stock Settings">
        <div class="field-group">
          <div class="field full">
            <label class="checkbox-label">
              <input type="checkbox" bind:checked={trackStock} />
              Track inventory for this product
            </label>
          </div>
          {#if trackStock}
            <div class="field">
              <label for="qty">Initial Quantity</label>
              <Input id="qty" type="number" min="0" bind:value={stockQuantity} placeholder="0" />
            </div>
            <div class="field">
              <label for="reorder">Reorder Point</label>
              <Input id="reorder" type="number" min="0" bind:value={reorderPoint} placeholder="0" />
            </div>
            <div class="field">
              <label for="minqty">Min Quantity</label>
              <Input id="minqty" type="number" min="0" bind:value={minQty} placeholder="0" />
            </div>
            <div class="field">
              <label for="maxqty">Max Quantity</label>
              <Input id="maxqty" type="number" min="0" bind:value={maxQty} placeholder="0" />
            </div>
          {/if}
        </div>
      </Card>
    </div>

    <div class="form-actions">
      <Button variant="secondary" type="button" on:click={() => window.location.href = '/products'}>Cancel</Button>
      <Button variant="primary" type="submit" disabled={submitting}>{submitting ? 'Creating...' : 'Create Product'}</Button>
    </div>
  </form>
</div>

<style>
  .page-container { padding: 1.5rem; max-width: 900px; margin: 0 auto; }
  .page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem; }
  .page-title { font-size: 1.875rem; font-weight: 700; color: var(--color-gray-900); margin: 0; }
  .page-description { color: var(--color-gray-500); margin-top: 0.25rem; }
  .error-banner { background: var(--color-red-50); border: 1px solid var(--color-red-200); color: var(--color-red-700); padding: 0.75rem 1rem; border-radius: 0.375rem; margin-bottom: 1rem; }
  .form-grid { display: flex; flex-direction: column; gap: 1.5rem; }
  .field-group { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  .field { display: flex; flex-direction: column; gap: 0.375rem; }
  .field.full { grid-column: 1 / -1; }
  label { font-size: 0.875rem; font-weight: 500; color: var(--color-gray-700); }
  .required { color: var(--color-red-500); }
  .textarea { width: 100%; padding: 0.5rem 0.75rem; border: 1px solid var(--color-gray-300); border-radius: 0.375rem; font-size: 0.875rem; resize: vertical; }
  .checkbox-label { display: flex; align-items: center; gap: 0.5rem; font-size: 0.875rem; font-weight: 500; cursor: pointer; }
  .margin-indicator { display: flex; align-items: center; justify-content: space-between; padding: 0.5rem 0.75rem; background: var(--color-green-50); border-radius: 0.375rem; }
  .margin-label { font-size: 0.875rem; color: var(--color-gray-600); }
  .margin-value { font-size: 1.125rem; font-weight: 700; color: var(--color-green-700); }
  .form-actions { display: flex; justify-content: flex-end; gap: 0.75rem; margin-top: 1.5rem; }
</style>
