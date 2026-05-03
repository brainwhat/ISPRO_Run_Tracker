# running-tracker

Сервис сбора и хранения беговых тренирвок.

## Фишка - хранит рекорды

При добавлении тренировки сервис проверяет 12 самых популярных дистанций - от 100м до марафона. Если пробежал чуть больше целевой (до +10%), время пересчитывается по среднему темпу.

**Пример:** 10 500 м за 63 мин = темп 6:00/км → рекорд на 10К записывается как **60:00**.

---

## Быстрый старт

```bash
go mod tidy && go run ./cmd/server
```

→ `http://localhost:8080`  
→ Swagger: `http://localhost:8080/swagger/index.html`


## Эндпоинты

```
GET    /workouts          - список тренировок (?limit=20&offset=0)
POST   /workouts          - добавить тренировку → {workout, new_records}
GET    /workouts/{id}     - тренировка по ID
PUT    /workouts/{id}     - обновить тренировку
DELETE /workouts/{id}     - удалить тренировку

GET    /records           - личные рекорды
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
  "workout": {
    "id": 11,
    "distance_km": 10.5,
    "duration_min": 63,
    "pace_min_per_km": 6,
    ...
  },
  "new_records": [
    {
      "distance_name": "10К",
      "time_min": 60,
      "pace_min_per_km": 6,
      "message": "Первый рекорд на 10К: 1:00:00 (темп 6:00/км)!"
    }
  ]
}
```

---

## Стек

- Go 1.22
- chi 
- swaggo/http-swagger 
- in-memory storage 
- OpenAPI 3.0

---

## Скриншоты

![Swagger UI](imgs/swagger.png)
