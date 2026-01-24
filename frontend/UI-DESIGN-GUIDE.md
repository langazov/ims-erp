# ERP System UI Design Guide

## Professional SvelteKit UI Design Best Practices

This guide establishes design standards for creating a clean, professional, and accessible user interface for the ERP system using SvelteKit.

---

## 1. Design Principles

### 1.1 Core Principles
- **Clarity** - Users should immediately understand what they can do
- **Consistency** - Similar elements behave similarly across the app
- **Efficiency** - Common tasks require minimal clicks
- **Feedback** - Users receive immediate response to actions
- **Accessibility** - Interface works for all users, including those with disabilities

### 1.2 Visual Hierarchy
```
┌─────────────────────────────────────────────────────────┐
│                    PRIMARY ACTION                        │  ← Most important
├─────────────────────────────────────────────────────────┤
│  Secondary actions  │  Tertiary actions  │  Low priority │  ← Decreasing emphasis
└─────────────────────────────────────────────────────────┘
```

**Implementation:**
- Use size, color, and position to indicate importance
- Limit primary actions to one per view
- Group related actions together
- Use visual separation for distinct sections

---

## 2. Color System

### 2.1 Primary Color Palette
```css
:root {
  /* Primary - Main brand color */
  --color-primary-50: #eff6ff;  /* Lightest */
  --color-primary-100: #dbeafe;
  --color-primary-200: #bfdbfe;
  --color-primary-300: #93c5fd;
  --color-primary-400: #60a5fa;
  --color-primary-500: #3b82f6; /* Base */
  --color-primary-600: #2563eb;
  --color-primary-700: #1d4ed8;
  --color-primary-800: #1e40af;
  --color-primary-900: #1e3a8a; /* Darkest */

  /* Neutral - Grays for text and backgrounds */
  --color-gray-50: #f9fafb;
  --color-gray-100: #f3f4f6;
  --color-gray-200: #e5e7eb;
  --color-gray-300: #d1d5db;
  --color-gray-400: #9ca3af;
  --color-gray-500: #6b7280;
  --color-gray-600: #4b5563;
  --color-gray-700: #374151;
  --color-gray-800: #1f2937;
  --color-gray-900: #111827;

  /* Semantic colors */
  --color-success: #10b981;    /* Green - Success states */
  --color-warning: #f59e0b;    /* Amber - Warnings */
  --color-error: #ef4444;      /* Red - Errors */
  --color-info: #3b82f6;       /* Blue - Informational */
}
```

### 2.2 Usage Guidelines
```svelte
<!-- Primary action button -->
<button class="bg-primary-600 hover:bg-primary-700 text-white">
  Save Changes
</button>

<!-- Secondary action -->
<button class="bg-gray-100 hover:bg-gray-200 text-gray-700">
  Cancel
</button>

<!-- Destructive action -->
<button class="bg-red-600 hover:bg-red-700 text-white">
  Delete
</button>

<!-- Success message -->
<div class="bg-green-50 text-green-800 border border-green-200">
  Operation completed successfully
</div>

<!-- Error message -->
<div class="bg-red-50 text-red-800 border border-red-200">
  Please correct the errors below
</div>

<!-- Warning message -->
<div class="bg-amber-50 text-amber-800 border border-amber-200">
  Review the highlighted items
</div>
```

### 2.3 Dark Mode Support
```svelte
<script>
  import { browser } from '$app/environment';
  
  let isDark = false;
  
  if (browser) {
    isDark = localStorage.getItem('theme') === 'dark' 
      || (!localStorage.getItem('theme') 
      && window.matchMedia('(prefers-color-scheme: dark)').matches);
  }
  
  function toggleTheme() {
    isDark = !isDark;
    document.documentElement.classList.toggle('dark', isDark);
    localStorage.setItem('theme', isDark ? 'dark' : 'light');
  }
</script>

<button 
  on:click={toggleTheme}
  class="p-2 rounded-lg bg-gray-100 dark:bg-gray-800"
>
  {isDark ? '🌙' : '☀️'}
</button>
```

---

## 3. Typography System

