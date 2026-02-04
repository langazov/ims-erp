<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';
  import Pagination from './Pagination.svelte';

  type Column<T = Record<string, unknown>> = {
    key: keyof T | string;
    label: string;
    sortable?: boolean;
    align?: 'left' | 'center' | 'right';
    width?: string;
    formatter?: (value: unknown, row: T) => string;
  };

  type Filter = {
    key: string;
    label: string;
    type: 'text' | 'select' | 'date' | 'number' | 'boolean';
    options?: { label: string; value: unknown }[];
  };

  export let columns: Column[] = [];
  export let data: Record<string, unknown>[] = [];
  export let rowKey = 'id';
  export let loading = false;
  export let emptyMessage = 'No data available';
  export let stickyHeader = false;
  export let selectable = false;
  export let selectedKeys: (string | number)[] = [];
  
  // Pagination
  export let pagination = false;
  export let page = 1;
  export let pageSize = 10;
  export let total = 0;
  
  // Sorting
  export let sortKey = '';
  export let sortDirection: 'asc' | 'desc' = 'asc';
  
  // Filtering
  export let filters: Filter[] = [];
  export let filterValues: Record<string, unknown> = {};
  export let showFilters = false;
  
  // Search
  export let searchable = false;
  export let searchQuery = '';
  export let searchPlaceholder = 'Search...';

  const dispatch = createEventDispatcher<{
    sort: { key: string; direction: 'asc' | 'desc' };
    pageChange: { page: number };
    pageSizeChange: { pageSize: number };
    filterChange: { filters: Record<string, unknown> };
    search: { query: string };
    rowClick: { row: Record<string, unknown>; index: number };
    selectionChange: { selectedKeys: (string | number)[]; selectedRows: Record<string, unknown>[] };
  }>();

  $: totalPages = Math.ceil(total / pageSize);
  
  $: paginatedData = pagination 
    ? data.slice((page - 1) * pageSize, page * pageSize)
    : data;

  $: allSelected = selectedKeys.length > 0 && selectedKeys.length === paginatedData.length;
  $: someSelected = selectedKeys.length > 0 && selectedKeys.length < paginatedData.length;

  function handleSort(column: Column) {
    if (!column.sortable) return;
    
    const newDirection = sortKey === column.key && sortDirection === 'asc' ? 'desc' : 'asc';
    sortKey = column.key as string;
    sortDirection = newDirection;
    
    dispatch('sort', { key: sortKey, direction: sortDirection });
  }

  function handlePageChange(newPage: number) {
    page = newPage;
    dispatch('pageChange', { page });
  }

  function handlePageSizeChange(newPageSize: number) {
    pageSize = newPageSize;
    page = 1;
    dispatch('pageSizeChange', { pageSize });
  }

  function handleFilterChange(key: string, value: unknown) {
    filterValues = { ...filterValues, [key]: value };
    dispatch('filterChange', { filters: filterValues });
  }

  function handleSearch() {
    dispatch('search', { query: searchQuery });
  }

  function toggleSelectAll() {
    if (allSelected) {
      selectedKeys = [];
    } else {
      selectedKeys = paginatedData.map(row => row[rowKey] as string | number);
    }
    dispatchSelectionChange();
  }

  function toggleSelectRow(key: string | number) {
    if (selectedKeys.includes(key)) {
      selectedKeys = selectedKeys.filter(k => k !== key);
    } else {
      selectedKeys = [...selectedKeys, key];
    }
    dispatchSelectionChange();
  }

  function dispatchSelectionChange() {
    const selectedRows = data.filter(row => selectedKeys.includes(row[rowKey] as string | number));
    dispatch('selectionChange', { selectedKeys, selectedRows });
  }

  function handleRowClick(row: Record<string, unknown>, index: number) {
    dispatch('rowClick', { row, index });
  }

  function getAlignment(column: Column) {
    return column.align ?? 'left';
  }

  function formatValue(column: Column, value: unknown, row: Record<string, unknown>) {
    if (column.formatter) {
      return column.formatter(value, row);
    }
    if (value === null || value === undefined) return '';
    return String(value);
  }
</script>

