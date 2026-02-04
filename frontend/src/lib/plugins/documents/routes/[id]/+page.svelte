<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Textarea from '$lib/shared/components/forms/Textarea.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';
  import Tabs from '$lib/shared/components/layout/Tabs.svelte';
  import Chip from '$lib/shared/components/display/Chip.svelte';
  import { cn } from '$lib/shared/utils/helpers';

  interface DocumentVersion {
    id: string;
    version: number;
    uploadedBy: string;
    uploadedAt: string;
    size: number;
    changeNotes?: string;
  }

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
    versions: DocumentVersion[];
  }

  const documentId = $page.params.id;

  let document: Document | null = null;
  let loading = true;
  let error: string | null = null;
  let activeTab = 'preview';
  let showDeleteModal = false;
  let showShareModal = false;
  let deleting = false;
  let shareEmail = '';
  let sharePermission: 'view' | 'edit' = 'view';
  let isEditingTags = false;
  let newTag = '';
  let isEditingDescription = false;
  let editedDescription = '';

  const tabs = [
    { id: 'preview', label: 'Preview' },
    { id: 'details', label: 'Details' },
    { id: 'versions', label: 'Versions' },
    { id: 'activity', label: 'Activity' }
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

  function getTypeLabel(type: string): string {
    const labels: Record<string, string> = {
      invoice: 'Invoice',
      po: 'Purchase Order',
      receipt: 'Receipt',
      contract: 'Contract',
      scanned: 'Scanned Document',
      other: 'Other'
    };
    return labels[type] || type;
  }

  function canPreview(contentType: string): boolean {
    return contentType.startsWith('image/') || contentType === 'application/pdf';
  }

  async function loadDocument() {
    loading = true;
    error = null;
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500));
      
      document = {
        id: documentId,
        name: 'Invoice_2024_001.pdf',
        type: 'invoice',
        size: 245760,
        contentType: 'application/pdf',
        status: 'completed',
        uploadedBy: 'John Doe',
        uploadedAt: '2024-01-15T10:30:00Z',
        tags: ['invoice', '2024', 'client-a', 'q1'],
        description: 'Quarterly invoice for Client A services rendered in Q1 2024.',
        metadata: {
          pageCount: 2,
          extractedText: 'Invoice #001...',
          ocrConfidence: 0.95
        },
        versions: [
          {
            id: 'v3',
            version: 3,
            uploadedBy: 'John Doe',
            uploadedAt: '2024-01-15T10:30:00Z',
            size: 245760,
            changeNotes: 'Final version with corrected totals'
          },
          {
            id: 'v2',
            version: 2,
            uploadedBy: 'John Doe',
            uploadedAt: '2024-01-15T09:15:00Z',
            size: 243200,
            changeNotes: 'Updated line items'
          },
          {
            id: 'v1',
            version: 1,
            uploadedBy: 'John Doe',
            uploadedAt: '2024-01-14T16:45:00Z',
            size: 240640,
            changeNotes: 'Initial upload'
          }
        ]
      };
      
      editedDescription = document.description || '';
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load document';
    } finally {
      loading = false;
    }
  }

  function handleDownload() {
    // Simulate download
    console.log('Downloading:', document?.name);
  }

  function handleShare() {
    showShareModal = true;
  }

  function handleDelete() {
    showDeleteModal = true;
  }

  async function confirmDelete() {
    deleting = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      goto('/documents');
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete document';
      deleting = false;
      showDeleteModal = false;
    }
  }

  function handleSendShare() {
    console.log('Sharing with:', shareEmail, 'permission:', sharePermission);
    showShareModal = false;
    shareEmail = '';
  }

  function addTag() {
    const trimmed = newTag.trim().toLowerCase();
    if (trimmed && document && !document.tags.includes(trimmed)) {
      document.tags = [...document.tags, trimmed];
    }
    newTag = '';
  }

  function removeTag(tag: string) {
    if (document) {
      document.tags = document.tags.filter(t => t !== tag);
    }
  }

  function handleTagKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      addTag();
    }
  }

  function saveDescription() {
    if (document) {
      document.description = editedDescription;
      isEditingDescription = false;
    }
  }

  function cancelEditDescription() {
    editedDescription = document?.description || '';
    isEditingDescription = false;
  }

  function restoreVersion(version: DocumentVersion) {
    console.log('Restoring version:', version.version);
  }

  onMount(() => {
    loadDocument();
  });
