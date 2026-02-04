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
  import { getUserById, updateUser } from '$lib/shared/api/users';
  import type { User, UserRole, UserStatus } from '$lib/shared/api/users';

  const userId = $page.params.id;

  let user: User | null = null;
  let loading = true;
  let error: string | null = null;
  let submitError: string | null = null;
  let submitting = false;

  // Form fields
  let name = '';
  let email = '';
  let role: UserRole = 'user';
  let status: UserStatus = 'active';

  let errors: Record<string, string> = {};

  const roleOptions = [
    { value: 'admin', label: 'Admin' },
    { value: 'manager', label: 'Manager' },
    { value: 'user', label: 'User' },
    { value: 'viewer', label: 'Viewer' }
  ];

  const statusOptions = [
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'suspended', label: 'Suspended' }
  ];

  async function loadUser() {
    loading = true;
    error = null;
    
    try {
      user = await getUserById(userId);
      // Initialize form fields
      name = user.name;
      email = user.email;
      role = user.role;
      status = user.status;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load user';
    } finally {
      loading = false;
    }
  }

  function validateForm(): boolean {
    errors = {};
    submitError = null;

    if (!name.trim()) {
      errors.name = 'Name is required';
    } else if (name.trim().length < 2) {
      errors.name = 'Name must be at least 2 characters';
    }

    if (!email.trim()) {
      errors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      errors.email = 'Invalid email format';
    }

    return Object.keys(errors).length === 0;
  }

  async function handleSubmit() {
    if (!validateForm()) {
      return;
    }

    submitting = true;
    try {
      const updateData: Partial<{ name: string; role: UserRole; status: UserStatus }> = {};
      
      if (name.trim() !== user?.name) {
        updateData.name = name.trim();
      }
      if (role !== user?.role) {
        updateData.role = role;
      }
      if (status !== user?.status) {
        updateData.status = status;
      }

      if (Object.keys(updateData).length === 0) {
        goto(`/users/${userId}`);
        return;
      }

      await updateUser(userId, updateData);
      goto(`/users/${userId}`);
    } catch (err) {
      console.error('Failed to update user:', err);
      submitError = err instanceof Error ? err.message : 'Failed to update user. Please try again.';
    } finally {
      submitting = false;
    }
  }

  function handleCancel() {
    goto(`/users/${userId}`);
  }

  onMount(() => {
    loadUser();
  });
</script>

<svelte:head>
  <title>Edit {user ? user.name : 'User'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading user details...</p>
    </div>
  {:else if error}
    <div class="error-container">
      <Alert variant="error">{error}</Alert>
      <Button variant="secondary" on:click={() => goto('/users')}>
        Back to Users
      </Button>
    </div>
  {:else if user}
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">Edit User</h1>
        <p class="page-description">Update {user.name}'s information</p>
      </div>
    </div>

    {#if submitError}
      <Alert variant="error" dismissible on:dismiss={() => submitError = null} class="mb-4">
        {submitError}
      </Alert>
    {/if}

    <Card>
      <form on:submit|preventDefault={handleSubmit}>
        <div class="form-section">
          <h2 class="section-title">Basic Information</h2>
          <div class="form-grid">
            <div class="form-item full-width">
              <Input
                id="name"
                label="Full Name"
                type="text"
                placeholder="Enter user's full name"
                bind:value={name}
                required
                error={errors.name}
                disabled={submitting}
              />
            </div>
            <div class="form-item full-width">
              <Input
                id="email"
                label="Email Address"
                type="email"
                placeholder="Enter email address"
                bind:value={email}
                required
                error={errors.email}
                disabled={true}
                helpText="Email cannot be changed"
              />
            </div>
          </div>
        </div>

        <div class="form-section">
          <h2 class="section-title">Role & Status</h2>
          <div class="form-grid two-col">
            <div class="form-item">
              <Select
                id="role"
                label="Role"
                options={roleOptions}
                bind:value={role}
                required
                disabled={submitting}
              />
            </div>
            <div class="form-item">
              <Select
                id="status"
                label="Status"
                options={statusOptions}
                bind:value={status}
                required
                disabled={submitting}
              />
            </div>
          </div>
        </div>

        <div class="form-actions">
          <Button variant="secondary" on:click={handleCancel} disabled={submitting}>
            Cancel
          </Button>
          <Button variant="primary" type="submit" loading={submitting}>
            {submitting ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </form>
    </Card>
  {:else}
    <Alert variant="error">User not found</Alert>
  {/if}
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 900px;
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
    grid-template-columns: 1fr;
    gap: 1rem;
  }

  .form-grid.two-col {
    grid-template-columns: repeat(2, 1fr);
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
    .form-grid.two-col {
      grid-template-columns: 1fr;
    }

    .form-actions {
      flex-direction: column-reverse;
    }
  }
</style>
