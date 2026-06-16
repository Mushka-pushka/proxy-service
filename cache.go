package main

import (
    "sync"
    "time"
)

// CacheItem хранит данные и время жизни
type CacheItem struct {
    Data      []byte
    ExpiresAt time.Time
}

// Cache - in-memory кеш с TTL
type Cache struct {
    items map[string]*CacheItem
    mu    sync.RWMutex
    ttl   time.Duration
}

// NewCache создаёт новый кеш
func NewCache(ttl time.Duration) *Cache {
    c := &Cache{
        items: make(map[string]*CacheItem),
        ttl:   ttl,
    }
    go c.cleanup()
    return c
}

// Get получает данные из кеша
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

// Set сохраняет данные в кеш
func (c *Cache) Set(key string, data []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.items[key] = &CacheItem{
        Data:      data,
        ExpiresAt: time.Now().Add(c.ttl),
    }
}

// cleanup удаляет просроченные записи каждую минуту
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

// Close - для совместимости с интерфейсом
func (c *Cache) Close() {
    // Ничего не делаем
}