<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { fly, fade } from 'svelte/transition';
  import { clickOutside } from '$lib/shared/utils/helpers';
  import { formatDistanceToNow } from 'date-fns';

  export let collapsed = false;

  interface Notification {
    id: string;
    type: 'info' | 'success' | 'warning' | 'error';
    title: string;
    message: string;
    timestamp: Date;
    read: boolean;
    action?: {
      label: string;
      href: string;
    };
  }

  const dispatch = createEventDispatcher();

  let isOpen = false;
  let notifications: Notification[] = [];
  let activeTab: 'all' | 'unread' = 'all';

  // Demo notifications for initial state
  const demoNotifications: Notification[] = [
    {
      id: '1',
      type: 'success',
      title: 'Order Completed',
      message: 'Order #1234 has been successfully processed and shipped.',
      timestamp: new Date(Date.now() - 1000 * 60 * 5),
      read: false,
      action: { label: 'View Order', href: '/orders/1234' },
    },
    {
      id: '2',
      type: 'warning',
      title: 'Low Stock Alert',
      message: 'Product "Wireless Mouse" is running low on stock (5 units remaining).',
      timestamp: new Date(Date.now() - 1000 * 60 * 30),
      read: false,
      action: { label: 'View Product', href: '/products/wireless-mouse' },
    },
    {
      id: '3',
      type: 'info',
      title: 'New Client Registered',
      message: 'Acme Corporation has been added as a new client.',
      timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2),
      read: true,
    },
    {
      id: '4',
      type: 'error',
      title: 'Payment Failed',
      message: 'Invoice #567 payment processing failed. Please review.',
      timestamp: new Date(Date.now() - 1000 * 60 * 60 * 4),
      read: true,
      action: { label: 'View Invoice', href: '/invoices/567' },
    },
    {
      id: '5',
      type: 'success',
      title: 'Backup Completed',
      message: 'Daily system backup completed successfully.',
      timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24),
      read: true,
    },
  ];

  onMount(() => {
    notifications = demoNotifications;
  });

  $: unreadCount = notifications.filter((n) => !n.read).length;
  $: filteredNotifications = activeTab === 'unread' 
    ? notifications.filter((n) => !n.read) 
    : notifications;

  function toggleDropdown() {
    isOpen = !isOpen;
  }

  function closeDropdown() {
    isOpen = false;
  }

  function markAsRead(id: string) {
    notifications = notifications.map((n) =>
      n.id === id ? { ...n, read: true } : n
    );
  }

  function markAllAsRead() {
    notifications = notifications.map((n) => ({ ...n, read: true }));
  }

  function deleteNotification(id: string) {
    notifications = notifications.filter((n) => n.id !== id);
  }

  function clearAll() {
    notifications = [];
  }

  function getIconForType(type: Notification['type']): string {
    const icons = {
      success: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
      error: 'M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
      warning: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
      info: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    };
    return icons[type];
  }

  function getColorForType(type: Notification['type']): string {
    const colors = {
      success: 'text-green-500 bg-green-50',
      error: 'text-red-500 bg-red-50',
      warning: 'text-yellow-500 bg-yellow-50',
      info: 'text-blue-500 bg-blue-50',
    };
    return colors[type];
  }

  function formatTime(date: Date): string {
    return formatDistanceToNow(date, { addSuffix: true });
  }
</script>

