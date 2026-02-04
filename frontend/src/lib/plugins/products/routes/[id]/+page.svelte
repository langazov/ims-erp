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
  import { getProductById, deleteProduct } from '$lib/shared/api/products';
  import type { Product, ProductStatus } from '$lib/shared/api/products';

  const productId = $page.params.id;

  let product: Product | null = null;
  let loading = true;
  let error: string | null = null;
  let showDeleteModal = false;
  let deleting = false;

  async function loadProduct() {
    loading = true;
    error = null;
    
    try {
      product = await getProductById(productId);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load product';
    } finally {
      loading = false;
    }
  }

  function getStatusVariant(status: ProductStatus): 'green' | 'gray' | 'yellow' | 'red' {
    switch (status) {
      case 'active': return 'green';
      case 'inactive': return 'gray';
      case 'discontinued': return 'red';
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

  function getCategoryLabel(category: string): string {
    const labels: Record<string, string> = {
      electronics: 'Electronics',
      clothing: 'Clothing',
      food: 'Food & Beverage',
      home: 'Home & Garden',
      sports: 'Sports & Outdoors',
      books: 'Books & Media',
      toys: 'Toys & Games',
      health: 'Health & Beauty',
      automotive: 'Automotive',
      office: 'Office Supplies',
      other: 'Other'
    };
    return labels[category] || category;
  }

  function getUnitLabel(unit: string): string {
    const labels: Record<string, string> = {
      piece: 'Piece',
      kg: 'Kilogram',
      g: 'Gram',
      lb: 'Pound',
      oz: 'Ounce',
      l: 'Liter',
      ml: 'Milliliter',
      m: 'Meter',
      cm: 'Centimeter',
      ft: 'Foot',
      in: 'Inch',
      box: 'Box',
      pack: 'Pack',
      set: 'Set'
    };
    return labels[unit] || unit;
  }

  function calculateMargin(): number {
    if (!product) return 0;
    const price = parseFloat(product.price) || 0;
    const cost = parseFloat(product.cost) || 0;
    if (price === 0) return 0;
    return ((price - cost) / price) * 100;
  }

  function handleEdit() {
    goto(`/products/${productId}/edit`);
  }

  function handleDelete() {
    showDeleteModal = true;
  }

  async function confirmDelete() {
    deleting = true;
    try {
      await deleteProduct(productId);
      goto('/products');
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete product';
      deleting = false;
      showDeleteModal = false;
    }
  }

  function handleViewVariants() {
    goto(`/products/${productId}/variants`);
  }

  function handleViewPricing() {
    goto(`/products/${productId}/pricing`);
  }

  onMount(() => {
    loadProduct();
  });
</script>

<svelte:head>
  <title>{product ? product.name : 'Product Details'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading product details...</p>
    </div>
  {:else if error}
    <div class="error-container">
      <Alert variant="error">{error}</Alert>
      <Button variant="secondary" on:click={() => goto('/products')}>
        Back to Products
      </Button>
    </div>
  {:else if product}
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <h1 class="page-title">{product.name}</h1>
          <Badge variant={getStatusVariant(product.status)} size="md">
            {product.status}
          </Badge>
        </div>
        <p class="page-description">
          SKU: {product.sku} • {getCategoryLabel(product.category)}
        </p>
      </div>
      <div class="header-actions">
        <Button variant="secondary" on:click={handleEdit}>
          Edit
        </Button>
        <Button variant="danger" on:click={handleDelete}>
          Delete
        </Button>
      </div>
    </div>

    <div class="stats-grid">
      <Card class="stat-card">
        <div class="stat-icon blue">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">Price</span>
          <span class="stat-value">{formatCurrency(product.price)}</span>
        </div>
      </Card>

      <Card class="stat-card">
        <div class="stat-icon green">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">Stock</span>
          <span class="stat-value" class:text-red-600={product.stockQuantity <= product.lowStockThreshold}>
            {product.stockQuantity}
            {#if product.stockQuantity <= product.lowStockThreshold}
              <span class="text-sm font-normal text-red-500">(Low)</span>
            {/if}
          </span>
        </div>
      </Card>

      <Card class="stat-card">
        <div class="stat-icon purple">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">Margin</span>
          <span class="stat-value">{calculateMargin().toFixed(1)}%</span>
        </div>
      </Card>
    </div>

    <div class="details-grid">
      <Card>
        <h2 class="section-title">Product Information</h2>
        <dl class="info-list">
          <div class="info-item">
            <dt>SKU</dt>
            <dd>{product.sku}</dd>
          </div>
          <div class="info-item">
            <dt>Name</dt>
            <dd>{product.name}</dd>
          </div>
          <div class="info-item full-width">
            <dt>Description</dt>
            <dd class="description">{product.description || 'No description available'}</dd>
          </div>
          <div class="info-item">
            <dt>Category</dt>
            <dd>{getCategoryLabel(product.category)}</dd>
          </div>
          <div class="info-item">
            <dt>Status</dt>
            <dd>
              <Badge variant={getStatusVariant(product.status)}>
                {product.status}
              </Badge>
            </dd>
          </div>
          <div class="info-item">
            <dt>Unit</dt>
            <dd>{getUnitLabel(product.unit)}</dd>
          </div>
        </dl>
      </Card>

      <Card>
        <h2 class="section-title">Pricing Details</h2>
        <dl class="info-list">
          <div class="info-item">
            <dt>Price</dt>
            <dd class="price">{formatCurrency(product.price)}</dd>
          </div>
          <div class="info-item">
            <dt>Cost</dt>
            <dd>{formatCurrency(product.cost)}</dd>
          </div>
          <div class="info-item">
            <dt>Margin</dt>
            <dd class={calculateMargin() > 0 ? 'text-green-600' : 'text-red-600'}>
              {calculateMargin().toFixed(1)}%
            </dd>
          </div>
          <div class="info-item">
            <dt>Profit</dt>
            <dd class="text-green-600">
              {formatCurrency((parseFloat(product.price) - parseFloat(product.cost)).toString())}
            </dd>
          </div>
        </dl>
      </Card>

      <Card>
        <h2 class="section-title">Inventory</h2>
        <dl class="info-list">
          <div class="info-item">
            <dt>Stock Quantity</dt>
            <dd class:low-stock={product.stockQuantity <= product.lowStockThreshold}>
              {product.stockQuantity} {getUnitLabel(product.unit)}s
              {#if product.stockQuantity <= product.lowStockThreshold}
                <Badge variant="red" size="sm">Low Stock</Badge>
              {/if}
            </dd>
          </div>
          <div class="info-item">
            <dt>Low Stock Threshold</dt>
            <dd>{product.lowStockThreshold} {getUnitLabel(product.unit)}s</dd>
          </div>
          <div class="info-item">
            <dt>Stock Value</dt>
            <dd>{formatCurrency((parseFloat(product.cost) * product.stockQuantity).toString())}</dd>
          </div>
        </dl>
      </Card>

      <Card>
        <h2 class="section-title">Metadata</h2>
        <dl class="info-list">
          <div class="info-item">
            <dt>Created</dt>
            <dd>{formatDate(product.createdAt)}</dd>
          </div>
          <div class="info-item">
            <dt>Last Updated</dt>
            <dd>{formatDate(product.updatedAt)}</dd>
          </div>
          <div class="info-item">
            <dt>Product ID</dt>
            <dd class="text-xs text-gray-500">{product.id}</dd>
          </div>
        </dl>
      </Card>
    </div>

    <Card class="actions-card">
      <h2 class="section-title">Quick Actions</h2>
      <div class="quick-actions">
        <Button variant="secondary" on:click={handleViewVariants}>
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
          </svg>
          Manage Variants
        </Button>
        <Button variant="secondary" on:click={handleViewPricing}>
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          Pricing History
        </Button>
      </div>
    </Card>
  {:else}
    <Alert variant="error">Product not found</Alert>
  {/if}
</div>

<Modal
  bind:open={showDeleteModal}
  title="Delete Product"
  size="sm"
>
  <p>Are you sure you want to delete <strong>{product?.name}</strong>? This action cannot be undone.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showDeleteModal = false; }} disabled={deleting}>
      Cancel
    </Button>
    <Button variant="danger" on:click={confirmDelete} loading={deleting}>
      {deleting ? 'Deleting...' : 'Delete'}
    </Button>
  </svelte:fragment>
</Modal>

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

  .page-description {
    color: var(--color-gray-500);
    margin-top: 0.25rem;
  }

  .header-actions {
    display: flex;
    gap: 0.5rem;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  :global(.stat-card) {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1.25rem;
  }

  .stat-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 3rem;
    height: 3rem;
    border-radius: 0.75rem;
  }

  .stat-icon.blue {
    background-color: var(--color-blue-100);
    color: var(--color-blue-600);
  }

  .stat-icon.green {
    background-color: var(--color-green-100);
    color: var(--color-green-600);
  }

  .stat-icon.purple {
    background-color: var(--color-purple-100);
    color: var(--color-purple-600);
  }

  .stat-content {
    display: flex;
    flex-direction: column;
  }

  .stat-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .stat-value {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-gray-900);
  }

  .details-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .info-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .info-item {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--color-gray-100);
  }

  .info-item:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }

  .info-item.full-width {
    flex-direction: column;
    gap: 0.25rem;
  }

  .info-item dt {
    color: var(--color-gray-500);
    font-size: 0.875rem;
  }

  .info-item dd {
    color: var(--color-gray-900);
    font-weight: 500;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .info-item dd.description {
    font-weight: normal;
    color: var(--color-gray-600);
    line-height: 1.5;
  }

  .info-item dd.price {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-green-600);
  }

  .low-stock {
    color: var(--color-red-600);
  }

  :global(.actions-card) {
    padding: 1.5rem;
  }

  .quick-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  @media (max-width: 768px) {
    .stats-grid {
      grid-template-columns: 1fr;
    }

    .details-grid {
      grid-template-columns: 1fr;
    }

    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .header-actions {
      width: 100%;
      justify-content: flex-end;
    }
  }
</style>
</style>
