# go-server

Учебный REST API на Go: CRUD для `Post`, PostgreSQL и Swagger. Всё поднимается через Docker.

## Запуск

```bash
docker compose up --build
```

- API: http://localhost:8080
- Swagger UI: http://localhost:8080/swagger/index.html
- Postgres: `localhost:5432` (user/pass/db: `postgres` / `postgres` / `go_server`)

Остановка: `docker compose down`. Том с данными БД сохранится.

## Эндпоинты

| Метод  | Путь              | Что делает        |
|--------|-------------------|-------------------|
| GET    | `/health`         | проверка живости  |
| GET    | `/api/v1/posts`   | список постов     |
| GET    | `/api/v1/posts/:id` | один пост       |
| POST   | `/api/v1/posts`   | создать           |
| PUT    | `/api/v1/posts/:id` | обновить        |
| DELETE | `/api/v1/posts/:id` | удалить         |

Пример создания:

```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H 'Content-Type: application/json' \
  -d '{"title":"Привет, Go","content":"Первый пост"}'
```

## Локально без контейнера API

```bash
docker compose up -d db
cp .env.example .env
# DATABASE_URL уже смотрит на localhost:5432
export $(grep -v '^#' .env | xargs)
go run ./cmd/api
```

Пересобрать Swagger после правок аннотаций:

```bash
make swagger
```

## Как устроен код

```
cmd/api          — точка входа (main)
internal/config  — переменные окружения
internal/database — подключение и миграция
internal/models  — структуры Post и DTO
internal/repository — SQL (CRUD)
internal/handler — HTTP-обработчики
internal/router  — маршруты + Swagger
docs             — сгенерированная OpenAPI-спецификация
```
