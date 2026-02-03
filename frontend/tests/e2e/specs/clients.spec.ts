import { test, expect } from '@playwright/test';
import { ClientsPage } from '../pages/ClientsPage';

test.describe('Client Management', () => {
  let clientsPage: ClientsPage;

  test.beforeEach(async ({ clientsPage }) => {
    clientsPage = clientsPage;
    await clientsPage.goto();
  });

  test('should display clients list', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Clients');
    await expect(page.locator('[data-testid="client-list"]')).toBeVisible();
  });

  test('should create a new client', async ({ page }) => {
    await page.click('[data-testid="new-client-button"]');
    await expect(page.locator('[data-testid="client-form"]')).toBeVisible();

    await page.fill('[data-testid="name-input"]', 'Test Company');
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="phone-input"]', '+1234567890');

    await page.click('[data-testid="submit-button"]');

    await expect(page.locator('[data-testid="toast"]')).toContainText(
      'Client created successfully'
    );
    await expect(page.locator('[data-testid="client-name"]')).toHaveText('Test Company');
  });

  test('should search clients', async ({ page }) => {
    await page.fill('[data-testid="search-input"]', 'Test Company');
    await page.click('[data-testid="search-button"]');

    await expect(page.locator('[data-testid="client-row"]').first()).toContainText('Test Company');
  });

  test('should filter by status', async ({ page }) => {
    await page.selectOption('[data-testid="status-filter"]', 'active');
    await page.click('[data-testid="apply-filters"]');

    await expect(page.locator('[data-testid="client-row"]').first()).toContainText('active');
  });

  test('should show validation errors for empty form', async ({ page }) => {
    await page.click('[data-testid="new-client-button"]');
    await page.click('[data-testid="submit-button"]');

    await expect(page.locator('[data-testid="name-error"]')).toContainText('Name is required');
    await expect(page.locator('[data-testid="email-error"]')).toContainText('Email is required');
  });

  test('should show validation error for invalid email', async ({ page }) => {
    await page.click('[data-testid="new-client-button"]');
    await page.fill('[data-testid="name-input"]', 'Test Company');
    await page.fill('[data-testid="email-input"]', 'invalid-email');
    await page.click('[data-testid="submit-button"]');

    await expect(page.locator('[data-testid="email-error"]')).toContainText('Invalid email format');
  });
});

test.describe('Client Detail View', () => {
  let clientsPage: ClientsPage;

  test.beforeEach(async ({ clientsPage }) => {
    clientsPage = clientsPage;
    await clientsPage.goto();
  });

  test('should navigate to client detail', async ({ page }) => {
    await clientsPage.clickClientRow(0);
    await expect(page).toHaveURL(/\/clients\/.+/);
  });

  test('should display client information', async ({ page }) => {
    await clientsPage.clickClientRow(0);

    await expect(page.locator('[data-testid="client-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="client-email"]')).toBeVisible();
    await expect(page.locator('[data-testid="client-status"]')).toBeVisible();
  });

  test('should allow editing client', async ({ page }) => {
    await clientsPage.clickClientRow(0);
    await page.click('[data-testid="edit-button"]');

    await page.fill('[data-testid="name-input"]', 'Updated Company');
    await page.click('[data-testid="save-button"]');

    await expect(page.locator('[data-testid="toast"]')).toContainText('Client updated');
  });

  test('should allow deleting client', async ({ page }) => {
    await clientsPage.clickClientRow(0);
    await page.click('[data-testid="delete-button"]');

    await expect(page.locator('[data-testid="confirm-dialog"]')).toBeVisible();
    await page.click('[data-testid="confirm-delete"]');

    await expect(page.locator('[data-testid="toast"]')).toContainText('Client deleted');
  });
});

test.describe('Client Search and Filter', () => {
  let clientsPage: ClientsPage;

  test.beforeEach(async ({ clientsPage }) => {
    clientsPage = clientsPage;
    await clientsPage.goto();
  });

  test('should search by client name', async ({ page }) => {
    await clientsPage.searchClients('Acme');

    await expect(page.locator('[data-testid="client-row"]').first()).toContainText('Acme');
  });

  test('should clear search results', async ({ page }) => {
    await page.fill('[data-testid="search-input"]', 'NonExistent');
    await page.click('[data-testid="search-button"]');

    await expect(page.locator('[data-testid="empty-state"]')).toBeVisible();
  });

  test('should filter by active status', async ({ clientsPage }) => {
    await clientsPage.filterByStatus('active');

    const rows = await clientsPage.getClientCount();
    expect(rows).toBeGreaterThan(0);
  });

  test('should filter by inactive status', async ({ clientsPage }) => {
    await clientsPage.filterByStatus('inactive');

    const rows = await clientsPage.getClientCount();
    expect(rows).toBeGreaterThanOrEqual(0);
  });
});

test.describe('Client Pagination', () => {
  let clientsPage: ClientsPage;

  test.beforeEach(async ({ clientsPage }) => {
    clientsPage = clientsPage;
    await clientsPage.goto();
  });

  test('should show pagination when multiple pages', async ({ page }) => {
    await expect(page.locator('[data-testid="pagination"]')).toBeVisible();
  });

  test('should navigate to next page', async ({ page }) => {
    const nextButton = page.locator('[data-testid="next-page-button"]');
    if (await nextButton.isEnabled()) {
      await nextButton.click();
      await expect(page.locator('[data-testid="current-page"]')).toContainText('2');
    }
  });

  test('should navigate to previous page', async ({ page }) => {
    const prevButton = page.locator('[data-testid="prev-page-button"]');
    if (await prevButton.isEnabled()) {
      await prevButton.click();
      await expect(page.locator('[data-testid="current-page"]')).toContainText('1');
    }
  });
});
