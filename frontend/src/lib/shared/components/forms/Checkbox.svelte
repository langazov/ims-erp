<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  export let id: string;
  export let label: string | undefined = undefined;
  export let checked = false;
  export let indeterminate = false;
  export let disabled = false;
  export let required = false;
  export let error: string | undefined = undefined;
  export let helper: string | undefined = undefined;

  const dispatch = createEventDispatcher();

  function handleChange(event: Event) {
    const target = event.target as HTMLInputElement;
    checked = target.checked;
    dispatch('change', checked);
  }
</script>

<div class={cn('flex items-start gap-3', $$props.class)}>
  <div class="relative flex items-center">
    <input
      {id}
      type="checkbox"
      class={cn(
        'peer h-5 w-5 cursor-pointer appearance-none rounded border border-gray-300',
        'bg-white dark:bg-gray-800 dark:border-gray-600',
        'checked:border-primary-600 checked:bg-primary-600',
        'indeterminate:border-primary-600 indeterminate:bg-primary-600',
        'focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2',
        'disabled:cursor-not-allowed disabled:opacity-50',
        error && 'border-red-500'
      )}
      bind:checked
      {indeterminate}
      {disabled}
      {required}
      on:change={handleChange}
    />
    <svg
      class={cn(
        'pointer-events-none absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2',
        'text-white opacity-0 peer-checked:opacity-100',
        'w-3.5 h-3.5'
      )}
      viewBox="0 0 14 14"
      fill="none"
    >
      <path
        d="M3 8L6 11L11 3.5"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
  </div>

  {#if label || helper || error}
    <div class="flex flex-col">
      {#if label}
        <label
          for={id}
          class={cn(
            'text-sm font-medium text-gray-700 dark:text-gray-300',
            disabled && 'opacity-50',
            error && 'text-red-600 dark:text-red-400'
          )}
        >
          {label}
          {#if required}
            <span class="text-red-500">*</span>
          {/if}
        </label>
      {/if}
      
      {#if helper}
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{helper}</p>
      {/if}
      
      {#if error}
        <p class="text-xs text-red-600 dark:text-red-400 mt-0.5">{error}</p>
      {/if}
    </div>
  {/if}
</div>
