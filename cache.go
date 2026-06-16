package main

import (
    "sync"
    "time"
)

type CacheItem struct {
    Data      []byte
    ExpiresAt time.Time
}

type Cache struct {
    items map[string]*CacheItem
    mu    sync.RWMutex
    ttl   time.Duration
}

func NewCache(ttl time.Duration) *Cache {
    c := &Cache{
        items: make(map[string]*CacheItem),
        ttl:   ttl,
    }
    go c.cleanup()
    return c
}

func (c *Cache) Get(key string) ([]byte, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    item, exists := c.items[key]
    if !exists {
        return nil, false
    }

    if time.Now().After(item.ExpiresAt) {
        delete(c.items, key)
        return nil, false
    }

    return item.Data, true
}

func (c *Cache) Set(key string, data []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.items[key] = &CacheItem{
        Data:      data,
        ExpiresAt: time.Now().Add(c.ttl),
    }
}

func (c *Cache) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for key, item := range c.items {
            if now.After(item.ExpiresAt) {
                delete(c.items, key)
            }
        }
        c.mu.Unlock()
    }
}