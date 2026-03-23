<script lang="ts">
  export let title: string = 'No data found';
  export let description: string = '';
  export let actionLabel: string = '';
  export let actionHref: string = '';
  export let icon: string = '📭';

  import Button from '$lib/shared/components/forms/Button.svelte';
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
</script>

<div class="flex flex-col items-center justify-center py-16 px-8 text-center">
  <div class="w-20 h-20 bg-gray-100 dark:bg-gray-700 rounded-full flex items-center justify-center mb-6">
    <span class="text-3xl">{icon}</span>
  </div>
  <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">{title}</h3>
  {#if description}
    <p class="text-sm text-gray-500 dark:text-gray-400 max-w-sm mb-6">{description}</p>
  {/if}
  {#if actionLabel}
    {#if actionHref}
      <Button variant="primary" href={actionHref}>{actionLabel}</Button>
    {:else}
      <Button variant="primary" on:click={() => dispatch('action')}>{actionLabel}</Button>
    {/if}
  {/if}
  <slot />
</div>
