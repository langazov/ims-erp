export interface User {
  id: string;
  tenantId: string;
  email: string;
  firstName: string;
  lastName: string;
  phone?: string;
  role: string;
  tenantRole: string;
  permissions: string[];
  status: UserStatus;
  mfaEnabled: boolean;
  lastLoginAt?: string;
  createdAt: string;
  updatedAt: string;
}

export type UserStatus = 'active' | 'inactive' | 'suspended' | 'locked';

export interface TokenPair {
  accessToken: string;
  refreshToken: string;
  tokenType: string;
  expiresIn: number;
  expiresAt: string;
}

export interface LoginResponse {
  user: User;
  tokens: TokenPair;
  sessionId: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  phone?: string;
}

export interface AuthState {
  user: User | null;
  tokens: TokenPair | null;
  sessionId: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

import api from './index';

const BASE_PATH = '/auth';

export async function login(tenantId: string, data: LoginRequest): Promise<LoginResponse> {
  const response = await api.post<LoginResponse>(`${BASE_PATH}/login?tenantId=${tenantId}`, data);
  return response;
}

export async function register(tenantId: string, data: RegisterRequest): Promise<User> {
  const response = await api.post<User>(`${BASE_PATH}/register?tenantId=${tenantId}`, data);
  return response;
}

export async function logout(userId: string, sessionId: string): Promise<void> {
  await api.post(`${BASE_PATH}/logout?userId=${userId}&sessionId=${sessionId}`);
}

export async function refreshToken(refreshToken: string): Promise<TokenPair> {
  const response = await api.post<TokenPair>(`${BASE_PATH}/refresh`, { refreshToken });
  return response;
}

export async function getCurrentUser(userId: string): Promise<User> {
  const response = await api.get<User>(`${BASE_PATH}/me?userId=${userId}`);
  return response;
}

export async function changePassword(
  userId: string,
  currentPassword: string,
  newPassword: string
): Promise<void> {
  await api.post(`${BASE_PATH}/change-password`, {
    userId,
    currentPassword,
    newPassword,
  });
}
