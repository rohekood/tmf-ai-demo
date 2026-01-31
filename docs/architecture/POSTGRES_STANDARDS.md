# Architecture: Postgres & Persistence Standards

## Overview
All services in the TMF monorepo MUST follow these strict standards for PostgreSQL persistence to ensure consistency, testability, and reliability.

## 1. Schema Management
*   **No AutoMigrate**: Never use `gorm.AutoMigrate()`.
*   **Migration Library**: Use `github.com/golang-migrate/migrate/v4`.
*   **SQL Files**: Store migration files in `internal/infrastructure/postgres/migrations`.
*   **Naming**: `V1__init_schema.up.sql`, `V1__init_schema.down.sql`.
*   **Execution**: Run migrations on service startup in `main.go`.

## 2. Repository Pattern
*   **Separation of Concerns**:
    *   **Domain Model**: Pure Go structs in `internal/core/domain`. No `gorm` tags allowed.
    *   **DAO (Data Access Object)**: Structs mirroring DB tables in `internal/infrastructure/postgres` (or `adapter/repository`). Include `gorm` tags here.
    *   **Mappers**: Explicit conversion methods between DAO and Domain.
*   **Transaction Management**:
    *   Use a `GetTx(ctx, db)` helper.
    *   Transactions are managed by the Usecase/Service layer, injected into Context, and retrieved by the Repository.

## 3. Data Consistency
*   **Transactional Outbox**:
    *   All state changes MUST be accompanied by an Outbox Event in the **same transaction**.
    *   Use a standard `outbox_events` table.
*   **Audit Fields**: `created_at`, `updated_at` (managed by DB or DAO).

## 4. Connection Management
*   **Config**: Use `pgx` driver via GORM: `gorm.Open(postgres.Open(dsn), ...)`.
*   **Resilience**: Configure connection pooling (`SetMaxIdleConns`, `SetMaxOpenConns`).

## 5. Security & Auditing
*   **Credentials**: Never hardcode. Load from Environment Variables (Secrets).
*   **Mandatory Auditing**:
    *   All services must implement the standard `audit` schema and `logged_actions` table.
    *   All public tables (except `outbox_events` usually) must have the `audit_trigger` applied.
    *   **Application Logic**: The Repository MUST execute `SELECT set_config('app.current_user', ?, true)` at the start of every transaction to attribute changes to a user (or service ID).
