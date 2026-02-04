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
  import Table from '$lib/shared/components/data/Table.svelte';
  import Pagination from '$lib/shared/components/data/Pagination.svelte';
  import { Plus, ArrowLeft, Package, ArrowRightLeft, ArrowUpRight } from 'lucide-svelte';

  const warehouseId = $page.params.id;

  interface Operation {
    id: string;
    type: 'receipt' | 'issue' | 'transfer' | 'adjustment' | 'return';
    reference: string;
    referenceType: string;
    status: 'pending' | 'in_progress' | 'completed' | 'cancelled';
    itemCount: number;
    startedAt?: string;
    completedAt?: string;
    createdAt: string;
    createdBy: string;
  }

  let operations: Operation[] = [];
  let loading = true;
  let error: string | null = null;
  let currentPage = 1;
  let totalPages = 1;
  let selectedOperation: Operation | null = null;
  let showActionModal = false;
  let actionType: 'start' | 'complete' | 'cancel' | null = null;
  let processing = false;

  const columns = [
    { key: 'type', label: 'Type', sortable: true },
    { key: 'reference', label: 'Reference', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'itemCount', label: 'Items', sortable: true },
    { key: 'createdAt', label: 'Created', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false },
  ];

  async function loadOperations() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      operations = [
        {
          id: 'op-001',
          type: 'receipt',
          reference: 'PO-2024-001',
          referenceType: 'purchase_order',
          status: 'completed',
          itemCount: 25,
          startedAt: '2024-01-15T10:00:00Z',
          completedAt: '2024-01-15T14:30:00Z',
          createdAt: '2024-01-15T09:00:00Z',
          createdBy: 'John Doe'
        },
        {
          id: 'op-002',
          type: 'issue',
          reference: 'SO-2024-045',
          referenceType: 'sales_order',
          status: 'in_progress',
          itemCount: 12,
          startedAt: '2024-01-16T08:00:00Z',
          createdAt: '2024-01-16T07:30:00Z',
          createdBy: 'Jane Smith'
        },
        {
          id: 'op-003',
          type: 'transfer',
          reference: 'TF-2024-012',
          referenceType: 'transfer_order',
          status: 'pending',
          itemCount: 50,
          createdAt: '2024-01-16T10:00:00Z',
          createdBy: 'Mike Johnson'
        },
        {
          id: 'op-004',
          type: 'adjustment',
          reference: 'ADJ-2024-003',
          referenceType: 'adjustment',
          status: 'completed',
          itemCount: 5,
          startedAt: '2024-01-14T11:00:00Z',
          completedAt: '2024-01-14T11:30:00Z',
          createdAt: '2024-01-14T10:45:00Z',
          createdBy: 'Sarah Wilson'
        },
        {
          id: 'op-005',
          type: 'return',
          reference: 'RET-2024-008',
          referenceType: 'return_order',
          status: 'pending',
          itemCount: 8,
          createdAt: '2024-01-17T09:00:00Z',
          createdBy: 'Tom Brown'
        }
      ];
      totalPages = 3;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load operations';
    } finally {
      loading = false;
    }
  }

  function getTypeIcon(type: string) {
    switch (type) {
      case 'receipt':
        return Package;
      case 'issue':
        return ArrowUpRight;
      case 'transfer':
        return ArrowRightLeft;
      default:
        return Package;
    }
  }

  function getTypeLabel(type: string): string {
    const labels: Record<string, string> = {
      receipt: 'Receipt',
      issue: 'Issue',
      transfer: 'Transfer',
      adjustment: 'Adjustment',
      return: 'Return'
    };
    return labels[type] || type;
  }

  function getStatusVariant(status: string): 'green' | 'yellow' | 'blue' | 'gray' | 'red' {
    switch (status) {
      case 'completed':
        return 'green';
      case 'in_progress':
        return 'blue';
      case 'pending':
        return 'yellow';
      case 'cancelled':
        return 'red';
      default:
        return 'gray';
    }
  }

  function formatDate(dateStr?: string): string {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function handleCreateOperation() {
    goto(`/warehouse/${warehouseId}/operations/new`);
  }

  function handleBack() {
    goto(`/warehouse/${warehouseId}`);
  }

  function handleViewOperation(operation: Operation) {
    goto(`/warehouse/${warehouseId}/operations/${operation.id}`);
  }

  function handleAction(operation: Operation, action: 'start' | 'complete' | 'cancel') {
    selectedOperation = operation;
    actionType = action;
    showActionModal = true;
  }

  async function confirmAction() {
    if (!selectedOperation || !actionType) return;
    
    processing = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      const operationIndex = operations.findIndex(op => op.id === selectedOperation!.id);
      if (operationIndex !== -1) {
        const updatedOperation = { ...operations[operationIndex] };
        
        switch (actionType) {
          case 'start':
            updatedOperation.status = 'in_progress';
            updatedOperation.startedAt = new Date().toISOString();
            break;
          case 'complete':
            updatedOperation.status = 'completed';
            updatedOperation.completedAt = new Date().toISOString();
            break;
          case 'cancel':
            updatedOperation.status = 'cancelled';
            break;
        }
        
        operations[operationIndex] = updatedOperation;
        operations = [...operations];
      }
      
      showActionModal = false;
      selectedOperation = null;
      actionType = null;
    } catch (err) {
      error = `Failed to ${actionType} operation`;
    } finally {
      processing = false;
    }
  }

  function getActionTitle(): string {
    switch (actionType) {
      case 'start':
        return 'Start Operation';
      case 'complete':
        return 'Complete Operation';
      case 'cancel':
        return 'Cancel Operation';
      default:
        return 'Confirm Action';
    }
  }

  function getActionMessage(): string {
    if (!selectedOperation) return '';
    switch (actionType) {
      case 'start':
        return `Are you sure you want to start operation ${selectedOperation.reference}?`;
      case 'complete':
        return `Are you sure you want to complete operation ${selectedOperation.reference}?`;
      case 'cancel':
        return `Are you sure you want to cancel operation ${selectedOperation.reference}? This action cannot be undone.`;
      default:
        return '';
    }
  }

  onMount(() => {
    loadOperations();
  });
