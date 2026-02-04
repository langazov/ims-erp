<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';

  export let variant: 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' = 'default';
  export let size: 'sm' | 'md' = 'md';
  export let removable = false;

  const variantStyles = {
    default: 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
    primary: 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300',
    success: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
    warning: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
    danger: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
    info: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
  };
</script>

<span
  class={cn(
    'inline-flex items-center gap-1.5 rounded-full font-medium',
    variantStyles[variant],
    size === 'sm' && 'px-2 py-0.5 text-xs',
    size === 'md' && 'px-2.5 py-0.5 text-sm',
    $$props.class
  )}
>
  <slot />
  
  {#if removable}
    <button
      type="button"
      class="ml-0.5 -mr-0.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10 p-0.5 transition-colors"
      on:click
    >
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  {/if}
</span>
