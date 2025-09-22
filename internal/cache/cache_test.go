package cache

import (
	"strings"
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := New(1024, 5*time.Minute, 10*time.Minute)
	defer cache.Stop()

	key := "test-key"
	value := []byte("test-value")

	// Test set and get
	ok := cache.Set(key, value)
	if !ok {
		t.Fatal("Expected Set to return true")
	}

	retrieved, found := cache.Get(key)
	if !found {
		t.Fatal("Expected to find the key")
	}

	if string(retrieved) != string(value) {
		t.Fatalf("Expected %s, got %s", string(value), string(retrieved))
	}
}

func TestCache_TTL(t *testing.T) {
	cache := New(1024, 5*time.Minute, 10*time.Minute)
	defer cache.Stop()

	key := "test-key"
	value := []byte("test-value")

	// Set with short TTL
	ok := cache.SetWithTTL(key, value, 50*time.Millisecond)
	if !ok {
		t.Fatal("Expected Set to return true")
	}

	// Should be available immediately
	_, found := cache.Get(key)
	if !found {
		t.Fatal("Expected to find the key immediately")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should be expired
	_, found = cache.Get(key)
	if found {
		t.Fatal("Expected key to be expired")
	}
}

func TestCache_SizeLimit(t *testing.T) {
	cache := New(10, 5*time.Minute, 10*time.Minute) // Very small cache
	defer cache.Stop()

	// Try to add data larger than cache
	key := "test-key"
	value := []byte("this is a very long value that exceeds cache size")

	ok := cache.Set(key, value)
	if ok {
		t.Fatal("Expected Set to return false due to size limit")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := New(1024, 5*time.Minute, 10*time.Minute)
	defer cache.Stop()

	key := "test-key"
	value := []byte("test-value")

	cache.Set(key, value)
	cache.Delete(key)

	_, found := cache.Get(key)
	if found {
		t.Fatal("Expected key to be deleted")
	}
}

func TestCache_Stats(t *testing.T) {
	cache := New(1024, 5*time.Minute, 10*time.Minute)
	defer cache.Stop()

	stats := cache.Stats()
	if stats.Entries != 0 {
		t.Fatalf("Expected 0 entries, got %d", stats.Entries)
	}

	cache.Set("key1", []byte("value1"))
	cache.Set("key2", []byte("value2"))

	stats = cache.Stats()
	if stats.Entries != 2 {
		t.Fatalf("Expected 2 entries, got %d", stats.Entries)
	}
}

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		method, path, query, expected string
	}{
		{"GET", "/api/users", "", "GET:/api/users"},
		{"POST", "/api/users", "sort=name", "POST:/api/users?sort=name"},
		{"GET", "/", "", "GET:/"},
	}

	for _, test := range tests {
		result := GenerateKey(test.method, test.path, test.query)
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestGenerateKeyWithBody(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		query          string
		body           []byte
		expectedPrefix string
		hasBodyHash    bool
	}{
		{
			name:           "GET request ignores body",
			method:         "GET",
			path:           "/api/users",
			query:          "",
			body:           []byte(`{"name": "test"}`),
			expectedPrefix: "GET:/api/users",
			hasBodyHash:    false,
		},
		{
			name:           "POST request with body includes hash",
			method:         "POST", 
			path:           "/api/users",
			query:          "",
			body:           []byte(`{"name": "test"}`),
			expectedPrefix: "POST:/api/users:body:",
			hasBodyHash:    true,
		},
		{
			name:           "PUT request with body includes hash",
			method:         "PUT",
			path:           "/api/users/1",
			query:          "format=json",
			body:           []byte(`{"name": "updated"}`),
			expectedPrefix: "PUT:/api/users/1?format=json:body:",
			hasBodyHash:    true,
		},
		{
			name:           "POST request with empty body",
			method:         "POST",
			path:           "/api/users",
			query:          "",
			body:           []byte{},
			expectedPrefix: "POST:/api/users",
			hasBodyHash:    false,
		},
		{
			name:           "PATCH request with body includes hash",
			method:         "PATCH",
			path:           "/api/users/1", 
			query:          "",
			body:           []byte(`{"status": "active"}`),
			expectedPrefix: "PATCH:/api/users/1:body:",
			hasBodyHash:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := GenerateKeyWithBody(test.method, test.path, test.query, test.body)
			
			if test.hasBodyHash {
				if !strings.HasPrefix(result, test.expectedPrefix) {
					t.Errorf("Expected result to start with %s, got %s", test.expectedPrefix, result)
				}
				if len(result) <= len(test.expectedPrefix) {
					t.Errorf("Expected result to have hash suffix, got %s", result)
				}
			} else {
				if result != test.expectedPrefix {
					t.Errorf("Expected %s, got %s", test.expectedPrefix, result)
				}
			}
		})
	}
	
	// Test that same body produces same hash
	body := []byte(`{"name": "test"}`)
	key1 := GenerateKeyWithBody("POST", "/api/users", "", body)
	key2 := GenerateKeyWithBody("POST", "/api/users", "", body)
	if key1 != key2 {
		t.Errorf("Expected same body to produce same key, got %s != %s", key1, key2)
	}
	
	// Test that different bodies produce different hashes
	body2 := []byte(`{"name": "different"}`)
	key3 := GenerateKeyWithBody("POST", "/api/users", "", body2)
	if key1 == key3 {
		t.Errorf("Expected different bodies to produce different keys, both got %s", key1)
	}
}