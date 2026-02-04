<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';
  import type { TreeNode } from './TreeView.svelte';

  export let node: TreeNode;
  export let selectable = false;
  export let multiSelect = false;
  export let selectedIds: (string | number)[] = [];
  export let path: TreeNode[] = [];
  export let level = 0;

  const dispatch = createEventDispatcher<{
    nodeClick: { node: TreeNode; path: TreeNode[]; event: MouseEvent };
    nodeToggle: { node: TreeNode; path: TreeNode[] };
    selectionChange: { selectedIds: (string | number)[] };
  }>();

  $: hasChildren = node.children && node.children.length > 0;
  $: isSelected = selectedIds.includes(node.id);
  $: isDisabled = node.disabled;

  function handleClick(event: MouseEvent) {
    dispatch('nodeClick', { node, path, event });
  }

  function handleToggle() {
    dispatch('nodeToggle', { node, path });
  }

  function handleCheckboxChange() {
    let newSelectedIds: (string | number)[];
    if (selectedIds.includes(node.id)) {
      newSelectedIds = selectedIds.filter(id => id !== node.id);
    } else {
      newSelectedIds = [...selectedIds, node.id];
    }
    dispatch('selectionChange', { selectedIds: newSelectedIds });
  }
</script>

<div class="select-none">
  <!-- Node Row -->
  <div
    class={cn(
      'flex items-center gap-1 py-1.5 px-2 rounded-md cursor-pointer transition-colors',
      isSelected && 'bg-primary-50 dark:bg-primary-900/20',
      !isSelected && !isDisabled && 'hover:bg-gray-100 dark:hover:bg-gray-800',
      isDisabled && 'opacity-50 cursor-not-allowed'
    )}
    style="padding-left: {level * 16 + 8}px"
    on:click={handleClick}
  >
    <!-- Expand/Collapse Icon -->
    {#if hasChildren}
      <button
        type="button"
        class="w-5 h-5 flex items-center justify-center rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
        on:click|stopPropagation={handleToggle}
      >
        <svg
          class={cn(
            'w-4 h-4 text-gray-500 transition-transform',
            node.expanded && 'rotate-90'
          )}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    {:else}
      <span class="w-5" />
    {/if}

    <!-- Checkbox (if selectable) -->
    {#if selectable}
      <input
        type="checkbox"
        checked={isSelected}
        disabled={isDisabled}
        class="mr-1 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
        on:click|stopPropagation
        on:change={handleCheckboxChange}
      />
    {/if}

    <!-- Node Icon -->
    {#if node.icon}
      <svg class="w-4 h-4 text-gray-500 dark:text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={node.icon} />
      </svg>
    {:else if hasChildren}
      <svg class="w-4 h-4 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
      </svg>
    {:else}
      <svg class="w-4 h-4 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
      </svg>
    {/if}

    <!-- Node Label -->
    <span class={cn(
      'text-sm truncate',
      isSelected ? 'text-primary-700 dark:text-primary-400 font-medium' : 'text-gray-700 dark:text-gray-300',
      isDisabled && 'text-gray-400'
    )}>
      {node.label}
    </span>
  </div>

  <!-- Children -->
  {#if hasChildren && node.expanded}
    <div class="mt-0.5">
      {#each node.children ?? [] as childNode}
        <svelte:self
          node={childNode}
          {selectable}
          {multiSelect}
          {selectedIds}
          path={[...path, node]}
          level={level + 1}
          on:nodeClick
          on:nodeToggle
          on:selectionChange
        />
      {/each}
    </div>
  {/if}
</div>
