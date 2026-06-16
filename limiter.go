package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "time"

    "github.com/valkey-io/valkey-go"
)

type RateLimiter struct {
    client  valkey.Client
    ctx     context.Context
    maxHits int          
    window  time.Duration 
}

func NewRateLimiter(client valkey.Client, maxHits int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        client:  client,
        ctx:     context.Background(),
        maxHits: maxHits,
        window:  window,
    }
}

func getIPFromAddr(addr string) string {
    host, _, err := net.SplitHostPort(addr)
    if err != nil {
        return addr
    }
    return host
}

func (rl *RateLimiter) Allow(addr string) (bool, error) {

    ip := getIPFromAddr(addr)
    key := fmt.Sprintf("rate_limit:%s", ip)
    
    result := rl.client.Do(rl.ctx, rl.client.B().Incr().Key(key).Build())
    if result.Error() != nil {
        return false, result.Error()
    }
    
    count, err := result.AsInt64()
    if err != nil {
        return false, err
    }
    
    if count == 1 {
        err = rl.client.Do(rl.ctx, rl.client.B().Expire().Key(key).Seconds(int64(rl.window.Seconds())).Build()).Error()
        if err != nil {
            return false, err
        }
    }
    
    if count > int64(rl.maxHits) {
        log.Printf("Rate limit exceeded for IP: %s (hits: %d/%d)", ip, count, rl.maxHits)
        return false, nil
    }
    
    log.Printf("Rate limit: %s -> %d/%d", ip, count, rl.maxHits)
    return true, nil
}