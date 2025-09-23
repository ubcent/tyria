package proxy

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"github.com/ubcent/edge.link/internal/cache"
	"github.com/ubcent/edge.link/internal/models"
)

func TestDBProxyCaching(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create required tables
	createTables(t, db)

	// Insert test data
	setupCacheTestData(t, db)

	// Create mock upstream server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"message": "Hello from upstream",
			"path":    r.URL.Path,
			"time":    time.Now().Unix(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer upstream.Close()

	// Update route upstream URL to point to mock server
	_, err = db.Exec("UPDATE routes SET upstream_url = ? WHERE id = 1", upstream.URL+"/api")
	if err != nil {
		t.Fatalf("Failed to update route upstream URL: %v", err)
	}

	// Create DBService with test cache
	testCache := cache.NewLRU(1024*1024, 5*time.Minute, 10*time.Minute)
	defer testCache.Stop()
	
	service := NewDBServiceWithCache(db, testCache)

	t.Run("cache miss on first request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant", "1")
		w := httptest.NewRecorder()

		service.dbProxyHandler(w, req)

		t.Logf("Response status: %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
		t.Logf("Response headers: %v", w.Header())

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
			return
		}

		cacheStatus := w.Header().Get("X-Cache-Status")
		if cacheStatus != string(cache.CacheStatusMiss) {
			t.Errorf("Expected cache status %s, got %s", cache.CacheStatusMiss, cacheStatus)
		}

		// Verify response is from upstream
		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		if err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if response["message"] != "Hello from upstream" {
			t.Errorf("Unexpected response message: %v", response["message"])
		}
	})

	t.Run("cache hit on second request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant", "1")
		w := httptest.NewRecorder()

		service.dbProxyHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		cacheStatus := w.Header().Get("X-Cache-Status")
		if cacheStatus != string(cache.CacheStatusHit) {
			t.Errorf("Expected cache status %s, got %s", cache.CacheStatusHit, cacheStatus)
		}
	})

	t.Run("cache bypass for POST request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"data": "test"}`))
		req.Header.Set("X-Tenant", "1")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		service.dbProxyHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		cacheStatus := w.Header().Get("X-Cache-Status")
		if cacheStatus != string(cache.CacheStatusBypass) {
			t.Errorf("Expected cache status %s, got %s", cache.CacheStatusBypass, cacheStatus)
		}
	})

	t.Run("vary headers affect cache key", func(t *testing.T) {
		// Request with different Accept-Language header
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant", "1")
		req.Header.Set("Accept-Language", "fr-FR")
		w := httptest.NewRecorder()

		service.dbProxyHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		cacheStatus := w.Header().Get("X-Cache-Status")
		if cacheStatus != string(cache.CacheStatusMiss) {
			t.Errorf("Expected cache status %s for vary header, got %s", cache.CacheStatusMiss, cacheStatus)
		}
	})
}

func TestCacheManagement(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

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