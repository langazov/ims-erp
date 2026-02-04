<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  type ListItem = {
    id: string | number;
    title: string;
    subtitle?: string;
    description?: string;
    icon?: string;
    avatar?: string;
    badge?: { text: string; variant?: 'gray' | 'green' | 'yellow' | 'red' | 'blue' };
    meta?: string;
    disabled?: boolean;
    [key: string]: unknown;
  };

  export let items: ListItem[] = [];
  export let loading = false;
  export let emptyMessage = 'No items found';
  export let selectable = false;
  export let selectedKeys: (string | number)[] = [];
  export let bordered = true;
  export let divided = true;
  export let hoverable = true;
  export let compact = false;
  export let draggable = false;

  const dispatch = createEventDispatcher<{
    itemClick: { item: ListItem; index: number };
    selectionChange: { selectedKeys: (string | number)[] };
    reorder: { items: ListItem[] };
  }>();

  let dragStartIndex: number | null = null;

  $: allSelected = selectedKeys.length > 0 && selectedKeys.length === items.length;
  $: someSelected = selectedKeys.length > 0 && selectedKeys.length < items.length;

  function handleItemClick(item: ListItem, index: number) {
    if (item.disabled) return;
    dispatch('itemClick', { item, index });
  }

  function toggleSelectAll() {
    if (allSelected) {
      selectedKeys = [];
    } else {
      selectedKeys = items.filter(item => !item.disabled).map(item => item.id);
    }
    dispatch('selectionChange', { selectedKeys });
  }

  function toggleSelectItem(key: string | number) {
    if (selectedKeys.includes(key)) {
      selectedKeys = selectedKeys.filter(k => k !== key);
    } else {
      selectedKeys = [...selectedKeys, key];
    }
    dispatch('selectionChange', { selectedKeys });
  }

  function handleDragStart(index: number) {
    dragStartIndex = index;
  }

  function handleDragOver(event: DragEvent, index: number) {
    event.preventDefault();
    if (dragStartIndex === null || dragStartIndex === index) return;

    const newItems = [...items];
    const [draggedItem] = newItems.splice(dragStartIndex, 1);
    newItems.splice(index, 0, draggedItem);
    
    dragStartIndex = index;
    items = newItems;
  }

  function handleDragEnd() {
    dragStartIndex = null;
    dispatch('reorder', { items });
  }

  function getBadgeClasses(variant: string = 'gray') {
    const classes = {
      gray: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200',
      green: 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-400',
      yellow: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-400',
      red: 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-400',
      blue: 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-400',
    };
    return classes[variant as keyof typeof classes] ?? classes.gray;
  }
</script>

<div class={cn(
  'bg-white dark:bg-gray-800',
  bordered && 'rounded-lg border border-gray-200 dark:border-gray-700',
  $$props.class
)}>
  <!-- Header with select all -->
  {#if selectable && items.length > 0}
    <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center gap-3">
      <input
        type="checkbox"
        checked={allSelected}
        indeterminate={someSelected}
        on:change={toggleSelectAll}
        class="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
      />
      <span class="text-sm text-gray-500 dark:text-gray-400">
        {selectedKeys.length} selected
      </span>
    </div>
  {/if}

  <!-- List Items -->
  {#if loading}
    <div class="px-4 py-12 text-center">
      <svg class="animate-spin h-8 w-8 text-primary-600 mx-auto" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
      </svg>
    </div>
  {:else if items.length === 0}
    <div class="px-4 py-12 text-center text-gray-500 dark:text-gray-400">
      {emptyMessage}
    </div>
  {:else}
    <ul class={cn(
      divided && 'divide-y divide-gray-200 dark:divide-gray-700'
    )}>
      {#each items as item, index (item.id)}
        <li
          class={cn(
            'flex items-center gap-4',
            compact ? 'px-4 py-2' : 'px-4 py-4',
            hoverable && !item.disabled && 'hover:bg-gray-50 dark:hover:bg-gray-800/50',
            item.disabled && 'opacity-50 cursor-not-allowed',
            !item.disabled && 'cursor-pointer',
            draggable && 'cursor-move'
          )}
          draggable={draggable}
          on:dragstart={() => handleDragStart(index)}
          on:dragover={(e) => handleDragOver(e, index)}
          on:dragend={handleDragEnd}
          on:click={() => handleItemClick(item, index)}
        >
          <!-- Drag Handle -->
          {#if draggable}
            <div class="text-gray-400 cursor-move">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8h16M4 16h16" />
              </svg>
            </div>
          {/if}

          <!-- Checkbox -->
          {#if selectable}
            <div on:click|stopPropagation>
              <input
                type="checkbox"
                checked={selectedKeys.includes(item.id)}
                disabled={item.disabled}
                on:change={() => toggleSelectItem(item.id)}
                class="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              />
            </div>
          {/if}

          <!-- Icon or Avatar -->
          {#if item.avatar}
            <img
              src={item.avatar}
              alt=""
              class={cn(
                'rounded-full object-cover',
                compact ? 'w-8 h-8' : 'w-10 h-10'
              )}
            />
          {:else if item.icon}
            <div class={cn(
              'flex-shrink-0 rounded-lg bg-gray-100 dark:bg-gray-700 flex items-center justify-center',
              compact ? 'w-8 h-8' : 'w-10 h-10'
            )}>
              <svg class={cn('text-gray-500 dark:text-gray-400', compact ? 'w-4 h-4' : 'w-5 h-5')} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon} />
              </svg>
            </div>
          {/if}

          <!-- Content -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <p class={cn(
                'font-medium text-gray-900 dark:text-white truncate',
                compact ? 'text-sm' : 'text-base'
              )}>
                {item.title}
              </p>
              {#if item.badge}
                <span class={cn(
                  'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                  getBadgeClasses(item.badge.variant)
                )}>
                  {item.badge.text}
                </span>
              {/if}
            </div>
            {#if item.subtitle}
              <p class={cn(
                'text-gray-500 dark:text-gray-400 truncate',
                compact ? 'text-xs' : 'text-sm'
              )}>
                {item.subtitle}
              </p>
            {/if}
            {#if item.description}
              <p class={cn(
                'text-gray-500 dark:text-gray-400 mt-1',
                compact ? 'text-xs' : 'text-sm'
              )}>
                {item.description}
              </p>
            {/if}
          </div>

          <!-- Meta -->
          {#if item.meta}
            <div class="text-sm text-gray-500 dark:text-gray-400">
              {item.meta}
            </div>
          {/if}

          <!-- Actions Slot -->
          {#if $$slots.actions}
            <div on:click|stopPropagation>
              <slot name="actions" {item} {index} />
            </div>
          {/if}

          <!-- Chevron -->
          {#if !$$slots.actions}
            <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>
