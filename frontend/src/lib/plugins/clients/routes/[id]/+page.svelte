<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import { clientsStore } from '../../stores';
  import type { Client, ClientStatus, UpdateClientRequest } from '$lib/shared/api/clients';

  let client: Client | null = null;
  let loading = true;
  let editing = false;
  let saving = false;

  let editName = '';
  let editEmail = '';
  let editPhone = '';
  let editCreditLimit = '';
  let editStatus: ClientStatus = 'active';

  const statusOptions = [
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'suspended', label: 'Suspended' }
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

  function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }

  function startEditing() {
    if (client) {
      editName = client.name;
      editEmail = client.email;
      editPhone = client.phone;
      editCreditLimit = client.creditLimit;
      editStatus = client.status;
      editing = true;
    }
  }

  function cancelEditing() {
    editing = false;
  }

  async function saveChanges() {
    if (!client) return;

    const data: UpdateClientRequest = {};
    if (editName !== client.name) data.name = editName;
    if (editEmail !== client.email) data.email = editEmail;
    if (editPhone !== client.phone) data.phone = editPhone;
    if (parseFloat(editCreditLimit) !== parseFloat(client.creditLimit)) {
      data.creditLimit = parseFloat(editCreditLimit);
    }
    if (editStatus !== client.status) data.status = editStatus;

    saving = true;
    try {
      await clientsStore.updateClient(client.id, data);
      client = await clientsStore.loadClient(client.id);
      editing = false;
    } catch (error) {
      console.error('Failed to update client:', error);
    } finally {
      saving = false;
    }
  }

  async function handleDelete() {
    if (!client) return;
    if (confirm(`Are you sure you want to delete ${client.name}?`)) {
      await clientsStore.deleteClient(client.id);
      goto('/clients');
    }
  }

  onMount(async () => {
    const clientId = $page.params.id;
    try {
      client = await clientsStore.loadClient(clientId);
    } catch (error) {
      console.error('Failed to load client:', error);
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>{client ? client.name : 'Client'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading client...</p>
    </div>
  {:else if !client}
    <div class="error-container">
      <p>Client not found</p>
      <Button variant="secondary" on:click={() => goto('/clients')}>
        Back to Clients
      </Button>
    </div>
  {:else}
    <div class="page-header">
      <div class="header-content">
        <div class="header-nav">
          <Button variant="ghost" size="sm" on:click={() => goto('/clients')}>
            ← Back to Clients
          </Button>
        </div>
        <h1 class="page-title">{client.name}</h1>
        <div class="header-meta">
          <Badge variant={getStatusBadgeVariant(client.status)}>
            {client.status}
          </Badge>
          <span class="meta-item">Code: {client.code}</span>
          <span class="meta-item">Created: {formatDate(client.createdAt)}</span>
        </div>
      </div>
      <div class="header-actions">
        {#if editing}
          <Button variant="secondary" on:click={cancelEditing} disabled={saving}>
            Cancel
          </Button>
          <Button variant="primary" on:click={saveChanges} loading={saving}>
            {saving ? 'Saving...' : 'Save Changes'}
          </Button>
        {:else}
          <Button variant="secondary" on:click={startEditing}>
            Edit
          </Button>
          <Button variant="danger" on:click={handleDelete}>
            Delete
          </Button>
        {/if}
      </div>
    </div>

    <div class="content-grid">
      <Card>
        <h2 class="card-title">Contact Information</h2>
        {#if editing}
          <div class="edit-form">
            <Input
              id="name"
              label="Name"
              type="text"
              bind:value={editName}
              required
            />
            <Input
              id="email"
              label="Email"
              type="email"
              bind:value={editEmail}
              required
            />
            <Input
              id="phone"
              label="Phone"
              type="tel"
              bind:value={editPhone}
            />
            <Select
              id="status"
              label="Status"
              options={statusOptions}
              bind:value={editStatus}
            />
          </div>
        {:else}
          <dl class="info-list">
            <div class="info-item">
              <dt>Email</dt>
              <dd><a href="mailto:{client.email}">{client.email}</a></dd>
            </div>
            <div class="info-item">
              <dt>Phone</dt>
              <dd>{client.phone || '-'}</dd>
            </div>
            <div class="info-item">
              <dt>Status</dt>
              <dd>
                <Badge variant={getStatusBadgeVariant(client.status)}>
                  {client.status}
                </Badge>
              </dd>
            </div>
          </dl>
        {/if}
      </Card>

      <Card>
        <h2 class="card-title">Credit Information</h2>
        {#if editing}
          <Input
            id="creditLimit"
            label="Credit Limit"
            type="number"
            bind:value={editCreditLimit}
            min="0"
            step="0.01"
          />
        {:else}
          <dl class="info-list">
            <div class="info-item">
              <dt>Credit Limit</dt>
              <dd>{formatCurrency(client.creditLimit)}</dd>
            </div>
            <div class="info-item">
              <dt>Current Balance</dt>
              <dd>{formatCurrency(client.currentBalance)}</dd>
            </div>
            <div class="info-item">
              <dt>Available Credit</dt>
              <dd class="available-credit">
                {formatCurrency((parseFloat(client.creditLimit) - parseFloat(client.currentBalance)).toString())}
              </dd>
            </div>
          </dl>
        {/if}
      </Card>

      <Card>
        <h2 class="card-title">Billing Address</h2>
        {#if client.billingAddress && (client.billingAddress.street || client.billingAddress.city)}
          <address class="address">
            {#if client.billingAddress.street}
              <span>{client.billingAddress.street}</span>
            {/if}
            {#if client.billingAddress.city || client.billingAddress.state || client.billingAddress.postalCode}
              <span>
                {#if client.billingAddress.city}{client.billingAddress.city}{/if}
                {#if client.billingAddress.city && client.billingAddress.state}, {/if}
                {#if client.billingAddress.state}{client.billingAddress.state}{/if}
                {#if client.billingAddress.postalCode} {client.billingAddress.postalCode}{/if}
              </span>
            {/if}
            {#if client.billingAddress.country}
              <span>{client.billingAddress.country}</span>
            {/if}
          </address>
        {:else}
          <p class="no-address">No billing address on file</p>
        {/if}
      </Card>

      <Card>
        <h2 class="card-title">Tags</h2>
        {#if client.tags && client.tags.length > 0}
          <div class="tags">
            {#each client.tags as tag}
              <Badge variant="gray">{tag}</Badge>
            {/each}
          </div>
        {:else}
          <p class="no-tags">No tags assigned</p>
        {/if}
      </Card>
    </div>
  {/if}
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .loading-container,
  .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .header-nav {
    margin-bottom: 0.5rem;
  }

  .page-title {
    font-size: 1.875rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin: 0;
  }

  .header-meta {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-top: 0.5rem;
    flex-wrap: wrap;
  }

  .meta-item {
    color: var(--color-gray-500);
    font-size: 0.875rem;
  }

  .header-actions {
    display: flex;
    gap: 0.75rem;
  }

  .content-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
  }

  .card-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1rem 0;
  }

  .edit-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .info-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .info-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--color-gray-100);
  }

  .info-item:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }

  .info-item dt {
    color: var(--color-gray-500);
    font-size: 0.875rem;
  }

  .info-item dd {
    color: var(--color-gray-900);
    font-weight: 500;
    margin: 0;
  }

  .info-item dd a {
    color: var(--color-primary-600);
    text-decoration: none;
  }

  .info-item dd a:hover {
    text-decoration: underline;
  }

  .available-credit {
    color: var(--color-green-600);
    font-weight: 600;
  }

  .address {
    font-style: normal;
    line-height: 1.6;
    color: var(--color-gray-700);
  }

  .address span {
    display: block;
  }

  .no-address,
  .no-tags {
    color: var(--color-gray-400);
    font-style: italic;
  }

  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  @media (max-width: 768px) {
    .content-grid {
      grid-template-columns: 1fr;
    }

    .page-header {
      flex-direction: column;
    }

    .header-actions {
      width: 100%;
      justify-content: flex-end;
    }
  }
</style>
