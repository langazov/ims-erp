<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';
  import { fly, fade } from 'svelte/transition';

  export let variant: 'success' | 'error' | 'warning' | 'info' = 'info';
  export let title: string | undefined = undefined;
  export let duration: number = 5000;
  export let dismissible: boolean = true;
  export let position: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' | 'top-center' | 'bottom-center' = 'top-right';

  const dispatch = createEventDispatcher();

  let timeoutId: ReturnType<typeof setTimeout>;

  const variantStyles = {
    success: 'bg-green-50 border-green-200 text-green-800 dark:bg-green-900/20 dark:border-green-800 dark:text-green-200',
    error: 'bg-red-50 border-red-200 text-red-800 dark:bg-red-900/20 dark:border-red-800 dark:text-red-200',
    warning: 'bg-yellow-50 border-yellow-200 text-yellow-800 dark:bg-yellow-900/20 dark:border-yellow-800 dark:text-yellow-200',
    info: 'bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-900/20 dark:border-blue-800 dark:text-blue-200'
  };

  const iconPaths = {
    success: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
    error: 'M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    warning: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
    info: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
  };

  const positionClasses = {
    'top-right': 'top-4 right-4',
    'top-left': 'top-4 left-4',
    'bottom-right': 'bottom-4 right-4',
    'bottom-left': 'bottom-4 left-4',
    'top-center': 'top-4 left-1/2 -translate-x-1/2',
    'bottom-center': 'bottom-4 left-1/2 -translate-x-1/2'
  };

  function startTimer() {
    if (duration > 0) {
      timeoutId = setTimeout(() => {
        dismiss();
      }, duration);
    }
  }

  function clearTimer() {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
  }

  function dismiss() {
    clearTimer();
    dispatch('dismiss');
  }

  startTimer();
</script>

<div
  class={cn(
    'fixed z-50 max-w-sm w-full shadow-lg rounded-lg border p-4',
    'transition-all duration-300',
    variantStyles[variant],
    positionClasses[position],
    $$props.class
  )}
  role="alert"
  transition:fly={{ y: -20, duration: 300 }}
  on:mouseenter={clearTimer}
  on:mouseleave={startTimer}
>
  <div class="flex items-start gap-3">
    <svg class="w-5 h-5 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={iconPaths[variant]} />
    </svg>
    
    <div class="flex-1 min-w-0">
      {#if title}
        <h3 class="font-semibold text-sm">{title}</h3>
      {/if}
      <div class={cn('text-sm', title && 'mt-1')}>
        <slot />
      </div>
    </div>

    {#if dismissible}
      <button
        type="button"
        class="flex-shrink-0 -mr-1 -mt-1 p-1 rounded hover:bg-black/5 dark:hover:bg-white/5 transition-colors"
        on:click={dismiss}
        aria-label="Dismiss"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    {/if}
  </div>

  {#if duration > 0}
    <div class="mt-3 h-1 bg-black/10 dark:bg-white/10 rounded-full overflow-hidden">
      <div
        class="h-full bg-current opacity-30"
        style="animation: shrink {duration}ms linear forwards"
      />
    </div>
  {/if}
</div>

<style>
  @keyframes shrink {
    from { width: 100%; }
    to { width: 0%; }
  }
</style>
