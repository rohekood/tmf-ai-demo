# Build Stage
FROM golang:1.26.1 AS builder
WORKDIR /workspace
COPY pkg ./pkg
COPY services/shopping-cart ./services/shopping-cart
WORKDIR /workspace/services/shopping-cart
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /workspace/server ./cmd/server/main.go

# Run Stage
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /workspace/server .
USER nonroot:nonroot
ENTRYPOINT ["./server"]
