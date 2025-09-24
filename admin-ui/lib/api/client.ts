import axios, { AxiosResponse, AxiosError } from 'axios';
import { toast } from '../toast';

// Get base URL from environment - prioritize NEXT_PUBLIC_ for client-side access
const API_BASE_URL = 
  typeof window !== 'undefined' 
    ? (process.env.NEXT_PUBLIC_ADMIN_API_URL || process.env.ADMIN_API_URL || 'http://localhost:3001')
    : (process.env.ADMIN_API_URL || 'http://localhost:3001');

// Create axios instance with base configuration
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Include cookies in requests
});

// Add request interceptor to include JWT from cookies
api.interceptors.request.use(
  (config) => {
    // For client-side requests, get token from document cookies
    if (typeof window !== 'undefined') {
      const cookies = document.cookie.split(';').reduce((acc, cookie) => {
        const [name, value] = cookie.trim().split('=');
        acc[name] = value;
        return acc;
      }, {} as Record<string, string>);
      
      const token = cookies['auth_token'];
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add response interceptor with enhanced error handling and retry logic
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as any;
    
    // Handle specific error cases
    if (error.response?.status === 401) {
      toast.error('Unauthorized', 'Please sign in again');
      // Redirect to login if on client side
      if (typeof window !== 'undefined') {
        window.location.href = '/signin';
      }
    } else if (error.response?.status === 403) {
      toast.error('Forbidden', 'You do not have permission to perform this action');
    } else if (error.response?.status && error.response.status >= 500) {
      // 5xx server errors
      toast.error('Server Error', 'An internal server error occurred. Please try again.');
    } else if (!error.response) {
      // Network errors
      toast.error('Network Error', 'Unable to connect to the server. Please check your connection.');
    }

    // Retry logic for idempotent GET requests
    if (
      error.config?.method === 'get' &&
      !originalRequest._retry &&
      (!error.response || error.response.status >= 500)
    ) {
      originalRequest._retry = true;
      originalRequest._retryCount = (originalRequest._retryCount || 0) + 1;
      
      // Retry up to 3 times with exponential backoff
      if (originalRequest._retryCount <= 3) {
        const delay = Math.pow(2, originalRequest._retryCount - 1) * 1000; // 1s, 2s, 4s
        await new Promise(resolve => setTimeout(resolve, delay));
        return api(originalRequest);
      }
    }

    return Promise.reject(error);
  }
);

// Types
export interface Route {
  id: number;
  name: string;
  match_path: string;
  upstream_url: string;
  headers_json: Record<string, any>;
  auth_mode: 'none' | 'api_key' | 'basic';
  caching_policy_json: Record<string, any>;
  rate_limit_policy_json: Record<string, any>;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProxyRoute {
  id: number;
  path: string;
  target: string;
  methods: string[];
  cache_enabled: boolean;
  cache_ttl: number;
  rate_limit_enabled: boolean;
  rate_limit_requests_per_minute: number;
  rate_limit_burst: number;
  auth_required: boolean;
  auth_keys: string[];
  validation_enabled: boolean;
  validation_request_schema?: string;
  validation_response_schema?: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface ApiKey {
  id: number;
  name: string;
  prefix: string;  // Only the prefix is returned for listing
  created_at: string;
  updated_at: string;
}

export interface CreateApiKeyRequest {
  name: string;
}

export interface CreateApiKeyResponse {
  id: number;
  name: string;
  key: string;  // Full key returned only once
  prefix: string;
  created_at: string;
}

export interface DashboardStats {
  total_requests: number;
  avg_response_time: number;
  success_rate: number;
  cache_hit_rate: number;
  active_routes: number;
  active_api_keys: number;
}

export interface Activity {
  id: number;
  message: string;
  timestamp: string;
  type: 'success' | 'warning' | 'error';
}

export interface CustomDomain {
  id: number;
  tenant_id: number;
  hostname: string;
  verification_token: string;
  status: 'pending' | 'verified' | 'failed';
  created_at: string;
  updated_at: string;
}

export interface CreateCustomDomainRequest {
  hostname: string;
}

export interface VerifyDomainResponse {
  verified: boolean;
  status: string;
  message: string;
}

export interface Tenant {
  id: number;
  name: string;
  plan: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface UserContext {
  user_id: number;
  tenant_id: number;
  email: string;
  role: string;
  tenant: Tenant;
}

// Generic API response wrapper
export interface ApiResponse<T = any> {
  data: T;
  error?: string;
  message?: string;
  code?: string;
}

// Generic HTTP client helpers with typed responses
export const httpClient = {
  /**
   * Generic GET request with typed response
   */
  get: async <T = any>(url: string): Promise<T> => {
    const response: AxiosResponse<T> = await api.get(url);
    return response.data;
  },

  /**
   * Generic POST request with typed response
   */
  post: async <TResponse = any, TRequest = any>(
    url: string,
    data?: TRequest
  ): Promise<TResponse> => {
    const response: AxiosResponse<TResponse> = await api.post(url, data);
    return response.data;
  },

  /**
   * Generic PATCH request with typed response
   */
  patch: async <TResponse = any, TRequest = any>(
    url: string,
    data?: TRequest
  ): Promise<TResponse> => {
    const response: AxiosResponse<TResponse> = await api.patch(url, data);
    return response.data;
  },

  /**
   * Generic PUT request with typed response
   */
  put: async <TResponse = any, TRequest = any>(
    url: string,
    data?: TRequest
  ): Promise<TResponse> => {
    const response: AxiosResponse<TResponse> = await api.put(url, data);
    return response.data;
  },

  /**
   * Generic DELETE request with typed response
   */
  delete: async <T = any>(url: string): Promise<T> => {
    const response: AxiosResponse<T> = await api.delete(url);
    return response.data;
  },
};

// Type-safe API wrapper methods
export const apiClient = {
  /**
   * Make a GET request with automatic error handling and type safety
   */
  get: async <T>(url: string): Promise<T> => {
    try {
      return await httpClient.get<T>(url);
    } catch (error) {
      console.error(`GET ${url} failed:`, error);
      throw error;
    }
  },

  /**
   * Make a POST request with automatic error handling and type safety
   */
  post: async <TResponse, TRequest = any>(
    url: string, 
    data?: TRequest
  ): Promise<TResponse> => {
    try {
      return await httpClient.post<TResponse, TRequest>(url, data);
    } catch (error) {
      console.error(`POST ${url} failed:`, error);
      throw error;
    }
  },

  /**
   * Make a PATCH request with automatic error handling and type safety
   */
  patch: async <TResponse, TRequest = any>(
    url: string,
    data?: TRequest
  ): Promise<TResponse> => {
    try {
      return await httpClient.patch<TResponse, TRequest>(url, data);
    } catch (error) {
      console.error(`PATCH ${url} failed:`, error);
      throw error;
    }
  },

  /**
   * Make a PUT request with automatic error handling and type safety
   */
  put: async <TResponse, TRequest = any>(
    url: string,
    data?: TRequest
  ): Promise<TResponse> => {
    try {
      return await httpClient.put<TResponse, TRequest>(url, data);
    } catch (error) {
      console.error(`PUT ${url} failed:`, error);
      throw error;
    }
  },

  /**
   * Make a DELETE request with automatic error handling and type safety
   */
  delete: async <T = void>(url: string): Promise<T> => {
    try {
      return await httpClient.delete<T>(url);
    } catch (error) {
      console.error(`DELETE ${url} failed:`, error);
      throw error;
    }
  },
};

// Routes API (using new v1 endpoints)
export const routesApi = {
  getAll: async (): Promise<Route[]> => {
    return apiClient.get<Route[]>('/api/v1/routes');
  },

  getById: async (id: number): Promise<Route> => {
    return apiClient.get<Route>(`/api/v1/routes/${id}`);
  },

  create: async (route: Omit<Route, 'id' | 'created_at' | 'updated_at'>): Promise<Route> => {
    return apiClient.post<Route>('/api/v1/routes', route);
  },

  update: async (id: number, route: Partial<Route>): Promise<Route> => {
    return apiClient.put<Route>(`/api/v1/routes/${id}`, route);
  },

  delete: async (id: number): Promise<void> => {
    return apiClient.delete<void>(`/api/v1/routes/${id}`);
  },
};

// Legacy Routes API (for backward compatibility)
export const legacyRoutesApi = {
  getAll: async (): Promise<ProxyRoute[]> => {
    return apiClient.get<ProxyRoute[]>('/api/routes');
  },

  getById: async (id: number): Promise<ProxyRoute> => {
    return apiClient.get<ProxyRoute>(`/api/routes/${id}`);
  },

  create: async (route: Omit<ProxyRoute, 'id' | 'created_at' | 'updated_at'>): Promise<ProxyRoute> => {
    return apiClient.post<ProxyRoute>('/api/routes', route);
  },

  update: async (id: number, route: Partial<ProxyRoute>): Promise<ProxyRoute> => {
    return apiClient.put<ProxyRoute>(`/api/routes/${id}`, route);
  },

  delete: async (id: number): Promise<void> => {
    return apiClient.delete<void>(`/api/routes/${id}`);
  },
};

// API Keys API
export const apiKeysApi = {
  getAll: async (): Promise<ApiKey[]> => {
    return apiClient.get<ApiKey[]>('/api/v1/api-keys');
  },

  create: async (request: CreateApiKeyRequest): Promise<CreateApiKeyResponse> => {
    return apiClient.post<CreateApiKeyResponse>('/api/v1/api-keys', request);
  },

  delete: async (id: number): Promise<void> => {
    return apiClient.delete<void>(`/api/v1/api-keys/${id}`);
  },
};

// Custom Domains API
export const customDomainsApi = {
  getAll: async (): Promise<CustomDomain[]> => {
    return apiClient.get<CustomDomain[]>('/api/v1/domains');
  },

  getById: async (id: number): Promise<CustomDomain> => {
    return apiClient.get<CustomDomain>(`/api/v1/domains/${id}`);
  },

  create: async (request: CreateCustomDomainRequest): Promise<CustomDomain> => {
    return apiClient.post<CustomDomain>('/api/v1/domains', request);
  },

  delete: async (id: number): Promise<void> => {
    return apiClient.delete<void>(`/api/v1/domains/${id}`);
  },

  verify: async (id: number): Promise<VerifyDomainResponse> => {
    return apiClient.post<VerifyDomainResponse>(`/api/v1/domains/${id}/verify`);
  },
};

// Dashboard API
export const dashboardApi = {
  getStats: async (): Promise<DashboardStats> => {
    return apiClient.get<DashboardStats>('/api/dashboard/stats');
  },

  getActivity: async (): Promise<Activity[]> => {
    return apiClient.get<Activity[]>('/api/dashboard/activity');
  },
};

// Health API
export const healthApi = {
  check: async (): Promise<{ status: string; timestamp: string }> => {
    return apiClient.get<{ status: string; timestamp: string }>('/health');
  },
};

// Tenants API
export const tenantsApi = {
  getAll: async (): Promise<Tenant[]> => {
    return apiClient.get<Tenant[]>('/api/v1/tenants');
  },

  getById: async (id: number): Promise<Tenant> => {
    return apiClient.get<Tenant>(`/api/v1/tenants/${id}`);
  },
};

// Auth API
export const authApi = {
  getProfile: async (): Promise<UserContext> => {
    return apiClient.get<UserContext>('/api/auth/profile');
  },
};

export default api;