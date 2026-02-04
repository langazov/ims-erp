<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import Checkbox from '$lib/shared/components/forms/Checkbox.svelte';

  interface Permission {
    id: string;
    name: string;
    description: string;
  }

  interface Role {
    id: string;
    name: string;
    description: string;
    permissions: string[];
    userCount: number;
    isSystem: boolean;
  }

  let roles: Role[] = [];
  let permissions: Permission[] = [];
  let loading = true;
  let error: string | null = null;
  let saving = false;

  // Modals
  let showCreateModal = false;
  let showEditModal = false;
  let showDeleteModal = false;
  let selectedRole: Role | null = null;

  // Form fields
  let roleName = '';
  let roleDescription = '';
  let selectedPermissions: string[] = [];

  const availablePermissions: Permission[] = [
    { id: 'users.view', name: 'View Users', description: 'Can view user list and details' },
    { id: 'users.create', name: 'Create Users', description: 'Can create new users' },
    { id: 'users.edit', name: 'Edit Users', description: 'Can edit user information' },
    { id: 'users.delete', name: 'Delete Users', description: 'Can delete users' },
    { id: 'roles.manage', name: 'Manage Roles', description: 'Can manage roles and permissions' },
    { id: 'clients.view', name: 'View Clients', description: 'Can view client information' },
    { id: 'clients.create', name: 'Create Clients', description: 'Can create new clients' },
    { id: 'clients.edit', name: 'Edit Clients', description: 'Can edit client information' },
    { id: 'clients.delete', name: 'Delete Clients', description: 'Can delete clients' },
    { id: 'products.view', name: 'View Products', description: 'Can view product catalog' },
    { id: 'products.create', name: 'Create Products', description: 'Can create new products' },
    { id: 'products.edit', name: 'Edit Products', description: 'Can edit product information' },
    { id: 'products.delete', name: 'Delete Products', description: 'Can delete products' },
    { id: 'inventory.view', name: 'View Inventory', description: 'Can view inventory levels' },
    { id: 'inventory.manage', name: 'Manage Inventory', description: 'Can manage stock and inventory' },
    { id: 'warehouse.view', name: 'View Warehouses', description: 'Can view warehouse information' },
    { id: 'warehouse.manage', name: 'Manage Warehouses', description: 'Can manage warehouses' },
    { id: 'invoices.view', name: 'View Invoices', description: 'Can view invoices' },
    { id: 'invoices.create', name: 'Create Invoices', description: 'Can create invoices' },
    { id: 'invoices.edit', name: 'Edit Invoices', description: 'Can edit invoices' },
    { id: 'invoices.delete', name: 'Delete Invoices', description: 'Can delete invoices' },
    { id: 'orders.view', name: 'View Orders', description: 'Can view orders' },
    { id: 'orders.create', name: 'Create Orders', description: 'Can create orders' },
    { id: 'orders.manage', name: 'Manage Orders', description: 'Can manage and fulfill orders' },
    { id: 'payments.view', name: 'View Payments', description: 'Can view payment information' },
    { id: 'payments.process', name: 'Process Payments', description: 'Can process payments' },
    { id: 'payments.refund', name: 'Process Refunds', description: 'Can process refunds' },
    { id: 'documents.view', name: 'View Documents', description: 'Can view documents' },
    { id: 'documents.manage', name: 'Manage Documents', description: 'Can upload and manage documents' },
    { id: 'settings.view', name: 'View Settings', description: 'Can view system settings' },
    { id: 'settings.edit', name: 'Edit Settings', description: 'Can modify system settings' },
    { id: 'reports.view', name: 'View Reports', description: 'Can view reports' },
    { id: 'reports.export', name: 'Export Reports', description: 'Can export reports' }
  ];

  async function loadRoles() {
    loading = true;
    error = null;
    
    try {
      // Mock data for roles
      await new Promise(resolve => setTimeout(resolve, 500));
      roles = [
        {
          id: 'admin',
          name: 'Admin',
          description: 'Full system access with all permissions',
          permissions: availablePermissions.map(p => p.id),
          userCount: 3,
          isSystem: true
        },
        {
          id: 'manager',
          name: 'Manager',
          description: 'Can manage most aspects of the system',
          permissions: [
            'users.view', 'clients.view', 'clients.create', 'clients.edit',
            'products.view', 'products.create', 'products.edit',
            'inventory.view', 'inventory.manage',
            'warehouse.view', 'warehouse.manage',
            'invoices.view', 'invoices.create', 'invoices.edit',
            'orders.view', 'orders.create', 'orders.manage',
            'payments.view', 'payments.process',
            'documents.view', 'documents.manage',
            'reports.view', 'reports.export'
          ],
          userCount: 8,
          isSystem: true
        },
        {
          id: 'user',
          name: 'User',
          description: 'Standard user with limited access',
          permissions: [
            'users.view',
            'clients.view', 'clients.create', 'clients.edit',
            'products.view',
            'inventory.view',
            'warehouse.view',
            'invoices.view', 'invoices.create',
            'orders.view', 'orders.create',
            'payments.view',
            'documents.view'
          ],
          userCount: 24,
          isSystem: true
        },
        {
          id: 'viewer',
          name: 'Viewer',
          description: 'Read-only access to view information',
          permissions: [
            'users.view',
            'clients.view',
            'products.view',
            'inventory.view',
            'warehouse.view',
            'invoices.view',
            'orders.view',
            'payments.view',
            'documents.view',
            'reports.view'
          ],
          userCount: 12,
          isSystem: true
        }
      ];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load roles';
    } finally {
      loading = false;
    }
  }

  function getRoleBadgeVariant(roleId: string): 'purple' | 'blue' | 'gray' | 'orange' {
    switch (roleId) {
      case 'admin': return 'purple';
      case 'manager': return 'blue';
      case 'user': return 'gray';
      case 'viewer': return 'orange';
      default: return 'gray';
    }
  }

  function openCreateModal() {
    roleName = '';
    roleDescription = '';
    selectedPermissions = [];
    showCreateModal = true;
  }

  function openEditModal(role: Role) {
    selectedRole = role;
    roleName = role.name;
    roleDescription = role.description;
    selectedPermissions = [...role.permissions];
    showEditModal = true;
  }

  function openDeleteModal(role: Role) {
    selectedRole = role;
    showDeleteModal = true;
  }

  function togglePermission(permissionId: string) {
    if (selectedPermissions.includes(permissionId)) {
      selectedPermissions = selectedPermissions.filter(p => p !== permissionId);
    } else {
      selectedPermissions = [...selectedPermissions, permissionId];
    }
  }

  function selectAllPermissions() {
    selectedPermissions = availablePermissions.map(p => p.id);
  }

  function deselectAllPermissions() {
    selectedPermissions = [];
  }

  async function handleCreateRole() {
    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      const newRole: Role = {
        id: `role-${Date.now()}`,
        name: roleName,
        description: roleDescription,
        permissions: selectedPermissions,
        userCount: 0,
        isSystem: false
      };
      roles = [...roles, newRole];
      showCreateModal = false;
    } catch (err) {
      error = 'Failed to create role';
    } finally {
      saving = false;
    }
  }

  async function handleUpdateRole() {
    if (!selectedRole) return;
    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      roles = roles.map(r =>
        r.id === selectedRole?.id
          ? { ...r, name: roleName, description: roleDescription, permissions: selectedPermissions }
          : r
      );
      showEditModal = false;
      selectedRole = null;
    } catch (err) {
      error = 'Failed to update role';
    } finally {
      saving = false;
    }
  }

  async function handleDeleteRole() {
    if (!selectedRole) return;
    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      if (selectedRole.userCount > 0) {
        error = `Cannot delete role with ${selectedRole.userCount} users. Please reassign users first.`;
        showDeleteModal = false;
        selectedRole = null;
        saving = false;
        return;
      }
      roles = roles.filter(r => r.id !== selectedRole?.id);
      showDeleteModal = false;
      selectedRole = null;
    } catch (err) {
      error = 'Failed to delete role';
    } finally {
      saving = false;
    }
  }

  function getPermissionCount(role: Role): string {
    const count = role.permissions.length;
    const total = availablePermissions.length;
    return `${count}/${total}`;
  }

  onMount(() => {
    loadRoles();
  });
