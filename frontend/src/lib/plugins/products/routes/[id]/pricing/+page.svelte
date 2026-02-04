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
  import { getProductById, updateProduct } from '$lib/shared/api/products';
  import type { Product } from '$lib/shared/api/products';

  const productId = $page.params.id;

  interface PriceHistory {
    id: string;
    type: 'base' | 'sale' | 'wholesale' | 'tier';
    price: string;
    oldPrice?: string;
    startDate?: string;
    endDate?: string;
    minQuantity?: number;
    customerGroup?: string;
    createdAt: string;
    createdBy: string;
  }

  interface PricingTier {
    id: string;
    name: string;
    minQuantity: number;
    price: string;
    discount: number;
  }

  let product: Product | null = null;
  let priceHistory: PriceHistory[] = [];
  let pricingTiers: PricingTier[] = [];
  let loading = true;
  let error: string | null = null;
  let saving = false;

  // Modals
  let showUpdatePriceModal = false;
  let showAddTierModal = false;
  let showEditTierModal = false;
  let showDeleteTierModal = false;
  let selectedTier: PricingTier | null = null;

  // Form fields
  let newPrice = '';
  let newCost = '';
  let priceType: 'base' | 'sale' = 'base';
  let saleStartDate = '';
  let saleEndDate = '';

  // Tier form fields
  let tierName = '';
  let tierMinQty = '';
  let tierPrice = '';

  const columns = [
    { key: 'type', label: 'Type', sortable: true },
    { key: 'price', label: 'Price', sortable: true },
    { key: 'oldPrice', label: 'Previous', sortable: true },
    { key: 'period', label: 'Period', sortable: false },
    { key: 'createdAt', label: 'Date', sortable: true },
    { key: 'createdBy', label: 'By', sortable: false }
  ];

  async function loadData() {
    loading = true;
    error = null;
    
    try {
      product = await getProductById(productId);
      // Load price history (mock data for now)
      await new Promise(resolve => setTimeout(resolve, 300));
      priceHistory = [
        {
          id: 'ph-1',
          type: 'base',
          price: product.price,
          oldPrice: (parseFloat(product.price) * 0.9).toFixed(2),
          createdAt: product.updatedAt,
          createdBy: 'Admin User'
        },
        {
          id: 'ph-2',
          type: 'sale',
          price: (parseFloat(product.price) * 0.85).toFixed(2),
          oldPrice: product.price,
          startDate: new Date().toISOString(),
          endDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
          createdAt: new Date(Date.now() - 86400000).toISOString(),
          createdBy: 'Sales Manager'
        }
      ];

      // Load pricing tiers (mock data)
      pricingTiers = [
        {
          id: 'tier-1',
          name: 'Wholesale',
          minQuantity: 10,
          price: (parseFloat(product.price) * 0.8).toFixed(2),
          discount: 20
        },
        {
          id: 'tier-2',
          name: 'Bulk',
          minQuantity: 50,
          price: (parseFloat(product.price) * 0.7).toFixed(2),
          discount: 30
        }
      ];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load pricing data';
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

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function getTypeBadgeVariant(type: string): 'blue' | 'green' | 'purple' | 'yellow' {
    switch (type) {
      case 'base': return 'blue';
      case 'sale': return 'green';
      case 'wholesale': return 'purple';
      case 'tier': return 'yellow';
      default: return 'blue';
    }
  }

  function calculateMargin(): number {
    if (!product) return 0;
    const price = parseFloat(product.price) || 0;
    const cost = parseFloat(product.cost) || 0;
    if (price === 0) return 0;
    return ((price - cost) / price) * 100;
  }

  function handleBack() {
    goto(`/products/${productId}`);
  }

  function openUpdatePriceModal() {
    newPrice = product?.price || '';
    newCost = product?.cost || '';
    priceType = 'base';
    saleStartDate = '';
    saleEndDate = '';
    showUpdatePriceModal = true;
  }

  function openAddTierModal() {
    tierName = '';
    tierMinQty = '';
    tierPrice = '';
    showAddTierModal = true;
  }

  function openEditTierModal(tier: PricingTier) {
    selectedTier = tier;
    tierName = tier.name;
    tierMinQty = tier.minQuantity.toString();
    tierPrice = tier.price;
    showEditTierModal = true;
  }

  function openDeleteTierModal(tier: PricingTier) {
    selectedTier = tier;
    showDeleteTierModal = true;
  }

  async function handleUpdatePrice() {
    if (!product) return;
    saving = true;
    try {
      const updates: Partial<{ price: number; cost: number }> = {};
      if (parseFloat(newPrice) !== parseFloat(product.price)) {
        updates.price = parseFloat(newPrice);
      }
      if (parseFloat(newCost) !== parseFloat(product.cost)) {
        updates.cost = parseFloat(newCost);
      }

      if (Object.keys(updates).length > 0) {
        await updateProduct(productId, updates);
        // Add to history
        priceHistory = [
          {
            id: `ph-${Date.now()}`,
            type: priceType,
            price: newPrice,
            oldPrice: product.price,
            startDate: priceType === 'sale' ? saleStartDate : undefined,
            endDate: priceType === 'sale' ? saleEndDate : undefined,
            createdAt: new Date().toISOString(),
            createdBy: 'Current User'
          },
          ...priceHistory
        ];
        product = { ...product, price: newPrice, cost: newCost };
      }
      showUpdatePriceModal = false;
    } catch (err) {
      error = 'Failed to update price';
    } finally {
      saving = false;
    }
  }

  async function handleAddTier() {
    saving = true;
    try {
      const newTier: PricingTier = {
        id: `tier-${Date.now()}`,
        name: tierName,
        minQuantity: parseInt(tierMinQty, 10),
        price: tierPrice,
        discount: Math.round((1 - parseFloat(tierPrice) / parseFloat(product?.price || '1')) * 100)
      };
      pricingTiers = [...pricingTiers, newTier];
      showAddTierModal = false;
    } catch (err) {
      error = 'Failed to add tier';
    } finally {
      saving = false;
    }
  }

  async function handleUpdateTier() {
    if (!selectedTier) return;
    saving = true;
    try {
      pricingTiers = pricingTiers.map(t =>
        t.id === selectedTier?.id
          ? {
              ...t,
              name: tierName,
              minQuantity: parseInt(tierMinQty, 10),
              price: tierPrice,
              discount: Math.round((1 - parseFloat(tierPrice) / parseFloat(product?.price || '1')) * 100)
            }
          : t
      );
      showEditTierModal = false;
      selectedTier = null;
    } catch (err) {
      error = 'Failed to update tier';
    } finally {
      saving = false;
    }
  }

  async function handleDeleteTier() {
    if (!selectedTier) return;
    saving = true;
    try {
      pricingTiers = pricingTiers.filter(t => t.id !== selectedTier?.id);
      showDeleteTierModal = false;
      selectedTier = null;
    } catch (err) {
      error = 'Failed to delete tier';
    } finally {
      saving = false;
    }
  }

  onMount(() => {
    loadData();
  });
</script>

<svelte:head>
  <title>{product ? `${product.name} - Pricing` : 'Product Pricing'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading pricing data...</p>
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
        <p class="page-description">Manage pricing and price history</p>
      </div>
      <div class="header-actions">
        <Button variant="primary" on:click={openUpdatePriceModal}>
          Update Price
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
          <span class="stat-label">Current Price</span>
          <span class="stat-value">{formatCurrency(product.price)}</span>
        </div>
      </Card>

      <Card class="stat-card">
        <div class="stat-icon green">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">Margin</span>
          <span class="stat-value">{calculateMargin().toFixed(1)}%</span>
        </div>
      </Card>

      <Card class="stat-card">
        <div class="stat-icon purple">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">Cost</span>
          <span class="stat-value">{formatCurrency(product.cost)}</span>
        </div>
      </Card>

      <Card class="stat-card">
        <div class="stat-icon yellow">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">Last Updated</span>
          <span class="stat-value text-sm">{formatDate(product.updatedAt)}</span>
        </div>
      </Card>
    </div>

    <div class="content-grid">
      <Card>
        <div class="section-header">
          <h2 class="section-title">Price History</h2>
        </div>
        {#if priceHistory.length === 0}
          <p class="empty-text">No price history available</p>
        {:else}
          <Table {columns}>
            <tbody>
              {#each priceHistory as entry}
                <tr>
                  <td>
                    <Badge variant={getTypeBadgeVariant(entry.type)}>
                      {entry.type}
                    </Badge>
                  </td>
                  <td class="font-medium">{formatCurrency(entry.price)}</td>
                  <td class="text-gray-500">
                    {entry.oldPrice ? formatCurrency(entry.oldPrice) : '-'}
                  </td>
                  <td>
                    {#if entry.startDate && entry.endDate}
                      {formatDate(entry.startDate)} - {formatDate(entry.endDate)}
                    {:else}
                      -
                    {/if}
                  </td>
                  <td>{formatDate(entry.createdAt)}</td>
                  <td>{entry.createdBy}</td>
                </tr>
              {/each}
            </tbody>
          </Table>
        {/if}
      </Card>

      <Card>
        <div class="section-header">
          <h2 class="section-title">Pricing Tiers</h2>
          <Button variant="secondary" size="sm" on:click={openAddTierModal}>
            Add Tier
          </Button>
        </div>
        {#if pricingTiers.length === 0}
          <p class="empty-text">No pricing tiers configured</p>
        {:else}
          <div class="tiers-list">
            {#each pricingTiers as tier}
              <div class="tier-item">
                <div class="tier-info">
                  <span class="tier-name">{tier.name}</span>
                  <span class="tier-details">
                    Min Qty: {tier.minQuantity} • {formatCurrency(tier.price)} • {tier.discount}% off
                  </span>
                </div>
                <div class="tier-actions">
                  <Button variant="ghost" size="sm" on:click={() => openEditTierModal(tier)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={() => openDeleteTierModal(tier)}>
                    Delete
                  </Button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </Card>
    </div>
  {:else}
    <Alert variant="error">Product not found</Alert>
  {/if}
</div>

<!-- Update Price Modal -->
<Modal
  bind:open={showUpdatePriceModal}
  title="Update Price"
  size="md"
>
  <div class="modal-form">
    <div class="form-row two-col">
      <Input
        id="newPrice"
        label="New Price"
        type="number"
        placeholder="0.00"
        bind:value={newPrice}
        min="0"
        step="0.01"
        required
      />
      <Input
        id="newCost"
        label="New Cost"
        type="number"
        placeholder="0.00"
        bind:value={newCost}
        min="0"
        step="0.01"
      />
    </div>
    <div class="form-row">
      <Select
        id="priceType"
        label="Price Type"
        options={[
          { value: 'base', label: 'Base Price' },
          { value: 'sale', label: 'Sale Price' }
        ]}
        bind:value={priceType}
      />
    </div>
    {#if priceType === 'sale'}
      <div class="form-row two-col">
        <Input
          id="saleStart"
          label="Sale Start Date"
          type="date"
          bind:value={saleStartDate}
        />
        <Input
          id="saleEnd"
          label="Sale End Date"
          type="date"
          bind:value={saleEndDate}
        />
      </div>
    {/if}
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={saving}>
      Cancel
    </Button>
    <Button variant="primary" on:click={handleUpdatePrice} loading={saving}>
      {saving ? 'Saving...' : 'Update Price'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Add Tier Modal -->
<Modal
  bind:open={showAddTierModal}
  title="Add Pricing Tier"
  size="md"
>
  <div class="modal-form">
    <div class="form-row">
      <Input
        id="tierName"
        label="Tier Name"
        type="text"
        placeholder="e.g., Wholesale, Bulk"
        bind:value={tierName}
        required
      />
    </div>
    <div class="form-row two-col">
      <Input
        id="tierMinQty"
        label="Minimum Quantity"
        type="number"
        placeholder="10"
        bind:value={tierMinQty}
        min="1"
        step="1"
        required
      />
      <Input
        id="tierPrice"
        label="Tier Price"
        type="number"
        placeholder="0.00"
        bind:value={tierPrice}
        min="0"
        step="0.01"
        required
      />
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={saving}>
      Cancel
    </Button>
    <Button variant="primary" on:click={handleAddTier} loading={saving}>
      {saving ? 'Adding...' : 'Add Tier'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Edit Tier Modal -->
<Modal
  bind:open={showEditTierModal}
  title="Edit Pricing Tier"
  size="md"
>
  <div class="modal-form">
    <div class="form-row">
      <Input
        id="editTierName"
        label="Tier Name"
        type="text"
        bind:value={tierName}
        required
      />
    </div>
    <div class="form-row two-col">
      <Input
        id="editTierMinQty"
        label="Minimum Quantity"
        type="number"
        bind:value={tierMinQty}
        min="1"
        step="1"
        required
      />
      <Input
        id="editTierPrice"
        label="Tier Price"
        type="number"
        bind:value={tierPrice}
        min="0"
        step="0.01"
        required
      />
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={saving}>
      Cancel
    </Button>
    <Button variant="primary" on:click={handleUpdateTier} loading={saving}>
      {saving ? 'Saving...' : 'Save Changes'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Delete Tier Modal -->
<Modal
  bind:open={showDeleteTierModal}
  title="Delete Pricing Tier"
  size="sm"
>
  <p>Are you sure you want to delete the pricing tier <strong>{selectedTier?.name}</strong>? This action cannot be undone.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showDeleteTierModal = false; }} disabled={saving}>
      Cancel
    </Button>
    <Button variant="danger" on:click={handleDeleteTier} loading={saving}>
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

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
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

  .stat-icon.yellow {
    background-color: var(--color-yellow-100);
    color: var(--color-yellow-600);
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

  .content-grid {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 1rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0;
  }

  .empty-text {
    color: var(--color-gray-500);
    text-align: center;
    padding: 2rem;
  }

  .tiers-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .tier-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem;
    border: 1px solid var(--color-gray-200);
    border-radius: 0.5rem;
  }

  .tier-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .tier-name {
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .tier-details {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .tier-actions {
    display: flex;
    gap: 0.25rem;
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

  @media (max-width: 768px) {
    .stats-grid {
      grid-template-columns: repeat(2, 1fr);
    }

    .content-grid {
      grid-template-columns: 1fr;
    }

    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .form-row.two-col {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 480px) {
    .stats-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
</style>
