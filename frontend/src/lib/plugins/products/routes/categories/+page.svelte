<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Textarea from '$lib/shared/components/forms/Textarea.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';
  import Table from '$lib/shared/components/data/Table.svelte';
  import Badge from '$lib/shared/components/display/Badge.svelte';
  import Spinner from '$lib/shared/components/display/Spinner.svelte';
  import Alert from '$lib/shared/components/display/Alert.svelte';
  import Modal from '$lib/shared/components/layout/Modal.svelte';

  interface Category {
    id: string;
    name: string;
    slug: string;
    description: string;
    parentId: string | null;
    productCount: number;
    status: 'active' | 'inactive';
    sortOrder: number;
    createdAt: string;
    updatedAt: string;
  }

  let categories: Category[] = [];
  let loading = true;
  let error: string | null = null;
  let saving = false;
  let searchQuery = '';

  // Modals
  let showCreateModal = false;
  let showEditModal = false;
  let showDeleteModal = false;
  let selectedCategory: Category | null = null;

  // Form fields
  let categoryName = '';
  let categorySlug = '';
  let categoryDescription = '';
  let categoryParent = '';
  let categoryStatus: 'active' | 'inactive' = 'active';
  let categorySortOrder = '0';

  const columns = [
    { key: 'name', label: 'Name', sortable: true },
    { key: 'slug', label: 'Slug', sortable: true },
    { key: 'productCount', label: 'Products', sortable: true },
    { key: 'status', label: 'Status', sortable: true },
    { key: 'sortOrder', label: 'Order', sortable: true },
    { key: 'actions', label: 'Actions', sortable: false }
  ];

  async function loadCategories() {
    loading = true;
    error = null;
    
    try {
      // Mock data for categories
      await new Promise(resolve => setTimeout(resolve, 500));
      categories = [
        {
          id: 'cat-1',
          name: 'Electronics',
          slug: 'electronics',
          description: 'Electronic devices and accessories',
          parentId: null,
          productCount: 156,
          status: 'active',
          sortOrder: 1,
          createdAt: '2024-01-15T10:00:00Z',
          updatedAt: '2024-01-20T14:30:00Z'
        },
        {
          id: 'cat-2',
          name: 'Computers',
          slug: 'computers',
          description: 'Laptops, desktops, and computer accessories',
          parentId: 'cat-1',
          productCount: 45,
          status: 'active',
          sortOrder: 1,
          createdAt: '2024-01-16T10:00:00Z',
          updatedAt: '2024-01-20T14:30:00Z'
        },
        {
          id: 'cat-3',
          name: 'Smartphones',
          slug: 'smartphones',
          description: 'Mobile phones and accessories',
          parentId: 'cat-1',
          productCount: 67,
          status: 'active',
          sortOrder: 2,
          createdAt: '2024-01-16T11:00:00Z',
          updatedAt: '2024-01-20T14:30:00Z'
        },
        {
          id: 'cat-4',
          name: 'Clothing',
          slug: 'clothing',
          description: 'Apparel and fashion items',
          parentId: null,
          productCount: 234,
          status: 'active',
          sortOrder: 2,
          createdAt: '2024-01-17T10:00:00Z',
          updatedAt: '2024-01-20T14:30:00Z'
        },
        {
          id: 'cat-5',
          name: 'Men\'s Clothing',
          slug: 'mens-clothing',
          description: 'Clothing for men',
          parentId: 'cat-4',
          productCount: 89,
          status: 'active',
          sortOrder: 1,
          createdAt: '2024-01-17T11:00:00Z',
          updatedAt: '2024-01-20T14:30:00Z'
        },
        {
          id: 'cat-6',
          name: 'Women\'s Clothing',
          slug: 'womens-clothing',
          description: 'Clothing for women',
          parentId: 'cat-4',
          productCount: 112,
          status: 'active',
          sortOrder: 2,
          createdAt: '2024-01-17T12:00:00Z',
          updatedAt: '2024-01-20T14:30:00Z'
        },
        {
          id: 'cat-7',
          name: 'Food & Beverage',
          slug: 'food-beverage',
          description: 'Food items and drinks',
          parentId: null,
          productCount: 89,
          status: 'active',
          sortOrder: 3,
          createdAt: '2024-01-18T10:00:00Z',
          updatedAt: '2024-01-20T14:30:00Z'
        },
        {
          id: 'cat-8',
          name: 'Discontinued Items',
          slug: 'discontinued',
          description: 'No longer available products',
          parentId: null,
          productCount: 12,
          status: 'inactive',
          sortOrder: 99,
          createdAt: '2024-01-19T10:00:00Z',
          updatedAt: '2024-01-20T14:30:00Z'
        }
      ];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load categories';
    } finally {
      loading = false;
    }
  }

  function getStatusVariant(status: string): 'green' | 'gray' {
    return status === 'active' ? 'green' : 'gray';
  }

  function getParentName(parentId: string | null): string {
    if (!parentId) return '-';
    const parent = categories.find(c => c.id === parentId);
    return parent ? parent.name : '-';
  }

  function generateSlug(name: string): string {
    return name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '');
  }

  function handleNameChange() {
    if (!categorySlug || categorySlug === generateSlug(categoryName.slice(0, -1))) {
      categorySlug = generateSlug(categoryName);
    }
  }

  function openCreateModal() {
    categoryName = '';
    categorySlug = '';
    categoryDescription = '';
    categoryParent = '';
    categoryStatus = 'active';
    categorySortOrder = '0';
    showCreateModal = true;
  }

  function openEditModal(category: Category) {
    selectedCategory = category;
    categoryName = category.name;
    categorySlug = category.slug;
    categoryDescription = category.description;
    categoryParent = category.parentId || '';
    categoryStatus = category.status;
    categorySortOrder = category.sortOrder.toString();
    showEditModal = true;
  }

  function openDeleteModal(category: Category) {
    selectedCategory = category;
    showDeleteModal = true;
  }

  async function handleCreateCategory() {
    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      const newCategory: Category = {
        id: `cat-${Date.now()}`,
        name: categoryName,
        slug: categorySlug || generateSlug(categoryName),
        description: categoryDescription,
        parentId: categoryParent || null,
        productCount: 0,
        status: categoryStatus,
        sortOrder: parseInt(categorySortOrder, 10) || 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };
      categories = [...categories, newCategory];
      showCreateModal = false;
    } catch (err) {
      error = 'Failed to create category';
    } finally {
      saving = false;
    }
  }

  async function handleUpdateCategory() {
    if (!selectedCategory) return;
    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      categories = categories.map(c =>
        c.id === selectedCategory?.id
          ? {
              ...c,
              name: categoryName,
              slug: categorySlug,
              description: categoryDescription,
              parentId: categoryParent || null,
              status: categoryStatus,
              sortOrder: parseInt(categorySortOrder, 10) || 0,
              updatedAt: new Date().toISOString()
            }
          : c
      );
      showEditModal = false;
      selectedCategory = null;
    } catch (err) {
      error = 'Failed to update category';
    } finally {
      saving = false;
    }
  }

  async function handleDeleteCategory() {
    if (!selectedCategory) return;
    saving = true;
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      // Check if category has children
      const hasChildren = categories.some(c => c.parentId === selectedCategory?.id);
      if (hasChildren) {
        error = 'Cannot delete category with subcategories. Please delete subcategories first.';
        showDeleteModal = false;
        selectedCategory = null;
        saving = false;
        return;
      }
      categories = categories.filter(c => c.id !== selectedCategory?.id);
      showDeleteModal = false;
      selectedCategory = null;
    } catch (err) {
      error = 'Failed to delete category';
    } finally {
      saving = false;
    }
  }

  function getFilteredCategories(): Category[] {
    if (!searchQuery.trim()) return categories;
    const query = searchQuery.toLowerCase();
    return categories.filter(c => 
      c.name.toLowerCase().includes(query) ||
      c.slug.toLowerCase().includes(query) ||
      c.description.toLowerCase().includes(query)
    );
  }

  function getRootCategories(): Category[] {
    return categories.filter(c => !c.parentId).sort((a, b) => a.sortOrder - b.sortOrder);
  }

  function getChildCategories(parentId: string): Category[] {
    return categories.filter(c => c.parentId === parentId).sort((a, b) => a.sortOrder - b.sortOrder);
  }

  onMount(() => {
    loadCategories();
  });
