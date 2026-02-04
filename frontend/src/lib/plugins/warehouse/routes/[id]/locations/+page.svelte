<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
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

  const warehouseId = $page.params.id;

  interface WarehouseLocation {
    id: string;
    code: string;
    name: string;
    type: 'zone' | 'aisle' | 'rack' | 'bin' | 'shelf';
    parentId: string | null;
    capacity: number;
    utilizedCapacity: number;
    status: 'active' | 'inactive';
  }

  let locations: WarehouseLocation[] = [];
  let loading = true;
  let error: string | null = null;
  let showAddModal = false;
  let editingLocation: WarehouseLocation | null = null;
  let deleteLocationId: string | null = null;
  let typeFilter = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;

  // Form fields
  let locationCode = '';
  let locationName = '';
  let locationType: 'zone' | 'aisle' | 'rack' | 'bin' | 'shelf' = 'zone';
  let parentLocation = '';
  let locationCapacity = '';
  let formErrors: Record<string, string> = {};

  const typeOptions = [
    { value: '', label: 'All Types' },
    { value: 'zone', label: 'Zone' },
    { value: 'aisle', label: 'Aisle' },
    { value: 'rack', label: 'Rack' },
    { value: 'shelf', label: 'Shelf' },
    { value: 'bin', label: 'Bin' }
  ];

  const locationTypeOptions = [
    { value: 'zone', label: 'Zone' },
    { value: 'aisle', label: 'Aisle' },
    { value: 'rack', label: 'Rack' },
    { value: 'shelf', label: 'Shelf' },
    { value: 'bin', label: 'Bin' }
  ];

  const columns = [
    { key: 'code', label: 'Code', sortable: true },
    { key: 'name', label: 'Name', sortable: true },
    { key: 'type', label: 'Type', sortable: true },
    { key: 'capacity', label: 'Capacity', sortable: true },
    { key: 'utilization', label: 'Utilization', sortable: false },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  async function loadLocations() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      locations = [
        {
          id: '1',
          code: 'A',
          name: 'Zone A - Electronics',
          type: 'zone',
          parentId: null,
          capacity: 5000,
          utilizedCapacity: 3200,
          status: 'active'
        },
        {
          id: '2',
          code: 'A-01',
          name: 'Aisle 01',
          type: 'aisle',
          parentId: '1',
          capacity: 1000,
          utilizedCapacity: 800,
          status: 'active'
        },
        {
          id: '3',
          code: 'A-01-R01',
          name: 'Rack 01',
          type: 'rack',
          parentId: '2',
          capacity: 200,
          utilizedCapacity: 150,
          status: 'active'
        },
        {
          id: '4',
          code: 'A-01-R01-B01',
          name: 'Bin 01',
          type: 'bin',
          parentId: '3',
          capacity: 20,
          utilizedCapacity: 15,
          status: 'active'
        },
        {
          id: '5',
          code: 'B',
          name: 'Zone B - Clothing',
          type: 'zone',
          parentId: null,
          capacity: 3000,
          utilizedCapacity: 2100,
          status: 'active'
        },
        {
          id: '6',
          code: 'B-01',
          name: 'Aisle 01',
          type: 'aisle',
          parentId: '5',
          capacity: 600,
          utilizedCapacity: 420,
          status: 'active'
        }
      ];
      
      totalItems = locations.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load locations';
    } finally {
      loading = false;
    }
  }

  function getTypeVariant(type: string): 'blue' | 'purple' | 'green' | 'yellow' | 'gray' {
    switch (type) {
      case 'zone': return 'blue';
      case 'aisle': return 'purple';
      case 'rack': return 'green';
      case 'shelf': return 'yellow';
      case 'bin': return 'gray';
      default: return 'gray';
    }
  }

  function getStatusVariant(status: string): 'green' | 'gray' {
    return status === 'active' ? 'green' : 'gray';
  }

  function getCapacityPercentage(location: WarehouseLocation): number {
    if (location.capacity === 0) return 0;
    return Math.round((location.utilizedCapacity / location.capacity) * 100);
  }

  function resetForm() {
    locationCode = '';
    locationName = '';
    locationType = 'zone';
    parentLocation = '';
    locationCapacity = '';
    formErrors = {};
    editingLocation = null;
  }

  function validateForm(): boolean {
    formErrors = {};

    if (!locationCode.trim()) {
      formErrors.code = 'Code is required';
    }

    if (!locationName.trim()) {
      formErrors.name = 'Name is required';
    }

    if (locationCapacity && (isNaN(parseFloat(locationCapacity)) || parseFloat(locationCapacity) < 0)) {
      formErrors.capacity = 'Capacity must be a positive number';
    }

    return Object.keys(formErrors).length === 0;
  }

  function openAddModal() {
    resetForm();
    showAddModal = true;
  }

  function openEditModal(location: WarehouseLocation) {
    editingLocation = location;
    locationCode = location.code;
    locationName = location.name;
    locationType = location.type;
    parentLocation = location.parentId || '';
    locationCapacity = location.capacity.toString();
    formErrors = {};
    showAddModal = true;
  }

  async function handleSave() {
    if (!validateForm()) return;

    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      
      if (editingLocation) {
        locations = locations.map(loc => 
          loc.id === editingLocation!.id 
            ? {
                ...loc,
                code: locationCode,
                name: locationName,
                type: locationType,
                parentId: parentLocation || null,
                capacity: parseFloat(locationCapacity) || 0
              }
            : loc
        );
      } else {
        const newLocation: WarehouseLocation = {
          id: Date.now().toString(),
          code: locationCode,
          name: locationName,
          type: locationType,
          parentId: parentLocation || null,
          capacity: parseFloat(locationCapacity) || 0,
          utilizedCapacity: 0,
          status: 'active'
        };
        locations = [...locations, newLocation];
      }
      
      totalItems = locations.length;
      totalPages = Math.ceil(totalItems / pageSize);
      showAddModal = false;
      resetForm();
    } catch (err) {
      error = 'Failed to save location';
    }
  }

  async function confirmDelete() {
    if (!deleteLocationId) return;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      locations = locations.filter(loc => loc.id !== deleteLocationId);
      totalItems = locations.length;
      totalPages = Math.ceil(totalItems / pageSize);
      deleteLocationId = null;
    } catch (err) {
      error = 'Failed to delete location';
    }
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadLocations();
  }

  function getParentOptions(): { value: string; label: string }[] {
    const parents = locations.filter(loc => 
      loc.type !== 'bin' && loc.type !== 'shelf'
    );
    return [
      { value: '', label: 'None (Root Level)' },
      ...parents.map(loc => ({ value: loc.id, label: `${loc.code} - ${loc.name}` }))
    ];
  }

  onMount(() => {
    loadLocations();
  });
