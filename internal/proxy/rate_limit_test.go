package proxy

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ubcent/edge.link/internal/models"
	"github.com/ubcent/edge.link/internal/ratelimit"
)

func TestDBService_RateLimit(t *testing.T) {
	// Create a mock database (in a real test, you'd use a test database)
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close test database: %v", err)
		}
	}(db)

	// Create rate limiting service with in-memory storage for testing
	rateLimitConfig := ratelimit.ServiceConfig{
		UseRedis: false,
		InMemoryConfig: ratelimit.Config{
			MaxTokens:     5, // Small burst for testing
			RefillRate:    2, // 2 tokens per period
			RefillPeriod:  time.Second,
			CleanupPeriod: time.Minute,
		},
	}

	// Create test service
	service := NewDBServiceWithRateLimit(db, nil, rateLimitConfig)

	// Create a test route with rate limiting enabled
	route := &models.Route{
		ID:                  1,
		TenantID:            1,
		Name:                "test-route",
		MatchPath:           "/api/test",
		UpstreamURL:         "http://example.com",
		RateLimitPolicyJSON: []byte(`{"enabled": true, "requests_per_minute": 60, "burst": 3}`),
		Enabled:             true,
	}

	// Test rate limiting enforcement
	t.Run("rate limit enforcement", func(t *testing.T) {
		// Reset rate limiter
		service.limiter.Reset()

		// First few requests should be allowed (within burst limit)
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/api/test", nil)
			req.Header.Set("X-Tenant-ID", "1")

			err := service.enforceRateLimit(route, req, 1)
			if err != nil {
				t.Errorf("Request %d should be allowed, got error: %v", i+1, err)
			}
		}

		// Next request should be rate limited
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("X-Tenant-ID", "1")

		err := service.enforceRateLimit(route, req, 1)
		if err == nil {
			t.Error("Request should be rate limited")
		}

		if rateLimitErr, ok := err.(*RateLimitError); ok {
			if rateLimitErr.RetryAfter <= 0 {
				t.Error("Expected positive retry-after value")
			}
		} else {
			t.Error("Expected RateLimitError")
		}
	})

	t.Run("disabled rate limiting", func(t *testing.T) {
		// Create route with disabled rate limiting
		disabledRoute := &models.Route{
			ID:                  2,
			TenantID:            1,
			Name:                "disabled-route",
			MatchPath:           "/api/disabled",
			UpstreamURL:         "http://example.com",
			RateLimitPolicyJSON: []byte(`{"enabled": false, "requests_per_minute": 1, "burst": 1}`),
			Enabled:             true,
		}

		// Many requests should be allowed
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("GET", "/api/disabled", nil)
			req.Header.Set("X-Tenant-ID", "1")

			err := service.enforceRateLimit(disabledRoute, req, 1)
			if err != nil {
				t.Errorf("Request %d should be allowed when rate limiting is disabled, got error: %v", i+1, err)
			}
		}
	})

	t.Run("no rate limit policy", func(t *testing.T) {
		// Create route with no rate limit policy
		noLimitRoute := &models.Route{
			ID:          3,
			TenantID:    1,
			Name:        "no-limit-route",
			MatchPath:   "/api/nolimit",
			UpstreamURL: "http://example.com",
			Enabled:     true,
		}

		// Many requests should be allowed
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("GET", "/api/nolimit", nil)
			req.Header.Set("X-Tenant-ID", "1")

			err := service.enforceRateLimit(noLimitRoute, req, 1)
			if err != nil {
				t.Errorf("Request %d should be allowed when no rate limit policy is set, got error: %v", i+1, err)
			}
		}
	})
}

func TestDBService_ExtractAPIKey(t *testing.T) {
	service := NewDBService(nil) // No DB needed for this test

	tests := []struct {
		name         string
		setupRequest func() *http.Request
		expected     string
	}{
		{
			name: "Bearer token in Authorization header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "Bearer abc123xyz")
				return req
			},
			expected: "abc123xyz",
		},
		{
			name: "API key in X-API-Key header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("X-API-Key", "def456uvw")
				return req
			},
			expected: "def456uvw",
		},
		{
			name: "API key in query parameter",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/?api_key=ghi789rst", nil)
				return req
			},
			expected: "ghi789rst",
		},
		{
			name: "No API key",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				return req
			},
			expected: "",
		},
		{
			name: "Invalid Authorization header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
				return req
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			result := service.extractAPIKey(req)
			if result != tt.expected {
				t.Errorf("extractAPIKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRateLimitError(t *testing.T) {
	err := &RateLimitError{
		Message:    "Rate limit exceeded",
		RetryAfter: 30,
	}

	if err.Error() != "Rate limit exceeded" {
		t.Errorf("Error() = %v, want %v", err.Error(), "Rate limit exceeded")
	}

	if err.RetryAfter != 30 {
		t.Errorf("RetryAfter = %v, want %v", err.RetryAfter, 30)
	}
}
