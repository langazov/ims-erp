<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { fly } from 'svelte/transition';
  import { clickOutside } from '$lib/shared/utils/helpers';

  interface HelpTopic {
    id: string;
    title: string;
    icon: string;
    description: string;
  }

  const helpTopics: HelpTopic[] = [
    {
      id: 'getting-started',
      title: 'Getting Started',
      icon: '🚀',
      description: 'Learn the basics of using the ERP system',
    },
    {
      id: 'clients',
      title: 'Managing Clients',
      icon: '👥',
      description: 'How to add, edit, and manage client information',
    },
    {
      id: 'orders',
      title: 'Orders & Invoices',
      icon: '📄',
      description: 'Create and manage orders and invoices',
    },
    {
      id: 'inventory',
      title: 'Inventory Management',
      icon: '📦',
      description: 'Track stock levels and warehouse operations',
    },
    {
      id: 'reports',
      title: 'Reports & Analytics',
      icon: '📊',
      description: 'Generate reports and view analytics',
    },
    {
      id: 'keyboard',
      title: 'Keyboard Shortcuts',
      icon: '⌨️',
      description: 'Quick commands to speed up your workflow',
    },
  ];

  const dispatch = createEventDispatcher();

  let isOpen = false;
  let searchQuery = '';

  $: filteredTopics = helpTopics.filter(
    (topic) =>
      topic.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      topic.description.toLowerCase().includes(searchQuery.toLowerCase())
  );

  function toggleMenu() {
    isOpen = !isOpen;
    if (isOpen) {
      searchQuery = '';
    }
  }

  function closeMenu() {
    isOpen = false;
  }

  function openDocumentation(topicId?: string) {
    closeMenu();
    const baseUrl = '/docs';
    const url = topicId ? `${baseUrl}/${topicId}` : baseUrl;
    window.open(url, '_blank');
  }

  function openSupport() {
    closeMenu();
    window.open('/support', '_blank');
  }
</script>

