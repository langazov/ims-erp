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
  import Avatar from '$lib/shared/components/display/Avatar.svelte';

  interface User {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
    role: 'admin' | 'manager' | 'user' | 'viewer';
    status: 'active' | 'inactive' | 'locked' | 'pending';
    department: string;
    lastLoginAt: string | null;
    createdAt: string;
    mfaEnabled: boolean;
  }

  let users: User[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter: string = '';
  let roleFilter: string = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;
  let showCreateModal = false;
  let deleteUserId: string | null = null;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'locked', label: 'Locked' },
    { value: 'pending', label: 'Pending' }
  ];

  const roleOptions = [
    { value: '', label: 'All Roles' },
    { value: 'admin', label: 'Admin' },
    { value: 'manager', label: 'Manager' },
    { value: 'user', label: 'User' },
    { value: 'viewer', label: 'Viewer' }
  ];

  const columns = [
    { key: 'user', label: 'User', sortable: true },
    { key: 'email', label: 'Email', sortable: true },
    { key: 'role', label: 'Role', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'department', label: 'Department', sortable: true },
    { key: 'lastLogin', label: 'Last Login', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  function getStatusVariant(status: string): 'green' | 'gray' | 'yellow' | 'red' {
    switch (status) {
      case 'active': return 'green';
      case 'pending': return 'yellow';
      case 'locked': return 'red';
      case 'inactive': return 'gray';
      default: return 'gray';
    }
  }

  function getRoleVariant(role: string): 'purple' | 'blue' | 'gray' | 'orange' {
    switch (role) {
      case 'admin': return 'purple';
      case 'manager': return 'blue';
      case 'user': return 'gray';
      case 'viewer': return 'orange';
      default: return 'gray';
    }
  }

  function formatDate(date: string | null): string {
    if (!date) return 'Never';
    return new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  async function loadUsers() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      users = [
        {
          id: '1',
          email: 'john.doe@example.com',
          firstName: 'John',
          lastName: 'Doe',
          role: 'admin',
          status: 'active',
          department: 'IT',
          lastLoginAt: '2024-01-15T10:30:00Z',
          createdAt: '2023-06-01',
          mfaEnabled: true
        },
        {
          id: '2',
          email: 'jane.smith@example.com',
          firstName: 'Jane',
          lastName: 'Smith',
          role: 'manager',
          status: 'active',
          department: 'Sales',
          lastLoginAt: '2024-01-14T16:45:00Z',
          createdAt: '2023-07-15',
          mfaEnabled: true
        },
        {
          id: '3',
          email: 'bob.wilson@example.com',
          firstName: 'Bob',
          lastName: 'Wilson',
          role: 'user',
          status: 'active',
          department: 'Warehouse',
          lastLoginAt: '2024-01-10T09:15:00Z',
          createdAt: '2023-08-20',
          mfaEnabled: false
        },
        {
          id: '4',
          email: 'alice.brown@example.com',
          firstName: 'Alice',
          lastName: 'Brown',
          role: 'viewer',
          status: 'pending',
          department: 'Finance',
          lastLoginAt: null,
          createdAt: '2024-01-05',
          mfaEnabled: false
        }
      ];
      
      totalItems = users.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load users';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadUsers();
  }

  function handleRowClick(user: User) {
    window.location.href = `/users/${user.id}`;
  }

  function handleEdit(user: User, event: Event) {
    event.stopPropagation();
    window.location.href = `/users/${user.id}/edit`;
  }

  async function handleDelete(user: User, event: Event) {
    event.stopPropagation();
    deleteUserId = user.id;
  }

  async function confirmDelete() {
    if (!deleteUserId) return;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      users = users.filter(u => u.id !== deleteUserId);
      totalItems = users.length;
      deleteUserId = null;
    } catch (err) {
      error = 'Failed to delete user';
    }
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadUsers();
  }

  onMount(() => {
    loadUsers();
  });
</script>

<svelte:head>
  <title>Users | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Users</h1>
      <p class="page-description">Manage system users and permissions</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={() => showCreateModal = true}>
        Add User
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
            placeholder="Search users..."
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
            id="role"
            label="Role"
            options={roleOptions}
            bind:value={roleFilter}
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
        <p>Loading users...</p>
      </div>
    {:else if users.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
        <p class="text-gray-500 mb-4">No users found</p>
        <Button variant="primary" on:click={() => showCreateModal = true}>
          Add Your First User
        </Button>
      </div>
    {:else}
      <Table {columns}>
        <tbody>
          {#each users as user}
            <tr on:click={() => handleRowClick(user)} class="clickable-row">
              <td>
                <div class="flex items-center gap-3">
                  <Avatar fallback={`${user.firstName} ${user.lastName}`} size="sm" />
                  <div>
                    <div class="font-medium">{user.firstName} {user.lastName}</div>
                    {#if user.mfaEnabled}
                      <span class="text-xs text-green-600">MFA Enabled</span>
                    {/if}
                  </div>
                </div>
              </td>
              <td>{user.email}</td>
              <td>
                <Badge variant={getRoleVariant(user.role)}>
                  {user.role}
                </Badge>
              </td>
              <td>
                <Badge variant={getStatusVariant(user.status)}>
                  {user.status}
                </Badge>
              </td>
              <td class="capitalize">{user.department}</td>
              <td class="text-sm text-gray-500">{formatDate(user.lastLoginAt)}</td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={(e) => handleEdit(user, e)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={(e) => handleDelete(user, e)}>
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
  title="Create User"
  size="lg"
>
  <p class="text-gray-600">User creation form will be implemented here.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close}>Cancel</Button>
    <Button variant="primary" on:click={() => {
      showCreateModal = false;
    }}>Create</Button>
  </svelte:fragment>
</Modal>

{#if deleteUserId}
  <Modal
    open={true}
    title="Delete User"
    size="sm"
  >
    <p>Are you sure you want to delete this user? This action cannot be undone.</p>
    
    <svelte:fragment slot="footer" let:close>
      <Button variant="secondary" on:click={() => { close(); deleteUserId = null; }}>Cancel</Button>
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
