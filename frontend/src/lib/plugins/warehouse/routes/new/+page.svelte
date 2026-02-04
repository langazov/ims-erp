<script lang="ts">
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';

  let code = '';
  let name = '';
  let type: 'main' | 'distribution' | 'retail' | 'virtual' = 'main';
  let capacity = '';
  
  // Address fields
  let street = '';
  let city = '';
  let state = '';
  let postalCode = '';
  let country = '';

  let errors: Record<string, string> = {};
  let submitting = false;
  let error: string | null = null;

  const typeOptions = [
    { value: 'main', label: 'Main' },
    { value: 'distribution', label: 'Distribution' },
    { value: 'retail', label: 'Retail' },
    { value: 'virtual', label: 'Virtual' }
  ];

  function validateForm(): boolean {
    errors = {};

    if (!code.trim()) {
      errors.code = 'Warehouse code is required';
    } else if (!/^[A-Z0-9-]+$/i.test(code)) {
      errors.code = 'Code must contain only letters, numbers, and hyphens';
    }

    if (!name.trim()) {
      errors.name = 'Name is required';
    }

    if (!type) {
      errors.type = 'Type is required';
    }

    if (capacity && (isNaN(parseFloat(capacity)) || parseFloat(capacity) < 0)) {
      errors.capacity = 'Capacity must be a positive number';
    }

    return Object.keys(errors).length === 0;
  }

  async function handleSubmit() {
    if (!validateForm()) return;

    const data = {
      code: code.toUpperCase(),
      name,
      type,
      capacity: capacity ? parseFloat(capacity) : undefined,
      address: street || city || state || postalCode || country
        ? {
            street,
            city,
            state,
            postalCode,
            country
          }
        : undefined
    };

    submitting = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 800));
      console.log('Creating warehouse:', data);
      goto('/warehouse');
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to create warehouse';
    } finally {
      submitting = false;
    }
  }

  function handleCancel() {
    goto('/warehouse');
  }
</script>

<svelte:head>
  <title>New Warehouse | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">New Warehouse</h1>
      <p class="page-description">Create a new warehouse location</p>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  <Card>
    <form on:submit|preventDefault={handleSubmit}>
      <div class="form-section">
        <h2 class="section-title">Warehouse Information</h2>
        <div class="form-grid">
          <div class="form-item">
            <Input
              id="code"
              label="Warehouse Code"
              type="text"
              placeholder="e.g., WH001"
              bind:value={code}
              required
              error={errors.code}
              helpText="Unique identifier for the warehouse"
            />
          </div>
          <div class="form-item">
            <Input
              id="name"
              label="Warehouse Name"
              type="text"
              placeholder="Enter warehouse name"
              bind:value={name}
              required
              error={errors.name}
            />
          </div>
          <div class="form-item">
            <Select
              id="type"
              label="Warehouse Type"
              options={typeOptions}
              bind:value={type}
              required
            />
          </div>
          <div class="form-item">
            <Input
              id="capacity"
              label="Capacity (sq ft)"
              type="number"
              placeholder="Enter capacity"
              bind:value={capacity}
              min="0"
              step="0.01"
              error={errors.capacity}
            />
          </div>
        </div>
      </div>

      <div class="form-section">
        <h2 class="section-title">Address</h2>
        <div class="form-grid">
          <div class="form-item full-width">
            <Input
              id="street"
              label="Street"
              type="text"
              placeholder="Enter street address"
              bind:value={street}
            />
          </div>
          <div class="form-item">
            <Input
              id="city"
              label="City"
              type="text"
              placeholder="Enter city"
              bind:value={city}
            />
          </div>
          <div class="form-item">
            <Input
              id="state"
              label="State/Province"
              type="text"
              placeholder="Enter state"
              bind:value={state}
            />
          </div>
          <div class="form-item">
            <Input
              id="postalCode"
              label="Postal Code"
              type="text"
              placeholder="Enter postal code"
              bind:value={postalCode}
            />
          </div>
          <div class="form-item">
            <Input
              id="country"
              label="Country"
              type="text"
              placeholder="Enter country"
              bind:value={country}
            />
          </div>
        </div>
      </div>

      <div class="form-actions">
        <Button variant="secondary" on:click={handleCancel} disabled={submitting}>
          Cancel
        </Button>
        <Button variant="primary" type="submit" loading={submitting}>
          {submitting ? 'Creating...' : 'Create Warehouse'}
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