</script>

<svelte:head>
  <title>Warehouse Operations | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <Button variant="ghost" on:click={handleBack} class="back-button">
        <ArrowLeft class="w-4 h-4 mr-2" />
        Back to Warehouse
      </Button>
      <h1 class="page-title">Warehouse Operations</h1>
      <p class="page-description">Manage inventory operations for this warehouse</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={handleCreateOperation}>
        <Plus class="w-4 h-4 mr-2" />
        New Operation
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
        <p>Loading operations...</p>
      </div>
    {:else}
      <Table {columns} data={operations}>
        <svelte:fragment slot="cell" let:column let:row>
          {#if column.key === 'type'}
            <div class="flex items-center gap-2">
              <svelte:component this={getTypeIcon(row.type)} class="w-4 h-4 text-gray-500" />
              <span>{getTypeLabel(row.type)}</span>
            </div>
          {:else if column.key === 'status'}
            <Badge variant={getStatusVariant(row.status)} size="sm">
              {row.status.replace('_', ' ')}
            </Badge>
          {:else if column.key === 'createdAt'}
            {formatDate(row.createdAt)}
          {:else if column.key === 'actions'}
            <div class="flex items-center gap-2">
              <Button
                variant="ghost"
                size="sm"
                on:click={() => handleViewOperation(row)}
              >
                View
              </Button>
              {#if row.status === 'pending'}
                <Button
                  variant="ghost"
                  size="sm"
                  on:click={() => handleAction(row, 'start')}
                >
                  Start
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  class="text-red-600 hover:text-red-700"
                  on:click={() => handleAction(row, 'cancel')}
                >
                  Cancel
                </Button>
              {:else if row.status === 'in_progress'}
                <Button
                  variant="ghost"
                  size="sm"
                  on:click={() => handleAction(row, 'complete')}
                >
                  Complete
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  class="text-red-600 hover:text-red-700"
                  on:click={() => handleAction(row, 'cancel')}
                >
                  Cancel
                </Button>
              {/if}
            </div>
          {:else}
            {row[column.key]}
          {/if}
        </svelte:fragment>
      </Table>

      <div class="pagination-container">
        <Pagination
          {currentPage}
          {totalPages}
          on:pageChange={(e) => {
            currentPage = e.detail;
            loadOperations();
          }}
        />
      </div>
    {/if}
  </Card>
</div>

<Modal
  bind:open={showActionModal}
  title={getActionTitle()}
  size="sm"
>
  <p>{getActionMessage()}</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showActionModal = false; }} disabled={processing}>
      Cancel
    </Button>
    <Button
      variant={actionType === 'cancel' ? 'danger' : 'primary'}
      on:click={confirmAction}
      loading={processing}
    >
      {processing ? 'Processing...' : 'Confirm'}
    </Button>
  </svelte:fragment>
</Modal>

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

  .header-content {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .back-button {
    align-self: flex-start;
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

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .pagination-container {
    display: flex;
    justify-content: center;
    padding: 1rem;
    border-top: 1px solid var(--color-gray-200);
  }

  @media (max-width: 768px) {
    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .header-actions {
      width: 100%;
    }
  }
</style>

