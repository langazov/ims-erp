<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import Progress from '$lib/shared/components/display/Progress.svelte';

  const clientId = $page.params.id;

  interface CreditTransaction {
    id: string;
    date: string;
    type: 'credit_limit_change' | 'invoice' | 'payment' | 'refund' | 'adjustment';
    description: string;
    amount: number;
    balance: number;
  }

  interface CreditStatus {
    creditLimit: number;
    currentBalance: number;
    availableCredit: number;
    utilizationPercentage: number;
  }

  let creditStatus: CreditStatus | null = null;
  let transactions: CreditTransaction[] = [];
  let loading = true;
  let error: string | null = null;
  let showAdjustModal = false;
  let newCreditLimit = '';
  let adjustmentReason = '';
  let saving = false;

  const transactionColumns = [
    { key: 'date', label: 'Date', sortable: true },
    { key: 'type', label: 'Type', sortable: true },
    { key: 'description', label: 'Description', sortable: false },
    { key: 'amount', label: 'Amount', sortable: true, align: 'right' as const },
    { key: 'balance', label: 'Balance', sortable: true, align: 'right' as const }
  ];

  async function loadCreditData() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      creditStatus = {
        creditLimit: 50000,
        currentBalance: 32500,
        availableCredit: 17500,
        utilizationPercentage: 65
      };
      
      transactions = [
        {
          id: '1',
          date: '2024-01-15',
          type: 'invoice',
          description: 'Invoice #INV-2024-001',
          amount: 12500,
          balance: 32500
        },
        {
          id: '2',
          date: '2024-01-10',
          type: 'payment',
          description: 'Payment received - Check #1234',
          amount: -5000,
          balance: 20000
        },
        {
          id: '3',
          date: '2024-01-05',
          type: 'invoice',
          description: 'Invoice #INV-2024-002',
          amount: 15000,
          balance: 25000
        },
        {
          id: '4',
          date: '2024-01-01',
          type: 'credit_limit_change',
          description: 'Credit limit increased',
          amount: 0,
          balance: 10000
        },
        {
          id: '5',
          date: '2023-12-20',
          type: 'invoice',
          description: 'Invoice #INV-2023-045',
          amount: 10000,
          balance: 10000
        }
      ];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load credit data';
    } finally {
      loading = false;
    }
  }

  function getUtilizationColor(percentage: number): 'green' | 'yellow' | 'red' {
    if (percentage < 50) return 'green';
    if (percentage < 80) return 'yellow';
    return 'red';
  }

  function getTransactionTypeVariant(type: string): 'green' | 'red' | 'blue' | 'purple' | 'gray' {
    switch (type) {
      case 'payment': return 'green';
      case 'invoice': return 'red';
      case 'refund': return 'blue';
      case 'credit_limit_change': return 'purple';
      default: return 'gray';
    }
  }

  function formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount);
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function openAdjustModal() {
    newCreditLimit = creditStatus?.creditLimit.toString() || '';
    adjustmentReason = '';
    showAdjustModal = true;
  }

  async function handleAdjustCredit() {
    const limit = parseFloat(newCreditLimit);
    if (isNaN(limit) || limit < 0) {
      error = 'Please enter a valid credit limit';
      return;
    }

    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      if (creditStatus) {
        const oldLimit = creditStatus.creditLimit;
        creditStatus = {
          ...creditStatus,
          creditLimit: limit,
          availableCredit: limit - creditStatus.currentBalance,
          utilizationPercentage: (creditStatus.currentBalance / limit) * 100
        };
        
        const newTransaction: CreditTransaction = {
          id: Date.now().toString(),
          date: new Date().toISOString().split('T')[0],
          type: 'credit_limit_change',
          description: `Credit limit ${limit > oldLimit ? 'increased' : 'decreased'}${adjustmentReason ? ': ' + adjustmentReason : ''}`,
          amount: 0,
          balance: creditStatus.currentBalance
        };
        transactions = [newTransaction, ...transactions];
      }
      
      showAdjustModal = false;
    } catch (err) {
      error = 'Failed to update credit limit';
    } finally {
      saving = false;
    }
  }

  onMount(() => {
    loadCreditData();
  });
</script>

