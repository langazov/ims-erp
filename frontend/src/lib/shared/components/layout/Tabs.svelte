<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  export let tabs: { id: string; label: string; disabled?: boolean }[] = [];
  export let activeTab = tabs[0]?.id ?? '';
  export let variant: 'default' | 'pills' | 'underline' = 'default';
  export let fullWidth = false;
  export let size: 'sm' | 'md' = 'md';

  const dispatch = createEventDispatcher();

  function selectTab(tabId: string) {
    const tab = tabs.find((t) => t.id === tabId);
    if (tab?.disabled) return;

    activeTab = tabId;
    dispatch('change', tabId);
  }

  const variantClasses = {
    default: {
      container: 'border-b border-gray-200 dark:border-gray-700',
      tab: 'border-b-2 border-transparent px-4 py-2 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:border-gray-600 dark:hover:text-gray-300',
      active: 'border-primary-600 text-primary-600 dark:border-primary-500 dark:text-primary-400',
      inactive: '',
    },
    pills: {
      container: 'gap-1',
      tab: 'rounded-full px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700',
      active: 'bg-primary-600 text-white hover:bg-primary-700 dark:bg-primary-600 dark:hover:bg-primary-700',
      inactive: '',
    },
    underline: {
      container: 'border-b border-gray-200 dark:border-gray-700',
      tab: 'border-b-2 border-transparent px-4 py-2 text-sm font-medium text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300',
      active: 'border-primary-600 text-primary-600 dark:border-primary-500 dark:text-primary-400',
      inactive: '',
    },
  };

  const sizeClasses = {
    sm: 'px-3 py-1.5 text-xs',
    md: 'px-4 py-2 text-sm',
  };
</script>

<div class={cn('w-full', $$props.class)}>
  <div
    class={cn(
      'flex',
      variantClasses[variant].container,
      fullWidth && 'w-full',
      fullWidth && variant === 'pills' && 'grid grid-flow-col auto-cols-fr'
    )}
    role="tablist"
  >
    {#each tabs as tab}
      <button
        type="button"
        role="tab"
        aria-selected={activeTab === tab.id}
        class={cn(
          'transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
          variantClasses[variant].tab,
          activeTab === tab.id && variantClasses[variant].active,
          size === 'sm' && variant === 'pills' && sizeClasses.sm,
          tab.disabled && 'cursor-not-allowed opacity-50'
        )}
        on:click={() => selectTab(tab.id)}
        disabled={tab.disabled}
      >
        {tab.label}
      </button>
    {/each}
  </div>

  <div class="mt-4">
    <slot />
  </div>
</div>
