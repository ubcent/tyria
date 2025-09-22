-- PostgreSQL schema for Edge.link Admin API

-- Tenants table for multi-tenant support
CREATE TABLE IF NOT EXISTS tenants (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  domain VARCHAR(255) UNIQUE,
  plan VARCHAR(50) DEFAULT 'free',
  enabled BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Domains table for domain linking
CREATE TABLE IF NOT EXISTS domains (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE,
  domain VARCHAR(255) NOT NULL,
  proxy_url VARCHAR(500) NOT NULL,
  verified BOOLEAN DEFAULT false,
  verify_token VARCHAR(255) NOT NULL,
  ssl_enabled BOOLEAN DEFAULT false,
  ssl_cert_path VARCHAR(500),
  ssl_key_path VARCHAR(500),
  enabled BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(tenant_id, domain)
);

-- Users table for authentication (optional for future use)
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  role VARCHAR(50) DEFAULT 'admin',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Proxy routes configuration
CREATE TABLE IF NOT EXISTS proxy_routes (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE,
  path VARCHAR(255) NOT NULL,
  target VARCHAR(500) NOT NULL,
  methods TEXT[] DEFAULT '{"GET"}',
  cache_enabled BOOLEAN DEFAULT false,
  cache_ttl INTEGER DEFAULT 300, -- seconds
  rate_limit_enabled BOOLEAN DEFAULT false,
  rate_limit_rate INTEGER DEFAULT 100,
  rate_limit_burst INTEGER DEFAULT 10,
  rate_limit_period INTEGER DEFAULT 60, -- seconds
  rate_limit_per_client BOOLEAN DEFAULT true,
  auth_required BOOLEAN DEFAULT false,
  auth_keys TEXT[] DEFAULT '{}',
  validation_enabled BOOLEAN DEFAULT false,
  validation_request_schema TEXT,
  validation_response_schema TEXT,
  enabled BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- API keys management
CREATE TABLE IF NOT EXISTS api_keys (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE,
  key_value VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  permissions TEXT[] DEFAULT '{}',
  rate_limit INTEGER DEFAULT 1000,
  enabled BOOLEAN DEFAULT true,
  expires_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Metrics and logs (simplified for MVP)
CREATE TABLE IF NOT EXISTS request_logs (
  id SERIAL PRIMARY KEY,
  route_path VARCHAR(255),
  method VARCHAR(10),
  status_code INTEGER,
  response_time_ms INTEGER,
  client_ip VARCHAR(45),
  api_key_used VARCHAR(255),
  cache_hit BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_proxy_routes_path ON proxy_routes(path);
CREATE INDEX IF NOT EXISTS idx_api_keys_value ON api_keys(key_value);
CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_route_path ON request_logs(route_path);

-- Insert default admin user (password: admin123)
INSERT INTO users (email, password_hash, name, role) 
VALUES ('admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin User', 'admin')
ON CONFLICT (email) DO NOTHING;

-- Insert sample route configurations
INSERT INTO proxy_routes (path, target, methods, cache_enabled, cache_ttl, rate_limit_enabled, rate_limit_rate)
VALUES 
  ('/api/v1/posts', 'https://jsonplaceholder.typicode.com', '{"GET","POST"}', true, 300, true, 100),
  ('/api/secure/', 'https://httpbin.org', '{"GET","POST","PUT","DELETE"}', true, 120, true, 50)
ON CONFLICT DO NOTHING;

-- Insert sample API keys
INSERT INTO api_keys (key_value, name, permissions, rate_limit)
VALUES 
  ('demo-key-12345', 'demo-client', '{"proxy.*"}', 1000),
  ('admin-key-67890', 'admin-client', '{"proxy.*","admin.*"}', 5000)
ON CONFLICT (key_value) DO NOTHING;