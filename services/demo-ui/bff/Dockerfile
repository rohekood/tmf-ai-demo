FROM golang:1.26.1 AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o bff cmd/server/main.go

# Final stage
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/bff .

EXPOSE 8080

CMD ["./bff"]
