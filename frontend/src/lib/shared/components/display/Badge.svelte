<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';

  export let variant: 'gray' | 'green' | 'yellow' | 'red' | 'blue' | 'purple' | 'orange' = 'gray';
  export let size: 'sm' | 'md' = 'md';
  export let dot = false;
</script>

<span
  class={cn(
    'inline-flex items-center font-medium rounded-full',
    variant === 'gray' &&
      'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200',
    variant === 'green' &&
      'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-400',
    variant === 'yellow' &&
      'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-400',
    variant === 'red' && 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-400',
    variant === 'blue' &&
      'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-400',
    variant === 'purple' &&
      'bg-purple-100 text-purple-800 dark:bg-purple-900/50 dark:text-purple-400',
    variant === 'orange' &&
      'bg-orange-100 text-orange-800 dark:bg-orange-900/50 dark:text-orange-400',
    size === 'sm' && 'px-2 py-0.5 text-xs',
    size === 'md' && 'px-2.5 py-1 text-sm'
  )}
>
  {#if dot}
    <span
      class={cn(
        'w-1.5 h-1.5 rounded-full mr-1.5',
        variant === 'gray' && 'bg-gray-500',
        variant === 'green' && 'bg-green-500',
        variant === 'yellow' && 'bg-yellow-500',
        variant === 'red' && 'bg-red-500',
        variant === 'blue' && 'bg-blue-500',
        variant === 'purple' && 'bg-purple-500',
        variant === 'orange' && 'bg-orange-500'
      )}
    ></span>
  {/if}
  <slot />
</span>
