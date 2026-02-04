<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Pagination from '$lib/shared/components/data/Pagination.svelte';
  import { Plus, CheckCircle, XCircle, AlertTriangle } from 'lucide-svelte';

  interface Adjustment {
    id: string;
    itemId: string;
    productName: string;
    sku: string;
    warehouseName: string;
    currentQuantity: number;
    adjustedQuantity: number;
    difference: number;
    reason: string;
    status: 'pending' | 'approved' | 'rejected';
    requestedBy: string;
    createdAt: string;
  }

  let adjustments: Adjustment[] = [];
  let loading = true;
  let error: string | null = null;
  let currentPage = 1;
  let totalPages = 1;
  let selectedAdjustment: Adjustment | null = null;
  let showApproveModal = false;
  let showRejectModal = false;
  let processing = false;

  const columns = [
    { key: 'productName', label: 'Product', sortable: true },
    { key: 'sku', label: 'SKU', sortable: true },
    { key: 'currentQuantity', label: 'Current', sortable: true },
    { key: 'adjustedQuantity', label: 'Adjusted', sortable: true },
    { key: 'difference', label: 'Difference', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false },
  ];

  async function loadAdjustments() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      adjustments = [
        {
          id: 'adj-001',
          itemId: 'item-001',
          productName: 'Wireless Bluetooth Headphones',
          sku: 'WBH-001-BLK',
          warehouseName: 'Main Distribution Center',
          currentQuantity: 150,
          adjustedQuantity: 148,
          difference: -2,
          reason: 'Damaged units found during cycle count',
          status: 'pending',
          requestedBy: 'John Doe',
          createdAt: '2024-01-15T10:00:00Z'
        },
        {
          id: 'adj-002',
          itemId: 'item-002',
          productName: 'USB-C Charging Cable',
          sku: 'UCC-002-WHT',
          warehouseName: 'Main Distribution Center',
          currentQuantity: 500,
          adjustedQuantity: 510,
          difference: 10,
          reason: 'Found extra units in receiving',
          status: 'approved',
          requestedBy: 'Jane Smith',
          createdAt: '2024-01-14T14:30:00Z'
        },
        {
          id: 'adj-003',
          itemId: 'item-003',
          productName: 'Mechanical Keyboard',
          sku: 'MKB-003-BLK',
          warehouseName: 'Retail Store #1',
          currentQuantity: 25,
          adjustedQuantity: 20,
          difference: -5,
          reason: 'Theft reported',
          status: 'rejected',
          requestedBy: 'Mike Johnson',
          createdAt: '2024-01-13T09:00:00Z'
        }
      ];
      totalPages = 1;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load adjustments';
    } finally {
      loading = false;
    }
  }

  function getStatusVariant(status: string): 'green' | 'yellow' | 'red' {
    switch (status) {
      case 'approved':
        return 'green';
      case 'pending':
        return 'yellow';
      case 'rejected':
        return 'red';
      default:
        return 'yellow';
    }
  }

  function handleCreateAdjustment() {
    goto('/inventory/adjustments/new');
  }

  function handleApprove(adjustment: Adjustment) {
    selectedAdjustment = adjustment;
    showApproveModal = true;
  }

  function handleReject(adjustment: Adjustment) {
    selectedAdjustment = adjustment;
    showRejectModal = true;
  }

  async function confirmApprove() {
    if (!selectedAdjustment) return;
    
    processing = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      const index = adjustments.findIndex(a => a.id === selectedAdjustment!.id);
      if (index !== -1) {
        adjustments[index] = { ...adjustments[index], status: 'approved' };
        adjustments = [...adjustments];
      }
      
      showApproveModal = false;
      selectedAdjustment = null;
    } catch (err) {
      error = 'Failed to approve adjustment';
    } finally {
      processing = false;
    }
  }

  async function confirmReject() {
    if (!selectedAdjustment) return;
    
    processing = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      const index = adjustments.findIndex(a => a.id === selectedAdjustment!.id);
      if (index !== -1) {
        adjustments[index] = { ...adjustments[index], status: 'rejected' };
        adjustments = [...adjustments];
      }
      
      showRejectModal = false;
      selectedAdjustment = null;
    } catch (err) {
      error = 'Failed to reject adjustment';
    } finally {
      processing = false;
    }
  }

  onMount(() => {
    loadAdjustments();
  });
