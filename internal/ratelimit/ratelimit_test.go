package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	bucket := NewTokenBucket(5, 1, time.Second)

	// Should allow up to 5 requests initially
	for i := 0; i < 5; i++ {
		if !bucket.Allow() {
			t.Fatalf("Expected request %d to be allowed", i+1)
		}
	}

	// 6th request should be denied
	if bucket.Allow() {
		t.Fatal("Expected 6th request to be denied")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	bucket := NewTokenBucket(2, 2, 100*time.Millisecond)

	// Consume all tokens
	bucket.Allow()
	bucket.Allow()

	// Should be denied
	if bucket.Allow() {
		t.Fatal("Expected request to be denied")
	}

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	if !bucket.Allow() {
		t.Fatal("Expected request to be allowed after refill")
	}
}

func TestLimiter_Allow(t *testing.T) {
	config := Config{
		MaxTokens:    2,
		RefillRate:   1,
		RefillPeriod: time.Second,
	}
	limiter := NewLimiter(config)

	key := "test-key"

	// Should allow up to 2 requests
	if !limiter.Allow(key) {
		t.Fatal("Expected first request to be allowed")
	}
	if !limiter.Allow(key) {
		t.Fatal("Expected second request to be allowed")
	}

	// Third should be denied
	if limiter.Allow(key) {
		t.Fatal("Expected third request to be denied")
	}
}

func TestLimiter_MultipleKeys(t *testing.T) {
	config := Config{
		MaxTokens:    1,
		RefillRate:   1,
		RefillPeriod: time.Second,
	}
	limiter := NewLimiter(config)

	key1 := "key1"
	key2 := "key2"

	// Each key should have its own bucket
	if !limiter.Allow(key1) {
		t.Fatal("Expected request for key1 to be allowed")
	}
	if !limiter.Allow(key2) {
		t.Fatal("Expected request for key2 to be allowed")
	}

	// Both should be exhausted now
	if limiter.Allow(key1) {
		t.Fatal("Expected second request for key1 to be denied")
	}
	if limiter.Allow(key2) {
		t.Fatal("Expected second request for key2 to be denied")
	}
}

func TestGenerateKeys(t *testing.T) {
	tests := []struct {
		name     string
		function func(string) string
		input    string
		expected string
	}{
		{"client key", GenerateClientKey, "user123", "client:user123"},
		{"endpoint key", GenerateEndpointKey, "/api/users", "endpoint:/api/users"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.function(test.input)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}

	// Test client-endpoint key
	result := GenerateClientEndpointKey("user123", "/api/users")
	expected := "client:user123:endpoint:/api/users"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}