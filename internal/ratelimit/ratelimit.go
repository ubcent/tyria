package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	mu          sync.Mutex
	tokens      int
	maxTokens   int
	refillRate  int           // tokens per period
	refillPeriod time.Duration
	lastRefill  time.Time
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(maxTokens, refillRate int, refillPeriod time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:       maxTokens,
		maxTokens:    maxTokens,
		refillRate:   refillRate,
		refillPeriod: refillPeriod,
		lastRefill:   time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token if so
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tb.refill(now)

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// AllowN checks if n requests are allowed and consumes n tokens if so
func (tb *TokenBucket) AllowN(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tb.refill(now)

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}

	return false
}

// refill adds tokens to the bucket based on elapsed time
func (tb *TokenBucket) refill(now time.Time) {
	elapsed := now.Sub(tb.lastRefill)
	
	if elapsed >= tb.refillPeriod {
		periods := int(elapsed / tb.refillPeriod)
		tokensToAdd := periods * tb.refillRate
		
		tb.tokens += tokensToAdd
		if tb.tokens > tb.maxTokens {
			tb.tokens = tb.maxTokens
		}
		
		tb.lastRefill = tb.lastRefill.Add(time.Duration(periods) * tb.refillPeriod)
	}
}

// Tokens returns the current number of available tokens
func (tb *TokenBucket) Tokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	now := time.Now()
	tb.refill(now)
	
	return tb.tokens
}

// Limiter manages rate limiting for multiple clients/endpoints
type Limiter struct {
	mu      sync.RWMutex
	buckets map[string]*TokenBucket
	config  Config
}

// Config holds rate limiter configuration
type Config struct {
	MaxTokens    int           `json:"max_tokens"`
	RefillRate   int           `json:"refill_rate"`
	RefillPeriod time.Duration `json:"refill_period"`
	CleanupPeriod time.Duration `json:"cleanup_period"`
}

// NewLimiter creates a new rate limiter
func NewLimiter(config Config) *Limiter {
	l := &Limiter{
		buckets: make(map[string]*TokenBucket),
		config:  config,
	}

	// Start cleanup goroutine
	go l.startCleanup()

	return l
}

// Allow checks if a request is allowed for the given key
func (l *Limiter) Allow(key string) bool {
	return l.AllowN(key, 1)
}

// AllowN checks if n requests are allowed for the given key
func (l *Limiter) AllowN(key string, n int) bool {
	bucket := l.getBucket(key)
	return bucket.AllowN(n)
}

// getBucket gets or creates a token bucket for the given key
func (l *Limiter) getBucket(key string) *TokenBucket {
	l.mu.RLock()
	bucket, exists := l.buckets[key]
	l.mu.RUnlock()

	if exists {
		return bucket
	}

	// Create new bucket
	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock
	if bucket, exists = l.buckets[key]; exists {
		return bucket
	}

	bucket = NewTokenBucket(l.config.MaxTokens, l.config.RefillRate, l.config.RefillPeriod)
	l.buckets[key] = bucket

	return bucket
}

// Stats returns statistics about the rate limiter
func (l *Limiter) Stats() Stats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := Stats{
		TotalBuckets: len(l.buckets),
		Buckets:      make(map[string]BucketStats),
	}

	for key, bucket := range l.buckets {
		stats.Buckets[key] = BucketStats{
			Tokens: bucket.Tokens(),
			MaxTokens: bucket.maxTokens,
		}
	}

	return stats
}

// Stats represents rate limiter statistics
type Stats struct {
	TotalBuckets int                    `json:"total_buckets"`
	Buckets      map[string]BucketStats `json:"buckets"`
}

// BucketStats represents statistics for a single bucket
type BucketStats struct {
	Tokens    int `json:"tokens"`
	MaxTokens int `json:"max_tokens"`
}

// Reset removes all buckets
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buckets = make(map[string]*TokenBucket)
}

// startCleanup starts a goroutine that periodically cleans up unused buckets
func (l *Limiter) startCleanup() {
	if l.config.CleanupPeriod <= 0 {
		return // Cleanup disabled
	}

	ticker := time.NewTicker(l.config.CleanupPeriod)
	defer ticker.Stop()

	for range ticker.C {
		l.cleanup()
	}
}

// cleanup removes buckets that haven't been used recently
func (l *Limiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for key, bucket := range l.buckets {
		// Remove buckets that are full and haven't been used for a while
		if bucket.tokens == bucket.maxTokens && 
		   now.Sub(bucket.lastRefill) > l.config.CleanupPeriod*2 {
			delete(l.buckets, key)
		}
	}
}

// GenerateClientKey generates a rate limit key for a client
func GenerateClientKey(clientID string) string {
	return "client:" + clientID
}

// GenerateEndpointKey generates a rate limit key for an endpoint
func GenerateEndpointKey(endpoint string) string {
	return "endpoint:" + endpoint
}

// GenerateClientEndpointKey generates a rate limit key for a client-endpoint combination
func GenerateClientEndpointKey(clientID, endpoint string) string {
	return "client:" + clientID + ":endpoint:" + endpoint
}