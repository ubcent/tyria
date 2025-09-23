package cache

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache provides a Redis-backed cache implementation
type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedis creates a new Redis cache instance
func NewRedis(config RedisConfig, defaultTTL time.Duration) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return &RedisCache{
		client:     client,
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from the Redis cache
func (r *RedisCache) Get(key string) ([]byte, bool) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		// Log error but return cache miss
		return nil, false
	}

	return []byte(val), true
}

// Set stores a value in the Redis cache with the default TTL
func (r *RedisCache) Set(key string, value []byte) bool {
	return r.SetWithTTL(key, value, r.defaultTTL)
}

// SetWithTTL stores a value in the Redis cache with a specific TTL
func (r *RedisCache) SetWithTTL(key string, value []byte, ttl time.Duration) bool {
	ctx := context.Background()
	err := r.client.Set(ctx, key, value, ttl).Err()
	return err == nil
}

// Delete removes a value from the Redis cache
func (r *RedisCache) Delete(key string) {
	ctx := context.Background()
	r.client.Del(ctx, key)
}

// Clear removes all entries from the Redis cache
func (r *RedisCache) Clear() {
	ctx := context.Background()
	r.client.FlushDB(ctx)
}

// Size returns the current size of the Redis cache in bytes (approximate)
func (r *RedisCache) Size() int64 {
	ctx := context.Background()
	info, err := r.client.Info(ctx, "memory").Result()
	if err != nil {
		return 0
	}

	// Parse used_memory from Redis INFO output
	// This is approximate as it includes all Redis memory usage
	lines := parseInfo(info)
	if usedMemory, exists := lines["used_memory"]; exists {
		if size, err := strconv.ParseInt(usedMemory, 10, 64); err == nil {
			return size
		}
	}
	return 0
}

// Len returns the number of entries in the Redis cache
func (r *RedisCache) Len() int {
	ctx := context.Background()
	count, err := r.client.DBSize(ctx).Result()
	if err != nil {
		return 0
	}
	return int(count)
}

// Stats returns cache statistics
func (r *RedisCache) Stats() Stats {
	ctx := context.Background()
	info, err := r.client.Info(ctx, "memory").Result()
	if err != nil {
		return Stats{}
	}

	lines := parseInfo(info)
	size := int64(0)
	if usedMemory, exists := lines["used_memory"]; exists {
		if parsedSize, err := strconv.ParseInt(usedMemory, 10, 64); err == nil {
			size = parsedSize
		}
	}

	return Stats{
		Entries:        r.Len(),
		Size:           size,
		MaxSize:        -1, // Redis doesn't have a fixed max size in this implementation
		ExpiredEntries: 0,  // Redis handles expiration automatically
	}
}

// Stop closes the Redis connection
func (r *RedisCache) Stop() {
	r.client.Close()
}

// parseInfo parses Redis INFO command output into key-value pairs
func parseInfo(info string) map[string]string {
	lines := make(map[string]string)
	for _, line := range strings.Split(info, "\r\n") {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				lines[parts[0]] = parts[1]
			}
		}
	}
	return lines
}

// Ensure RedisCache implements Interface
var _ Interface = (*RedisCache)(nil)
