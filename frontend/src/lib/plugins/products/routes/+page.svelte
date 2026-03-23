<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import { getProducts, deleteProduct, getProductStats } from '$lib/shared/api/products';
  import type { Product, ProductStatus } from '$lib/shared/api/products';

  let products: Product[] = [];
  let total = 0;
  let totalPages = 1;
  let currentPage = 1;
  let loading = true;
  let error = '';
  let searchQuery = '';
  let categoryFilter = '';
  let statusFilter: ProductStatus | '' = '';
  let stats: { total: number; active: number; lowStock: number; outOfStock: number } | null = null;
  let deleteConfirmId = '';
  let deleteConfirmName = '';
  let deleting = false;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'discontinued', label: 'Discontinued' }
  ];

  function statusVariant(status: ProductStatus): 'green' | 'gray' | 'red' {
    if (status === 'active') return 'green';
    if (status === 'discontinued') return 'red';
    return 'gray';
  }

  function formatCurrency(val: string | number) {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(Number(val) || 0);
  }

  async function load() {
    loading = true;
    error = '';
    try {
      const res = await getProducts({
        search: searchQuery || undefined,
        category: categoryFilter || undefined,
        status: statusFilter || undefined,
        page: currentPage,
        pageSize: 20
      });
      products = res.data;
      total = res.total;
      totalPages = res.totalPages;
    } catch (e: any) {
      error = e.message || 'Failed to load products';
    } finally {
      loading = false;
    }
  }

  async function loadStats() {
    try { stats = await getProductStats(); } catch {}
  }

  async function confirmDelete(id: string, name: string) {
    deleteConfirmId = id;
    deleteConfirmName = name;
  }

  async function doDelete() {
    deleting = true;
    try {
      await deleteProduct(deleteConfirmId);
      deleteConfirmId = '';
      await load();
      await loadStats();
    } catch (e: any) {
      error = e.message || 'Delete failed';
    } finally {
      deleting = false;
    }
  }

  onMount(() => { load(); loadStats(); });
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Products</h1>
      <p class="page-description">Manage your product catalog</p>
    </div>
    <Button variant="primary" on:click={() => window.location.href = '/products/new'}>+ Add Product</Button>
  </div>

  {#if stats}
    <div class="stats-grid">
      <Card><div class="stat"><span class="stat-label">Total</span><span class="stat-value">{stats.total}</span></div></Card>
      <Card><div class="stat"><span class="stat-label">Active</span><span class="stat-value text-green">{stats.active}</span></div></Card>
      <Card><div class="stat"><span class="stat-label">Low Stock</span><span class="stat-value text-yellow">{stats.lowStock}</span></div></Card>
      <Card><div class="stat"><span class="stat-label">Out of Stock</span><span class="stat-value text-red">{stats.outOfStock}</span></div></Card>
    </div>
  {/if}

  <Card>
    <div class="filters">
      <div class="filter-row">
        <div class="filter-item search-filter">
          <Input placeholder="Search by name or SKU..." bind:value={searchQuery} on:input={load} />
        </div>
        <div class="filter-item">
          <Input placeholder="Filter by category..." bind:value={categoryFilter} on:input={load} />
        </div>
        <div class="filter-item">
          <Select options={statusOptions} bind:value={statusFilter} on:change={load} />
        </div>
      </div>
    </div>

    {#if loading}
      <div class="loading-container"><Spinner size="lg" /><span>Loading products...</span></div>
    {:else if error}
      <div class="error-container"><p class="error-message">{error}</p><Button variant="secondary" on:click={load}>Retry</Button></div>
    {:else if products.length === 0}
      <div class="empty-container">
        <p>No products found.</p>
        <Button variant="primary" on:click={() => window.location.href = '/products/new'}>Add First Product</Button>
      </div>
    {:else}
      <table class="data-table">
        <thead>
          <tr>
            <th>SKU</th><th>Name</th><th>Category</th><th>Price</th><th>Cost</th><th>Stock</th><th>Status</th><th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each products as p}
            <tr class="clickable-row" on:click={() => window.location.href = `/products/${p.id}`}>
              <td class="mono">{p.sku}</td>
              <td class="font-medium">{p.name}</td>
              <td>{p.category || '—'}</td>
              <td>{formatCurrency(p.price)}</td>
              <td>{formatCurrency(p.cost)}</td>
              <td>
                {#if p.stockQuantity <= p.lowStockThreshold && p.stockQuantity > 0}
                  <Badge variant="yellow">{p.stockQuantity} {p.unit || 'units'}</Badge>
                {:else if p.stockQuantity === 0}
                  <Badge variant="red">Out of stock</Badge>
                {:else}
                  {p.stockQuantity} {p.unit || 'units'}
                {/if}
              </td>
              <td><Badge variant={statusVariant(p.status)}>{p.status}</Badge></td>
              <td>
                <div class="actions-cell" on:click|stopPropagation>
                  <Button size="sm" variant="secondary" on:click={() => window.location.href = `/products/${p.id}/edit`}>Edit</Button>
                  <Button size="sm" variant="danger" on:click={() => confirmDelete(p.id, p.name)}>Delete</Button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>

      {#if totalPages > 1}
        <div class="pagination">
          <Button variant="secondary" disabled={currentPage === 1} on:click={() => { currentPage--; load(); }}>Previous</Button>
          <span class="pagination-info">Page {currentPage} of {totalPages} ({total} total)</span>
          <Button variant="secondary" disabled={currentPage === totalPages} on:click={() => { currentPage++; load(); }}>Next</Button>
        </div>
      {/if}
    {/if}
  </Card>
</div>

{#if deleteConfirmId}
  <div class="modal-overlay">
    <div class="modal">
      <h3>Delete Product</h3>
      <p>Are you sure you want to delete <strong>{deleteConfirmName}</strong>?</p>
      <div class="modal-actions">
        <Button variant="secondary" on:click={() => deleteConfirmId = ''}>Cancel</Button>
        <Button variant="danger" disabled={deleting} on:click={doDelete}>{deleting ? 'Deleting...' : 'Delete'}</Button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page-container { padding: 1.5rem; max-width: 1400px; margin: 0 auto; }
  .page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem; }
  .page-title { font-size: 1.875rem; font-weight: 700; color: var(--color-gray-900); margin: 0; }
  .page-description { color: var(--color-gray-500); margin-top: 0.25rem; }
  .stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; margin-bottom: 1.5rem; }
  .stat { display: flex; flex-direction: column; gap: 0.25rem; padding: 0.5rem; }
  .stat-label { font-size: 0.75rem; color: var(--color-gray-500); text-transform: uppercase; }
  .stat-value { font-size: 1.5rem; font-weight: 700; }
  .text-green { color: var(--color-green-600); }
  .text-yellow { color: var(--color-yellow-600); }
  .text-red { color: var(--color-red-600); }
  .filters { margin-bottom: 1rem; }
  .filter-row { display: flex; gap: 0.75rem; flex-wrap: wrap; }
  .filter-item { flex: 1; min-width: 180px; }
  .search-filter { flex: 2; min-width: 280px; }
  .data-table { width: 100%; border-collapse: collapse; }
  .data-table th { text-align: left; padding: 0.75rem 1rem; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; color: var(--color-gray-500); border-bottom: 1px solid var(--color-gray-200); }
  .data-table td { padding: 0.75rem 1rem; border-bottom: 1px solid var(--color-gray-100); font-size: 0.875rem; }
  .clickable-row { cursor: pointer; }
  .clickable-row:hover { background-color: var(--color-gray-50); }
  .mono { font-family: monospace; font-size: 0.8rem; }
  .font-medium { font-weight: 500; }
  .actions-cell { display: flex; gap: 0.5rem; }
  .loading-container, .error-container, .empty-container { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 3rem; gap: 1rem; color: var(--color-gray-500); }
  .error-message { color: var(--color-red-600); }
  .pagination { display: flex; align-items: center; justify-content: center; gap: 1rem; margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--color-gray-200); }
  .pagination-info { color: var(--color-gray-600); font-size: 0.875rem; }
  .modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 50; }
  .modal { background: white; border-radius: 0.5rem; padding: 1.5rem; max-width: 400px; width: 90%; }
  .modal h3 { margin: 0 0 0.75rem; font-size: 1.125rem; font-weight: 600; }
  .modal-actions { display: flex; gap: 0.75rem; justify-content: flex-end; margin-top: 1.5rem; }
</style>
