# Plan 01: Foundation & Infrastructure Setup

**Reference**: `docs/design/01_architecture_eda.md`

## 1. Shared Libraries (`pkg/rabbitmq`)
**Goal**: Centralize the EDA logic to avoid the code duplication seen in `customer-mgmt` vs `product-catalog`.

- [ ] **Create `pkg/go.mod`**: Define module `tmf/pkg`.
- [ ] **Register in `go.work`**: Add `./pkg` to the workspace.
- [ ] **Create `pkg/rabbitmq/publisher.go`**:
    - **No CloudEvents**: Use `json.Marshal(body)` directly.
    - **Header Injection**: Must extract `user` and `Authorization` from `context` and pass to AMQP headers.
- [ ] **Create `pkg/rabbitmq/consumer.go`**:
    - **Header Extraction**: Must read headers and populate context.
    - Standardize `Handler` signature.

## 2. Infrastructure (RabbitMQ)
Define the topology as code (e.g., using a simple Go setup tool or Terraform).

- [ ] **Define Exchanges**:
    - `ex.domain.market` (Topic)
    - `ex.domain.ordering` (Topic)
    - `ex.domain.inventory` (Topic)
- [ ] **Define Queues & Bindings**:
    - `q.qual.command` binds `cmd.qual.*`
    - `q.cart.command` binds `cmd.cart.*`
    - `q.pricing.events` binds `evt.cart.updated`
    - `q.pocv.command` binds `cmd.order.*`
- [ ] **Define DLQs**: One DLQ per main queue.

## 3. BFF Service (`services/demo-ui/bff`)
The Gateway for the Frontend. ALREADY EXISTS.

- [/] **Update `go.mod`**: Use `tmf/pkg`.
- [ ] **Implement Correlation Middleware**:
    - Extract `X-Request-ID` or generate UUID.
    - Inject into Context.
- [ ] **Implement WebSocket Hub**:
    - Endpoint `WS /ws/events`.
    - Map `ConnectionID` -> `CorrelationID` / `UserID`.
    - Consumer that listens to `evt.*` (wildcard) and routes to correct WS Connection based on correlation ID.
