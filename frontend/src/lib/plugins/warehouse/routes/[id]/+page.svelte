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

  const warehouseId = $page.params.id;

  interface Warehouse {
    id: string;
    code: string;
    name: string;
    type: 'main' | 'distribution' | 'retail' | 'virtual';
    status: 'active' | 'inactive';
    address: {
      street: string;
      city: string;
      state: string;
      postalCode: string;
      country: string;
    };
    capacity: number;
    utilizedCapacity: number;
    locationCount: number;
    operationCount: number;
    createdAt: string;
    updatedAt: string;
  }

  let warehouse: Warehouse | null = null;
  let loading = true;
  let error: string | null = null;
  let showDeleteModal = false;
  let deleting = false;

  async function loadWarehouse() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      warehouse = {
        id: warehouseId,
        code: 'WH001',
        name: 'Main Distribution Center',
        type: 'main',
        status: 'active',
        address: {
          street: '123 Industrial Parkway',
          city: 'New York',
          state: 'NY',
          postalCode: '10001',
          country: 'USA'
        },
        capacity: 10000,
        utilizedCapacity: 7500,
        locationCount: 150,
        operationCount: 42,
        createdAt: '2024-01-15T10:00:00Z',
        updatedAt: '2024-01-20T14:30:00Z'
      };
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load warehouse';
    } finally {
      loading = false;
    }
  }

  function getTypeLabel(type: string): string {
    const labels: Record<string, string> = {
      main: 'Main',
      distribution: 'Distribution',
      retail: 'Retail',
      virtual: 'Virtual'
    };
    return labels[type] || type;
  }

  function getStatusVariant(status: string): 'green' | 'gray' {
    return status === 'active' ? 'green' : 'gray';
  }

  function getCapacityPercentage(): number {
    if (!warehouse || warehouse.capacity === 0) return 0;
    return Math.round((warehouse.utilizedCapacity / warehouse.capacity) * 100);
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }

  function handleEdit() {
    goto(`/warehouse/${warehouseId}/edit`);
  }

  function handleDelete() {
    showDeleteModal = true;
  }

  async function confirmDelete() {
    deleting = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      goto('/warehouse');
    } catch (err) {
      error = 'Failed to delete warehouse';
      deleting = false;
    }
  }

  function handleViewLocations() {
    goto(`/warehouse/${warehouseId}/locations`);
  }

  onMount(() => {
    loadWarehouse();
  });
</script>

<svelte:head>
  <title>Warehouse Details | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading warehouse details...</p>
    </div>
  {:else if warehouse}
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <h1 class="page-title">{warehouse.name}</h1>
          <Badge variant={getStatusVariant(warehouse.status)} size="md">
            {warehouse.status}
          </Badge>
        </div>
        <p class="page-description">
          {warehouse.code} • {getTypeLabel(warehouse.type)}
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

    {#if error}
      <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
        {error}
      </Alert>
    {/if}

    <div class="stats-grid">
      <Card class="stat-card">
        <div class="stat-icon blue">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">Locations</span>
          <span class="stat-value">{warehouse.locationCount}</span>
        </div>
      </Card>

      <Card class="stat-card">
        <div class="stat-icon green">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">Operations</span>
          <span class="stat-value">{warehouse.operationCount}</span>
        </div>
      </Card>

      <Card class="stat-card">
        <div class="stat-icon purple">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">Created</span>
          <span class="stat-value text-sm">{formatDate(warehouse.createdAt)}</span>
        </div>
      </Card>
    </div>

    <div class="details-grid">
      <Card>
        <h2 class="section-title">Capacity Overview</h2>
        <div class="capacity-info">
          <div class="capacity-bar-container">
            <div class="capacity-bar">
              <div 
                class="capacity-fill"
                class:bg-green-500={getCapacityPercentage() < 50}
                class:bg-yellow-500={getCapacityPercentage() >= 50 && getCapacityPercentage() < 80}
                class:bg-red-500={getCapacityPercentage() >= 80}
                style="width: {getCapacityPercentage()}%"
              />
            </div>
            <span class="capacity-percentage">{getCapacityPercentage()}%</span>
          </div>
          <div class="capacity-details">
            <div class="capacity-item">
              <span class="capacity-label">Total Capacity</span>
              <span class="capacity-value">{warehouse.capacity.toLocaleString()} sq ft</span>
            </div>
            <div class="capacity-item">
              <span class="capacity-label">Utilized</span>
              <span class="capacity-value">{warehouse.utilizedCapacity.toLocaleString()} sq ft</span>
            </div>
            <div class="capacity-item">
              <span class="capacity-label">Available</span>
              <span class="capacity-value text-green-600">
                {(warehouse.capacity - warehouse.utilizedCapacity).toLocaleString()} sq ft
              </span>
            </div>
          </div>
        </div>
      </Card>

      <Card>
        <h2 class="section-title">Address</h2>
        <div class="address-block">
          <p class="address-line">{warehouse.address.street}</p>
          <p class="address-line">
            {warehouse.address.city}, {warehouse.address.state} {warehouse.address.postalCode}
          </p>
          <p class="address-line">{warehouse.address.country}</p>
        </div>
      </Card>
    </div>

    <Card class="actions-card">
      <h2 class="section-title">Quick Actions</h2>
      <div class="quick-actions">
        <Button variant="secondary" on:click={handleViewLocations}>
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          View Locations
        </Button>
        <Button variant="secondary">
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          View Inventory
        </Button>
        <Button variant="secondary">
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
          View Operations
        </Button>
      </div>
    </Card>
  {:else}
    <Alert variant="error">Warehouse not found</Alert>
  {/if}
</div>

<Modal
  bind:open={showDeleteModal}
  title="Delete Warehouse"
  size="sm"
>
  <p>Are you sure you want to delete this warehouse? This action cannot be undone.</p>
  
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

  .loading-container {
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
    grid-template-columns: 2fr 1fr;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin-bottom: 1rem;
  }

  .capacity-info {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .capacity-bar-container {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .capacity-bar {
    flex: 1;
    height: 1rem;
    background-color: var(--color-gray-200);
    border-radius: 9999px;
    overflow: hidden;
  }

  .capacity-fill {
    height: 100%;
    border-radius: 9999px;
    transition: width 0.3s ease;
  }

  .capacity-percentage {
    font-weight: 600;
    color: var(--color-gray-700);
    min-width: 3rem;
    text-align: right;
  }

  .capacity-details {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
  }

  .capacity-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .capacity-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .capacity-value {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .address-block {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .address-line {
    color: var(--color-gray-700);
    margin: 0;
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

    .capacity-details {
      grid-template-columns: 1fr;
    }

    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .header-actions {
      width: 100%;
    }
  }
</style>
