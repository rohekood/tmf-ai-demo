# Build Stage
FROM golang:1.26.1 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

# Run Stage
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/server .
COPY --from=builder --chown=nonroot:nonroot /app/internal/adapter/repository/migrations ./internal/adapter/repository/migrations
USER nonroot:nonroot
# No exposed port as this service only runs RabbitMQ consumers
ENTRYPOINT ["./server"]
