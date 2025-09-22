package cache

import (
	"sync"
	"time"
)

// Entry represents a cache entry with expiration
type Entry struct {
	Value     []byte
	ExpiresAt time.Time
}

// IsExpired checks if the entry has expired
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache provides a thread-safe in-memory cache with TTL support
type Cache struct {
	mu       sync.RWMutex
	entries  map[string]*Entry
	maxSize  int64
	size     int64
	defaultTTL time.Duration
	stopCh   chan struct{}
}

// New creates a new cache instance
func New(maxSize int64, defaultTTL time.Duration, cleanupPeriod time.Duration) *Cache {
	c := &Cache{
		entries:    make(map[string]*Entry),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
		stopCh:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go c.startCleanup(cleanupPeriod)

	return c
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if entry.IsExpired() {
		// Don't delete here to avoid write lock in read operation
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in the cache with the default TTL
func (c *Cache) Set(key string, value []byte) bool {
	return c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a specific TTL
func (c *Cache) SetWithTTL(key string, value []byte, ttl time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	valueSize := int64(len(value))
	
	// Check if we have an existing entry
	if existingEntry, exists := c.entries[key]; exists {
		// Remove the size of the existing entry
		c.size -= int64(len(existingEntry.Value))
	}

	// Check if adding this entry would exceed max size
	if c.size+valueSize > c.maxSize {
		// Try to make room by removing expired entries
		c.cleanupExpiredLocked()
		
		// Check again after cleanup
		if c.size+valueSize > c.maxSize {
			return false // Cannot fit
		}
	}

	// Add the entry
	c.entries[key] = &Entry{
		Value:     make([]byte, len(value)),
		ExpiresAt: time.Now().Add(ttl),
	}
	copy(c.entries[key].Value, value)
	c.size += valueSize

	return true
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entries[key]; exists {
		delete(c.entries, key)
		c.size -= int64(len(entry.Value))
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*Entry)
	c.size = 0
}

// Size returns the current size of the cache in bytes
func (c *Cache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

// Len returns the number of entries in the cache
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Stats returns cache statistics
func (c *Cache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	expired := 0
	for _, entry := range c.entries {
		if entry.IsExpired() {
			expired++
		}
	}

	return Stats{
		Entries:      len(c.entries),
		Size:         c.size,
		MaxSize:      c.maxSize,
		ExpiredEntries: expired,
	}
}

// Stats represents cache statistics
type Stats struct {
	Entries        int   `json:"entries"`
	Size           int64 `json:"size"`
	MaxSize        int64 `json:"max_size"`
	ExpiredEntries int   `json:"expired_entries"`
}

// Stop stops the cache cleanup goroutine
func (c *Cache) Stop() {
	close(c.stopCh)
}

// startCleanup starts a goroutine that periodically removes expired entries
func (c *Cache) startCleanup(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.stopCh:
			return
		}
	}
}

// cleanupExpired removes expired entries from the cache
func (c *Cache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleanupExpiredLocked()
}

// cleanupExpiredLocked removes expired entries (must be called with lock held)
func (c *Cache) cleanupExpiredLocked() {
	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
			c.size -= int64(len(entry.Value))
		}
	}
}

// GenerateKey generates a cache key from request components
func GenerateKey(method, path, query string) string {
	if query == "" {
		return method + ":" + path
	}
	return method + ":" + path + "?" + query
}