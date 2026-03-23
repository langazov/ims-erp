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
    getPayments,
    getPaymentStats,
    cancelPayment,
    type Payment,
    type PaymentStatus,
    type PaymentMethod,
    type PaymentStats
  } from '$lib/shared/api/payments';
  import { toast } from '$lib/shared/stores/toast';

  let payments: Payment[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let statusFilter: PaymentStatus | '' = '';
  let methodFilter: PaymentMethod | '' = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;

  let stats: PaymentStats | null = null;

  let cancelTarget: Payment | null = null;
  let cancelModalOpen = false;
  let cancelling = false;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'pending', label: 'Pending' },
    { value: 'processing', label: 'Processing' },
    { value: 'completed', label: 'Completed' },
    { value: 'failed', label: 'Failed' },
    { value: 'refunded', label: 'Refunded' },
    { value: 'cancelled', label: 'Cancelled' }
  ];

  const methodOptions = [
    { value: '', label: 'All Methods' },
    { value: 'card', label: 'Card' },
    { value: 'bank_transfer', label: 'Bank Transfer' },
    { value: 'cash', label: 'Cash' },
    { value: 'check', label: 'Check' },
    { value: 'stripe', label: 'Stripe' },
    { value: 'paypal', label: 'PayPal' }
  ];

  const columns = [
    { key: 'paymentNumber', label: 'Payment #', sortable: true },
    { key: 'clientName', label: 'Client', sortable: true },
    { key: 'invoiceNumber', label: 'Invoice #', sortable: false },
    { key: 'amount', label: 'Amount', sortable: true },
    { key: 'method', label: 'Method', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'paidAt', label: 'Date', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  function formatCurrency(amount: string): string {
    const num = parseFloat(amount);
    if (isNaN(num)) return amount;
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(num);
  }

  function formatDate(dateStr: string | null): string {
    if (!dateStr) return '—';
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function getStatusVariant(status: PaymentStatus): 'green' | 'yellow' | 'red' | 'gray' | 'blue' {
    switch (status) {
      case 'completed': return 'green';
      case 'pending': return 'yellow';
      case 'processing': return 'blue';
      case 'failed': return 'red';
      case 'refunded': return 'gray';
      case 'cancelled': return 'gray';
      default: return 'gray';
    }
  }

  function formatMethodLabel(method: PaymentMethod): string {
    const labels: Record<PaymentMethod, string> = {
      card: 'Card',
      bank_transfer: 'Bank Transfer',
      cash: 'Cash',
      check: 'Check',
      stripe: 'Stripe',
      paypal: 'PayPal'
    };
    return labels[method] ?? method;
  }

  async function loadStats() {
    try {
      stats = await getPaymentStats();
    } catch {
      stats = null;
    }
  }

  async function loadPayments() {
    loading = true;
    error = null;
    try {
      const filter: Record<string, unknown> = {
        page: currentPage,
        pageSize
      };
      if (statusFilter) filter.status = statusFilter;
      if (methodFilter) filter.method = methodFilter;
      if (searchQuery.trim()) filter.search = searchQuery.trim();

      const response = await getPayments(filter as any);
      payments = response.data;
      totalPages = response.totalPages;
      totalItems = response.total;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load payments';
    } finally {
      loading = false;
    }
  }

  function handleFilterChange() {
    currentPage = 1;
    loadPayments();
  }

  function openCancelModal(payment: Payment) {
    cancelTarget = payment;
    cancelModalOpen = true;
  }

  async function handleCancelConfirm() {
    if (!cancelTarget) return;
    cancelling = true;
    try {
      await cancelPayment(cancelTarget.id);
      toast.success(`Payment ${cancelTarget.paymentNumber} cancelled`);
      cancelModalOpen = false;
      cancelTarget = null;
      await loadPayments();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to cancel payment');
    } finally {
      cancelling = false;
    }
  }

  onMount(() => {
    loadStats();
    loadPayments();
  });
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Payments</h1>
      <p class="page-subtitle">Track and manage all payment transactions</p>
    </div>
  </div>

  <!-- Stats Bar -->
  {#if stats}
    <div class="stats-bar">
      <div class="stat-item">
        <span class="stat-label">Today</span>
        <span class="stat-value">{formatCurrency(stats.totalAmountToday)}</span>
      </div>
      <div class="stat-divider"></div>
      <div class="stat-item">
        <span class="stat-label">This Month</span>
        <span class="stat-value">{formatCurrency(stats.totalAmountThisMonth)}</span>
      </div>
      <div class="stat-divider"></div>
      <div class="stat-item">
        <span class="stat-label">Total Collected</span>
        <span class="stat-value">{formatCurrency(stats.totalAmountCollected)}</span>
      </div>
      <div class="stat-divider"></div>
      <div class="stat-item">
        <span class="stat-label">Pending</span>
        <span class="stat-value stat-pending">{stats.pending}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Completed</span>
        <span class="stat-value stat-completed">{stats.completed}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Failed</span>
        <span class="stat-value stat-failed">{stats.failed}</span>
      </div>
    </div>
  {/if}

  <Card>
    <!-- Filters -->
    <div class="filters-row">
      <div class="search-wrap">
        <Input
          id="search"
          label=""
          type="text"
          placeholder="Search payments..."
          bind:value={searchQuery}
        />
      </div>
      <div class="filter-wrap">
        <Select
          id="status-filter"
          label=""
          options={statusOptions}
          bind:value={statusFilter}
          on:change={handleFilterChange}
        />
      </div>
      <div class="filter-wrap">
        <Select
          id="method-filter"
          label=""
          options={methodOptions}
          bind:value={methodFilter}
          on:change={handleFilterChange}
        />
      </div>
      <Button variant="secondary" on:click={handleFilterChange}>Search</Button>
    </div>

    {#if error}
      <Alert variant="error" dismissible on:dismiss={() => (error = null)}>{error}</Alert>
    {/if}

    {#if loading}
      <div class="spinner-wrap">
        <Spinner size="lg" />
      </div>
    {:else if payments.length === 0}
      <EmptyState
        title="No payments found"
        description="No payments match your current filters."
        icon="💳"
      />
    {:else}
      <Table {columns}>
        <tbody>
          {#each payments as payment (payment.id)}
            <tr
              class="table-row"
              on:click={() => goto(`/payments/${payment.id}`)}
              role="button"
              tabindex="0"
              on:keydown={(e) => e.key === 'Enter' && goto(`/payments/${payment.id}`)}
            >
              <td class="td">{payment.paymentNumber}</td>
              <td class="td">{payment.clientName}</td>
              <td class="td">
                <a
                  href="/invoices/{payment.invoiceId}"
                  class="link"
                  on:click|stopPropagation
                >
                  {payment.invoiceNumber}
                </a>
              </td>
              <td class="td font-medium">{formatCurrency(payment.amount)} {payment.currency}</td>
              <td class="td">{formatMethodLabel(payment.method)}</td>
              <td class="td">
                <Badge variant={getStatusVariant(payment.status)} dot>
                  {payment.status.charAt(0).toUpperCase() + payment.status.slice(1)}
                </Badge>
              </td>
              <td class="td">{formatDate(payment.paidAt ?? payment.createdAt)}</td>
              <td class="td actions-cell" on:click|stopPropagation>
                {#if payment.status === 'pending' || payment.status === 'processing'}
                  <Button
                    variant="ghost"
                    size="sm"
                    on:click={() => openCancelModal(payment)}
                  >
                    Cancel
                  </Button>
                {/if}
                <Button
                  variant="ghost"
                  size="sm"
                  on:click={() => goto(`/payments/${payment.id}`)}
                >
                  View
                </Button>
              </td>
            </tr>
          {/each}
        </tbody>
      </Table>

      {#if totalPages > 1}
        <div class="pagination-wrap">
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            totalItems={totalItems}
            pageSize={pageSize}
            on:pageChange={(e) => {
              currentPage = e.detail;
              loadPayments();
            }}
          />
        </div>
      {/if}
    {/if}
  </Card>
</div>

<ConfirmModal
  bind:open={cancelModalOpen}
  title="Cancel Payment"
  message="Are you sure you want to cancel payment {cancelTarget?.paymentNumber}? This action cannot be undone."
  confirmLabel="Cancel Payment"
  cancelLabel="Keep Payment"
  variant="danger"
  loading={cancelling}
  on:confirm={handleCancelConfirm}
  on:cancel={() => (cancelModalOpen = false)}
/>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1400px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .page-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-gray-900, #111827);
    margin: 0;
  }

  .page-subtitle {
    font-size: 0.875rem;
    color: var(--color-gray-500, #6b7280);
    margin: 0.25rem 0 0;
  }

  .stats-bar {
    display: flex;
    align-items: center;
    gap: 1.5rem;
    padding: 1rem 1.25rem;
    background: var(--color-white, #ffffff);
    border: 1px solid var(--color-gray-200, #e5e7eb);
    border-radius: 0.75rem;
    flex-wrap: wrap;
  }

  .stat-item {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .stat-label {
    font-size: 0.75rem;
    color: var(--color-gray-500, #6b7280);
    font-weight: 500;
  }

  .stat-value {
    font-size: 1rem;
    font-weight: 700;
    color: var(--color-gray-900, #111827);
  }

  .stat-pending { color: #d97706; }
  .stat-completed { color: #059669; }
  .stat-failed { color: #dc2626; }

  .stat-divider {
    width: 1px;
    height: 2rem;
    background: var(--color-gray-200, #e5e7eb);
  }

  .filters-row {
    display: flex;
    gap: 0.75rem;
    align-items: flex-end;
    margin-bottom: 1rem;
    flex-wrap: wrap;
  }

  .search-wrap {
    flex: 1;
    min-width: 200px;
  }

  .filter-wrap {
    width: 160px;
  }

  .spinner-wrap {
    display: flex;
    justify-content: center;
    padding: 3rem 0;
  }

  .table-row {
    cursor: pointer;
    transition: background-color 0.15s;
  }

  .table-row:hover {
    background-color: var(--color-gray-50, #f9fafb);
  }

  .td {
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    color: var(--color-gray-700, #374151);
    border-bottom: 1px solid var(--color-gray-100, #f3f4f6);
    vertical-align: middle;
  }

  .font-medium {
    font-weight: 600;
  }

  .link {
    color: var(--color-primary-600, #2563eb);
    text-decoration: none;
    font-weight: 500;
  }

  .link:hover {
    text-decoration: underline;
  }

  .actions-cell {
    display: flex;
    gap: 0.25rem;
    align-items: center;
  }

  .pagination-wrap {
    margin-top: 1rem;
    display: flex;
    justify-content: center;
  }
</style>
