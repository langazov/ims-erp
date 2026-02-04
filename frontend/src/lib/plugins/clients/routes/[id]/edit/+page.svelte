<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import { clientsStore } from '../../../stores';
  import type { Client, Address, UpdateClientRequest, ClientStatus } from '$lib/shared/api/clients';

  const clientId = $page.params.id;
  
  let client: Client | null = null;
  let loading = true;
  let saving = false;
  let error: string | null = null;
  let errors: Record<string, string> = {};

  // Form fields
  let name = '';
  let email = '';
  let phone = '';
  let status: ClientStatus = 'active';
  let creditLimit = '';
  
  // Billing address
  let billingStreet = '';
  let billingCity = '';
  let billingState = '';
  let billingPostalCode = '';
  let billingCountry = '';

  const statusOptions = [
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'suspended', label: 'Suspended' },
    { value: 'merged', label: 'Merged' }
  ];

  onMount(async () => {
    try {
      client = await clientsStore.loadClient(clientId);
      populateForm(client);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load client';
    } finally {
      loading = false;
    }
  });

  function populateForm(clientData: Client) {
    name = clientData.name;
    email = clientData.email;
    phone = clientData.phone || '';
    status = clientData.status;
    creditLimit = clientData.creditLimit || '';
    
    if (clientData.billingAddress) {
      billingStreet = clientData.billingAddress.street || '';
      billingCity = clientData.billingAddress.city || '';
      billingState = clientData.billingAddress.state || '';
      billingPostalCode = clientData.billingAddress.postalCode || '';
      billingCountry = clientData.billingAddress.country || '';
    }
  }

  function validateForm(): boolean {
    errors = {};

    if (!name.trim()) {
      errors.name = 'Name is required';
    }
    
    if (!email.trim()) {
      errors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      errors.email = 'Invalid email format';
    }

    return Object.keys(errors).length === 0;
  }

  async function handleSubmit() {
    if (!validateForm()) return;

    const billingAddress: Address | undefined = billingStreet || billingCity || billingState || billingPostalCode || billingCountry
      ? {
          street: billingStreet,
          city: billingCity,
          state: billingState,
          postalCode: billingPostalCode,
          country: billingCountry
        }
      : undefined;

    const data: UpdateClientRequest = {
      name,
      email,
      phone: phone || undefined,
      status,
      creditLimit: creditLimit ? parseFloat(creditLimit) : undefined,
      billingAddress
    };

    saving = true;
    try {
      await clientsStore.updateClient(clientId, data);
      goto(`/clients/${clientId}`);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to update client';
    } finally {
      saving = false;
    }
  }

  function handleCancel() {
    goto(`/clients/${clientId}`);
  }
</script>

<svelte:head>
  <title>Edit Client | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Edit Client</h1>
      <p class="page-description">Update client information and settings</p>
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
      <p>Loading client...</p>
    </div>
  {:else if client}
    <Card>
      <form on:submit|preventDefault={handleSubmit}>
        <div class="form-section">
          <h2 class="section-title">Basic Information</h2>
          <div class="form-grid">
            <div class="form-item full-width">
              <Input
                id="name"
                label="Name"
                type="text"
                placeholder="Enter client name"
                bind:value={name}
                required
                error={errors.name}
              />
            </div>
            <div class="form-item">
              <Input
                id="email"
                label="Email"
                type="email"
                placeholder="Enter email address"
                bind:value={email}
                required
                error={errors.email}
              />
            </div>
            <div class="form-item">
              <Input
                id="phone"
                label="Phone"
                type="tel"
                placeholder="Enter phone number"
                bind:value={phone}
              />
            </div>
            <div class="form-item">
              <Select
                id="status"
                label="Status"
                options={statusOptions}
                bind:value={status}
                required
              />
            </div>
            <div class="form-item">
              <Input
                id="creditLimit"
                label="Credit Limit"
                type="number"
                placeholder="Enter credit limit"
                bind:value={creditLimit}
                min="0"
                step="0.01"
              />
            </div>
          </div>
        </div>

        <div class="form-section">
          <h2 class="section-title">Billing Address</h2>
          <div class="form-grid">
            <div class="form-item full-width">
              <Input
                id="billingStreet"
                label="Street"
                type="text"
                placeholder="Enter street address"
                bind:value={billingStreet}
              />
            </div>
            <div class="form-item">
              <Input
                id="billingCity"
                label="City"
                type="text"
                placeholder="Enter city"
                bind:value={billingCity}
              />
            </div>
            <div class="form-item">
              <Input
                id="billingState"
                label="State/Province"
                type="text"
                placeholder="Enter state"
                bind:value={billingState}
              />
            </div>
            <div class="form-item">
              <Input
                id="billingPostalCode"
                label="Postal Code"
                type="text"
                placeholder="Enter postal code"
                bind:value={billingPostalCode}
              />
            </div>
            <div class="form-item">
              <Input
                id="billingCountry"
                label="Country"
                type="text"
                placeholder="Enter country"
                bind:value={billingCountry}
              />
            </div>
          </div>
        </div>

        <div class="form-actions">
          <Button variant="secondary" on:click={handleCancel} disabled={saving}>
            Cancel
          </Button>
          <Button variant="primary" type="submit" loading={saving}>
            {saving ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </form>
    </Card>
  {:else}
    <Alert variant="error">Client not found</Alert>
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

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .form-section {
    margin-bottom: 2rem;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin-bottom: 1rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .form-item {
    min-width: 0;
  }

  .form-item.full-width {
    grid-column: 1 / -1;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--color-gray-200);
    margin-top: 2rem;
  }

  @media (max-width: 640px) {
    .form-grid {
      grid-template-columns: 1fr;
    }

    .form-item.full-width {
      grid-column: 1;
    }
  }
</style>