</script>

<svelte:head>
  <title>Warehouse Locations | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Warehouse Locations</h1>
      <p class="page-description">Manage storage locations within the warehouse</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={openAddModal}>
        Add Location
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  <Card>
    <div class="filters">
      <div class="filter-row">
        <div class="filter-item">
          <Select
            id="typeFilter"
            label="Filter by Type"
            options={typeOptions}
            bind:value={typeFilter}
            on:change={() => { currentPage = 1; loadLocations(); }}
          />
        </div>
      </div>
    </div>

    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading locations...</p>
      </div>
    {:else if locations.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
        </svg>
        <p class="text-gray-500 mb-4">No locations found</p>
        <Button variant="primary" on:click={openAddModal}>
          Add First Location
        </Button>
      </div>
    {:else}
      <Table columns={columns} data={locations}>
        <tbody>
          {#each locations as location}
            <tr>
              <td class="font-medium">{location.code}</td>
              <td>{location.name}</td>
              <td>
                <Badge variant={getTypeVariant(location.type)}>
                  {location.type}
                </Badge>
              </td>
              <td>{location.capacity.toLocaleString()}</td>
              <td>
                <div class="utilization-cell">
                  <div class="w-20 bg-gray-200 rounded-full h-2">
                    <div 
                      class="h-2 rounded-full"
                      class:bg-green-500={getCapacityPercentage(location) < 50}
                      class:bg-yellow-500={getCapacityPercentage(location) >= 50 && getCapacityPercentage(location) < 80}
                      class:bg-red-500={getCapacityPercentage(location) >= 80}
                      style="width: {getCapacityPercentage(location)}%"
                    />
                  </div>
                  <span class="text-xs text-gray-500">{getCapacityPercentage(location)}%</span>
                </div>
              </td>
              <td>
                <Badge variant={getStatusVariant(location.status)}>
                  {location.status}
                </Badge>
              </td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={() => openEditModal(location)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={() => deleteLocationId = location.id}>
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
  bind:open={showAddModal}
  title={editingLocation ? 'Edit Location' : 'Add Location'}
  size="lg"
>
  <div class="modal-form">
    <div class="form-row">
      <Input
        id="locationCode"
        label="Code"
        type="text"
        placeholder="e.g., A-01"
        bind:value={locationCode}
        required
        error={formErrors.code}
      />
    </div>
    <div class="form-row">
      <Input
        id="locationName"
        label="Name"
        type="text"
        placeholder="Enter location name"
        bind:value={locationName}
        required
        error={formErrors.name}
      />
    </div>
    <div class="form-row">
      <Select
        id="locationType"
        label="Type"
        options={locationTypeOptions}
        bind:value={locationType}
        required
      />
    </div>
    <div class="form-row">
      <Select
        id="parentLocation"
        label="Parent Location"
        options={getParentOptions()}
        bind:value={parentLocation}
      />
    </div>
    <div class="form-row">
      <Input
        id="locationCapacity"
        label="Capacity"
        type="number"
        placeholder="Enter capacity"
        bind:value={locationCapacity}
        min="0"
        step="0.01"
        error={formErrors.capacity}
      />
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close}>Cancel</Button>
    <Button variant="primary" on:click={handleSave}>
      {editingLocation ? 'Save Changes' : 'Add Location'}
    </Button>
  </svelte:fragment>
</Modal>

{#if deleteLocationId}
  <Modal
    open={true}
    title="Delete Location"
    size="sm"
  >
    <p>Are you sure you want to delete this location? This action cannot be undone.</p>
    
    <svelte:fragment slot="footer" let:close>
      <Button variant="secondary" on:click={() => { close(); deleteLocationId = null; }}>Cancel</Button>
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
    max-width: 300px;
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

  .utilization-cell {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .actions-cell {
    display: flex;
    gap: 0.5rem;
  }

  .modal-form {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .form-row {
    min-width: 0;
  }

  @media (max-width: 640px) {
    .modal-form {
      grid-template-columns: 1fr;
    }

    .filter-item {
      min-width: 100%;
    }
  }
</style>
