FROM golang:1.27rc2 AS builder

WORKDIR /workspace

# Copy monorepo dependencies and service source
COPY pkg ./pkg
COPY services/demo-ui/bff ./services/demo-ui/bff

WORKDIR /workspace/services/demo-ui/bff

RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /workspace/bff ./cmd/server/main.go

# Final stage
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /workspace/bff .

EXPOSE 8080

CMD ["./bff"]
