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
  import Avatar from '$lib/shared/components/display/Avatar.svelte';
  import { getUserById, deleteUser } from '$lib/shared/api/users';
  import type { User, UserRole, UserStatus } from '$lib/shared/api/users';

  const userId = $page.params.id;

  let user: User | null = null;
  let loading = true;
  let error: string | null = null;
  let showDeleteModal = false;
  let deleting = false;

  async function loadUser() {
    loading = true;
    error = null;
    
    try {
      user = await getUserById(userId);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load user';
    } finally {
      loading = false;
    }
  }

  function getStatusVariant(status: UserStatus): 'green' | 'gray' | 'yellow' | 'red' {
    switch (status) {
      case 'active': return 'green';
      case 'inactive': return 'gray';
      case 'suspended': return 'red';
      default: return 'gray';
    }
  }

  function getRoleVariant(role: UserRole): 'purple' | 'blue' | 'gray' | 'orange' {
    switch (role) {
      case 'admin': return 'purple';
      case 'manager': return 'blue';
      case 'user': return 'gray';
      case 'viewer': return 'orange';
      default: return 'gray';
    }
  }

  function formatDate(dateStr: string | null): string {
    if (!dateStr) return 'Never';
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function handleEdit() {
    goto(`/users/${userId}/edit`);
  }

  function handleDelete() {
    showDeleteModal = true;
  }

  async function confirmDelete() {
    deleting = true;
    try {
      await deleteUser(userId);
      goto('/users');
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete user';
      deleting = false;
      showDeleteModal = false;
    }
  }

  function handleBack() {
    goto('/users');
  }

  onMount(() => {
    loadUser();
  });
</script>

<svelte:head>
  <title>{user ? user.name : 'User Details'} | ERP System</title>
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
      <Button variant="secondary" on:click={handleBack}>
        Back to Users
      </Button>
    </div>
  {:else if user}
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <Avatar fallback={user.name} size="lg" />
          <div class="header-text">
            <h1 class="page-title">{user.name}</h1>
            <div class="header-badges">
              <Badge variant={getRoleVariant(user.role)} size="md">
                {user.role}
              </Badge>
              <Badge variant={getStatusVariant(user.status)} size="md">
                {user.status}
              </Badge>
            </div>
          </div>
        </div>
        <p class="page-description">{user.email}</p>
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

    <div class="details-grid">
      <Card>
        <h2 class="section-title">User Information</h2>
        <dl class="info-list">
          <div class="info-item">
            <dt>Full Name</dt>
            <dd>{user.name}</dd>
          </div>
          <div class="info-item">
            <dt>Email</dt>
            <dd>{user.email}</dd>
          </div>
          <div class="info-item">
            <dt>Role</dt>
            <dd>
              <Badge variant={getRoleVariant(user.role)}>
                {user.role}
              </Badge>
            </dd>
          </div>
          <div class="info-item">
            <dt>Status</dt>
            <dd>
              <Badge variant={getStatusVariant(user.status)}>
                {user.status}
              </Badge>
            </dd>
          </div>
        </dl>
      </Card>

      <Card>
        <h2 class="section-title">Activity</h2>
        <dl class="info-list">
          <div class="info-item">
            <dt>Last Login</dt>
            <dd>{formatDate(user.lastLogin)}</dd>
          </div>
          <div class="info-item">
            <dt>Created</dt>
            <dd>{formatDate(user.createdAt)}</dd>
          </div>
          <div class="info-item">
            <dt>Last Updated</dt>
            <dd>{formatDate(user.updatedAt)}</dd>
          </div>
        </dl>
      </Card>

      <Card class="full-width">
        <h2 class="section-title">System Information</h2>
        <dl class="info-list">
          <div class="info-item">
            <dt>User ID</dt>
            <dd class="mono">{user.id}</dd>
          </div>
          <div class="info-item">
            <dt>Tenant ID</dt>
            <dd class="mono">{user.tenantId}</dd>
          </div>
          <div class="info-item">
            <dt>Avatar</dt>
            <dd>
              {#if user.avatar}
                <span class="text-green-600">Uploaded</span>
              {:else}
                <span class="text-gray-500">Not set</span>
              {/if}
            </dd>
          </div>
        </dl>
      </Card>
    </div>
  {:else}
    <Alert variant="error">User not found</Alert>
  {/if}
</div>

<Modal
  bind:open={showDeleteModal}
  title="Delete User"
  size="sm"
>
  <p>Are you sure you want to delete <strong>{user?.name}</strong>? This action cannot be undone.</p>
  
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
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
  }

  .header-content {
    flex: 1;
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 0.5rem;
  }

  .header-text {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .header-badges {
    display: flex;
    gap: 0.5rem;
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

  .header-actions {
    display: flex;
    gap: 0.5rem;
  }

  .details-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .details-grid :global(.full-width) {
    grid-column: 1 / -1;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .info-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .info-item {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--color-gray-100);
  }

  .info-item:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }

  .info-item dt {
    color: var(--color-gray-500);
    font-size: 0.875rem;
  }

  .info-item dd {
    color: var(--color-gray-900);
    font-weight: 500;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .info-item dd.mono {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 0.875rem;
    color: var(--color-gray-600);
  }

  @media (max-width: 768px) {
    .details-grid {
      grid-template-columns: 1fr;
    }

    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .header-actions {
      width: 100%;
      justify-content: flex-end;
    }
  }
</style>
