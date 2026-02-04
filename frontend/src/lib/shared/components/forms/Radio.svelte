<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  export let id: string;
  export let label: string | undefined = undefined;
  export let options: { value: string; label: string; disabled?: boolean }[] = [];
  export let value: string = '';
  export let name: string | undefined = undefined;
  export let disabled = false;
  export let required = false;
  export let error: string | undefined = undefined;
  export let helper: string | undefined = undefined;
  export let inline = false;

  const dispatch = createEventDispatcher();

  function handleChange(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    dispatch('change', value);
  }
</script>

<div class={cn($$props.class)}>
  {#if label}
    <label
      class={cn(
        'block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2',
        disabled && 'opacity-50'
      )}
    >
      {label}
      {#if required}
        <span class="text-red-500">*</span>
      {/if}
    </label>
  {/if}

  <div class={cn('gap-3', inline ? 'flex flex-wrap' : 'flex flex-col')}>
    {#each options as option}
      <label
        class={cn(
          'flex items-center gap-3 cursor-pointer',
          (disabled || option.disabled) && 'cursor-not-allowed opacity-50'
        )}
      >
        <div class="relative flex items-center">
          <input
            type="radio"
            {name}
            value={option.value}
            checked={value === option.value}
            disabled={disabled || option.disabled}
            {required}
            class={cn(
              'peer h-5 w-5 cursor-pointer appearance-none rounded-full border border-gray-300',
              'bg-white dark:bg-gray-800 dark:border-gray-600',
              'checked:border-primary-600 checked:bg-primary-600',
              'focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2',
              error && 'border-red-500'
            )}
            on:change={handleChange}
          />
          <div
            class={cn(
              'pointer-events-none absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2',
              'w-2.5 h-2.5 rounded-full bg-white opacity-0 peer-checked:opacity-100'
            )}
          />
        </div>
        
        <span class="text-sm text-gray-700 dark:text-gray-300">{option.label}</span>
      </label>
    {/each}
  </div>

  {#if helper}
    <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">{helper}</p>
  {/if}

  {#if error}
    <p class="text-xs text-red-600 dark:text-red-400 mt-2">{error}</p>
  {/if}
</div>