### 3.1 Font Stack
```css
:root {
  /* Primary font - System fonts for best performance */
  --font-sans: system-ui, -apple-system, BlinkMacSystemFont, 
    'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  
  /* Monospace for code and data */
  --font-mono: 'SF Mono', 'Fira Code', 'Fira Mono', 
    Menlo, Monaco, 'Courier New', monospace;
}
```

### 3.2 Type Scale
```css
:root {
  /* Base size: 16px */
  --text-xs: 0.75rem;    /* 12px */
  --text-sm: 0.875rem;   /* 14px */
  --text-base: 1rem;     /* 16px */
  --text-lg: 1.125rem;   /* 18px */
  --text-xl: 1.25rem;    /* 20px */
  --text-2xl: 1.5rem;    /* 24px */
  --text-3xl: 1.875rem;  /* 30px */
  --text-4xl: 2.25rem;   /* 36px */
  
  /* Line heights */
  --leading-tight: 1.25;
  --leading-normal: 1.5;
  --leading-relaxed: 1.75;
  
  /* Font weights */
  --font-normal: 400;
  --font-medium: 500;
  --font-semibold: 600;
  --font-bold: 700;
}
```

### 3.3 Typography Examples
```svelte
<!-- Page title -->
<h1 class="text-3xl font-bold text-gray-900 dark:text-white">
  Client Management
</h1>

<!-- Section heading -->
<h2 class="text-xl font-semibold text-gray-800 dark:text-gray-100">
  Billing Information
</h2>

<!-- Card title -->
<h3 class="text-lg font-medium text-gray-900">
  Order #12345
</h3>

<!-- Body text -->
<p class="text-base text-gray-600 dark:text-gray-300 leading-relaxed">
  The client has been successfully updated.
</p>

<!-- Helper text -->
<p class="text-sm text-gray-500">
  Last updated 2 hours ago
</p>

<!-- Code/data -->
<code class="text-sm font-mono bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded">
  INV-2024-001
</code>
```

### 3.4 Character Limits
```svelte
<script>
  const limits = {
    pageTitle: '60 characters',
    sectionTitle: '40 characters',
    buttonText: '3-4 words',
    tableHeader: '2-3 words',
    inputPlaceholder: '2-6 words',
    successMessage: '1-2 sentences',
    errorMessage: 'Specific, actionable',
    tooltip: '1 short sentence'
  };
</script>
```

---

## 4. Spacing System

### 4.1 Spacing Scale
```css
:root {
  --space-0: 0;
  --space-1: 0.25rem;   /* 4px */
  --space-2: 0.5rem;    /* 8px */
  --space-3: 0.75rem;   /* 12px */
  --space-4: 1rem;      /* 16px */
  --space-5: 1.25rem;   /* 20px */
  --space-6: 1.5rem;    /* 24px */
  --space-8: 2rem;      /* 32px */
  --space-10: 2.5rem;   /* 40px */
  --space-12: 3rem;     /* 48px */
  --space-16: 4rem;     /* 64px */
  --space-20: 5rem;     /* 80px */
}
```

### 4.2 Spacing Guidelines

#### Container Spacing
```svelte
<!-- Page container -->
<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
  <!-- Content -->
</div>

<!-- Card container -->
<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
  <!-- Content -->
</div>
```

#### Component Spacing
```svelte
<!-- Form field spacing -->
<div class="space-y-4">
  <div>
    <label class="block text-sm font-medium mb-1">Email</label>
    <input type="email" class="w-full" />
  </div>
  <div>
    <label class="block text-sm font-medium mb-1">Password</label>
    <input type="password" class="w-full" />
  </div>
</div>

<!-- Button group -->
<div class="flex gap-3">
  <button class="btn-primary">Save</button>
  <button class="btn-secondary">Cancel</button>
</div>

<!-- Inline items -->
<div class="flex items-center gap-4">
  <span class="text-gray-500">Status:</span>
  <span class="badge-success">Active</span>
</div>
```

#### Table Spacing
```svelte
<table class="w-full">
  <thead class="bg-gray-50 dark:bg-gray-800">
    <tr>
      <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider">
        Name
      </th>
      <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider">
        Status
      </th>
    </tr>
  </thead>
  <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
    <tr>
      <td class="px-6 py-4 whitespace-nowrap">John Doe</td>
      <td class="px-6 py-4 whitespace-nowrap">Active</td>
    </tr>
  </tbody>
</table>
```

