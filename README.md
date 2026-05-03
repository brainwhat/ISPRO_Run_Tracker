# running-tracker

Сервис сбора и хранения беговых тренировок с автоматическим отслеживанием личных рекордов.

## Фишка — хранит рекорды

При добавлении тренировки сервис проверяет 12 популярных дистанций (от 100м до марафона). Если пробежал чуть больше целевой (до +10%), время пересчитывается по среднему темпу.

**Пример:** 10 500 м за 63 мин = темп 6:00/км → рекорд на 10К записывается как **60:00**.

---

## Стек

- Go 1.23
- chi
- swaggo/http-swagger
- In-memory storage
- OpenAPI 3.0
- prometheus/client_golang
- Prometheus
- Grafana

---

## Структура проекта

```
.
├── cmd/server/main.go
├── internal/
│   ├── handlers/workouts.go
│   ├── metrics/
│   │   ├── metrics.go       # объявление метрик
│   │   └── middleware.go    # HTTP middleware
│   ├── models/
│   └── storage/
├── configs/
│   ├── prometheus.yml
│   └── grafana/
│       ├── provisioning/    # авто-подключение datasource и дашбордов
│       └── dashboards/      # JSON дашборд
├── scripts/run-all.sh
├── Makefile
└── openapi.yaml
```

---

## Быстрый старт

```bash
make run-all
```

| Сервис     | Адрес                                    |
|------------|------------------------------------------|
| API        | http://localhost:8080                    |
| Swagger    | http://localhost:8080/swagger/index.html |
| Метрики    | http://localhost:8080/metrics            |
| Prometheus | http://localhost:9090                    |
| Grafana    | http://localhost:3000                    |

Первый раз — установить инструменты: `make install-tools`

Остановить стек: **Ctrl-C** или `make stop`

---

## Эндпоинты

```
GET    /workouts          список тренировок (?limit=20&offset=0)
POST   /workouts          добавить тренировку → {workout, new_records}
GET    /workouts/{id}     тренировка по ID
PUT    /workouts/{id}     обновить тренировку
DELETE /workouts/{id}     удалить тренировку

GET    /records           личные рекорды
```

Полная спецификация: `GET /openapi.yaml`

## Формат запроса

```json
POST /workouts
{
  "name": "Воскресный забег",
  "distance_km": 10.5,
  "duration_min": 63.0,
  "avg_heart_rate": 158,
  "date": "2026-04-29"
}
```

```json
201 Created
{
  "workout": { "id": 11, "distance_km": 10.5, "pace_min_per_km": 6, ... },
  "new_records": [
    {
      "distance_name": "10К",
      "time_min": 60,
      "message": "Первый рекорд на 10К: 1:00:00 (темп 6:00/км)!"
    }
  ]
}
```

---

## Метрики

Prometheus scrape: `GET /metrics`, интервал 5с. Дашборд подгружается автоматически при `make run-all`.

**HTTP:**

| Метрика | Тип | Лейблы | Описание |
|---|---|---|---|
| `http_requests_total` | счётчик | `method`, `path`, `status_code` | Суммарное число запросов |
| `http_request_duration_seconds` | гистограмма | `method`, `path` | Латентность |
| `http_requests_in_flight` | gauge | — | Запросы в обработке прямо сейчас |

**Продуктовые:**

| Метрика | Тип | Лейблы | Описание |
|---|---|---|---|
| `workouts_created_total` | счётчик | — | Создано тренировок за всё время |
| `workouts_deleted_total` | счётчик | — | Удалено тренировок за всё время |
| `workouts_active_total` | gauge | — | Текущее количество тренировок |
| `workout_distance_km` | гистограмма | — | Распределение дистанций новых тренировок |
| `personal_records_broken_total` | счётчик | `distance_name` | Побитые личные рекорды по дистанциям |

### Примеры PromQL

```promql
# RPS по эндпоинтам
sum(rate(http_requests_total[1m])) by (method, path)

# p95 латентность
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[1m])) by (le))

# Доля ошибок
sum(rate(http_requests_total{status_code=~"[45].."}[1m])) / sum(rate(http_requests_total[1m]))

# Медианная дистанция новых тренировок
histogram_quantile(0.50, sum(rate(workout_distance_km_bucket[10m])) by (le))
```

---

## Скриншоты

![Swagger UI](imgs/swagger.png)
![Grafana](imgs/grafana.png)
