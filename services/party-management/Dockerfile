# Build Stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Run Stage
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/server .
COPY --from=builder --chown=nonroot:nonroot /app/internal/infrastructure/postgres/migrations ./internal/infrastructure/postgres/migrations
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["./server"]
