<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { getProductById, deleteProduct } from '$lib/shared/api/products';
  import type { Product } from '$lib/shared/api/products';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Progress from '$lib/shared/components/display/Progress.svelte';

  let product: Product | null = null;
  let loading = true;
  let error = '';
  let showDeleteConfirm = false;
  let deleting = false;

  $: id = $page.params.id;

  function statusVariant(status: string): 'green' | 'gray' | 'red' {
    if (status === 'active') return 'green';
    if (status === 'discontinued') return 'red';
    return 'gray';
  }

  function formatCurrency(val: string | number) {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(Number(val) || 0);
  }

  function stockPercent(qty: number, threshold: number) {
    if (!threshold) return 100;
    return Math.min(100, (qty / (threshold * 3)) * 100);
  }

  function stockVariant(qty: number, threshold: number): 'success' | 'warning' | 'danger' | 'default' {
    if (qty === 0) return 'danger';
    if (qty <= threshold) return 'warning';
    return 'success';
  }

  async function handleDelete() {
    deleting = true;
    try {
      await deleteProduct(id);
      window.location.href = '/products';
    } catch (e: any) {
      error = e.message || 'Delete failed';
      deleting = false;
    }
  }

  onMount(async () => {
    try {
      product = await getProductById(id);
    } catch (e: any) {
      error = e.message || 'Failed to load product';
    } finally {
      loading = false;
    }
  });
</script>

<div class="page-container">
  <div class="page-header">
    <Button variant="secondary" on:click={() => window.location.href = '/products'}>← Products</Button>
    {#if product}
      <div class="header-actions">
        <Button variant="secondary" on:click={() => window.location.href = `/products/${id}/edit`}>Edit</Button>
        <Button variant="danger" on:click={() => showDeleteConfirm = true}>Delete</Button>
      </div>
    {/if}
  </div>

  {#if loading}
    <div class="loading-container"><Spinner size="lg" /><span>Loading product...</span></div>
  {:else if error}
    <div class="error-container"><p class="error-message">{error}</p></div>
  {:else if product}
    <div class="content-grid">
      <Card>
        <div class="product-header">
          <div>
            <span class="sku-tag">{product.sku}</span>
            <h2 class="product-name">{product.name}</h2>
            {#if product.description}
              <p class="product-desc">{product.description}</p>
            {/if}
          </div>
          <Badge variant={statusVariant(product.status)}>{product.status}</Badge>
        </div>
        <div class="detail-grid">
          <div class="detail-item"><span class="detail-label">Category</span><span class="detail-value">{product.category || '—'}</span></div>
          <div class="detail-item"><span class="detail-label">Unit</span><span class="detail-value">{product.unit || '—'}</span></div>
          <div class="detail-item"><span class="detail-label">Price</span><span class="detail-value price">{formatCurrency(product.price)}</span></div>
          <div class="detail-item"><span class="detail-label">Cost</span><span class="detail-value">{formatCurrency(product.cost)}</span></div>
          {#if Number(product.price) > 0 && Number(product.cost) > 0}
            <div class="detail-item">
              <span class="detail-label">Margin</span>
              <span class="detail-value">{(((Number(product.price) - Number(product.cost)) / Number(product.price)) * 100).toFixed(1)}%</span>
            </div>
          {/if}
          <div class="detail-item"><span class="detail-label">Created</span><span class="detail-value">{new Date(product.createdAt).toLocaleDateString()}</span></div>
        </div>
      </Card>

      <Card title="Stock Level">
        <div class="stock-section">
          <div class="stock-numbers">
            <div class="stock-main">
              <span class="stock-qty">{product.stockQuantity}</span>
              <span class="stock-unit">{product.unit || 'units'}</span>
            </div>
            <div class="stock-threshold">Reorder at: {product.lowStockThreshold}</div>
          </div>
          <Progress
            value={product.stockQuantity}
            max={Math.max(product.stockQuantity, product.lowStockThreshold * 3)}
            size="lg"
            variant={stockVariant(product.stockQuantity, product.lowStockThreshold)}
            showLabel
          />
          {#if product.stockQuantity === 0}
            <Badge variant="red">Out of Stock</Badge>
          {:else if product.stockQuantity <= product.lowStockThreshold}
            <Badge variant="yellow">Low Stock</Badge>
          {:else}
            <Badge variant="green">In Stock</Badge>
          {/if}
        </div>
      </Card>
    </div>
  {/if}
</div>

{#if showDeleteConfirm}
  <div class="modal-overlay">
    <div class="modal">
      <h3>Delete Product</h3>
      <p>Are you sure you want to delete <strong>{product?.name}</strong>? This action cannot be undone.</p>
      <div class="modal-actions">
        <Button variant="secondary" on:click={() => showDeleteConfirm = false}>Cancel</Button>
        <Button variant="danger" disabled={deleting} on:click={handleDelete}>{deleting ? 'Deleting...' : 'Delete'}</Button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page-container { padding: 1.5rem; max-width: 900px; margin: 0 auto; }
  .page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
  .header-actions { display: flex; gap: 0.75rem; }
  .loading-container, .error-container { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 3rem; gap: 1rem; }
  .error-message { color: var(--color-red-600); }
  .content-grid { display: flex; flex-direction: column; gap: 1.5rem; }
  .product-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem; }
  .sku-tag { font-family: monospace; font-size: 0.75rem; background: var(--color-gray-100); padding: 0.2rem 0.5rem; border-radius: 0.25rem; color: var(--color-gray-600); }
  .product-name { font-size: 1.5rem; font-weight: 700; margin: 0.5rem 0 0; }
  .product-desc { color: var(--color-gray-500); margin-top: 0.5rem; }
  .detail-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; }
  .detail-item { display: flex; flex-direction: column; gap: 0.25rem; }
  .detail-label { font-size: 0.75rem; text-transform: uppercase; color: var(--color-gray-500); font-weight: 500; }
  .detail-value { font-size: 0.9rem; color: var(--color-gray-800); font-weight: 500; }
  .price { font-size: 1.125rem; color: var(--color-green-700); font-weight: 700; }
  .stock-section { display: flex; flex-direction: column; gap: 1rem; }
  .stock-numbers { display: flex; align-items: baseline; justify-content: space-between; }
  .stock-main { display: flex; align-items: baseline; gap: 0.5rem; }
  .stock-qty { font-size: 2.5rem; font-weight: 700; color: var(--color-gray-900); }
  .stock-unit { font-size: 1rem; color: var(--color-gray-500); }
  .stock-threshold { font-size: 0.875rem; color: var(--color-gray-500); }
  .modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 50; }
  .modal { background: white; border-radius: 0.5rem; padding: 1.5rem; max-width: 400px; width: 90%; }
  .modal h3 { margin: 0 0 0.75rem; font-size: 1.125rem; font-weight: 600; }
  .modal-actions { display: flex; gap: 0.75rem; justify-content: flex-end; margin-top: 1.5rem; }
</style>
