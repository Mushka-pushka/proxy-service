package main

import (
    "bytes"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "time"
)

// ResponseRecorder перехватывает ответ для кеширования
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
    // Парсим URL origin-сервера
    originURL, err := url.Parse("http://localhost:8081")
    if err != nil {
        log.Fatal(err)
    }

    // Создаём reverse proxy
    proxy := httputil.NewSingleHostReverseProxy(originURL)

    // Создаём кеш с TTL 30 секунд
    cache := NewCache(30 * time.Second)

    // Обработчик всех запросов
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Кешируем только GET-запросы
        if r.Method == "GET" {
            cacheKey := r.URL.String()

            // Проверяем кеш
            if cachedData, found := cache.Get(cacheKey); found {
                log.Printf("Cache HIT: %s", cacheKey)
                w.Write(cachedData)
                return
            }

            log.Printf("Cache MISS: %s", cacheKey)

            // Создаём рекордер для перехвата ответа
            recorder := NewResponseRecorder(w)

            // Перенаправляем запрос на origin
            proxy.ServeHTTP(recorder, r)

            // Если статус 200, сохраняем в кеш
            if recorder.status == 200 || recorder.status == 0 {
                cache.Set(cacheKey, recorder.GetBody())
                log.Printf("Cached: %s", cacheKey)
            }
        } else {
            // Для не-GET запросов просто проксируем
            proxy.ServeHTTP(w, r)
        }
    })

    log.Println("Proxy server with cache starting on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}