# --- Стадия 1: сборка ---
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app

# Сначала только go.mod/go.sum — это даёт Docker слой, который
# пересобирается лишь когда меняются зависимости, а не при каждой правке кода
COPY go.mod go.sum ./
RUN go mod download

# Теперь копируем весь исходный код и собираем
COPY . .

# CGO_ENABLED=0 — статическая линковка, без зависимости от glibc/musl
# в runtime-образе (alpine использует musl, а не glibc — источник частых
# "no such file or directory" при попытке запустить обычный бинарник)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/orders-service ./cmd/api

# --- Стадия 2: финальный образ ---
FROM alpine:3.20

# ca-certificates понадобится позже, когда сервис начнёт делать исходящие
# HTTPS/gRPC-вызовы к другим сервисам — добавляем сразу, чтобы не
# возвращаться к этому файлу через 2 этапа
RUN apk add --no-cache ca-certificates wget
# непривилегированный пользователь — контейнер не должен работать от root
RUN adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /app/bin/orders-service .

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s \
    CMD wget --spider -q http://localhost:8080/health || exit 1

ENTRYPOINT ["./orders-service"]
