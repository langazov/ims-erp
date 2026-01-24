// Types
export interface Client {
  id: string;
  tenantId: string;
  code: string;
  name: string;
  email: string;
  phone: string;
  status: ClientStatus;
  creditLimit: string;
  currentBalance: string;
  billingAddress: Address;
  shippingAddresses: Address[];
  tags: string[];
  createdAt: string;
  updatedAt: string;
}

export type ClientStatus = 'active' | 'inactive' | 'suspended' | 'merged';

export interface Address {
  street: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
}

export interface CreateClientRequest {
  name: string;
  email: string;
  phone?: string;
  billingAddress?: Address;
  creditLimit?: number;
}

export interface UpdateClientRequest {
  name?: string;
  email?: string;
  phone?: string;
  billingAddress?: Address;
  creditLimit?: number;
  status?: ClientStatus;
}

export interface ClientFilter {
  search?: string;
  status?: ClientStatus;
  tags?: string[];
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface ClientListResponse {
  data: Client[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// API Functions
import api from './index';

export async function getClients(filter?: ClientFilter): Promise<ClientListResponse> {
  const params = filter as Record<string, string | number | boolean | undefined>;
  return api.get('/clients', params);
}

export async function getClientById(id: string): Promise<Client> {
  return api.get(`/clients/${id}`);
}

export async function createClient(data: CreateClientRequest): Promise<Client> {
  return api.post('/clients', data);
}

export async function updateClient(id: string, data: UpdateClientRequest): Promise<Client> {
  return api.patch(`/clients/${id}`, data);
}

export async function deleteClient(id: string): Promise<void> {
  return api.delete(`/clients/${id}`);
}

export async function getClientCreditStatus(id: string): Promise<{
  availableCredit: string;
  utilization: number;
  status: string;
}> {
  return api.get(`/clients/${id}/credit-status`);
}

export async function assignCreditLimit(
  id: string,
  limit: number
): Promise<Client> {
  return api.post(`/clients/${id}/credit-limit`, { limit });
}
