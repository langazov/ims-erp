<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import DatePicker from '$lib/shared/components/forms/DatePicker.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Chip from '$lib/shared/components/display/Chip.svelte';
  import Pagination from '$lib/shared/components/data/Pagination.svelte';
  import { cn } from '$lib/shared/utils/helpers';

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
    description?: string;
    metadata?: {
      pageCount?: number;
      extractedText?: string;
      ocrConfidence?: number;
    };
  }

  interface SavedSearch {
    id: string;
    name: string;
    filters: SearchFilters;
    createdAt: string;
  }

  interface SearchFilters {
    query: string;
    type: string;
    status: string;
    uploader: string;
    dateFrom: string;
    dateTo: string;
    tags: string[];
    hasOcr: boolean | null;
    minOcrConfidence: number | null;
  }

  let documents: Document[] = [];
  let loading = false;
  let error: string | null = null;
  let showFilters = false;
  let savedSearches: SavedSearch[] = [];
  let showSaveSearchModal = false;
  let newSearchName = '';
  
  // Pagination
  let currentPage = 1;
  let totalPages = 1;
  let totalItems = 0;
  let pageSize = 10;

  // Filters
  let filters: SearchFilters = {
    query: '',
    type: '',
    status: '',
    uploader: '',
    dateFrom: '',
    dateTo: '',
    tags: [],
    hasOcr: null,
    minOcrConfidence: null
  };

  let currentTag = '';
  let searchHighlight = '';

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

  const uploaderOptions = [
    { value: '', label: 'All Uploaders' },
    { value: 'john.doe', label: 'John Doe' },
    { value: 'jane.smith', label: 'Jane Smith' },
    { value: 'bob.wilson', label: 'Bob Wilson' },
    { value: 'alice.brown', label: 'Alice Brown' }
  ];

  const ocrOptions = [
    { value: '', label: 'Any' },
    { value: true, label: 'Has OCR Text' },
    { value: false, label: 'No OCR Text' }
  ];

  const confidenceOptions = [
    { value: '', label: 'Any Confidence' },
    { value: 90, label: '90% or higher' },
    { value: 80, label: '80% or higher' },
    { value: 70, label: '70% or higher' },
    { value: 50, label: '50% or higher' }
  ];

  const columns = [
    { key: 'name', label: 'Name', sortable: true },
    { key: 'type', label: 'Type', sortable: true },
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
      year: 'numeric'
    });
  }

  function getTypeLabel(type: string): string {
    const labels: Record<string, string> = {
      invoice: 'Invoice',
      po: 'Purchase Order',
      receipt: 'Receipt',
      contract: 'Contract',
      scanned: 'Scanned',
      other: 'Other'
    };
    return labels[type] || type;
  }

  function highlightText(text: string, query: string): string {
    if (!query) return text;
    const regex = new RegExp(`(${escapeRegex(query)})`, 'gi');
    return text.replace(regex, '<mark>$1</mark>');
  }

  function escapeRegex(string: string): string {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }

  function addTag() {
    const trimmed = currentTag.trim().toLowerCase();
    if (trimmed && !filters.tags.includes(trimmed)) {
      filters.tags = [...filters.tags, trimmed];
    }
    currentTag = '';
  }

  function removeTag(tag: string) {
    filters.tags = filters.tags.filter(t => t !== tag);
  }

  function handleTagKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      addTag();
    } else if (event.key === 'Backspace' && !currentTag && filters.tags.length > 0) {
      filters.tags = filters.tags.slice(0, -1);
    }
  }

  function hasActiveFilters(): boolean {
    return !!(
      filters.type ||
      filters.status ||
      filters.uploader ||
      filters.dateFrom ||
      filters.dateTo ||
      filters.tags.length > 0 ||
      filters.hasOcr !== null ||
      filters.minOcrConfidence !== null
    );
  }

  function clearFilters() {
    filters = {
      query: '',
      type: '',
      status: '',
      uploader: '',
      dateFrom: '',
      dateTo: '',
      tags: [],
      hasOcr: null,
      minOcrConfidence: null
    };
    currentTag = '';
    performSearch();
  }

  async function performSearch() {
    loading = true;
    error = null;
    searchHighlight = filters.query;

    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500));

      // Mock search results
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
          metadata: { pageCount: 2, extractedText: 'Invoice #001...', ocrConfidence: 0.95 }
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
          metadata: { pageCount: 3, ocrConfidence: 0.88 }
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
          metadata: { ocrConfidence: 0.72 }
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

      // Filter results based on search criteria
      if (filters.query) {
        const query = filters.query.toLowerCase();
        documents = documents.filter(d => 
          d.name.toLowerCase().includes(query) ||
          d.description?.toLowerCase().includes(query) ||
          d.metadata?.extractedText?.toLowerCase().includes(query)
        );
      }

      if (filters.type) {
        documents = documents.filter(d => d.type === filters.type);
      }

      if (filters.status) {
        documents = documents.filter(d => d.status === filters.status);
      }

      if (filters.uploader) {
        documents = documents.filter(d => d.uploadedBy.toLowerCase().replace(' ', '.') === filters.uploader);
      }

      if (filters.dateFrom) {
        const fromDate = new Date(filters.dateFrom);
        documents = documents.filter(d => new Date(d.uploadedAt) >= fromDate);
      }

      if (filters.dateTo) {
        const toDate = new Date(filters.dateTo);
        toDate.setHours(23, 59, 59, 999);
        documents = documents.filter(d => new Date(d.uploadedAt) <= toDate);
      }

      if (filters.tags.length > 0) {
        documents = documents.filter(d => 
          filters.tags.some(tag => d.tags.includes(tag))
        );
      }

      if (filters.hasOcr !== null) {
        documents = documents.filter(d => 
          filters.hasOcr ? !!d.metadata?.extractedText : !d.metadata?.extractedText
        );
      }

      if (filters.minOcrConfidence !== null) {
        documents = documents.filter(d => 
          (d.metadata?.ocrConfidence || 0) >= (filters.minOcrConfidence || 0) / 100
        );
      }

      totalItems = documents.length;
      totalPages = Math.ceil(totalItems / pageSize);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to search documents';
    } finally {
      loading = false;
    }
  }

  function handleSearchInput(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      currentPage = 1;
      performSearch();
    }
  }

  function handlePageChange(newPage: number) {
    currentPage = newPage;
    performSearch();
  }

  function handleRowClick(document: Document) {
    goto(`/documents/${document.id}`);
  }

  function handleDownload(document: Document, event: Event) {
    event.stopPropagation();
    console.log('Downloading:', document.name);
  }

  function exportResults() {
    const csvContent = [
      ['Name', 'Type', 'Size', 'Status', 'Uploaded By', 'Date', 'Tags'].join(','),
      ...documents.map(d => [
        `"${d.name}"`,
        d.type,
        formatFileSize(d.size),
        d.status,
        d.uploadedBy,
        formatDate(d.uploadedAt),
        `"${d.tags.join(', ')}"`
      ].join(','))
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `document-search-results-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  function saveSearch() {
    if (!newSearchName.trim()) return;
    
    const newSearch: SavedSearch = {
      id: Math.random().toString(36).substring(2, 9),
      name: newSearchName,
      filters: { ...filters },
      createdAt: new Date().toISOString()
    };
    
    savedSearches = [...savedSearches, newSearch];
    newSearchName = '';
    showSaveSearchModal = false;
  }

  function loadSavedSearch(search: SavedSearch) {
    filters = { ...search.filters };
    performSearch();
  }

  function deleteSavedSearch(id: string) {
    savedSearches = savedSearches.filter(s => s.id !== id);
  }

  onMount(() => {
    // Load saved searches from localStorage
    const saved = localStorage.getItem('documentSavedSearches');
    if (saved) {
      try {
        savedSearches = JSON.parse(saved);
      } catch {
        savedSearches = [];
      }
    }
    performSearch();
  });

  $: {
    // Save to localStorage whenever savedSearches changes
    if (typeof window !== 'undefined') {
      localStorage.setItem('documentSavedSearches', JSON.stringify(savedSearches));
    }
  }
</script>

<svelte:head>
  <title>Search Documents | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Search Documents</h1>
      <p class="page-description">Advanced search across all your documents</p>
    </div>
    <div class="header-actions">
      <Button variant="secondary" on:click={() => showSaveSearchModal = true} disabled={!hasActiveFilters()}>
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
        </svg>
        Save Search
      </Button>
      <Button variant="secondary" on:click={exportResults} disabled={documents.length === 0}>
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
        </svg>
        Export
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  <!-- Search Bar -->
  <Card class="search-card">
    <div class="search-bar">
      <div class="search-input-wrapper">
        <svg class="search-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          type="search"
          class="search-input"
          placeholder="Search by name, description, or content..."
          bind:value={filters.query}
          on:keydown={handleSearchInput}
        />
      </div>
      <Button variant="primary" on:click={() => performSearch()} loading={loading}>
        Search
      </Button>
      <Button 
        variant="secondary" 
        on:click={() => showFilters = !showFilters}
        class={cn('filter-toggle', hasActiveFilters() && 'has-filters')}
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
        </svg>
        Filters
        {#if hasActiveFilters()}
          <span class="filter-badge">!</span>
        {/if}
      </Button>
    </div>

    <!-- Advanced Filters -->
    {#if showFilters}
      <div class="advanced-filters">
        <div class="filter-grid">
          <div class="filter-group">
            <Select
              id="type-filter"
              label="Document Type"
              options={typeOptions}
              bind:value={filters.type}
            />
          </div>
          <div class="filter-group">
            <Select
              id="status-filter"
              label="Status"
              options={statusOptions}
              bind:value={filters.status}
            />
          </div>
          <div class="filter-group">
            <Select
              id="uploader-filter"
              label="Uploaded By"
              options={uploaderOptions}
              bind:value={filters.uploader}
            />
          </div>
          <div class="filter-group">
            <Select
              id="ocr-filter"
              label="OCR Status"
              options={ocrOptions}
              bind:value={filters.hasOcr}
            />
          </div>
          <div class="filter-group">
            <Select
              id="confidence-filter"
              label="OCR Confidence"
              options={confidenceOptions}
              bind:value={filters.minOcrConfidence}
            />
          </div>
        </div>

        <div class="date-filters">
          <div class="filter-group">
            <DatePicker
              id="date-from"
              label="From Date"
              bind:value={filters.dateFrom}
            />
          </div>
          <div class="filter-group">
            <DatePicker
              id="date-to"
              label="To Date"
              bind:value={filters.dateTo}
            />
          </div>
        </div>

        <div class="tags-filter">
          <label class="filter-label">Tags</label>
          <div class="tags-input-container">
            {#each filters.tags as tag}
              <Chip variant="primary" size="md" removable on:click={() => removeTag(tag)}>
                {tag}
              </Chip>
            {/each}
            <input
              type="text"
              class="tags-input"
              placeholder={filters.tags.length === 0 ? 'Add tags to filter...' : ''}
              bind:value={currentTag}
              on:keydown={handleTagKeydown}
              on:blur={addTag}
            />
          </div>
        </div>

        <div class="filter-actions">
          <Button variant="ghost" size="sm" on:click={clearFilters}>
            Clear All Filters
          </Button>
          <Button variant="primary" size="sm" on:click={() => { currentPage = 1; performSearch(); }}>
            Apply Filters
          </Button>
        </div>
      </div>
    {/if}
  </Card>

  <!-- Saved Searches -->
  {#if savedSearches.length > 0}
    <div class="saved-searches">
      <span class="saved-label">Saved Searches:</span>
      {#each savedSearches as search}
        <div class="saved-search-chip">
          <button class="saved-search-btn" on:click={() => loadSavedSearch(search)}>
            {search.name}
          </button>
          <button class="remove-saved-btn" on:click={() => deleteSavedSearch(search.id)}>
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Active Filters Display -->
  {#if hasActiveFilters()}
    <div class="active-filters">
      <span class="active-label">Active Filters:</span>
      {#if filters.type}
        <Chip variant="info" size="sm" removable on:click={() => { filters.type = ''; performSearch(); }}>
          Type: {getTypeLabel(filters.type)}
        </Chip>
      {/if}
      {#if filters.status}
        <Chip variant="info" size="sm" removable on:click={() => { filters.status = ''; performSearch(); }}>
          Status: {filters.status}
        </Chip>
      {/if}
      {#if filters.uploader}
        <Chip variant="info" size="sm" removable on:click={() => { filters.uploader = ''; performSearch(); }}>
          Uploader: {filters.uploader}
        </Chip>
      {/if}
      {#if filters.dateFrom}
        <Chip variant="info" size="sm" removable on:click={() => { filters.dateFrom = ''; performSearch(); }}>
          From: {formatDate(filters.dateFrom)}
        </Chip>
      {/if}
      {#if filters.dateTo}
        <Chip variant="info" size="sm" removable on:click={() => { filters.dateTo = ''; performSearch(); }}>
          To: {formatDate(filters.dateTo)}
        </Chip>
      {/if}
      {#each filters.tags as tag}
        <Chip variant="info" size="sm" removable on:click={() => removeTag(tag)}>
          Tag: {tag}
        </Chip>
      {/each}
    </div>
  {/if}

  <!-- Results -->
  <Card class="results-card">
    <div class="results-header">
      <h2 class="results-title">
        Search Results
        <span class="results-count">({totalItems} found)</span>
      </h2>
    </div>

    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Searching documents...</p>
      </div>
    {:else if documents.length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <p class="text-gray-500 mb-2">No documents found</p>
        <p class="text-gray-400 text-sm">Try adjusting your search criteria</p>
      </div>
    {:else}
      <Table {columns}>
        <tbody>
          {#each documents as doc}
            <tr on:click={() => handleRowClick(doc)} class="clickable-row">
              <td>
                <div class="flex items-center gap-3">
                  <span class="text-2xl">{getTypeIcon(doc.type)}</span>
                  <div>
                    <div class="font-medium">
                      {@html highlightText(doc.name, searchHighlight)}
                    </div>
                    {#if doc.tags.length > 0}
                      <div class="flex gap-1 mt-1">
                        {#each doc.tags.slice(0, 3) as tag}
                          <span class="text-xs px-2 py-0.5 bg-gray-100 rounded-full text-gray-600">{tag}</span>
                        {/each}
                        {#if doc.tags.length > 3}
                          <span class="text-xs text-gray-400">+{doc.tags.length - 3}</span>
                        {/if}
                      </div>
                    {/if}
                    {#if doc.metadata?.extractedText && searchHighlight}
                      {@const highlighted = highlightText(doc.metadata.extractedText.substring(0, 100), searchHighlight)}
                      {#if highlighted.includes('<mark>')}
                        <p class="text-xs text-gray-500 mt-1">
                          {@html highlighted}...
                        </p>
                      {/if}
                    {/if}
                  </div>
                </div>
              </td>
              <td class="capitalize">{getTypeLabel(doc.type)}</td>
              <td>
                <Badge variant={getStatusVariant(doc.status)}>
                  {doc.status}
                </Badge>
              </td>
              <td>{doc.uploadedBy}</td>
              <td class="text-sm text-gray-500">{formatDate(doc.uploadedAt)}</td>
              <td>
                <div class="actions-cell">
                  <Button variant="ghost" size="sm" on:click={(e) => handleDownload(doc, e)}>
                    Download
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

<!-- Save Search Modal -->
{#if showSaveSearchModal}
  <div class="modal-overlay" on:click={() => showSaveSearchModal = false}>
    <div class="modal-content" on:click|stopPropagation>
      <div class="modal-header">
        <h3 class="modal-title">Save Search</h3>
        <button class="modal-close" on:click={() => showSaveSearchModal = false}>
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="modal-body">
        <Input
          id="search-name"
          label="Search Name"
          type="text"
          placeholder="e.g., Q1 Invoices"
          bind:value={newSearchName}
          required
        />
      </div>
      <div class="modal-footer">
        <Button variant="secondary" on:click={() => showSaveSearchModal = false}>
          Cancel
        </Button>
        <Button variant="primary" on:click={saveSearch} disabled={!newSearchName.trim()}>
          Save
        </Button>
      </div>
    </div>
  </div>
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

  :global(.search-card) {
    padding: 1.25rem;
    margin-bottom: 1rem;
  }

  .search-bar {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
  }

  .search-input-wrapper {
    flex: 1;
    position: relative;
  }

  .search-icon {
    position: absolute;
    left: 1rem;
    top: 50%;
    transform: translateY(-50%);
    width: 1.25rem;
    height: 1.25rem;
    color: var(--color-gray-400);
  }

  .search-input {
    width: 100%;
    padding: 0.625rem 1rem 0.625rem 2.75rem;
    border: 1px solid var(--color-gray-300);
    border-radius: 0.5rem;
    font-size: 0.9375rem;
    background-color: white;
    transition: all 0.15s ease;
  }

  .search-input:focus {
    outline: none;
    border-color: var(--color-primary-500);
    box-shadow: 0 0 0 3px var(--color-primary-100);
  }

  :global(.filter-toggle) {
    position: relative;
  }

  :global(.filter-toggle.has-filters) {
    border-color: var(--color-primary-500);
    color: var(--color-primary-600);
  }

  .filter-badge {
    position: absolute;
    top: -4px;
    right: -4px;
    width: 16px;
    height: 16px;
    background-color: var(--color-primary-500);
    color: white;
    border-radius: 50%;
    font-size: 10px;
    font-weight: 600;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .advanced-filters {
    margin-top: 1.25rem;
    padding-top: 1.25rem;
    border-top: 1px solid var(--color-gray-200);
  }

  .filter-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .date-filters {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .tags-filter {
    margin-bottom: 1rem;
  }

  .filter-label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-700);
    margin-bottom: 0.5rem;
  }

  .tags-input-container {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    padding: 0.5rem;
    border: 1px solid var(--color-gray-300);
    border-radius: 0.5rem;
    background-color: white;
    min-height: 2.75rem;
  }

  .tags-input-container:focus-within {
    border-color: var(--color-primary-500);
    box-shadow: 0 0 0 2px var(--color-primary-100);
  }

  .tags-input {
    flex: 1;
    min-width: 100px;
    border: none;
    background: none;
    font-size: 0.875rem;
    color: var(--color-gray-900);
    outline: none;
  }

  .tags-input::placeholder {
    color: var(--color-gray-400);
  }

  .filter-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }

  .saved-searches {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .saved-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-600);
  }

  .saved-search-chip {
    display: flex;
    align-items: center;
    background-color: var(--color-gray-100);
    border-radius: 9999px;
    overflow: hidden;
  }

  .saved-search-btn {
    padding: 0.25rem 0.75rem;
    font-size: 0.875rem;
    color: var(--color-gray-700);
    background: none;
    border: none;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .saved-search-btn:hover {
    background-color: var(--color-gray-200);
    color: var(--color-gray-900);
  }

  .remove-saved-btn {
    padding: 0.25rem;
    color: var(--color-gray-400);
    background: none;
    border: none;
    border-left: 1px solid var(--color-gray-200);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .remove-saved-btn:hover {
    color: var(--color-red-500);
    background-color: var(--color-red-50);
  }

  .active-filters {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .active-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-600);
  }

  :global(.results-card) {
    padding: 1.25rem;
  }

  .results-header {
    margin-bottom: 1rem;
  }

  .results-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0;
  }

  .results-count {
    font-weight: 400;
    color: var(--color-gray-500);
    font-size: 0.875rem;
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

  /* Modal Styles */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
  }

  .modal-content {
    background-color: white;
    border-radius: 0.75rem;
    width: 100%;
    max-width: 400px;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.25rem;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .modal-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0;
  }

  .modal-close {
    padding: 0.25rem;
    color: var(--color-gray-400);
    background: none;
    border: none;
    cursor: pointer;
    border-radius: 0.25rem;
    transition: all 0.15s ease;
  }

  .modal-close:hover {
    color: var(--color-gray-600);
    background-color: var(--color-gray-100);
  }

  .modal-body {
    padding: 1.25rem;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 1rem 1.25rem;
    border-top: 1px solid var(--color-gray-200);
  }

  :global(mark) {
    background-color: var(--color-yellow-200);
    color: var(--color-gray-900);
    padding: 0.125rem 0.25rem;
    border-radius: 0.125rem;
  }

  @media (max-width: 768px) {
    .page-header {
      flex-direction: column;
    }

    .search-bar {
      flex-wrap: wrap;
    }

    .search-input-wrapper {
      width: 100%;
    }

    .filter-grid {
      grid-template-columns: 1fr;
    }

    .date-filters {
      grid-template-columns: 1fr;
    }
  }
</style>
