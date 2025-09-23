package ratelimit

import (
	"testing"
	"time"
)

func TestRedisLimiter_AllowWithPolicy(t *testing.T) {
	// Skip if Redis is not available
	redisConfig := RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	
	limiter := NewRedisLimiter(redisConfig)
	defer limiter.Close()
	
	// Test basic allowance
	key := "test:redis:basic"
	limiter.Reset() // Clear any existing state
	
	// Allow first request
	allowed, retryAfter := limiter.AllowWithPolicy(key, 60, 5) // 60 req/min, burst 5
	if !allowed {
		t.Errorf("Expected first request to be allowed")
	}
	if retryAfter != 0 {
		t.Errorf("Expected retry-after to be 0 for allowed request, got %d", retryAfter)
	}
	
	// Allow burst requests
	for i := 0; i < 4; i++ {
		allowed, _ := limiter.AllowWithPolicy(key, 60, 5)
		if !allowed {
			t.Errorf("Expected burst request %d to be allowed", i+2)
		}
	}
	
	// Sixth request should be denied (burst limit exceeded)
	allowed, retryAfter = limiter.AllowWithPolicy(key, 60, 5)
	if allowed {
		t.Errorf("Expected sixth request to be denied")
	}
	if retryAfter <= 0 {
		t.Errorf("Expected retry-after to be positive for denied request, got %d", retryAfter)
	}
}

func TestRedisLimiter_DifferentKeys(t *testing.T) {
	redisConfig := RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	
	limiter := NewRedisLimiter(redisConfig)
	defer limiter.Close()
	
	limiter.Reset()
	
	key1 := "test:redis:key1"
	key2 := "test:redis:key2"
	
	// Both keys should be allowed independently
	allowed1, _ := limiter.AllowWithPolicy(key1, 60, 1) // Burst of 1
	allowed2, _ := limiter.AllowWithPolicy(key2, 60, 1) // Burst of 1
	
	if !allowed1 {
		t.Errorf("Expected key1 to be allowed")
	}
	if !allowed2 {
		t.Errorf("Expected key2 to be allowed")
	}
	
	// Second requests should both be denied
	allowed1, retry1 := limiter.AllowWithPolicy(key1, 60, 1)
	allowed2, retry2 := limiter.AllowWithPolicy(key2, 60, 1)
	
	if allowed1 {
		t.Errorf("Expected second request for key1 to be denied")
	}
	if allowed2 {
		t.Errorf("Expected second request for key2 to be denied")
	}
	if retry1 <= 0 || retry2 <= 0 {
		t.Errorf("Expected positive retry-after values, got %d and %d", retry1, retry2)
	}
}

func TestGenerateTenantRouteKey(t *testing.T) {
	key := GenerateTenantRouteKey(123, "/api/users")
	expected := "tenant:123:route:/api/users"
	if key != expected {
		t.Errorf("Expected %s, got %s", expected, key)
	}
}

func TestGenerateAPIKeyKey(t *testing.T) {
	key := GenerateAPIKeyKey("abc123")
	expected := "apikey:abc123"
	if key != expected {
		t.Errorf("Expected %s, got %s", expected, key)
	}
}

func TestRedisLimiter_Stats(t *testing.T) {
	redisConfig := RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	
	limiter := NewRedisLimiter(redisConfig)
	defer limiter.Close()
	
	limiter.Reset()
	
	// Create some activity
	limiter.AllowWithPolicy("test:stats:1", 60, 5)
	limiter.AllowWithPolicy("test:stats:2", 60, 5)
	
	stats := limiter.Stats()
	if stats.TotalBuckets < 0 {
		t.Errorf("Expected non-negative bucket count, got %d", stats.TotalBuckets)
	}
}

func TestService_Redis(t *testing.T) {
	config := ServiceConfig{
		UseRedis: true,
		RedisConfig: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}
	
	service := NewService(config)
	defer service.Close()
	
	service.Reset()
	
	// Test basic functionality
	key := "test:service:redis"
	allowed, retryAfter := service.AllowWithPolicy(key, 60, 2)
	if !allowed {
		t.Errorf("Expected first request to be allowed")
	}
	if retryAfter != 0 {
		t.Errorf("Expected retry-after to be 0, got %d", retryAfter)
	}
	
	// Test burst limit
	service.AllowWithPolicy(key, 60, 2) // Use second token
	allowed, retryAfter = service.AllowWithPolicy(key, 60, 2) // Should be denied
	if allowed {
		t.Errorf("Expected third request to be denied")
	}
	if retryAfter <= 0 {
		t.Errorf("Expected positive retry-after, got %d", retryAfter)
	}
}

func TestService_InMemory(t *testing.T) {
	config := ServiceConfig{
		UseRedis: false,
		InMemoryConfig: Config{
			MaxTokens:     10,
			RefillRate:    1,
			RefillPeriod:  time.Second,
			CleanupPeriod: time.Minute,
		},
	}
	
	service := NewService(config)
	defer service.Close()
	
	service.Reset()
	
	// Test basic functionality
	key := "test:service:memory"
	allowed, retryAfter := service.AllowWithPolicy(key, 60, 2)
	if !allowed {
		t.Errorf("Expected first request to be allowed")
	}
	if retryAfter != 0 {
		t.Errorf("Expected retry-after to be 0, got %d", retryAfter)
	}
}