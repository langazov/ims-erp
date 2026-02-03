// Invoices API
import api from './index';

export interface Invoice {
  id: string;
  tenantId: string;
  invoiceNumber: string;
  clientId: string;
  clientName: string;
  status: InvoiceStatus;
  items: InvoiceItem[];
  subtotal: string;
  tax: string;
  total: string;
  dueDate: string;
  paidDate: string | null;
  notes: string;
  createdAt: string;
  updatedAt: string;
}

export type InvoiceStatus = 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled';

export interface InvoiceItem {
  description: string;
  quantity: number;
  unitPrice: string;
  total: string;
}

export interface InvoiceFilter {
  status?: InvoiceStatus;
  clientId?: string;
  search?: string;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface InvoiceListResponse {
  data: Invoice[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

const BASE_PATH = '/invoices';

export async function getInvoices(filter?: InvoiceFilter): Promise<InvoiceListResponse> {
  const params = filter as Record<string, string | number | boolean | undefined>;
  return api.get(`${BASE_PATH}`, params);
}

export async function getInvoiceById(id: string): Promise<Invoice> {
  return api.get(`${BASE_PATH}/${id}`);
}

export async function createInvoice(data: {
  clientId: string;
  items: InvoiceItem[];
  dueDate?: string;
  notes?: string;
}): Promise<Invoice> {
  return api.post(`${BASE_PATH}`, data);
}

export async function updateInvoice(id: string, data: Partial<{
  items: InvoiceItem[];
  dueDate: string;
  notes: string;
  status: InvoiceStatus;
}>): Promise<Invoice> {
  return api.patch(`${BASE_PATH}/${id}`, data);
}

export async function deleteInvoice(id: string): Promise<void> {
  return api.delete(`${BASE_PATH}/${id}`);
}

export async function markInvoiceAsPaid(id: string): Promise<Invoice> {
  return api.post(`${BASE_PATH}/${id}/pay`, {});
}

export async function getInvoiceStats(): Promise<{
  total: number;
  draft: number;
  sent: number;
  paid: number;
  overdue: number;
  totalAmount: string;
  paidAmount: string;
}> {
  return api.get(`${BASE_PATH}/stats`);
}
