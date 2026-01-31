# Architecture: General Standards

## 1. Monorepo Structure
*   **Services**: All business logic resides in `services/<service-name>`.
*   **Shared Code**: Infrastructure libraries reside in `pkg/`. Domain logic MUST NOT be shared.

## 2. Deployability
*   **Independent Deployment**: Every directory in `services/` MUST be independently deployable.
*   **Entrypoint**: Each service MUST have a `cmd/server/main.go` file.
*   **Containerization**: Each service MUST have a `Dockerfile` at its root.
*   **Configuration**: Configuration MUST be loaded from Environment Variables.

## 3. Communication
*   **Inter-Service**: Strictly Asynchronous via RabbitMQ (EDA).
*   **No Direct DB Access**: A service MUST NOT access another service's database.

## 4. Observability
*   **Logging**: Use structured JSON logging (`log/slog`).
*   **Tracing**: Propagate `X-Correlation-ID` and `X-User-ID` across all boundaries.
