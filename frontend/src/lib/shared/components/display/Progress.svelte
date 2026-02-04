<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';

  export let value: number = 0;
  export let max: number = 100;
  export let size: 'sm' | 'md' | 'lg' = 'md';
  export let variant: 'default' | 'success' | 'warning' | 'danger' = 'default';
  export let indeterminate = false;
  export let showLabel = false;

  $: percentage = Math.min(100, Math.max(0, (value / max) * 100));

  const sizeClasses = {
    sm: 'h-1.5',
    md: 'h-2',
    lg: 'h-3'
  };

  const variantClasses = {
    default: 'bg-primary-600',
    success: 'bg-green-600',
    warning: 'bg-yellow-600',
    danger: 'bg-red-600'
  };
</script>

<div class={cn('w-full', $$props.class)}>
  {#if showLabel}
    <div class="flex justify-between text-sm text-gray-600 dark:text-gray-400 mb-1">
      <slot name="label">
        <span>{Math.round(percentage)}%</span>
      </slot>
    </div>
  {/if}
  
  <div
    class={cn(
      'w-full bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden',
      sizeClasses[size]
    )}
  >
    <div
      class={cn(
        'h-full rounded-full transition-all duration-300 ease-out',
        variantClasses[variant],
        indeterminate && 'animate-pulse w-1/3'
      )}
      style={indeterminate ? '' : `width: ${percentage}%`}
    />
  </div>
</div>
