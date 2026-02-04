<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import { getProductById } from '$lib/shared/api/products';
  import type { Product } from '$lib/shared/api/products';

  const productId = $page.params.id;

  interface ProductVariant {
    id: string;
    sku: string;
    name: string;
    attributes: Record<string, string>;
    price: string;
    cost: string;
    stockQuantity: number;
    status: 'active' | 'inactive';
    createdAt: string;
  }

  let product: Product | null = null;
  let variants: ProductVariant[] = [];
  let loading = true;
  let error: string | null = null;
  let showCreateModal = false;
  let showEditModal = false;
  let showDeleteModal = false;
  let saving = false;
  let selectedVariant: ProductVariant | null = null;

  // Form fields
  let variantSku = '';
  let variantName = '';
  let variantPrice = '';
  let variantCost = '';
  let variantStock = '';
  let variantStatus: 'active' | 'inactive' = 'active';
  let variantAttributes: { key: string; value: string }[] = [{ key: '', value: '' }];

  const columns = [
    { key: 'sku', label: 'SKU', sortable: true },
    { key: 'name', label: 'Name', sortable: true },
    { key: 'attributes', label: 'Attributes', sortable: false },
    { key: 'price', label: 'Price', sortable: true },
    { key: 'stock', label: 'Stock', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  async function loadData() {
    loading = true;
    error = null;
    
    try {
      product = await getProductById(productId);
      // Load variants (mock data for now)
      await new Promise(resolve => setTimeout(resolve, 300));
      variants = [
        {
          id: 'var-1',
          sku: `${product.sku}-RED-L`,
          name: `${product.name} - Red, Large`,
          attributes: { color: 'Red', size: 'Large' },
          price: product.price,
          cost: product.cost,
          stockQuantity: 25,
          status: 'active',
          createdAt: product.createdAt
        },
        {
          id: 'var-2',
          sku: `${product.sku}-RED-M`,
          name: `${product.name} - Red, Medium`,
          attributes: { color: 'Red', size: 'Medium' },
          price: product.price,
          cost: product.cost,
          stockQuantity: 30,
          status: 'active',
          createdAt: product.createdAt
        },
        {
          id: 'var-3',
          sku: `${product.sku}-BLUE-L`,
          name: `${product.name} - Blue, Large`,
          attributes: { color: 'Blue', size: 'Large' },
          price: product.price,
          cost: product.cost,
          stockQuantity: 15,
          status: 'active',
          createdAt: product.createdAt
        }
      ];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load product variants';
    } finally {
      loading = false;
    }
  }

  function formatCurrency(value: string): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(parseFloat(value) || 0);
  }

  function getStatusVariant(status: string): 'green' | 'gray' {
    return status === 'active' ? 'green' : 'gray';
  }

  function formatAttributes(attrs: Record<string, string>): string {
    return Object.entries(attrs)
      .map(([key, value]) => `${key}: ${value}`)
      .join(', ');
  }

  function handleBack() {
    goto(`/products/${productId}`);
  }

  function openCreateModal() {
    variantSku = '';
    variantName = '';
    variantPrice = product?.price || '';
    variantCost = product?.cost || '';
    variantStock = '';
    variantStatus = 'active';
    variantAttributes = [{ key: '', value: '' }];
    showCreateModal = true;
  }

  function openEditModal(variant: ProductVariant) {
    selectedVariant = variant;
    variantSku = variant.sku;
    variantName = variant.name;
    variantPrice = variant.price;
    variantCost = variant.cost;
    variantStock = variant.stockQuantity.toString();
    variantStatus = variant.status;
    variantAttributes = Object.entries(variant.attributes).map(([key, value]) => ({ key, value }));
    showEditModal = true;
  }

  function openDeleteModal(variant: ProductVariant) {
    selectedVariant = variant;
    showDeleteModal = true;
  }

  function addAttribute() {
    variantAttributes = [...variantAttributes, { key: '', value: '' }];
  }

  function removeAttribute(index: number) {
    variantAttributes = variantAttributes.filter((_, i) => i !== index);
  }

  async function handleCreateVariant() {
    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      const newVariant: ProductVariant = {
        id: `var-${Date.now()}`,
        sku: variantSku,
        name: variantName,
        attributes: Object.fromEntries(
          variantAttributes.filter(a => a.key && a.value).map(a => [a.key, a.value])
        ),
        price: variantPrice,
        cost: variantCost,
        stockQuantity: parseInt(variantStock, 10) || 0,
        status: variantStatus,
        createdAt: new Date().toISOString()
      };
      variants = [...variants, newVariant];
      showCreateModal = false;
    } catch (err) {
      error = 'Failed to create variant';
    } finally {
      saving = false;
    }
  }

  async function handleUpdateVariant() {
    if (!selectedVariant) return;
    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      variants = variants.map(v => 
        v.id === selectedVariant?.id 
          ? {
              ...v,
              sku: variantSku,
              name: variantName,
              attributes: Object.fromEntries(
                variantAttributes.filter(a => a.key && a.value).map(a => [a.key, a.value])
              ),
              price: variantPrice,
              cost: variantCost,
              stockQuantity: parseInt(variantStock, 10) || 0,
              status: variantStatus
            }
          : v
      );
      showEditModal = false;
      selectedVariant = null;
    } catch (err) {
      error = 'Failed to update variant';
    } finally {
      saving = false;
    }
  }

  async function handleDeleteVariant() {
    if (!selectedVariant) return;
    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      variants = variants.filter(v => v.id !== selectedVariant?.id);
      showDeleteModal = false;
      selectedVariant = null;
    } catch (err) {
      error = 'Failed to delete variant';
    } finally {
      saving = false;
    }
  }

  onMount(() => {
    loadData();
  });
