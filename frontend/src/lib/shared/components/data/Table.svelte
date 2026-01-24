<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';

  export let columns: { key: string; label: string; sortable?: boolean; align?: 'left' | 'center' | 'right'; width?: string }[] = [];
  export let data: Record<string, unknown>[] = [];
  export let sortKey = '';
  export let sortDirection: 'asc' | 'desc' = 'asc';
  export let rowKey = 'id';
  export let loading = false;
  export let emptyMessage = 'No data available';
  export let stickyHeader = false;

  function handleSort(column: typeof columns[0]) {
    if (!column.sortable) return;
    if (sortKey === column.key) {
      sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
      sortKey = column.key;
      sortDirection = 'asc';
    }
  }

  function getAlignment(column: typeof columns[0]) {
    return column.align ?? 'left';
  }
</script>

<div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
  <table class="w-full">
    <thead class={cn('bg-gray-50 dark:bg-gray-800/50', stickyHeader && 'sticky top-0 z-10')}>
      <tr>
        {#each columns as column}
          <th
            class={cn(
              'px-6 py-3 text-xs font-medium uppercase tracking-wider',
              'text-gray-500 dark:text-gray-400',
              column.sortable && 'cursor-pointer select-none hover:bg-gray-100 dark:hover:bg-gray-800',
              getAlignment(column) === 'left' && 'text-left',
              getAlignment(column) === 'center' && 'text-center',
              getAlignment(column) === 'right' && 'text-right'
            )}
            style={column.width ? `width: ${column.width}` : ''}
            on:click={() => handleSort(column)}
          >
            <div class="flex items-center gap-2">
              <span>{column.label}</span>
              {#if column.sortable}
                <span class="text-gray-400">
                  {#if sortKey === column.key}
                    {#if sortDirection === 'asc'}
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
                      </svg>
                    {:else}
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                      </svg>
                    {/if}
                  {:else}
                    <svg class="w-4 h-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
                    </svg>
                  {/if}
                </span>
              {/if}
            </div>
          </th>
        {/each}
      </tr>
    </thead>
    <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
      {#if loading}
        <tr>
          <td colspan={columns.length} class="px-6 py-12 text-center">
            <div class="flex justify-center">
              <slot name="loading">
                <svg class="animate-spin h-8 w-8 text-primary-600" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
              </slot>
            </div>
          </td>
        </tr>
      {:else if data.length === 0}
        <tr>
          <td colspan={columns.length} class="px-6 py-12 text-center text-gray-500 dark:text-gray-400">
            {emptyMessage}
          </td>
        </tr>
      {:else}
        {#each data as row, index (row[rowKey] ?? index)}
          <tr class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
            {#each columns as column}
              <td
                class={cn(
                  'px-6 py-4 whitespace-nowrap text-sm',
                  'text-gray-900 dark:text-gray-100',
                  getAlignment(column) === 'left' && 'text-left',
                  getAlignment(column) === 'center' && 'text-center',
                  getAlignment(column) === 'right' && 'text-right'
                )}
              >
                <slot name="cell" {row} {column} value={row[column.key]}>
                  {row[column.key] ?? ''}
                </slot>
              </td>
            {/each}
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>
</div>
