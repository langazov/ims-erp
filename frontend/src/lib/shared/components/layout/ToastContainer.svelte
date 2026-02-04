<script lang="ts">
  import { flip } from 'svelte/animate';
  import { fly, fade } from 'svelte/transition';
  import { toast, type Toast } from '$lib/shared/stores/toast';
  import ToastComponent from '$lib/shared/components/display/Toast.svelte';

  export let position: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' | 'top-center' | 'bottom-center' = 'top-right';

  const positionClasses = {
    'top-right': 'top-4 right-4',
    'top-left': 'top-4 left-4',
    'bottom-right': 'bottom-4 right-4',
    'bottom-left': 'bottom-4 left-4',
    'top-center': 'top-4 left-1/2 -translate-x-1/2',
    'bottom-center': 'bottom-4 left-1/2 -translate-x-1/2',
  };

  function handleDismiss(id: string) {
    toast.remove(id);
  }
</script>

<div class="toast-container {positionClasses[position]}">
  {#each $toast as toastItem (toastItem.id)}
    <div
      animate:flip={{ duration: 300 }}
      in:fly={{ x: position.includes('right') ? 100 : position.includes('left') ? -100 : 0, y: position.includes('top') ? -20 : 20, duration: 300 }}
      out:fade={{ duration: 200 }}
      class="toast-wrapper"
    >
      <ToastComponent
        variant={toastItem.variant}
        title={toastItem.title}
        duration={toastItem.duration}
        dismissible={toastItem.dismissible}
        position={position}
        on:dismiss={() => handleDismiss(toastItem.id)}
      >
        {toastItem.message}
      </ToastComponent>
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    pointer-events: none;
    max-width: 100%;
    padding: 1rem;
  }

  .toast-wrapper {
    pointer-events: auto;
  }

  /* Adjust for different positions */
  .toast-container:global(.top-right),
  .toast-container:global(.top-left),
  .toast-container:global(.top-center) {
    flex-direction: column;
  }

  .toast-container:global(.bottom-right),
  .toast-container:global(.bottom-left),
  .toast-container:global(.bottom-center) {
    flex-direction: column-reverse;
  }
</style>
