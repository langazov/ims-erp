<script lang="ts">
  export let title: string;
  export let value: string | number;
  export let change: number | null = null; // percentage change, positive = good
  export let changeLabel: string = 'vs last period';
  export let icon: string = '';
  export let iconBg: string = 'bg-blue-100';
  export let iconColor: string = 'text-blue-600';
  export let loading: boolean = false;
  export let href: string | undefined = undefined;

  $: isPositive = change !== null && change >= 0;
  $: isNegative = change !== null && change < 0;
</script>

<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6 {href ? 'hover:shadow-md transition-shadow cursor-pointer' : ''}" on:click={href ? () => window.location.href = href : undefined}>
  {#if loading}
    <div class="animate-pulse">
      <div class="flex items-center justify-between mb-4">
        <div class="h-4 bg-gray-200 rounded w-1/2"></div>
        <div class="w-10 h-10 bg-gray-200 rounded-lg"></div>
      </div>
      <div class="h-8 bg-gray-200 rounded w-1/3 mb-2"></div>
      <div class="h-3 bg-gray-200 rounded w-2/3"></div>
    </div>
  {:else}
    <div class="flex items-center justify-between mb-4">
      <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{title}</p>
      {#if icon}
        <div class="w-10 h-10 {iconBg} rounded-lg flex items-center justify-center">
          <span class="text-lg {iconColor}">{icon}</span>
        </div>
      {/if}
    </div>
    <div class="flex items-end gap-3">
      <p class="text-3xl font-bold text-gray-900 dark:text-white">{value}</p>
      {#if change !== null}
        <div class="flex items-center gap-1 mb-1 {isPositive ? 'text-emerald-600' : 'text-red-500'}">
          <span class="text-sm font-medium">
            {isPositive ? '↑' : '↓'} {Math.abs(change).toFixed(1)}%
          </span>
        </div>
      {/if}
    </div>
    {#if change !== null}
      <p class="text-xs text-gray-400 mt-1">{changeLabel}</p>
    {/if}
    <slot />
  {/if}
</div>