### 4.3 Responsive Spacing
```svelte
<!-- Mobile-first spacing -->
<div class="p-4 md:p-6 lg:p-8">
  <!-- Small on mobile, larger on desktop -->
</div>

<!-- Gap responsive -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6 lg:gap-8">
  <!-- Cards -->
</div>
```

---

## 5. Layout Patterns

### 5.1 App Shell Layout
```svelte
<!-- src/routes/+layout.svelte -->
<script>
  import { page } from '$app/stores';
  import Sidebar from '$lib/shared/components/layout/Sidebar.svelte';
  import Navbar from '$lib/shared/components/layout/Navbar.svelte';
</script>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
  <!-- Fixed sidebar on desktop -->
  <Sidebar class="hidden lg:fixed lg:inset-y-0 lg:flex lg:w-64" />
  
  <!-- Main content area -->
  <div class="lg:pl-64">
    <!-- Top navigation -->
    <Navbar class="sticky top-0 z-40" />
    
    <!-- Page content -->
    <main class="p-4 md:p-6 lg:p-8">
      <slot />
    </main>
  </div>
</div>
```

### 5.2 Card Layout
```svelte
<div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700">
  <!-- Header -->
  <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
        Client Details
      </h3>
      <button class="text-gray-400 hover:text-gray-600">
        <svg class="w-5 h-5">...</svg>
      </button>
    </div>
  </div>
  
  <!-- Content -->
  <div class="px-6 py-4">
    <slot />
  </div>
  
  <!-- Footer -->
  <div class="px-6 py-4 bg-gray-50 dark:bg-gray-900/50 border-t border-gray-200 dark:border-gray-700 rounded-b-xl">
    <slot name="footer" />
  </div>
</div>
```

### 5.3 Form Layout
```svelte
<form class="max-w-2xl space-y-6">
  <!-- Section -->
  <div class="space-y-4">
    <h3 class="text-lg font-medium text-gray-900 dark:text-white">
      Personal Information
    </h3>
    
    <!-- Grid form -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label for="firstName" class="block text-sm font-medium mb-1">
          First Name
        </label>
        <input 
          id="firstName" 
          type="text" 
          class="w-full rounded-lg border-gray-300 dark:border-gray-600"
        />
      </div>
      <div>
        <label for="lastName" class="block text-sm font-medium mb-1">
          Last Name
        </label>
        <input 
          id="lastName" 
          type="text" 
          class="w-full rounded-lg border-gray-300 dark:border-gray-600"
        />
      </div>
    </div>
  </div>
  
  <!-- Actions -->
  <div class="flex justify-end gap-3 pt-4 border-t">
    <button type="button" class="btn-secondary">Cancel</button>
    <button type="submit" class="btn-primary">Save Changes</button>
  </div>
</form>
```

### 5.4 Data Table Layout
```svelte
<div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm overflow-hidden">
  <!-- Toolbar -->
  <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div class="relative max-w-xs">
        <input 
          type="search" 
          placeholder="Search..." 
          class="w-full pl-10 pr-4 py-2 rounded-lg border"
        />
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400">
          ...
        </svg>
      </div>
      <div class="flex gap-2">
        <button class="btn-secondary">Export</button>
        <button class="btn-primary">Add Client</button>
      </div>
    </div>
  </div>
  
  <!-- Table -->
  <table class="w-full">
    <thead class="bg-gray-50 dark:bg-gray-800/50">
      <tr>
        <th class="px-6 py-3 text-left text-xs font-medium uppercase">Name</th>
        <th class="px-6 py-3 text-left text-xs font-medium uppercase">Status</th>
        <th class="px-6 py-3 text-left text-xs font-medium uppercase">Credit</th>
        <th class="px-6 py-3 text-right text-xs font-medium uppercase">Actions</th>
      </tr>
    </thead>
    <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
      <slot name="rows" />
    </tbody>
  </table>
  
  <!-- Pagination -->
  <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700">
    <slot name="pagination" />
  </div>
</div>
```

