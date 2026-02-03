import { test as base, expect } from '@playwright/test';
import type { Page } from '@playwright/test';
import { ClientsPage } from '../pages/ClientsPage';
import { ModulesPage } from '../pages/ModulesPage';

interface Fixtures {
  clientsPage: ClientsPage;
  modulesPage: ModulesPage;
  authenticatedPage: Page;
}

export const test = base.extend<Fixtures>({
  authenticatedPage: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();

    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@example.com');
    await page.fill('[data-testid="password"]', 'password123');
    await page.click('[data-testid="login-button"]');

    await expect(page).toHaveURL('/dashboard');

    await use(page);

    await context.close();
  },

  clientsPage: async ({ authenticatedPage }, use) => {
    const clientsPage = new ClientsPage(authenticatedPage);
    await use(clientsPage);
  },

  modulesPage: async ({ authenticatedPage }, use) => {
    const modulesPage = new ModulesPage(authenticatedPage);
    await use(modulesPage);
  }
});
