FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY event_service/ ./event_service/
COPY gen/ ./gen/ 

WORKDIR /app/event_service

RUN CGO_ENABLED=0 GOOS=linux go build -o event ./cmd

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/event_service/event .
COPY event_service/config/config.yaml ./config.yaml

EXPOSE 50052
CMD ["./event", "-config", "config.yaml"]