-- PostgreSQL schema for Edge.link Admin API

-- Tenants table for multi-tenant support
CREATE TABLE IF NOT EXISTS tenants (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  plan VARCHAR(50) DEFAULT 'free' CHECK (plan IN ('free', 'pro', 'enterprise')),
  status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'canceled')),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Users table for authentication (optional for future use)
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  email VARCHAR(255) UNIQUE NOT NULL,
  hashed_password VARCHAR(255) NOT NULL,
  role VARCHAR(50) DEFAULT 'viewer' CHECK (role IN ('owner', 'admin', 'viewer')),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Proxy routes configuration
CREATE TABLE IF NOT EXISTS routes (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  match_path VARCHAR(255) NOT NULL,
  upstream_url VARCHAR(500) NOT NULL,
  headers_json JSONB DEFAULT '{}',
  auth_mode VARCHAR(50) DEFAULT 'none' CHECK (auth_mode IN ('none', 'api_key', 'bearer')),
  caching_policy_json JSONB DEFAULT '{"enabled": false, "ttl_seconds": 300}',
  rate_limit_policy_json JSONB DEFAULT '{"enabled": false, "requests_per_minute": 100}',
  enabled BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(tenant_id, match_path)
);

-- API keys management
CREATE TABLE IF NOT EXISTS api_keys (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  prefix VARCHAR(20) NOT NULL,
  hash VARCHAR(255) NOT NULL,
  last_used_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(tenant_id, name)
);

-- Custom domains for tenant branding
CREATE TABLE IF NOT EXISTS custom_domains (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  hostname VARCHAR(255) NOT NULL UNIQUE,
  verification_token VARCHAR(255) NOT NULL,
  status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'failed')),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Request logs for analytics and monitoring
CREATE TABLE IF NOT EXISTS requests_log (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  route_id INTEGER REFERENCES routes(id) ON DELETE SET NULL,
  status_code INTEGER NOT NULL,
  latency_ms INTEGER NOT NULL,
  cache_status VARCHAR(20) DEFAULT 'miss' CHECK (cache_status IN ('hit', 'miss', 'bypass')),
  bytes_in INTEGER DEFAULT 0,
  bytes_out INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(prefix);
CREATE INDEX IF NOT EXISTS idx_routes_tenant_id ON routes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_routes_match_path ON routes(match_path);
CREATE INDEX IF NOT EXISTS idx_custom_domains_tenant_id ON custom_domains(tenant_id);
CREATE INDEX IF NOT EXISTS idx_custom_domains_hostname ON custom_domains(hostname);
CREATE INDEX IF NOT EXISTS idx_requests_log_tenant_id ON requests_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_requests_log_route_id ON requests_log(route_id);
CREATE INDEX IF NOT EXISTS idx_requests_log_created_at ON requests_log(created_at);

-- Insert default tenant
INSERT INTO tenants (name, plan, status) 
VALUES ('Demo Company', 'free', 'active')
ON CONFLICT DO NOTHING;

-- Insert default admin user (password: admin123)
-- First create the user with a reference to tenant ID 1
INSERT INTO users (tenant_id, email, hashed_password, role) 
VALUES (1, 'admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin')
ON CONFLICT (email) DO NOTHING;

-- Insert sample route configurations
INSERT INTO routes (tenant_id, name, match_path, upstream_url, auth_mode, enabled)
VALUES 
  (1, 'Posts API', '/api/v1/posts', 'https://jsonplaceholder.typicode.com', 'none', true),
  (1, 'Secure API', '/api/secure/', 'https://httpbin.org', 'api_key', true)
ON CONFLICT (tenant_id, match_path) DO NOTHING;

-- Insert sample API keys
INSERT INTO api_keys (tenant_id, name, prefix, hash)
VALUES 
  (1, 'demo-client', 'demo', '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8'),
  (1, 'admin-client', 'admin', 'ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f')
ON CONFLICT (tenant_id, name) DO NOTHING;