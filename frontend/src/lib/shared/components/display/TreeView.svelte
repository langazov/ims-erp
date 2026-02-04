<script lang="ts" context="module">
  export type TreeNode = {
    id: string | number;
    label: string;
    icon?: string;
    expanded?: boolean;
    selected?: boolean;
    disabled?: boolean;
    children?: TreeNode[];
    [key: string]: unknown;
  };
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';
  import TreeNodeItem from './TreeNodeItem.svelte';

  export let nodes: TreeNode[] = [];
  export let selectable = false;
  export let multiSelect = false;
  export let expandOnClick = true;
  export let autoExpand = false;
  export let loading = false;
  export let emptyMessage = 'No items';

  const dispatch = createEventDispatcher<{
    nodeClick: { node: TreeNode; path: TreeNode[] };
    nodeToggle: { node: TreeNode; expanded: boolean; path: TreeNode[] };
    selectionChange: { selectedNodes: TreeNode[]; selectedIds: (string | number)[] };
  }>();

  let selectedIds: (string | number)[] = [];

  function toggleNode(node: TreeNode, path: TreeNode[]) {
    if (node.disabled) return;
    
    node.expanded = !node.expanded;
    dispatch('nodeToggle', { node, expanded: node.expanded ?? false, path });
    nodes = [...nodes];
  }

  function handleNodeClick(node: TreeNode, path: TreeNode[], event: MouseEvent) {
    if (node.disabled) return;

    if (selectable) {
      if (multiSelect) {
        if (event.ctrlKey || event.metaKey) {
          if (selectedIds.includes(node.id)) {
            selectedIds = selectedIds.filter(id => id !== node.id);
          } else {
            selectedIds = [...selectedIds, node.id];
          }
        } else {
          selectedIds = [node.id];
        }
      } else {
        selectedIds = [node.id];
      }
      
      const selectedNodes = getSelectedNodes(nodes);
      dispatch('selectionChange', { selectedNodes, selectedIds });
    }

    if (expandOnClick && node.children && node.children.length > 0) {
      toggleNode(node, path);
    }

    dispatch('nodeClick', { node, path });
  }

  function getSelectedNodes(nodeList: TreeNode[]): TreeNode[] {
    const result: TreeNode[] = [];
    for (const node of nodeList) {
      if (selectedIds.includes(node.id)) {
        result.push(node);
      }
      if (node.children) {
        result.push(...getSelectedNodes(node.children));
      }
    }
    return result;
  }

  function expandAll(nodeList: TreeNode[] = nodes) {
    for (const node of nodeList) {
      node.expanded = true;
      if (node.children) {
        expandAll(node.children);
      }
    }
    nodes = [...nodes];
  }

  function collapseAll(nodeList: TreeNode[] = nodes) {
    for (const node of nodeList) {
      node.expanded = false;
      if (node.children) {
        collapseAll(node.children);
      }
    }
    nodes = [...nodes];
  }

  export function getExpandedNodes(): TreeNode[] {
    const result: TreeNode[] = [];
    function findExpanded(nodeList: TreeNode[]) {
      for (const node of nodeList) {
        if (node.expanded) {
          result.push(node);
        }
        if (node.children) {
          findExpanded(node.children);
        }
      }
    }
    findExpanded(nodes);
    return result;
  }

  export function expandToNode(targetId: string | number) {
    function findAndExpand(nodeList: TreeNode[], path: TreeNode[]): boolean {
      for (const node of nodeList) {
        if (node.id === targetId) {
          for (const parent of path) {
            parent.expanded = true;
          }
          return true;
        }
        if (node.children) {
          if (findAndExpand(node.children, [...path, node])) {
            node.expanded = true;
            return true;
          }
        }
      }
      return false;
    }
    findAndExpand(nodes, []);
    nodes = [...nodes];
  }
</script>

<div class={cn('space-y-1', $$props.class)}>
  {#if loading}
    <div class="flex justify-center py-4">
      <svg class="animate-spin h-6 w-6 text-primary-600" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
      </svg>
    </div>
  {:else if nodes.length === 0}
    <div class="text-center py-4 text-gray-500 dark:text-gray-400 text-sm">
      {emptyMessage}
    </div>
  {:else}
    {#each nodes as node}
      <TreeNodeItem
        {node}
        {selectable}
        {multiSelect}
        {selectedIds}
        path={[]}
        level={0}
        on:nodeClick={(e) => handleNodeClick(e.detail.node, e.detail.path, e.detail.event)}
        on:nodeToggle={(e) => toggleNode(e.detail.node, e.detail.path)}
      />
    {/each}
  {/if}
</div>