<svelte:head>
  <title>Client Credit | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Credit Management</h1>
      <p class="page-description">Monitor and manage client credit limits and balances</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={openAdjustModal}>
        Adjust Credit Limit
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading credit information...</p>
    </div>
  {:else if creditStatus}
    <div class="credit-overview">
      <Card class="stat-card">
        <div class="stat-header">
          <span class="stat-label">Credit Limit</span>
          <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="stat-value">{formatCurrency(creditStatus.creditLimit)}</div>
      </Card>

      <Card class="stat-card">
        <div class="stat-header">
          <span class="stat-label">Current Balance</span>
          <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
          </svg>
        </div>
        <div class="stat-value" class:text-red-600={creditStatus.currentBalance > 0}>
          {formatCurrency(creditStatus.currentBalance)}
        </div>
      </Card>

      <Card class="stat-card">
        <div class="stat-header">
          <span class="stat-label">Available Credit</span>
          <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="stat-value text-green-600">
          {formatCurrency(creditStatus.availableCredit)}
        </div>
      </Card>
    </div>

    <Card class="utilization-card">
      <div class="utilization-header">
        <h3 class="section-title">Credit Utilization</h3>
        <Badge variant={getUtilizationColor(creditStatus.utilizationPercentage)} size="md">
          {creditStatus.utilizationPercentage.toFixed(1)}%
        </Badge>
      </div>
      <div class="progress-container">
        <div class="progress-bar">
          <div 
            class="progress-fill"
            class:bg-green-500={creditStatus.utilizationPercentage < 50}
            class:bg-yellow-500={creditStatus.utilizationPercentage >= 50 && creditStatus.utilizationPercentage < 80}
            class:bg-red-500={creditStatus.utilizationPercentage >= 80}
            style="width: {Math.min(creditStatus.utilizationPercentage, 100)}%"
          />
        </div>
        <div class="progress-labels">
          <span class="text-sm text-gray-500">0%</span>
          <span class="text-sm text-gray-500">50%</span>
          <span class="text-sm text-gray-500">100%</span>
        </div>
      </div>
      <p class="utilization-note">
        {#if creditStatus.utilizationPercentage >= 80}
          <span class="text-red-600">High utilization - consider reviewing credit limit</span>
        {:else if creditStatus.utilizationPercentage >= 50}
          <span class="text-yellow-600">Moderate utilization</span>
        {:else}
          <span class="text-green-600">Healthy credit utilization</span>
        {/if}
      </p>
    </Card>

    <Card>
      <h3 class="section-title mb-4">Credit History</h3>
      {#if transactions.length === 0}
        <div class="empty-state">
          <p class="text-gray-500">No transactions found</p>
        </div>
      {:else}
        <Table columns={transactionColumns} data={transactions}>
          <tbody>
            {#each transactions as transaction}
              <tr>
                <td>{formatDate(transaction.date)}</td>
                <td>
                  <Badge variant={getTransactionTypeVariant(transaction.type)}>
                    {transaction.type.replace('_', ' ')}
                  </Badge>
                </td>
                <td class="max-w-xs truncate">{transaction.description}</td>
                <td class="text-right">
                  {#if transaction.amount !== 0}
                    <span class={transaction.amount > 0 ? 'text-red-600' : 'text-green-600'}>
                      {transaction.amount > 0 ? '+' : ''}{formatCurrency(transaction.amount)}
                    </span>
                  {:else}
                    <span class="text-gray-400">—</span>
                  {/if}
                </td>
                <td class="text-right font-medium">{formatCurrency(transaction.balance)}</td>
              </tr>
            {/each}
          </tbody>
        </Table>
      {/if}
    </Card>
  {:else}
    <Alert variant="error">Credit information not found</Alert>
  {/if}
</div>

<Modal
  bind:open={showAdjustModal}
  title="Adjust Credit Limit"
  size="md"
>
  <div class="adjust-form">
    <div class="current-limit">
      <span class="label">Current Limit:</span>
      <span class="value">{creditStatus ? formatCurrency(creditStatus.creditLimit) : '—'}</span>
    </div>
    <Input
      id="newCreditLimit"
      label="New Credit Limit"
      type="number"
      placeholder="Enter new credit limit"
      bind:value={newCreditLimit}
      min="0"
      step="0.01"
      required
    />
    <div class="form-row">
      <label class="input-label" for="reason">Reason (Optional)</label>
      <textarea
        id="reason"
        class="reason-textarea"
        placeholder="Enter reason for adjustment..."
        bind:value={adjustmentReason}
        rows="3"
      />
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={saving}>Cancel</Button>
    <Button variant="primary" on:click={handleAdjustCredit} loading={saving}>
      {saving ? 'Saving...' : 'Save Changes'}
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

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .credit-overview {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  :global(.stat-card) {
    padding: 1.5rem;
  }

  .stat-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .stat-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .stat-value {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-gray-900);
  }

  .utilization-card {
    margin-bottom: 1.5rem;
  }

  :global(.utilization-card) {
    padding: 1.5rem;
  }

  .utilization-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0;
  }

  .progress-container {
    margin-bottom: 0.5rem;
  }

  .progress-bar {
    height: 0.75rem;
    background-color: var(--color-gray-200);
    border-radius: 9999px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    border-radius: 9999px;
    transition: width 0.3s ease;
  }

  .progress-labels {
    display: flex;
    justify-content: space-between;
    margin-top: 0.25rem;
  }

  .utilization-note {
    margin-top: 0.75rem;
    font-size: 0.875rem;
  }

  .empty-state {
    text-align: center;
    padding: 2rem;
  }

  .adjust-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .current-limit {
    display: flex;
    justify-content: space-between;
    padding: 0.75rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
    font-size: 0.875rem;
  }

  .current-limit .label {
    color: var(--color-gray-500);
  }

  .current-limit .value {
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .form-row {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .input-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-700);
  }

  .reason-textarea {
    width: 100%;
    padding: 0.625rem 0.875rem;
    border: 1px solid var(--color-gray-300);
    border-radius: 0.5rem;
    font-size: 0.875rem;
    resize: vertical;
    min-height: 80px;
  }

  .reason-textarea:focus {
    outline: none;
    border-color: var(--color-primary-500);
    box-shadow: 0 0 0 2px var(--color-primary-100);
  }

  @media (max-width: 768px) {
    .credit-overview {
      grid-template-columns: 1fr;
    }
  }
</style>
