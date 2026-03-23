<script lang="ts">
  export let variant: 'text' | 'rect' | 'circle' = 'text';
  export let width: string = '100%';
  export let height: string = variant === 'text' ? '1rem' : '3rem';
  export let lines: number = 1;
  export let class_: string = '';
</script>

{#if lines > 1}
  <div class="space-y-2 {class_}">
    {#each Array(lines) as _, i}
      <div
        class="animate-pulse bg-gray-200 dark:bg-gray-700 rounded"
        style="width: {i === lines - 1 ? '75%' : width}; height: {height};"
        class:rounded-full={variant === 'circle'}
      ></div>
    {/each}
  </div>
{:else}
  <div
    class="animate-pulse bg-gray-200 dark:bg-gray-700 rounded {class_}"
    style="width: {width}; height: {height};"
    class:rounded-full={variant === 'circle'}
  ></div>
{/if}
