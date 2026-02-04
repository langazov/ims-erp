<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  export let id: string;
  export let label: string | undefined = undefined;
  export let value: string = '';
  export let placeholder: string | undefined = undefined;
  export let rows: number = 4;
  export let disabled = false;
  export let required = false;
  export let readonly = false;
  export let error: string | undefined = undefined;
  export let helper: string | undefined = undefined;
  export let maxLength: number | undefined = undefined;
  export let resize: 'none' | 'vertical' | 'horizontal' | 'both' = 'vertical';

  const dispatch = createEventDispatcher();

  function handleInput(event: Event) {
    const target = event.target as HTMLTextAreaElement;
    value = target.value;
    dispatch('input', value);
  }

  function handleChange(event: Event) {
    const target = event.target as HTMLTextAreaElement;
    dispatch('change', target.value);
  }

  const resizeClasses = {
    none: 'resize-none',
    vertical: 'resize-y',
    horizontal: 'resize-x',
    both: 'resize'
  };
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

  <textarea
    {id}
    {placeholder}
    {rows}
    {disabled}
    {required}
    {readonly}
    maxlength={maxLength}
    class={cn(
      'w-full px-3 py-2 rounded-lg border bg-white dark:bg-gray-800',
      'text-gray-900 dark:text-gray-100 placeholder-gray-400',
      'focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:border-primary-500',
      'disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-gray-50 dark:disabled:bg-gray-900',
      'read-only:bg-gray-50 dark:read-only:bg-gray-900',
      resizeClasses[resize],
      error
        ? 'border-red-500 focus-visible:ring-red-500 focus-visible:border-red-500'
        : 'border-gray-300 dark:border-gray-600'
    )}
    bind:value
    on:input={handleInput}
    on:change={handleChange}
    on:blur
    on:focus
  />

  <div class="flex justify-between mt-1.5">
    {#if helper}
      <p class="text-xs text-gray-500 dark:text-gray-400">{helper}</p>
    {/if}
    
    {#if maxLength}
      <p class="text-xs text-gray-400 dark:text-gray-500 ml-auto">
        {value.length}/{maxLength}
      </p>
    {/if}
    
    {#if error}
      <p class="text-xs text-red-600 dark:text-red-400">{error}</p>
    {/if}
  </div>
</div>
