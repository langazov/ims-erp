<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { fly, fade } from 'svelte/transition';
  import { clickOutside } from '$lib/shared/utils/helpers';
  import Avatar from '$lib/shared/components/display/Avatar.svelte';
  import { auth } from '$lib/shared/stores/auth';
  import type { User } from '$lib/shared/api/auth';
  import { goto } from '$app/navigation';

  export let collapsed = false;

  const dispatch = createEventDispatcher();

  let isOpen = false;
  let user: User | null = null;
  let unsubscribe: (() => void) | null = null;

  onMount(() => {
    unsubscribe = auth.subscribe((state) => {
      user = state.user;
    });
  });

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });

  function toggleDropdown() {
    isOpen = !isOpen;
  }

  function closeDropdown() {
    isOpen = false;
  }

  function handleProfile() {
    closeDropdown();
    goto('/users/profile');
  }

  function handleSettings() {
    closeDropdown();
    goto('/settings');
  }

  function handleLogout() {
    closeDropdown();
    auth.logout();
  }

  function getUserDisplayName(): string {
    if (!user) return 'User';
    return `${user.firstName} ${user.lastName}`.trim() || user.email;
  }

  function getUserRole(): string {
    if (!user) return '';
    return user.role || user.tenantRole || 'User';
  }

  function getInitials(): string {
    if (!user) return 'U';
    const name = `${user.firstName} ${user.lastName}`.trim();
    return name || user.email;
  }
</script>

<div class="user-profile-dropdown" use:clickOutside={closeDropdown}>
  <button
    class="profile-trigger"
    class:collapsed
    on:click={toggleDropdown}
    aria-expanded={isOpen}
    aria-haspopup="true"
  >
    <Avatar
      fallback={getInitials()}
      size={collapsed ? 'sm' : 'md'}
      shape="circle"
    />
    {#if !collapsed}
      <div class="user-info">
        <span class="user-name">{getUserDisplayName()}</span>
        <span class="user-role">{getUserRole()}</span>
      </div>
      <svg
        class="dropdown-arrow"
        class:open={isOpen}
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M19 9l-7 7-7-7"
        />
      </svg>
    {/if}
  </button>

  {#if isOpen}
    <div
      class="dropdown-menu"
      class:collapsed
      transition:fly={{ y: -10, duration: 200 }}
    >
      <div class="dropdown-header">
        <Avatar fallback={getInitials()} size="lg" shape="circle" />
        <div class="header-info">
          <span class="header-name">{getUserDisplayName()}</span>
          <span class="header-email">{user?.email || ''}</span>
          <span class="header-role">{getUserRole()}</span>
        </div>
      </div>

      <div class="dropdown-divider"></div>

      <div class="dropdown-section">
        <span class="section-label">Account</span>
        <button class="dropdown-item" on:click={handleProfile}>
          <svg class="item-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
            />
          </svg>
          <span class="item-label">Profile</span>
        </button>
        <button class="dropdown-item" on:click={handleSettings}>
          <svg class="item-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
            />
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
            />
          </svg>
          <span class="item-label">Settings</span>
        </button>
      </div>

      <div class="dropdown-divider"></div>

      <div class="dropdown-section">
        <button class="dropdown-item logout" on:click={handleLogout}>
          <svg class="item-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
            />
          </svg>
          <span class="item-label">Sign Out</span>
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .user-profile-dropdown {
    position: relative;
  }

  .profile-trigger {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    width: 100%;
    padding: 0.75rem;
    border: none;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.2s ease;
    color: white;
  }

  .profile-trigger:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .profile-trigger.collapsed {
    justify-content: center;
    padding: 0.5rem;
  }

  .user-info {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    text-align: left;
  }

  .user-name {
    font-size: 0.875rem;
    font-weight: 500;
    color: white;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .user-role {
    font-size: 0.75rem;
    color: rgba(255, 255, 255, 0.6);
    text-transform: capitalize;
  }

  .dropdown-arrow {
    width: 1rem;
    height: 1rem;
    color: rgba(255, 255, 255, 0.6);
    transition: transform 0.2s ease;
    flex-shrink: 0;
  }

  .dropdown-arrow.open {
    transform: rotate(180deg);
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
    min-width: 240px;
  }

  .dropdown-menu.collapsed {
    left: 0;
    right: auto;
    min-width: 280px;
  }

  .dropdown-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem;
    background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
  }

  .header-info {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .header-name {
    font-size: 0.875rem;
    font-weight: 600;
    color: white;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .header-email {
    font-size: 0.75rem;
    color: rgba(255, 255, 255, 0.7);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .header-role {
    font-size: 0.75rem;
    color: #60a5fa;
    text-transform: capitalize;
    margin-top: 0.25rem;
  }

  .dropdown-divider {
    height: 1px;
    background: #e5e7eb;
    margin: 0.5rem 0;
  }

  .dropdown-section {
    padding: 0.5rem;
  }

  .section-label {
    display: block;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #6b7280;
    padding: 0.5rem 0.75rem;
  }

  .dropdown-item {
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
    color: #374151;
    text-align: left;
  }

  .dropdown-item:hover {
    background: #f3f4f6;
    color: #111827;
  }

  .dropdown-item.logout {
    color: #dc2626;
  }

  .dropdown-item.logout:hover {
    background: #fef2f2;
    color: #b91c1c;
  }

  .item-icon {
    width: 1.25rem;
    height: 1.25rem;
    flex-shrink: 0;
  }

  .item-label {
    font-size: 0.875rem;
    font-weight: 500;
  }

  @media (max-width: 768px) {
    .user-info,
    .dropdown-arrow {
      display: none;
    }

    .profile-trigger {
      justify-content: center;
      padding: 0.5rem;
    }

    .dropdown-menu {
      left: 0;
      right: auto;
      min-width: 280px;
    }
  }
</style>
