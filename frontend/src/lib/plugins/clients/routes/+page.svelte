<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import { clientsStore } from '../stores';
  import type { Client, ClientStatus } from '$lib/shared/api/clients';

  let searchQuery = '';
  let statusFilter: ClientStatus | '' = '';
  let page = 1;

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'suspended', label: 'Suspended' },
    { value: 'merged', label: 'Merged' }
  ];

  const columns = [
    { key: 'code', label: 'Code', sortable: true },
    { key: 'name', label: 'Name', sortable: true },
    { key: 'email', label: 'Email', sortable: true },
    { key: 'phone', label: 'Phone', sortable: false },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'creditLimit', label: 'Credit Limit', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  function getStatusBadgeVariant(status: ClientStatus): 'gray' | 'green' | 'yellow' | 'red' {
    switch (status) {
      case 'active': return 'green';
      case 'inactive': return 'gray';
      case 'suspended': return 'red';
      case 'merged': return 'yellow';
      default: return 'gray';
    }
  }

  function formatCurrency(value: string): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(parseFloat(value) || 0);
  }

  async function handleSearch() {
    await clientsStore.loadClients({
      search: searchQuery || undefined,
      status: statusFilter || undefined,
      page
    });
  }

  function handleRowClick(client: Client) {
    window.location.href = `/clients/${client.id}`;
  }

  function handleEdit(client: Client, event: Event) {
    event.stopPropagation();
    window.location.href = `/clients/${client.id}/edit`;
  }

  async function handleDelete(client: Client, event: Event) {
    event.stopPropagation();
    if (confirm(`Are you sure you want to delete ${client.name}?`)) {
      await clientsStore.deleteClient(client.id);
    }
  }

  function handlePageChange(newPage: number) {
    page = newPage;
    handleSearch();
  }

  onMount(() => {
    handleSearch();
  });
</script>

<svelte:head>
  <title>Clients | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Clients</h1>
      <p class="page-description">Manage your customer relationships</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={() => window.location.href = '/clients/new'}>
        Add Client
      </Button>
    </div>
  </div>

  <Card>
    <div class="filters">
      <div class="filter-row">
          <div class="filter-item search-filter">
          <Input
            id="search"
            label="Search"
            type="search"
            placeholder="Search clients..."
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

    {#if $clientsStore.loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading clients...</p>
      </div>
    {:else if $clientsStore.error}
      <div class="error-container">
        <p class="error-message">{$clientsStore.error}</p>
        <Button variant="secondary" on:click={handleSearch}>
          Retry
        </Button>
      </div>
    {:else if $clientsStore.clients.length === 0}
      <div class="empty-container">
        <p>No clients found</p>
        <Button variant="primary" on:click={() => window.location.href = '/clients/new'}>
          Add Your First Client
        </Button>
      </div>
    {:else}
      <Table {columns}>
        <tbody>
          {#each $clientsStore.clients as client}
            <tr on:click={() => handleRowClick(client)} class="clickable-row">
              <td>{client.code}</td>
              <td>{client.name}</td>
              <td>{client.email}</td>
              <td>{client.phone || '-'}</td>
              <td>
                <Badge variant={getStatusBadgeVariant(client.status)}>
                  {client.status}
                </Badge>
              </td>
              <td>{formatCurrency(client.creditLimit)}</td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={(e) => handleEdit(client, e)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={(e) => handleDelete(client, e)}>
                    Delete
                  </Button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </Table>

      {#if $clientsStore.pagination.totalPages > 1}
        <div class="pagination">
          <Button
            variant="secondary"
            size="sm"
            disabled={page <= 1}
            on:click={() => handlePageChange(page - 1)}
          >
            Previous
          </Button>
          <span class="pagination-info">
            Page {page} of {$clientsStore.pagination.totalPages}
          </span>
          <Button
            variant="secondary"
            size="sm"
            disabled={page >= $clientsStore.pagination.totalPages}
            on:click={() => handlePageChange(page + 1)}
          >
            Next
          </Button>
        </div>
      {/if}
    {/if}
  </Card>
</div>

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
  .error-container,
  .empty-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .error-message {
    color: var(--color-red-600);
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

  .pagination {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--color-gray-200);
  }

  .pagination-info {
    color: var(--color-gray-600);
    font-size: 0.875rem;
  }
</style>
