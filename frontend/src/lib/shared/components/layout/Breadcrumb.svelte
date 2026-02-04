<script lang="ts">
  import { page } from '$app/stores';
  import { derived } from 'svelte/store';

  interface BreadcrumbItem {
    label: string;
    href: string;
    isLast: boolean;
  }

  // Route label mappings
  const routeLabels: Record<string, string> = {
    '': 'Home',
    'dashboard': 'Dashboard',
    'clients': 'Clients',
    'users': 'Users',
    'products': 'Products',
    'inventory': 'Inventory',
    'warehouse': 'Warehouse',
    'orders': 'Orders',
    'invoices': 'Invoices',
    'payments': 'Payments',
    'documents': 'Documents',
    'settings': 'Settings',
    'profile': 'Profile',
    'new': 'New',
    'edit': 'Edit',
    'details': 'Details',
    'notifications': 'Notifications',
  };

  function formatLabel(segment: string): string {
    // Check if it's a UUID or ID
    if (/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(segment)) {
      return 'Details';
    }
    // Check if it's a numeric ID
    if (/^\d+$/.test(segment)) {
      return `#${segment}`;
    }
    // Return mapped label or capitalize
    return routeLabels[segment] || segment.charAt(0).toUpperCase() + segment.slice(1);
  }

  const breadcrumbs = derived(page, ($page): BreadcrumbItem[] => {
    const path = $page.url.pathname;
    
    // Skip for root and auth pages
    if (path === '/' || path === '/dashboard' || path.startsWith('/login') || path.startsWith('/register')) {
      return [];
    }

    const segments = path.split('/').filter(Boolean);
    const items: BreadcrumbItem[] = [
      { label: 'Home', href: '/dashboard', isLast: segments.length === 0 }
    ];

    let currentPath = '';
    segments.forEach((segment, index) => {
      currentPath += `/${segment}`;
      items.push({
        label: formatLabel(segment),
        href: currentPath,
        isLast: index === segments.length - 1
      });
    });

    return items;
  });
</script>

{#if $breadcrumbs.length > 0}
  <nav class="breadcrumb" aria-label="Breadcrumb">
    <ol class="breadcrumb-list">
      {#each $breadcrumbs as item, index}
        <li class="breadcrumb-item" class:active={item.isLast}>
          {#if item.isLast}
            <span class="breadcrumb-current" aria-current="page">
              {item.label}
            </span>
          {:else}
            <a href={item.href} class="breadcrumb-link">
              {#if index === 0}
                <svg class="home-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                </svg>
              {:else}
                {item.label}
              {/if}
            </a>
          {/if}
          {#if !item.isLast}
            <svg class="separator" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          {/if}
        </li>
      {/each}
    </ol>
  </nav>
{/if}

<style>
  .breadcrumb {
    padding: 0.75rem 1.5rem;
    background: white;
    border-bottom: 1px solid #e5e7eb;
  }

  .breadcrumb-list {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    list-style: none;
    margin: 0;
    padding: 0;
    flex-wrap: wrap;
  }

  .breadcrumb-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .breadcrumb-link {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #6b7280;
    text-decoration: none;
    transition: color 0.2s ease;
  }

  .breadcrumb-link:hover {
    color: #3b82f6;
  }

  .home-icon {
    width: 1rem;
    height: 1rem;
  }

  .separator {
    width: 1rem;
    height: 1rem;
    color: #d1d5db;
    flex-shrink: 0;
  }

  .breadcrumb-current {
    font-size: 0.875rem;
    font-weight: 600;
    color: #111827;
  }

  @media (max-width: 640px) {
    .breadcrumb {
      padding: 0.5rem 1rem;
    }

    .breadcrumb-list {
      gap: 0.25rem;
    }

    /* Hide middle items on mobile, show only first and last */
    .breadcrumb-item:not(:first-child):not(:last-child):not(.active) {
      display: none;
    }

    .breadcrumb-item.ellipsis::before {
      content: '...';
      color: #9ca3af;
      margin-right: 0.25rem;
    }
  }
</style>
