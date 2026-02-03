// API Client Base
const API_BASE = import.meta.env.VITE_API_URL || '/api';

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  body?: unknown;
  headers?: Record<string, string>;
  params?: Record<string, string | number | boolean | undefined>;
}

class ApiClient {
  private baseUrl: string;
  private defaultHeaders: Record<string, string>;

  constructor(baseUrl = API_BASE) {
    this.baseUrl = baseUrl;
    this.defaultHeaders = {
      'Content-Type': 'application/json',
    };
  }

  private buildUrl(endpoint: string, params?: Record<string, string | number | boolean | undefined>): string {
    let url: URL;
    if (endpoint.startsWith('http')) {
      url = new URL(endpoint);
    } else {
      const baseUrl = this.baseUrl.endsWith('/') ? this.baseUrl.slice(0, -1) : this.baseUrl;
      const normalizedEndpoint = endpoint.startsWith('/') ? endpoint.slice(1) : endpoint;
      url = new URL(`${baseUrl}/${normalizedEndpoint}`, 'http://localhost');
    }
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== '') {
          url.searchParams.set(key, String(value));
        }
      });
    }
    return url.toString();
  }

  private async request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    const { method = 'GET', body, headers = {}, params } = options;
    const url = this.buildUrl(endpoint, params);

    const response = await fetch(url, {
      method,
      headers: { ...this.defaultHeaders, ...headers },
      body: body ? JSON.stringify(body) : undefined,
      credentials: 'include',
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'An error occurred' }));
      throw new Error(error.message || `HTTP error! status: ${response.status}`);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  }

  get<T>(endpoint: string, params?: Record<string, string | number | boolean | undefined>): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET', params });
  }

  post<T>(endpoint: string, body?: unknown, options?: Omit<RequestOptions, 'method' | 'body'>): Promise<T> {
    return this.request<T>(endpoint, { method: 'POST', body, ...options });
  }

  put<T>(endpoint: string, body?: unknown, options?: Omit<RequestOptions, 'method' | 'body'>): Promise<T> {
    return this.request<T>(endpoint, { method: 'PUT', body, ...options });
  }

  patch<T>(endpoint: string, body?: unknown, options?: Omit<RequestOptions, 'method' | 'body'>): Promise<T> {
    return this.request<T>(endpoint, { method: 'PATCH', body, ...options });
  }

  delete<T>(endpoint: string, options?: Omit<RequestOptions, 'method'>): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE', ...options });
  }

  setAuthToken(token: string) {
    this.defaultHeaders['Authorization'] = `Bearer ${token}`;
  }

  clearAuthToken() {
    delete this.defaultHeaders['Authorization'];
  }
}

export const api = new ApiClient();
export default api;
