package server

import (
    "sync"
    "time"
)

type cacheItem struct {
    value      interface{}
    expiration time.Time
}

type Cache struct {
    mu    sync.RWMutex
    items map[string]cacheItem
}

func NewCache() *Cache { return &Cache{items: make(map[string]cacheItem)} }

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = cacheItem{value: value, expiration: time.Now().Add(ttl)}
}

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
