<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Select from '$lib/shared/components/forms/Select.svelte';
  import Textarea from '$lib/shared/components/forms/Textarea.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Progress from '$lib/shared/components/display/Progress.svelte';
  import Chip from '$lib/shared/components/display/Chip.svelte';
  import { cn } from '$lib/shared/utils/helpers';

  interface UploadFile {
    file: File;
    id: string;
    name: string;
    size: number;
    type: string;
    progress: number;
    status: 'pending' | 'uploading' | 'completed' | 'error';
    error?: string;
    preview?: string;
  }

  let files: UploadFile[] = [];
  let documentType: string = '';
  let description = '';
  let tags: string[] = [];
  let currentTag = '';
  let dragOver = false;
  let uploading = false;
  let uploadComplete = false;
  let error: string | null = null;

  const MAX_FILE_SIZE = 50 * 1024 * 1024; // 50MB
  const ALLOWED_TYPES = [
    'application/pdf',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'application/vnd.ms-excel',
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    'image/jpeg',
    'image/png',
    'image/gif',
    'text/plain',
    'text/csv'
  ];

  const ALLOWED_EXTENSIONS = ['.pdf', '.doc', '.docx', '.xls', '.xlsx', '.jpg', '.jpeg', '.png', '.gif', '.txt', '.csv'];

  const typeOptions = [
    { value: '', label: 'Select document type' },
    { value: 'invoice', label: 'Invoice' },
    { value: 'po', label: 'Purchase Order' },
    { value: 'receipt', label: 'Receipt' },
    { value: 'contract', label: 'Contract' },
    { value: 'scanned', label: 'Scanned Document' },
    { value: 'other', label: 'Other' }
  ];

  function generateId(): string {
    return Math.random().toString(36).substring(2, 9);
  }

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  function getFileIcon(contentType: string): string {
    if (contentType.startsWith('image/')) return '🖼️';
    if (contentType.includes('pdf')) return '📄';
    if (contentType.includes('word') || contentType.includes('document')) return '📝';
    if (contentType.includes('excel') || contentType.includes('sheet')) return '📊';
    if (contentType.includes('csv') || contentType.includes('text')) return '📃';
    return '📎';
  }

  function isValidFile(file: File): { valid: boolean; error?: string } {
    if (file.size > MAX_FILE_SIZE) {
      return { valid: false, error: `File exceeds maximum size of ${formatFileSize(MAX_FILE_SIZE)}` };
    }

    const extension = '.' + file.name.split('.').pop()?.toLowerCase();
    if (!ALLOWED_EXTENSIONS.includes(extension)) {
      return { valid: false, error: `File type not supported. Allowed: ${ALLOWED_EXTENSIONS.join(', ')}` };
    }

    return { valid: true };
  }

  function createPreview(file: File): Promise<string | undefined> {
    return new Promise((resolve) => {
      if (!file.type.startsWith('image/')) {
        resolve(undefined);
        return;
      }

      const reader = new FileReader();
      reader.onload = (e) => resolve(e.target?.result as string);
      reader.onerror = () => resolve(undefined);
      reader.readAsDataURL(file);
    });
  }

  async function handleFiles(selectedFiles: FileList | null) {
    if (!selectedFiles) return;

    error = null;
    const newFiles: UploadFile[] = [];

    for (const file of Array.from(selectedFiles)) {
      const validation = isValidFile(file);
      
      const uploadFile: UploadFile = {
        file,
        id: generateId(),
        name: file.name,
        size: file.size,
        type: file.type,
        progress: 0,
        status: validation.valid ? 'pending' : 'error',
        error: validation.error,
        preview: file.type.startsWith('image/') ? await createPreview(file) : undefined
      };

      newFiles.push(uploadFile);
    }

    files = [...files, ...newFiles];
  }

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
    dragOver = true;
  }

  function handleDragLeave(event: DragEvent) {
    event.preventDefault();
    dragOver = false;
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    dragOver = false;
    handleFiles(event.dataTransfer?.files ?? null);
  }

  function handleFileInput(event: Event) {
    const target = event.target as HTMLInputElement;
    handleFiles(target.files);
    target.value = ''; // Reset input
  }

  function removeFile(id: string) {
    files = files.filter(f => f.id !== id);
  }

  function addTag() {
    const trimmed = currentTag.trim().toLowerCase();
    if (trimmed && !tags.includes(trimmed)) {
      tags = [...tags, trimmed];
    }
    currentTag = '';
  }

  function removeTag(tag: string) {
    tags = tags.filter(t => t !== tag);
  }

  function handleTagKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      addTag();
    } else if (event.key === 'Backspace' && !currentTag && tags.length > 0) {
      tags = tags.slice(0, -1);
    }
  }

  async function simulateUpload(file: UploadFile): Promise<void> {
    return new Promise((resolve, reject) => {
      const totalSteps = 20;
      let currentStep = 0;

      const interval = setInterval(() => {
        currentStep++;
        const progress = Math.min((currentStep / totalSteps) * 100, 100);
        
        files = files.map(f => 
          f.id === file.id ? { ...f, progress, status: 'uploading' } : f
        );

        if (currentStep >= totalSteps) {
          clearInterval(interval);
          
          // Simulate occasional errors (5% chance)
          if (Math.random() < 0.05) {
            files = files.map(f => 
              f.id === file.id ? { ...f, status: 'error', error: 'Upload failed. Please try again.' } : f
            );
            reject(new Error('Upload failed'));
          } else {
            files = files.map(f => 
              f.id === file.id ? { ...f, progress: 100, status: 'completed' } : f
            );
            resolve();
          }
        }
      }, 100);
    });
  }

  async function handleSubmit() {
    if (files.length === 0) {
      error = 'Please select at least one file to upload';
      return;
    }

    if (!documentType) {
      error = 'Please select a document type';
      return;
    }

    const pendingFiles = files.filter(f => f.status === 'pending');
    if (pendingFiles.length === 0) {
      error = 'No valid files to upload';
      return;
    }

    uploading = true;
    error = null;

    try {
      for (const file of pendingFiles) {
        await simulateUpload(file);
      }

      const allSuccessful = files.every(f => f.status === 'completed');
      if (allSuccessful) {
        uploadComplete = true;
        setTimeout(() => {
          goto('/documents');
        }, 1500);
      }
    } catch (err) {
      error = 'Some files failed to upload. Please check the status below.';
    } finally {
      uploading = false;
    }
  }

  function handleCancel() {
    goto('/documents');
  }

  function getStatusIcon(status: string): string {
    switch (status) {
      case 'completed': return '✓';
      case 'error': return '✕';
      case 'uploading': return '↻';
      default: return '○';
    }
  }

  onMount(() => {
    // Component mounted
  });
