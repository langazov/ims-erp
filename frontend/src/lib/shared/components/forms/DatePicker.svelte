<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  export let id: string;
  export let label: string | undefined = undefined;
  export let value: string = '';
  export let min: string | undefined = undefined;
  export let max: string | undefined = undefined;
  export let disabled = false;
  export let required = false;
  export let error: string | undefined = undefined;
  export let helper: string | undefined = undefined;

  const dispatch = createEventDispatcher();

  function handleChange(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    dispatch('change', value);
  }
</script>

<div class={cn('w-full', $$props.class)}>
  {#if label}
    <label
      for={id}
      class={cn(
        'block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5',
        disabled && 'opacity-50'
      )}
    >
      {label}
      {#if required}
        <span class="text-red-500">*</span>
      {/if}
    </label>
  {/if}

  <input
    {id}
    type="date"
    {value}
    {min}
    {max}
    {disabled}
    {required}
    class={cn(
      'w-full px-3 py-2 rounded-lg border bg-white dark:bg-gray-800',
      'text-gray-900 dark:text-gray-100',
      'focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:border-primary-500',
      'disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-gray-50 dark:disabled:bg-gray-900',
      error
        ? 'border-red-500 focus-visible:ring-red-500 focus-visible:border-red-500'
        : 'border-gray-300 dark:border-gray-600'
    )}
    on:change={handleChange}
    on:blur
    on:focus
  />

  {#if helper}
    <p class="text-xs text-gray-500 dark:text-gray-400 mt-1.5">{helper}</p>
  {/if}

  {#if error}
    <p class="text-xs text-red-600 dark:text-red-400 mt-1.5">{error}</p>
  {/if}
</div>
