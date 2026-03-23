<script lang="ts">
  import Modal from './Modal.svelte';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import { createEventDispatcher } from 'svelte';

  export let open: boolean = false;
  export let title: string = 'Confirm Action';
  export let message: string = 'Are you sure you want to proceed?';
  export let confirmLabel: string = 'Confirm';
  export let cancelLabel: string = 'Cancel';
  export let variant: 'danger' | 'warning' | 'primary' = 'danger';
  export let loading: boolean = false;

  const dispatch = createEventDispatcher<{ confirm: void; cancel: void }>();

  function handleConfirm() {
    dispatch('confirm');
  }
  function handleCancel() {
    open = false;
    dispatch('cancel');
  }
</script>

<Modal bind:open {title} size="sm">
  <p class="text-sm text-gray-600 dark:text-gray-300">{message}</p>
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={handleCancel} disabled={loading}>{cancelLabel}</Button>
    <Button variant={variant} on:click={handleConfirm} {loading}>{confirmLabel}</Button>
  </svelte:fragment>
</Modal>
