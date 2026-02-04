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
  import { Plus, Lock, Unlock, Calendar } from 'lucide-svelte';

  interface Reservation {
    id: string;
    itemId: string;
    productName: string;
    sku: string;
    quantity: number;
    reservedBy: string;
    reservedFor: string;
    expiresAt?: string;
    status: 'active' | 'fulfilled' | 'expired' | 'cancelled';
    createdAt: string;
  }

  let reservations: Reservation[] = [];
  let loading = true;
  let error: string | null = null;
  let currentPage = 1;
  let totalPages = 1;
  let selectedReservation: Reservation | null = null;
  let showReleaseModal = false;
  let releasing = false;

  const columns = [
    { key: 'productName', label: 'Product', sortable: true },
    { key: 'sku', label: 'SKU', sortable: true },
    { key: 'quantity', label: 'Quantity', sortable: true },
    { key: 'reservedFor', label: 'Reserved For', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'expiresAt', label: 'Expires', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false },
  ];

  async function loadReservations() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      reservations = [
        {
          id: 'res-001',
          itemId: 'item-001',
          productName: 'Wireless Bluetooth Headphones',
          sku: 'WBH-001-BLK',
          quantity: 25,
          reservedBy: 'John Doe',
          reservedFor: 'SO-2024-045',
          expiresAt: '2024-02-15T23:59:59Z',
          status: 'active',
          createdAt: '2024-01-15T10:00:00Z'
        },
        {
          id: 'res-002',
          itemId: 'item-002',
          productName: 'USB-C Charging Cable',
          sku: 'UCC-002-WHT',
          quantity: 50,
          reservedBy: 'Jane Smith',
          reservedFor: 'SO-2024-046',
          expiresAt: '2024-02-10T23:59:59Z',
          status: 'active',
          createdAt: '2024-01-14T14:30:00Z'
        },
        {
          id: 'res-003',
          itemId: 'item-003',
          productName: 'Mechanical Keyboard',
          sku: 'MKB-003-BLK',
          quantity: 10,
          reservedBy: 'Mike Johnson',
          reservedFor: 'SO-2024-040',
          status: 'fulfilled',
          createdAt: '2024-01-10T09:00:00Z'
        },
        {
          id: 'res-004',
          itemId: 'item-004',
          productName: 'Wireless Mouse',
          sku: 'WM-004-GRY',
          quantity: 30,
          reservedBy: 'Sarah Wilson',
          reservedFor: 'SO-2024-038',
          expiresAt: '2024-01-20T23:59:59Z',
          status: 'expired',
          createdAt: '2024-01-05T11:00:00Z'
        }
      ];
      totalPages = 2;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load reservations';
    } finally {
      loading = false;
    }
  }

  function getStatusVariant(status: string): 'green' | 'blue' | 'gray' | 'red' | 'yellow' {
    switch (status) {
      case 'active':
        return 'green';
      case 'fulfilled':
        return 'blue';
      case 'expired':
        return 'red';
      case 'cancelled':
        return 'gray';
      default:
        return 'yellow';
    }
  }

  function formatDate(dateStr?: string): string {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  function handleCreateReservation() {
    goto('/inventory/reservations/new');
  }

  function handleRelease(reservation: Reservation) {
    selectedReservation = reservation;
    showReleaseModal = true;
  }

  async function confirmRelease() {
    if (!selectedReservation) return;
    
    releasing = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      const index = reservations.findIndex(r => r.id === selectedReservation!.id);
      if (index !== -1) {
        reservations[index] = { ...reservations[index], status: 'cancelled' };
        reservations = [...reservations];
      }
      
      showReleaseModal = false;
      selectedReservation = null;
    } catch (err) {
      error = 'Failed to release reservation';
    } finally {
      releasing = false;
    }
  }

  onMount(() => {
    loadReservations();
  });
</script>

<svelte:head>
  <title>Stock Reservations | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Stock Reservations</h1>
      <p class="page-description">Manage reserved inventory for orders and transfers</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={handleCreateReservation}>
        <Plus class="w-4 h-4 mr-2" />
        New Reservation
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
        <p>Loading reservations...</p>
      </div>
    {:else}
      <Table {columns} data={reservations}>
        <svelte:fragment slot="cell" let:column let:row>
          {#if column.key === 'productName'}
            <div class="flex items-center gap-2">
              <Package class="w-4 h-4 text-gray-500" />
              <span>{row.productName}</span>
            </div>
          {:else if column.key === 'status'}
            <Badge variant={getStatusVariant(row.status)} size="sm">
              {row.status}
            </Badge>
          {:else if column.key === 'expiresAt'}
            <div class="flex items-center gap-1">
              <Calendar class="w-3 h-3 text-gray-400" />
              {formatDate(row.expiresAt)}
            </div>
          {:else if column.key === 'actions'}
            <div class="flex items-center gap-2">
              {#if row.status === 'active'}
                <Button
                  variant="ghost"
                  size="sm"
                  on:click={() => handleRelease(row)}
                >
                  <Unlock class="w-4 h-4 mr-1" />
                  Release
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
            loadReservations();
          }}
        />
      </div>
    {/if}
  </Card>
</div>

<Modal
  bind:open={showReleaseModal}
  title="Release Reservation"
  size="sm"
>
  <p>
    Are you sure you want to release the reservation for 
    <strong>{selectedReservation?.productName}</strong> 
    ({selectedReservation?.quantity} units)?
  </p>
  <p class="text-sm text-gray-500 mt-2">
    This will make the reserved stock available for other orders.
  </p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showReleaseModal = false; }} disabled={releasing}>
      Cancel
    </Button>
    <Button
      variant="primary"
      on:click={confirmRelease}
      loading={releasing}
    >
      {releasing ? 'Releasing...' : 'Release'}
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

