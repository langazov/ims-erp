<script lang="ts">
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import { createUser } from '$lib/shared/api/users';
  import type { UserRole } from '$lib/shared/api/users';

  let name = '';
  let email = '';
  let role: UserRole = 'user';
  let password = '';
  let confirmPassword = '';

  let errors: Record<string, string> = {};
  let submitting = false;
  let submitError: string | null = null;

  const roleOptions = [
    { value: 'admin', label: 'Admin' },
    { value: 'manager', label: 'Manager' },
    { value: 'user', label: 'User' },
    { value: 'viewer', label: 'Viewer' }
  ];

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

    if (!password) {
      errors.password = 'Password is required';
    } else if (password.length < 8) {
      errors.password = 'Password must be at least 8 characters';
    }

    if (password !== confirmPassword) {
      errors.confirmPassword = 'Passwords do not match';
    }

    return Object.keys(errors).length === 0;
  }

  async function handleSubmit() {
    if (!validateForm()) {
      return;
    }

    submitting = true;
    try {
      const user = await createUser({
        email: email.trim(),
        name: name.trim(),
        role
      });
      goto(`/users/${user.id}`);
    } catch (error) {
      console.error('Failed to create user:', error);
      submitError = error instanceof Error ? error.message : 'Failed to create user. Please try again.';
    } finally {
      submitting = false;
    }
  }

  function handleCancel() {
    goto('/users');
  }
</script>

<svelte:head>
  <title>New User | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">New User</h1>
      <p class="page-description">Add a new user to your system</p>
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
              disabled={submitting}
            />
          </div>
          <div class="form-item full-width">
            <Select
              id="role"
              label="Role"
              options={roleOptions}
              bind:value={role}
              required
              disabled={submitting}
            />
          </div>
        </div>
      </div>

      <div class="form-section">
        <h2 class="section-title">Security</h2>
        <div class="form-grid">
          <div class="form-item full-width">
            <Input
              id="password"
              label="Password"
              type="password"
              placeholder="Enter password"
              bind:value={password}
              required
              error={errors.password}
              disabled={submitting}
              helpText="Must be at least 8 characters"
            />
          </div>
          <div class="form-item full-width">
            <Input
              id="confirmPassword"
              label="Confirm Password"
              type="password"
              placeholder="Confirm password"
              bind:value={confirmPassword}
              required
              error={errors.confirmPassword}
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
          {submitting ? 'Creating...' : 'Create User'}
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
    grid-template-columns: 1fr;
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
    .form-actions {
      flex-direction: column-reverse;
    }
  }
</style>
