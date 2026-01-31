# General Architecture Guidelines

This document establishes the mandatory architectural standards for all services within the TMF project. All new services and refactors must strictly adhere to these rules.

## 1. Core Principles

### 1.1 Context Propagation
**Requirement**: All methods across all layers (Domain, Infrastructure, Transport) MUST accept `context.Context` as their first parameter.
**Rationale**:
- **Timeout Control**: Ensures database and broker operations are bound by time.
- **Cancellation**: Allows the service to halt processing upon client disconnect or system shutdown.
- **Observability**: Carries trace IDs and other relevant metadata.

### 1.2 Error Handling
- **Domain Errors**: Use domain-specific errors (e.g., `ErrEntityNotFound`) defined in `internal/core/domain/errors.go`.
- **Wrapping**: Always wrap low-level errors with `%w` to preserve context.
- **Mapping**: Map infrastructure-specific errors (e.g., `gorm.ErrRecordNotFound`) to domain errors at the Adapter boundary.

### 1.3 Resilience
- **Graceful Shutdown**: Implement `Close()` methods for all long-running components (Servers, Consumers).
- **Signal Handling**: Use `os.Signal` to catch `SIGINT` / `SIGTERM` and wait for active processes to finish (`sync.WaitGroup`).
- **Connection Managers**: Use retry logic for infrastructure connections (RabbitMQ, Postgres).

### 1.4 Logging
- **Structured Logging**: Use `log/slog`.
- **Attributes**: Include TraceID, UserID, and relevant entity IDs in log attributes.

### 1.5 Testing Standards
- **Mandatory Coverage**: All functionality must be covered by both **Unit Tests** (for domain logic and isolated components) and **Integration Tests** (for use cases, event flows, and database interactions).
- **Zero Regressions**: No code shall be merged without passing tests that cover the new or modified functionality.
- **Tools**: Use the standard Go `testing` package.

## 2. Structural Standards (Clean Architecture)

Services must follow the **Hexagonal / Clean Architecture** pattern to ensure testability and independence from frameworks.

### 2.1 Directory Structure
```text
internal/
├── core/                   # The Inner Ring (Pure Business Logic)
│   ├── domain/             # Entities, Value Objects, Logic
│   └── ports/              # Interfaces (Input/Output definitions)
│
├── usecase/                # Application Orchestration
│   └── ...                 # Use Case Interactors (e.g. CreateOrder)
│
└── adapter/                # The Outer Ring (Implementation details)
    ├── handler/            # Driving Adapters (HTTP, RabbitMQ Consumers)
    ├── repository/         # Driven Adapters (Database, File System)
    └── gateway/            # Driven Adapters (External Service Clients)
```

### 2.2 Rules of Dependency
- **Direction**: Dependencies point **INWARDS**.
    - `adapter` depends on `core/ports` and `usecase`.
    - `usecase` depends on `core/domain` and `core/ports`.
    - `core` depends on **NOTHING**.
- **Isolation**: Domain logic must never import `gorm`, `amqp`, or `gin`.

## 3. Communication Standards (Event-Driven)

### 3.1 Asynchronous First
- **State Changes**: All state mutations must emit an **Integration Event** to RabbitMQ.
- **Transactional Outbox**: To guarantee consistency, write the Event to the DB in the same transaction as the Entity, then publish it asynchronously.

### 3.2 Synchronous vs Asynchronous
- **Commands (Writes)**: Use Asynchronous Command Queues or Outbox Events.
- **Queries (Reads)**: Use **Async Request-Reply (RPC)** over RabbitMQ to avoid direct HTTP coupling between services where possible.
- **BFF/UI**: The BFF communicates with the Frontend via HTTP/WebSockets, but communicates with Backend Services via RabbitMQ.
