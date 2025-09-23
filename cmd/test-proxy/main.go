package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ubcent/edge.link/internal/apikeys"
	"github.com/ubcent/edge.link/internal/models"
	"github.com/ubcent/edge.link/internal/proxy"
	"github.com/ubcent/edge.link/internal/routes"
	"github.com/ubcent/edge.link/internal/tenant"
	_ "modernc.org/sqlite"
)

func main() {
	// Start mock upstream server
	go startMockUpstream()

	// Give the mock server time to start
	time.Sleep(100 * time.Millisecond)

	// Create in-memory SQLite database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

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
			log.Fatalf("Failed to create table: %v", err)
		}
	}

	// Set up test data
	setupTestData(db)

	// Create proxy service
	proxyService := proxy.NewDBService(db)

	// Start test server
	log.Println("Starting test server on :8080")
	log.Println("Test scenarios:")
	log.Println("1. curl -H 'X-Tenant: 1' http://localhost:8080/users/123")
	log.Println("2. curl http://localhost:8080/1/api/posts")
	log.Println("3. curl -H 'X-API-Key: [generated key]' -H 'X-Tenant: 1' http://localhost:8080/admin/users")

	if err := http.ListenAndServe(":8080", proxyService.Handler()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func setupTestData(db *sql.DB) {
	ctx := context.Background()

	// Create services
	tenantService := tenant.NewService(db)
	routesService := routes.NewService(db)
	apiKeysService := apikeys.NewService(db)

	// Create a test tenant
	testTenant := &models.Tenant{
		Name:   "Test Tenant",
		Plan:   "free",
		Status: "active",
	}
	if err := tenantService.Create(ctx, testTenant); err != nil {
		log.Fatalf("Failed to create tenant: %v", err)
	}
	log.Printf("Created tenant: %d", testTenant.ID)

	// Create test routes
	routes := []*models.Route{
		{
			TenantID:    testTenant.ID,
			Name:        "User API",
			MatchPath:   "/users/{id}",
			UpstreamURL: "http://localhost:9999/users/{id}",
			AuthMode:    "none",
			Enabled:     true,
		},
		{
			TenantID:    testTenant.ID,
			Name:        "Posts API",
			MatchPath:   "/api/posts",
			UpstreamURL: "http://localhost:9999/posts",
			AuthMode:    "none",
			Enabled:     true,
		},
		{
			TenantID:    testTenant.ID,
			Name:        "Admin API",
			MatchPath:   "/admin/*",
			UpstreamURL: "http://localhost:9999",
			AuthMode:    "api_key",
			Enabled:     true,
		},
	}

	for _, route := range routes {
		// Set default JSON values
		route.HeadersJSON = json.RawMessage(`{}`)
		route.CachingPolicyJSON = json.RawMessage(`{"enabled": false, "ttl_seconds": 300}`)
		route.RateLimitPolicyJSON = json.RawMessage(`{"enabled": false, "requests_per_minute": 60}`)

		if err := routesService.Create(ctx, route); err != nil {
			log.Fatalf("Failed to create route %s: %v", route.Name, err)
		}
		log.Printf("Created route: %s -> %s", route.MatchPath, route.UpstreamURL)
	}

	// Create an API key for testing protected routes
	apiKey := &models.APIKey{
		TenantID: testTenant.ID,
		Name:     "Test API Key",
	}
	fullKey, err := apiKeysService.Create(ctx, apiKey)
	if err != nil {
		log.Fatalf("Failed to create API key: %v", err)
	}
	log.Printf("Created API key: %s", fullKey)

	// Print the key for manual testing
	fmt.Printf("\nFor testing protected routes, use this API key:\n")
	fmt.Printf("curl -H 'X-API-Key: %s' -H 'X-Tenant: 1' http://localhost:8080/admin/users\n\n", fullKey)
}

func startMockUpstream() {
	mux := http.NewServeMux()

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Path[len("/users/"):]
		response := map[string]interface{}{
			"id":    userID,
			"name":  "Test User " + userID,
			"email": "test" + userID + "@example.com",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		response := []map[string]interface{}{
			{"id": 1, "title": "Test Post 1", "body": "This is a test post"},
			{"id": 2, "title": "Test Post 2", "body": "Another test post"},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"message": "Admin API accessed successfully",
			"path":    r.URL.Path,
			"method":  r.Method,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	log.Println("Starting mock upstream server on :9999")
	if err := http.ListenAndServe(":9999", mux); err != nil {
		log.Printf("Mock server failed: %v", err)
	}
}
