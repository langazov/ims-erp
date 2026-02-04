<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';
  import { fade, fly } from 'svelte/transition';
  import { clickOutside } from '$lib/shared/utils/helpers';

  export let isOpen = false;
  export let placeholder = 'Search anything...';

  const dispatch = createEventDispatcher();

  interface SearchResult {
    id: string;
    type: 'client' | 'product' | 'order' | 'invoice' | 'user' | 'document';
    title: string;
    subtitle?: string;
    url: string;
    icon?: string;
  }

  let searchQuery = '';
  let isLoading = false;
  let results: SearchResult[] = [];
  let selectedIndex = 0;
  let inputRef: HTMLInputElement;

  // Mock search data
  const mockData: SearchResult[] = [
    { id: '1', type: 'client', title: 'Acme Corporation', subtitle: 'Client', url: '/clients/1', icon: '👤' },
    { id: '2', type: 'client', title: 'TechStart Inc', subtitle: 'Client', url: '/clients/2', icon: '👤' },
    { id: '3', type: 'product', title: 'Wireless Bluetooth Headphones', subtitle: 'Product - PROD-001', url: '/products/1', icon: '📦' },
    { id: '4', type: 'product', title: 'Cotton T-Shirt', subtitle: 'Product - PROD-002', url: '/products/2', icon: '📦' },
    { id: '5', type: 'order', title: 'Order #ORD-2024-001', subtitle: 'Acme Corporation - $12,500.00', url: '/orders/1', icon: '🛒' },
    { id: '6', type: 'order', title: 'Order #ORD-2024-002', subtitle: 'TechStart Inc - $8,750.00', url: '/orders/2', icon: '🛒' },
    { id: '7', type: 'invoice', title: 'Invoice #INV-2024-001', subtitle: 'Acme Corporation - $5,400.00', url: '/invoices/1', icon: '📄' },
    { id: '8', type: 'invoice', title: 'Invoice #INV-2024-002', subtitle: 'TechStart Inc - $3,780.00', url: '/invoices/2', icon: '📄' },
    { id: '9', type: 'user', title: 'John Doe', subtitle: 'Admin - john.doe@example.com', url: '/users/1', icon: '👨‍💼' },
    { id: '10', type: 'user', title: 'Jane Smith', subtitle: 'Manager - jane.smith@example.com', url: '/users/2', icon: '👩‍💼' },
    { id: '11', type: 'document', title: 'Invoice_2024_001.pdf', subtitle: 'Document - Invoice', url: '/documents/1', icon: '📎' },
    { id: '12', type: 'document', title: 'Purchase_Order_12345.pdf', subtitle: 'Document - PO', url: '/documents/2', icon: '📎' }
  ];

  function performSearch() {
    if (!searchQuery.trim()) {
      results = [];
      return;
    }

    isLoading = true;
    
    // Simulate API delay
    setTimeout(() => {
      const query = searchQuery.toLowerCase();
      results = mockData.filter(item => 
        item.title.toLowerCase().includes(query) ||
        (item.subtitle && item.subtitle.toLowerCase().includes(query))
      );
      selectedIndex = 0;
      isLoading = false;
    }, 200);
  }

  function handleKeydown(event: KeyboardEvent) {
    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        selectedIndex = Math.min(selectedIndex + 1, results.length - 1);
        break;
      case 'ArrowUp':
        event.preventDefault();
        selectedIndex = Math.max(selectedIndex - 1, 0);
        break;
      case 'Enter':
        event.preventDefault();
        if (results[selectedIndex]) {
          selectResult(results[selectedIndex]);
        }
        break;
      case 'Escape':
        event.preventDefault();
        close();
        break;
    }
  }

  function selectResult(result: SearchResult) {
    window.location.href = result.url;
    close();
  }

  function close() {
    isOpen = false;
    searchQuery = '';
    results = [];
    dispatch('close');
  }

  function getTypeColor(type: string): string {
    const colors: Record<string, string> = {
      client: 'bg-blue-100 text-blue-800',
      product: 'bg-green-100 text-green-800',
      order: 'bg-purple-100 text-purple-800',
      invoice: 'bg-yellow-100 text-yellow-800',
      user: 'bg-pink-100 text-pink-800',
      document: 'bg-gray-100 text-gray-800'
    };
    return colors[type] || 'bg-gray-100 text-gray-800';
  }

  $: if (isOpen && inputRef) {
    setTimeout(() => inputRef.focus(), 100);
  }
