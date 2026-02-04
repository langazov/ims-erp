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
  import type { Address } from '$lib/shared/api/clients';

  const clientId = $page.params.id;

  interface ClientAddress extends Address {
    id: string;
    type: 'billing' | 'shipping';
    isDefault: boolean;
  }

  let addresses: ClientAddress[] = [];
  let loading = true;
  let error: string | null = null;
  let showAddModal = false;
  let editingAddress: ClientAddress | null = null;
  let deleteAddressId: string | null = null;

  // Form fields
  let addressType: 'billing' | 'shipping' = 'shipping';
  let street = '';
  let city = '';
  let state = '';
  let postalCode = '';
  let country = '';
  let isDefault = false;
  let formErrors: Record<string, string> = {};

  const typeOptions = [
    { value: 'billing', label: 'Billing' },
    { value: 'shipping', label: 'Shipping' }
  ];

  const columns = [
    { key: 'type', label: 'Type', sortable: true },
    { key: 'address', label: 'Address', sortable: false },
    { key: 'city', label: 'City', sortable: true },
    { key: 'country', label: 'Country', sortable: true },
    { key: 'default', label: 'Default', sortable: false },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  async function loadAddresses() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      addresses = [
        {
          id: '1',
          type: 'billing',
          street: '123 Main Street',
          city: 'New York',
          state: 'NY',
          postalCode: '10001',
          country: 'USA',
          isDefault: true
        },
        {
          id: '2',
          type: 'shipping',
          street: '456 Warehouse Blvd',
          city: 'Los Angeles',
          state: 'CA',
          postalCode: '90001',
          country: 'USA',
          isDefault: true
        },
        {
          id: '3',
          type: 'shipping',
          street: '789 Delivery Lane',
          city: 'Chicago',
          state: 'IL',
          postalCode: '60601',
          country: 'USA',
          isDefault: false
        }
      ];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load addresses';
    } finally {
      loading = false;
    }
  }

  function resetForm() {
    addressType = 'shipping';
    street = '';
    city = '';
    state = '';
    postalCode = '';
    country = '';
    isDefault = false;
    formErrors = {};
    editingAddress = null;
  }

  function validateForm(): boolean {
    formErrors = {};

    if (!street.trim()) {
      formErrors.street = 'Street is required';
    }
    if (!city.trim()) {
      formErrors.city = 'City is required';
    }
    if (!country.trim()) {
      formErrors.country = 'Country is required';
    }

    return Object.keys(formErrors).length === 0;
  }

  function openAddModal() {
    resetForm();
    showAddModal = true;
  }

  function openEditModal(address: ClientAddress) {
    editingAddress = address;
    addressType = address.type;
    street = address.street;
    city = address.city;
    state = address.state;
    postalCode = address.postalCode;
    country = address.country;
    isDefault = address.isDefault;
    formErrors = {};
    showAddModal = true;
  }

  async function handleSave() {
    if (!validateForm()) return;

    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      
      if (editingAddress) {
        addresses = addresses.map(addr => 
          addr.id === editingAddress!.id 
            ? { ...addr, type: addressType, street, city, state, postalCode, country, isDefault }
            : addr
        );
      } else {
        const newAddress: ClientAddress = {
          id: Date.now().toString(),
          type: addressType,
          street,
          city,
          state,
          postalCode,
          country,
          isDefault
        };
        addresses = [...addresses, newAddress];
      }
      
      showAddModal = false;
      resetForm();
    } catch (err) {
      error = 'Failed to save address';
    }
  }

  async function confirmDelete() {
    if (!deleteAddressId) return;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      addresses = addresses.filter(addr => addr.id !== deleteAddressId);
      deleteAddressId = null;
    } catch (err) {
      error = 'Failed to delete address';
    }
  }

  function getTypeVariant(type: string): 'blue' | 'purple' | 'gray' {
    switch (type) {
      case 'billing': return 'blue';
      case 'shipping': return 'purple';
      default: return 'gray';
    }
  }

  function formatAddress(address: ClientAddress): string {
    return `${address.street}${address.state ? ', ' + address.state : ''} ${address.postalCode}`;
  }

  onMount(() => {
    loadAddresses();
  });
