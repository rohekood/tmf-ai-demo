# Final Implementation Plan: Event-Driven Fiber Platform

**Status**: APPROVED
**Based on**: Analysis Findings (Topic Alignment, Mock Interfaces, RPC Gaps)
**Supersedes**: Previous partial plans `01` - `04`.

This document is the **Canonical Roadmap** for the implementation. It incorporates the strict 4-part topic naming convention (`type.domain.entity.action`) and addresses the architectural gaps identified in the analysis phase.

---

## Phase 1: Foundation & Infrastructure (The Backbone)
**Goal**: Establish the shared Event-Driven library and define the interfaces for external dependencies.

### 1.1 Shared Library (`pkg/rabbitmq`)
*   [ ] **Initialize Module**: `tmf/pkg`.
*   [ ] **Publisher**: Implement `Publish(ctx, topic, payload)`.
    *   **Constraint**: Must support 4-part topic routing keys.
    *   **Context**: Auto-inject `X-Correlation-ID`, `X-User-ID` into AMQP headers.
*   [ ] **Consumer**: Implement `Consume(topic, handler)`.
    *   **Context**: Auto-extract AMQP headers into Go `context.Context`.
*   [ ] **Topic Registry**: Define constant strings for all approved topics to prevent typos.
    *   `cmd.qual.eligibility.check`
    *   `evt.cart.session.updated`
    *   etc.

### 1.2 Infrastructure Topology
*   [ ] **RabbitMQ Config**: Define Exchanges (`ex.market`, `ex.order`, `ex.inventory`) and DLQs.

### 1.3 Interface Definitions (The Contracts)
**Gap Fixed**: We need strict JSON contracts for the services we will mock (Inventory, GIS, Billing).
*   [ ] **Create `docs/interfaces/inventory.json`**: Define `query.inventory.resource.capacity` (Request/Response) and `cmd.inventory.resource.reserve`.
*   [ ] **Create `docs/interfaces/gis.json`**: Define `query.gis.geography.check`.
*   [ ] **Create `docs/interfaces/billing.json`**: Define `cmd.payment.transaction.authorize`.

---

## Phase 2: Qualification Service (The Read Path)
**Goal**: Implement the "Scatter-Gather" Logic for Address Checks using the corrected topics.

### 2.1 Core Logic
*   [ ] **Implement UseCase**: `CheckEligibility(Address)`.
    *   Use `errgroup` for parallel execution.
*   [ ] **RPC Clients**: Implement `InventoryClient` and `GISClient` using the JSON contracts defined in Phase 1.3.

### 2.2 Event Wiring
*   [ ] **Ingress**: Listen to `cmd.qual.eligibility.check`.
*   [ ] **Egress**: Publish `evt.qual.eligibility.checked`.

### 2.3 Mocks
*   [ ] **Implement `mock-inventory`**: Simple consumer returning static capacity.
*   [ ] **Implement `mock-gis`**: Simple consumer returning static polygon match.

---

## Phase 3: Shopping Cart & Pricing (The Reactive Path)
**Goal**: Implement the "Ping-Pong" choreography with **Real Catalog Data**.

### 3.1 Cart Service (`services/shopping-cart`)
*   [ ] **Persistence**: Postgres + GORM + Transactional Outbox.
*   [ ] **Topic Update**:
    *   Listen: `cmd.cart.item.add` (BFF), `cmd.cart.price.update` (Pricing).
    *   Emit: `evt.cart.session.updated`.
*   [ ] **Optimistic Locking**: Enforce `version` check on price updates.

### 3.2 Pricing Service (`services/pricing`)
**Architecture Update**: Pure EDA (No RPC). Pricing maintains a local replica of price data.
*   [ ] **Persistence**: Postgres table `price_lookup` (id, amount, currency, valid_for).
*   [ ] **Catalog Consumer**:
    *   Listen: `evt.catalog.offering.created`, `evt.catalog.offering.updated`.
    *   Action: Upsert data into `price_lookup`.
*   [ ] **Logic**:
    *   Receive `evt.cart.session.updated`.
    *   Query **Local** `price_lookup` table.
    *   Calculate Total.
    *   Emit `cmd.cart.price.update`.

---

## Phase 4: Order Capture Saga (The Write Path)
**Goal**: Implement the POCV Saga Orchestrator with the corrected Trigger Flow.

### 4.1 Trigger Flow
**Gap Fixed**: BFF triggers POCV, POCV locks Cart.
*   [ ] **BFF**: Sends `cmd.order.checkout.submit` to POCV.
*   [ ] **POCV**:
    1.  Receives Command.
    2.  Fetches Cart (RPC `query.cart.session.get`).
    3.  Validates Cart State (and locks it via internal logic or command).

### 4.2 Saga Orchestration
*   [ ] **State Machine**: Persist `SagaInstance` (Pending -> Reserving -> Paying -> Creating).
*   [ ] **Step: Inventory**: Emit `cmd.inventory.resource.reserve`.
*   [ ] **Step: Payment**: Emit `cmd.payment.transaction.authorize`.
*   [ ] **Step: Order**: Emit `cmd.order.management.create`.

### 4.3 Compensation
*   [ ] **Handle Failures**: Listen to `evt.payment.transaction.declined` -> Trigger `cmd.inventory.resource.release`.

---

## Phase 5: End-to-End Verification
*   [ ] **Scenario**: "Happy Path Fiber Order".
*   [ ] **Trace**: Follow `CorrelationID` from BFF -> Qual -> Cart -> Pricing -> POCV -> Inventory -> Billing -> Order.
