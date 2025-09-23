package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ubcent/edge.link/internal/cache"
)

func main() {
	fmt.Println("Edge.link Cache Interface Demo")
	fmt.Println("===============================")

	// Demo 1: LRU Cache
	fmt.Println("\n1. LRU Cache Demo")
	lruCache := cache.NewLRU(1024, 5*time.Minute, 10*time.Minute)
	defer lruCache.Stop()

	// Basic operations
	fmt.Println("Setting key 'user:123' with value 'John Doe'")
	lruCache.Set("user:123", []byte("John Doe"))

	if value, found := lruCache.Get("user:123"); found {
		fmt.Printf("Retrieved: %s\n", string(value))
	}

	// TTL demo
	fmt.Println("Setting key with 2-second TTL")
	lruCache.SetWithTTL("temp:key", []byte("temporary data"), 2*time.Second)
	
	if value, found := lruCache.Get("temp:key"); found {
		fmt.Printf("Retrieved before expiry: %s\n", string(value))
	}

	fmt.Println("Waiting 3 seconds...")
	time.Sleep(3 * time.Second)

	if _, found := lruCache.Get("temp:key"); !found {
		fmt.Println("Key expired and not found")
	}

	// Stats demo
	stats := lruCache.Stats()
	fmt.Printf("Cache stats: %d entries, %d bytes\n", stats.Entries, stats.Size)

	// Demo 2: Key Builder
	fmt.Println("\n2. Cache Key Builder Demo")
	keyBuilder := cache.NewKeyBuilder()

	// Simple key
	key1 := keyBuilder.GenerateKey(1, "api-users", "GET", "/users", "", nil)
	fmt.Printf("Simple key: %s\n", key1)

	// Key with query
	key2 := keyBuilder.GenerateKey(1, "api-users", "GET", "/users", "page=1&limit=10", nil)
	fmt.Printf("Key with query: %s\n", key2)

	// Key with vary headers
	varyHeaders := map[string]string{
		"Accept-Language": "en-US",
		"User-Agent":      "EdgeLink/1.0",
	}
	key3 := keyBuilder.GenerateKey(1, "api-users", "GET", "/users", "", varyHeaders)
	fmt.Printf("Key with vary headers: %s\n", key3)

	// Key with body
	key4 := keyBuilder.GenerateKeyWithBody(1, "api-users", "POST", "/users", "", []byte(`{"name":"John"}`), nil)
	fmt.Printf("Key with body hash: %s\n", key4)

	// Demo 3: HTTP Caching Semantics
	fmt.Println("\n3. HTTP Caching Semantics Demo")
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "PATCH"}
	
	for _, method := range methods {
		cacheable := cache.IsCacheable(method)
		fmt.Printf("Method %s: cacheable=%t\n", method, cacheable)
	}

	// Demo 4: Cache Status
	fmt.Println("\n4. Cache Status Demo")
	fmt.Printf("Cache Hit: %s\n", cache.CacheStatusHit)
	fmt.Printf("Cache Miss: %s\n", cache.CacheStatusMiss)
	fmt.Printf("Cache Bypass: %s\n", cache.CacheStatusBypass)

	fmt.Println("\nDemo completed successfully!")
}

// Redis demo (commented out since it requires Redis server)
func redisDemo() {
	fmt.Println("\n5. Redis Cache Demo (requires Redis server)")
	
	// Uncomment to test with actual Redis server
	/*
	redisConfig := cache.RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	
	redisCache := cache.NewRedis(redisConfig, 5*time.Minute)
	defer redisCache.Stop()

	// Test Redis operations
	redisCache.Set("redis:test", []byte("Hello Redis"))
	
	if value, found := redisCache.Get("redis:test"); found {
		fmt.Printf("Redis value: %s\n", string(value))
	}

	stats := redisCache.Stats()
	fmt.Printf("Redis stats: %d entries, %d bytes\n", stats.Entries, stats.Size)
	*/
	
	log.Println("Redis demo skipped (no Redis server)")
}