FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/internal/migrations ./internal/migrations

EXPOSE 8080

CMD ["./main"]