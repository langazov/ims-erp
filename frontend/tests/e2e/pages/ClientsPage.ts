import type { Page } from '@playwright/test';

export class ClientsPage {
  constructor(private page: Page) {}

  async goto(): Promise<void> {
    await this.page.goto('/clients');
  }

  async createClient(data: {
    name: string;
    email: string;
    phone?: string;
    creditLimit?: string;
  }): Promise<void> {
    await this.page.click('[data-testid="add-client-button"]');
    await this.page.fill('[data-testid="name-input"]', data.name);
    await this.page.fill('[data-testid="email-input"]', data.email);
    if (data.phone) {
      await this.page.fill('[data-testid="phone-input"]', data.phone);
    }
    if (data.creditLimit) {
      await this.page.fill('[data-testid="creditLimit-input"]', data.creditLimit);
    }
    await this.page.click('[data-testid="submit-button"]');
  }

  async searchClients(query: string): Promise<void> {
    await this.page.fill('[data-testid="search-input"]', query);
    await this.page.click('[data-testid="search-button"]');
  }

  async getClientNames(): Promise<string[]> {
    return this.page.locator('[data-testid="client-name"]').allTextContents();
  }

  async getClientCount(): Promise<number> {
    return this.page.locator('[data-testid="client-row"]').count();
  }

  async filterByStatus(status: string): Promise<void> {
    await this.page.selectOption('[data-testid="status-filter"]', status);
    await this.page.click('[data-testid="apply-filters"]');
  }

  async clickClientRow(index: number = 0): Promise<void> {
    await this.page.locator('[data-testid="client-row"]').nth(index).click();
  }

  async editClient(index: number = 0): Promise<void> {
    await this.page.locator('[data-testid="client-row"]').nth(index).hover();
    await this.page.locator('[data-testid="edit-button"]').nth(index).click();
  }

  async deleteClient(index: number = 0): Promise<void> {
    await this.page.locator('[data-testid="client-row"]').nth(index).hover();
    await this.page.locator('[data-testid="delete-button"]').nth(index).click();
    await this.page.click('[data-testid="confirm-delete"]');
  }

  async navigateToNewClient(): Promise<void> {
    await this.page.click('[data-testid="new-client-button"]');
  }

  async updateClient(data: {
    name?: string;
    email?: string;
    phone?: string;
    creditLimit?: string;
  }): Promise<void> {
    if (data.name) {
      await this.page.fill('[data-testid="name-input"]', data.name);
    }
    if (data.email) {
      await this.page.fill('[data-testid="email-input"]', data.email);
    }
    if (data.phone) {
      await this.page.fill('[data-testid="phone-input"]', data.phone);
    }
    if (data.creditLimit) {
      await this.page.fill('[data-testid="creditLimit-input"]', data.creditLimit);
    }
    await this.page.click('[data-testid="save-button"]');
  }

  async goBackToClients(): Promise<void> {
    await this.page.click('[data-testid="back-button"]');
  }

  async getToastMessage(): Promise<string> {
    return this.page.locator('[data-testid="toast"]').textContent();
  }
}
