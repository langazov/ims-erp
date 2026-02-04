<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';
  import Button from '../forms/Button.svelte';

  export let currentPage: number = 1;
  export let totalPages: number = 1;
  export let totalItems: number = 0;
  export let pageSize: number = 10;
  export let pageSizeOptions: number[] = [10, 25, 50, 100];
  export let showPageSizeSelector = true;
  export let showItemCount = true;
  export let siblingCount: number = 1;

  $: startItem = (currentPage - 1) * pageSize + 1;
  $: endItem = Math.min(currentPage * pageSize, totalItems);

  function generatePageNumbers(): (number | string)[] {
    const pages: (number | string)[] = [];
    
    if (totalPages <= 7) {
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      if (currentPage <= 3) {
        pages.push(1, 2, 3, 4, '...', totalPages);
      } else if (currentPage >= totalPages - 2) {
        pages.push(1, '...', totalPages - 3, totalPages - 2, totalPages - 1, totalPages);
      } else {
        pages.push(1, '...', currentPage - 1, currentPage, currentPage + 1, '...', totalPages);
      }
    }
    
    return pages;
  }

  function handlePageChange(page: number) {
    if (page >= 1 && page <= totalPages && page !== currentPage) {
      currentPage = page;
    }
  }

  function handlePageSizeChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    pageSize = parseInt(target.value);
    currentPage = 1;
  }
</script>

<div class={cn('flex items-center justify-between gap-4', $$props.class)}>
  {#if showItemCount}
    <div class="text-sm text-gray-600 dark:text-gray-400">
      Showing <span class="font-medium">{startItem}</span> to <span class="font-medium">{endItem}</span>
      of <span class="font-medium">{totalItems}</span> results
    </div>
  {/if}

  <div class="flex items-center gap-2">
    <Button
      variant="secondary"
      size="sm"
      disabled={currentPage <= 1}
      on:click={() => handlePageChange(currentPage - 1)}
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
      </svg>
    </Button>

    {#each generatePageNumbers() as page}
      {#if page === '...'}
        <span class="px-3 py-2 text-sm text-gray-500">... </span>
      {:else}
        <Button
          variant={currentPage === page ? 'primary' : 'secondary'}
          size="sm"
          on:click={() => handlePageChange(page as number)}
        >
          {page}
        </Button>
      {/if}
    {/each}

    <Button
      variant="secondary"
      size="sm"
      disabled={currentPage >= totalPages}
      on:click={() => handlePageChange(currentPage + 1)}
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
    </Button>
  </div>

  {#if showPageSizeSelector}
    <div class="flex items-center gap-2">
      <span class="text-sm text-gray-600 dark:text-gray-400">Show</span>
      <select
        class="rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm px-2 py-1"
        value={pageSize}
        on:change={handlePageSizeChange}
      >
        {#each pageSizeOptions as size}
          <option value={size}>{size}</option>
        {/each}
      </select>
      <span class="text-sm text-gray-600 dark:text-gray-400">per page</span>
    </div>
  {/if}
</div>
