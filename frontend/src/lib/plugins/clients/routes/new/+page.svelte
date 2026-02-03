<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import { clientsStore } from '../../stores';
  import type { CreateClientRequest, Address } from '$lib/shared/api/clients';

  let name = '';
  let email = '';
  let phone = '';
  let creditLimit = '';
  let billingStreet = '';
  let billingCity = '';
  let billingState = '';
  let billingPostalCode = '';
  let billingCountry = '';

  let errors: Record<string, string> = {};
  let submitting = false;

  async function handleSubmit() {
    errors = {};

    if (!name.trim()) {
      errors.name = 'Name is required';
    }
    if (!email.trim()) {
      errors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      errors.email = 'Invalid email format';
    }

    if (Object.keys(errors).length > 0) {
      return;
    }

    const billingAddress: Address | undefined = billingStreet || billingCity || billingState || billingPostalCode || billingCountry
      ? {
          street: billingStreet,
          city: billingCity,
          state: billingState,
          postalCode: billingPostalCode,
          country: billingCountry
        }
      : undefined;

    const data: CreateClientRequest = {
      name,
      email,
      phone: phone || undefined,
      creditLimit: creditLimit ? parseFloat(creditLimit) : undefined,
      billingAddress
    };

    submitting = true;
    try {
      const client = await clientsStore.createClient(data);
      goto(`/clients/${client.id}`);
    } catch (error) {
      console.error('Failed to create client:', error);
    } finally {
      submitting = false;
    }
  }

  function handleCancel() {
    goto('/clients');
  }
</script>

<svelte:head>
  <title>New Client | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">New Client</h1>
      <p class="page-description">Add a new client to your system</p>
    </div>
  </div>

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
        <Button variant="secondary" on:click={handleCancel} disabled={submitting}>
          Cancel
        </Button>
        <Button variant="primary" type="submit" loading={submitting}>
          {submitting ? 'Creating...' : 'Create Client'}
        </Button>
      </div>
    </form>
  </Card>
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
