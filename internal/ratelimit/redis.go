package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements a Redis-based token bucket rate limiter
type RedisLimiter struct {
	client *redis.Client
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisLimiter creates a new Redis-based rate limiter
func NewRedisLimiter(config RedisConfig) *RedisLimiter {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return &RedisLimiter{
		client: client,
	}
}

// AllowWithPolicy checks if a request is allowed with the given policy and returns retry-after seconds
func (r *RedisLimiter) AllowWithPolicy(key string, requestsPerMinute, burst int) (allowed bool, retryAfter int) {
	if requestsPerMinute <= 0 {
		return true, 0
	}

	ctx := context.Background()
	now := time.Now().Unix()
	windowStart := now - (now % 60) // Start of current minute

	// Redis Lua script for atomic token bucket operations
	script := `
		local key = ARGV[1]
		local window_start = tonumber(ARGV[2])
		local requests_per_minute = tonumber(ARGV[3])
		local burst = tonumber(ARGV[4])
		local now = tonumber(ARGV[5])
		
		-- Get current bucket state
		local bucket_key = key .. ":bucket:" .. window_start
		local current_tokens = redis.call('GET', bucket_key)
		if not current_tokens then
			current_tokens = burst
		else
			current_tokens = tonumber(current_tokens)
		end
		
		-- Calculate tokens to add based on time elapsed
		local last_refill_key = key .. ":last_refill"
		local last_refill = redis.call('GET', last_refill_key)
		if not last_refill then
			last_refill = now
		else
			last_refill = tonumber(last_refill)
		end
		
		-- Add tokens based on time elapsed (requests_per_minute tokens per 60 seconds)
		local elapsed = now - last_refill
		if elapsed > 0 then
			local tokens_to_add = math.floor((elapsed * requests_per_minute) / 60)
			current_tokens = math.min(current_tokens + tokens_to_add, burst)
		end
		
		-- Check if request is allowed
		if current_tokens > 0 then
			-- Allow request and consume token
			current_tokens = current_tokens - 1
			redis.call('SET', bucket_key, current_tokens, 'EX', 120) -- Expire after 2 minutes
			redis.call('SET', last_refill_key, now, 'EX', 120)
			return {1, 0} -- allowed, no retry-after
		else
			-- Request denied, calculate retry-after
			local tokens_needed = 1
			local time_for_tokens = math.ceil((tokens_needed * 60) / requests_per_minute)
			return {0, time_for_tokens} -- not allowed, retry-after seconds
		end
	`

	result, err := r.client.Eval(ctx, script, []string{}, key, windowStart, requestsPerMinute, burst, now).Result()
	if err != nil {
		// On Redis error, allow the request (fail open)
		return true, 0
	}

	if resultSlice, ok := result.([]interface{}); ok && len(resultSlice) == 2 {
		allowed := resultSlice[0].(int64) == 1
		retryAfter := int(resultSlice[1].(int64))
		return allowed, retryAfter
	}

	// Fallback if script result is unexpected
	return true, 0
}

// Allow checks if a single request is allowed for the given key with default limits
func (r *RedisLimiter) Allow(key string) bool {
	allowed, _ := r.AllowWithPolicy(key, 60, 10) // Default: 60 req/min, burst 10
	return allowed
}

// AllowN checks if n requests are allowed for the given key with default limits
func (r *RedisLimiter) AllowN(key string, n int) bool {
	// For simplicity, just check if we can allow n consecutive requests
	for i := 0; i < n; i++ {
		if !r.Allow(key) {
			return false
		}
	}
	return true
}

// Stats returns basic statistics (simplified for Redis implementation)
func (r *RedisLimiter) Stats() Stats {
	ctx := context.Background()

	// Count keys with our pattern
	keys, err := r.client.Keys(ctx, "*:bucket:*").Result()
	if err != nil {
		return Stats{TotalBuckets: 0, Buckets: make(map[string]BucketStats)}
	}

	buckets := make(map[string]BucketStats)
	for _, key := range keys {
		tokens, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		tokenCount, _ := strconv.Atoi(tokens)
		buckets[key] = BucketStats{
			Tokens:    tokenCount,
			MaxTokens: -1, // Don't track max tokens in this simple implementation
		}
	}

	return Stats{
		TotalBuckets: len(buckets),
		Buckets:      buckets,
	}
}

// Reset removes all rate limit buckets
func (r *RedisLimiter) Reset() {
	ctx := context.Background()

	// Delete all bucket keys
	keys, err := r.client.Keys(ctx, "*:bucket:*").Result()
	if err != nil {
		return
	}

	if len(keys) > 0 {
		r.client.Del(ctx, keys...)
	}

	// Delete all last_refill keys
	keys, err = r.client.Keys(ctx, "*:last_refill").Result()
	if err != nil {
		return
	}

	if len(keys) > 0 {
		r.client.Del(ctx, keys...)
	}
}

// Close closes the Redis connection
func (r *RedisLimiter) Close() error {
	return r.client.Close()
}

// GenerateTenantRouteKey generates a rate limit key for tenant+route combination
func GenerateTenantRouteKey(tenantID int, route string) string {
	return fmt.Sprintf("tenant:%d:route:%s", tenantID, route)
}

// GenerateAPIKeyKey generates a rate limit key for API key rate limiting
func GenerateAPIKeyKey(apiKeyPrefix string) string {
	return fmt.Sprintf("apikey:%s", apiKeyPrefix)
}
