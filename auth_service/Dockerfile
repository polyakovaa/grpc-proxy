FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY auth_service/ ./auth_service/
COPY gen/ ./gen/ 

WORKDIR /app/auth_service

RUN CGO_ENABLED=0 GOOS=linux go build -o auth ./cmd

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/auth_service/auth .
COPY auth_service/config/config.yaml ./config.yaml

EXPOSE 50051
CMD ["./auth", "-config", "config.yaml"]