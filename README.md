# Proxy Service with Cache and Rate Limiting

Reverse-proxy сервис на Go с кешированием (Valkey) и rate limiting для защиты от перегрузок.


# Описание

Сервис представляет собой reverse-proxy, который:
- Принимает HTTP-запросы
- Кеширует GET-запросы (TTL 30 секунд)
- Ограничивает количество запросов с одного IP
- Пересылает запросы на origin-сервер

**Технологии:**
- Go 1.26.4
- Valkey (Redis-совместимый кеш)
- Docker & Docker Compose
- Nginx (тестовый origin)
- k6 (нагрузочное тестирование)
- Grafana Cloud (визуализация)


## Структура проекта

```
proxy-service/
├── main.go                        # Точка входа, прокси-сервер
├── cache.go                       # In-memory кеш
├── cache_interface.go             # Интерфейс кеша
├── cache_valkey.go                # Реализация кеша через Valkey
├── limiter.go                     # Rate limiter (IP, Valkey)
├── docker-compose.yml             # Docker Compose
├── .gitignore                     # Игнорируемые файлы
│
├── docker/
│   ├── Dockerfile                 # Multi-stage Dockerfile
│   └── nginx/
│       └── default.conf           # Конфиг Nginx (origin)
│
├── k6/                            # Нагрузочные тесты
│   ├── smoke-test.js              # Дымовой тест
│   ├── baseline-without-proxy.js  # 500 RPS без прокси
│   ├── baseline-with-proxy.js     # 500 RPS через прокси
│   └── stress-test.js             # Стресс-тест 100→5000 RPS
│
└── scripts/
    └── plot-graphs.py             # Построение графиков
```


# Запуск

## 1. Запуск через Docker Compose (рекомендуемый способ)

### Клонировать репозиторий
``` bash
git clone https://github.com/Mushka-pushka/proxy-service.git
cd proxy-service
```
### Запустить все сервисы
``` bash
docker-compose up --build
```

### Сервисы:
```
Прокси: http://localhost:8080
Origin (nginx): http://localhost:8081
Valkey: localhost:6379
```

## 2. Локальный запуск (без Docker)

### Установить зависимости
``` bash 
go mod download
```
### Запустить origin
``` bash 
go run origin/main.go
```
### Запустить прокси
``` bash 
go run main.go cache.go cache_interface.go cache_valkey.go limiter.go
```

## Нагрузочное тестирование
### Установка k6
```
Windows (скачать .msi)
Mac      brew install k6
Linux    sudo apt-get install k6
```

## Результаты тестов

| Тест          | Запросы   | RPS     | Ошибки        | P95     |
|-------------- |-----------|---------|-------------- |-------- |
| **Smoke**     | 10        | ~1      | 0%            | 12.46 мс|
| **Без прокси**| 74 989    | 500     | 0%            | 1.14 мс |
| **С прокси**  | 74 988    | 500     | 0%            | 2.73 мс |
| **Stress**    | 1 743 814 | до 5000 | 86% (на пике) | 1.74 с  |

### Ключевые выводы

1. Прокси работает и обрабатывает запросы
2. Кеш ускоряет ответы
3. Rate limiter защищает систему
4. Система выдерживает нагрузку до 2000 RPS
5. При 5000 RPS система перегружается (86% ошибок)


### Запуск тестов
```bash
# Smoke-тест (проверка работы)
k6 run k6/smoke-test.js
# Baseline без прокси (500 RPS)
k6 run k6/baseline-without-proxy.js
# Baseline с прокси (500 RPS)
k6 run k6/baseline-with-proxy.js
# Stress-тест (100→5000 RPS)
k6 run k6/stress-test.js

# Сохранение результатов в JSON
k6 run --out json=results/test-name.json k6/test-name.js
```

# Построение графиков

### Установка Python зависимостей
``` bash
pip install matplotlib
```
### Запуск скрипта
``` bash 
python scripts/plot-graphs.py
```

Все графики сохраняются в папку `results/graphs/`:
- `smoke-test.png` — Smoke-тест
- `baseline-без-прокси.png` — Без прокси (500 RPS)
- `baseline-с-прокси.png` — С прокси (500 RPS)
- `comparison.png` — Сравнение
- `stress-test.png` — Стресс-тест


# Docker

### Сборка образа
``` bash 
docker build -f docker/Dockerfile -t proxy-service .
```
### Запуск через Docker Compose
``` bash 
docker-compose up --build
```
### Остановка
``` bash 
docker-compose down
```

## Переменные окружения
| Переменная    | По умолчанию            | Описание                    |
|---------------|-------------------------|-----------------------------|
| `ORIGIN_URL`  | `http://localhost:8081` | URL origin-сервера          |
| `VALKEY_ADDR` | `localhost:6379`        | Адрес Valkey                |
| `CACHE_TTL`   | `30`                    | Время жизни кеша (сек)      |
| `RATE_LIMIT`  | `10`                    | Максимум запросов в минуту  |
| `RATE_WINDOW` | `60`                    | Окно для rate limiter (сек) |


## Используемые инструменты

| Инструмент         | Версия | Назначение               |
|--------------------|--------|--------------------------|
| **Go**             | 1.26.4 | Язык программирования    |
| **Docker**         | 27.3.1 | Контейнеризация          |
| **Docker Compose** | —      | Оркестрация контейнеров  |
| **Valkey**         | 9.1.0  | Кеш и rate limiter       |
| **Nginx**          | Alpine | Тестовый origin-сервер   |
| **k6**             | 2.0.0  | Нагрузочное тестирование |
| **Grafana Cloud**  | —      | Визуализация результатов |
| **Python**         | 3.12.7 | Построение графиков      |
| **Matplotlib**     | 3.10.7 | Библиотека для графиков  |
| **Git**            | 2.53.0 | Контроль версий          |
| **GitHub**         | —      | Хостинг репозитория      |