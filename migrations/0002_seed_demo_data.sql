-- +goose Up
-- +goose StatementBegin

-- Insert demo tenant
INSERT INTO tenants (name, plan, status) 
VALUES ('Demo Company', 'pro', 'active')
ON CONFLICT DO NOTHING;

-- Insert demo user (password: demo123 - bcrypt hashed)
INSERT INTO users (tenant_id, email, hashed_password, role)
VALUES (
  (SELECT id FROM tenants WHERE name = 'Demo Company'),
  'demo@example.com', 
  '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
  'admin'
) ON CONFLICT (email) DO NOTHING;

-- Insert demo API keys
INSERT INTO api_keys (tenant_id, name, prefix, hash, last_used_at)
VALUES 
  (
    (SELECT id FROM tenants WHERE name = 'Demo Company'),
    'Production API',
    'el_prod_',
    '$2a$10$abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
    NOW() - INTERVAL '2 hours'
  ),
  (
    (SELECT id FROM tenants WHERE name = 'Demo Company'),
    'Development API', 
    'el_dev_',
    '$2a$10$1234567890abcdef1234567890abcdef1234567890abcdef1234567890',
    NOW() - INTERVAL '1 day'
  )
ON CONFLICT (tenant_id, name) DO NOTHING;

-- Insert demo routes
INSERT INTO routes (tenant_id, name, match_path, upstream_url, headers_json, auth_mode, caching_policy_json, rate_limit_policy_json, enabled)
VALUES 
  (
    (SELECT id FROM tenants WHERE name = 'Demo Company'),
    'JSONPlaceholder Posts',
    '/api/v1/posts',
    'https://jsonplaceholder.typicode.com',
    '{"X-Forwarded-Host": "demo.edge.link"}',
    'none',
    '{"enabled": true, "ttl_seconds": 300}',
    '{"enabled": true, "requests_per_minute": 1000}',
    true
  ),
  (
    (SELECT id FROM tenants WHERE name = 'Demo Company'),
    'Secure API',
    '/api/secure/',
    'https://httpbin.org',
    '{"Authorization": "Bearer ${api_key}"}',
    'api_key',
    '{"enabled": true, "ttl_seconds": 120}',
    '{"enabled": true, "requests_per_minute": 500}',
    true
  ),
  (
    (SELECT id FROM tenants WHERE name = 'Demo Company'),
    'Public GitHub API',
    '/public/github/',
    'https://api.github.com',
    '{"User-Agent": "edge.link-proxy/1.0"}',
    'none',
    '{"enabled": true, "ttl_seconds": 600}',
    '{"enabled": true, "requests_per_minute": 5000}',
    true
  )
ON CONFLICT (tenant_id, match_path) DO NOTHING;

-- Insert demo custom domain
INSERT INTO custom_domains (tenant_id, hostname, verification_token, status)
VALUES (
  (SELECT id FROM tenants WHERE name = 'Demo Company'),
  'api.democompany.com',
  'demo_verify_token_12345abcdef',
  'verified'
) ON CONFLICT (hostname) DO NOTHING;

-- Insert sample request logs for demo data
INSERT INTO requests_log (tenant_id, route_id, status_code, latency_ms, cache_status, bytes_in, bytes_out, created_at)
VALUES 
  (
    (SELECT id FROM tenants WHERE name = 'Demo Company'),
    (SELECT id FROM routes WHERE match_path = '/api/v1/posts' AND tenant_id = (SELECT id FROM tenants WHERE name = 'Demo Company')),
    200, 45, 'hit', 512, 2048,
    NOW() - INTERVAL '1 hour'
  ),
  (
    (SELECT id FROM tenants WHERE name = 'Demo Company'),
    (SELECT id FROM routes WHERE match_path = '/api/v1/posts' AND tenant_id = (SELECT id FROM tenants WHERE name = 'Demo Company')),
    200, 120, 'miss', 512, 1024,
    NOW() - INTERVAL '2 hours'
  ),
  (
    (SELECT id FROM tenants WHERE name = 'Demo Company'),
    (SELECT id FROM routes WHERE match_path = '/api/secure/' AND tenant_id = (SELECT id FROM tenants WHERE name = 'Demo Company')),
    401, 15, 'bypass', 256, 128,
    NOW() - INTERVAL '30 minutes'
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM requests_log WHERE tenant_id = (SELECT id FROM tenants WHERE name = 'Demo Company');
DELETE FROM custom_domains WHERE tenant_id = (SELECT id FROM tenants WHERE name = 'Demo Company');
DELETE FROM routes WHERE tenant_id = (SELECT id FROM tenants WHERE name = 'Demo Company');
DELETE FROM api_keys WHERE tenant_id = (SELECT id FROM tenants WHERE name = 'Demo Company');
DELETE FROM users WHERE tenant_id = (SELECT id FROM tenants WHERE name = 'Demo Company');
DELETE FROM tenants WHERE name = 'Demo Company';

-- +goose StatementEnd