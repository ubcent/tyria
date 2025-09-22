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

// Types
export interface ProxyRoute {
  id: number;
  path: string;
  target: string;
  methods: string[];
  cache_enabled: boolean;
  cache_ttl: number;
  rate_limit_enabled: boolean;
  rate_limit_rate: number;
  rate_limit_burst: number;
  rate_limit_period: number;
  rate_limit_per_client: boolean;
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
  key_value: string;
  name: string;
  permissions: string[];
  rate_limit: number;
  enabled: boolean;
  expires_at?: string;
  created_at: string;
  updated_at: string;
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

// Routes API
export const routesApi = {
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
    const response = await api.get('/api/keys');
    return response.data;
  },

  getById: async (id: number): Promise<ApiKey> => {
    const response = await api.get(`/api/keys/${id}`);
    return response.data;
  },

  create: async (key: Omit<ApiKey, 'id' | 'created_at' | 'updated_at'>): Promise<ApiKey> => {
    const response = await api.post('/api/keys', key);
    return response.data;
  },

  update: async (id: number, key: Partial<ApiKey>): Promise<ApiKey> => {
    const response = await api.put(`/api/keys/${id}`, key);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/keys/${id}`);
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