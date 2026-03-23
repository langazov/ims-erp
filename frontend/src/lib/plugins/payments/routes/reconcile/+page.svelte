<script lang="ts">
  import { onMount } from 'svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import EmptyState from '$lib/shared/components/display/EmptyState.svelte';
  import { getPayments, type Payment, type PaymentStatus, type PaymentMethod } from '$lib/shared/api/payments';
  import { toast } from '$lib/shared/stores/toast';

  interface DateGroup {
    date: string;
    payments: Payment[];
    total: number;
    reconciled: number;
    unreconciled: number;
  }

  let loading = true;
  let error: string | null = null;
  let dateGroups: DateGroup[] = [];
  let totalReconciled = 0;
  let totalUnreconciled = 0;
  let reconciledCount = 0;
  let unreconciledCount = 0;

  function formatCurrency(amount: string | number): string {
    const num = typeof amount === 'string' ? parseFloat(amount) : amount;
    if (isNaN(num)) return String(amount);
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(num);
  }

  function toDateKey(dateStr: string): string {
    return new Date(dateStr).toISOString().split('T')[0];
  }

  function formatDateHeader(key: string): string {
    const date = new Date(key + 'T12:00:00');
    return date.toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' });
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

  function groupByDate(payments: Payment[]): DateGroup[] {
    const map = new Map<string, Payment[]>();
    for (const p of payments) {
      const key = toDateKey(p.paidAt ?? p.createdAt);
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(p);
    }

    const groups: DateGroup[] = [];
    const sortedKeys = Array.from(map.keys()).sort((a, b) => b.localeCompare(a));
    for (const date of sortedKeys) {
      const list = map.get(date)!;
      let reconciled = 0;
      let unreconciled = 0;
      for (const p of list) {
        const amt = parseFloat(p.amount) || 0;
        if (p.status === 'completed') reconciled += amt;
        else unreconciled += amt;
      }
      groups.push({ date, payments: list, total: reconciled + unreconciled, reconciled, unreconciled });
    }
    return groups;
  }

  async function loadPayments() {
    loading = true;
    error = null;
    try {
      const response = await getPayments({ pageSize: 500, page: 1 });
      const all = response.data;
      dateGroups = groupByDate(all);

      totalReconciled = 0;
      totalUnreconciled = 0;
      reconciledCount = 0;
      unreconciledCount = 0;

      for (const p of all) {
        const amt = parseFloat(p.amount) || 0;
        if (p.status === 'completed') {
          totalReconciled += amt;
          reconciledCount++;
        } else {
          totalUnreconciled += amt;
          unreconciledCount++;
        }
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load payments';
    } finally {
      loading = false;
    }
  }

  function handleExport() {
    toast.info('Export feature coming soon');
  }

  onMount(loadPayments);
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Payment Reconciliation</h1>
      <p class="page-subtitle">Review and reconcile payments grouped by date</p>
    </div>
    <Button variant="secondary" on:click={handleExport}>Export</Button>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => (error = null)}>{error}</Alert>
  {/if}

  {#if loading}
    <div class="spinner-wrap">
      <Spinner size="lg" />
    </div>
  {:else}
    <!-- Summary -->
    <div class="summary-bar">
      <div class="summary-item">
        <span class="summary-label">Reconciled</span>
        <span class="summary-value reconciled">{formatCurrency(totalReconciled)}</span>
        <span class="summary-count">{reconciledCount} payments</span>
      </div>
      <div class="summary-divider"></div>
      <div class="summary-item">
        <span class="summary-label">Unreconciled</span>
        <span class="summary-value unreconciled">{formatCurrency(totalUnreconciled)}</span>
        <span class="summary-count">{unreconciledCount} payments</span>
      </div>
      <div class="summary-divider"></div>
      <div class="summary-item">
        <span class="summary-label">Total</span>
        <span class="summary-value">{formatCurrency(totalReconciled + totalUnreconciled)}</span>
        <span class="summary-count">{reconciledCount + unreconciledCount} payments</span>
      </div>
    </div>

    {#if dateGroups.length === 0}
      <EmptyState
        title="No payments to reconcile"
        description="There are no payment records available."
        icon="💳"
      />
    {:else}
      <div class="groups-list">
        {#each dateGroups as group (group.date)}
          <div class="date-group">
            <div class="date-header">
              <div class="date-header-left">
                <span class="date-label">{formatDateHeader(group.date)}</span>
                <span class="date-count">{group.payments.length} payment{group.payments.length !== 1 ? 's' : ''}</span>
              </div>
              <div class="date-totals">
                {#if group.reconciled > 0}
                  <span class="date-reconciled">✓ {formatCurrency(group.reconciled)}</span>
                {/if}
                {#if group.unreconciled > 0}
                  <span class="date-unreconciled">⏳ {formatCurrency(group.unreconciled)}</span>
                {/if}
              </div>
            </div>
            <Card>
              <table class="payments-table">
                <thead>
                  <tr>
                    <th>Payment #</th>
                    <th>Client</th>
                    <th>Amount</th>
                    <th>Method</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {#each group.payments as payment (payment.id)}
                    <tr>
                      <td class="td-mono">{payment.paymentNumber}</td>
                      <td>{payment.clientName}</td>
                      <td class="td-amount">{formatCurrency(payment.amount)} {payment.currency}</td>
                      <td>{formatMethodLabel(payment.method)}</td>
                      <td>
                        <Badge variant={getStatusVariant(payment.status)} dot>
                          {payment.status.charAt(0).toUpperCase() + payment.status.slice(1)}
                        </Badge>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </Card>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1100px;
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

  .spinner-wrap {
    display: flex;
    justify-content: center;
    padding: 4rem 0;
  }

  .summary-bar {
    display: flex;
    align-items: center;
    gap: 2rem;
    padding: 1rem 1.5rem;
    background: var(--color-white, #fff);
    border: 1px solid var(--color-gray-200, #e5e7eb);
    border-radius: 0.75rem;
    flex-wrap: wrap;
  }

  .summary-item {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .summary-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-gray-500, #6b7280);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .summary-value {
    font-size: 1.375rem;
    font-weight: 700;
    color: var(--color-gray-900, #111827);
  }

  .summary-value.reconciled { color: #059669; }
  .summary-value.unreconciled { color: #d97706; }

  .summary-count {
    font-size: 0.75rem;
    color: var(--color-gray-400, #9ca3af);
  }

  .summary-divider {
    width: 1px;
    height: 3rem;
    background: var(--color-gray-200, #e5e7eb);
  }

  .groups-list {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .date-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .date-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .date-header-left {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .date-label {
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--color-gray-800, #1f2937);
  }

  .date-count {
    font-size: 0.75rem;
    color: var(--color-gray-400, #9ca3af);
    background: var(--color-gray-100, #f3f4f6);
    padding: 0.125rem 0.5rem;
    border-radius: 9999px;
  }

  .date-totals {
    display: flex;
    gap: 1rem;
    font-size: 0.8125rem;
    font-weight: 600;
  }

  .date-reconciled { color: #059669; }
  .date-unreconciled { color: #d97706; }

  .payments-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .payments-table thead tr th {
    padding: 0.625rem 0.875rem;
    text-align: left;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--color-gray-500, #6b7280);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid var(--color-gray-100, #f3f4f6);
  }

  .payments-table tbody tr td {
    padding: 0.75rem 0.875rem;
    color: var(--color-gray-700, #374151);
    border-bottom: 1px solid var(--color-gray-50, #f9fafb);
    vertical-align: middle;
  }

  .payments-table tbody tr:last-child td {
    border-bottom: none;
  }

  .td-mono {
    font-family: monospace;
    font-size: 0.8125rem;
    color: var(--color-gray-600, #4b5563);
  }

  .td-amount {
    font-weight: 600;
    color: var(--color-gray-900, #111827);
  }
</style>
