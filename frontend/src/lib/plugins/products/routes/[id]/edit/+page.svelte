<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Textarea from '$lib/shared/components/forms/Textarea.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import { getProductById, updateProduct } from '$lib/shared/api/products';
  import type { Product, ProductStatus } from '$lib/shared/api/products';

  const productId = $page.params.id;

  let product: Product | null = null;
  let loading = true;
  let saving = false;
  let error: string | null = null;
  let errors: Record<string, string> = {};

  let sku = '';
  let name = '';
  let description = '';
  let category = '';
  let price = '';
  let cost = '';
  let stockQuantity = '';
  let lowStockThreshold = '';
  let unit = '';
  let status: ProductStatus = 'active';

  const statusOptions = [
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'discontinued', label: 'Discontinued' }
  ];

  const categoryOptions = [
    { value: 'electronics', label: 'Electronics' },
    { value: 'clothing', label: 'Clothing' },
    { value: 'food', label: 'Food & Beverage' },
    { value: 'home', label: 'Home & Garden' },
    { value: 'sports', label: 'Sports & Outdoors' },
    { value: 'books', label: 'Books & Media' },
    { value: 'toys', label: 'Toys & Games' },
    { value: 'health', label: 'Health & Beauty' },
    { value: 'automotive', label: 'Automotive' },
    { value: 'office', label: 'Office Supplies' },
    { value: 'other', label: 'Other' }
  ];

  const unitOptions = [
    { value: 'piece', label: 'Piece' },
    { value: 'kg', label: 'Kilogram' },
    { value: 'g', label: 'Gram' },
    { value: 'lb', label: 'Pound' },
    { value: 'oz', label: 'Ounce' },
    { value: 'l', label: 'Liter' },
    { value: 'ml', label: 'Milliliter' },
    { value: 'm', label: 'Meter' },
    { value: 'box', label: 'Box' },
    { value: 'pack', label: 'Pack' },
    { value: 'set', label: 'Set' }
  ];

  onMount(async () => {
    try {
      product = await getProductById(productId);
      sku = product.sku;
      name = product.name;
      description = product.description;
      category = product.category;
      price = product.price;
      cost = product.cost;
      stockQuantity = product.stockQuantity.toString();
      lowStockThreshold = product.lowStockThreshold.toString();
      unit = product.unit;
      status = product.status;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load product';
    } finally {
      loading = false;
    }
  });

  function validate(): boolean {
    errors = {};
    if (!sku.trim()) errors.sku = 'SKU is required';
    if (!name.trim()) errors.name = 'Product name is required';
    if (!category) errors.category = 'Category is required';
    if (!price.trim() || isNaN(parseFloat(price)) || parseFloat(price) < 0)
      errors.price = 'Valid price is required';
    if (cost.trim() && (isNaN(parseFloat(cost)) || parseFloat(cost) < 0))
      errors.cost = 'Cost must be a positive number';
    if (stockQuantity.trim() && (isNaN(parseInt(stockQuantity)) || parseInt(stockQuantity) < 0))
      errors.stockQuantity = 'Stock quantity must be non-negative';
    if (lowStockThreshold.trim() && (isNaN(parseInt(lowStockThreshold)) || parseInt(lowStockThreshold) < 0))
      errors.lowStockThreshold = 'Threshold must be non-negative';
    return Object.keys(errors).length === 0;
  }

  async function handleSubmit() {
    if (!validate()) return;
    saving = true;
    error = null;
    try {
      await updateProduct(productId, {
        name,
        description,
        category,
        price: parseFloat(price),
        cost: cost ? parseFloat(cost) : undefined,
        stockQuantity: stockQuantity ? parseInt(stockQuantity) : undefined,
        lowStockThreshold: lowStockThreshold ? parseInt(lowStockThreshold) : undefined,
        unit: unit || undefined,
        status
      });
      goto(`/products/${productId}`);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to update product';
    } finally {
      saving = false;
    }
  }
</script>

<svelte:head>
  <title>Edit Product | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <Button variant="secondary" on:click={() => goto(`/products/${productId}`)}>← Back</Button>
    <h1 class="page-title">Edit Product</h1>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => (error = null)} class="mb-4">{error}</Alert>
  {/if}

  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <span>Loading product...</span>
    </div>
  {:else if product}
    <Card>
      <form on:submit|preventDefault={handleSubmit}>
        <div class="form-section">
          <h2 class="section-title">Basic Information</h2>
          <div class="form-grid">
            <Input id="sku" label="SKU" bind:value={sku} readonly error={errors.sku} helpText="SKU cannot be changed" />
            <Input id="name" label="Product Name" bind:value={name} required error={errors.name} />
            <div class="full-width">
              <Textarea id="description" label="Description" bind:value={description} rows={3} />
            </div>
            <Select id="category" label="Category" options={categoryOptions} bind:value={category} required error={errors.category} />
            <Select id="status" label="Status" options={statusOptions} bind:value={status} required />
          </div>
        </div>

        <div class="form-section">
          <h2 class="section-title">Pricing</h2>
          <div class="form-grid">
            <Input id="price" label="Price ($)" type="number" placeholder="0.00" bind:value={price} min="0" step="0.01" required error={errors.price} />
            <Input id="cost" label="Cost ($)" type="number" placeholder="0.00" bind:value={cost} min="0" step="0.01" error={errors.cost} />
          </div>
        </div>

        <div class="form-section">
          <h2 class="section-title">Inventory</h2>
          <div class="form-grid">
            <Input id="stockQuantity" label="Stock Quantity" type="number" placeholder="0" bind:value={stockQuantity} min="0" step="1" error={errors.stockQuantity} />
            <Input id="lowStockThreshold" label="Low Stock Threshold" type="number" placeholder="10" bind:value={lowStockThreshold} min="0" step="1" error={errors.lowStockThreshold} />
            <Select id="unit" label="Unit of Measure" options={unitOptions} bind:value={unit} />
          </div>
        </div>

        <div class="form-actions">
          <Button variant="secondary" on:click={() => goto(`/products/${productId}`)} disabled={saving}>Cancel</Button>
          <Button variant="primary" type="submit" loading={saving}>{saving ? 'Saving...' : 'Save Changes'}</Button>
        </div>
      </form>
    </Card>
  {:else}
    <Alert variant="error">Product not found</Alert>
  {/if}
</div>

<style>
  .page-container { padding: 1.5rem; max-width: 860px; margin: 0 auto; }
  .page-header { display: flex; align-items: center; gap: 1rem; margin-bottom: 1.5rem; }
  .page-title { font-size: 1.5rem; font-weight: 700; color: var(--color-gray-900); margin: 0; }
  .loading-container { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 3rem; gap: 1rem; color: var(--color-gray-500); }
  .form-section { margin-bottom: 2rem; }
  .section-title { font-size: 1rem; font-weight: 600; color: var(--color-gray-800); margin-bottom: 1rem; padding-bottom: 0.5rem; border-bottom: 1px solid var(--color-gray-200); }
  .form-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem; }
  .full-width { grid-column: 1 / -1; }
  .form-actions { display: flex; justify-content: flex-end; gap: 0.75rem; padding-top: 1.5rem; border-top: 1px solid var(--color-gray-200); }
  @media (max-width: 640px) { .form-grid { grid-template-columns: 1fr; } }
</style>
