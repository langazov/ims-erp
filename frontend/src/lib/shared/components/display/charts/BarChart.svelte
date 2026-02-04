<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';

  interface DataPoint {
    label: string;
    value: number;
    color?: string;
  }

  export let data: DataPoint[] = [];
  export let title: string = '';
  export let height: number = 300;
  export let showValues: boolean = true;
  export let horizontal: boolean = false;
  export let className: string = '';

  $: maxValue = Math.max(...data.map(d => d.value), 1);

  const defaultColors = [
    'bg-primary-500',
    'bg-blue-500',
    'bg-green-500',
    'bg-yellow-500',
    'bg-purple-500',
    'bg-pink-500',
    'bg-indigo-500',
    'bg-red-500',
  ];

  function getColor(index: number, customColor?: string): string {
    if (customColor) return customColor;
    return defaultColors[index % defaultColors.length];
  }
</script>

<div class={cn('bg-white dark:bg-gray-800 rounded-lg shadow p-6', className)}>
  {#if title}
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{title}</h3>
  {/if}

  <div class="relative" style="height: {height}px;">
    {#if data.length === 0}
      <div class="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
        No data available
      </div>
    {:else}
      <div class={cn(
        'flex h-full gap-2',
        horizontal ? 'flex-col' : 'flex-row items-end'
      )}>
        {#each data as item, index}
          <div class={cn(
            'flex',
            horizontal ? 'flex-row items-center gap-2' : 'flex-col items-center flex-1'
          )}>
            {#if horizontal}
              <!-- Horizontal Bar -->
              <div class="flex-1 flex items-center gap-2">
                <span class="text-sm text-gray-600 dark:text-gray-400 w-24 truncate">{item.label}</span>
                <div class="flex-1 bg-gray-200 dark:bg-gray-700 rounded-full h-6 overflow-hidden">
                  <div
                    class={cn('h-full rounded-full transition-all duration-500 ease-out', getColor(index, item.color))}
                    style="width: {(item.value / maxValue) * 100}%"
                  ></div>
                </div>
                {#if showValues}
                  <span class="text-sm font-medium text-gray-900 dark:text-white w-12 text-right">
                    {item.value.toLocaleString()}
                  </span>
                {/if}
              </div>
            {:else}
              <!-- Vertical Bar -->
              {#if showValues}
                <span class="text-sm font-medium text-gray-900 dark:text-white mb-1">
                  {item.value.toLocaleString()}
                </span>
              {/if}
              <div class="flex-1 w-full bg-gray-200 dark:bg-gray-700 rounded-t-lg relative overflow-hidden">
                <div
                  class={cn('absolute bottom-0 w-full rounded-t-lg transition-all duration-500 ease-out', getColor(index, item.color))}
                  style="height: {(item.value / maxValue) * 100}%"
                ></div>
              </div>
              <span class="text-xs text-gray-600 dark:text-gray-400 mt-2 text-center truncate w-full">
                {item.label}
              </span>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

