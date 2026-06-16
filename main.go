package main

import (
    "bytes"
    "context"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "os"
    "strconv"
    "time"

    "github.com/valkey-io/valkey-go"
)

type ResponseRecorder struct {
    http.ResponseWriter
    body   *bytes.Buffer
    status int
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
    return &ResponseRecorder{
        ResponseWriter: w,
        body:           bytes.NewBuffer(nil),
    }
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
    r.body.Write(data)
    return r.ResponseWriter.Write(data)
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
    r.status = statusCode
    r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) GetBody() []byte {
    return r.body.Bytes()
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

func main() {
    originURL := getEnv("ORIGIN_URL", "http://localhost:8081")
    valkeyAddr := getEnv("VALKEY_ADDR", "localhost:6379")
    cacheTTL := getEnvInt("CACHE_TTL", 30)
    rateLimit := getEnvInt("RATE_LIMIT", 10)
    rateWindow := getEnvInt("RATE_WINDOW", 60)

    targetURL, err := url.Parse(originURL)
    if err != nil {
        log.Fatalf("Invalid ORIGIN_URL: %v", err)
    }

    proxy := httputil.NewSingleHostReverseProxy(targetURL)

    client, err := valkey.NewClient(valkey.ClientOption{
        InitAddress: []string{valkeyAddr},
    })
    if err != nil {
        log.Printf("Could not connect to Valkey at %s: %v", valkeyAddr, err)
        log.Println("Running without Valkey (cache disabled, rate limiting disabled)")
        startServer(proxy, nil, nil)
        return
    }
    defer client.Close()

    ctx := context.Background()
    if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
        log.Printf("Valkey ping failed: %v", err)
        log.Println("Running without Valkey (cache disabled, rate limiting disabled)")
        startServer(proxy, nil, nil)
        return
    }

    log.Printf("Connected to Valkey at %s", valkeyAddr)

    cache := &ValkeyCache{
        client: client,
        ctx:    context.Background(),
        ttl:    time.Duration(cacheTTL) * time.Second,
    }

    limiter := NewRateLimiter(client, rateLimit, time.Duration(rateWindow)*time.Second)

    startServer(proxy, cache, limiter)
}

func startServer(proxy *httputil.ReverseProxy, cache CacheInterface, limiter *RateLimiter) {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

        if limiter != nil {
            ip := r.RemoteAddr
            allowed, err := limiter.Allow(ip)
            if err != nil {
                log.Printf("Rate limiter error: %v", err)
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                return
            }
            if !allowed {
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }
        }

        if r.Method == "GET" && cache != nil {
            cacheKey := r.URL.String()

            if cachedData, found := cache.Get(cacheKey); found {
                log.Printf("Cache HIT: %s", cacheKey)
                w.Write(cachedData)
                return
            }

            log.Printf("Cache MISS: %s", cacheKey)

            recorder := NewResponseRecorder(w)
            proxy.ServeHTTP(recorder, r)

            if recorder.status == 200 || recorder.status == 0 {
                cache.Set(cacheKey, recorder.GetBody())
                log.Printf("Cached: %s", cacheKey)
            }
        } else {
            proxy.ServeHTTP(w, r)
        }
    })

    log.Println("Proxy server with Valkey cache and rate limiting starting on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}