</script>

<svelte:head>
  <title>Product Categories | ERP System</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div class="header-content">
      <h1 class="page-title">Product Categories</h1>
      <p class="page-description">Manage your product catalog categories</p>
    </div>
    <div class="header-actions">
      <Button variant="primary" on:click={openCreateModal}>
        Add Category
      </Button>
    </div>
  </div>

  {#if error}
    <Alert variant="error" dismissible on:dismiss={() => error = null} class="mb-4">
      {error}
    </Alert>
  {/if}

  <Card>
    <div class="filters">
      <div class="filter-row">
        <div class="filter-item search-filter">
          <Input
            id="search"
            label="Search"
            type="search"
            placeholder="Search categories..."
            bind:value={searchQuery}
          />
        </div>
      </div>
    </div>

    {#if loading}
      <div class="loading-container">
        <Spinner size="lg" />
        <p>Loading categories...</p>
      </div>
    {:else if getFilteredCategories().length === 0}
      <div class="empty-container">
        <svg class="w-16 h-16 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <p class="text-gray-500 mb-4">No categories found</p>
        <Button variant="primary" on:click={openCreateModal}>
          Create First Category
        </Button>
      </div>
    {:else}
      <div class="categories-tree">
        {#each getRootCategories() as category}
          <div class="category-group">
            <div class="category-item main">
              <div class="category-info">
                <span class="category-name">{category.name}</span>
                <span class="category-meta">
                  {category.slug} • {category.productCount} products
                </span>
              </div>
              <div class="category-actions">
                <Badge variant={getStatusVariant(category.status)}>
                  {category.status}
                </Badge>
                <Button variant="ghost" size="sm" on:click={() => openEditModal(category)}>
                  Edit
                </Button>
                <Button variant="ghost" size="sm" on:click={() => openDeleteModal(category)}>
                  Delete
                </Button>
              </div>
            </div>
            {#each getChildCategories(category.id) as child}
              <div class="category-item child">
                <div class="category-info">
                  <span class="category-name">{child.name}</span>
                  <span class="category-meta">
                    {child.slug} • {child.productCount} products
                  </span>
                </div>
                <div class="category-actions">
                  <Badge variant={getStatusVariant(child.status)}>
                    {child.status}
                  </Badge>
                  <Button variant="ghost" size="sm" on:click={() => openEditModal(child)}>
                    Edit
                  </Button>
                  <Button variant="ghost" size="sm" on:click={() => openDeleteModal(child)}>
                    Delete
                  </Button>
                </div>
              </div>
            {/each}
          </div>
        {/each}
      </div>
    {/if}
  </Card>
</div>

<!-- Create Category Modal -->
<Modal
  bind:open={showCreateModal}
  title="Create Category"
  size="md"
>
  <div class="modal-form">
    <div class="form-row">
      <Input
        id="categoryName"
        label="Category Name"
        type="text"
        placeholder="Enter category name"
        bind:value={categoryName}
        on:input={handleNameChange}
        required
      />
    </div>
    <div class="form-row">
      <Input
        id="categorySlug"
        label="Slug"
        type="text"
        placeholder="category-slug"
        bind:value={categorySlug}
        helper="Used in URLs"
      />
    </div>
    <div class="form-row">
      <Textarea
        id="categoryDescription"
        label="Description"
        placeholder="Enter category description"
        bind:value={categoryDescription}
        rows={3}
        maxLength={500}
      />
    </div>
    <div class="form-row two-col">
      <div class="form-item">
        <label class="form-label">Parent Category</label>
        <select
          id="categoryParent"
          bind:value={categoryParent}
          class="form-select"
        >
          <option value="">None (Root Category)</option>
          {#each categories.filter(c => !c.parentId) as cat}
            <option value={cat.id}>{cat.name}</option>
          {/each}
        </select>
      </div>
      <Input
        id="categorySortOrder"
        label="Sort Order"
        type="number"
        placeholder="0"
        bind:value={categorySortOrder}
        min="0"
        step="1"
      />
    </div>
    <div class="form-row">
      <div class="form-item">
        <label class="form-label">Status</label>
        <select
          id="categoryStatus"
          bind:value={categoryStatus}
          class="form-select"
        >
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </div>
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={saving}>
      Cancel
    </Button>
    <Button variant="primary" on:click={handleCreateCategory} loading={saving}>
      {saving ? 'Creating...' : 'Create Category'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Edit Category Modal -->
<Modal
  bind:open={showEditModal}
  title="Edit Category"
  size="md"
>
  <div class="modal-form">
    <div class="form-row">
      <Input
        id="editCategoryName"
        label="Category Name"
        type="text"
        bind:value={categoryName}
        required
      />
    </div>
    <div class="form-row">
      <Input
        id="editCategorySlug"
        label="Slug"
        type="text"
        bind:value={categorySlug}
        helper="Used in URLs"
      />
    </div>
    <div class="form-row">
      <Textarea
        id="editCategoryDescription"
        label="Description"
        bind:value={categoryDescription}
        rows={3}
        maxLength={500}
      />
    </div>
    <div class="form-row two-col">
      <div class="form-item">
        <label class="form-label">Parent Category</label>
        <select
          id="editCategoryParent"
          bind:value={categoryParent}
          class="form-select"
          disabled={selectedCategory?.id && categories.some(c => c.parentId === selectedCategory?.id)}
        >
          <option value="">None (Root Category)</option>
          {#each categories.filter(c => !c.parentId && c.id !== selectedCategory?.id) as cat}
            <option value={cat.id}>{cat.name}</option>
          {/each}
        </select>
        {#if selectedCategory?.id && categories.some(c => c.parentId === selectedCategory?.id)}
          <span class="helper-text">Cannot change parent: has subcategories</span>
        {/if}
      </div>
      <Input
        id="editCategorySortOrder"
        label="Sort Order"
        type="number"
        bind:value={categorySortOrder}
        min="0"
        step="1"
      />
    </div>
    <div class="form-row">
      <div class="form-item">
        <label class="form-label">Status</label>
        <select
          id="editCategoryStatus"
          bind:value={categoryStatus}
          class="form-select"
        >
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </div>
    </div>
  </div>
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={close} disabled={saving}>
      Cancel
    </Button>
    <Button variant="primary" on:click={handleUpdateCategory} loading={saving}>
      {saving ? 'Saving...' : 'Save Changes'}
    </Button>
  </svelte:fragment>
</Modal>

<!-- Delete Category Modal -->
<Modal
  bind:open={showDeleteModal}
  title="Delete Category"
  size="sm"
>
  <p>Are you sure you want to delete the category <strong>{selectedCategory?.name}</strong>?</p>
  {#if selectedCategory?.productCount && selectedCategory.productCount > 0}
    <p class="warning-text">
      This category contains {selectedCategory.productCount} products. 
      These products will need to be reassigned to another category.
    </p>
  {/if}
  
  <svelte:fragment slot="footer" let:close>
    <Button variant="secondary" on:click={() => { close(); showDeleteModal = false; }} disabled={saving}>
      Cancel
    </Button>
    <Button variant="danger" on:click={handleDeleteCategory} loading={saving}>
      {saving ? 'Deleting...' : 'Delete'}
    </Button>
  </svelte:fragment>
</Modal>

<style>
  .page-container {
    padding: 1.5rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
  }

  .page-title {
    font-size: 1.875rem;
    font-weight: 700;
    color: var(--color-gray-900);
    margin: 0;
  }

  .page-description {
    color: var(--color-gray-500);
    margin-top: 0.25rem;
  }

  .header-actions {
    display: flex;
    gap: 0.5rem;
  }

  .filters {
    margin-bottom: 1rem;
  }

  .filter-row {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .filter-item {
    flex: 1;
    min-width: 200px;
  }

  .search-filter {
    flex: 2;
    min-width: 300px;
  }

  .loading-container,
  .empty-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-gray-500);
  }

  .categories-tree {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .category-group {
    border: 1px solid var(--color-gray-200);
    border-radius: 0.5rem;
    overflow: hidden;
  }

  .category-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border-bottom: 1px solid var(--color-gray-100);
  }

  .category-item:last-child {
    border-bottom: none;
  }

  .category-item.main {
    background-color: var(--color-gray-50);
    font-weight: 500;
  }

  .category-item.child {
    padding-left: 2rem;
  }

  .category-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .category-name {
    font-weight: 600;
    color: var(--color-gray-900);
  }

  .category-meta {
    font-size: 0.875rem;
    color: var(--color-gray-500);
  }

  .category-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .modal-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .form-row {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .form-row.two-col {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .form-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .form-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-gray-700);
  }

  .form-select {
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--color-gray-300);
    border-radius: 0.375rem;
    background-color: white;
    font-size: 0.875rem;
    color: var(--color-gray-900);
  }

  .form-select:focus {
    outline: none;
    border-color: var(--color-primary-500);
    ring: 2px solid var(--color-primary-200);
  }

  .form-select:disabled {
    background-color: var(--color-gray-100);
    cursor: not-allowed;
  }

  .helper-text {
    font-size: 0.75rem;
    color: var(--color-gray-500);
  }

  .warning-text {
    color: var(--color-yellow-600);
    font-size: 0.875rem;
    margin-top: 0.5rem;
  }

  @media (max-width: 640px) {
    .page-header {
      flex-direction: column;
      gap: 1rem;
    }

    .form-row.two-col {
      grid-template-columns: 1fr;
    }

    .category-item {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }

    .category-actions {
      width: 100%;
      justify-content: flex-end;
    }
  }
</style>
</style>
