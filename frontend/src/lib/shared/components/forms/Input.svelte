<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  export let id: string;
  export let label: string;
  export let type: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url' | 'search' | 'date' =
    'text';
  export let value = '';
  export let placeholder = '';
  export let error = '';
  export let helpText = '';
  export let required = false;
  export let disabled = false;
  export let readonly = false;
  export let name: string | undefined = undefined;
  export let autocomplete: string | undefined = undefined;
  export let maxlength: number | undefined = undefined;
  export let minlength: number | undefined = undefined;
  export let pattern: string | undefined = undefined;
  export let inputmode: 'none' | 'text' | 'decimal' | 'numeric' | 'tel' | 'search' | 'email' | 'url' | undefined =
    undefined;

  const dispatch = createEventDispatcher();

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    dispatch('input', value);
  }

  function handleChange(event: Event) {
    dispatch('change', value);
  }

  function handleBlur(event: FocusEvent) {
    dispatch('blur', event);
  }

  function handleFocus(event: FocusEvent) {
    dispatch('focus', event);
  }

  $: inputClasses = cn(
    'w-full rounded-lg border px-4 py-2.5 text-sm',
    'bg-white dark:bg-gray-800',
    'text-gray-900 dark:text-white',
    'placeholder:text-gray-400 dark:placeholder:text-gray-500',
    'border-gray-300 dark:border-gray-600',
    'focus:outline-none focus:ring-2 focus:border-transparent',
    'disabled:bg-gray-100 dark:disabled:bg-gray-700 disabled:cursor-not-allowed disabled:opacity-50',
    'readonly:bg-gray-50 dark:readonly:bg-gray-800/50 readonly:cursor-not-allowed',
    error
      ? 'border-red-500 focus:ring-red-500'
      : 'focus:ring-primary-500',
    $$props.class
  );
</script>

<div class="space-y-1">
  {#if label}
    <label
      for={id}
      class="block text-sm font-medium text-gray-700 dark:text-gray-300"
    >
      {label}
      {#if required}
        <span class="text-red-500">*</span>
      {/if}
    </label>
  {/if}

  <input
    {id}
    {name}
    {type}
    {value}
    {placeholder}
    {required}
    {disabled}
    {readonly}
    {autocomplete}
    {maxlength}
    {minlength}
    {pattern}
    inputmode={inputmode}
    class={inputClasses}
    on:input={handleInput}
    on:change={handleChange}
    on:blur={handleBlur}
    on:focus={handleFocus}
    aria-invalid={error ? 'true' : undefined}
    aria-describedby={error ? `${id}-error` : helpText ? `${id}-help` : undefined}
  />

  {#if error}
    <p id="{id}-error" class="text-sm text-red-600 dark:text-red-400">
      {error}
    </p>
  {:else if helpText}
    <p id="{id}-help" class="text-sm text-gray-500 dark:text-gray-400">
      {helpText}
    </p>
  {/if}
</div>