### 5.5 Modal/Dialog Layout
```svelte
<div class="fixed inset-0 z-50 overflow-y-auto">
  <!-- Backdrop -->
  <div class="fixed inset-0 bg-black/50 transition-opacity" />
  
  <!-- Modal container -->
  <div class="flex min-h-full items-center justify-center p-4">
    <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-xl max-w-lg w-full">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
        <h3 class="text-lg font-semibold">Create New Client</h3>
      </div>
      
      <!-- Body -->
      <div class="px-6 py-4 max-h-[60vh] overflow-y-auto">
        <slot />
      </div>
      
      <!-- Footer -->
      <div class="px-6 py-4 bg-gray-50 dark:bg-gray-900/50 rounded-b-xl flex justify-end gap-3">
        <slot name="footer" />
      </div>
    </div>
  </div>
</div>
```

---

## 6. Component Design Patterns

### 6.1 Button Component
```svelte
<!-- src/lib/shared/components/forms/Button.svelte -->
<script lang="ts">
  export let variant: 'primary' | 'secondary' | 'danger' | 'ghost' = 'primary';
  export let size: 'sm' | 'md' | 'lg' = 'md';
  export let disabled = false;
  export let loading = false;
  export let type: 'button' | 'submit' | 'reset' = 'button';
</script>

<button
  {type}
  {disabled}
  class="
    inline-flex items-center justify-center font-medium rounded-lg
    transition-colors duration-200
    focus:outline-none focus:ring-2 focus:ring-offset-2
    disabled:opacity-50 disabled:cursor-not-allowed
    {variant === 'primary' && 'bg-primary-600 hover:bg-primary-700 text-white focus:ring-primary-500'}
    {variant === 'secondary' && 'bg-gray-100 hover:bg-gray-200 text-gray-700 dark:bg-gray-700 dark:hover:bg-gray-600 dark:text-gray-200 focus:ring-gray-500'}
    {variant === 'danger' && 'bg-red-600 hover:bg-red-700 text-white focus:ring-red-500'}
    {variant === 'ghost' && 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800 focus:ring-gray-500'}
    {size === 'sm' && 'px-3 py-1.5 text-sm'}
    {size === 'md' && 'px-4 py-2 text-sm'}
    {size === 'lg' && 'px-6 py-3 text-base'}
    {loading && 'cursor-wait'}
  "
  on:click
>
  {#if loading}
    <svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
    </svg>
  {/if}
  <slot />
</button>
```

### 6.2 Input Component
```svelte
<!-- src/lib/shared/components/forms/Input.svelte -->
<script lang="ts">
  export let id: string;
  export let label: string;
  export let type = 'text';
  export let value = '';
  export let error = '';
  export let placeholder = '';
  export let required = false;
  export let disabled = false;
  export let helpText = '';
</script>

<div class="space-y-1">
  <label for={id} class="block text-sm font-medium text-gray-700 dark:text-gray-300">
    {label}
    {#if required}<span class="text-red-500">*</span>{/if}
  </label>
  
  <input
    {id}
    {type}
    bind:value
    {placeholder}
    {required}
    {disabled}
    class="
      w-full rounded-lg border
      px-4 py-2.5
      text-gray-900 dark:text-white
      bg-white dark:bg-gray-800
      border-gray-300 dark:border-gray-600
      placeholder:text-gray-400 dark:placeholder:text-gray-500
      focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent
      disabled:bg-gray-100 dark:disabled:bg-gray-700 disabled:cursor-not-allowed
      {error && 'border-red-500 focus:ring-red-500'}
    "
  />
  
  {#if error}
    <p class="text-sm text-red-600 dark:text-red-400">{error}</p>
  {:else if helpText}
    <p class="text-sm text-gray-500 dark:text-gray-400">{helpText}</p>
  {/if}
</div>
```

### 6.3 Badge Component
```svelte
<!-- src/lib/shared/components/display/Badge.svelte -->
<script lang="ts">
  export let variant: 'gray' | 'green' | 'yellow' | 'red' | 'blue' = 'gray';
  export let size: 'sm' | 'md' = 'md';
</script>

<span class="
  inline-flex items-center font-medium rounded-full
  {variant === 'gray' && 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'}
  {variant === 'green' && 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-400'}
  {variant === 'yellow' && 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-400'}
  {variant === 'red' && 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-400'}
  {variant === 'blue' && 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-400'}
  {size === 'sm' && 'px-2 py-0.5 text-xs'}
  {size === 'md' && 'px-2.5 py-1 text-sm'}
">
  <slot />
</span>
```

