<script lang="ts">
  export let title: string;
  export let description: string = '';
  export let maxWidth: string = '1400px';
</script>

<div class="page-shell" style="max-width: {maxWidth};">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">{title}</h1>
      {#if description}
        <p class="page-description">{description}</p>
      {/if}
    </div>
    <div class="header-actions">
      <slot name="actions" />
    </div>
  </div>
  <slot />
</div>

<style>
  .page-shell {
    padding: 1.5rem;
    margin: 0 auto;
    width: 100%;
  }
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
    gap: 1rem;
    flex-wrap: wrap;
  }
  .page-title {
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin: 0;
    line-height: 1.25;
  }
  .page-description {
    color: var(--color-gray-500);
    margin: 0.25rem 0 0;
    font-size: 0.875rem;
  }
  .header-actions {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    flex-shrink: 0;
  }
  :global(.dark) .page-title {
    color: var(--color-gray-100);
  }
</style>
