<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';

  interface DataPoint {
    x: string | number;
    y: number;
  }

  interface LineSeries {
    name: string;
    data: DataPoint[];
    color?: string;
  }

  export let series: LineSeries[] = [];
  export let title: string = '';
  export let height: number = 300;
  export let showGrid: boolean = true;
  export let showLegend: boolean = true;
  export let className: string = '';

  const padding = { top: 20, right: 20, bottom: 40, left: 60 };

  $: allYValues = series.flatMap(s => s.data.map(d => d.y));
  $: maxY = Math.max(...allYValues, 1);
  $: minY = Math.min(...allYValues, 0);
  $: yRange = maxY - minY || 1;

  $: allXValues = series.flatMap(s => s.data.map(d => d.x));
  $: xLabels = [...new Set(allXValues)];

  function getYPosition(value: number): number {
    return padding.top + (1 - (value - minY) / yRange) * (height - padding.top - padding.bottom);
  }

  function getXPosition(index: number): number {
    return padding.left + (index / (xLabels.length - 1 || 1)) * (300 - padding.left - padding.right);
  }

  function generatePath(data: DataPoint[]): string {
    if (data.length === 0) return '';
    
    const points = data.map((d, i) => {
      const xIndex = xLabels.indexOf(d.x);
      const x = getXPosition(xIndex);
      const y = getYPosition(d.y);
      return `${x},${y}`;
    });

    return `M ${points.join(' L ')}`;
  }

  const defaultColors = [
    '#3b82f6',
    '#10b981',
    '#f59e0b',
    '#ef4444',
    '#8b5cf6',
    '#ec4899',
    '#06b6d4',
    '#f97316',
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
    {#if series.length === 0 || series.every(s => s.data.length === 0)}
      <div class="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
        No data available
      </div>
    {:else}
      <svg class="w-full h-full" viewBox="0 0 400 {height}" preserveAspectRatio="xMidYMid meet">
        <!-- Grid Lines -->
        {#if showGrid}
          {#each Array(5) as _, i}
            {@const y = padding.top + (i / 4) * (height - padding.top - padding.bottom)}
            {@const value = maxY - (i / 4) * yRange}
            <line
              x1={padding.left}
              y1={y}
              x2={400 - padding.right}
              y2={y}
              stroke="#e5e7eb"
              stroke-dasharray="4"
            />
            <text
              x={padding.left - 10}
              y={y + 4}
              text-anchor="end"
              class="text-xs fill-gray-500 dark:fill-gray-400"
            >
              {Math.round(value).toLocaleString()}
            </text>
          {/each}
        {/if}

        <!-- X Axis Labels -->
        {#each xLabels as label, i}
          {@const x = getXPosition(i)}
          <text
            x={x}
            y={height - 10}
            text-anchor="middle"
            class="text-xs fill-gray-500 dark:fill-gray-400"
          >
            {label}
          </text>
        {/each}

        <!-- Lines -->
        {#each series as s, seriesIndex}
          <path
            d={generatePath(s.data)}
            fill="none"
            stroke={getColor(seriesIndex, s.color)}
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          
          <!-- Data Points -->
          {#each s.data as point}
            {@const xIndex = xLabels.indexOf(point.x)}
            {@const x = getXPosition(xIndex)}
            {@const y = getYPosition(point.y)}
            <circle
              cx={x}
              cy={y}
              r="4"
              fill={getColor(seriesIndex, s.color)}
              class="hover:r-6 transition-all"
            />
          {/each}
        {/each}
      </svg>
    {/if}
  </div>

  <!-- Legend -->
  {#if showLegend && series.length > 0}
    <div class="flex flex-wrap justify-center gap-4 mt-4">
      {#each series as s, i}
        <div class="flex items-center gap-2">
          <div
            class="w-3 h-3 rounded-full"
            style="background-color: {getColor(i, s.color)}"
          ></div>
          <span class="text-sm text-gray-600 dark:text-gray-400">{s.name}</span>
        </div>
      {/each}
    </div>
  {/if}
</div>

