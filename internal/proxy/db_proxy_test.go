package proxy

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"
	"github.com/ubcent/edge.link/internal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create tables
	schemas := []string{
		`CREATE TABLE tenants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL,
			plan VARCHAR(50) DEFAULT 'free',
			status VARCHAR(50) DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE routes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			match_path VARCHAR(512) NOT NULL,
			upstream_url VARCHAR(512) NOT NULL,
			headers_json TEXT DEFAULT '{}',
			auth_mode VARCHAR(50) DEFAULT 'none',
			caching_policy_json TEXT DEFAULT '{"enabled": false, "ttl_seconds": 300}',
			rate_limit_policy_json TEXT DEFAULT '{"enabled": false, "requests_per_minute": 60}',
			enabled BOOLEAN DEFAULT true,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			prefix VARCHAR(20) NOT NULL,
			hash VARCHAR(255) NOT NULL,
			last_used_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE custom_domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			hostname VARCHAR(255) NOT NULL UNIQUE,
			verification_token VARCHAR(255) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE requests_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			route_id INTEGER NULL,
			status_code INTEGER NOT NULL,
			latency_ms INTEGER NOT NULL,
			cache_status VARCHAR(20) DEFAULT 'miss',
			bytes_in INTEGER DEFAULT 0,
			bytes_out INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	return db
}

func TestPathMatching(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewDBService(db)

	tests := []struct {
		name        string
		pattern     string
		requestPath string
		expectMatch bool
		expectParams map[string]string
	}{
		{
			name:        "exact match",
			pattern:     "/api/users",
			requestPath: "/api/users",
			expectMatch: true,
			expectParams: map[string]string{},
		},
		{
			name:        "no match different path",
			pattern:     "/api/users",
			requestPath: "/api/posts",
			expectMatch: false,
			expectParams: map[string]string{},
		},
		{
			name:        "wildcard match",
			pattern:     "/api/*",
			requestPath: "/api/users/123",
			expectMatch: true,
			expectParams: map[string]string{},
		},
		{
			name:        "single parameter",
			pattern:     "/users/{id}",
			requestPath: "/users/123",
			expectMatch: true,
			expectParams: map[string]string{"id": "123"},
		},
		{
			name:        "multiple parameters",
			pattern:     "/users/{userId}/posts/{postId}",
			requestPath: "/users/123/posts/456",
			expectMatch: true,
			expectParams: map[string]string{"userId": "123", "postId": "456"},
		},
		{
			name:        "parameter mismatch count",
			pattern:     "/users/{id}",
			requestPath: "/users/123/extra",
			expectMatch: false,
			expectParams: map[string]string{},
		},
		{
			name:        "mixed literal and parameter",
			pattern:     "/api/v1/users/{id}",
			requestPath: "/api/v1/users/123",
			expectMatch: true,
			expectParams: map[string]string{"id": "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, matches := service.matchPath(tt.pattern, tt.requestPath)
			
			if matches != tt.expectMatch {
				t.Errorf("Expected match %v, got %v for pattern %s and path %s", 
					tt.expectMatch, matches, tt.pattern, tt.requestPath)
			}
			
			if tt.expectMatch {
				for key, expectedValue := range tt.expectParams {
					if actualValue, exists := params[key]; !exists {
						t.Errorf("Expected parameter %s not found", key)
					} else if actualValue != expectedValue {
						t.Errorf("Expected parameter %s=%s, got %s", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

func TestTenantResolution(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewDBService(db)

	// Insert test data
	_, err := db.Exec("INSERT INTO tenants (id, name) VALUES (1, 'test-tenant')")
	if err != nil {
		t.Fatalf("Failed to insert tenant: %v", err)
	}

	_, err = db.Exec("INSERT INTO custom_domains (tenant_id, hostname, verification_token, status) VALUES (1, 'example.com', 'token123', 'verified')")
	if err != nil {
		t.Fatalf("Failed to insert custom domain: %v", err)
	}

	tests := []struct {
		name        string
		host        string
		tenantHeader string
		expectedID  int
		expectError bool
	}{
		{
			name:        "resolve from X-Tenant header",
			host:        "localhost:8080",
			tenantHeader: "1",
			expectedID:  1,
			expectError: false,
		},
		{
			name:        "resolve from custom domain",
			host:        "example.com",
			tenantHeader: "",
			expectedID:  1,
			expectError: false,
		},
		{
			name:        "no tenant found",
			host:        "unknown.com",
			tenantHeader: "",
			expectedID:  0,
			expectError: false,
		},
		{
			name:        "invalid X-Tenant header",
			host:        "localhost:8080",
			tenantHeader: "invalid",
			expectedID:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://"+tt.host+"/test", nil)
			if tt.tenantHeader != "" {
				req.Header.Set("X-Tenant", tt.tenantHeader)
			}

			tenantID, err := service.resolveTenant(req)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if tenantID != tt.expectedID {
				t.Errorf("Expected tenant ID %d, got %d", tt.expectedID, tenantID)
			}
		})
	}
}

func TestTenantFromPath(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewDBService(db)

	tests := []struct {
		name         string
		path         string
		expectedID   int
		expectedPath string
		expectError  bool
	}{
		{
			name:         "valid tenant in path",
			path:         "/1/api/users",
			expectedID:   1,
			expectedPath: "/api/users",
			expectError:  false,
		},
		{
			name:         "tenant only path",
			path:         "/123",
			expectedID:   123,
			expectedPath: "/",
			expectError:  false,
		},
		{
			name:         "empty path",
			path:         "/",
			expectedID:   0,
			expectedPath: "/",
			expectError:  true,
		},
		{
			name:         "invalid tenant ID",
			path:         "/abc/api/users",
			expectedID:   0,
			expectedPath: "/abc/api/users",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://localhost"+tt.path, nil)
			
			tenantID, err := service.extractTenantFromPath(req)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if tenantID != tt.expectedID {
				t.Errorf("Expected tenant ID %d, got %d", tt.expectedID, tenantID)
			}
			
			if !tt.expectError && req.URL.Path != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, req.URL.Path)
			}
		})
	}
}

func TestAuthModeEnforcement(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewDBService(db)

	// Insert test tenant
	_, err := db.Exec("INSERT INTO tenants (id, name) VALUES (1, 'test-tenant')")
	if err != nil {
		t.Fatalf("Failed to insert tenant: %v", err)
	}

	// Create a valid API key using the service
	apiKeyService := service.apiKeysService
	apiKey := &models.APIKey{
		TenantID: 1,
		Name:     "test-key",
	}
	fullKey, err := apiKeyService.Create(context.Background(), apiKey)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	tests := []struct {
		name        string
		authMode    string
		headers     map[string]string
		expectError bool
	}{
		{
			name:        "no auth required",
			authMode:    "none",
			headers:     map[string]string{},
			expectError: false,
		},
		{
			name:        "api key auth with valid header",
			authMode:    "api_key",
			headers:     map[string]string{"X-API-Key": fullKey},
			expectError: false,
		},
		{
			name:        "api key auth missing key",
			authMode:    "api_key",
			headers:     map[string]string{},
			expectError: true,
		},
		{
			name:        "basic auth not implemented",
			authMode:    "basic",
			headers:     map[string]string{"Authorization": "Basic dGVzdDp0ZXN0"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &models.Route{
				AuthMode: tt.authMode,
				TenantID: 1,
			}

			req := httptest.NewRequest("GET", "http://localhost/test", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			err := service.enforceAuth(route, req, 1)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}