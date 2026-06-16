package main

import (
    "context"
    "log"
    "time"

    "github.com/valkey-io/valkey-go"
)

// ValkeyCache - реализация кеша через Valkey
type ValkeyCache struct {
    client valkey.Client
    ctx    context.Context
    ttl    time.Duration
}

// NewValkeyCache создаёт новый кеш на основе Valkey
func NewValkeyCache(addr string, ttl time.Duration) (*ValkeyCache, error) {
    // Создаём клиент Valkey
    client, err := valkey.NewClient(valkey.ClientOption{
        InitAddress: []string{addr},
    })
    if err != nil {
        return nil, err
    }

    // Проверяем подключение
    ctx := context.Background()
    if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
        return nil, err
    }

    log.Printf("Connected to Valkey at %s", addr)
    
    return &ValkeyCache{
        client: client,
        ctx:    ctx,
        ttl:    ttl,
    }, nil
}

// Get получает данные из кеша
func (c *ValkeyCache) Get(key string) ([]byte, bool) {
    result := c.client.Do(c.ctx, c.client.B().Get().Key(key).Build())
    
    if result.Error() != nil {
        return nil, false
    }
    
    data, err := result.AsBytes()
    if err != nil {
        return nil, false
    }
    
    return data, true
}

// Set сохраняет данные в кеш
func (c *ValkeyCache) Set(key string, data []byte) {
    // Устанавливаем значение с TTL
    err := c.client.Do(c.ctx, c.client.B().Setex().Key(key).Seconds(int64(c.ttl.Seconds())).Value(string(data)).Build()).Error()
    if err != nil {
        log.Printf("Failed to cache: %v", err)
    }
}

// Close закрывает соединение с Valkey
func (c *ValkeyCache) Close() {
    c.client.Close()
}