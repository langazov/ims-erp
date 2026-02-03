// Users API
import api from './index';

export interface User {
  id: string;
  tenantId: string;
  email: string;
  name: string;
  role: UserRole;
  status: UserStatus;
  avatar: string | null;
  lastLogin: string | null;
  createdAt: string;
  updatedAt: string;
}

export type UserRole = 'admin' | 'manager' | 'user' | 'viewer';
export type UserStatus = 'active' | 'inactive' | 'suspended';

export interface UserFilter {
  role?: UserRole;
  status?: UserStatus;
  search?: string;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface UserListResponse {
  data: User[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

const BASE_PATH = '/users';

export async function getUsers(filter?: UserFilter): Promise<UserListResponse> {
  try {
    const params = filter as Record<string, string | number | boolean | undefined>;
    return await api.get(`${BASE_PATH}`, params);
  } catch {
    return { data: [], total: 0, page: 1, pageSize: filter?.pageSize || 20, totalPages: 0 };
  }
}

export async function getUserById(id: string): Promise<User> {
  return api.get(`${BASE_PATH}/${id}`);
}

export async function createUser(data: {
  email: string;
  name: string;
  role: UserRole;
}): Promise<User> {
  return api.post(`${BASE_PATH}`, data);
}

export async function updateUser(id: string, data: Partial<{
  name: string;
  role: UserRole;
  status: UserStatus;
}>): Promise<User> {
  return api.patch(`${BASE_PATH}/${id}`, data);
}

export async function deleteUser(id: string): Promise<void> {
  return api.delete(`${BASE_PATH}/${id}`);
}

export async function getUserStats(): Promise<{
  total: number;
  active: number;
  inactive: number;
  admin: number;
}> {
  return api.get(`${BASE_PATH}/stats`);
}
