<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  export let variant: 'primary' | 'secondary' | 'danger' | 'ghost' | 'link' = 'primary';
  export let size: 'sm' | 'md' | 'lg' = 'md';
  export let disabled = false;
  export let loading = false;
  export let type: 'button' | 'submit' | 'reset' = 'button';
  export let fullWidth = false;
  export let href: string | undefined = undefined;

  const dispatch = createEventDispatcher();

  function handleClick(event: MouseEvent) {
    if (!disabled && !loading) {
      dispatch('click', event);
    }
  }

  $: baseClasses = cn(
    'inline-flex items-center justify-center font-medium rounded-lg',
    'transition-all duration-150',
    'focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2',
    'disabled:opacity-50 disabled:cursor-not-allowed',
    'active:scale-[0.98]',
    {
      'bg-primary-600 hover:bg-primary-700 text-white focus-visible:ring-primary-500':
        variant === 'primary',
      'bg-gray-100 hover:bg-gray-200 text-gray-700 dark:bg-gray-700 dark:hover:bg-gray-600 dark:text-gray-200 focus-visible:ring-gray-500':
        variant === 'secondary',
      'bg-red-600 hover:bg-red-700 text-white focus-visible:ring-red-500':
        variant === 'danger',
      'text-gray-600 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-300 dark:hover:text-white dark:hover:bg-gray-800 focus-visible:ring-gray-500':
        variant === 'ghost',
      'text-primary-600 hover:text-primary-700 underline-offset-4 hover:underline focus-visible:ring-primary-500':
        variant === 'link',
    },
    {
      'px-3 py-1.5 text-xs': size === 'sm',
      'px-4 py-2 text-sm': size === 'md',
      'px-6 py-3 text-base': size === 'lg',
    },
    {
      'w-full': fullWidth,
    },
    loading && 'cursor-wait'
  );
</script>

{#if href}
  <a
    {href}
    class={cn(baseClasses, $$props.class)}
    class:pointer-events-none={disabled || loading}
    on:click={handleClick}
  >
    {#if loading}
      <svg
        class="animate-spin -ml-1 mr-2 h-4 w-4"
        class:h-3={size === 'sm'}
        class:h-4={size === 'md'}
        class:h-5={size === 'lg'}
        class:mr-1={size === 'sm'}
        class:mr-2={size !== 'sm'}
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          class="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          stroke-width="4"
        />
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
    {/if}
    <slot />
  </a>
{:else}
  <button
    {type}
    {disabled}
    class={cn(baseClasses, $$props.class)}
    on:click={handleClick}
  >
    {#if loading}
      <svg
        class="animate-spin -ml-1 mr-2 h-4 w-4"
        class:h-3={size === 'sm'}
        class:h-4={size === 'md'}
        class:h-5={size === 'lg'}
        class:mr-1={size === 'sm'}
        class:mr-2={size !== 'sm'}
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          class="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          stroke-width="4"
        />
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
    {/if}
    <slot />
  </button>
{/if}
