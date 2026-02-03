import type { Page } from '@playwright/test';

export class ModulesPage {
  constructor(private page: Page) {}

  async goto(): Promise<void> {
    await this.page.goto('/modules');
  }

  async getModuleNames(): Promise<string[]> {
    return this.page.locator('.module-name').allTextContents();
  }

  async getModuleCount(): Promise<number> {
    return this.page.locator('.module-card').count();
  }

  async clickCategory(category: string): Promise<void> {
    await this.page.locator('.category-header', { hasText: category }).click();
  }

  async clickModule(moduleName: string): Promise<void> {
    await this.page.locator('.module-card', { hasText: moduleName }).click();
  }

  async searchModules(query: string): Promise<void> {
    await this.page.fill('.form-input', query);
    await this.page.press('.form-input', 'Enter');
  }

  async expandAll(): Promise<void> {
    await this.page.click('button:has-text("Expand All")');
  }

  async collapseAll(): Promise<void> {
    await this.page.click('button:has-text("Collapse All")');
  }

  async getTotalModules(): Promise<number> {
    const text = await this.page.locator('.stat-value').first().textContent();
    return parseInt(text || '0');
  }

  async getEnabledCount(): Promise<number> {
    const text = await this.page.locator('.stat-item').nth(1).locator('.stat-value').textContent();
    return parseInt(text || '0');
  }

  async getCategoriesCount(): Promise<number> {
    const text = await this.page.locator('.stat-item').nth(2).locator('.stat-value').textContent();
    return parseInt(text || '0');
  }

  async isCategoryExpanded(category: string): Promise<boolean> {
    const header = this.page.locator('.category-header', { hasText: category });
    const content = header.locator('..').locator('.category-modules');
    return await content.isVisible();
  }
}
