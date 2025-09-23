// Package cache provides caching functionality for the edge.link proxy service.
// It includes in-memory LRU cache and Redis cache implementations.
package cache

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Interface defines the cache operations
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

// CacheStatus represents the cache lookup result
type CacheStatus string

const (
	CacheStatusHit    CacheStatus = "hit"
	CacheStatusMiss   CacheStatus = "miss"
	CacheStatusBypass CacheStatus = "bypass"
)

// KeyBuilder provides cache key generation
type KeyBuilder struct{}

// NewKeyBuilder creates a new key builder
func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{}
}

// GenerateKey generates a cache key from request components including tenant
func (kb *KeyBuilder) GenerateKey(tenantID int, route string, method, path, query string, varyHeaders map[string]string) string {
	baseKey := GenerateKey(method, path, query)
	key := fmt.Sprintf("tenant:%d:route:%s:%s", tenantID, route, baseKey)

	// Add vary headers to key if present
	if len(varyHeaders) > 0 {
		var headerParts []string
		for k, v := range varyHeaders {
			headerParts = append(headerParts, fmt.Sprintf("%s=%s", k, v))
		}
		// Sort for consistent key generation
		sort.Strings(headerParts)
		key += ":vary:" + strings.Join(headerParts, "&")
	}

	return key
}

// GenerateKeyWithBody generates a cache key including request body for POST/PUT requests
func (kb *KeyBuilder) GenerateKeyWithBody(tenantID int, route string, method, path, query string, body []byte, varyHeaders map[string]string) string {
	baseKey := GenerateKeyWithBody(method, path, query, body)
	key := fmt.Sprintf("tenant:%d:route:%s:%s", tenantID, route, baseKey)

	// Add vary headers to key if present
	if len(varyHeaders) > 0 {
		var headerParts []string
		for k, v := range varyHeaders {
			headerParts = append(headerParts, fmt.Sprintf("%s=%s", k, v))
		}
		// Sort for consistent key generation
		sort.Strings(headerParts)
		key += ":vary:" + strings.Join(headerParts, "&")
	}

	return key
}

// IsCacheable determines if a request method is cacheable
func IsCacheable(method string) bool {
	return method == "GET" || method == "HEAD"
}
