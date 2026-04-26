# Build Stage
FROM golang:1.26.2 AS builder
WORKDIR /workspace
COPY pkg ./pkg
COPY services/product-catalog-management ./services/product-catalog-management
WORKDIR /workspace/services/product-catalog-management
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /workspace/server ./cmd/main.go

# Run Stage
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /workspace/server .
COPY --from=builder --chown=nonroot:nonroot /workspace/services/product-catalog-management/internal/adapter/repository/migrations ./internal/adapter/repository/migrations
USER nonroot:nonroot
# No exposed port as this service only runs RabbitMQ consumers
ENTRYPOINT ["./server"]
