<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';

  export let src: string | undefined = undefined;
  export let alt: string = '';
  export let size: 'xs' | 'sm' | 'md' | 'lg' | 'xl' = 'md';
  export let shape: 'circle' | 'square' | 'rounded' = 'circle';
  export let fallback: string | undefined = undefined;

  function getInitials(name: string): string {
    return name
      .split(' ')
      .map(n => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  }

  const sizeClasses = {
    xs: 'w-6 h-6 text-xs',
    sm: 'w-8 h-8 text-xs',
    md: 'w-10 h-10 text-sm',
    lg: 'w-12 h-12 text-base',
    xl: 'w-16 h-16 text-lg'
  };

  const shapeClasses = {
    circle: 'rounded-full',
    square: 'rounded-none',
    rounded: 'rounded-lg'
  };
</script>

<div
  class={cn(
    'inline-flex items-center justify-center overflow-hidden bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300 font-semibold',
    sizeClasses[size],
    shapeClasses[shape],
    $$props.class
  )}
>
  {#if src}
    <img {src} {alt} class="w-full h-full object-cover" />
  {:else if fallback}
    {getInitials(fallback)}
  {:else}
    <svg class="w-1/2 h-1/2" fill="currentColor" viewBox="0 0 24 24">
      <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" />
    </svg>
  {/if}
</div>
