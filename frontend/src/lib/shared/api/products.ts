// Products API
import api from './index';

export interface Product {
  id: string;
  tenantId: string;
  sku: string;
  name: string;
  description: string;
  category: string;
  price: string;
  cost: string;
  stockQuantity: number;
  lowStockThreshold: number;
  unit: string;
  images: string[];
  status: ProductStatus;
  createdAt: string;
  updatedAt: string;
}

export type ProductStatus = 'active' | 'inactive' | 'discontinued';

export interface ProductFilter {
  category?: string;
  status?: ProductStatus;
  search?: string;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface ProductListResponse {
  data: Product[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

const BASE_PATH = '/products';

export async function getProducts(filter?: ProductFilter): Promise<ProductListResponse> {
  const params = filter as Record<string, string | number | boolean | undefined>;
  return api.get(`${BASE_PATH}`, params);
}

export async function getProductById(id: string): Promise<Product> {
  return api.get(`${BASE_PATH}/${id}`);
}

export async function createProduct(data: {
  sku: string;
  name: string;
  description?: string;
  category?: string;
  price: number;
  cost?: number;
  stockQuantity?: number;
  lowStockThreshold?: number;
  unit?: string;
}): Promise<Product> {
  return api.post(`${BASE_PATH}`, data);
}

export async function updateProduct(id: string, data: Partial<{
  name: string;
  description: string;
  category: string;
  price: number;
  cost: number;
  stockQuantity: number;
  lowStockThreshold: number;
  unit: string;
  status: ProductStatus;
}>): Promise<Product> {
  return api.patch(`${BASE_PATH}/${id}`, data);
}

export async function deleteProduct(id: string): Promise<void> {
  return api.delete(`${BASE_PATH}/${id}`);
}

export async function getProductStats(): Promise<{
  total: number;
  active: number;
  lowStock: number;
  outOfStock: number;
}> {
  return api.get(`${BASE_PATH}/stats`);
}
