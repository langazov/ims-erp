<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';
  import { fade } from 'svelte/transition';

  export let content: string = '';
  export let position: 'top' | 'bottom' | 'left' | 'right' = 'top';
  export let delay: number = 200;
  export let disabled: boolean = false;

  let show = false;
  let timeoutId: ReturnType<typeof setTimeout>;
  let tooltipElement: HTMLDivElement;

  function handleMouseEnter() {
    if (disabled) return;
    timeoutId = setTimeout(() => {
      show = true;
    }, delay);
  }

  function handleMouseLeave() {
    clearTimeout(timeoutId);
    show = false;
  }

  function handleFocus() {
    if (disabled) return;
    show = true;
  }

  function handleBlur() {
    show = false;
  }

  const positionClasses = {
    top: 'bottom-full left-1/2 -translate-x-1/2 mb-2',
    bottom: 'top-full left-1/2 -translate-x-1/2 mt-2',
    left: 'right-full top-1/2 -translate-y-1/2 mr-2',
    right: 'left-full top-1/2 -translate-y-1/2 ml-2',
  };

  const arrowClasses = {
    top: 'top-full left-1/2 -translate-x-1/2 border-t-gray-900 dark:border-t-gray-700',
    bottom: 'bottom-full left-1/2 -translate-x-1/2 border-b-gray-900 dark:border-b-gray-700',
    left: 'left-full top-1/2 -translate-y-1/2 border-l-gray-900 dark:border-l-gray-700',
    right: 'right-full top-1/2 -translate-y-1/2 border-r-gray-900 dark:border-r-gray-700',
  };
</script>

<div
  class="relative inline-block"
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
  on:focus={handleFocus}
  on:blur={handleBlur}
  role="tooltip"
  aria-describedby={show ? 'tooltip-content' : undefined}
>
  <slot />
  
  {#if show && content}
    <div
      bind:this={tooltipElement}
      id="tooltip-content"
      class={cn(
        'absolute z-50 px-3 py-2 text-sm font-medium text-white bg-gray-900 dark:bg-gray-700 rounded-lg shadow-lg whitespace-nowrap',
        'transition-opacity duration-150',
        positionClasses[position]
      )}
      transition:fade={{ duration: 150 }}
    >
      {content}
      <!-- Arrow -->
      <div
        class={cn(
          'absolute w-0 h-0 border-4 border-transparent',
          arrowClasses[position]
        )}
      ></div>
    </div>
  {/if}
</div>
