.PHONY: run tidy swagger docker-up docker-down

# Локальный запуск API (Postgres должен уже работать, например через docker compose up db -d)
run:
	go run ./cmd/api

# Подтянуть зависимости и подчистить go.mod / go.sum
tidy:
	go mod tidy

# Пересобрать swagger-спецификацию из комментариев @Summary / @Router
swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs

# Собрать образы и поднять API + Postgres
docker-up:
	docker compose up --build

# Остановить контейнеры (том с данными Postgres сохранится)
docker-down:
	docker compose down
