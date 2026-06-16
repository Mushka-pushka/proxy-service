package main

import (
    "context"
    "log"
    "time"

    "github.com/valkey-io/valkey-go"
)

type ValkeyCache struct {
    client valkey.Client
    ctx    context.Context
    ttl    time.Duration
}

func NewValkeyCache(addr string, ttl time.Duration) (*ValkeyCache, error) {
    client, err := valkey.NewClient(valkey.ClientOption{
        InitAddress: []string{addr},
    })
    if err != nil {
        return nil, err
    }

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

func (c *ValkeyCache) Set(key string, data []byte) {
    err := c.client.Do(c.ctx, c.client.B().Setex().Key(key).Seconds(int64(c.ttl.Seconds())).Value(string(data)).Build()).Error()
    if err != nil {
        log.Printf("Failed to cache: %v", err)
    }
}

func (c *ValkeyCache) Close() {
    c.client.Close()
}