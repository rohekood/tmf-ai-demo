# Build Stage
FROM golang:1.26.5 AS builder
WORKDIR /workspace
COPY pkg ./pkg
COPY services/customer-management ./services/customer-management
WORKDIR /workspace/services/customer-management
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /workspace/server ./cmd/server/main.go

# Run Stage
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /workspace/server .
COPY --from=builder --chown=nonroot:nonroot /workspace/services/customer-management/internal/infrastructure/postgres/migrations ./internal/infrastructure/postgres/migrations
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["./server"]
