<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';
  import { fly } from 'svelte/transition';

  export let open = false;
  export let position: 'left' | 'right' = 'right';
  export let size: 'sm' | 'md' | 'lg' | 'xl' | 'full' = 'md';
  export let title = '';
  export let showOverlay = true;
  export let closeOnEscape = true;
  export let closeOnBackdropClick = true;

  const dispatch = createEventDispatcher();

  const sizeClasses = {
    sm: 'w-80',
    md: 'w-96',
    lg: 'w-[30rem]',
    xl: 'w-[40rem]',
    full: 'w-full max-w-full',
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
    class="fixed inset-0 z-50 flex"
    role="dialog"
    aria-modal="true"
    class:justify-start={position === 'left'}
    class:justify-end={position === 'right'}
    on:click={handleBackdropClick}
    on:keydown={handleKeydown}
  >
    <!-- Overlay -->
    {#if showOverlay}
      <div
        class="absolute inset-0 bg-black/50 backdrop-blur-sm"
        transition:fly={{ duration: 200 }}
      ></div>
    {/if}

    <!-- Drawer Content -->
    <div
      class={cn(
        'relative z-10 h-full bg-white shadow-xl dark:bg-gray-800',
        'flex flex-col',
        sizeClasses[size]
      )}
      transition:fly={{
        duration: 300,
        x: position === 'left' ? -300 : 300,
      }}
    >
      <!-- Header -->
      {#if title}
        <div
          class="flex items-center justify-between border-b border-gray-200 px-6 py-4 dark:border-gray-700"
        >
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{title}</h2>
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
        </div>
      {/if}

      <!-- Body -->
      <div class="flex-1 overflow-y-auto p-6">
        <slot />
      </div>

      <!-- Footer -->
      {#if $$slots.footer}
        <div
          class="border-t border-gray-200 px-6 py-4 dark:border-gray-700"
        >
          <slot name="footer" {close} />
        </div>
      {/if}
    </div>
  </div>
{/if}