</script>

<svelte:head>
  <title>{product ? `${product.name} - Variants` : 'Product Variants'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading variants...</p>
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
        <div class="header-nav">
          <Button variant="ghost" size="sm" on:click={handleBack}>
            ← Back to Product
          </Button>
        </div>
        <h1 class="page-title">{product.name}</h1>
        <p class="page-description">Manage product variants and options</p>
      </div>
      <div class="header-actions">
        <Button variant="primary" on:click={openCreateModal}>
          Add Variant
        </Button>
      </div>
    </div>

    {#if variants.length === 0}
      <Card>
        <div class="empty-container">
          <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
          </svg>
          <p class="text-gray-500 mb-4">No variants found for this product</p>
          <Button variant="primary" on:click={openCreateModal}>
            Create First Variant
          </Button>
        </div>
      </Card>
    {:else}
      <Card>
        <Table {columns}>
          <tbody>
            {#each variants as variant}
              <tr>
                <td class="font-medium">{variant.sku}</td>
                <td>{variant.name}</td>
                <td>
                  <span class="text-sm text-gray-600">
                    {formatAttributes(variant.attributes)}
                  </span>
                </td>
                <td>{formatCurrency(variant.price)}</td>
                <td>
                  <span class={variant.stockQuantity <= 5 ? 'text-red-600 font-medium' : ''}>
                    {variant.stockQuantity}
                  </span>
                </td>
                <td>
                  <Badge variant={getStatusVariant(variant.status)}>
                    {variant.status}
                  </Badge>
                </td>
                <td>
                  <div class="actions-cell">
                    <Button variant="ghost" size="sm" on:click={() => openEditModal(variant)}>
                      Edit
                    </Button>
                    <Button variant="ghost" size="sm" on:click={() => openDeleteModal(variant)}>
                      Delete
                    </Button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </Table>
      </Card>
    {/if}
  {:else}
    <Alert variant="error">Product not found</Alert>
  {/if}
</div>

<!-- Create Variant Modal -->
<Modal
  bind:open={showCreateModal}
  title="Create Variant"
  size="lg"
>
  <div class="modal-form">
    <div class="form-row">
      <Input
        id="variantSku"
        label="SKU"
        type="text"
        placeholder="Enter variant SKU"
        bind:value={variantSku}
        required
      />
    </div>
    <div class="form-row">
      <Input
        id="variantName"
        label="Variant Name"
        type="text"
        placeholder="e.g., Red, Large"
        bind:value={variantName}
        required
      />
    </div>
    <div class="form-row two-col">
      <Input
        id="variantPrice"
        label="Price"
        type="number"
        placeholder="0.00"
        bind:value={variantPrice}
        min="0"
        step="0.01"
      />
      <Input
        id="variantCost"
        label="Cost"
        type="number"
        placeholder="0.00"
        bind:value={variantCost}
        min="0"
        step="0.01"
      />
    </div>
    <div class="form-row two-col">
      <Input
        id="variantStock"
        label="Stock Quantity"
        type="number"
        placeholder="0"
        bind:value={variantStock}
        min="0"
        step="1"
      />
      <Select
        id="variantStatus"
        label="Status"
        options={[
          { value: 'active', label: 'Active' },
          { value: 'inactive', label: 'Inactive' }
        ]}
        bind:value={variantStatus}
      />
    </div>
    <div class="form-row">
      <label class="attributes-label">Attributes</label>
      {#each variantAttributes as attr, index}
        <div class="attribute-row">
          <Input
            id="attr-key-{index}"
            type="text"
            placeholder="e.g., Color"
            bind:value={attr.key}
          />
          <Input
            id="attr-value-{index}"
            type="text"
            placeholder="e.g., Red"
            bind:value={attr.value}
          />
          <Button 
            variant="ghost" 
            size="sm" 
            on:click={() => removeAttribute(index)}
            disabled={variantAttributes.length === 1}
          >
            Remove
          </Button>
        </div>
      {/each}
      <Button variant="secondary" size="sm" on:click={addAttribute} class="mt-2">
        + Add Attribute
      </Button>
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={saving}>
      Cancel
    </Button>
    <Button variant="primary" on:click={handleCreateVariant} loading={saving}>
      {saving ? 'Creating...' : 'Create Variant'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Edit Variant Modal -->
<Modal
  bind:open={showEditModal}
  title="Edit Variant"
  size="lg"
>
  <div class="modal-form">
    <div class="form-row">
      <Input
        id="editVariantSku"
        label="SKU"
        type="text"
        bind:value={variantSku}
        required
      />
    </div>
    <div class="form-row">
      <Input
        id="editVariantName"
        label="Variant Name"
        type="text"
        bind:value={variantName}
        required
      />
    </div>
    <div class="form-row two-col">
      <Input
        id="editVariantPrice"
        label="Price"
        type="number"
        bind:value={variantPrice}
        min="0"
        step="0.01"
      />
      <Input
        id="editVariantCost"
        label="Cost"
        type="number"
        bind:value={variantCost}
        min="0"
        step="0.01"
      />
    </div>
    <div class="form-row two-col">
      <Input
        id="editVariantStock"
        label="Stock Quantity"
        type="number"
        bind:value={variantStock}
        min="0"
        step="1"
      />
      <Select
        id="editVariantStatus"
        label="Status"
        options={[
          { value: 'active', label: 'Active' },
          { value: 'inactive', label: 'Inactive' }
        ]}
        bind:value={variantStatus}
      />
    </div>
    <div class="form-row">
      <label class="attributes-label">Attributes</label>
      {#each variantAttributes as attr, index}
        <div class="attribute-row">
          <Input
            id="edit-attr-key-{index}"
            type="text"
            placeholder="e.g., Color"
            bind:value={attr.key}
          />
          <Input
            id="edit-attr-value-{index}"
            type="text"
            placeholder="e.g., Red"
            bind:value={attr.value}
          />
          <Button 
            variant="ghost" 
            size="sm" 
            on:click={() => removeAttribute(index)}
            disabled={variantAttributes.length === 1}
          >
            Remove
          </Button>
        </div>
      {/each}
      <Button variant="secondary" size="sm" on:click={addAttribute} class="mt-2">
        + Add Attribute
      </Button>
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={saving}>
      Cancel
    </Button>
    <Button variant="primary" on:click={handleUpdateVariant} loading={saving}>
      {saving ? 'Saving...' : 'Save Changes'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Delete Variant Modal -->
<Modal
  bind:open={showDeleteModal}
  title="Delete Variant"
  size="sm"
>
  <p>Are you sure you want to delete the variant <strong>{selectedVariant?.name}</strong>? This action cannot be undone.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showDeleteModal = false; }} disabled={saving}>
      Cancel
    </Button>
    <Button variant="danger" on:click={handleDeleteVariant} loading={saving}>
      {saving ? 'Deleting...' : 'Delete'}
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

  .header-content {
    flex: 1;
  }

  .header-nav {
    margin-bottom: 0.5rem;
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

  .empty-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    color: var(--color-gray-500);
  }

  .actions-cell {
    display: flex;
    gap: 0.5rem;
  }

  .modal-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .form-row {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .form-row.two-col {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .attributes-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-700);
  }

  .attribute-row {
    display: grid;
    grid-template-columns: 1fr 1fr auto;
    gap: 0.5rem;
    align-items: flex-start;
  }

  @media (max-width: 640px) {
    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .form-row.two-col {
      grid-template-columns: 1fr;
    }

    .attribute-row {
      grid-template-columns: 1fr;
    }
  }
</style>
</style>
