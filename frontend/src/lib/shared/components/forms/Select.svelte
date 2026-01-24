<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  export let id: string;
  export let label: string;
  export let value: string | number = '';
  export let options: { value: string | number; label: string; disabled?: boolean }[] = [];
  export let placeholder = 'Select an option';
  export let error = '';
  export let helpText = '';
  export let required = false;
  export let disabled = false;
  export let name: string | undefined = undefined;

  const dispatch = createEventDispatcher();

  let isOpen = false;

  function handleSelect(option: { value: string | number; label: string }) {
    if (!option.disabled) {
      value = option.value;
      isOpen = false;
      dispatch('change', value);
    }
  }

  function handleToggle() {
    if (!disabled) {
      isOpen = !isOpen;
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      isOpen = false;
    }
  }

  function handleBlur() {
    setTimeout(() => {
      isOpen = false;
    }, 200);
  }

  $: selectedOption = options.find((opt) => opt.value === value);
  $: displayValue = selectedOption?.label ?? placeholder;
  $: isPlaceholder = !selectedOption;
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="space-y-1">
  {#if label}
    <label for={id} class="block text-sm font-medium text-gray-700 dark:text-gray-300">
      {label}
      {#if required}
        <span class="text-red-500">*</span>
      {/if}
    </label>
  {/if}

  <div class="relative">
    <button
      type="button"
      {id}
      {name}
      {disabled}
      class={cn(
        'w-full rounded-lg border px-4 py-2.5 text-left text-sm',
        'bg-white dark:bg-gray-800',
        'text-gray-900 dark:text-white',
        'border-gray-300 dark:border-gray-600',
        'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent',
        'disabled:bg-gray-100 dark:disabled:bg-gray-700 disabled:cursor-not-allowed',
        'transition-colors duration-150',
        error ? 'border-red-500' : 'hover:border-gray-400',
        isOpen && 'ring-2 ring-primary-500 border-transparent'
      )}
      class:text-gray-500={isPlaceholder}
      on:click={handleToggle}
      on:focus={() => (isOpen = true)}
      on:blur={handleBlur}
      aria-expanded={isOpen}
      aria-haspopup="listbox"
    >
      <span class="block truncate">{displayValue}</span>
      <span
        class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none"
      >
        <svg
          class="w-5 h-5 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </span>
    </button>

    {#if isOpen}
      <div
        class="absolute z-50 w-full mt-1 bg-white dark:bg-gray-800 rounded-lg border border-gray-300 dark:border-gray-600 shadow-lg max-h-60 overflow-auto"
        role="listbox"
      >
        {#if placeholder && !required}
          <button
            type="button"
            class="w-full px-4 py-2 text-left text-sm text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700"
            on:click={() => handleSelect({ value: '', label: placeholder })}
          >
            {placeholder}
          </button>
        {/if}
        {#each options as option}
          <button
            type="button"
            class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
            class:bg-primary-50={option.value === value}
            class:text-primary-700={option.value === value}
            class:dark:bg-primary-900={option.value === value}
            class:dark:text-primary-300={option.value === value}
            class:opacity-50={option.disabled}
            class:cursor-not-allowed={option.disabled}
            disabled={option.disabled}
            role="option"
            aria-selected={option.value === value}
            on:click={() => handleSelect(option)}
          >
            {option.label}
          </button>
        {/each}
      </div>
    {/if}
  </div>

  {#if error}
    <p class="text-sm text-red-600 dark:text-red-400">{error}</p>
  {:else if helpText}
    <p class="text-sm text-gray-500 dark:text-gray-400">{helpText}</p>
  {/if}
</div>
