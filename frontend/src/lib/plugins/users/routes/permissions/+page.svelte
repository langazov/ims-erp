<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';

  interface Permission {
    id: string;
    name: string;
    description: string;
  }

  interface Module {
    id: string;
    name: string;
    icon: string;
    permissions: Permission[];
  }

  interface RolePermissions {
    [roleId: string]: string[];
  }

  const modules: Module[] = [
    {
      id: 'clients',
      name: 'Clients',
      icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z',
      permissions: [
        { id: 'clients.view', name: 'View', description: 'View client information' },
        { id: 'clients.create', name: 'Create', description: 'Create new clients' },
        { id: 'clients.edit', name: 'Edit', description: 'Edit client details' },
        { id: 'clients.delete', name: 'Delete', description: 'Delete clients' },
        { id: 'clients.export', name: 'Export', description: 'Export client data' }
      ]
    },
    {
      id: 'products',
      name: 'Products',
      icon: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
      permissions: [
        { id: 'products.view', name: 'View', description: 'View product catalog' },
        { id: 'products.create', name: 'Create', description: 'Create new products' },
        { id: 'products.edit', name: 'Edit', description: 'Edit product details' },
        { id: 'products.delete', name: 'Delete', description: 'Delete products' },
        { id: 'products.export', name: 'Export', description: 'Export product data' }
      ]
    },
    {
      id: 'inventory',
      name: 'Inventory',
      icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2',
      permissions: [
        { id: 'inventory.view', name: 'View', description: 'View inventory levels' },
        { id: 'inventory.adjust', name: 'Adjust', description: 'Adjust stock levels' },
        { id: 'inventory.transfer', name: 'Transfer', description: 'Transfer stock between locations' },
        { id: 'inventory.count', name: 'Count', description: 'Perform stock counts' },
        { id: 'inventory.export', name: 'Export', description: 'Export inventory data' }
      ]
    },
    {
      id: 'warehouse',
      name: 'Warehouse',
      icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4',
      permissions: [
        { id: 'warehouse.view', name: 'View', description: 'View warehouse information' },
        { id: 'warehouse.create', name: 'Create', description: 'Create warehouses' },
        { id: 'warehouse.edit', name: 'Edit', description: 'Edit warehouse details' },
        { id: 'warehouse.delete', name: 'Delete', description: 'Delete warehouses' },
        { id: 'warehouse.operations', name: 'Operations', description: 'Manage warehouse operations' }
      ]
    },
    {
      id: 'invoices',
      name: 'Invoices',
      icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
      permissions: [
        { id: 'invoices.view', name: 'View', description: 'View invoices' },
        { id: 'invoices.create', name: 'Create', description: 'Create invoices' },
        { id: 'invoices.edit', name: 'Edit', description: 'Edit invoices' },
        { id: 'invoices.delete', name: 'Delete', description: 'Delete invoices' },
        { id: 'invoices.export', name: 'Export', description: 'Export invoices' }
      ]
    },
    {
      id: 'orders',
      name: 'Orders',
      icon: 'M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z',
      permissions: [
        { id: 'orders.view', name: 'View', description: 'View orders' },
        { id: 'orders.create', name: 'Create', description: 'Create orders' },
        { id: 'orders.edit', name: 'Edit', description: 'Edit orders' },
        { id: 'orders.delete', name: 'Delete', description: 'Delete orders' },
        { id: 'orders.fulfill', name: 'Fulfill', description: 'Fulfill orders' }
      ]
    },
    {
      id: 'payments',
      name: 'Payments',
      icon: 'M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z',
      permissions: [
        { id: 'payments.view', name: 'View', description: 'View payments' },
        { id: 'payments.process', name: 'Process', description: 'Process payments' },
        { id: 'payments.refund', name: 'Refund', description: 'Process refunds' },
        { id: 'payments.export', name: 'Export', description: 'Export payment data' }
      ]
    },
    {
      id: 'documents',
      name: 'Documents',
      icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
      permissions: [
        { id: 'documents.view', name: 'View', description: 'View documents' },
        { id: 'documents.upload', name: 'Upload', description: 'Upload documents' },
        { id: 'documents.delete', name: 'Delete', description: 'Delete documents' },
        { id: 'documents.share', name: 'Share', description: 'Share documents' }
      ]
    },
    {
      id: 'users',
      name: 'Users',
      icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z',
      permissions: [
        { id: 'users.view', name: 'View', description: 'View users' },
        { id: 'users.create', name: 'Create', description: 'Create users' },
        { id: 'users.edit', name: 'Edit', description: 'Edit users' },
        { id: 'users.delete', name: 'Delete', description: 'Delete users' },
        { id: 'users.manage_roles', name: 'Manage Roles', description: 'Manage roles and permissions' }
      ]
    },
    {
      id: 'settings',
      name: 'Settings',
      icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
      permissions: [
        { id: 'settings.view', name: 'View', description: 'View settings' },
        { id: 'settings.edit', name: 'Edit', description: 'Edit settings' },
        { id: 'settings.backup', name: 'Backup', description: 'Manage backups' }
      ]
    }
  ];

  const roles = [
    { id: 'admin', name: 'Admin', color: 'purple' },
    { id: 'manager', name: 'Manager', color: 'blue' },
    { id: 'user', name: 'User', color: 'gray' },
    { id: 'viewer', name: 'Viewer', color: 'orange' }
  ];

  let rolePermissions: RolePermissions = {};
  let loading = true;
  let error: string | null = null;
  let saving = false;
  let hasChanges = false;
  let showSaveModal = false;

  async function loadPermissions() {
    loading = true;
    error = null;
    
    try {
      // Mock data for role permissions
      await new Promise(resolve => setTimeout(resolve, 500));
      rolePermissions = {
        admin: modules.flatMap(m => m.permissions.map(p => p.id)),
        manager: [
          'clients.view', 'clients.create', 'clients.edit',
          'products.view', 'products.create', 'products.edit',
          'inventory.view', 'inventory.adjust', 'inventory.transfer',
          'warehouse.view', 'warehouse.operations',
          'invoices.view', 'invoices.create', 'invoices.edit',
          'orders.view', 'orders.create', 'orders.edit', 'orders.fulfill',
          'payments.view', 'payments.process',
          'documents.view', 'documents.upload',
          'users.view',
          'settings.view'
        ],
        user: [
          'clients.view', 'clients.create', 'clients.edit',
          'products.view',
          'inventory.view',
          'warehouse.view',
          'invoices.view', 'invoices.create',
          'orders.view', 'orders.create',
          'payments.view',
          'documents.view', 'documents.upload'
        ],
        viewer: [
          'clients.view',
          'products.view',
          'inventory.view',
          'warehouse.view',
          'invoices.view',
          'orders.view',
          'payments.view',
          'documents.view'
        ]
      };
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load permissions';
    } finally {
      loading = false;
    }
  }

  function hasPermission(roleId: string, permissionId: string): boolean {
    return rolePermissions[roleId]?.includes(permissionId) ?? false;
  }

  function togglePermission(roleId: string, permissionId: string) {
    const currentPermissions = rolePermissions[roleId] || [];
    if (currentPermissions.includes(permissionId)) {
      rolePermissions = {
        ...rolePermissions,
        [roleId]: currentPermissions.filter(p => p !== permissionId)
      };
    } else {
      rolePermissions = {
        ...rolePermissions,
        [roleId]: [...currentPermissions, permissionId]
      };
    }
    hasChanges = true;
  }

  function toggleAllForRole(roleId: string, module: Module, checked: boolean) {
    const modulePermissions = module.permissions.map(p => p.id);
    const currentPermissions = rolePermissions[roleId] || [];
    
    if (checked) {
      // Add all module permissions
      const newPermissions = [...new Set([...currentPermissions, ...modulePermissions])];
      rolePermissions = { ...rolePermissions, [roleId]: newPermissions };
    } else {
      // Remove all module permissions
      rolePermissions = {
        ...rolePermissions,
        [roleId]: currentPermissions.filter(p => !modulePermissions.includes(p))
      };
    }
    hasChanges = true;
  }

  function isAllModulePermissions(roleId: string, module: Module): boolean {
    const modulePermissions = module.permissions.map(p => p.id);
    const currentPermissions = rolePermissions[roleId] || [];
    return modulePermissions.every(p => currentPermissions.includes(p));
  }

  function isSomeModulePermissions(roleId: string, module: Module): boolean {
    const modulePermissions = module.permissions.map(p => p.id);
    const currentPermissions = rolePermissions[roleId] || [];
    const hasSome = modulePermissions.some(p => currentPermissions.includes(p));
    const hasAll = modulePermissions.every(p => currentPermissions.includes(p));
    return hasSome && !hasAll;
  }

  async function handleSave() {
    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 800));
      hasChanges = false;
      showSaveModal = false;
    } catch (err) {
      error = 'Failed to save permissions';
    } finally {
      saving = false;
    }
  }

  function handleCancel() {
    if (hasChanges) {
      showSaveModal = true;
    }
  }

  function confirmCancel() {
    loadPermissions();
    hasChanges = false;
    showSaveModal = false;
  }

  onMount(() => {
    loadPermissions();
  });
