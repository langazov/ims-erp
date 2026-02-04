<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';
  import { createEventDispatcher } from 'svelte';
  import { X, Filter, RotateCcw } from 'lucide-svelte';

  export let isOpen: boolean = false;
  export let title: string = 'Filters';
  export let showClearButton: boolean = true;
  export let position: 'right' | 'left' = 'right';

  const dispatch = createEventDispatcher<{
    apply: { filters: Record<string, unknown> };
    clear: void;
    close: void;
  }>();

  let filters: Record<string, unknown> = {};

  function handleApply() {
    dispatch('apply', { filters });
  }

  function handleClear() {
    filters = {};
    dispatch('clear');
  }

  function handleClose() {
    dispatch('close');
  }

  function updateFilter(key: string, value: unknown) {
    filters = { ...filters, [key]: value };
  }
</script>

{#if isOpen}
  <div class="fixed inset-0 z-40" on:click={handleClose}>
    <div class="absolute inset-0 bg-black/30 backdrop-blur-sm transition-opacity"></div>
  </div>
{/if}

<aside
  class={cn(
    'fixed top-0 z-50 h-full w-96 bg-white dark:bg-gray-800 shadow-2xl transition-transform duration-300 ease-in-out',
    position === 'right' ? 'right-0' : 'left-0',
    isOpen ? 'translate-x-0' : position === 'right' ? 'translate-x-full' : '-translate-x-full'
  )}
>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
      <div class="flex items-center gap-2">
        <Filter class="w-5 h-5 text-gray-500 dark:text-gray-400" />
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{title}</h2>
      </div>
      <button
        on:click={handleClose}
        class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        aria-label="Close filters"
      >
        <X class="w-5 h-5" />
      </button>
    </div>

    <!-- Filter Content -->
    <div class="flex-1 overflow-y-auto p-6 space-y-6">
      <slot {updateFilter} {filters} />
    </div>

    <!-- Footer -->
    <div class="flex items-center justify-between px-6 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50">
      {#if showClearButton}
        <button
          on:click={handleClear}
          class="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors"
        >
          <RotateCcw class="w-4 h-4" />
          Clear All
        </button>
      {:else}
        <div></div>
      {/if}
      
      <div class="flex gap-3">
        <button
          on:click={handleClose}
          class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
        >
          Cancel
        </button>
        <button
          on:click={handleApply}
          class="px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-lg hover:bg-primary-700 transition-colors"
        >
          Apply Filters
        </button>
      </div>
    </div>
  </div>
</aside>

<!-- Filter Presets Slot -->
{#if $$slots.presets}
  <div class="mb-4">
    <slot name="presets" />
  </div>
{/if}

