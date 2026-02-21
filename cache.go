package main

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// Cache is a simple in-memory TTL cache.
// To avoid the map-memory-leak pattern, it periodically rebuilds.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	stop    chan struct{}
}

// NewCache creates a new cache that runs periodic cleanup.
func NewCache() *Cache {
	c := &Cache{
		entries: make(map[string]cacheEntry),
		stop:    make(chan struct{}),
	}
	go c.cleanup()
	return c
}

// Get returns the cached value if present and not expired.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.value, true
}

// Set stores a value with the given TTL.
func (c *Cache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	c.entries[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

// Close stops the cleanup goroutine.
func (c *Cache) Close() {
	close(c.stop)
}

// cleanup runs every 5 minutes, rebuilding the map to reclaim memory.
func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			fresh := make(map[string]cacheEntry, len(c.entries)/2)
			for k, v := range c.entries {
				if now.Before(v.expiresAt) {
					fresh[k] = v
				}
			}
			c.entries = fresh
			c.mu.Unlock()
		case <-c.stop:
			return
		}
	}
}
