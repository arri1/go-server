# Многоступенчатая сборка: сначала компилируем, потом кладём только бинарник.
# Так итоговый образ меньше и без компилятора / исходников.

# ---- этап сборки ----
FROM golang:1.24-alpine AS builder

# git нужен некоторым go-модулям при загрузке зависимостей
RUN apk add --no-cache git

WORKDIR /src

# Сначала копируем манифесты модулей — слой кэшируется, пока go.mod/go.sum не менялись.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 — статическая сборка без libc, удобно для alpine.
# -o /out/api — имя выходного файла.
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

# ---- этап запуска ----
FROM alpine:3.20

# ca-certificates — HTTPS-клиент; wget — healthcheck внутри контейнера.
RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

# Копируем только готовый бинарник из предыдущего этапа.
COPY --from=builder /out/api /app/api

EXPOSE 8080

# HEALTHCHECK использует ту же /health, что и оркестраторы.
HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
  CMD wget -qO- http://127.0.0.1:8080/health || exit 1

# Процесс контейнера = наш сервер. PID 1 получит SIGTERM при docker stop.
ENTRYPOINT ["/app/api"]