### 6.4 Card Component
```svelte
<!-- src/lib/shared/components/layout/Card.svelte -->
<script lang="ts">
  export let padding: 'none' | 'sm' | 'md' | 'lg' = 'md';
  export let hover = false;
</script>

<div class="
  bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700
  {hover && 'hover:shadow-md transition-shadow duration-200'}
  {padding === 'none' && ''}
  {padding === 'sm' && 'p-4'}
  {padding === 'md' && 'p-6'}
  {padding === 'lg' && 'p-8'}
">
  <slot />
</div>
```

### 6.5 Data Table Component
```svelte
<!-- src/lib/shared/components/data/Table.svelte -->
<script lang="ts">
  export let columns: { key: string; label: string; sortable?: boolean }[];
  export let data: Record<string, unknown>[];
  export let sortKey = '';
  export let sortDirection: 'asc' | 'desc' = 'asc';
</script>

<div class="overflow-x-auto">
  <table class="w-full">
    <thead class="bg-gray-50 dark:bg-gray-800/50">
      <tr>
        {#each columns as col}
          <th 
            class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider
              {col.sortable ? 'cursor-pointer select-none hover:bg-gray-100 dark:hover:bg-gray-800' : ''}"
            on:click={() => {
              if (col.sortable) {
                if (sortKey === col.key) {
                  sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
                } else {
                  sortKey = col.key;
                  sortDirection = 'asc';
                }
              }
            }}
          >
            <div class="flex items-center gap-2">
              <span>{col.label}</span>
              {#if sortKey === col.key}
                <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
              {/if}
            </div>
          </th>
        {/each}
      </tr>
    </thead>
    <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
      {#each data as row}
        <tr class="hover:bg-gray-50 dark:hover:bg-gray-800/50">
          {#each columns as col}
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
              {row[col.key]}
            </td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
</div>
```

---

## 7. Visual Design Checklist

### 7.1 Consistency
- [ ] Same color used for primary actions throughout
- [ ] Typography scale applied consistently
- [ ] Spacing increments follow the scale
- [ ] Border radius is consistent (use `rounded-lg` or `rounded-xl`)
- [ ] Shadow styles are consistent
- [ ] Icon style is consistent (e.g., all use same library)

### 7.2 Whitespace
- [ ] Adequate padding inside cards (at least `p-6`)
- [ ] Space between related elements (`gap-4` or `space-y-4`)
- [ ] Space between sections (`my-8` or `gap-8`)
- [ ] No overcrowding in forms and tables
- [ ] Visual breathing room around CTAs

### 7.3 Visual Feedback
- [ ] Hover states on interactive elements
- [ ] Focus states for accessibility
- [ ] Loading states for async operations
- [ ] Success/error states for forms
- [ ] Disabled states for unavailable actions
- [ ] Transition animations (200ms recommended)

### 7.4 Hierarchy
- [ ] Primary actions are visually dominant
- [ ] Section headings are clear
- [ ] Most important information is prominent
- [ ] Related items are grouped
- [ ] Unrelated items are separated

### 7.5 Accessibility
- [ ] Sufficient color contrast (4.5:1 minimum for text)
- [ ] Keyboard navigation works
- [ ] Focus indicators visible
- [ ] ARIA labels where needed
- [ ] Semantic HTML
- [ ] Reduced motion respected

### 7.6 Responsiveness
- [ ] Mobile-first approach
- [ ] Touch targets at least 44x44px
- [ ] Stacking on mobile, grid on desktop
- [ ] Readable text on all screen sizes
- [ ] No horizontal scrolling
- [ ] Proper breakpoint usage

---

## 8. Icons

### 8.1 Icon Library Recommendation
Use **Lucide Icons** or **Heroicons** for consistent, professional icons:

```bash
npm install lucide-svelte
```

