<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Form from '$lib/shared/components/forms/Form.svelte';
  import { ArrowLeft, Plus, Trash2 } from 'lucide-svelte';

  const warehouseId = $page.params.id;

  interface OperationItem {
    id: string;
    productId: string;
    productName: string;
    quantity: number;
    locationId?: string;
  }

  interface Warehouse {
    id: string;
    name: string;
    code: string;
    locations: Array<{ id: string; code: string; name: string }>;
  }

  let warehouse: Warehouse | null = null;
  let loading = true;
  let saving = false;
  let error: string | null = null;

  let operationType: 'receipt' | 'issue' | 'transfer' | 'adjustment' | 'return' = 'receipt';
  let referenceType = 'purchase_order';
  let reference = '';
  let notes = '';
  let items: OperationItem[] = [];

  const operationTypes = [
    { value: 'receipt', label: 'Receipt' },
    { value: 'issue', label: 'Issue' },
    { value: 'transfer', label: 'Transfer' },
    { value: 'adjustment', label: 'Adjustment' },
    { value: 'return', label: 'Return' },
  ];

  const referenceTypes = [
    { value: 'purchase_order', label: 'Purchase Order' },
    { value: 'sales_order', label: 'Sales Order' },
    { value: 'transfer_order', label: 'Transfer Order' },
    { value: 'return_order', label: 'Return Order' },
    { value: 'adjustment', label: 'Adjustment' },
  ];

  async function loadWarehouse() {
    loading = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      
      warehouse = {
        id: warehouseId,
        name: 'Main Distribution Center',
        code: 'WH001',
        locations: [
          { id: 'loc-1', code: 'A-01-01', name: 'Aisle A, Rack 1, Shelf 1' },
          { id: 'loc-2', code: 'A-01-02', name: 'Aisle A, Rack 1, Shelf 2' },
          { id: 'loc-3', code: 'B-02-01', name: 'Aisle B, Rack 2, Shelf 1' },
        ]
      };
    } catch (err) {
      error = 'Failed to load warehouse';
    } finally {
      loading = false;
    }
  }

  function addItem() {
    items = [...items, {
      id: crypto.randomUUID(),
      productId: '',
      productName: '',
      quantity: 1,
      locationId: undefined
    }];
  }

  function removeItem(itemId: string) {
    items = items.filter(item => item.id !== itemId);
  }

  function updateItem(itemId: string, updates: Partial<OperationItem>) {
    items = items.map(item => 
      item.id === itemId ? { ...item, ...updates } : item
    );
  }

  async function handleSubmit() {
    if (items.length === 0) {
      error = 'Please add at least one item';
      return;
    }

    saving = true;
    error = null;

    try {
      await new Promise(resolve => setTimeout(resolve, 1000));
      goto(`/warehouse/${warehouseId}/operations`);
    } catch (err) {
      error = 'Failed to create operation';
      saving = false;
    }
  }

  function handleCancel() {
    goto(`/warehouse/${warehouseId}/operations`);
  }

  onMount(() => {
    loadWarehouse();
    addItem();
  });
</script>

<svelte:head>
  <title>New Operation | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <Button variant="ghost" on:click={handleCancel} class="back-button">
        <ArrowLeft class="w-4 h-4 mr-2" />
        Back to Operations
      </Button>
      <h1 class="page-title">Create New Operation</h1>
      <p class="page-description">
        {warehouse ? `${warehouse.name} (${warehouse.code})` : 'Loading warehouse...'}
      </p>
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
      <p>Loading...</p>
    </div>
  {:else}
    <Form on:submit={handleSubmit} class="space-y-6">
      <!-- Operation Details -->
      <Card>
        <h2 class="section-title">Operation Details</h2>
        <div class="form-grid">
          <Select
            label="Operation Type"
            bind:value={operationType}
            options={operationTypes}
            required
          />
          <Select
            label="Reference Type"
            bind:value={referenceType}
            options={referenceTypes}
            required
          />
          <Input
            label="Reference Number"
            bind:value={reference}
            placeholder="e.g., PO-2024-001"
            required
          />
        </div>
        <div class="mt-4">
          <Input
            label="Notes"
            bind:value={notes}
            placeholder="Optional notes about this operation"
          />
        </div>
      </Card>

      <!-- Items -->
      <Card>
        <div class="items-header">
          <h2 class="section-title">Items</h2>
          <Button type="button" variant="secondary" size="sm" on:click={addItem}>
            <Plus class="w-4 h-4 mr-2" />
            Add Item
          </Button>
        </div>

        {#if items.length === 0}
          <div class="empty-state">
            <p>No items added. Click "Add Item" to start.</p>
          </div>
        {:else}
          <div class="items-list">
            {#each items as item, index (item.id)}
              <div class="item-row">
                <div class="item-number">{index + 1}</div>
                <div class="item-fields">
                  <Input
                    label="Product"
                    bind:value={item.productName}
                    placeholder="Search product..."
                    required
                  />
                  <Input
                    label="Quantity"
                    type="number"
                    bind:value={item.quantity}
                    min={1}
                    required
                  />
                  <Select
                    label="Location"
                    bind:value={item.locationId}
                    options={warehouse?.locations.map(loc => ({ 
                      value: loc.id, 
                      label: `${loc.code} - ${loc.name}` 
                    })) || []}
                    placeholder="Select location..."
                  />
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  class="delete-button"
                  on:click={() => removeItem(item.id)}
                >
                  <Trash2 class="w-4 h-4 text-red-500" />
                </Button>
              </div>
            {/each}
          </div>
        {/if}
      </Card>

      <!-- Actions -->
      <div class="actions">
        <Button type="button" variant="secondary" on:click={handleCancel}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" loading={saving}>
          {saving ? 'Creating...' : 'Create Operation'}
        </Button>
      </div>
    </Form>
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

  .header-content {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .back-button {
    align-self: flex-start;
  }

  .page-title {
    font-size: 1.875rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin: 0;
  }

  .page-description {
    color: var(--color-gray-500);
    margin: 0;
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

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin-bottom: 1rem;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
  }

  .items-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .items-header .section-title {
    margin-bottom: 0;
  }

  .empty-state {
    text-align: center;
    padding: 2rem;
    color: var(--color-gray-500);
  }

  .items-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .item-row {
    display: flex;
    align-items: flex-end;
    gap: 1rem;
    padding: 1rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
  }

  .item-number {
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--color-primary-100);
    color: var(--color-primary-700);
    font-weight: 600;
    border-radius: 0.375rem;
    flex-shrink: 0;
  }

  .item-fields {
    flex: 1;
    display: grid;
    grid-template-columns: 2fr 1fr 2fr;
    gap: 1rem;
  }

  .delete-button {
    flex-shrink: 0;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
  }

  @media (max-width: 768px) {
    .form-grid {
      grid-template-columns: 1fr;
    }

    .item-row {
      flex-direction: column;
      align-items: stretch;
    }

    .item-fields {
      grid-template-columns: 1fr;
    }

    .actions {
      flex-direction: column;
    }
  }
</style>