<div class="space-y-4">
  <!-- Toolbar -->
  {#if searchable || filters.length > 0 || $$slots.toolbar}
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div class="flex items-center gap-2 flex-wrap">
        {#if searchable}
          <div class="relative">
            <input
              type="text"
              bind:value={searchQuery}
              placeholder={searchPlaceholder}
              class="pl-10 pr-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm focus:ring-2 focus:ring-primary-500 focus:border-transparent w-64"
              on:input={() => handleSearch()}
            />
            <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
        {/if}
        
        {#if filters.length > 0}
          <button
            type="button"
            class="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700"
            on:click={() => showFilters = !showFilters}
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
            </svg>
            Filters
            {#if Object.keys(filterValues).length > 0}
              <span class="bg-primary-600 text-white text-xs rounded-full px-2 py-0.5">
                {Object.keys(filterValues).length}
              </span>
            {/if}
          </button>
        {/if}
      </div>
      
      <slot name="toolbar" />
    </div>
  {/if}

  <!-- Filters Panel -->
  {#if showFilters && filters.length > 0}
    <div class="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4 space-y-4">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium text-gray-900 dark:text-white">Filters</h3>
        <button
          type="button"
          class="text-sm text-primary-600 hover:text-primary-700"
          on:click={() => { filterValues = {}; dispatch('filterChange', { filters: {} }); }}
        >
          Clear all
        </button>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {#each filters as filter}
          <div>
            <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{filter.label}</label>
            {#if filter.type === 'select'}
              <select
                value={filterValues[filter.key] ?? ''}
                on:change={(e) => handleFilterChange(filter.key, e.currentTarget.value)}
                class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm px-3 py-2"
              >
                <option value="">All</option>
                {#each filter.options ?? [] as option}
                  <option value={option.value}>{option.label}</option>
                {/each}
              </select>
            {:else if filter.type === 'text'}
              <input
                type="text"
                value={filterValues[filter.key] ?? ''}
                on:input={(e) => handleFilterChange(filter.key, e.currentTarget.value)}
                class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm px-3 py-2"
              />
            {:else if filter.type === 'date'}
              <input
                type="date"
                value={filterValues[filter.key] ?? ''}
                on:input={(e) => handleFilterChange(filter.key, e.currentTarget.value)}
                class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm px-3 py-2"
              />
            {:else if filter.type === 'number'}
              <input
                type="number"
                value={filterValues[filter.key] ?? ''}
                on:input={(e) => handleFilterChange(filter.key, e.currentTarget.value)}
                class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm px-3 py-2"
              />
            {:else if filter.type === 'boolean'}
              <select
                value={filterValues[filter.key] ?? ''}
                on:change={(e) => handleFilterChange(filter.key, e.currentTarget.value === 'true')}
                class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm px-3 py-2"
              >
                <option value="">All</option>
                <option value="true">Yes</option>
                <option value="false">No</option>
              </select>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Table -->
  <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
    <table class="w-full">
      <thead class={cn('bg-gray-50 dark:bg-gray-800/50', stickyHeader && 'sticky top-0 z-10')}>
        <tr>
          {#if selectable}
            <th class="px-4 py-3 w-12">
              <input
                type="checkbox"
                checked={allSelected}
                indeterminate={someSelected}
                on:change={toggleSelectAll}
                class="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              />
            </th>
          {/if}
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
              <div class="flex items-center gap-2" class:justify-start={getAlignment(column) === 'left'} class:justify-center={getAlignment(column) === 'center'} class:justify-end={getAlignment(column) === 'right'}>
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
            <td colspan={columns.length + (selectable ? 1 : 0)} class="px-6 py-12 text-center">
              <div class="flex justify-center">
                <svg class="animate-spin h-8 w-8 text-primary-600" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
              </div>
            </td>
          </tr>
        {:else if paginatedData.length === 0}
          <tr>
            <td colspan={columns.length + (selectable ? 1 : 0)} class="px-6 py-12 text-center text-gray-500 dark:text-gray-400">
              {emptyMessage}
            </td>
          </tr>
        {:else}
          {#each paginatedData as row, index (row[rowKey] ?? index)}
            <tr 
              class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              on:click={() => handleRowClick(row, index)}
            >
              {#if selectable}
                <td class="px-4 py-4" on:click|stopPropagation>
                  <input
                    type="checkbox"
                    checked={selectedKeys.includes(row[rowKey] as string | number)}
                    on:change={() => toggleSelectRow(row[rowKey] as string | number)}
                    class="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                  />
                </td>
              {/if}
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
                    {formatValue(column, row[column.key], row)}
                  </slot>
                </td>
              {/each}
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  {#if pagination}
    <div class="flex items-center justify-between">
      <div class="text-sm text-gray-500 dark:text-gray-400">
        Showing {((page - 1) * pageSize) + 1} to {Math.min(page * pageSize, total)} of {total} results
      </div>
      <Pagination
        {page}
        {totalPages}
        on:change={(e) => handlePageChange(e.detail.page)}
      />
    </div>
  {/if}
</div>