### 8.2 Icon Sizes
```svelte
<script>
  import { Home, Settings, User, Bell, Search, Menu, X } from 'lucide-svelte';
</script>

<!-- Small icons (16x16) -->
<button class="p-1"><Home class="w-4 h-4" /></button>

<!-- Medium icons (20x20) - Default -->
<button class="p-2"><Settings class="w-5 h-5" /></button>

<!-- Large icons (24x24) -->
<button class="p-3"><User class="w-6 h-6" /></button>

<!-- Extra large icons (32x32) -->
<div class="p-4"><Bell class="w-8 h-8" /></div>
```

### 8.3 Icon Usage Guidelines
```svelte
<!-- Navigation -->
<nav class="flex gap-4">
  <a href="/dashboard" class="flex items-center gap-2 text-gray-600 hover:text-primary-600">
    <Home class="w-5 h-5" />
    <span>Dashboard</span>
  </a>
</nav>

<!-- Action buttons -->
<button class="flex items-center gap-2 text-gray-600 hover:text-gray-900">
  <Settings class="w-5 h-5" />
  <span>Settings</span>
</button>

<!-- Status indicators -->
<Badge variant="green">
  <Check class="w-4 h-4 mr-1" />
  Active
</Badge>
```

---

## 9. Animations and Transitions

### 9.1 Svelte Transitions
```svelte
<script>
  import { fade, fly, slide, scale } from 'svelte/transition';
  import { quintOut, elasticOut } from 'svelte/easing';
  
  let show = false;
</script>

<!-- Fade in -->
<div in:fade={{ duration: 200 }} out:fade={{ duration: 150 }}>
  Content
</div>

<!-- Slide -->
<div transition:slide={{ duration: 300 }}>
  Collapsible content
</div>

<!-- Fly -->
<div in:fly={{ y: 20, duration: 300, easing: quintOut }}>
  Content slides up
</div>

<!-- Scale -->
<div in:scale={{ duration: 300, easing: elasticOut }}>
  Pop in effect
</div>
```

### 9.2 Hover Effects
```svelte
<button class="
  transition-all duration-200
  hover:scale-105
  hover:shadow-lg
  active:scale-95
">
  Animated Button
</button>
```

### 9.3 Loading States
```svelte
<script>
  import { Spinner } from '$lib/shared/components/display';
</script>

{#if loading}
  <div class="flex justify-center py-8">
    <Spinner size="lg" />
  </div>
{:else}
  <div>Content loaded</div>
{/if}

<!-- Skeleton loader -->
<div class="animate-pulse space-y-4">
  <div class="h-4 bg-gray-200 rounded w-3/4"></div>
  <div class="h-4 bg-gray-200 rounded w-1/2"></div>
  <div class="h-4 bg-gray-200 rounded w-5/6"></div>
</div>
```

---

## 10. Error States

### 10.1 Form Validation Errors
```svelte
<script>
  let errors = {
    email: 'Please enter a valid email address',
    password: 'Password must be at least 8 characters'
  };
  
  let touched = {
    email: false,
    password: false
  };
</script>

<div class="space-y-4">
  <Input 
    id="email"
    label="Email"
    bind:value={email}
    error={touched.email ? errors.email : ''}
    on:blur={() => touched.email = true}
  />
  
  <Input 
    id="password"
    type="password"
    label="Password"
    bind:value={password}
    error={touched.password ? errors.password : ''}
    on:blur={() => touched.password = true}
  />
</div>
```

### 10.2 Error Boundaries
```svelte
<script>
  import { ErrorMessage } from '$lib/shared/components/display';
</script>

{#await data}
  <Spinner />
{:then data}
  <Content {data} />
{:catch error}
  <ErrorMessage 
    title="Failed to load data"
    message={error.message}
    on:retry={() => invalidateAll()}
  />
{/await}
```

### 10.3 Empty States
```svelte
<div class="text-center py-12">
  <div class="mx-auto w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center mb-4">
    <Users class="w-8 h-8 text-gray-400" />
  </div>
  <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
    No clients yet
  </h3>
  <p class="text-gray-500 dark:text-gray-400 mb-6">
    Get started by creating your first client.
  </p>
  <Button variant="primary">Create Client</Button>
</div>
```

