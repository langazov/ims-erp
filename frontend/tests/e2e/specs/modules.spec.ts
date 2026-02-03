import { test, expect } from '@playwright/test';
import { ModulesPage } from '../pages/ModulesPage';

test.describe('Modules Management', () => {
  let modulesPage: ModulesPage;

  test.beforeEach(async ({ modulesPage }) => {
    modulesPage = modulesPage;
    await modulesPage.goto();
  });

  test('should display modules page', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Modules');
    await expect(page.locator('.page-description')).toContainText('installed modules');
  });

  test('should show system statistics', async ({ modulesPage }) => {
    const total = await modulesPage.getTotalModules();
    const enabled = await modulesPage.getEnabledCount();
    const categories = await modulesPage.getCategoriesCount();

    expect(total).toBeGreaterThan(0);
    expect(enabled).toBeGreaterThan(0);
    expect(categories).toBeGreaterThan(0);
  });

  test('should display categories', async ({ page }) => {
    await expect(page.locator('.category-header').first()).toBeVisible();
  });

  test('should expand and collapse categories', async ({ modulesPage }) => {
    await modulesPage.clickCategory('Core');
    expect(await modulesPage.isCategoryExpanded('Core')).toBe(true);

    await modulesPage.clickCategory('Core');
    expect(await modulesPage.isCategoryExpanded('Core')).toBe(false);
  });

  test('should expand all categories', async ({ modulesPage }) => {
    await modulesPage.expandAll();
    expect(await modulesPage.isCategoryExpanded('Core')).toBe(true);
    expect(await modulesPage.isCategoryExpanded('Management')).toBe(true);
  });

  test('should collapse all categories', async ({ modulesPage }) => {
    await modulesPage.expandAll();
    await modulesPage.collapseAll();
    expect(await modulesPage.isCategoryExpanded('Core')).toBe(false);
  });

  test('should display module cards', async ({ modulesPage }) => {
    const count = await modulesPage.getModuleCount();
    expect(count).toBeGreaterThan(0);
  });

  test('should search modules', async ({ modulesPage }) => {
    await modulesPage.searchModules('Client');

    const names = await modulesPage.getModuleNames();
    names.forEach(name => {
      expect(name.toLowerCase()).toContain('client');
    });
  });

  test('should show empty results for non-matching search', async ({ page }) => {
    await modulesPage.searchModules('NonExistentModule12345');

    await expect(page.locator('.empty-state')).toBeVisible();
  });
});

test.describe('Module Navigation', () => {
  let modulesPage: ModulesPage;

  test.beforeEach(async ({ modulesPage }) => {
    modulesPage = modulesPage;
    await modulesPage.goto();
  });

  test('should navigate to module detail', async ({ page }) => {
    await modulesPage.clickModule('Dashboard');
    await expect(page).toHaveURL(/\/modules\/.+/);
    await expect(page.locator('h1')).toContainText('Dashboard');
  });

  test('should show module information', async ({ page }) => {
    await modulesPage.clickModule('Clients');
    await expect(page.locator('.info-item')).toContainText('Module ID');
    await expect(page.locator('.info-item')).toContainText('Category');
    await expect(page.locator('.info-item')).toContainText('Status');
  });

  test('should show module description', async ({ page }) => {
    await modulesPage.clickModule('Clients');

    const description = await page.locator('.description').textContent();
    expect(description).toContain('Client management');
  });

  test('should have open module button', async ({ page }) => {
    await modulesPage.clickModule('Dashboard');

    const button = page.locator('button:has-text("Open Module")');
    await expect(button).toBeVisible();
  });
});

test.describe('Module Categories', () => {
  let modulesPage: ModulesPage;

  test.beforeEach(async ({ modulesPage }) => {
    modulesPage = modulesPage;
    await modulesPage.goto();
  });

  test('should have Core category with dashboard', async ({ page }) => {
    await modulesPage.clickCategory('Core');
    await expect(page.locator('.module-card', { hasText: 'Dashboard' })).toBeVisible();
    await expect(page.locator('.module-card', { hasText: 'Modules' })).toBeVisible();
  });

  test('should have Management category', async ({ page }) => {
    await modulesPage.clickCategory('Management');

    await expect(page.locator('.module-card', { hasText: 'Clients' })).toBeVisible();
    await expect(page.locator('.module-card', { hasText: 'Users' })).toBeVisible();
    await expect(page.locator('.module-card', { hasText: 'Products' })).toBeVisible();
  });

  test('should have Operations category', async ({ page }) => {
    await modulesPage.clickCategory('Operations');

    await expect(page.locator('.module-card', { hasText: 'Inventory' })).toBeVisible();
    await expect(page.locator('.module-card', { hasText: 'Warehouse' })).toBeVisible();
    await expect(page.locator('.module-card', { hasText: 'Orders' })).toBeVisible();
  });

  test('should have Settings category', async ({ page }) => {
    await modulesPage.clickCategory('Settings');

    await expect(page.locator('.module-card', { hasText: 'Settings' })).toBeVisible();
    await expect(page.locator('.module-card', { hasText: 'Documents' })).toBeVisible();
  });
});

test.describe('Module Status', () => {
  let modulesPage: ModulesPage;

  test.beforeEach(async ({ modulesPage }) => {
    modulesPage = modulesPage;
    await modulesPage.goto();
  });

  test('should show enabled status badges', async ({ page }) => {
    await modulesPage.expandAll();

    const enabledBadges = page.locator('.badge:has-text("enabled")');
    await expect(enabledBadges.first()).toBeVisible();
  });
});
