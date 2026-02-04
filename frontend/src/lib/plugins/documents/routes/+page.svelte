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
  import FileUpload from '$lib/shared/components/forms/FileUpload.svelte';

  interface Document {
    id: string;
    name: string;
    type: 'invoice' | 'po' | 'receipt' | 'contract' | 'scanned' | 'other';
    size: number;
    contentType: string;
    status: 'pending' | 'processing' | 'completed' | 'failed';
    uploadedBy: string;
    uploadedAt: string;
    tags: string[];
    metadata: {
      pageCount?: number;
      extractedText?: string;
    };
  }

  let documents: Document[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery = '';
  let typeFilter: string = '';
  let statusFilter: string = '';
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;
  let showUploadModal = false;
  let deleteDocumentId: string | null = null;

  const typeOptions = [
    { value: '', label: 'All Types' },
    { value: 'invoice', label: 'Invoice' },
    { value: 'po', label: 'Purchase Order' },
    { value: 'receipt', label: 'Receipt' },
    { value: 'contract', label: 'Contract' },
    { value: 'scanned', label: 'Scanned' },
    { value: 'other', label: 'Other' }
  ];

  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'pending', label: 'Pending' },
    { value: 'processing', label: 'Processing' },
    { value: 'completed', label: 'Completed' },
    { value: 'failed', label: 'Failed' }
  ];

  const columns = [
    { key: 'name', label: 'Name', sortable: true },
    { key: 'type', label: 'Type', sortable: true },
    { key: 'size', label: 'Size', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'uploadedBy', label: 'Uploaded By', sortable: true },
    { key: 'uploadedAt', label: 'Date', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  function getStatusVariant(status: string): 'green' | 'gray' | 'yellow' | 'red' | 'blue' {
    switch (status) {
      case 'completed': return 'green';
      case 'processing': return 'blue';
      case 'pending': return 'yellow';
      case 'failed': return 'red';
      default: return 'gray';
    }
  }

  function getTypeIcon(type: string): string {
    const icons: Record<string, string> = {
      invoice: '📄',
      po: '📋',
      receipt: '🧾',
      contract: '📑',
      scanned: '📷',
      other: '📎'
    };
    return icons[type] || '📎';
  }

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  function formatDate(date: string): string {
    return new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  async function loadDocuments() {
    loading = true;
    error = null;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      documents = [
        {
          id: '1',
          name: 'Invoice_2024_001.pdf',
          type: 'invoice',
          size: 245760,
          contentType: 'application/pdf',
          status: 'completed',
          uploadedBy: 'John Doe',
          uploadedAt: '2024-01-15T10:30:00Z',
          tags: ['invoice', '2024', 'client-a'],
          metadata: { pageCount: 2, extractedText: 'Invoice #001...' }
        },
        {
          id: '2',
          name: 'Purchase_Order_12345.pdf',
          type: 'po',
          size: 512000,
          contentType: 'application/pdf',
          status: 'completed',
          uploadedBy: 'Jane Smith',
          uploadedAt: '2024-01-14T16:45:00Z',
          tags: ['po', 'supplier'],
          metadata: { pageCount: 3 }
        },
        {
          id: '3',
          name: 'Contract_Agreement_2024.docx',
          type: 'contract',
          size: 1048576,
          contentType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
          status: 'processing',
          uploadedBy: 'Bob Wilson',
          uploadedAt: '2024-01-13T09:15:00Z',
          tags: ['contract', 'legal'],
          metadata: {}
        },
        {
          id: '4',
          name: 'Receipt_Store_001.jpg',
          type: 'receipt',
          size: 1536000,
          contentType: 'image/jpeg',
          status: 'completed',
          uploadedBy: 'Alice Brown',
          uploadedAt: '2024-01-12T14:20:00Z',
          tags: ['receipt', 'expense'],
          metadata: {}
        },
        {
          id: '5',
          name: 'Scanned_Document_001.pdf',
          type: 'scanned',
          size: 3145728,
          contentType: 'application/pdf',
          status: 'failed',
          uploadedBy: 'John Doe',
          uploadedAt: '2024-01-11T11:00:00Z',
          tags: ['scanned'],
          metadata: {}
        }
      ];
      
      totalItems = documents.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load documents';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    currentPage = 1;
    loadDocuments();
  }

  function handleRowClick(document: Document) {
    window.location.href = `/documents/${document.id}`;
  }

  function handleDownload(document: Document, event: Event) {
    event.stopPropagation();
    // Simulate download
    console.log('Downloading:', document.name);
  }

  async function handleDelete(document: Document, event: Event) {
    event.stopPropagation();
    deleteDocumentId = document.id;
  }

  async function confirmDelete() {
    if (!deleteDocumentId) return;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      documents = documents.filter(d => d.id !== deleteDocumentId);
      totalItems = documents.length;
      deleteDocumentId = null;
    } catch (err) {
      error = 'Failed to delete document';
    }
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    loadDocuments();
  }

  function handleFileSelect(event: CustomEvent) {
    console.log('Files selected:', event.detail);
  }

  onMount(() => {
    loadDocuments();
  });
</script>

<svelte:head>
  <title>Documents | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Documents</h1>
      <p class="page-description">Manage and organize your documents</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={() => showUploadModal = true}>
        Upload Document
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
            placeholder="Search documents..."
            bind:value={searchQuery}
            on:keydown={(e) => e.key === 'Enter' && handleSearch()}
          />
        </div>
        <div class="filter-item">
          <Select
            id="type"
            label="Type"
            options={typeOptions}
            bind:value={typeFilter}
            on:change={() => handleSearch()}
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
        <p>Loading documents...</p>
      </div>
    {:else if documents.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <p class="text-gray-500 mb-4">No documents found</p>
        <Button variant="primary" on:click={() => showUploadModal = true}>
          Upload Your First Document
        </Button>
      </div>
    {:else}
      <Table {columns}>
        <tbody>
          {#each documents as document}
            <tr on:click={() => handleRowClick(document)} class="clickable-row">
              <td>
                <div class="flex items-center gap-3">
                  <span class="text-2xl">{getTypeIcon(document.type)}</span>
                  <div>
                    <div class="font-medium">{document.name}</div>
                    {#if document.tags.length > 0}
                      <div class="flex gap-1 mt-1">
                        {#each document.tags.slice(0, 3) as tag}
                          <span class="text-xs px-2 py-0.5 bg-gray-100 rounded-full text-gray-600">{tag}</span>
                        {/each}
                        {#if document.tags.length > 3}
                          <span class="text-xs text-gray-400">+{document.tags.length - 3}</span>
                        {/if}
                      </div>
                    {/if}
                  </div>
                </div>
              </td>
              <td class="capitalize">{document.type.replace('_', ' ')}</td>
              <td>{formatFileSize(document.size)}</td>
              <td>
                <Badge variant={getStatusVariant(document.status)}>
                  {document.status}
                </Badge>
              </td>
              <td>{document.uploadedBy}</td>
              <td class="text-sm text-gray-500">{formatDate(document.uploadedAt)}</td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={(e) => handleDownload(document, e)}>
                    Download
                  </Button>
                  <Button variant="ghost" size="sm" on:click={(e) => handleDelete(document, e)}>
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
  bind:open={showUploadModal}
  title="Upload Document"
  size="lg"
>
  <FileUpload
    id="document-upload"
    label="Select or drop files"
    accept=".pdf,.doc,.docx,.jpg,.jpeg,.png"
    multiple={true}
    on:select={handleFileSelect}
  />
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close}>Cancel</Button>
    <Button variant="primary" on:click={() => {
      showUploadModal = false;
    }}>Upload</Button>
  </svelte:fragment>
</Modal>

{#if deleteDocumentId}
  <Modal
    open={true}
    title="Delete Document"
    size="sm"
  >
    <p>Are you sure you want to delete this document? This action cannot be undone.</p>
    
    <svelte:fragment slot="footer" let:close>
      <Button variant="secondary" on:click={() => { close(); deleteDocumentId = null; }}>Cancel</Button>
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
