<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  export let id: string;
  export let label: string | undefined = undefined;
  export let accept: string | undefined = undefined;
  export let multiple = false;
  export let disabled = false;
  export let required = false;
  export let error: string | undefined = undefined;
  export let helper: string | undefined = undefined;
  export let maxSize: number | undefined = undefined; // in bytes

  let files: FileList | null = null;
  let dragOver = false;
  let fileInput: HTMLInputElement;

  const dispatch = createEventDispatcher();

  function handleFiles(selectedFiles: FileList | null) {
    if (!selectedFiles) return;

    const validFiles: File[] = [];
    const errors: string[] = [];

    Array.from(selectedFiles).forEach(file => {
      if (maxSize && file.size > maxSize) {
        errors.push(`${file.name} exceeds maximum size`);
      } else {
        validFiles.push(file);
      }
    });

    if (errors.length > 0) {
      dispatch('error', errors);
    }

    if (validFiles.length > 0) {
      dispatch('select', multiple ? validFiles : validFiles[0]);
    }

    files = selectedFiles;
  }

  function handleChange(event: Event) {
    const target = event.target as HTMLInputElement;
    handleFiles(target.files);
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

  function handleClick() {
    fileInput?.click();
  }

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
</script>

<div class={cn('w-full', $$props.class)}>
  {#if label}
    <label
      for={id}
      class={cn(
        'block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5',
        disabled && 'opacity-50'
      )}
    >
      {label}
      {#if required}
        <span class="text-red-500">*</span>
      {/if}
    </label>
  {/if}

  <div
    class={cn(
      'relative border-2 border-dashed rounded-lg p-6 text-center cursor-pointer',
      'transition-colors duration-200',
      dragOver
        ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
        : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500',
      error && 'border-red-500 bg-red-50 dark:bg-red-900/10',
      disabled && 'cursor-not-allowed opacity-50'
    )}
    on:dragover={handleDragOver}
    on:dragleave={handleDragLeave}
    on:drop={handleDrop}
    on:click={handleClick}
    on:keydown={(e) => e.key === 'Enter' && handleClick()}
    role="button"
    tabindex="0"
  >
    <input
      bind:this={fileInput}
      {id}
      type="file"
      {accept}
      {multiple}
      {disabled}
      {required}
      class="sr-only"
      on:change={handleChange}
    />

    <svg
      class="mx-auto h-12 w-12 text-gray-400"
      stroke="currentColor"
      fill="none"
      viewBox="0 0 48 48"
      aria-hidden="true"
    >
      <path
        d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>

    <div class="mt-4">
      <p class="text-sm text-gray-600 dark:text-gray-400">
        <span class="font-medium text-primary-600 hover:text-primary-500">Upload a file</span>
        or drag and drop
      </p>
      {#if accept}
        <p class="text-xs text-gray-500 dark:text-gray-500 mt-1">{accept}</p>
      {/if}
      {#if maxSize}
        <p class="text-xs text-gray-500 dark:text-gray-500">Max size: {formatFileSize(maxSize)}</p>
      {/if}
    </div>
  </div>

  {#if files && files.length > 0}
    <div class="mt-4 space-y-2">
      {#each Array.from(files) as file}
        <div class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <svg class="w-5 h-5 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{file.name}</p>
            <p class="text-xs text-gray-500">{formatFileSize(file.size)}</p>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  {#if helper}
    <p class="text-xs text-gray-500 dark:text-gray-400 mt-1.5">{helper}</p>
  {/if}

  {#if error}
    <p class="text-xs text-red-600 dark:text-red-400 mt-1.5">{error}</p>
  {/if}
</div>