</script>

<svelte:head>
  <title>Upload Documents | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Upload Documents</h1>
      <p class="page-description">Upload and organize your documents</p>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  {#if uploadComplete}
    <Alert variant="success" class="mb-4">
      All files uploaded successfully! Redirecting to documents...
    </Alert>
  {/if}

  <div class="upload-grid">
    <Card class="upload-card">
      <h2 class="section-title">Select Files</h2>
      
      <!-- Drag and Drop Zone -->
      <div
        class={cn(
          'drop-zone',
          dragOver && 'drag-over',
          uploading && 'disabled'
        )}
        on:dragover={handleDragOver}
        on:dragleave={handleDragLeave}
        on:drop={handleDrop}
        on:click={() => !uploading && document.getElementById('file-input')?.click()}
        role="button"
        tabindex="0"
      >
        <input
          id="file-input"
          type="file"
          multiple
          accept={ALLOWED_EXTENSIONS.join(',')}
          class="hidden"
          on:change={handleFileInput}
          disabled={uploading}
        />
        
        <svg class="upload-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
        </svg>
        
        <p class="upload-text">
          <span class="font-medium">Click to upload</span> or drag and drop
        </p>
        <p class="upload-hint">
          PDF, Word, Excel, Images up to {formatFileSize(MAX_FILE_SIZE)}
        </p>
      </div>

      <!-- File List -->
      {#if files.length > 0}
        <div class="file-list">
          <h3 class="file-list-title">Selected Files ({files.length})</h3>
          
          {#each files as file (file.id)}
            <div class="file-item" class:has-error={file.status === 'error'}>
              <div class="file-preview">
                {#if file.preview}
                  <img src={file.preview} alt={file.name} class="preview-image" />
                {:else}
                  <span class="file-icon">{getFileIcon(file.type)}</span>
                {/if}
              </div>
              
              <div class="file-info">
                <div class="file-name" title={file.name}>{file.name}</div>
                <div class="file-meta">
                  <span>{formatFileSize(file.size)}</span>
                  {#if file.status === 'error'}
                    <span class="error-text">{file.error}</span>
                  {/if}
                </div>
                
                {#if file.status === 'uploading' || file.status === 'completed'}
                  <div class="file-progress">
                    <Progress 
                      value={file.progress} 
                      max={100} 
                      size="sm"
                      variant={file.status === 'completed' ? 'success' : 'default'}
                      showLabel
                    />
                  </div>
                {/if}
              </div>
              
              <div class="file-status">
                {#if file.status === 'completed'}
                  <span class="status-icon success">{getStatusIcon(file.status)}</span>
                {:else if file.status === 'error'}
                  <span class="status-icon error">{getStatusIcon(file.status)}</span>
                {:else if file.status === 'uploading'}
                  <Spinner size="sm" />
                {/if}
                
                {#if !uploading}
                  <button
                    type="button"
                    class="remove-btn"
                    on:click={() => removeFile(file.id)}
                    title="Remove file"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </Card>

    <Card class="metadata-card">
      <h2 class="section-title">Document Details</h2>
      
      <form on:submit|preventDefault={handleSubmit} class="metadata-form">
        <div class="form-group">
          <Select
            id="document-type"
            label="Document Type"
            options={typeOptions}
            bind:value={documentType}
            required
            disabled={uploading}
          />
        </div>

        <div class="form-group">
          <Textarea
            id="description"
            label="Description"
            bind:value={description}
            placeholder="Add a description for these documents..."
            rows={4}
            maxLength={500}
            disabled={uploading}
          />
        </div>

        <div class="form-group">
          <label class="tags-label" for="tags-input">
            Tags
            <span class="help-text">Press Enter to add tags</span>
          </label>
          
          <div class="tags-input-container">
            {#each tags as tag}
              <Chip variant="primary" size="sm" removable on:click={() => removeTag(tag)}>
                {tag}
              </Chip>
            {/each}
            
            <input
              id="tags-input"
              type="text"
              class="tags-input"
              placeholder={tags.length === 0 ? 'Add tags...' : ''}
              bind:value={currentTag}
              on:keydown={handleTagKeydown}
              on:blur={addTag}
              disabled={uploading}
            />
          </div>
        </div>

        <div class="form-actions">
          <Button variant="secondary" on:click={handleCancel} disabled={uploading}>
            Cancel
          </Button>
          <Button 
            variant="primary" 
            type="submit" 
            loading={uploading}
            disabled={files.length === 0 || !documentType}
          >
            {uploading ? 'Uploading...' : `Upload ${files.length > 0 ? `(${files.filter(f => f.status === 'pending').length})` : ''}`}
          </Button>
        </div>
      </form>
    </Card>
  </div>
</div>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .page-header {
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

  .upload-grid {
    display: grid;
    grid-template-columns: 1.5fr 1fr;
    gap: 1.5rem;
  }

  :global(.upload-card),
  :global(.metadata-card) {
    padding: 1.5rem;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-gray-900);
    margin: 0 0 1.25rem 0;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .drop-zone {
    border: 2px dashed var(--color-gray-300);
    border-radius: 0.75rem;
    padding: 2.5rem;
    text-align: center;
    cursor: pointer;
    transition: all 0.2s ease;
    background-color: var(--color-gray-50);
  }

  .drop-zone:hover:not(.disabled) {
    border-color: var(--color-primary-400);
    background-color: var(--color-primary-50);
  }

  .drop-zone.drag-over {
    border-color: var(--color-primary-500);
    background-color: var(--color-primary-100);
  }

  .drop-zone.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .upload-icon {
    width: 3rem;
    height: 3rem;
    color: var(--color-gray-400);
    margin: 0 auto 1rem;
  }

  .upload-text {
    color: var(--color-gray-600);
    margin-bottom: 0.5rem;
  }

  .upload-hint {
    font-size: 0.875rem;
    color: var(--color-gray-400);
  }

  .file-list {
    margin-top: 1.5rem;
  }

  .file-list-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-gray-700);
    margin-bottom: 0.75rem;
  }

  .file-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    background-color: var(--color-gray-50);
    border-radius: 0.5rem;
    margin-bottom: 0.5rem;
    border: 1px solid transparent;
  }

  .file-item.has-error {
    border-color: var(--color-red-300);
    background-color: var(--color-red-50);
  }

  .file-preview {
    width: 3rem;
    height: 3rem;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: white;
    border-radius: 0.375rem;
    overflow: hidden;
  }

  .preview-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .file-icon {
    font-size: 1.25rem;
  }

  .file-info {
    flex: 1;
    min-width: 0;
  }

  .file-name {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-900);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .file-meta {
    font-size: 0.75rem;
    color: var(--color-gray-500);
    display: flex;
    gap: 0.5rem;
    margin-top: 0.125rem;
  }

  .error-text {
    color: var(--color-red-600);
  }

  .file-progress {
    margin-top: 0.5rem;
  }

  .file-status {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .status-icon {
    width: 1.5rem;
    height: 1.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .status-icon.success {
    background-color: var(--color-green-100);
    color: var(--color-green-600);
  }

  .status-icon.error {
    background-color: var(--color-red-100);
    color: var(--color-red-600);
  }

  .remove-btn {
    padding: 0.25rem;
    color: var(--color-gray-400);
    background: none;
    border: none;
    cursor: pointer;
    border-radius: 0.25rem;
    transition: all 0.15s ease;
  }

  .remove-btn:hover {
    color: var(--color-red-500);
    background-color: var(--color-red-50);
  }

  .metadata-form {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
  }

  .tags-label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-700);
    margin-bottom: 0.5rem;
  }

  .help-text {
    font-weight: 400;
    color: var(--color-gray-400);
    margin-left: 0.5rem;
    font-size: 0.75rem;
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
    min-width: 80px;
    border: none;
    background: none;
    font-size: 0.875rem;
    color: var(--color-gray-900);
    outline: none;
  }

  .tags-input::placeholder {
    color: var(--color-gray-400);
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    padding-top: 1rem;
    border-top: 1px solid var(--color-gray-200);
    margin-top: auto;
  }

  @media (max-width: 968px) {
    .upload-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
