# Implementation Plan - Asynchronous Party Management Service

## Goal Description
Implement the **Party Management Service** (TMF632) as a purely asynchronous microservice. It will handle the lifecycle of `Individual` and `Organization` entities using **RabbitMQ** for all communication (Essentials: Commands, Events, Queries).

## User Review Required
> [!IMPORTANT]
> **RabbitMQ**: Switched from NATS to RabbitMQ per user request.
> **Testcontainers**: Integration tests will use `testcontainers-go` to manage ephemeral infrastructure.

## Proposed Changes

### 1. Project Scaffolding
Establish the internal directory structure for a clean "Hexagonal" or "Clean" Architecture.

#### [MODIFY] [go.mod](file:///home/raino/IdeaProject/tmf/services/party-management/go.mod)
- Add dependencies: `github.com/rabbitmq/amqp091-go`, `github.com/lib/pq`.
- Test dependencies: `github.com/testcontainers/testcontainers-go`.

#### [NEW] [internal/config/config.go](file:///home/raino/IdeaProject/tmf/services/party-management/internal/config/config.go)
- Configuration struct (RabbitMQ URL, Postgres DSN, etc.).

### 2. Infrastructure Setup (Dev)

#### [NEW] [docker-compose.yml](file:///home/raino/IdeaProject/tmf/docker-compose.yml)
- Services: `rabbitmq` (management plugin enabled), `postgres`.

### 3. Domain Layer
Define the core entities.

#### [NEW] [internal/domain/party.go](file:///home/raino/IdeaProject/tmf/services/party-management/internal/domain/party.go)
- Structs for `Party`, `Individual`, `Organization`.

### 4. Infrastructure Layer (Driven Adapters)

#### [NEW] [internal/infrastructure/postgres/party_repository.go](file:///home/raino/IdeaProject/tmf/services/party-management/internal/infrastructure/postgres/party_repository.go)
- GORM implementation for Party storage.
- Uses `golang-migrate` for schema versioning.

#### [NEW] [internal/infrastructure/rabbitmq/publisher.go](file:///home/raino/IdeaProject/tmf/services/party-management/internal/infrastructure/rabbitmq/publisher.go)
- Methods to publish events to a Topic Exchange (e.g., `tmf.events`).

### 5. Interface Layer (Driving Adapters)

#### [NEW] [internal/transport/rabbitmq/listener.go](file:///home/raino/IdeaProject/tmf/services/party-management/internal/transport/rabbitmq/listener.go)
- Listen on `cmd.party.create` queue.
- Listen on `query.party.get` queue (Handle RPC style replies).

#### [NEW] [cmd/server/main.go](file:///home/raino/IdeaProject/tmf/services/party-management/cmd/server/main.go)
- Wire up Config, Repos, RabbitMQ connection, and Listeners.

## Verification Plan

### Automated Tests
- **Integration Tests**: Use **Testcontainers** to spin up ephemeral RabbitMQ and Postgres instances for running tests.
    - Verify `PartyRepository` against a real Postgres container.
    - Verify RabbitMQ publishing/consuming against a real RabbitMQ container.

### Manual Verification
- Run `docker-compose up`.
- Access RabbitMQ Management UI (`http://localhost:15672`).
- Run the service and verify it connects.
