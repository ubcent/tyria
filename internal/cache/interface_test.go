package cache

import (
	"strings"
	"testing"
	"time"
)

func TestKeyBuilder_GenerateKey(t *testing.T) {
	kb := NewKeyBuilder()

	tests := []struct {
		name        string
		tenantID    int
		route       string
		method      string
		path        string
		query       string
		varyHeaders map[string]string
		expected    string
	}{
		{
			name:     "simple GET request",
			tenantID: 1,
			route:    "api-route",
			method:   "GET",
			path:     "/users",
			query:    "",
			expected: "tenant:1:route:api-route:GET:/users",
		},
		{
			name:     "GET request with query",
			tenantID: 2,
			route:    "search-route",
			method:   "GET",
			path:     "/search",
			query:    "q=test",
			expected: "tenant:2:route:search-route:GET:/search?q=test",
		},
		{
			name:     "GET request with vary headers",
			tenantID: 1,
			route:    "api-route",
			method:   "GET",
			path:     "/users",
			query:    "",
			varyHeaders: map[string]string{
				"Accept-Language": "en-US",
				"User-Agent":      "TestAgent",
			},
			expected: "tenant:1:route:api-route:GET:/users:vary:Accept-Language=en-US&User-Agent=TestAgent",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := kb.GenerateKey(test.tenantID, test.route, test.method, test.path, test.query, test.varyHeaders)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestKeyBuilder_GenerateKeyWithBody(t *testing.T) {
	kb := NewKeyBuilder()

	tests := []struct {
		name        string
		tenantID    int
		route       string
		method      string
		path        string
		query       string
		body        []byte
		varyHeaders map[string]string
		expectsBody bool
	}{
		{
			name:        "GET request ignores body",
			tenantID:    1,
			route:       "api-route",
			method:      "GET",
			path:        "/users",
			query:       "",
			body:        []byte(`{"name": "test"}`),
			expectsBody: false,
		},
		{
			name:        "POST request includes body hash",
			tenantID:    1,
			route:       "api-route",
			method:      "POST",
			path:        "/users",
			query:       "",
			body:        []byte(`{"name": "test"}`),
			expectsBody: true,
		},
		{
			name:        "POST request with vary headers",
			tenantID:    1,
			route:       "api-route",
			method:      "POST",
			path:        "/users",
			query:       "",
			body:        []byte(`{"name": "test"}`),
			varyHeaders: map[string]string{"Content-Type": "application/json"},
			expectsBody: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := kb.GenerateKeyWithBody(test.tenantID, test.route, test.method, test.path, test.query, test.body, test.varyHeaders)
			
			// Check if the result contains tenant, route, and method/path
			expectedBase := "tenant:1:route:api-route:"
			if !strings.HasPrefix(result, expectedBase) {
				t.Errorf("Expected key to start with %s, got %s", expectedBase, result)
			}

			// Check if body hash is included for POST requests
			if test.expectsBody && test.method == "POST" && len(test.body) > 0 {
				if !strings.Contains(result, ":body:") {
					t.Errorf("Expected POST request with body to include body hash, got %s", result)
				}
			}

			// Check if vary headers are included
			if len(test.varyHeaders) > 0 {
				if !strings.Contains(result, ":vary:") {
					t.Errorf("Expected vary headers to be included, got %s", result)
				}
			}
		})
	}
}

func TestIsCacheable(t *testing.T) {
	tests := []struct {
		method   string
		expected bool
	}{
		{"GET", true},
		{"HEAD", true},
		{"POST", false},
		{"PUT", false},
		{"DELETE", false},
		{"PATCH", false},
		{"OPTIONS", false},
	}

	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			result := IsCacheable(test.method)
			if result != test.expected {
				t.Errorf("Expected %s to be cacheable: %v, got %v", test.method, test.expected, result)
			}
		})
	}
}

func TestCacheInterface_LRUCache(t *testing.T) {
	// Test that LRUCache implements the Interface
	var cache Interface = NewLRU(1024, 5*time.Minute, 10*time.Minute)
	defer cache.Stop()

	// Test basic operations
	key := "test-key"
	value := []byte("test-value")

	// Test Set
	ok := cache.Set(key, value)
	if !ok {
		t.Fatal("Expected Set to return true")
	}

	// Test Get
	retrieved, found := cache.Get(key)
	if !found {
		t.Fatal("Expected to find the key")
	}

	if string(retrieved) != string(value) {
		t.Fatalf("Expected %s, got %s", string(value), string(retrieved))
	}

	// Test Delete
	cache.Delete(key)
	_, found = cache.Get(key)
	if found {
		t.Fatal("Expected key to be deleted")
	}

	// Test Stats
	stats := cache.Stats()
	if stats.Entries < 0 {
		t.Fatal("Expected non-negative entries count")
	}
}