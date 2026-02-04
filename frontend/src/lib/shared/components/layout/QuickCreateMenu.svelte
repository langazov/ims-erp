<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { fly, scale } from 'svelte/transition';
  import { clickOutside } from '$lib/shared/utils/helpers';
  import { goto } from '$app/navigation';

  interface QuickAction {
    id: string;
    label: string;
    icon: string;
    href: string;
    category: 'sales' | 'inventory' | 'crm' | 'finance';
  }

  const quickActions: QuickAction[] = [
    { id: 'client', label: 'New Client', icon: '👤', href: '/clients/new', category: 'crm' },
    { id: 'order', label: 'New Order', icon: '📦', href: '/orders/new', category: 'sales' },
    { id: 'invoice', label: 'New Invoice', icon: '📄', href: '/invoices/new', category: 'finance' },
    { id: 'product', label: 'New Product', icon: '🏷️', href: '/products/new', category: 'inventory' },
    { id: 'payment', label: 'Record Payment', icon: '💳', href: '/payments/new', category: 'finance' },
    { id: 'warehouse', label: 'Warehouse Operation', icon: '🏭', href: '/warehouse/operations/new', category: 'inventory' },
  ];

  const dispatch = createEventDispatcher();

  let isOpen = false;

  function toggleMenu() {
    isOpen = !isOpen;
  }

  function closeMenu() {
    isOpen = false;
  }

  function handleAction(action: QuickAction) {
    closeMenu();
    goto(action.href);
  }

  function getCategoryColor(category: QuickAction['category']): string {
    const colors = {
      sales: 'bg-blue-100 text-blue-700',
      inventory: 'bg-green-100 text-green-700',
      crm: 'bg-purple-100 text-purple-700',
      finance: 'bg-yellow-100 text-yellow-700',
    };
    return colors[category];
  }

  function getCategoryLabel(category: QuickAction['category']): string {
    const labels = {
      sales: 'Sales',
      inventory: 'Inventory',
      crm: 'CRM',
      finance: 'Finance',
    };
    return labels[category];
  }

  // Group actions by category
  $: groupedActions = quickActions.reduce((acc, action) => {
    if (!acc[action.category]) {
      acc[action.category] = [];
    }
    acc[action.category].push(action);
    return acc;
  }, {} as Record<string, QuickAction[]>);

  $: categories = Object.keys(groupedActions) as QuickAction['category'][];
</script>

<div class="quick-create-menu" use:clickOutside={closeMenu}>
  <!-- Main FAB Button -->
  <button
    class="fab-button"
    class:open={isOpen}
    on:click={toggleMenu}
    aria-label="Quick create"
    aria-expanded={isOpen}
  >
    <svg class="fab-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
    </svg>
  </button>

  <!-- Dropdown Menu -->
  {#if isOpen}
    <div
      class="dropdown-menu"
      transition:fly={{ y: 20, duration: 200 }}
    >
      <div class="dropdown-header">
        <h3 class="header-title">Quick Create</h3>
        <p class="header-subtitle">Create new items quickly</p>
      </div>

      <div class="actions-container">
        {#each categories as category}
          <div class="category-section">
            <span class="category-label">{getCategoryLabel(category)}</span>
            <div class="actions-grid">
              {#each groupedActions[category] as action}
                <button
                  class="action-button"
                  on:click={() => handleAction(action)}
                >
                  <span class="action-icon {getCategoryColor(category)}">{action.icon}</span>
                  <span class="action-label">{action.label}</span>
                </button>
              {/each}
            </div>
          </div>
        {/each}
      </div>

      <div class="dropdown-footer">
        <span class="keyboard-hint">Press <kbd>Q</kbd> to open</span>
      </div>
    </div>
  {/if}
</div>

<style>
  .quick-create-menu {
    position: fixed;
    bottom: 6rem;
    right: 1.5rem;
    z-index: 50;
  }

  .fab-button {
    width: 3.5rem;
    height: 3.5rem;
    border-radius: 50%;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    border: none;
    color: white;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06),
      0 0 0 4px rgba(59, 130, 246, 0.2);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .fab-button:hover {
    transform: scale(1.05);
    box-shadow:
      0 10px 15px -3px rgba(0, 0, 0, 0.1),
      0 4px 6px -2px rgba(0, 0, 0, 0.05),
      0 0 0 6px rgba(59, 130, 246, 0.2);
  }

  .fab-button.open {
    transform: rotate(45deg);
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06),
      0 0 0 4px rgba(239, 68, 68, 0.2);
  }

  .fab-button.open:hover {
    box-shadow:
      0 10px 15px -3px rgba(0, 0, 0, 0.1),
      0 4px 6px -2px rgba(0, 0, 0, 0.05),
      0 0 0 6px rgba(239, 68, 68, 0.2);
  }

  .fab-icon {
    width: 1.5rem;
    height: 1.5rem;
    transition: transform 0.3s ease;
  }

  .dropdown-menu {
    position: absolute;
    bottom: calc(100% + 1rem);
    right: 0;
    width: 320px;
    background: white;
    border-radius: 1rem;
    box-shadow:
      0 20px 25px -5px rgba(0, 0, 0, 0.1),
      0 10px 10px -5px rgba(0, 0, 0, 0.04);
    border: 1px solid #e5e7eb;
    overflow: hidden;
  }

  .dropdown-header {
    padding: 1.25rem;
    background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
    color: white;
  }

  .header-title {
    font-size: 1.125rem;
    font-weight: 600;
    margin: 0;
  }

  .header-subtitle {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.7);
    margin: 0.25rem 0 0 0;
  }

  .actions-container {
    padding: 1rem;
    max-height: 400px;
    overflow-y: auto;
  }

  .category-section {
    margin-bottom: 1rem;
  }

  .category-section:last-child {
    margin-bottom: 0;
  }

  .category-label {
    display: block;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #6b7280;
    margin-bottom: 0.5rem;
    padding-left: 0.5rem;
  }

  .actions-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.5rem;
  }

  .action-button {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    padding: 0.875rem 0.5rem;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    background: white;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .action-button:hover {
    background: #f9fafb;
    border-color: #d1d5db;
    transform: translateY(-2px);
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
  }

  .action-icon {
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 0.625rem;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.25rem;
  }

  .action-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: #374151;
    text-align: center;
  }

  .dropdown-footer {
    padding: 0.75rem 1rem;
    background: #f9fafb;
    border-top: 1px solid #e5e7eb;
    text-align: center;
  }

  .keyboard-hint {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .keyboard-hint kbd {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.5rem;
    height: 1.5rem;
    padding: 0 0.375rem;
    background: white;
    border: 1px solid #d1d5db;
    border-radius: 0.25rem;
    font-family: inherit;
    font-size: 0.6875rem;
    font-weight: 600;
    color: #374151;
    box-shadow: 0 1px 0 rgba(0, 0, 0, 0.05);
  }

  @media (max-width: 640px) {
    .quick-create-menu {
      bottom: 5rem;
      right: 1rem;
    }

    .dropdown-menu {
      width: calc(100vw - 2rem);
      right: -0.5rem;
    }

    .actions-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }
</style>
