package main

// CacheInterface - общий интерфейс для всех видов кеша
type CacheInterface interface {
    Get(key string) ([]byte, bool)
    Set(key string, data []byte)
}