</script>

{#if isOpen}
  <div
    class="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]"
    transition:fade={{ duration: 200 }}
  >
    <!-- Backdrop -->
    <div 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={close}
    />

    <!-- Search Modal -->
    <div
      class="relative w-full max-w-2xl mx-4 bg-white dark:bg-gray-900 rounded-xl shadow-2xl overflow-hidden"
      transition:fly={{ y: -20, duration: 300 }}
      use:clickOutside
      on:click_outside={close}
    >
      <!-- Search Input -->
      <div class="flex items-center border-b border-gray-200 dark:border-gray-700 px-4">
        <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        
        <input
          bind:this={inputRef}
          type="text"
          class="flex-1 px-4 py-4 text-lg bg-transparent border-none outline-none text-gray-900 dark:text-gray-100 placeholder-gray-400"
          {placeholder}
          bind:value={searchQuery}
          on:input={performSearch}
          on:keydown={handleKeydown}
        />
        
        {#if isLoading}
          <svg class="w-5 h-5 text-gray-400 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
        {:else if searchQuery}
          <button
            class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
            on:click={() => { searchQuery = ''; results = []; inputRef.focus(); }}
          >
            <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}
        
        <div class="ml-2 px-2 py-1 text-xs text-gray-400 border border-gray-200 dark:border-gray-700 rounded">
          ESC
        </div>
      </div>

      <!-- Results -->
      <div class="max-h-[60vh] overflow-y-auto">
        {#if results.length > 0}
          <div class="py-2">
            <div class="px-4 py-2 text-xs font-medium text-gray-500 uppercase tracking-wider">
              {results.length} result{results.length === 1 ? '' : 's'}
            </div>
            
            {#each results as result, index}
              <button
                class={cn(
                  'w-full px-4 py-3 flex items-center gap-3 text-left transition-colors',
                  'hover:bg-gray-50 dark:hover:bg-gray-800',
                  index === selectedIndex && 'bg-gray-50 dark:bg-gray-800'
                )}
                on:click={() => selectResult(result)}
                on:mouseenter={() => selectedIndex = index}
              >
                <span class="text-2xl">{result.icon}</span>
                
                <div class="flex-1 min-w-0">
                  <div class="font-medium text-gray-900 dark:text-gray-100 truncate">
                    {result.title}
                  </div>
                  {#if result.subtitle}
                    <div class="text-sm text-gray-500 truncate">
                      {result.subtitle}
                    </div>
                  {/if}
                </div>
                
                <span class={cn(
                  'px-2 py-1 text-xs font-medium rounded-full capitalize',
                  getTypeColor(result.type)
                )}>
                  {result.type}
                </span>
              </button>
            {/each}
          </div>
        {:else if searchQuery && !isLoading}
          <div class="py-12 text-center text-gray-500">
            <svg class="w-12 h-12 mx-auto mb-3 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <p>No results found for "{searchQuery}"</p>
          </div>
        {:else}
          <div class="py-6 px-4">
            <div class="text-xs font-medium text-gray-500 uppercase tracking-wider mb-3">
              Recent searches
            </div>
            <div class="flex flex-wrap gap-2">
              {#each ['Acme', 'Invoice 001', 'Headphones', 'John Doe'] as term}
                <button
                  class="px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-800 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
                  on:click={() => { searchQuery = term; performSearch(); }}
                >
                  {term}
                </button>
              {/each}
            </div>

            <div class="mt-6 text-xs font-medium text-gray-500 uppercase tracking-wider mb-3">
              Keyboard shortcuts
            </div>
            
            <div class="space-y-2 text-sm text-gray-600 dark:text-gray-400">
              <div class="flex justify-between">
                <span>Navigate</span>
                <span class="font-mono">↑ ↓</span>
              </div>
              <div class="flex justify-between">
                <span>Select</span>
                <span class="font-mono">Enter</span>
              </div>
              <div class="flex justify-between">
                <span>Close</span>
                <span class="font-mono">ESC</span>
              </div>
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
