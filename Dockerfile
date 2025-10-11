# Dockerfile
# Этап сборки
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o access-proxy cmd/main.go

# Финальный этап
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из этапа сборки
COPY --from=builder /app/access-proxy .
# Копируем конфиг
COPY --from=builder /app/config.yaml ./

# Создаем папку для логов
RUN mkdir -p /var/log/access-proxy

EXPOSE 8000

CMD ["./access-proxy"]