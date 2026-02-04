<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import DatePicker from '$lib/shared/components/forms/DatePicker.svelte';
  import FileUpload from '$lib/shared/components/forms/FileUpload.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import Chip from '$lib/shared/components/display/Chip.svelte';

  interface BankTransaction {
    id: string;
    date: string;
    description: string;
    amount: string;
    reference?: string;
    status: 'unmatched' | 'matched' | 'disputed';
    matchedPaymentId?: string;
  }

  interface SystemPayment {
    id: string;
    paymentNumber: string;
    clientName: string;
    invoiceNumber: string;
    amount: string;
    method: string;
    paidAt: string;
    status: 'pending' | 'completed' | 'failed' | 'refunded';
    matchedTransactionId?: string;
  }

  interface MatchSuggestion {
    transactionId: string;
    paymentId: string;
    confidence: number;
    reason: string;
  }

  let loading = true;
  let error: string | null = null;
  let processing = false;
  
  // Date range
  let dateFrom = '';
  let dateTo = '';
  
  // Filters
  let statusFilter: 'all' | 'matched' | 'unmatched' | 'disputed' = 'all';
  
  // Data
  let bankTransactions: BankTransaction[] = [];
  let systemPayments: SystemPayment[] = [];
  let matchSuggestions: MatchSuggestion[] = [];
  
  // Selection
  let selectedTransactionId: string | null = null;
  let selectedPaymentId: string | null = null;
  
  // Modals
  let showConfirmModal = false;
  let showExportModal = false;
  let showMatchModal = false;
  let confirmAction: 'match' | 'unmatch' | 'reconcile' | null = null;
  
  // Statistics
  $: stats = calculateStats(bankTransactions, systemPayments);
  
  // Filtered data
  $: filteredTransactions = filterTransactions(bankTransactions, statusFilter);
  $: unmatchedPayments = systemPayments.filter(p => !p.matchedTransactionId && p.status === 'completed');

  const statusOptions = [
    { value: 'all', label: 'All Items' },
    { value: 'matched', label: 'Matched' },
    { value: 'unmatched', label: 'Unmatched' },
    { value: 'disputed', label: 'Disputed' }
  ];

  function calculateStats(transactions: BankTransaction[], payments: SystemPayment[]) {
    const totalTransactions = transactions.length;
    const matchedCount = transactions.filter(t => t.status === 'matched').length;
    const unmatchedCount = transactions.filter(t => t.status === 'unmatched').length;
    const disputedCount = transactions.filter(t => t.status === 'disputed').length;
    
    const totalBankAmount = transactions.reduce((sum, t) => sum + parseFloat(t.amount), 0);
    const matchedAmount = transactions
      .filter(t => t.status === 'matched')
      .reduce((sum, t) => sum + parseFloat(t.amount), 0);
    
    const totalPayments = payments.filter(p => p.status === 'completed').length;
    const matchedPayments = payments.filter(p => p.matchedTransactionId).length;
    
    return {
      totalTransactions,
      matchedCount,
      unmatchedCount,
      disputedCount,
      totalBankAmount,
      matchedAmount,
      totalPayments,
      matchedPayments
    };
  }

  function filterTransactions(transactions: BankTransaction[], filter: string): BankTransaction[] {
    if (filter === 'all') return transactions;
    return transactions.filter(t => t.status === filter);
  }

  async function loadData() {
    loading = true;
    error = null;
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 800));
      
      // Mock bank transactions
      bankTransactions = [
        {
          id: 'txn-001',
          date: '2024-01-15',
          description: 'ACME CORP - WIRE TRANSFER',
          amount: '5400.00',
          reference: 'WIRE-789456',
          status: 'matched',
          matchedPaymentId: 'pay-001'
        },
        {
          id: 'txn-002',
          date: '2024-01-16',
          description: 'TECHSTART INC - ACH PAYMENT',
          amount: '3780.00',
          reference: 'ACH-123456',
          status: 'matched',
          matchedPaymentId: 'pay-002'
        },
        {
          id: 'txn-003',
          date: '2024-01-17',
          description: 'GLOBAL SOLUTIONS LLC',
          amount: '2000.00',
          reference: 'CHK-4521',
          status: 'unmatched'
        },
        {
          id: 'txn-004',
          date: '2024-01-18',
          description: 'DIGITAL VENTURES CO',
          amount: '2500.00',
          reference: 'PAYPAL-789',
          status: 'disputed'
        },
        {
          id: 'txn-005',
          date: '2024-01-19',
          description: 'UNKNOWN MERCHANT',
          amount: '1500.00',
          status: 'unmatched'
        },
        {
          id: 'txn-006',
          date: '2024-01-20',
          description: 'ACME CORP - WIRE TRANSFER',
          amount: '3200.00',
          reference: 'WIRE-789457',
          status: 'unmatched'
        }
      ];
      
      // Mock system payments
      systemPayments = [
        {
          id: 'pay-001',
          paymentNumber: 'PAY-2024-001',
          clientName: 'Acme Corporation',
          invoiceNumber: 'INV-2024-001',
          amount: '5400.00',
          method: 'bank_transfer',
          paidAt: '2024-01-15T10:30:00Z',
          status: 'completed',
          matchedTransactionId: 'txn-001'
        },
        {
          id: 'pay-002',
          paymentNumber: 'PAY-2024-002',
          clientName: 'TechStart Inc',
          invoiceNumber: 'INV-2024-002',
          amount: '3780.00',
          method: 'credit_card',
          paidAt: '2024-01-16T14:20:00Z',
          status: 'completed',
          matchedTransactionId: 'txn-002'
        },
        {
          id: 'pay-003',
          paymentNumber: 'PAY-2024-003',
          clientName: 'Global Solutions LLC',
          invoiceNumber: 'INV-2024-003',
          amount: '2000.00',
          method: 'check',
          paidAt: '2024-01-17T09:00:00Z',
          status: 'completed'
        },
        {
          id: 'pay-004',
          paymentNumber: 'PAY-2024-004',
          clientName: 'Digital Ventures Co',
          invoiceNumber: 'INV-2024-004',
          amount: '2500.00',
          method: 'paypal',
          paidAt: '2024-01-18T11:30:00Z',
          status: 'completed'
        },
        {
          id: 'pay-005',
          paymentNumber: 'PAY-2024-005',
          clientName: 'Acme Corporation',
          invoiceNumber: 'INV-2024-005',
          amount: '3200.00',
          method: 'bank_transfer',
          paidAt: '2024-01-20T10:30:00Z',
          status: 'completed'
        }
      ];
      
      // Mock match suggestions
      matchSuggestions = [
        { transactionId: 'txn-003', paymentId: 'pay-003', confidence: 95, reason: 'Exact amount match' },
        { transactionId: 'txn-004', paymentId: 'pay-004', confidence: 90, reason: 'Amount and date match' },
        { transactionId: 'txn-006', paymentId: 'pay-005', confidence: 88, reason: 'Amount and client match' }
      ];
      
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load reconciliation data';
    } finally {
      loading = false;
    }
  }

  function getStatusVariant(status: string): 'green' | 'gray' | 'yellow' | 'red' | 'blue' {
    switch (status) {
      case 'matched':
      case 'completed':
        return 'green';
      case 'unmatched':
      case 'pending':
        return 'yellow';
      case 'disputed':
      case 'failed':
        return 'red';
      case 'refunded':
        return 'blue';
      default:
        return 'gray';
    }
  }

  function formatCurrency(value: string): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(parseFloat(value) || 0);
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  function handleFileSelect(event: CustomEvent<File>) {
    // TODO: Process bank statement file
    console.log('File selected:', event.detail);
    // Simulate processing
    setTimeout(() => {
      alert('Bank statement uploaded successfully!');
    }, 500);
  }

  function handleTransactionSelect(transaction: BankTransaction) {
    if (selectedTransactionId === transaction.id) {
      selectedTransactionId = null;
    } else {
      selectedTransactionId = transaction.id;
      // Check for auto-match suggestion
      const suggestion = matchSuggestions.find(s => s.transactionId === transaction.id);
      if (suggestion && transaction.status === 'unmatched') {
        selectedPaymentId = suggestion.paymentId;
      }
    }
  }

  function handlePaymentSelect(payment: SystemPayment) {
    if (selectedPaymentId === payment.id) {
      selectedPaymentId = null;
    } else {
      selectedPaymentId = payment.id;
    }
  }

  function handleMatch() {
    if (!selectedTransactionId || !selectedPaymentId) return;
    confirmAction = 'match';
    showConfirmModal = true;
  }

  function handleUnmatch(transaction: BankTransaction) {
    selectedTransactionId = transaction.id;
    confirmAction = 'unmatch';
    showConfirmModal = true;
  }

  async function confirmMatchAction() {
    processing = true;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      if (confirmAction === 'match' && selectedTransactionId && selectedPaymentId) {
        // Update transaction
        bankTransactions = bankTransactions.map(t => 
          t.id === selectedTransactionId 
            ? { ...t, status: 'matched', matchedPaymentId: selectedPaymentId }
            : t
        );
        
        // Update payment
        systemPayments = systemPayments.map(p => 
          p.id === selectedPaymentId 
            ? { ...p, matchedTransactionId: selectedTransactionId }
            : p
        );
        
        // Remove suggestion
        matchSuggestions = matchSuggestions.filter(s => 
          s.transactionId !== selectedTransactionId && s.paymentId !== selectedPaymentId
        );
        
        selectedTransactionId = null;
        selectedPaymentId = null;
      } else if (confirmAction === 'unmatch' && selectedTransactionId) {
        const transaction = bankTransactions.find(t => t.id === selectedTransactionId);
        if (transaction?.matchedPaymentId) {
          // Update transaction
          bankTransactions = bankTransactions.map(t => 
            t.id === selectedTransactionId 
              ? { ...t, status: 'unmatched', matchedPaymentId: undefined }
              : t
          );
          
          // Update payment
          systemPayments = systemPayments.map(p => 
            p.id === transaction.matchedPaymentId 
              ? { ...p, matchedTransactionId: undefined }
              : p
          );
        }
        selectedTransactionId = null;
      }
      
      showConfirmModal = false;
      confirmAction = null;
    } catch (err) {
      error = 'Failed to process match action';
    } finally {
      processing = false;
    }
  }

  function handleMarkReconciled() {
    confirmAction = 'reconcile';
    showConfirmModal = true;
  }

  async function confirmReconcile() {
    processing = true;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 1000));
      alert('Reconciliation completed successfully!');
      showConfirmModal = false;
      confirmAction = null;
    } catch (err) {
      error = 'Failed to complete reconciliation';
    } finally {
      processing = false;
    }
  }

  function handleExport() {
    const csvContent = [
      ['Date', 'Description', 'Reference', 'Amount', 'Status', 'Matched Payment'].join(','),
      ...bankTransactions.map(t => [
        t.date,
        `"${t.description}"`,
        t.reference || '',
        t.amount,
        t.status,
        t.matchedPaymentId || ''
      ].join(','))
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `reconciliation-report-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    
    showExportModal = true;
    setTimeout(() => showExportModal = false, 2000);
  }

  function getMatchSuggestion(transactionId: string): MatchSuggestion | undefined {
    return matchSuggestions.find(s => s.transactionId === transactionId);
  }

  function canMatch(): boolean {
    return !!selectedTransactionId && !!selectedPaymentId;
  }

  function applyDateRange() {
    loadData();
  }

  onMount(() => {
    // Set default date range (last 30 days)
    const today = new Date();
    const thirtyDaysAgo = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000);
    dateTo = today.toISOString().split('T')[0];
    dateFrom = thirtyDaysAgo.toISOString().split('T')[0];
    
    loadData();
  });
</script>

<svelte:head>
  <title>Payment Reconciliation | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Payment Reconciliation</h1>
      <p class="page-description">Match bank transactions with system payments</p>
    </div>
    <div class="header-actions">
      <Button variant="secondary" on:click={handleExport}>
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
        </svg>
        Export Report
      </Button>
      <Button variant="primary" on:click={handleMarkReconciled} disabled={stats.unmatchedCount > 0}>
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        Mark Reconciled
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  <!-- Date Range & Filters -->
  <Card class="filters-card">
    <div class="filters-row">
      <div class="date-range">
        <span class="filter-label">Period:</span>
        <DatePicker
          id="date-from"
          label=""
          bind:value={dateFrom}
        />
        <span class="date-separator">to</span>
        <DatePicker
          id="date-to"
          label=""
          bind:value={dateTo}
        />
        <Button variant="secondary" size="sm" on:click={applyDateRange}>
          Apply
        </Button>
      </div>
      <div class="status-filter">
        <Select
          id="status-filter"
          label="Filter by Status"
          options={statusOptions}
          bind:value={statusFilter}
        />
      </div>
    </div>
  </Card>

  <!-- Statistics -->
  <div class="stats-grid">
    <Card class="stat-card">
      <div class="stat-content">
        <span class="stat-label">Total Transactions</span>
        <span class="stat-value">{stats.totalTransactions}</span>
        <span class="stat-amount">{formatCurrency(stats.totalBankAmount.toFixed(2))}</span>
      </div>
    </Card>
    <Card class="stat-card">
      <div class="stat-content">
        <span class="stat-label">Matched</span>
        <span class="stat-value success">{stats.matchedCount}</span>
        <span class="stat-amount">{formatCurrency(stats.matchedAmount.toFixed(2))}</span>
      </div>
    </Card>
    <Card class="stat-card">
      <div class="stat-content">
        <span class="stat-label">Unmatched</span>
        <span class="stat-value warning">{stats.unmatchedCount}</span>
        <span class="stat-amount">
          {formatCurrency(
            bankTransactions
              .filter(t => t.status === 'unmatched')
              .reduce((sum, t) => sum + parseFloat(t.amount), 0)
              .toFixed(2)
          )}
        </span>
      </div>
    </Card>
    <Card class="stat-card">
      <div class="stat-content">
        <span class="stat-label">Disputed</span>
        <span class="stat-value danger">{stats.disputedCount}</span>
        <span class="stat-amount">
          {formatCurrency(
            bankTransactions
              .filter(t => t.status === 'disputed')
              .reduce((sum, t) => sum + parseFloat(t.amount), 0)
              .toFixed(2)
          )}
        </span>
      </div>
    </Card>
  </div>

  <!-- Match Action Bar -->
  {#if selectedTransactionId && selectedPaymentId}
    <div class="match-action-bar" transition:fade>
      <div class="match-info">
        <span class="match-label">Ready to match:</span>
        <Chip variant="primary" size="md">
          Transaction: {bankTransactions.find(t => t.id === selectedTransactionId)?.description.substring(0, 20)}...
        </Chip>
        <span class="match-arrow">→</span>
        <Chip variant="primary" size="md">
          Payment: {systemPayments.find(p => p.id === selectedPaymentId)?.paymentNumber}
        </Chip>
      </div>
      <Button variant="primary" on:click={handleMatch}>
        Match Selected
      </Button>
    </div>
  {/if}

  <!-- Two-Panel Layout -->
  <div class="panels-grid">
    <!-- Bank Transactions Panel -->
    <Card class="panel-card">
      <div class="panel-header">
        <h2 class="panel-title">
          Bank Transactions
          <Badge variant="gray" size="sm">{filteredTransactions.length}</Badge>
        </h2>
        <FileUpload
          id="bank-statement"
          label=""
          accept=".csv,.ofx,.qfx"
          on:select={handleFileSelect}
        />
      </div>

      {#if loading}
        <div class="loading-container">
          <Spinner size="lg" />
          <p>Loading transactions...</p>
        </div>
      {:else if filteredTransactions.length === 0}
        <div class="empty-container">
          <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p class="text-gray-500">No transactions found</p>
        </div>
      {:else}
        <div class="transactions-list">
          {#each filteredTransactions as transaction}
            {@const suggestion = getMatchSuggestion(transaction.id)}
            <div 
              class="transaction-item"
              class:selected={selectedTransactionId === transaction.id}
              class:matched={transaction.status === 'matched'}
              class:unmatched={transaction.status === 'unmatched'}
              class:disputed={transaction.status === 'disputed'}
              on:click={() => handleTransactionSelect(transaction)}
              on:keydown={(e) => e.key === 'Enter' && handleTransactionSelect(transaction)}
              role="button"
              tabindex="0"
            >
              <div class="transaction-header">
                <div class="transaction-main">
                  <span class="transaction-date">{formatDate(transaction.date)}</span>
                  <span class="transaction-description">{transaction.description}</span>
                  {#if transaction.reference}
                    <span class="transaction-reference">Ref: {transaction.reference}</span>
                  {/if}
                </div>
                <div class="transaction-amount">
                  <span class="amount">{formatCurrency(transaction.amount)}</span>
                  <Badge variant={getStatusVariant(transaction.status)} size="sm">
                    {transaction.status}
                  </Badge>
                </div>
              </div>
              
              {#if suggestion && transaction.status === 'unmatched'}
                <div class="suggestion-badge">
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  <span>Suggested match ({suggestion.confidence}% confidence)</span>
                </div>
              {/if}
              
              {#if transaction.status === 'matched' && transaction.matchedPaymentId}
                <div class="matched-info">
                  <span>Matched to: {systemPayments.find(p => p.id === transaction.matchedPaymentId)?.paymentNumber}</span>
                  <Button variant="ghost" size="sm" on:click={(e) => { e.stopPropagation(); handleUnmatch(transaction); }}>
                    Unmatch
                  </Button>
                </div>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </Card>

    <!-- System Payments Panel -->
    <Card class="panel-card">
      <div class="panel-header">
        <h2 class="panel-title">
          System Payments
          <Badge variant="gray" size="sm">{unmatchedPayments.length}</Badge>
        </h2>
        <span class="panel-subtitle">Unmatched payments</span>
      </div>

      {#if loading}
        <div class="loading-container">
          <Spinner size="lg" />
          <p>Loading payments...</p>
        </div>
      {:else if unmatchedPayments.length === 0}
        <div class="empty-container">
          <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
          <p class="text-gray-500">No unmatched payments</p>
        </div>
      {:else}
        <div class="payments-list">
          {#each unmatchedPayments as payment}
            <div 
              class="payment-item"
              class:selected={selectedPaymentId === payment.id}
              on:click={() => handlePaymentSelect(payment)}
              on:keydown={(e) => e.key === 'Enter' && handlePaymentSelect(payment)}
              role="button"
              tabindex="0"
            >
              <div class="payment-header">
                <div class="payment-main">
                  <span class="payment-number">{payment.paymentNumber}</span>
                  <span class="payment-client">{payment.clientName}</span>
                  <span class="payment-invoice">{payment.invoiceNumber}</span>
                </div>
                <div class="payment-amount">
                  <span class="amount">{formatCurrency(payment.amount)}</span>
                  <span class="method">{payment.method}</span>
                </div>
              </div>
              <div class="payment-date">
                {formatDate(payment.paidAt)}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </Card>
  </div>

  <!-- Auto-Match Suggestions -->
  {#if matchSuggestions.length > 0}
    <Card class="suggestions-card">
      <h3 class="suggestions-title">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        Auto-Match Suggestions ({matchSuggestions.length})
      </h3>
      <div class="suggestions-list">
        {#each matchSuggestions as suggestion}
          {@const transaction = bankTransactions.find(t => t.id === suggestion.transactionId)}
          {@const payment = systemPayments.find(p => p.id === suggestion.paymentId)}
          {#if transaction && payment}
            <div class="suggestion-item">
              <div class="suggestion-match">
                <span class="suggestion-transaction">{transaction.description}</span>
                <span class="suggestion-arrow">↔</span>
                <span class="suggestion-payment">{payment.paymentNumber}</span>
              </div>
              <div class="suggestion-details">
                <span class="suggestion-amount">{formatCurrency(transaction.amount)}</span>
                <Badge variant="green" size="sm">{suggestion.confidence}% match</Badge>
                <span class="suggestion-reason">{suggestion.reason}</span>
              </div>
              <Button 
                variant="primary" 
                size="sm" 
                on:click={() => {
                  selectedTransactionId = suggestion.transactionId;
                  selectedPaymentId = suggestion.paymentId;
                  handleMatch();
                }}
              >
                Apply Match
              </Button>
            </div>
          {/if}
        {/each}
      </div>
    </Card>
  {/if}
</div>

<!-- Confirmation Modal -->
<Modal
  bind:open={showConfirmModal}
  title={confirmAction === 'match' ? 'Confirm Match' : confirmAction === 'unmatch' ? 'Confirm Unmatch' : 'Complete Reconciliation'}
  size="sm"
>
  {#if confirmAction === 'match'}
    <p>Match the selected transaction with the selected payment?</p>
    <p class="text-sm text-gray-500 mt-2">This will mark both items as matched.</p>
  {:else if confirmAction === 'unmatch'}
    <p>Remove the match for this transaction?</p>
    <p class="text-sm text-gray-500 mt-2">Both items will be marked as unmatched.</p>
  {:else if confirmAction === 'reconcile'}
    <p>Mark reconciliation as complete?</p>
    <p class="text-sm text-gray-500 mt-2">
      {#if stats.unmatchedCount > 0}
        <span class="text-yellow-600">Warning: There are still {stats.unmatchedCount} unmatched transactions.</span>
      {:else}
        All transactions have been matched. This action cannot be undone.
      {/if}
    </p>
  {/if}
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showConfirmModal = false; }} disabled={processing}>
      Cancel
    </Button>
    <Button 
      variant={confirmAction === 'unmatch' ? 'danger' : 'primary'} 
      on:click={confirmAction === 'reconcile' ? confirmReconcile : confirmMatchAction}
      loading={processing}
    >
      {processing ? 'Processing...' : 'Confirm'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Export Success Modal -->
<Modal
  bind:open={showExportModal}
  title="Export Complete"
  size="sm"
>
  <div class="export-success">
    <svg class="w-12 h-12 text-green-500 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
    <p>Reconciliation report exported successfully!</p>
  </div>
</Modal>

<script context="module">
  import { fade } from 'svelte/transition';
</script>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1600px;
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
    flex-wrap: wrap;
    gap: 1rem;
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
    flex-wrap: wrap;
  }

  :global(.filters-card) {
    padding: 1rem;
    margin-bottom: 1rem;
  }

  .filters-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .date-range {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .filter-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-700);
  }

  .date-separator {
    color: var(--color-gray-500);
  }

  .status-filter {
    min-width: 200px;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
    margin-bottom: 1rem;
  }

  :global(.stat-card) {
    padding: 1rem;
  }

  .stat-content {
    display: flex;
    flex-direction: column;
  }

  .stat-label {
    font-size: 0.875rem;
    color: var(--color-gray-500);
    margin-bottom: 0.5rem;
  }

  .stat-value {
    font-size: 2rem;
    font-weight: 700;
    color: var(--color-gray-900);
  }

  .stat-value.success {
    color: var(--color-green-600);
  }

  .stat-value.warning {
    color: var(--color-yellow-600);
  }

  .stat-value.danger {
    color: var(--color-red-600);
  }

  .stat-amount {
    font-size: 0.875rem;
    color: var(--color-gray-600);
    margin-top: 0.25rem;
  }

  .match-action-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background-color: var(--color-primary-50);
    border: 1px solid var(--color-primary-200);
    border-radius: 0.5rem;
    margin-bottom: 1rem;
  }

  .match-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .match-label {
    font-size: 0.875rem;
    color: var(--color-gray-600);
  }

  .match-arrow {
    color: var(--color-gray-400);
  }

  .panels-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  :global(.panel-card) {
    padding: 1rem;
    min-height: 500px;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--color-gray-200);
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .panel-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .panel-subtitle {
    font-size: 0.875rem;
    color: var(--color-gray-500);
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

  .transactions-list,
  .payments-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-height: 500px;
    overflow-y: auto;
  }

  .transaction-item,
  .payment-item {
    padding: 1rem;
    border: 2px solid var(--color-gray-200);
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .transaction-item:hover,
  .payment-item:hover {
    border-color: var(--color-primary-300);
    background-color: var(--color-gray-50);
  }

  .transaction-item.selected,
  .payment-item.selected {
    border-color: var(--color-primary-500);
    background-color: var(--color-primary-50);
  }

  .transaction-item.matched {
    border-left: 4px solid var(--color-green-500);
  }

  .transaction-item.unmatched {
    border-left: 4px solid var(--color-yellow-500);
  }

  .transaction-item.disputed {
    border-left: 4px solid var(--color-red-500);
  }

  .transaction-header,
  .payment-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
  }

  .transaction-main,
  .payment-main {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    flex: 1;
  }

  .transaction-date,
  .payment-date {
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  .transaction-description,
  .payment-number {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .transaction-reference,
  .payment-invoice {
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  .payment-client {
    font-size: 0.875rem;
    color: var(--color-gray-700);
  }

  .transaction-amount,
  .payment-amount {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.25rem;
  }

  .amount {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .method {
    font-size: 0.75rem;
    color: var(--color-gray-500);
    text-transform: capitalize;
  }

  .suggestion-badge {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    margin-top: 0.5rem;
    padding: 0.375rem 0.75rem;
    background-color: var(--color-blue-100);
    color: var(--color-blue-700);
    border-radius: 0.25rem;
    font-size: 0.75rem;
    font-weight: 500;
    width: fit-content;
  }

  .matched-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 0.5rem;
    padding-top: 0.5rem;
    border-top: 1px solid var(--color-gray-200);
    font-size: 0.75rem;
    color: var(--color-green-600);
  }

  :global(.suggestions-card) {
    padding: 1rem;
  }

  .suggestions-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .suggestions-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .suggestion-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .suggestion-match {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex: 1;
    flex-wrap: wrap;
  }

  .suggestion-transaction,
  .suggestion-payment {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-900);
  }

  .suggestion-arrow {
    color: var(--color-gray-400);
  }

  .suggestion-details {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .suggestion-amount {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .suggestion-reason {
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  .export-success {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 1rem;
    text-align: center;
  }

  @media (max-width: 1200px) {
    .stats-grid {
      grid-template-columns: repeat(2, 1fr);
    }

    .panels-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .page-header {
      flex-direction: column;
    }

    .stats-grid {
      grid-template-columns: 1fr;
    }

    .filters-row {
      flex-direction: column;
      align-items: stretch;
    }

    .date-range {
      flex-direction: column;
      align-items: stretch;
    }

    .match-action-bar {
      flex-direction: column;
      gap: 1rem;
    }

    .suggestion-item {
      flex-direction: column;
      align-items: stretch;
    }

    .suggestion-match {
      flex-direction: column;
      align-items: flex-start;
    }
  }
</style>
