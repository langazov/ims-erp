<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Pagination from '$lib/shared/components/data/Pagination.svelte';
  import ConfirmModal from '$lib/shared/components/layout/ConfirmModal.svelte';
  import EmptyState from '$lib/shared/components/display/EmptyState.svelte';
  import {
    getInvoices,
    deleteInvoice,
    getInvoiceStats,
    type Invoice,
    type InvoiceFilter
  } from '$lib/shared/api/invoices';
  import { toast } from '$lib/shared/stores/toast';

  let invoices: Invoice[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter: string = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;

  let stats: {
    total: number;
    draft: number;
    sent: number;
    paid: number;
    overdue: number;
    totalAmount: string;
    paidAmount: string;
  } | null = null;

  let deleteTarget: Invoice | null = null;
  let deleteModalOpen = false;
  let deleting = false;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'draft', label: 'Draft' },
    { value: 'sent', label: 'Sent' },
    { value: 'paid', label: 'Paid' },
    { value: 'overdue', label: 'Overdue' },
    { value: 'cancelled', label: 'Cancelled' }
  ];

  const columns = [
    { key: 'number', label: 'Invoice #', sortable: true },
    { key: 'client', label: 'Client', sortable: true },
    { key: 'issueDate', label: 'Issue Date', sortable: true },
    { key: 'dueDate', label: 'Due Date', sortable: true },
    { key: 'total', label: 'Total', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  function getStatusVariant(status: string): 'green' | 'gray' | 'yellow' | 'red' | 'blue' {
    switch (status) {
      case 'paid': return 'green';
      case 'sent': return 'blue';
      case 'draft': return 'gray';
      case 'overdue': return 'red';
      case 'cancelled': return 'yellow';
      default: return 'gray';
    }
  }

  function formatCurrency(value: string | number): string {
    const num = typeof value === 'string' ? parseFloat(value) : value;
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(isNaN(num) ? 0 : num);
  }

  function formatDate(date: string): string {
    if (!date) return '—';
    return new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  async function loadInvoices() {
    loading = true;
    error = null;
    try {
      const filter: InvoiceFilter = {
        page: currentPage,
        pageSize,
        search: searchQuery || undefined,
        status: statusFilter ? (statusFilter as any) : undefined
      };
      const res = await getInvoices(filter);
      invoices = res.data;
      totalItems = res.total;
      totalPages = res.totalPages;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load invoices';
    } finally {
      loading = false;
    }
  }

  async function loadStats() {
    try {
      stats = await getInvoiceStats();
    } catch {
      // non-fatal — summary bar simply won't show
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadInvoices();
  }

  function handleRowClick(invoice: Invoice) {
    goto(`/invoices/${invoice.id}`);
  }

  function handleEdit(invoice: Invoice, event: Event) {
    event.stopPropagation();
    goto(`/invoices/${invoice.id}/edit`);
  }

  function handleDeleteClick(invoice: Invoice, event: Event) {
    event.stopPropagation();
    deleteTarget = invoice;
    deleteModalOpen = true;
  }

  async function confirmDelete() {
    if (!deleteTarget) return;
    deleting = true;
    try {
      await deleteInvoice(deleteTarget.id);
      toast.success(`Invoice ${deleteTarget.invoiceNumber} deleted`);
      deleteTarget = null;
      deleteModalOpen = false;
      await Promise.all([loadInvoices(), loadStats()]);
    } catch (err) {
      toast.error('Failed to delete invoice');
    } finally {
      deleting = false;
    }
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadInvoices();
  }

  $: outstandingCount = stats ? stats.draft + stats.sent + stats.overdue : 0;
  $: outstandingAmount = stats
    ? parseFloat(stats.totalAmount || '0') - parseFloat(stats.paidAmount || '0')
    : 0;

  onMount(() => {
    loadInvoices();
    loadStats();
  });
</script>

<svelte:head>
  <title>Invoices | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Invoices</h1>
      <p class="page-description">Manage invoices and track payments</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={() => goto('/invoices/new')}>
        + Create Invoice
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => (error = null)}>
      {error}
    </Alert>
  {/if}

  <!-- Summary bar -->
  {#if stats}
    <div class="summary-bar">
      <div class="summary-card summary-outstanding">
        <div class="summary-label">Outstanding</div>
        <div class="summary-value">{formatCurrency(outstandingAmount)}</div>
        <div class="summary-count">{outstandingCount} invoice{outstandingCount !== 1 ? 's' : ''}</div>
      </div>
      <div class="summary-card summary-paid">
        <div class="summary-label">Paid</div>
        <div class="summary-value">{formatCurrency(stats.paidAmount)}</div>
        <div class="summary-count">{stats.paid} invoice{stats.paid !== 1 ? 's' : ''}</div>
      </div>
      <div class="summary-card summary-overdue">
        <div class="summary-label">Overdue</div>
        <div class="summary-value">{stats.overdue}</div>
        <div class="summary-count">invoice{stats.overdue !== 1 ? 's' : ''} past due</div>
      </div>
      <div class="summary-card summary-total">
        <div class="summary-label">Total Invoiced</div>
        <div class="summary-value">{formatCurrency(stats.totalAmount)}</div>
        <div class="summary-count">{stats.total} invoice{stats.total !== 1 ? 's' : ''}</div>
      </div>
    </div>
  {/if}

  <Card>
    <div class="filters">
      <div class="filter-row">
        <div class="filter-item search-filter">
          <Input
            id="search"
            label="Search"
            type="search"
            placeholder="Search by number, client..."
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
        <div class="filter-actions">
          <Button variant="secondary" on:click={handleSearch}>Search</Button>
        </div>
      </div>
    </div>

    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading invoices...</p>
      </div>
    {:else if invoices.length === 0}
      <EmptyState
        icon="🧾"
        title="No invoices found"
        description={searchQuery || statusFilter
          ? 'No invoices match your current filters. Try adjusting your search.'
          : 'Get started by creating your first invoice.'}
        actionLabel={!searchQuery && !statusFilter ? 'Create Invoice' : ''}
        on:action={() => goto('/invoices/new')}
      />
    {:else}
      <Table {columns}>
        <tbody>
          {#each invoices as invoice}
            <tr
              on:click={() => handleRowClick(invoice)}
              class="clickable-row"
              class:overdue-row={invoice.status === 'overdue'}
            >
              <td class="font-medium invoice-number">{invoice.invoiceNumber}</td>
              <td>{invoice.clientName}</td>
              <td>{formatDate(invoice.createdAt)}</td>
              <td class:text-red-600={invoice.status === 'overdue'} class:font-medium={invoice.status === 'overdue'}>
                {formatDate(invoice.dueDate)}
              </td>
              <td class="font-medium">{formatCurrency(invoice.total)}</td>
              <td>
                <Badge variant={getStatusVariant(invoice.status)}>
                  {invoice.status}
                </Badge>
              </td>
              <td>
                <div class="actions-cell" role="presentation" on:click|stopPropagation>
                  <Button variant="ghost" size="sm" on:click={() => goto(`/invoices/${invoice.id}`)}>
                    View
                  </Button>
                  {#if invoice.status === 'draft'}
                    <Button variant="ghost" size="sm" on:click={(e) => handleEdit(invoice, e)}>
                      Edit
                    </Button>
                  {/if}
                  <Button variant="ghost" size="sm" on:click={(e) => handleDeleteClick(invoice, e)}>
                    Delete
                  </Button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </Table>

      <div class="pagination-wrapper">
        <Pagination
          {currentPage}
          {totalPages}
          {totalItems}
          {pageSize}
          on:pageChange={(e) => handlePageChange(e.detail)}
        />
      </div>
    {/if}
  </Card>
</div>

<ConfirmModal
  bind:open={deleteModalOpen}
  title="Delete Invoice"
  message={deleteTarget
    ? `Are you sure you want to delete invoice ${deleteTarget.invoiceNumber}? This action cannot be undone.`
    : ''}
  confirmLabel="Delete"
  variant="danger"
  loading={deleting}
  on:confirm={confirmDelete}
  on:cancel={() => { deleteModalOpen = false; deleteTarget = null; }}
/>

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
    color: var(--color-gray-900, #111827);
    margin: 0;
  }

  .page-description {
    color: var(--color-gray-500, #6b7280);
    margin-top: 0.25rem;
    margin-bottom: 0;
  }

  /* Summary bar */
  .summary-bar {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  @media (max-width: 768px) {
    .summary-bar {
      grid-template-columns: repeat(2, 1fr);
    }
  }

  .summary-card {
    background: #fff;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 1rem 1.25rem;
    border-left: 4px solid transparent;
  }

  .summary-outstanding { border-left-color: #3b82f6; }
  .summary-paid        { border-left-color: #10b981; }
  .summary-overdue     { border-left-color: #ef4444; }
  .summary-total       { border-left-color: #8b5cf6; }

  .summary-label {
    font-size: 0.75rem;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #6b7280;
    margin-bottom: 0.25rem;
  }

  .summary-value {
    font-size: 1.375rem;
    font-weight: 700;
    color: #111827;
    line-height: 1.2;
  }

  .summary-count {
    font-size: 0.75rem;
    color: #9ca3af;
    margin-top: 0.2rem;
  }

  /* Filters */
  .filters {
    margin-bottom: 1rem;
  }

  .filter-row {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
    align-items: flex-end;
  }

  .filter-item {
    flex: 1;
    min-width: 180px;
  }

  .search-filter {
    flex: 2;
    min-width: 280px;
  }

  .filter-actions {
    padding-bottom: 0.125rem;
  }

  /* Loading / empty */
  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: #6b7280;
  }

  /* Table rows */
  .clickable-row {
    cursor: pointer;
    transition: background-color 0.1s;
  }

  .clickable-row:hover {
    background-color: #f9fafb;
  }

  .overdue-row {
    background-color: #fff5f5 !important;
  }

  .overdue-row:hover {
    background-color: #fee2e2 !important;
  }

  .invoice-number {
    font-family: monospace;
    font-size: 0.875rem;
  }

  .actions-cell {
    display: flex;
    gap: 0.25rem;
  }

  .pagination-wrapper {
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e5e7eb;
  }
</style>
