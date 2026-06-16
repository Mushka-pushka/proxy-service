package main

import (
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
)

func main() {
    // Парсим URL origin-сервера
    originURL, err := url.Parse("http://localhost:8081")
    if err != nil {
        log.Fatal(err)
    }

    // Создаём reverse proxy
    proxy := httputil.NewSingleHostReverseProxy(originURL)

    // Обработчик всех запросов
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Proxying request: %s %s", r.Method, r.URL.Path)
        proxy.ServeHTTP(w, r)
    })

    log.Println("Proxy server starting on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}