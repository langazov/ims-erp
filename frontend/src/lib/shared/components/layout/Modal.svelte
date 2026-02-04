<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';
  import { fade } from 'svelte/transition';

  export let open = false;
  export let size: 'sm' | 'md' | 'lg' | 'xl' | 'full' = 'md';
  export let title = '';
  export let closeOnEscape = true;
  export let closeOnBackdropClick = true;
  export let showCloseButton = true;

  const dispatch = createEventDispatcher();

  const sizeClasses = {
    sm: 'max-w-sm',
    md: 'max-w-lg',
    lg: 'max-w-2xl',
    xl: 'max-w-4xl',
    full: 'max-w-full mx-4',
  };

  function handleBackdropClick(event: MouseEvent) {
    if (closeOnBackdropClick && event.target === event.currentTarget) {
      close();
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (closeOnEscape && event.key === 'Escape') {
      close();
    }
  }

  function close() {
    open = false;
    dispatch('close');
  }
</script>

{#if open}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center"
    role="dialog"
    aria-modal="true"
    on:click={handleBackdropClick}
    on:keydown={handleKeydown}
    transition:fade={{ duration: 150 }}
  >
    <!-- Backdrop -->
    <div class="absolute inset-0 bg-black/50 backdrop-blur-sm"></div>

    <!-- Modal Content -->
    <div
      class={cn(
        'relative z-10 w-full rounded-lg bg-white dark:bg-gray-800',
        'shadow-xl',
        sizeClasses[size]
      )}
      transition:fade={{ duration: 150 }}
    >
      <!-- Header -->
      {#if title || showCloseButton}
        <div
          class={cn(
            'flex items-center justify-between border-b border-gray-200 px-6 py-4 dark:border-gray-700',
            !title && 'border-b-0'
          )}
        >
          {#if title}
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{title}</h2>
          {:else}
            <div></div>
          {/if}

          {#if showCloseButton}
            <button
              type="button"
              class="inline-flex rounded-md p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-500 focus:outline-none dark:hover:bg-gray-700 dark:hover:text-gray-300"
              on:click={close}
              aria-label="Close"
            >
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          {/if}
        </div>
      {/if}

      <!-- Body -->
      <div class={cn('px-6', title || showCloseButton ? 'py-4' : 'py-6')}>
        <slot />
      </div>

      <!-- Footer -->
      {#if $$slots.footer}
        <div
          class="flex items-center justify-end gap-3 border-t border-gray-200 px-6 py-4 dark:border-gray-700"
        >
          <slot name="footer" {close} />
        </div>
      {/if}
    </div>
  </div>
{/if}
