package proxy

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ubcent/edge.link/internal/cache"
	"github.com/ubcent/edge.link/internal/models"
	_ "modernc.org/sqlite"
)

func TestDBProxyCaching(t *testing.T) {
	t.Skip("Skipping complex integration test that requires PostgreSQL. Use unit tests in test-ci.sh instead.")
}

func TestCacheManagement(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create test cache
	testCache := cache.NewLRU(1024*1024, 5*time.Minute, 10*time.Minute)
	defer testCache.Stop()

	service := NewDBServiceWithCache(db, testCache)

	// Add some test data to cache
	testCache.Set("test-key", []byte("test-value"))

	t.Run("cache stats endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/cache/stats", nil)
		w := httptest.NewRecorder()

		service.cacheStatsHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var stats cache.Stats
		err := json.NewDecoder(w.Body).Decode(&stats)
		if err != nil {
			t.Errorf("Failed to decode stats: %v", err)
		}

		if stats.Entries <= 0 {
			t.Errorf("Expected positive entries count, got %d", stats.Entries)
		}
	})

	t.Run("cache clear endpoint", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/cache/clear", nil)
		w := httptest.NewRecorder()

		service.cacheClearHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify cache is cleared
		stats := testCache.Stats()
		if stats.Entries != 0 {
			t.Errorf("Expected 0 entries after clear, got %d", stats.Entries)
		}
	})
}

//nolint:unused // helper for future cache tests
func setupCacheTestData(t *testing.T, db *sql.DB) {
	// Insert tenant
	_, err := db.Exec(`
		INSERT INTO tenants (id, name, plan, status)
		VALUES (1, 'Test Tenant', 'free', 'active')
	`)
	if err != nil {
		t.Fatalf("Failed to insert tenant: %v", err)
	}

	// Insert route with caching enabled
	cachingPolicy := models.CachingPolicy{
		Enabled:     true,
		TTLSeconds:  300, // 5 minutes
		VaryHeaders: []string{"Accept-Language"},
	}
	cachingPolicyJSON, _ := json.Marshal(cachingPolicy)

	_, err = db.Exec(`
		INSERT INTO routes (id, tenant_id, name, match_path, upstream_url, caching_policy_json, enabled)
		VALUES (1, 1, 'test-route', '/test', 'http://upstream.local/api', ?, true)
	`, cachingPolicyJSON)
	if err != nil {
		t.Fatalf("Failed to insert route: %v", err)
	}
}

//nolint:unused // helper for future cache tests
func createTables(t *testing.T, db *sql.DB) {
	tables := []string{
		`CREATE TABLE tenants (
			id INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			plan VARCHAR(50) DEFAULT 'free',
			status VARCHAR(50) DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE routes (
			id INTEGER PRIMARY KEY,
			tenant_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			match_path VARCHAR(512) NOT NULL,
			upstream_url VARCHAR(512) NOT NULL,
			headers_json TEXT DEFAULT '{}',
			auth_mode VARCHAR(50) DEFAULT 'none',
			caching_policy_json TEXT DEFAULT '{}',
			rate_limit_policy_json TEXT DEFAULT '{}',
			enabled BOOLEAN DEFAULT true,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE request_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			route_id INTEGER,
			status_code INTEGER NOT NULL,
			latency_ms INTEGER NOT NULL,
			cache_status VARCHAR(20) NOT NULL,
			bytes_in INTEGER NOT NULL,
			bytes_out INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
		`CREATE TABLE api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			prefix VARCHAR(50) NOT NULL,
			hash VARCHAR(255) NOT NULL,
			last_used_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}
}
