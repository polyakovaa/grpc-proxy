FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY gateway/ ./gateway/
COPY gen/ ./gen/ 

WORKDIR /app/gateway

RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/gateway/gateway .
COPY gateway/config/config.yaml ./config.yaml

EXPOSE 8080
CMD ["./gateway", "-config", "config.yaml"]