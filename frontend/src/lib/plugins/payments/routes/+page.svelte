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

  interface Payment {
    id: string;
    reference: string;
    clientId: string;
    clientName: string;
    invoiceId: string;
    invoiceNumber: string;
    amount: number;
    method: 'credit_card' | 'bank_transfer' | 'cash' | 'check' | 'paypal' | 'stripe';
    status: 'pending' | 'completed' | 'failed' | 'refunded';
    date: string;
    notes: string;
  }

  let payments: Payment[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter: string = '';
  let methodFilter: string = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;
  let showCreateModal = false;
  let deletePaymentId: string | null = null;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'pending', label: 'Pending' },
    { value: 'completed', label: 'Completed' },
    { value: 'failed', label: 'Failed' },
    { value: 'refunded', label: 'Refunded' }
  ];

  const methodOptions = [
    { value: '', label: 'All Methods' },
    { value: 'credit_card', label: 'Credit Card' },
    { value: 'bank_transfer', label: 'Bank Transfer' },
    { value: 'cash', label: 'Cash' },
    { value: 'check', label: 'Check' },
    { value: 'paypal', label: 'PayPal' },
    { value: 'stripe', label: 'Stripe' }
  ];

  const columns = [
    { key: 'reference', label: 'Reference', sortable: true },
    { key: 'client', label: 'Client', sortable: true },
    { key: 'invoice', label: 'Invoice', sortable: true },
    { key: 'amount', label: 'Amount', sortable: true },
    { key: 'method', label: 'Method', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'date', label: 'Date', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  function getStatusVariant(status: string): 'green' | 'gray' | 'yellow' | 'red' | 'blue' {
    switch (status) {
      case 'completed': return 'green';
      case 'pending': return 'yellow';
      case 'failed': return 'red';
      case 'refunded': return 'blue';
      default: return 'gray';
    }
  }

  function formatMethod(method: string): string {
    return method.split('_').map(word => 
      word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ');
  }

  function formatCurrency(value: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(value);
  }

  function formatDate(date: string): string {
    return new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  async function loadPayments() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      payments = [
        {
          id: '1',
          reference: 'PAY-2024-001',
          clientId: '1',
          clientName: 'Acme Corporation',
          invoiceId: '1',
          invoiceNumber: 'INV-2024-001',
          amount: 5400.00,
          method: 'bank_transfer',
          status: 'completed',
          date: '2024-01-20',
          notes: 'Payment received via wire transfer'
        },
        {
          id: '2',
          reference: 'PAY-2024-002',
          clientId: '2',
          clientName: 'TechStart Inc',
          invoiceId: '2',
          invoiceNumber: 'INV-2024-002',
          amount: 3780.00,
          method: 'credit_card',
          status: 'pending',
          date: '2024-01-25',
          notes: 'Awaiting credit card processing'
        },
        {
          id: '3',
          reference: 'PAY-2024-003',
          clientId: '3',
          clientName: 'Global Solutions LLC',
          invoiceId: '3',
          invoiceNumber: 'INV-2024-003',
          amount: 2000.00,
          method: 'check',
          status: 'completed',
          date: '2024-01-10',
          notes: 'Partial payment - check cleared'
        },
        {
          id: '4',
          reference: 'PAY-2024-004',
          clientId: '1',
          clientName: 'Acme Corporation',
          invoiceId: '4',
          invoiceNumber: 'INV-2024-004',
          amount: 1296.00,
          method: 'paypal',
          status: 'failed',
          date: '2024-01-28',
          notes: 'Payment failed - insufficient funds'
        },
        {
          id: '5',
          reference: 'PAY-2024-005',
          clientId: '4',
          clientName: 'Digital Ventures Co',
          invoiceId: '5',
          invoiceNumber: 'INV-2024-005',
          amount: 2500.00,
          method: 'stripe',
          status: 'refunded',
          date: '2024-01-15',
          notes: 'Customer requested refund'
        }
      ];
      
      totalItems = payments.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load payments';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadPayments();
  }

  function handleRowClick(payment: Payment) {
    window.location.href = `/payments/${payment.id}`;
  }

  function handleEdit(payment: Payment, event: Event) {
    event.stopPropagation();
    window.location.href = `/payments/${payment.id}/edit`;
  }

  async function handleDelete(payment: Payment, event: Event) {
    event.stopPropagation();
    deletePaymentId = payment.id;
  }

  async function confirmDelete() {
    if (!deletePaymentId) return;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      payments = payments.filter(p => p.id !== deletePaymentId);
      totalItems = payments.length;
      deletePaymentId = null;
    } catch (err) {
      error = 'Failed to delete payment';
    }
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadPayments();
  }

  onMount(() => {
    loadPayments();
  });
</script>

<svelte:head>
  <title>Payments | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Payments</h1>
      <p class="page-description">Manage and track payments</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={() => showCreateModal = true}>
        Record Payment
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
            placeholder="Search payments..."
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
            id="method"
            label="Method"
            options={methodOptions}
            bind:value={methodFilter}
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
        <p>Loading payments...</p>
      </div>
    {:else if payments.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
        </svg>
        <p class="text-gray-500 mb-4">No payments found</p>
        <Button variant="primary" on:click={() => showCreateModal = true}>
          Record Your First Payment
        </Button>
      </div>
    {:else}
      <Table {columns}>
        <tbody>
          {#each payments as payment}
            <tr on:click={() => handleRowClick(payment)} class="clickable-row">
              <td class="font-medium">{payment.reference}</td>
              <td>{payment.clientName}</td>
              <td>{payment.invoiceNumber}</td>
              <td class="font-medium">{formatCurrency(payment.amount)}</td>
              <td>{formatMethod(payment.method)}</td>
              <td>
                <Badge variant={getStatusVariant(payment.status)}>
                  {payment.status}
                </Badge>
              </td>
              <td>{formatDate(payment.date)}</td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={(e) => handleEdit(payment, e)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={(e) => handleDelete(payment, e)}>
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
  title="Record Payment"
  size="lg"
>
  <p class="text-gray-600">Payment recording form will be implemented here.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close}>Cancel</Button>
    <Button variant="primary" on:click={() => {
      showCreateModal = false;
    }}>Record Payment</Button>
  </svelte:fragment>
</Modal>

{#if deletePaymentId}
  <Modal
    open={true}
    title="Delete Payment"
    size="sm"
  >
    <p>Are you sure you want to delete this payment? This action cannot be undone.</p>
    
    <svelte:fragment slot="footer" let:close>
      <Button variant="secondary" on:click={() => { close(); deletePaymentId = null; }}>Cancel</Button>
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