</script>

<svelte:head>
  <title>Role Management | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Role Management</h1>
      <p class="page-description">Manage user roles and their permissions</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={openCreateModal}>
        Create Role
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
        <p>Loading roles...</p>
      </div>
    {:else if roles.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
        </svg>
        <p class="text-gray-500 mb-4">No roles found</p>
        <Button variant="primary" on:click={openCreateModal}>
          Create First Role
        </Button>
      </div>
    {:else}
      <div class="roles-grid">
        {#each roles as role}
          <div class="role-card">
            <div class="role-header">
              <div class="role-title">
                <h3>{role.name}</h3>
                {#if role.isSystem}
                  <Badge variant="blue" size="sm">System</Badge>
                {/if}
              </div>
              <Badge variant={getRoleBadgeVariant(role.id)}>
                {getPermissionCount(role)} permissions
              </Badge>
            </div>
            <p class="role-description">{role.description}</p>
            <div class="role-stats">
              <div class="stat">
                <span class="stat-value">{role.userCount}</span>
                <span class="stat-label">users</span>
              </div>
            </div>
            <div class="role-actions">
              <Button variant="secondary" size="sm" on:click={() => openEditModal(role)}>
                Edit Permissions
              </Button>
              {#if !role.isSystem}
                <Button variant="ghost" size="sm" on:click={() => openDeleteModal(role)}>
                  Delete
                </Button>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </Card>
</div>

<!-- Create Role Modal -->
<Modal
  bind:open={showCreateModal}
  title="Create Role"
  size="lg"
>
  <div class="modal-form">
    <div class="form-row">
      <Input
        id="roleName"
        label="Role Name"
        type="text"
        placeholder="Enter role name"
        bind:value={roleName}
        required
      />
    </div>
    <div class="form-row">
      <Input
        id="roleDescription"
        label="Description"
        type="text"
        placeholder="Enter role description"
        bind:value={roleDescription}
      />
    </div>
    <div class="form-row">
      <div class="permissions-header">
        <label class="form-label">Permissions</label>
        <div class="permission-actions">
          <Button variant="ghost" size="sm" on:click={selectAllPermissions}>
            Select All
          </Button>
          <Button variant="ghost" size="sm" on:click={deselectAllPermissions}>
            Deselect All
          </Button>
        </div>
      </div>
      <div class="permissions-grid">
        {#each availablePermissions as permission}
          <label class="permission-item">
            <input
              type="checkbox"
              checked={selectedPermissions.includes(permission.id)}
              on:change={() => togglePermission(permission.id)}
            />
            <div class="permission-info">
              <span class="permission-name">{permission.name}</span>
              <span class="permission-desc">{permission.description}</span>
            </div>
          </label>
        {/each}
      </div>
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={saving}>
      Cancel
    </Button>
    <Button variant="primary" on:click={handleCreateRole} loading={saving}>
      {saving ? 'Creating...' : 'Create Role'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Edit Role Modal -->
<Modal
  bind:open={showEditModal}
  title="Edit Role: {selectedRole?.name}"
  size="lg"
>
  <div class="modal-form">
    <div class="form-row">
      <Input
        id="editRoleName"
        label="Role Name"
        type="text"
        bind:value={roleName}
        required
        disabled={selectedRole?.isSystem}
      />
    </div>
    <div class="form-row">
      <Input
        id="editRoleDescription"
        label="Description"
        type="text"
        bind:value={roleDescription}
        disabled={selectedRole?.isSystem}
      />
    </div>
    <div class="form-row">
      <div class="permissions-header">
        <label class="form-label">Permissions</label>
        <div class="permission-actions">
          <Button variant="ghost" size="sm" on:click={selectAllPermissions}>
            Select All
          </Button>
          <Button variant="ghost" size="sm" on:click={deselectAllPermissions}>
            Deselect All
          </Button>
        </div>
      </div>
      <div class="permissions-grid">
        {#each availablePermissions as permission}
          <label class="permission-item">
            <input
              type="checkbox"
              checked={selectedPermissions.includes(permission.id)}
              on:change={() => togglePermission(permission.id)}
            />
            <div class="permission-info">
              <span class="permission-name">{permission.name}</span>
              <span class="permission-desc">{permission.description}</span>
            </div>
          </label>
        {/each}
      </div>
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={saving}>
      Cancel
    </Button>
    <Button variant="primary" on:click={handleUpdateRole} loading={saving}>
      {saving ? 'Saving...' : 'Save Changes'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Delete Role Modal -->
<Modal
  bind:open={showDeleteModal}
  title="Delete Role"
  size="sm"
>
  <p>Are you sure you want to delete the role <strong>{selectedRole?.name}</strong>?</p>
  {#if selectedRole?.userCount && selectedRole.userCount > 0}
    <p class="warning-text">
      This role is assigned to {selectedRole.userCount} users. Please reassign them before deleting.
    </p>
  {/if}
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showDeleteModal = false; }} disabled={saving}>
      Cancel
    </Button>
    <Button variant="danger" on:click={handleDeleteRole} loading={saving}>
      {saving ? 'Deleting...' : 'Delete'}
    </Button>
  </svelte:fragment>
</Modal>

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

  .header-actions {
    display: flex;
    gap: 0.5rem;
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

  .roles-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1rem;
  }

  .role-card {
    border: 1px solid var(--color-gray-200);
    border-radius: 0.75rem;
    padding: 1.5rem;
    background-color: white;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .role-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .role-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .role-title h3 {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0;
  }

  .role-description {
    color: var(--color-gray-600);
    font-size: 0.875rem;
    margin: 0;
    flex: 1;
  }

  .role-stats {
    display: flex;
    gap: 1rem;
  }

  .stat {
    display: flex;
    align-items: baseline;
    gap: 0.25rem;
  }

  .stat-value {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .stat-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .role-actions {
    display: flex;
    gap: 0.5rem;
    padding-top: 1rem;
    border-top: 1px solid var(--color-gray-100);
  }

  .modal-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .form-row {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .form-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-700);
  }

  .permissions-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .permission-actions {
    display: flex;
    gap: 0.5rem;
  }

  .permissions-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
    max-height: 400px;
    overflow-y: auto;
    padding: 0.75rem;
    border: 1px solid var(--color-gray-200);
    border-radius: 0.5rem;
    background-color: var(--color-gray-50);
  }

  .permission-item {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 0.5rem;
    border-radius: 0.375rem;
    cursor: pointer;
    transition: background-color 0.15s;
  }

  .permission-item:hover {
    background-color: white;
  }

  .permission-item input[type="checkbox"] {
    margin-top: 0.125rem;
    width: 1rem;
    height: 1rem;
    accent-color: var(--color-primary-600);
  }

  .permission-info {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .permission-name {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-900);
  }

  .permission-desc {
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  .warning-text {
    color: var(--color-yellow-600);
    font-size: 0.875rem;
    margin-top: 0.5rem;
  }

  @media (max-width: 768px) {
    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .roles-grid {
      grid-template-columns: 1fr;
    }

    .permissions-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
