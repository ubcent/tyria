import { Pool } from 'pg';

const pool = new Pool({
  connectionString: process.env.POSTGRES_URL,
  ssl: process.env.NODE_ENV === 'production' ? { rejectUnauthorized: false } : false,
});

export default pool;

// Types for database entities
export interface User {
  id: number;
  email: string;
  name: string;
  role: string;
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

export interface RequestLog {
  id: number;
  route_path: string;
  method: string;
  status_code: number;
  response_time_ms: number;
  client_ip: string;
  api_key_used?: string;
  cache_hit: boolean;
  created_at: string;
}

// Database helper functions
export class DB {
  // Routes
  static async getRoutes(): Promise<ProxyRoute[]> {
    const result = await pool.query('SELECT * FROM proxy_routes ORDER BY created_at DESC');
    return result.rows;
  }

  static async getRoute(id: number): Promise<ProxyRoute | null> {
    const result = await pool.query('SELECT * FROM proxy_routes WHERE id = $1', [id]);
    return result.rows[0] || null;
  }

  static async createRoute(route: Omit<ProxyRoute, 'id' | 'created_at' | 'updated_at'>): Promise<ProxyRoute> {
    const result = await pool.query(`
      INSERT INTO proxy_routes (
        path, target, methods, cache_enabled, cache_ttl,
        rate_limit_enabled, rate_limit_rate, rate_limit_burst, rate_limit_period,
        rate_limit_per_client, auth_required, auth_keys, validation_enabled,
        validation_request_schema, validation_response_schema, enabled
      ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
      RETURNING *
    `, [
      route.path, route.target, route.methods, route.cache_enabled, route.cache_ttl,
      route.rate_limit_enabled, route.rate_limit_rate, route.rate_limit_burst,
      route.rate_limit_period, route.rate_limit_per_client, route.auth_required,
      route.auth_keys, route.validation_enabled, route.validation_request_schema,
      route.validation_response_schema, route.enabled
    ]);
    return result.rows[0];
  }

  static async updateRoute(id: number, route: Partial<ProxyRoute>): Promise<ProxyRoute | null> {
    const fields = Object.keys(route).filter(key => key !== 'id' && key !== 'created_at');
    const values = fields.map(field => route[field as keyof ProxyRoute]);
    const setClause = fields.map((field, index) => `${field} = $${index + 2}`).join(', ');
    
    if (fields.length === 0) return null;

    const result = await pool.query(`
      UPDATE proxy_routes 
      SET ${setClause}, updated_at = NOW() 
      WHERE id = $1 
      RETURNING *
    `, [id, ...values]);
    
    return result.rows[0] || null;
  }

  static async deleteRoute(id: number): Promise<boolean> {
    const result = await pool.query('DELETE FROM proxy_routes WHERE id = $1', [id]);
    return result.rowCount > 0;
  }

  // API Keys
  static async getApiKeys(): Promise<ApiKey[]> {
    const result = await pool.query('SELECT * FROM api_keys ORDER BY created_at DESC');
    return result.rows;
  }

  static async getApiKey(id: number): Promise<ApiKey | null> {
    const result = await pool.query('SELECT * FROM api_keys WHERE id = $1', [id]);
    return result.rows[0] || null;
  }

  static async createApiKey(apiKey: Omit<ApiKey, 'id' | 'created_at' | 'updated_at'>): Promise<ApiKey> {
    const result = await pool.query(`
      INSERT INTO api_keys (key_value, name, permissions, rate_limit, enabled, expires_at)
      VALUES ($1, $2, $3, $4, $5, $6)
      RETURNING *
    `, [
      apiKey.key_value, apiKey.name, apiKey.permissions, 
      apiKey.rate_limit, apiKey.enabled, apiKey.expires_at || null
    ]);
    return result.rows[0];
  }

  static async updateApiKey(id: number, apiKey: Partial<ApiKey>): Promise<ApiKey | null> {
    const fields = Object.keys(apiKey).filter(key => key !== 'id' && key !== 'created_at');
    const values = fields.map(field => apiKey[field as keyof ApiKey]);
    const setClause = fields.map((field, index) => `${field} = $${index + 2}`).join(', ');
    
    if (fields.length === 0) return null;

    const result = await pool.query(`
      UPDATE api_keys 
      SET ${setClause}, updated_at = NOW() 
      WHERE id = $1 
      RETURNING *
    `, [id, ...values]);
    
    return result.rows[0] || null;
  }

  static async deleteApiKey(id: number): Promise<boolean> {
    const result = await pool.query('DELETE FROM api_keys WHERE id = $1', [id]);
    return result.rowCount > 0;
  }

  // Users
  static async getUserByEmail(email: string): Promise<User | null> {
    const result = await pool.query('SELECT * FROM users WHERE email = $1', [email]);
    return result.rows[0] || null;
  }

  // Logs
  static async getRequestLogs(limit = 100, offset = 0): Promise<RequestLog[]> {
    const result = await pool.query(`
      SELECT * FROM request_logs 
      ORDER BY created_at DESC 
      LIMIT $1 OFFSET $2
    `, [limit, offset]);
    return result.rows;
  }

  static async getMetrics(timeframe = '24h'): Promise<any> {
    const timeCondition = timeframe === '1h' ? 'created_at > NOW() - INTERVAL \'1 hour\'' :
                         timeframe === '24h' ? 'created_at > NOW() - INTERVAL \'1 day\'' :
                         'created_at > NOW() - INTERVAL \'7 days\'';

    const result = await pool.query(`
      SELECT 
        COUNT(*) as total_requests,
        AVG(response_time_ms) as avg_response_time,
        COUNT(*) FILTER (WHERE status_code >= 200 AND status_code < 300) as success_count,
        COUNT(*) FILTER (WHERE status_code >= 400) as error_count,
        COUNT(*) FILTER (WHERE cache_hit = true) as cache_hits,
        route_path,
        COUNT(*) as route_requests
      FROM request_logs 
      WHERE ${timeCondition}
      GROUP BY route_path
      ORDER BY route_requests DESC
    `);
    
    return result.rows;
  }
}