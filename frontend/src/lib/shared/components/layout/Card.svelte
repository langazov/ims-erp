<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';

  export let padding: 'none' | 'sm' | 'md' | 'lg' = 'md';
  export let hover = false;
  export let bordered = true;
</script>

<div
  class={cn(
    'bg-white dark:bg-gray-800 rounded-xl',
    bordered && 'border border-gray-200 dark:border-gray-700',
    hover && 'hover:shadow-md transition-shadow duration-200',
    padding === 'none' && '',
    padding === 'sm' && 'p-4',
    padding === 'md' && 'p-6',
    padding === 'lg' && 'p-8',
    $$props.class
  )}
>
  <slot />
</div>
