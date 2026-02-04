<script lang="ts">
  import PieChart from './PieChart.svelte';

  interface DonutSlice {
    label: string;
    value: number;
    color?: string;
  }

  export let data: DonutSlice[] = [];
  export let title: string = '';
  export let size: number = 300;
  export let showLegend: boolean = true;
  export let showValues: boolean = true;
  export let donutWidth: number = 60;
  export let className: string = '';
</script>

<PieChart
  {data}
  {title}
  {size}
  {showLegend}
  {showValues}
  donut={true}
  {donutWidth}
  {className}
/>

