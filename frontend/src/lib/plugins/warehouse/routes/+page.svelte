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
    createdAt: string;
  }

  let warehouses: Warehouse[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter: string = '';
  let typeFilter: string = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;
  let showCreateModal = false;
  let deleteWarehouseId: string | null = null;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' }
  ];

  const typeOptions = [
    { value: '', label: 'All Types' },
    { value: 'main', label: 'Main' },
    { value: 'distribution', label: 'Distribution' },
    { value: 'retail', label: 'Retail' },
    { value: 'virtual', label: 'Virtual' }
  ];

  const columns = [
    { key: 'code', label: 'Code', sortable: true },
    { key: 'name', label: 'Name', sortable: true },
    { key: 'type', label: 'Type', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'capacity', label: 'Capacity', sortable: true },
    { key: 'locationCount', label: 'Locations', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  function getStatusVariant(status: string): 'green' | 'gray' | 'yellow' | 'red' {
    switch (status) {
      case 'active': return 'green';
      case 'inactive': return 'gray';
      default: return 'gray';
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

  function getCapacityPercentage(warehouse: Warehouse): number {
    if (warehouse.capacity === 0) return 0;
    return Math.round((warehouse.utilizedCapacity / warehouse.capacity) * 100);
  }

  async function loadWarehouses() {
    loading = true;
    error = null;
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // Mock data
      warehouses = [
        {
          id: '1',
          code: 'WH001',
          name: 'Main Distribution Center',
          type: 'main',
          status: 'active',
          address: { street: '123 Main St', city: 'New York', state: 'NY', postalCode: '10001', country: 'USA' },
          capacity: 10000,
          utilizedCapacity: 7500,
          locationCount: 150,
          createdAt: '2024-01-15'
        },
        {
          id: '2',
          code: 'WH002',
          name: 'West Coast Hub',
          type: 'distribution',
          status: 'active',
          address: { street: '456 West Ave', city: 'Los Angeles', state: 'CA', postalCode: '90001', country: 'USA' },
          capacity: 8000,
          utilizedCapacity: 6200,
          locationCount: 120,
          createdAt: '2024-02-20'
        },
        {
          id: '3',
          code: 'WH003',
          name: 'Retail Store - Downtown',
          type: 'retail',
          status: 'active',
          address: { street: '789 Shop St', city: 'Chicago', state: 'IL', postalCode: '60601', country: 'USA' },
          capacity: 500,
          utilizedCapacity: 350,
          locationCount: 25,
          createdAt: '2024-03-10'
        }
      ];
      
      totalItems = warehouses.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load warehouses';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadWarehouses();
  }

  function handleRowClick(warehouse: Warehouse) {
    window.location.href = `/warehouse/${warehouse.id}`;
  }

  function handleEdit(warehouse: Warehouse, event: Event) {
    event.stopPropagation();
    window.location.href = `/warehouse/${warehouse.id}/edit`;
  }

  async function handleDelete(warehouse: Warehouse, event: Event) {
    event.stopPropagation();
    deleteWarehouseId = warehouse.id;
  }

  async function confirmDelete() {
    if (!deleteWarehouseId) return;
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 300));
      warehouses = warehouses.filter(w => w.id !== deleteWarehouseId);
      totalItems = warehouses.length;
      deleteWarehouseId = null;
    } catch (err) {
      error = 'Failed to delete warehouse';
    }
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadWarehouses();
  }

  onMount(() => {
    loadWarehouses();
  });
</script>

<svelte:head>
  <title>Warehouses | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Warehouses</h1>
      <p class="page-description">Manage your warehouse locations and storage facilities</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={() => showCreateModal = true}>
        Add Warehouse
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
            placeholder="Search warehouses..."
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
            id="type"
            label="Type"
            options={typeOptions}
            bind:value={typeFilter}
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
        <p>Loading warehouses...</p>
      </div>
    {:else if warehouses.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
        </svg>
        <p class="text-gray-500 mb-4">No warehouses found</p>
        <Button variant="primary" on:click={() => showCreateModal = true}>
          Add Your First Warehouse
        </Button>
      </div>
    {:else}
      <Table {columns}>
        <tbody>
          {#each warehouses as warehouse}
            <tr on:click={() => handleRowClick(warehouse)} class="clickable-row">
              <td class="font-medium">{warehouse.code}</td>
              <td>{warehouse.name}</td>
              <td>{getTypeLabel(warehouse.type)}</td>
              <td>
                <Badge variant={getStatusVariant(warehouse.status)}>
                  {warehouse.status}
                </Badge>
              </td>
              <td>
                <div class="flex items-center gap-2">
                  <div class="w-24 bg-gray-200 rounded-full h-2">
                    <div 
                      class="bg-primary-600 h-2 rounded-full" 
                      style="width: {getCapacityPercentage(warehouse)}%"
                    />
                  </div>
                  <span class="text-xs text-gray-500">
                    {getCapacityPercentage(warehouse)}%
                  </span>
                </div>
              </td>
              <td>{warehouse.locationCount}</td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={(e) => handleEdit(warehouse, e)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={(e) => handleDelete(warehouse, e)}>
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
  title="Create Warehouse"
  size="lg"
>
  <p class="text-gray-600">Warehouse creation form will be implemented here.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close}>Cancel</Button>
    <Button variant="primary" on:click={() => {
      showCreateModal = false;
      // Handle create
    }}>Create</Button>
  </svelte:fragment>
</Modal>

{#if deleteWarehouseId}
  <Modal
    open={true}
    title="Delete Warehouse"
    size="sm"
  >
    <p>Are you sure you want to delete this warehouse? This action cannot be undone.</p>
    
    <svelte:fragment slot="footer" let:close>
      <Button variant="secondary" on:click={() => { close(); deleteWarehouseId = null; }}>Cancel</Button>
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
