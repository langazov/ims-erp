// Payments API
import api from './index';

export type PaymentStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'refunded' | 'cancelled';
export type PaymentMethod = 'card' | 'bank_transfer' | 'cash' | 'check' | 'stripe' | 'paypal';

export interface Payment {
  id: string;
  tenantId: string;
  paymentNumber: string;
  invoiceId: string;
  invoiceNumber: string;
  clientId: string;
  clientName: string;
  amount: string;
  currency: string;
  method: PaymentMethod;
  status: PaymentStatus;
  gatewayReference: string | null;
  processorResponse: string | null;
  notes: string;
  paidAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface PaymentFilter {
  status?: PaymentStatus;
  method?: PaymentMethod;
  clientId?: string;
  invoiceId?: string;
  search?: string;
  dateFrom?: string;
  dateTo?: string;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface PaymentListResponse {
  data: Payment[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface CreatePaymentRequest {
  invoiceId: string;
  amount: number;
  currency?: string;
  method: PaymentMethod;
  notes?: string;
  paidAt?: string;
}

export interface RefundPaymentRequest {
  amount: number;
  reason: string;
}

export interface PaymentStats {
  total: number;
  pending: number;
  completed: number;
  failed: number;
  totalAmountCollected: string;
  totalAmountToday: string;
  totalAmountThisMonth: string;
}

const BASE_PATH = '/payments';

export async function getPayments(filter?: PaymentFilter): Promise<PaymentListResponse> {
  const params = filter as Record<string, string | number | boolean | undefined>;
  return api.get(`${BASE_PATH}`, params);
}

export async function getPaymentById(id: string): Promise<Payment> {
  return api.get(`${BASE_PATH}/${id}`);
}

export async function createPayment(data: CreatePaymentRequest): Promise<Payment> {
  return api.post(`${BASE_PATH}`, data);
}

export async function processPayment(id: string): Promise<Payment> {
  return api.post(`${BASE_PATH}/${id}/process`, {});
}

export async function refundPayment(id: string, data: RefundPaymentRequest): Promise<Payment> {
  return api.post(`${BASE_PATH}/${id}/refund`, data);
}

export async function cancelPayment(id: string): Promise<Payment> {
  return api.post(`${BASE_PATH}/${id}/cancel`, {});
}

export async function getPaymentStats(): Promise<PaymentStats> {
  return api.get(`${BASE_PATH}/stats`);
}

export async function getPaymentsByInvoice(invoiceId: string): Promise<Payment[]> {
  return api.get(`${BASE_PATH}/invoice/${invoiceId}`);
}
