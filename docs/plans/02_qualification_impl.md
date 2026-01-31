# Plan 02: Qualification Service Implementation

**Reference**: `docs/design/02_qualification_flow.md`

## 1. Service Scaffolding (`services/qualification`)
**Architecture**: Clean Architecture (Standardized).

- [ ] **Directory Structure**:
    - `cmd/main.go`
    - `internal/core/domain` (Structs)
    - `internal/core/ports` (Interfaces)
    - `internal/usecase` (Logic)
    - `internal/adapter/handler` (RabbitMQ Consumer)
    - `internal/adapter/repository` (No DB for Qual, but maybe cache)
    - `internal/adapter/publisher` (RabbitMQ Publisher)
- [ ] **Setup Configuration**: Connect to RabbitMQ using `pkg/rabbitmq`.

## 2. Infrastructure Adapters (The Clients)
Since Qual needs to call Inventory and GIS, we need async RPC clients.

- [ ] **Create `internal/adapter/rpc/inventory_client.go`**:
    - Method `GetPortCapacity(ctx, locationID)`.
    - Usage: `publisher.PublishRaw` to `query.inv.*`.
- [ ] **Create `GISClient`**:
    - Method `CheckPolygon(ctx, address)`.

## 3. Core Domain Logic
- [ ] **Define Entities**: `QualificationRequest`, `QualificationResult`.
- [ ] **Implement `CheckEligibilityUseCase`**:
    - Logic: `errgroup` to call Inventory and GIS in parallel.
    - Logic: Combine results (AND logic).
    - Logic: Filter Categories (if `fiber_available` -> allow `CAT_FIBER`).

## 4. TMF679 Interface (EDA)
- [ ] **Implement Consumer**: Listen to `cmd.qual.check`.
- [ ] **Implement Publisher**: Publish `evt.qual.checked`.
- [ ] **Map Payloads**: Convert Internal Domain Model -> TMF679 JSON Structure.

## 5. Mocks (For Testing)
To verify this phase without building the full backend.
- [ ] **Create `mock-inventory`**: Listens to `query.inv.*` and returns random capacity.
- [ ] **Create `mock-gis`**: Listens to `query.gis.*` and returns "In Zone".
