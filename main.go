package main

import (
    "bytes"
	"context"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
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

func main() {
    originURL, err := url.Parse("http://localhost:8081")
    if err != nil {
        log.Fatal(err)
    }

    proxy := httputil.NewSingleHostReverseProxy(originURL)

    client, err := valkey.NewClient(valkey.ClientOption{
        InitAddress: []string{"localhost:6379"},
    })
    if err != nil {
        log.Printf("Could not connect to Valkey: %v", err)
        log.Println("Running without Valkey (cache disabled, rate limiting disabled)")
        startServer(proxy, nil, nil)
        return
    }
    defer client.Close()

    log.Println("Connected to Valkey at localhost:6379")

    cache := &ValkeyCache{
        client: client,
        ctx:    context.Background(),
        ttl:    30 * time.Second,
    }

    limiter := NewRateLimiter(client, 10, 60*time.Second)

    startServer(proxy, cache, limiter)
}

func startServer(proxy *httputil.ReverseProxy, cache CacheInterface, limiter *RateLimiter) {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Rate limiting
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
                log.Printf("💾 Cached: %s", cacheKey)
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