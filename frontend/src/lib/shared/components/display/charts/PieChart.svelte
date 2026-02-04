<script lang="ts">
  import { cn } from '$lib/shared/utils/helpers';

  interface PieSlice {
    label: string;
    value: number;
    color?: string;
  }

  export let data: PieSlice[] = [];
  export let title: string = '';
  export let size: number = 300;
  export let showLegend: boolean = true;
  export let showValues: boolean = true;
  export let donut: boolean = false;
  export let donutWidth: number = 60;
  export let className: string = '';

  $: total = data.reduce((sum, item) => sum + item.value, 0);

  const defaultColors = [
    '#3b82f6',
    '#10b981',
    '#f59e0b',
    '#ef4444',
    '#8b5cf6',
    '#ec4899',
    '#06b6d4',
    '#f97316',
    '#84cc16',
    '#14b8a6',
  ];

  function getColor(index: number, customColor?: string): string {
    if (customColor) return customColor;
    return defaultColors[index % defaultColors.length];
  }

  function generateSlices(): Array<{
    path: string;
    color: string;
    label: string;
    value: number;
    percentage: number;
  }> {
    if (total === 0) return [];

    let currentAngle = -Math.PI / 2; // Start from top
    const radius = size / 2 - 20;
    const centerX = size / 2;
    const centerY = size / 2;

    return data.map((item, index) => {
      const percentage = item.value / total;
      const angle = percentage * 2 * Math.PI;
      const endAngle = currentAngle + angle;

      const x1 = centerX + radius * Math.cos(currentAngle);
      const y1 = centerY + radius * Math.sin(currentAngle);
      const x2 = centerX + radius * Math.cos(endAngle);
      const y2 = centerY + radius * Math.sin(endAngle);

      const largeArcFlag = angle > Math.PI ? 1 : 0;

      const path = donut
        ? generateDonutSlice(centerX, centerY, radius, currentAngle, endAngle, largeArcFlag)
        : `M ${centerX} ${centerY} L ${x1} ${y1} A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2} Z`;

      currentAngle = endAngle;

      return {
        path,
        color: getColor(index, item.color),
        label: item.label,
        value: item.value,
        percentage: percentage * 100,
      };
    });
  }

  function generateDonutSlice(
    cx: number,
    cy: number,
    outerRadius: number,
    startAngle: number,
    endAngle: number,
    largeArcFlag: number
  ): string {
    const innerRadius = outerRadius - donutWidth;

    const x1 = cx + outerRadius * Math.cos(startAngle);
    const y1 = cy + outerRadius * Math.sin(startAngle);
    const x2 = cx + outerRadius * Math.cos(endAngle);
    const y2 = cy + outerRadius * Math.sin(endAngle);

    const x3 = cx + innerRadius * Math.cos(endAngle);
    const y3 = cy + innerRadius * Math.sin(endAngle);
    const x4 = cx + innerRadius * Math.cos(startAngle);
    const y4 = cy + innerRadius * Math.sin(startAngle);

    return [
      `M ${x1} ${y1}`,
      `A ${outerRadius} ${outerRadius} 0 ${largeArcFlag} 1 ${x2} ${y2}`,
      `L ${x3} ${y3}`,
      `A ${innerRadius} ${innerRadius} 0 ${largeArcFlag} 0 ${x4} ${y4}`,
      'Z',
    ].join(' ');
  }

  $: slices = generateSlices();
  $: centerText = donut ? `${Math.round(total).toLocaleString()}` : '';
</script>

<div class={cn('bg-white dark:bg-gray-800 rounded-lg shadow p-6', className)}>
  {#if title}
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">{title}</h3>
  {/if}

  <div class="flex flex-col items-center">
    <div class="relative" style="width: {size}px; height: {size}px;">
      {#if data.length === 0 || total === 0}
        <div class="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
          No data available
        </div>
      {:else}
        <svg
          width={size}
          height={size}
          viewBox="0 0 {size} {size}"
          class="transform transition-transform hover:scale-105"
        >
          {#each slices as slice}
            <path
              d={slice.path}
              fill={slice.color}
              stroke="white"
              stroke-width="2"
              class="hover:opacity-80 transition-opacity cursor-pointer"
            />
          {/each}

          {#if donut}
            <text
              x={size / 2}
              y={size / 2}
              text-anchor="middle"
              dominant-baseline="middle"
              class="text-2xl font-bold fill-gray-900 dark:fill-white"
            >
              {centerText}
            </text>
          {/if}
        </svg>
      {/if}
    </div>

    <!-- Legend -->
    {#if showLegend && slices.length > 0}
      <div class="flex flex-wrap justify-center gap-4 mt-6">
        {#each slices as slice}
          <div class="flex items-center gap-2">
            <div
              class="w-3 h-3 rounded-full"
              style="background-color: {slice.color}"
            ></div>
            <span class="text-sm text-gray-600 dark:text-gray-400">
              {slice.label}
              {#if showValues}
                <span class="text-gray-400 dark:text-gray-500">
                  ({slice.percentage.toFixed(1)}%)
                </span>
              {/if}
            </span>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>
