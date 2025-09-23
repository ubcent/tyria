import axios from 'axios';

const API_BASE_URL = process.env.ADMIN_API_URL || 'http://localhost:3001';

// Create axios instance with base configuration
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add response interceptor for handling 401/403 errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized - redirect to login or show auth error
      console.error('Unauthorized access - please login');
      // In a real app, you might redirect to login page here
    } else if (error.response?.status === 403) {
      // Handle forbidden - show permission error
      console.error('Forbidden - insufficient permissions');
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

// Routes API (using new v1 endpoints)
export const routesApi = {
  getAll: async (): Promise<Route[]> => {
    const response = await api.get('/api/v1/routes');
    return response.data;
  },

  getById: async (id: number): Promise<Route> => {
    const response = await api.get(`/api/v1/routes/${id}`);
    return response.data;
  },

  create: async (route: Omit<Route, 'id' | 'created_at' | 'updated_at'>): Promise<Route> => {
    const response = await api.post('/api/v1/routes', route);
    return response.data;
  },

  update: async (id: number, route: Partial<Route>): Promise<Route> => {
    const response = await api.put(`/api/v1/routes/${id}`, route);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/v1/routes/${id}`);
  },
};

// Legacy Routes API (for backward compatibility)
export const legacyRoutesApi = {
  getAll: async (): Promise<ProxyRoute[]> => {
    const response = await api.get('/api/routes');
    return response.data;
  },

  getById: async (id: number): Promise<ProxyRoute> => {
    const response = await api.get(`/api/routes/${id}`);
    return response.data;
  },

  create: async (route: Omit<ProxyRoute, 'id' | 'created_at' | 'updated_at'>): Promise<ProxyRoute> => {
    const response = await api.post('/api/routes', route);
    return response.data;
  },

  update: async (id: number, route: Partial<ProxyRoute>): Promise<ProxyRoute> => {
    const response = await api.put(`/api/routes/${id}`, route);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/routes/${id}`);
  },
};

// API Keys API
export const apiKeysApi = {
  getAll: async (): Promise<ApiKey[]> => {
    const response = await api.get('/api/v1/api-keys');
    return response.data;
  },

  create: async (request: CreateApiKeyRequest): Promise<CreateApiKeyResponse> => {
    const response = await api.post('/api/v1/api-keys', request);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/v1/api-keys/${id}`);
  },
};

// Dashboard API
export const dashboardApi = {
  getStats: async (): Promise<DashboardStats> => {
    const response = await api.get('/api/dashboard/stats');
    return response.data;
  },

  getActivity: async (): Promise<Activity[]> => {
    const response = await api.get('/api/dashboard/activity');
    return response.data;
  },
};

// Health API
export const healthApi = {
  check: async (): Promise<{ status: string; timestamp: string }> => {
    const response = await api.get('/health');
    return response.data;
  },
};

export default api;