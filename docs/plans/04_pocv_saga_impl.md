# Plan 04: POCV Saga & Order Capture

**Reference**: `docs/design/04_order_capture_saga.md`

## 1. POCV Service (`services/pocv`)
The Saga Orchestrator. **Must use Transactional Outbox**.

### 1.1 Persistence & Architecture
- [ ] **Directory Structure**:
    - `internal/core/domain`: `SagaInstance`, `events.go`.
    - `internal/adapter/repository`: `SagaRepo`, `outbox_model.go`.
    - `internal/adapter/worker`: `outbox_worker.go`.
- [ ] **Setup Postgres**: Tables `saga_instances`, `outbox_events`.
- [ ] **Implement Repository**: `SaveSaga(ctx, saga, events...)` - Saves state and outbox events atomically.

### 1.2 State Machine
- [ ] **Define States**: `PENDING`, `RESERVING`, `PAYING`, `CREATING`, `COMPLETED`, `FAILED`.
- [ ] **Implement Step Handlers**:
    - `HandleReserve()`: Return `OutboxEvent{"cmd.inv.reserve"}`.
    - `HandlePayment()`: Return `OutboxEvent{"cmd.payment.auth"}`.
- [ ] **Async Execution**: The `OutboxWorker` actually sends these commands to RabbitMQ.

### 1.3 Event Consumers
- [ ] **Listen `evt.inv.reserved`**: Transition Saga -> `PAYING`.
- [ ] **Listen `evt.inv.failed`**: Transition Saga -> `FAILED` (and notify user).
- [ ] **Listen `evt.payment.authorized`**: Transition Saga -> `CREATING`.
- [ ] **Listen `evt.payment.declined`**: Trigger Compensation (`cmd.inv.release`).

## 2. Transformation Logic (EDA)
- [ ] **Implement Mapper**: `CartToOrderMapper`.
- [ ] **Logic**: Convert `CartItem` (Redis/Cache) -> `ProductOrderItem` (TMF622).
- [ ] **Logic**: Inject `InstallationAddress` from the Qualification Session (passed via context/metadata).

## 3. Ordering Service (`services/product-ordering`)
The System of Record.

- [ ] **Scaffold Service**: Postgres DB (TMF622 Schema).
- [ ] **Implement `CreateOrderUseCase`**:
    - Listen to `cmd.order.create`.
    - Validate Payload.
    - Insert to DB.
    - Publish `evt.order.created`.

## 4. End-to-End Test
- [ ] Simulate full flow from BFF `cmd.order.submit`.
- [ ] Watch the Saga traverse all steps via RabbitMQ monitoring.
- [ ] Verify final entry in `product_order` table.