</script>

<svelte:head>
  <title>Inventory Adjustments | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Inventory Adjustments</h1>
      <p class="page-description">Request and approve inventory quantity adjustments</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={handleCreateAdjustment}>
        <Plus class="w-4 h-4 mr-2" />
        New Adjustment
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
        <p>Loading adjustments...</p>
      </div>
    {:else}
      <Table {columns} data={adjustments}>
        <svelte:fragment slot="cell" let:column let:row>
          {#if column.key === 'productName'}
            <div class="flex flex-col">
              <span class="font-medium">{row.productName}</span>
              <span class="text-sm text-gray-500">{row.warehouseName}</span>
            </div>
          {:else if column.key === 'difference'}
            <span class={row.difference > 0 ? 'text-green-600' : 'text-red-600'}>
              {row.difference > 0 ? '+' : ''}{row.difference}
            </span>
          {:else if column.key === 'status'}
            <Badge variant={getStatusVariant(row.status)} size="sm">
              {row.status}
            </Badge>
          {:else if column.key === 'actions'}
            <div class="flex items-center gap-2">
              {#if row.status === 'pending'}
                <Button
                  variant="ghost"
                  size="sm"
                  on:click={() => handleApprove(row)}
                >
                  <CheckCircle class="w-4 h-4 mr-1 text-green-600" />
                  Approve
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  on:click={() => handleReject(row)}
                >
                  <XCircle class="w-4 h-4 mr-1 text-red-600" />
                  Reject
                </Button>
              {:else}
                <span class="text-gray-400 text-sm">-</span>
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
            loadAdjustments();
          }}
        />
      </div>
    {/if}
  </Card>
</div>

<Modal
  bind:open={showApproveModal}
  title="Approve Adjustment"
  size="sm"
>
  <div class="modal-content">
    <AlertTriangle class="w-12 h-12 text-yellow-500 mx-auto mb-4" />
    <p class="text-center">
      Are you sure you want to approve this adjustment?
    </p>
    {#if selectedAdjustment}
      <div class="adjustment-details mt-4">
        <p><strong>Product:</strong> {selectedAdjustment.productName}</p>
        <p><strong>Current:</strong> {selectedAdjustment.currentQuantity}</p>
        <p><strong>Adjusted:</strong> {selectedAdjustment.adjustedQuantity}</p>
        <p><strong>Reason:</strong> {selectedAdjustment.reason}</p>
      </div>
    {/if}
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showApproveModal = false; }} disabled={processing}>
      Cancel
    </Button>
    <Button
      variant="primary"
      on:click={confirmApprove}
      loading={processing}
    >
      {processing ? 'Approving...' : 'Approve'}
    </Button>
  </svelte:fragment>
</Modal>

<Modal
  bind:open={showRejectModal}
  title="Reject Adjustment"
  size="sm"
>
  <div class="modal-content">
    <AlertTriangle class="w-12 h-12 text-red-500 mx-auto mb-4" />
    <p class="text-center">
      Are you sure you want to reject this adjustment?
    </p>
    {#if selectedAdjustment}
      <div class="adjustment-details mt-4">
        <p><strong>Product:</strong> {selectedAdjustment.productName}</p>
        <p><strong>Current:</strong> {selectedAdjustment.currentQuantity}</p>
        <p><strong>Adjusted:</strong> {selectedAdjustment.adjustedQuantity}</p>
        <p><strong>Reason:</strong> {selectedAdjustment.reason}</p>
      </div>
    {/if}
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showRejectModal = false; }} disabled={processing}>
      Cancel
    </Button>
    <Button
      variant="danger"
      on:click={confirmReject}
      loading={processing}
    >
      {processing ? 'Rejecting...' : 'Reject'}
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

  .modal-content {
    padding: 1rem 0;
  }

  .adjustment-details {
    background-color: var(--color-gray-50);
    padding: 1rem;
    border-radius: 0.5rem;
  }

  .adjustment-details p {
    margin: 0.25rem 0;
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

