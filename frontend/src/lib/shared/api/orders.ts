// Orders API
import api from './index';

export interface Order {
  id: string;
  tenantId: string;
  orderNumber: string;
  clientId: string;
  clientName: string;
  status: OrderStatus;
  items: OrderItem[];
  subtotal: string;
  tax: string;
  total: string;
  shippingAddress: Address;
  notes: string;
  createdAt: string;
  updatedAt: string;
}

export type OrderStatus = 'pending' | 'confirmed' | 'processing' | 'shipped' | 'delivered' | 'cancelled';

export interface OrderItem {
  productId: string;
  productName: string;
  quantity: number;
  unitPrice: string;
  total: string;
}

export interface Address {
  street: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
}

export interface OrderFilter {
  status?: OrderStatus;
  clientId?: string;
  search?: string;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface OrderListResponse {
  data: Order[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

const BASE_PATH = '/orders';

export async function getOrders(filter?: OrderFilter): Promise<OrderListResponse> {
  const params = filter as Record<string, string | number | boolean | undefined>;
  return api.get(`${BASE_PATH}`, params);
}

export async function getOrderById(id: string): Promise<Order> {
  return api.get(`${BASE_PATH}/${id}`);
}

export async function createOrder(data: {
  clientId: string;
  items: OrderItem[];
  shippingAddress?: Address;
  notes?: string;
}): Promise<Order> {
  return api.post(`${BASE_PATH}`, data);
}

export async function updateOrderStatus(id: string, status: OrderStatus): Promise<Order> {
  return api.patch(`${BASE_PATH}/${id}`, { status });
}

export async function deleteOrder(id: string): Promise<void> {
  return api.delete(`${BASE_PATH}/${id}`);
}

export async function getOrderStats(): Promise<{
  pending: number;
  processing: number;
  shipped: number;
  delivered: number;
  total: number;
}> {
  try {
    const response = await api.get<any>('/orders/report/summary');
    return {
      pending: response.pending || 0,
      processing: response.processing || 0,
      shipped: response.shipped || 0,
      delivered: response.delivered || 0,
      total: response.total || 0
    };
  } catch {
    return { pending: 0, processing: 0, shipped: 0, delivered: 0, total: 0 };
  }
}
