# Build Stage
FROM golang:1.26.4 AS builder
# CMD_PATH selects which binary to build (server by default; mocks override it).
ARG CMD_PATH=./cmd/server/main.go
WORKDIR /workspace
COPY pkg ./pkg
COPY services/qualification ./services/qualification
WORKDIR /workspace/services/qualification
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /workspace/server ${CMD_PATH}

# Run Stage
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /workspace/server .
COPY --from=builder --chown=nonroot:nonroot /workspace/services/qualification/internal/infrastructure/postgres/migrations ./internal/infrastructure/postgres/migrations
USER nonroot:nonroot
ENTRYPOINT ["./server"]
