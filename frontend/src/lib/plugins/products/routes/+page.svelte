<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import Pagination from '$lib/shared/components/data/Pagination.svelte';

  interface Product {
    id: string;
    sku: string;
    name: string;
    description: string;
    category: string;
    type: 'physical' | 'digital' | 'service';
    status: 'active' | 'inactive' | 'draft';
    listPrice: number;
    salePrice: number;
    costPrice: number;
    stockQuantity: number;
    lowStockThreshold: number;
    images: string[];
    createdAt: string;
  }

  let products: Product[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter: string = '';
  let categoryFilter: string = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;
  let showCreateModal = false;
  let deleteProductId: string | null = null;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'draft', label: 'Draft' }
  ];

  const categoryOptions = [
    { value: '', label: 'All Categories' },
    { value: 'electronics', label: 'Electronics' },
    { value: 'clothing', label: 'Clothing' },
    { value: 'food', label: 'Food & Beverage' },
    { value: 'home', label: 'Home & Garden' }
  ];

  const columns = [
    { key: 'sku', label: 'SKU', sortable: true },
    { key: 'name', label: 'Name', sortable: true },
    { key: 'category', label: 'Category', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'price', label: 'Price', sortable: true },
    { key: 'stock', label: 'Stock', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  function getStatusVariant(status: string): 'green' | 'gray' | 'yellow' | 'red' {
    switch (status) {
      case 'active': return 'green';
      case 'draft': return 'yellow';
      case 'inactive': return 'gray';
      default: return 'gray';
    }
  }

  function formatCurrency(value: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(value);
  }

  async function loadProducts() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      products = [
        {
          id: '1',
          sku: 'PROD-001',
          name: 'Wireless Bluetooth Headphones',
          description: 'High-quality wireless headphones with noise cancellation',
          category: 'electronics',
          type: 'physical',
          status: 'active',
          listPrice: 199.99,
          salePrice: 179.99,
          costPrice: 89.99,
          stockQuantity: 150,
          lowStockThreshold: 20,
          images: [],
          createdAt: '2024-01-15'
        },
        {
          id: '2',
          sku: 'PROD-002',
          name: 'Cotton T-Shirt',
          description: 'Premium cotton t-shirt in various colors',
          category: 'clothing',
          type: 'physical',
          status: 'active',
          listPrice: 29.99,
          salePrice: 24.99,
          costPrice: 12.99,
          stockQuantity: 500,
          lowStockThreshold: 50,
          images: [],
          createdAt: '2024-02-01'
        },
        {
          id: '3',
          sku: 'PROD-003',
          name: 'Organic Coffee Beans',
          description: 'Single-origin organic coffee beans',
          category: 'food',
          type: 'physical',
          status: 'active',
          listPrice: 24.99,
          salePrice: 19.99,
          costPrice: 10.99,
          stockQuantity: 8,
          lowStockThreshold: 15,
          images: [],
          createdAt: '2024-02-10'
        }
      ];
      
      totalItems = products.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load products';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadProducts();
  }

  function handleRowClick(product: Product) {
    window.location.href = `/products/${product.id}`;
  }

  function handleEdit(product: Product, event: Event) {
    event.stopPropagation();
    window.location.href = `/products/${product.id}/edit`;
  }

  async function handleDelete(product: Product, event: Event) {
    event.stopPropagation();
    deleteProductId = product.id;
  }

  async function confirmDelete() {
    if (!deleteProductId) return;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      products = products.filter(p => p.id !== deleteProductId);
      totalItems = products.length;
      deleteProductId = null;
    } catch (err) {
      error = 'Failed to delete product';
    }
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadProducts();
  }

  onMount(() => {
    loadProducts();
  });
</script>

<svelte:head>
  <title>Products | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Products</h1>
      <p class="page-description">Manage your product catalog and inventory</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={() => showCreateModal = true}>
        Add Product
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null}>
      {error}
    </Alert>
  {/if}

  <Card>
    <div class="filters">
      <div class="filter-row">
        <div class="filter-item search-filter">
          <Input
            id="search"
            label="Search"
            type="search"
            placeholder="Search products..."
            bind:value={searchQuery}
            on:keydown={(e) => e.key === 'Enter' && handleSearch()}
          />
        </div>
        <div class="filter-item">
          <Select
            id="status"
            label="Status"
            options={statusOptions}
            bind:value={statusFilter}
            on:change={() => handleSearch()}
          />
        </div>
        <div class="filter-item">
          <Select
            id="category"
            label="Category"
            options={categoryOptions}
            bind:value={categoryFilter}
            on:change={() => handleSearch()}
          />
        </div>
        <div class="filter-item">
          <Button variant="secondary" on:click={handleSearch}>
            Search
          </Button>
        </div>
      </div>
    </div>

    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading products...</p>
      </div>
    {:else if products.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
        <p class="text-gray-500 mb-4">No products found</p>
        <Button variant="primary" on:click={() => showCreateModal = true}>
          Add Your First Product
        </Button>
      </div>
    {:else}
      <Table {columns}>
        <tbody>
          {#each products as product}
            <tr on:click={() => handleRowClick(product)} class="clickable-row">
              <td class="font-medium">{product.sku}</td>
              <td>{product.name}</td>
              <td class="capitalize">{product.category}</td>
              <td>
                <Badge variant={getStatusVariant(product.status)}>
                  {product.status}
                </Badge>
              </td>
              <td>{formatCurrency(product.salePrice)}</td>
              <td>
                <span class={product.stockQuantity <= product.lowStockThreshold ? 'text-red-600 font-medium' : ''}>
                  {product.stockQuantity}
                  {#if product.stockQuantity <= product.lowStockThreshold}
                    <span class="text-xs ml-1">(Low)</span>
                  {/if}
                </span>
              </td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={(e) => handleEdit(product, e)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={(e) => handleDelete(product, e)}>
                    Delete
                  </Button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </Table>

      <Pagination
        {currentPage}
        {totalPages}
        {totalItems}
        {pageSize}
        on:pageChange={(e) => handlePageChange(e.detail)}
      />
    {/if}
  </Card>
</div>

<Modal
  bind:open={showCreateModal}
  title="Create Product"
  size="lg"
>
  <p class="text-gray-600">Product creation form will be implemented here.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close}>Cancel</Button>
    <Button variant="primary" on:click={() => {
      showCreateModal = false;
    }}>Create</Button>
  </svelte:fragment>
</Modal>

{#if deleteProductId}
  <Modal
    open={true}
    title="Delete Product"
    size="sm"
  >
    <p>Are you sure you want to delete this product? This action cannot be undone.</p>
    
    <svelte:fragment slot="footer" let:close>
      <Button variant="secondary" on:click={() => { close(); deleteProductId = null; }}>Cancel</Button>
      <Button variant="danger" on:click={confirmDelete}>Delete</Button>
    </svelte:fragment>
  </Modal>
{/if}

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1400px;
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

  .filters {
    margin-bottom: 1rem;
  }

  .filter-row {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .filter-item {
    flex: 1;
    min-width: 200px;
  }

  .search-filter {
    flex: 2;
    min-width: 300px;
  }

  .loading-container,
  .empty-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .clickable-row {
    cursor: pointer;
  }

  .clickable-row:hover {
    background-color: var(--color-gray-50);
  }

  .actions-cell {
    display: flex;
    gap: 0.5rem;
  }
</style>