</script>

<svelte:head>
  <title>Client Addresses | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Client Addresses</h1>
      <p class="page-description">Manage billing and shipping addresses</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={openAddModal}>
        Add Address
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  <Card>
    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading addresses...</p>
      </div>
    {:else if addresses.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <p class="text-gray-500 mb-4">No addresses found</p>
        <Button variant="primary" on:click={openAddModal}>
          Add First Address
        </Button>
      </div>
    {:else}
      <Table {columns} data={addresses}>
        <tbody>
          {#each addresses as address}
            <tr>
              <td>
                <Badge variant={getTypeVariant(address.type)}>
                  {address.type}
                </Badge>
              </td>
              <td class="max-w-xs truncate">{formatAddress(address)}</td>
              <td>{address.city}</td>
              <td>{address.country}</td>
              <td>
                {#if address.isDefault}
                  <Badge variant="green" size="sm">Default</Badge>
                {:else}
                  <span class="text-gray-400">—</span>
                {/if}
              </td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={() => openEditModal(address)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={() => deleteAddressId = address.id}>
                    Delete
                  </Button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </Table>
    {/if}
  </Card>
</div>

<Modal
  bind:open={showAddModal}
  title={editingAddress ? 'Edit Address' : 'Add Address'}
  size="lg"
>
  <div class="modal-form">
    <div class="form-row">
      <Select
        id="addressType"
        label="Address Type"
        options={typeOptions}
        bind:value={addressType}
        required
      />
    </div>
    <div class="form-row full-width">
      <Input
        id="street"
        label="Street"
        type="text"
        placeholder="Enter street address"
        bind:value={street}
        required
        error={formErrors.street}
      />
    </div>
    <div class="form-row">
      <Input
        id="city"
        label="City"
        type="text"
        placeholder="Enter city"
        bind:value={city}
        required
        error={formErrors.city}
      />
    </div>
    <div class="form-row">
      <Input
        id="state"
        label="State/Province"
        type="text"
        placeholder="Enter state"
        bind:value={state}
      />
    </div>
    <div class="form-row">
      <Input
        id="postalCode"
        label="Postal Code"
        type="text"
        placeholder="Enter postal code"
        bind:value={postalCode}
      />
    </div>
    <div class="form-row">
      <Input
        id="country"
        label="Country"
        type="text"
        placeholder="Enter country"
        bind:value={country}
        required
        error={formErrors.country}
      />
    </div>
    <div class="form-row full-width">
      <label class="checkbox-label">
        <input type="checkbox" bind:checked={isDefault} />
        <span>Set as default {addressType} address</span>
      </label>
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close}>Cancel</Button>
    <Button variant="primary" on:click={handleSave}>
      {editingAddress ? 'Save Changes' : 'Add Address'}
    </Button>
  </svelte:fragment>
</Modal>

{#if deleteAddressId}
  <Modal
    open={true}
    title="Delete Address"
    size="sm"
  >
    <p>Are you sure you want to delete this address? This action cannot be undone.</p>
    
    <svelte:fragment slot="footer" let:close>
      <Button variant="secondary" on:click={() => { close(); deleteAddressId = null; }}>Cancel</Button>
      <Button variant="danger" on:click={confirmDelete}>Delete</Button>
    </svelte:fragment>
  </Modal>
{/if}

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1200px;
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

  .form-row.full-width {
    grid-column: 1 / -1;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-size: 0.875rem;
    color: var(--color-gray-700);
  }

  .checkbox-label input[type="checkbox"] {
    width: 1rem;
    height: 1rem;
    accent-color: var(--color-primary-600);
  }

  @media (max-width: 640px) {
    .modal-form {
      grid-template-columns: 1fr;
    }

    .form-row.full-width {
      grid-column: 1;
    }
  }
</style>
