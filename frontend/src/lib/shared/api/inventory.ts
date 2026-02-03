// Inventory API
import api from './index';

export interface InventoryItem {
  id: string;
  tenantId: string;
  productId: string;
  productName: string;
  sku: string;
  warehouseId: string;
  warehouseName: string;
  quantity: number;
  reservedQuantity: number;
  availableQuantity: number;
  location: string;
  lastCountDate: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface InventoryFilter {
  warehouseId?: string;
  productId?: string;
  search?: string;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface InventoryListResponse {
  data: InventoryItem[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface InventoryStats {
  totalItems: number;
  totalQuantity: number;
  lowStock: number;
  outOfStock: number;
}

const BASE_PATH = '/inventory/items';

export async function getInventory(filter?: InventoryFilter): Promise<InventoryListResponse> {
  const params = filter as Record<string, string | number | boolean | undefined>;
  return api.get(`${BASE_PATH}`, params);
}

export async function getInventoryById(id: string): Promise<InventoryItem> {
  return api.get(`${BASE_PATH}/${id}`);
}

export async function getInventoryByProduct(productId: string): Promise<InventoryItem[]> {
  return api.get(`${BASE_PATH}/product/${productId}`);
}

export async function adjustInventory(id: string, data: {
  quantity: number;
  reason: string;
  reference?: string;
}): Promise<InventoryItem> {
  return api.post(`${BASE_PATH}/${id}/adjust`, data);
}

export async function reserveInventory(id: string, quantity: number): Promise<InventoryItem> {
  return api.post(`${BASE_PATH}/${id}/reserve`, { quantity });
}

export async function releaseReservation(id: string, quantity: number): Promise<InventoryItem> {
  return api.post(`${BASE_PATH}/${id}/release`, { quantity });
}

export async function getInventoryStats(): Promise<InventoryStats> {
  const response = await api.get<{ items: InventoryItem[] }>('/inventory/reports/stock');
  const items = response.items || [];
  const lowStock = items.filter((item: InventoryItem) => item.availableQuantity > 0 && item.availableQuantity < 10).length;
  const outOfStock = items.filter((item: InventoryItem) => item.availableQuantity === 0).length;
  
  return {
    totalItems: items.length,
    totalQuantity: items.reduce((sum: number, item: InventoryItem) => sum + item.quantity, 0),
    lowStock,
    outOfStock
  };
}

export async function getLowStockItems(): Promise<InventoryItem[]> {
  return api.get(`${BASE_PATH}/low-stock`);
}