<div class="help-button-container" use:clickOutside={closeMenu}>
  <button
    class="help-button"
    on:click={toggleMenu}
    aria-label="Help"
    aria-expanded={isOpen}
    title="Help & Documentation"
  >
    <svg class="help-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>
  </button>

  {#if isOpen}
    <div class="help-dropdown" transition:fly={{ y: 10, duration: 200 }}>
      <div class="dropdown-header">
        <h3 class="header-title">Help Center</h3>
        <p class="header-subtitle">Find answers and get support</p>
      </div>

      <div class="search-box">
        <svg class="search-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
        <input
          type="text"
          class="search-input"
          placeholder="Search help topics..."
          bind:value={searchQuery}
        />
      </div>

      <div class="topics-list">
        {#if filteredTopics.length === 0}
          <div class="no-results">
            <p>No topics found matching "{searchQuery}"</p>
            <button class="view-all-btn" on:click={() => openDocumentation()}>
              View all documentation
            </button>
          </div>
        {:else}
          {#each filteredTopics as topic}
            <button class="topic-item" on:click={() => openDocumentation(topic.id)}>
              <span class="topic-icon">{topic.icon}</span>
              <div class="topic-content">
                <span class="topic-title">{topic.title}</span>
                <span class="topic-description">{topic.description}</span>
              </div>
              <svg class="topic-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 5l7 7-7 7"
                />
              </svg>
            </button>
          {/each}
        {/if}
      </div>

      <div class="dropdown-footer">
        <button class="footer-btn primary" on:click={() => openDocumentation()}>
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
            />
          </svg>
          Full Documentation
        </button>
        <button class="footer-btn" on:click={openSupport}>
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z"
            />
          </svg>
          Contact Support
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .help-button-container {
    position: relative;
  }

  .help-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2.25rem;
    height: 2.25rem;
    padding: 0;
    border: none;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 0.5rem;
    cursor: pointer;
    color: rgba(255, 255, 255, 0.7);
    transition: all 0.2s ease;
  }

  .help-button:hover {
    background: rgba(255, 255, 255, 0.15);
    color: white;
  }

  .help-icon {
    width: 1.25rem;
    height: 1.25rem;
  }

  .help-dropdown {
    position: absolute;
    top: calc(100% + 0.5rem);
    right: 0;
    width: 320px;
    background: white;
    border-radius: 0.75rem;
    box-shadow:
      0 10px 15px -3px rgba(0, 0, 0, 0.1),
      0 4px 6px -2px rgba(0, 0, 0, 0.05);
    border: 1px solid #e5e7eb;
    overflow: hidden;
    z-index: 1000;
  }

  .dropdown-header {
    padding: 1rem;
    background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
    color: white;
  }

  .header-title {
    font-size: 1rem;
    font-weight: 600;
    margin: 0;
  }

  .header-subtitle {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.7);
    margin: 0.25rem 0 0 0;
  }

  .search-box {
    position: relative;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .search-icon {
    position: absolute;
    left: 1.5rem;
    top: 50%;
    transform: translateY(-50%);
    width: 1rem;
    height: 1rem;
    color: #9ca3af;
  }

  .search-input {
    width: 100%;
    padding: 0.5rem 0.75rem 0.5rem 2rem;
    border: 1px solid #d1d5db;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    background: #f9fafb;
    transition: all 0.2s ease;
  }

  .search-input:focus {
    outline: none;
    border-color: #3b82f6;
    background: white;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .search-input::placeholder {
    color: #9ca3af;
  }

  .topics-list {
    max-height: 280px;
    overflow-y: auto;
    padding: 0.5rem;
  }

  .no-results {
    padding: 1.5rem;
    text-align: center;
  }

  .no-results p {
    font-size: 0.875rem;
    color: #6b7280;
    margin: 0 0 0.75rem 0;
  }

  .view-all-btn {
    font-size: 0.875rem;
    font-weight: 500;
    color: #3b82f6;
    background: none;
    border: none;
    cursor: pointer;
    text-decoration: underline;
  }

  .view-all-btn:hover {
    color: #2563eb;
  }

  .topic-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    width: 100%;
    padding: 0.75rem;
    border: none;
    background: transparent;
    border-radius: 0.5rem;
    cursor: pointer;
    text-align: left;
    transition: all 0.2s ease;
  }

  .topic-item:hover {
    background: #f3f4f6;
  }

  .topic-icon {
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #eff6ff;
    border-radius: 0.5rem;
    font-size: 1rem;
    flex-shrink: 0;
  }

  .topic-content {
    flex: 1;
    min-width: 0;
  }

  .topic-title {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: #111827;
  }

  .topic-description {
    display: block;
    font-size: 0.75rem;
    color: #6b7280;
    margin-top: 0.125rem;
  }

  .topic-arrow {
    width: 1rem;
    height: 1rem;
    color: #d1d5db;
    flex-shrink: 0;
  }

  .topic-item:hover .topic-arrow {
    color: #9ca3af;
  }

  .dropdown-footer {
    display: flex;
    gap: 0.5rem;
    padding: 0.75rem;
    border-top: 1px solid #e5e7eb;
    background: #f9fafb;
  }

  .footer-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    flex: 1;
    padding: 0.5rem 0.75rem;
    border: 1px solid #d1d5db;
    background: white;
    border-radius: 0.5rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: #374151;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .footer-btn:hover {
    background: #f3f4f6;
    border-color: #9ca3af;
  }

  .footer-btn.primary {
    background: #3b82f6;
    border-color: #3b82f6;
    color: white;
  }

  .footer-btn.primary:hover {
    background: #2563eb;
    border-color: #2563eb;
  }

  .footer-btn svg {
    width: 0.875rem;
    height: 0.875rem;
  }

  @media (max-width: 640px) {
    .help-dropdown {
      position: fixed;
      top: auto;
      bottom: 4rem;
      right: 1rem;
      left: 1rem;
      width: auto;
    }
  }
</style>
