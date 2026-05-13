FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Исправлено: добавлены флаги окружения и парсинга внутренних пакетов
RUN CGO_ENABLED=0 swag init -g cmd/main.go -o docs --parseDependency --parseInternal

# Исправлено: добавлена статическая сборка для alpine
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/internal/migrations ./internal/migrations
# Исправлено: добавлен перенос папки с документацией в финальный образ
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./main"]
