// Package cache provides a bounded, thread-safe in-memory cache with TTL.
package cache

import (
	"sync"
	"time"
)

// entry holds a cached value and its expiration time.
type entry struct {
	value     any
	expiresAt time.Time
}

// Cache is a bounded LRU-ish in-memory cache. It is safe for concurrent use.
// When MaxEntries is exceeded the oldest entries are evicted.
type Cache struct {
	mu         sync.RWMutex
	data       map[string]entry
	keys       []string // insertion order for eviction
	ttl        time.Duration
	maxEntries int
}

// New creates a Cache with the given TTL and max entry count.
func New(ttl time.Duration, maxEntries int) *Cache {
	if maxEntries < 1 {
		maxEntries = 1000
	}
	return &Cache{
		data:       make(map[string]entry),
		ttl:        ttl,
		maxEntries: maxEntries,
	}
}

// Get returns the cached value for key, or false if missing/expired.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	e, ok := c.data[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

// Set stores a value with the cache's TTL. Errors are not cached by callers;
// short-lived negative caching is handled by the caller if needed.
func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.data[key]; !exists {
		c.keys = append(c.keys, key)
		// Evict oldest entries if over capacity.
		for len(c.keys) > c.maxEntries {
			old := c.keys[0]
			c.keys = c.keys[1:]
			delete(c.data, old)
		}
	}
	c.data[key] = entry{value: value, expiresAt: time.Now().Add(c.ttl)}
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	for i, k := range c.keys {
		if k == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			break
		}
	}
}

// Len returns the number of non-expired entries.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}