---

## 11. File Structure

```
frontend/src/
├── lib/
│   ├── shared/
│   │   ├── components/
│   │   │   ├── display/
│   │   │   │   ├── Alert.svelte
│   │   │   │   ├── Avatar.svelte
│   │   │   │   ├── Badge.svelte
│   │   │   │   ├── Chip.svelte
│   │   │   │   ├── EmptyState.svelte
│   │   │   │   ├── Progress.svelte
│   │   │   │   ├── Skeleton.svelte
│   │   │   │   ├── Spinner.svelte
│   │   │   │   ├── Toast.svelte
│   │   │   │   └── Tooltip.svelte
│   │   │   ├── forms/
│   │   │   │   ├── Button.svelte
│   │   │   │   ├── Checkbox.svelte
│   │   │   │   ├── DatePicker.svelte
│   │   │   │   ├── FileUpload.svelte
│   │   │   │   ├── Input.svelte
│   │   │   │   ├── Radio.svelte
│   │   │   │   ├── Select.svelte
│   │   │   │   ├── Textarea.svelte
│   │   │   │   └── Toggle.svelte
│   │   │   ├── layout/
│   │   │   │   ├── Card.svelte
│   │   │   │   ├── Drawer.svelte
│   │   │   │   ├── Modal.svelte
│   │   │   │   ├── PageHeader.svelte
│   │   │   │   ├── Sidebar.svelte
│   │   │   │   ├── Tab.svelte
│   │   │   │   └── Toolbar.svelte
│   │   │   └── data/
│   │   │       ├── Column.svelte
│   │   │       ├── DataGrid.svelte
│   │   │       ├── EmptyRow.svelte
│   │   │       ├── Pagination.svelte
│   │   │       ├── Row.svelte
│   │   │       └── Table.svelte
│   │   │
│   │   └── styles/
│   │       ├── variables.css
│   │       ├── global.css
│   │       └── utilities.css
│   │
│   └── plugins/
│       ├── clients/
│       │   ├── components/
│       │   │   ├── ClientCard.svelte
│       │   │   ├── ClientForm.svelte
│       │   │   ├── ClientList.svelte
│       │   │   └── CreditWarning.svelte
│       │   └── routes/
│       │       └── ...
│       └── ...
│
└── routes/
    └── +layout.svelte
```

---

## 12. Quick Start Checklist

### Before Starting Development

- [ ] Install Tailwind CSS
- [ ] Configure color variables in `variables.css`
- [ ] Set up typography scale
- [ ] Create base component library (Button, Input, Card, Badge, Table)
- [ ] Install icon library (lucide-svelte)
- [ ] Set up dark mode support
- [ ] Create layout components (Sidebar, Navbar)
- [ ] Define form validation schema (using Zod)
- [ ] Set up error handling boundary

### For Each New Feature

1. Create component in shared library if reusable
2. Use existing components from library
3. Follow color, typography, and spacing guidelines
4. Add hover/focus/active states
5. Include loading and error states
6. Test responsiveness
7. Verify accessibility
8. Add to component documentation

---

## 13. Tools and Resources

### Design Tools
- **Figma** - UI design and prototyping
- **Color Contrast Checker** - Accessibility verification
- **Tailwind CSS** - Utility-first CSS framework

### Useful Packages
```bash
npm install tailwindcss postcss autoprefixer
npm install lucide-svelte
npm install clsx tailwind-merge
npm install zod
npm install date-fns
```

### Validation Helpers
```typescript
// src/lib/shared/utils/forms.ts
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
```

---

## Summary

Following this guide will ensure:

1. **Consistent Design** - Uniform colors, typography, and spacing
2. **Professional Appearance** - Clean, modern, business-appropriate UI
3. **Accessibility** - WCAG compliant, keyboard navigable
4. **Responsiveness** - Works on all device sizes
5. **Maintainability** - Component-based, reusable code
6. **User Experience** - Clear feedback, intuitive interactions

The key principles are:
- Use the design system consistently
- Prioritize clarity and simplicity
- Provide visual feedback for all interactions
- Test across devices and with assistive technologies
- Iterate based on user feedback
