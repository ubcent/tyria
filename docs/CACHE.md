# Cache Implementation

This implementation provides a comprehensive caching system for the edge.link proxy with support for both in-memory LRU and Redis backends.

## Features

### Cache Interface
- Generic `cache.Interface` supporting multiple backends
- LRU in-memory cache implementation
- Redis cache implementation
- Configurable TTL per route
- Vary headers support
- HTTP caching semantics (GET/HEAD only)

### Cache Key Generation
- Enhanced key derivation including:
  - Tenant ID
  - Route name
  - HTTP method and path
  - Query parameters
  - Vary headers (for request differentiation)
  - Request body hash (for POST/PUT, though not cached)

### Cache Status Tracking
- `hit`: Response served from cache
- `miss`: Cache lookup failed, response from upstream
- `bypass`: Caching disabled or non-cacheable request

## Usage

### Basic LRU Cache
```go
import "github.com/ubcent/edge.link/internal/cache"

// Create LRU cache: 100MB max, 5min default TTL, 10min cleanup interval
lruCache := cache.NewLRU(100*1024*1024, 5*time.Minute, 10*time.Minute)
defer lruCache.Stop()

// Use with DBService
service := proxy.NewDBServiceWithCache(db, lruCache)
```

### Redis Cache
```go
import "github.com/ubcent/edge.link/internal/cache"

// Create Redis cache
redisConfig := cache.RedisConfig{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
}
redisCache := cache.NewRedis(redisConfig, 5*time.Minute)
defer redisCache.Stop()

// Use with DBService
service := proxy.NewDBServiceWithCache(db, redisCache)
```

### Route Configuration

Routes support caching through the `caching_policy_json` field:

```json
{
  "enabled": true,
  "ttl_seconds": 300,
  "vary_headers": ["Accept-Language", "User-Agent"]
}
```

- `enabled`: Enable/disable caching for this route
- `ttl_seconds`: Cache time-to-live in seconds
- `vary_headers`: Headers that affect cache key generation

## Cache Key Format

Cache keys follow this pattern:
```
tenant:{tenant_id}:route:{route_name}:{method}:{path}?{query}:vary:{headers}
```

Examples:
- `tenant:1:route:api-users:GET:/users`
- `tenant:1:route:api-users:GET:/users?page=1`
- `tenant:1:route:api-search:GET:/search?q=test:vary:Accept-Language=en-US`

## HTTP Caching Semantics

- Only `GET` and `HEAD` requests are cached
- POST, PUT, DELETE, PATCH requests always bypass cache
- Only successful responses (2xx status codes) are cached
- Cache headers are automatically set:
  - `X-Cache-Status`: `hit`, `miss`, or `bypass`

## Cache Management

The proxy exposes cache management endpoints:

### Get Cache Statistics
```bash
GET /api/cache/stats
```

Response:
```json
{
  "entries": 150,
  "size": 2048576,
  "max_size": 104857600,
  "expired_entries": 5
}
```

### Clear Cache
```bash
POST /api/cache/clear
```

Response:
```json
{
  "message": "Cache cleared successfully",
  "timestamp": "2023-09-23T10:00:00Z"
}
```

## Implementation Details

### Cache Interface
```go
type Interface interface {
    Get(key string) ([]byte, bool)
    Set(key string, value []byte) bool
    SetWithTTL(key string, value []byte, ttl time.Duration) bool
    Delete(key string)
    Clear()
    Size() int64
    Len() int
    Stats() Stats
    Stop()
}
```

### LRU Cache Features
- Thread-safe operations with RWMutex
- Automatic cleanup of expired entries
- Size-based eviction when max size is reached
- Background cleanup goroutine

### Redis Cache Features
- Distributed caching across multiple proxy instances
- Automatic TTL handling
- Connection pooling through go-redis client
- Graceful error handling (cache misses on Redis errors)

## Testing

The implementation includes comprehensive tests:

```bash
# Run cache tests
go test ./internal/cache -v

# Run proxy caching tests
go test ./internal/proxy -v

# Run all tests
go test ./... -v
```

## Performance Considerations

1. **LRU Cache**: Fast in-memory operations, limited by available memory
2. **Redis Cache**: Network overhead but shared across instances
3. **Cache Key Length**: Optimize vary headers to avoid overly long keys
4. **TTL Selection**: Balance between freshness and performance
5. **Size Limits**: Monitor cache size and set appropriate limits

## Migration from Old Cache

The implementation maintains backward compatibility:
- `cache.New()` still works and returns `*LRUCache`
- Existing cache method signatures are preserved
- Can be gradually migrated to use the interface

## Monitoring

Cache performance can be monitored through:
- Cache hit/miss ratios via request logs
- Cache size and entry count via `/api/cache/stats`
- Application metrics and observability tools

## Configuration Examples

### Development (In-Memory)
```yaml
cache:
  type: "lru"
  max_size: 104857600  # 100MB
  default_ttl: "5m"
  cleanup_period: "10m"
```

### Production (Redis)
```yaml
cache:
  type: "redis"
  redis:
    addr: "redis:6379"
    password: ""
    db: 0
  default_ttl: "5m"
```

### Route Caching Policy
```json
{
  "enabled": true,
  "ttl_seconds": 300,
  "vary_headers": ["Accept-Language", "Authorization"]
}
```