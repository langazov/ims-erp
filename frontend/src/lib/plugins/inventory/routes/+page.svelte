<script lang="ts">
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
</script>

<svelte:head>
  <title>Inventory | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <h1 class="page-title">Inventory</h1>
    <p class="page-description">Stock and warehouse inventory management</p>
  </div>

  <Card>
    <div class="content">
      <div class="icon">📦</div>
      <h2>Inventory Module</h2>
      <p>This module is under development.</p>
      <Badge variant="yellow">Coming Soon</Badge>
    </div>
  </Card>
</div>

<style>
  .page-container { padding: 1.5rem; max-width: 800px; margin: 0 auto; }
  .page-header { margin-bottom: 1.5rem; }
  .page-title { font-size: 1.875rem; font-weight: 700; color: var(--color-gray-900); margin: 0; }
  .page-description { color: var(--color-gray-500); margin-top: 0.25rem; }
  .content { text-align: center; padding: 3rem; }
  .icon { font-size: 3rem; margin-bottom: 1rem; }
  h2 { font-size: 1.25rem; color: var(--color-gray-900); margin: 0 0 0.5rem 0; }
  p { color: var(--color-gray-500); margin: 0 0 1rem 0; }
</style>
