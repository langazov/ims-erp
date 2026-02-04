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

  interface Invoice {
    id: string;
    number: string;
    clientId: string;
    clientName: string;
    issueDate: string;
    dueDate: string;
    status: 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled';
    subtotal: number;
    taxAmount: number;
    total: number;
    paidAmount: number;
    balanceDue: number;
    lineItems: {
      description: string;
      quantity: number;
      unitPrice: number;
      total: number;
    }[];
    notes: string;
    createdAt: string;
  }

  let invoices: Invoice[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter: string = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;
  let showCreateModal = false;
  let deleteInvoiceId: string | null = null;

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
    { key: 'balance', label: 'Balance', sortable: true },
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

  function isOverdue(dueDate: string, status: string): boolean {
    if (status === 'paid' || status === 'cancelled') return false;
    return new Date(dueDate) < new Date();
  }

  async function loadInvoices() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      invoices = [
        {
          id: '1',
          number: 'INV-2024-001',
          clientId: '1',
          clientName: 'Acme Corporation',
          issueDate: '2024-01-15',
          dueDate: '2024-02-15',
          status: 'paid',
          subtotal: 5000.00,
          taxAmount: 400.00,
          total: 5400.00,
          paidAmount: 5400.00,
          balanceDue: 0,
          lineItems: [
            { description: 'Consulting Services', quantity: 40, unitPrice: 125.00, total: 5000.00 }
          ],
          notes: 'Payment received via bank transfer',
          createdAt: '2024-01-15T10:30:00Z'
        },
        {
          id: '2',
          number: 'INV-2024-002',
          clientId: '2',
          clientName: 'TechStart Inc',
          issueDate: '2024-01-20',
          dueDate: '2024-02-20',
          status: 'sent',
          subtotal: 3500.00,
          taxAmount: 280.00,
          total: 3780.00,
          paidAmount: 0,
          balanceDue: 3780.00,
          lineItems: [
            { description: 'Software Development', quantity: 20, unitPrice: 175.00, total: 3500.00 }
          ],
          notes: '',
          createdAt: '2024-01-20T14:00:00Z'
        },
        {
          id: '3',
          number: 'INV-2024-003',
          clientId: '3',
          clientName: 'Global Solutions LLC',
          issueDate: '2024-01-05',
          dueDate: '2024-02-05',
          status: 'overdue',
          subtotal: 8200.00,
          taxAmount: 656.00,
          total: 8856.00,
          paidAmount: 2000.00,
          balanceDue: 6856.00,
          lineItems: [
            { description: 'Project Management', quantity: 60, unitPrice: 100.00, total: 6000.00 },
            { description: 'Training Sessions', quantity: 8, unitPrice: 275.00, total: 2200.00 }
          ],
          notes: 'Partial payment received',
          createdAt: '2024-01-05T09:00:00Z'
        },
        {
          id: '4',
          number: 'INV-2024-004',
          clientId: '1',
          clientName: 'Acme Corporation',
          issueDate: '2024-01-25',
          dueDate: '2024-02-25',
          status: 'draft',
          subtotal: 1200.00,
          taxAmount: 96.00,
          total: 1296.00,
          paidAmount: 0,
          balanceDue: 1296.00,
          lineItems: [
            { description: 'Maintenance Services', quantity: 4, unitPrice: 300.00, total: 1200.00 }
          ],
          notes: 'Draft - awaiting approval',
          createdAt: '2024-01-25T11:30:00Z'
        }
      ];
      
      totalItems = invoices.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load invoices';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadInvoices();
  }

  function handleRowClick(invoice: Invoice) {
    window.location.href = `/invoices/${invoice.id}`;
  }

  function handleEdit(invoice: Invoice, event: Event) {
    event.stopPropagation();
    window.location.href = `/invoices/${invoice.id}/edit`;
  }

  async function handleDelete(invoice: Invoice, event: Event) {
    event.stopPropagation();
    deleteInvoiceId = invoice.id;
  }

  async function confirmDelete() {
    if (!deleteInvoiceId) return;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      invoices = invoices.filter(i => i.id !== deleteInvoiceId);
      totalItems = invoices.length;
      deleteInvoiceId = null;
    } catch (err) {
      error = 'Failed to delete invoice';
    }
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadInvoices();
  }

  onMount(() => {
    loadInvoices();
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
      <Button variant="primary" on:click={() => showCreateModal = true}>
        Create Invoice
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
            placeholder="Search invoices..."
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
          <Button variant="secondary" on:click={handleSearch}>
            Search
          </Button>
        </div>
      </div>
    </div>

    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading invoices...</p>
      </div>
    {:else if invoices.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2z" />
        </svg>
        <p class="text-gray-500 mb-4">No invoices found</p>
        <Button variant="primary" on:click={() => showCreateModal = true}>
          Create Your First Invoice
        </Button>
      </div>
    {:else}
      <Table {columns}>
        <tbody>
          {#each invoices as invoice}
            <tr on:click={() => handleRowClick(invoice)} class="clickable-row">
              <td class="font-medium">{invoice.number}</td>
              <td>{invoice.clientName}</td>
              <td>{formatDate(invoice.issueDate)}</td>
              <td class={isOverdue(invoice.dueDate, invoice.status) ? 'text-red-600 font-medium' : ''}>
                {formatDate(invoice.dueDate)}
                {#if isOverdue(invoice.dueDate, invoice.status)}
                  <span class="text-xs ml-1">(Overdue)</span>
                {/if}
              </td>
              <td class="font-medium">{formatCurrency(invoice.total)}</td>
              <td>
                <Badge variant={getStatusVariant(invoice.status)}>
                  {invoice.status}
                </Badge>
              </td>
              <td class={invoice.balanceDue > 0 ? 'text-red-600 font-medium' : 'text-green-600'}>
                {formatCurrency(invoice.balanceDue)}
              </td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={(e) => handleEdit(invoice, e)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={(e) => handleDelete(invoice, e)}>
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
  title="Create Invoice"
  size="lg"
>
  <p class="text-gray-600">Invoice creation form will be implemented here.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close}>Cancel</Button>
    <Button variant="primary" on:click={() => {
      showCreateModal = false;
    }}>Create</Button>
  </svelte:fragment>
</Modal>

{#if deleteInvoiceId}
  <Modal
    open={true}
    title="Delete Invoice"
    size="sm"
  >
    <p>Are you sure you want to delete this invoice? This action cannot be undone.</p>
    
    <svelte:fragment slot="footer" let:close>
      <Button variant="secondary" on:click={() => { close(); deleteInvoiceId = null; }}>Cancel</Button>
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