</script>

<svelte:head>
  <title>{document ? document.name : 'Document Details'} | ERP System</title>
</svelte:head>

<div class="page-container">
  {#if loading}
    <div class="loading-container">
      <Spinner size="lg" />
      <p>Loading document details...</p>
    </div>
  {:else if error}
    <div class="error-container">
      <Alert variant="error">{error}</Alert>
      <Button variant="secondary" on:click={() => goto('/documents')}>
        Back to Documents
      </Button>
    </div>
  {:else if document}
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <span class="document-icon">{getTypeIcon(document.type)}</span>
          <div class="header-text">
            <h1 class="page-title" title={document.name}>{document.name}</h1>
            <div class="header-meta">
              <Badge variant={getStatusVariant(document.status)} size="sm">
                {document.status}
              </Badge>
              <span class="meta-item">{getTypeLabel(document.type)}</span>
              <span class="meta-item">{formatFileSize(document.size)}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="header-actions">
        <Button variant="secondary" on:click={handleShare}>
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
          </svg>
          Share
        </Button>
        <Button variant="secondary" on:click={handleDownload}>
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
          Download
        </Button>
        <Button variant="danger" on:click={handleDelete}>
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          Delete
        </Button>
      </div>
    </div>

    <Tabs {tabs} bind:activeTab>
      {#if activeTab === 'preview'}
        <div class="preview-section">
          {#if canPreview(document.contentType)}
            <Card class="preview-card">
              <div class="preview-placeholder">
                <svg class="preview-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p class="preview-text">Document Preview</p>
                <p class="preview-hint">{document.name}</p>
                <Button variant="secondary" size="sm" on:click={handleDownload}>
                  Open Document
                </Button>
              </div>
            </Card>
          {:else}
            <Card class="preview-card">
              <div class="preview-placeholder no-preview">
                <svg class="preview-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <p class="preview-text">Preview not available</p>
                <p class="preview-hint">This file type cannot be previewed. Download to view.</p>
                <Button variant="secondary" size="sm" on:click={handleDownload}>
                  Download Document
                </Button>
              </div>
            </Card>
          {/if}

          <!-- Quick Info -->
          <div class="quick-info-grid">
            <Card class="info-card">
              <h3 class="info-title">Document Info</h3>
              <dl class="info-list">
                <div class="info-item">
                  <dt>Type</dt>
                  <dd>{getTypeLabel(document.type)}</dd>
                </div>
                <div class="info-item">
                  <dt>Size</dt>
                  <dd>{formatFileSize(document.size)}</dd>
                </div>
                <div class="info-item">
                  <dt>Pages</dt>
                  <dd>{document.metadata?.pageCount || 'N/A'}</dd>
                </div>
                <div class="info-item">
                  <dt>OCR Confidence</dt>
                  <dd>
                    {#if document.metadata?.ocrConfidence}
                      <span class={cn(
                        'confidence-badge',
                        document.metadata.ocrConfidence >= 0.9 ? 'high' :
                        document.metadata.ocrConfidence >= 0.7 ? 'medium' : 'low'
                      )}>
                        {Math.round(document.metadata.ocrConfidence * 100)}%
                      </span>
                    {:else}
                      N/A
                    {/if}
                  </dd>
                </div>
              </dl>
            </Card>

            <Card class="info-card">
              <h3 class="info-title">Upload Info</h3>
              <dl class="info-list">
                <div class="info-item">
                  <dt>Uploaded By</dt>
                  <dd>{document.uploadedBy}</dd>
                </div>
                <div class="info-item">
                  <dt>Upload Date</dt>
                  <dd>{formatDate(document.uploadedAt)}</dd>
                </div>
                <div class="info-item">
                  <dt>Document ID</dt>
                  <dd class="mono">{document.id}</dd>
                </div>
                <div class="info-item">
                  <dt>Content Type</dt>
                  <dd class="mono">{document.contentType}</dd>
                </div>
              </dl>
            </Card>
          </div>
        </div>
      {:else if activeTab === 'details'}
        <div class="details-section">
          <Card>
            <div class="details-header">
              <h3 class="section-title">Description</h3>
              {#if !isEditingDescription}
                <Button variant="ghost" size="sm" on:click={() => isEditingDescription = true}>
                  Edit
                </Button>
              {/if}
            </div>
            
            {#if isEditingDescription}
              <div class="description-edit">
                <Textarea
                  id="description"
                  bind:value={editedDescription}
                  rows={4}
                  placeholder="Add a description..."
                />
                <div class="edit-actions">
                  <Button variant="secondary" size="sm" on:click={cancelEditDescription}>
                    Cancel
                  </Button>
                  <Button variant="primary" size="sm" on:click={saveDescription}>
                    Save
                  </Button>
                </div>
              </div>
            {:else}
              <p class="description-text">
                {document.description || 'No description provided.'}
              </p>
            {/if}
          </Card>

          <Card class="tags-card">
            <div class="tags-header">
              <h3 class="section-title">Tags</h3>
              <Button 
                variant="ghost" 
                size="sm" 
                on:click={() => isEditingTags = !isEditingTags}
              >
                {isEditingTags ? 'Done' : 'Edit'}
              </Button>
            </div>
            
            <div class="tags-container">
              {#each document.tags as tag}
                <Chip 
                  variant="primary" 
                  size="md" 
                  removable={isEditingTags}
                  on:click={() => removeTag(tag)}
                >
                  {tag}
                </Chip>
              {/each}
              
              {#if isEditingTags}
                <input
                  type="text"
                  class="tag-input"
                  placeholder="Add tag..."
                  bind:value={newTag}
                  on:keydown={handleTagKeydown}
                  on:blur={addTag}
                />
              {/if}
            </div>
          </Card>

          {#if document.metadata?.extractedText}
            <Card class="extracted-text-card">
              <h3 class="section-title">Extracted Text</h3>
              <div class="extracted-text">
                <pre>{document.metadata.extractedText}</pre>
              </div>
            </Card>
          {/if}
        </div>
      {:else if activeTab === 'versions'}
        <div class="versions-section">
          <Card>
            <div class="versions-header">
              <h3 class="section-title">Version History</h3>
              <Button variant="secondary" size="sm">
                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                </svg>
                Upload New Version
              </Button>
            </div>

            <div class="versions-list">
              {#each document.versions as version, index}
                <div class="version-item" class:is-current={index === 0}>
                  <div class="version-icon">
                    <span class="version-number">v{version.version}</span>
                    {#if index === 0}
                      <span class="current-badge">Current</span>
                    {/if}
                  </div>
                  
                  <div class="version-info">
                    <div class="version-meta">
                      <span class="version-uploader">{version.uploadedBy}</span>
                      <span class="version-date">{formatDate(version.uploadedAt)}</span>
                    </div>
                    <div class="version-size">{formatFileSize(version.size)}</div>
                    {#if version.changeNotes}
                      <p class="version-notes">{version.changeNotes}</p>
                    {/if}
                  </div>
                  
                  <div class="version-actions">
                    <Button variant="ghost" size="sm" on:click={handleDownload}>
                      Download
                    </Button>
                    {#if index !== 0}
                      <Button 
                        variant="ghost" 
                        size="sm"
                        on:click={() => restoreVersion(version)}
                      >
                        Restore
                      </Button>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          </Card>
        </div>
      {:else if activeTab === 'activity'}
        <div class="activity-section">
          <Card>
            <h3 class="section-title">Activity Log</h3>
            <div class="activity-list">
              <div class="activity-item">
                <div class="activity-icon success">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4-4m0 0L8 8m4-4v12" />
                  </svg>
                </div>
                <div class="activity-content">
                  <p class="activity-text">Document uploaded</p>
                  <p class="activity-meta">{document.uploadedBy} • {formatDate(document.uploadedAt)}</p>
                </div>
              </div>
              
              <div class="activity-item">
                <div class="activity-icon info">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div class="activity-content">
                  <p class="activity-text">OCR processing completed</p>
                  <p class="activity-meta">System • {formatDate(document.uploadedAt)}</p>
                </div>
              </div>
              
              <div class="activity-item">
                <div class="activity-icon success">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <div class="activity-content">
                  <p class="activity-text">Document indexed for search</p>
                  <p class="activity-meta">System • {formatDate(document.uploadedAt)}</p>
                </div>
              </div>
            </div>
          </Card>
        </div>
      {/if}
    </Tabs>
  {:else}
    <Alert variant="error">Document not found</Alert>
  {/if}
</div>

<!-- Share Modal -->
<Modal
  bind:open={showShareModal}
  title="Share Document"
  size="md"
>
  <div class="share-form">
    <Input
      id="share-email"
      label="Email Address"
      type="email"
      placeholder="Enter email address"
      bind:value={shareEmail}
    />
    
    <div class="share-permissions">
      <label class="permission-label">Permission Level</label>
      <div class="permission-options">
        <label class="permission-option" class:selected={sharePermission === 'view'}>
          <input
            type="radio"
            name="permission"
            value="view"
            bind:group={sharePermission}
          />
          <div class="permission-content">
            <span class="permission-title">View Only</span>
            <span class="permission-desc">Can view and download</span>
          </div>
        </label>
        <label class="permission-option" class:selected={sharePermission === 'edit'}>
          <input
            type="radio"
            name="permission"
            value="edit"
            bind:group={sharePermission}
          />
          <div class="permission-content">
            <span class="permission-title">Edit</span>
            <span class="permission-desc">Can view, download, and edit</span>
          </div>
        </label>
      </div>
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close}>Cancel</Button>
    <Button 
      variant="primary" 
      on:click={handleSendShare}
      disabled={!shareEmail}
    >
      Send Invite
    </Button>
  </svelte:fragment>
</Modal>

<!-- Delete Modal -->
<Modal
  bind:open={showDeleteModal}
  title="Delete Document"
  size="sm"
>
  <p>Are you sure you want to delete <strong>{document?.name}</strong>?</p>
  <p class="delete-warning">This action cannot be undone. The document and all its versions will be permanently deleted.</p>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showDeleteModal = false; }} disabled={deleting}>
      Cancel
    </Button>
    <Button variant="danger" on:click={confirmDelete} loading={deleting}>
      {deleting ? 'Deleting...' : 'Delete'}
    </Button>
  </svelte:fragment>
</Modal>

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
  }

  .header-content {
    flex: 1;
    min-width: 0;
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .document-icon {
    font-size: 2rem;
    flex-shrink: 0;
  }

  .header-text {
    min-width: 0;
  }

  .page-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .header-meta {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-top: 0.25rem;
    flex-wrap: wrap;
  }

  .meta-item {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .header-actions {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .preview-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  :global(.preview-card) {
    padding: 0;
    overflow: hidden;
  }

  .preview-placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem;
    gap: 1rem;
    background: linear-gradient(135deg, var(--color-gray-50) 0%, var(--color-gray-100) 100%);
  }

  .preview-placeholder.no-preview {
    background: var(--color-yellow-50);
  }

  .preview-icon {
    width: 4rem;
    height: 4rem;
    color: var(--color-gray-400);
  }

  .preview-placeholder.no-preview .preview-icon {
    color: var(--color-yellow-500);
  }

  .preview-text {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-700);
    margin: 0;
  }

  .preview-hint {
    font-size: 0.875rem;
    color: var(--color-gray-500);
    margin: 0;
  }

  .quick-info-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  :global(.info-card) {
    padding: 1.25rem;
  }

  .info-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-700);
    margin: 0 0 1rem 0;
    text-transform: uppercase;
    letter-spacing: 0.025em;
  }

  .info-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .info-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .info-item dt {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .info-item dd {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-900);
    margin: 0;
  }

  .info-item dd.mono {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 0.75rem;
    color: var(--color-gray-600);
  }

  .confidence-badge {
    padding: 0.125rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .confidence-badge.high {
    background-color: var(--color-green-100);
    color: var(--color-green-700);
  }

  .confidence-badge.medium {
    background-color: var(--color-yellow-100);
    color: var(--color-yellow-700);
  }

  .confidence-badge.low {
    background-color: var(--color-red-100);
    color: var(--color-red-700);
  }

  .details-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  :global(.tags-card),
  :global(.extracted-text-card) {
    padding: 1.25rem;
  }

  .details-header,
  .tags-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .section-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0;
  }

  .description-text {
    color: var(--color-gray-600);
    line-height: 1.6;
    margin: 0;
  }

  .description-edit {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .edit-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }

  .tags-container {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .tag-input {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--color-gray-300);
    border-radius: 9999px;
    font-size: 0.875rem;
    outline: none;
    min-width: 100px;
  }

  .tag-input:focus {
    border-color: var(--color-primary-500);
    box-shadow: 0 0 0 2px var(--color-primary-100);
  }

  .extracted-text {
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
    padding: 1rem;
    max-height: 300px;
    overflow-y: auto;
  }

  .extracted-text pre {
    margin: 0;
    font-size: 0.875rem;
    color: var(--color-gray-700);
    white-space: pre-wrap;
    word-wrap: break-word;
  }

  .versions-section,
  .activity-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .versions-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .versions-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .version-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
    border: 1px solid transparent;
  }

  .version-item.is-current {
    border-color: var(--color-primary-200);
    background-color: var(--color-primary-50);
  }

  .version-icon {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
  }

  .version-number {
    width: 2.5rem;
    height: 2.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--color-primary-600);
    color: white;
    border-radius: 50%;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .current-badge {
    font-size: 0.625rem;
    font-weight: 600;
    color: var(--color-primary-600);
    text-transform: uppercase;
  }

  .version-info {
    flex: 1;
  }

  .version-meta {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .version-uploader {
    font-weight: 500;
    color: var(--color-gray-900);
  }

  .version-date {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .version-size {
    font-size: 0.875rem;
    color: var(--color-gray-600);
    margin-top: 0.25rem;
  }

  .version-notes {
    font-size: 0.875rem;
    color: var(--color-gray-600);
    margin: 0.5rem 0 0 0;
    font-style: italic;
  }

  .version-actions {
    display: flex;
    gap: 0.5rem;
  }

  .activity-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .activity-item {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .activity-icon {
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .activity-icon.success {
    background-color: var(--color-green-100);
    color: var(--color-green-600);
  }

  .activity-icon.info {
    background-color: var(--color-blue-100);
    color: var(--color-blue-600);
  }

  .activity-content {
    flex: 1;
    padding-top: 0.25rem;
  }

  .activity-text {
    font-weight: 500;
    color: var(--color-gray-900);
    margin: 0;
  }

  .activity-meta {
    font-size: 0.875rem;
    color: var(--color-gray-500);
    margin: 0.125rem 0 0 0;
  }

  .share-form {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .share-permissions {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .permission-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-700);
  }

  .permission-options {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .permission-option {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 0.75rem;
    border: 1px solid var(--color-gray-200);
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .permission-option:hover {
    border-color: var(--color-gray-300);
  }

  .permission-option.selected {
    border-color: var(--color-primary-500);
    background-color: var(--color-primary-50);
  }

  .permission-option input {
    margin-top: 0.125rem;
  }

  .permission-content {
    display: flex;
    flex-direction: column;
  }

  .permission-title {
    font-weight: 500;
    color: var(--color-gray-900);
  }

  .permission-desc {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .delete-warning {
    font-size: 0.875rem;
    color: var(--color-gray-600);
    margin-top: 0.5rem;
  }

  @media (max-width: 768px) {
    .page-header {
      flex-direction: column;
    }

    .header-actions {
      width: 100%;
      justify-content: flex-start;
    }

    .quick-info-grid {
      grid-template-columns: 1fr;
    }

    .version-item {
      flex-direction: column;
      gap: 0.75rem;
    }

    .version-actions {
      width: 100%;
      justify-content: flex-start;
    }
  }
</style>
