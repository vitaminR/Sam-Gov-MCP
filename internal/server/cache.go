package server

import (
    "sync"
    "time"
)

type cacheItem struct {
    value      interface{}
    expiration time.Time
}

// Cache is a minimal in-memory TTL cache safe for concurrent access.
type Cache struct {
    mu    sync.RWMutex
    items map[string]cacheItem
}

// NewCache constructs an empty Cache instance.
func NewCache() *Cache { return &Cache{items: make(map[string]cacheItem)} }

// Set stores a value with a time-to-live for the given key.
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = cacheItem{value: value, expiration: time.Now().Add(ttl)}
}

// Get retrieves a non-expired value for the key, returning false if missing or expired.
func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    it, ok := c.items[key]
    c.mu.RUnlock()
    if !ok {
        return nil, false
    }
    if time.Now().After(it.expiration) {
        c.mu.Lock()
        delete(c.items, key)
        c.mu.Unlock()
        return nil, false
    }
    return it.value, true
}
