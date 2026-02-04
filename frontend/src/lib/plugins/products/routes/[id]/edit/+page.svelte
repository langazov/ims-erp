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

  // Form fields
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
    { value: 'cm', label: 'Centimeter' },
    { value: 'ft', label: 'Foot' },
    { value: 'in', label: 'Inch' },
    { value: 'box', label: 'Box' },
    { value: 'pack', label: 'Pack' },
    { value: 'set', label: 'Set' }
  ];

  onMount(async () => {
    try {
      product = await getProductById(productId);
      populateForm(product);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load product';
    } finally {
      loading = false;
    }
  });

  function populateForm(productData: Product) {
    sku = productData.sku;
    name = productData.name;
    description = productData.description;
    category = productData.category;
    price = productData.price;
    cost = productData.cost;
    stockQuantity = productData.stockQuantity.toString();
    lowStockThreshold = productData.lowStockThreshold.toString();
    unit = productData.unit;
    status = productData.status;
  }

  function validateForm(): boolean {
    errors = {};

    if (!sku.trim()) {
      errors.sku = 'SKU is required';
    } else if (sku.length < 3) {
      errors.sku = 'SKU must be at least 3 characters';
    }

    if (!name.trim()) {
      errors.name = 'Product name is required';
    } else if (name.length < 2) {
      errors.name = 'Product name must be at least 2 characters';
    }

    if (!category) {
      errors.category = 'Category is required';
    }

    if (!price.trim()) {
      errors.price = 'Price is required';
    } else {
      const priceNum = parseFloat(price);
      if (isNaN(priceNum) || priceNum < 0) {
        errors.price = 'Price must be a positive number';
      }
    }

    if (cost.trim()) {
      const costNum = parseFloat(cost);
      if (isNaN(costNum) || costNum < 0) {
        errors.cost = 'Cost must be a positive number';
      }
    }

    if (stockQuantity.trim()) {
      const qty = parseInt(stockQuantity, 10);
      if (isNaN(qty) || qty < 0) {
        errors.stockQuantity = 'Stock quantity must be a non-negative integer';
      }
    }

    if (lowStockThreshold.trim()) {
      const threshold = parseInt(lowStockThreshold, 10);
      if (isNaN(threshold) || threshold < 0) {
        errors.lowStockThreshold = 'Low stock threshold must be a non-negative integer';
      }
    }

    return Object.keys(errors).length === 0;
  }

  async function handleSubmit() {
    if (!validateForm()) return;

    const data: Partial<{
      name: string;
      description: string;
      category: string;
      price: number;
      cost: number;
      stockQuantity: number;
      lowStockThreshold: number;
      unit: string;
      status: ProductStatus;
    }> = {};

    if (name !== product?.name) data.name = name;
    if (description !== product?.description) data.description = description;
    if (category !== product?.category) data.category = category;
    if (parseFloat(price) !== parseFloat(product?.price || '0')) data.price = parseFloat(price);
    if (parseFloat(cost) !== parseFloat(product?.cost || '0')) data.cost = parseFloat(cost);
    if (parseInt(stockQuantity, 10) !== product?.stockQuantity) data.stockQuantity = parseInt(stockQuantity, 10);
    if (parseInt(lowStockThreshold, 10) !== product?.lowStockThreshold) data.lowStockThreshold = parseInt(lowStockThreshold, 10);
    if (unit !== product?.unit) data.unit = unit;
    if (status !== product?.status) data.status = status;

    if (Object.keys(data).length === 0) {
      goto(`/products/${productId}`);
      return;
    }

    saving = true;
    try {
      await updateProduct(productId, data);
      goto(`/products/${productId}`);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to update product';
    } finally {
      saving = false;
    }
  }

  function handleCancel() {
    goto(`/products/${productId}`);
  }
</script>

<svelte:head>
  <title>Edit Product | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Edit Product</h1>
      <p class="page-description">Update product information and settings</p>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading product...</p>
    </div>
  {:else if product}
    <Card>
      <form on:submit|preventDefault={handleSubmit}>
        <div class="form-section">
          <h2 class="section-title">Basic Information</h2>
          <div class="form-grid">
            <div class="form-item">
              <Input
                id="sku"
                label="SKU"
                type="text"
                bind:value={sku}
                readonly
                helper="SKU cannot be changed"
              />
            </div>
            <div class="form-item full-width">
              <Input
                id="name"
                label="Product Name"
                type="text"
                placeholder="Enter product name"
                bind:value={name}
                required
                error={errors.name}
              />
            </div>
            <div class="form-item full-width">
              <Textarea
                id="description"
                label="Description"
                placeholder="Enter product description"
                bind:value={description}
                rows={4}
                maxLength={2000}
              />
            </div>
            <div class="form-item">
              <Select
                id="category"
                label="Category"
                options={categoryOptions}
                bind:value={category}
                required
                error={errors.category}
              />
            </div>
            <div class="form-item">
              <Select
                id="status"
                label="Status"
                options={statusOptions}
                bind:value={status}
                required
              />
            </div>
          </div>
        </div>

        <div class="form-section">
          <h2 class="section-title">Pricing</h2>
          <div class="form-grid">
            <div class="form-item">
              <Input
                id="price"
                label="Price"
                type="number"
                placeholder="0.00"
                bind:value={price}
                required
                min="0"
                step="0.01"
                error={errors.price}
              />
            </div>
            <div class="form-item">
              <Input
                id="cost"
                label="Cost"
                type="number"
                placeholder="0.00"
                bind:value={cost}
                min="0"
                step="0.01"
                error={errors.cost}
              />
            </div>
          </div>
        </div>

        <div class="form-section">
          <h2 class="section-title">Inventory</h2>
          <div class="form-grid">
            <div class="form-item">
              <Input
                id="stockQuantity"
                label="Stock Quantity"
                type="number"
                placeholder="0"
                bind:value={stockQuantity}
                min="0"
                step="1"
                error={errors.stockQuantity}
              />
            </div>
            <div class="form-item">
              <Input
                id="lowStockThreshold"
                label="Low Stock Threshold"
                type="number"
                placeholder="10"
                bind:value={lowStockThreshold}
                min="0"
                step="1"
                error={errors.lowStockThreshold}
              />
            </div>
            <div class="form-item">
              <Select
                id="unit"
                label="Unit of Measure"
                options={unitOptions}
                bind:value={unit}
                required
              />
            </div>
          </div>
        </div>

        <div class="form-actions">
          <Button variant="secondary" on:click={handleCancel} disabled={saving}>
            Cancel
          </Button>
          <Button variant="primary" type="submit" loading={saving}>
            {saving ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </form>
    </Card>
  {:else}
    <Alert variant="error">Product not found</Alert>
  {/if}
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 900px;
    margin: 0 auto;
  }

  .page-header {
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

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .form-section {
    margin-bottom: 2rem;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin-bottom: 1rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .form-item {
    min-width: 0;
  }

  .form-item.full-width {
    grid-column: 1 / -1;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--color-gray-200);
    margin-top: 2rem;
  }

  @media (max-width: 640px) {
    .form-grid {
      grid-template-columns: 1fr;
    }

    .form-item.full-width {
      grid-column: 1;
    }
  }
</style>
</style>