<div class="notification-center" use:clickOutside={closeDropdown}>
  <button
    class="notification-trigger"
    class:collapsed
    on:click={toggleDropdown}
    aria-expanded={isOpen}
    aria-haspopup="true"
    aria-label="Notifications"
  >
    <div class="icon-wrapper">
      <svg class="bell-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
        />
      </svg>
      {#if unreadCount > 0}
        <span class="notification-badge">
          {unreadCount > 99 ? '99+' : unreadCount}
        </span>
      {/if}
    </div>
    {#if !collapsed}
      <span class="trigger-label">Notifications</span>
    {/if}
  </button>

  {#if isOpen}
    <div
      class="dropdown-menu"
      class:collapsed
      transition:fly={{ y: -10, duration: 200 }}
    >
      <div class="dropdown-header">
        <h3 class="header-title">Notifications</h3>
        <div class="header-actions">
          {#if unreadCount > 0}
            <button class="action-btn" on:click={markAllAsRead} title="Mark all as read">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </button>
          {/if}
          <button class="action-btn" on:click={clearAll} title="Clear all">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
          </button>
        </div>
      </div>

      <div class="tabs">
        <button
          class="tab"
          class:active={activeTab === 'all'}
          on:click={() => (activeTab = 'all')}
        >
          All
          <span class="tab-count">{notifications.length}</span>
        </button>
        <button
          class="tab"
          class:active={activeTab === 'unread'}
          on:click={() => (activeTab = 'unread')}
        >
          Unread
          <span class="tab-count">{unreadCount}</span>
        </button>
      </div>

      <div class="notifications-list">
        {#if filteredNotifications.length === 0}
          <div class="empty-state">
            <svg class="empty-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
              />
            </svg>
            <p class="empty-text">
              {activeTab === 'unread' ? 'No unread notifications' : 'No notifications'}
            </p>
          </div>
        {:else}
          {#each filteredNotifications as notification (notification.id)}
            <div
              class="notification-item"
              class:unread={!notification.read}
              transition:fly={{ y: -10, duration: 200 }}
            >
              <div class="notification-icon {getColorForType(notification.type)}">
                <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d={getIconForType(notification.type)}
                  />
                </svg>
              </div>
              <div class="notification-content">
                <div class="notification-header">
                  <h4 class="notification-title">{notification.title}</h4>
                  <span class="notification-time">{formatTime(notification.timestamp)}</span>
                </div>
                <p class="notification-message">{notification.message}</p>
                {#if notification.action}
                  <a
                    href={notification.action.href}
                    class="notification-action"
                    on:click={() => markAsRead(notification.id)}
                  >
                    {notification.action.label} →
                  </a>
                {/if}
              </div>
              <div class="notification-actions">
                {#if !notification.read}
                  <button
                    class="action-btn small"
                    on:click={() => markAsRead(notification.id)}
                    title="Mark as read"
                  >
                    <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M5 13l4 4L19 7"
                      />
                    </svg>
                  </button>
                {/if}
                <button
                  class="action-btn small"
                  on:click={() => deleteNotification(notification.id)}
                  title="Delete"
                >
                  <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>
            </div>
          {/each}
        {/if}
      </div>

      <div class="dropdown-footer">
        <a href="/notifications" class="view-all-link" on:click={closeDropdown}>
          View all notifications
        </a>
      </div>
    </div>
  {/if}
</div>

<style>
  .notification-center {
    position: relative;
  }

  .notification-trigger {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    width: 100%;
    padding: 0.625rem 0.75rem;
    border: none;
    background: transparent;
    border-radius: 0.375rem;
    cursor: pointer;
    transition: all 0.2s ease;
    color: rgba(255, 255, 255, 0.7);
    text-align: left;
  }

  .notification-trigger:hover {
    background: rgba(255, 255, 255, 0.1);
    color: white;
  }

  .notification-trigger.collapsed {
    justify-content: center;
    padding: 0.625rem;
  }

  .icon-wrapper {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .bell-icon {
    width: 1.25rem;
    height: 1.25rem;
  }

  .notification-badge {
    position: absolute;
    top: -6px;
    right: -6px;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    background: #ef4444;
    color: white;
    font-size: 0.6875rem;
    font-weight: 600;
    border-radius: 9px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid #1e293b;
  }

  .trigger-label {
    font-size: 0.8125rem;
    font-weight: 500;
  }

  .dropdown-menu {
    position: absolute;
    bottom: 100%;
    left: 0;
    right: 0;
    margin-bottom: 0.5rem;
    background: white;
    border-radius: 0.75rem;
    box-shadow:
      0 10px 15px -3px rgba(0, 0, 0, 0.1),
      0 4px 6px -2px rgba(0, 0, 0, 0.05);
    border: 1px solid #e5e7eb;
    overflow: hidden;
    z-index: 1000;
    min-width: 360px;
    max-height: 500px;
    display: flex;
    flex-direction: column;
  }

  .dropdown-menu.collapsed {
    left: 0;
    right: auto;
    min-width: 380px;
  }

  .dropdown-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .header-title {
    font-size: 1rem;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .header-actions {
    display: flex;
    gap: 0.25rem;
  }

  .action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    padding: 0;
    border: none;
    background: transparent;
    border-radius: 0.375rem;
    cursor: pointer;
    color: #6b7280;
    transition: all 0.2s ease;
  }

  .action-btn:hover {
    background: #f3f4f6;
    color: #374151;
  }

  .action-btn svg {
    width: 1.25rem;
    height: 1.25rem;
  }

  .action-btn.small {
    width: 1.5rem;
    height: 1.5rem;
  }

  .action-btn.small svg {
    width: 1rem;
    height: 1rem;
  }

  .tabs {
    display: flex;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #e5e7eb;
    background: #f9fafb;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.375rem 0.75rem;
    border: none;
    background: transparent;
    border-radius: 9999px;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    color: #6b7280;
    transition: all 0.2s ease;
  }

  .tab:hover {
    background: #e5e7eb;
    color: #374151;
  }

  .tab.active {
    background: #3b82f6;
    color: white;
  }

  .tab-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.25rem;
    height: 1.25rem;
    padding: 0 0.375rem;
    background: currentColor;
    color: inherit;
    font-size: 0.75rem;
    font-weight: 600;
    border-radius: 9999px;
    opacity: 0.8;
  }

  .notifications-list {
    flex: 1;
    overflow-y: auto;
    max-height: 320px;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem 1rem;
    text-align: center;
  }

  .empty-icon {
    width: 3rem;
    height: 3rem;
    color: #d1d5db;
    margin-bottom: 0.75rem;
  }

  .empty-text {
    font-size: 0.875rem;
    color: #6b7280;
    margin: 0;
  }

  .notification-item {
    display: flex;
    gap: 0.75rem;
    padding: 1rem;
    border-bottom: 1px solid #f3f4f6;
    transition: background 0.2s ease;
  }

  .notification-item:hover {
    background: #f9fafb;
  }

  .notification-item.unread {
    background: #eff6ff;
  }

  .notification-item.unread:hover {
    background: #dbeafe;
  }

  .notification-icon {
    flex-shrink: 0;
    width: 2rem;
    height: 2rem;
    border-radius: 0.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .notification-icon svg {
    width: 1.25rem;
    height: 1.25rem;
  }

  .notification-content {
    flex: 1;
    min-width: 0;
  }

  .notification-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
  }

  .notification-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .notification-time {
    font-size: 0.75rem;
    color: #9ca3af;
    flex-shrink: 0;
  }

  .notification-message {
    font-size: 0.8125rem;
    color: #6b7280;
    margin: 0 0 0.5rem 0;
    line-height: 1.4;
  }

  .notification-action {
    font-size: 0.8125rem;
    font-weight: 500;
    color: #3b82f6;
    text-decoration: none;
  }

  .notification-action:hover {
    text-decoration: underline;
  }

  .notification-actions {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    opacity: 0;
    transition: opacity 0.2s ease;
  }

  .notification-item:hover .notification-actions {
    opacity: 1;
  }

  .dropdown-footer {
    padding: 0.75rem 1rem;
    border-top: 1px solid #e5e7eb;
    background: #f9fafb;
    text-align: center;
  }

  .view-all-link {
    font-size: 0.875rem;
    font-weight: 500;
    color: #3b82f6;
    text-decoration: none;
  }

  .view-all-link:hover {
    text-decoration: underline;
  }

  @media (max-width: 768px) {
    .trigger-label {
      display: none;
    }

    .notification-trigger {
      justify-content: center;
      padding: 0.625rem;
    }

    .dropdown-menu {
      left: 0;
      right: auto;
      min-width: 360px;
    }
  }
</style>