</script>

<svelte:head>
  <title>Permission Management | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Permission Management</h1>
      <p class="page-description">Configure permissions for each role across all modules</p>
    </div>
    <div class="header-actions">
      {#if hasChanges}
        <Button variant="secondary" on:click={handleCancel}>
          Cancel
        </Button>
      {/if}
      <Button variant="primary" on:click={handleSave} loading={saving} disabled={!hasChanges}>
        {saving ? 'Saving...' : 'Save Changes'}
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  {#if hasChanges}
    <Alert variant="warning" class="mb-4">
      You have unsaved changes. Don't forget to save your modifications.
    </Alert>
  {/if}

  <Card>
    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading permissions...</p>
      </div>
    {:else}
      <div class="permissions-matrix">
        <div class="matrix-header">
          <div class="module-header">Module / Permission</div>
          {#each roles as role}
            <div class="role-header">
              <Badge variant={role.color as 'purple' | 'blue' | 'gray' | 'orange'} size="md">
                {role.name}
              </Badge>
            </div>
          {/each}
        </div>

        {#each modules as module}
          <div class="module-section">
            <div class="module-title-row">
              <div class="module-title">
                <svg class="module-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={module.icon} />
                </svg>
                <span>{module.name}</span>
              </div>
              {#each roles as role}
                <div class="role-checkbox">
                  <input
                    type="checkbox"
                    checked={isAllModulePermissions(role.id, module)}
                    indeterminate={isSomeModulePermissions(role.id, module)}
                    on:change={(e) => toggleAllForRole(role.id, module, e.currentTarget.checked)}
                    title="Toggle all {module.name} permissions for {role.name}"
                  />
                </div>
              {/each}
            </div>
            {#each module.permissions as permission}
              <div class="permission-row">
                <div class="permission-info">
                  <span class="permission-name">{permission.name}</span>
                  <span class="permission-desc">{permission.description}</span>
                </div>
                {#each roles as role}
                  <div class="role-checkbox">
                    <input
                      type="checkbox"
                      checked={hasPermission(role.id, permission.id)}
                      on:change={() => togglePermission(role.id, permission.id)}
                      title="{role.name}: {permission.name}"
                    />
                  </div>
                {/each}
              </div>
            {/each}
          </div>
        {/each}
      </div>
    {/if}
  </Card>
</div>

<Modal
  bind:open={showSaveModal}
  title="Unsaved Changes"
  size="sm"
>
  <p>You have unsaved changes. Are you sure you want to discard them?</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showSaveModal = false; }}>
      Keep Editing
    </Button>
    <Button variant="danger" on:click={confirmCancel}>
      Discard Changes
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

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .permissions-matrix {
    display: flex;
    flex-direction: column;
    border: 1px solid var(--color-gray-200);
    border-radius: 0.5rem;
    overflow: hidden;
  }

  .matrix-header {
    display: grid;
    grid-template-columns: 1fr repeat(4, 100px);
    background-color: var(--color-gray-50);
    border-bottom: 2px solid var(--color-gray-200);
    font-weight: 600;
  }

  .module-header {
    padding: 1rem;
    color: var(--color-gray-700);
  }

  .role-header {
    padding: 1rem;
    text-align: center;
    display: flex;
    align-items: center;
    justify-content: center;
    border-left: 1px solid var(--color-gray-200);
  }

  .module-section {
    border-bottom: 1px solid var(--color-gray-200);
  }

  .module-section:last-child {
    border-bottom: none;
  }

  .module-title-row {
    display: grid;
    grid-template-columns: 1fr repeat(4, 100px);
    background-color: var(--color-gray-100);
    font-weight: 600;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .module-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    color: var(--color-gray-900);
  }

  .module-icon {
    width: 1.25rem;
    height: 1.25rem;
    color: var(--color-gray-500);
  }

  .permission-row {
    display: grid;
    grid-template-columns: 1fr repeat(4, 100px);
    border-bottom: 1px solid var(--color-gray-100);
  }

  .permission-row:last-child {
    border-bottom: none;
  }

  .permission-info {
    display: flex;
    flex-direction: column;
    padding: 0.75rem 1rem;
    padding-left: 2.5rem;
  }

  .permission-name {
    font-size: 0.875rem;
    color: var(--color-gray-700);
  }

  .permission-desc {
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  .role-checkbox {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.75rem;
    border-left: 1px solid var(--color-gray-100);
  }

  .role-checkbox input[type="checkbox"] {
    width: 1.25rem;
    height: 1.25rem;
    accent-color: var(--color-primary-600);
    cursor: pointer;
  }

  @media (max-width: 1024px) {
    .matrix-header,
    .module-title-row,
    .permission-row {
      grid-template-columns: 1fr repeat(4, 80px);
    }

    .role-header,
    .role-checkbox {
      font-size: 0.75rem;
    }
  }

  @media (max-width: 768px) {
    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .permissions-matrix {
      overflow-x: auto;
    }

    .matrix-header,
    .module-title-row,
    .permission-row {
      min-width: 600px;
    }
  }
</style>
