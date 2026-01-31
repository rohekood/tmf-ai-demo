# Design 01: Event-Driven Architecture & Topology

## 1. Overview
This document defines the **Global Event Topology** and **Communication Patterns** for the Fiber Internet platform. Adhering to the project's strict EDA guidelines, all backend services communicate exclusively via RabbitMQ. synchronous HTTP calls are forbidden between microservices.

## 2. Global Topic Topology
We use a structured topic hierarchy to ensure routing flexibility and observability.

### 2.1 Topic Structure
Pattern: `<type>.<domain>.<entity>.<action>`

*   **Type**:
    *   `cmd`: **Command**. A request to do something. Targeted at a specific service. (One-to-One).
    *   `evt`: **Event**. A fact that happened. Broadcast to anyone interested. (One-to-Many).
    *   `query`: **Query**. A request for data. (Async Request-Reply).
*   **Domain**: e.g., `catalog`, `party`, `ordering`, `inventory`.
*   **Entity**: e.g., `productOffering`, `customer`, `productOrder`.
*   **Action/State**: e.g., `create`, `created`, `updated`.

### 2.2 Domain Separation & Topic List

#### Market Domain
| Topic | Type | Payload | Producer | Consumer |
| :--- | :--- | :--- | :--- | :--- |
| `cmd.qual.eligibility.check` | Command | `CheckEligibilityCmd` | BFF | Qualification Svc |
| `evt.qual.eligibility.checked` | Event | `EligibilityCheckedEvt` | Qualification Svc | BFF (via socket) |
| `cmd.cart.item.add` | Command | `AddItemCmd` | BFF | Cart Svc |
| `evt.cart.session.updated` | Event | `CartUpdatedEvt` | Cart Svc | Pricing Svc |
| `cmd.cart.price.update` | Command | `UpdateCartPriceCmd` | Pricing Svc | Cart Svc |

#### Ordering Domain (POCV)
| Topic | Type | Payload | Producer | Consumer |
| :--- | :--- | :--- | :--- | :--- |
| `cmd.order.checkout.submit` | Command | `SubmitOrderCmd` | BFF | POCV Svc |
| `evt.order.management.created` | Event | `OrderCreatedEvt` | POCV Svc | Inventory, Billing |
| `evt.order.checkout.failed` | Event | `OrderFailedEvt` | POCV Svc | BFF, Ops |

#### Inventory Domain
| Topic | Type | Payload | Producer | Consumer |
| :--- | :--- | :--- | :--- | :--- |
| `query.inventory.resource.capacity` | Query | `PortCountQuery` | Qual Svc | Inv Svc |
| `cmd.inventory.resource.reserve` | Command | `ReserveResourceCmd` | POCV Svc | Inv Svc |
| `evt.inventory.resource.reserved` | Event | `ResourceReservedEvt` | Inv Svc | POCV Svc |

---

## 3. The Sync-to-Async Bridge (BFF Pattern)
Since the Frontend (React) speaks HTTP/REST and the Backend speaks RabbitMQ, the **BFF (Backend-For-Frontend)** acts as the protocol bridge.

### 3.1 Command Correlation Pattern
How the Frontend knows a Command is done.

1.  **Frontend**: Generates a UUID `correlation_id` (e.g., `req-123`).
2.  **Frontend**: Sends HTTP POST to BFF with header `X-Request-ID: req-123`.
3.  **BFF**: Publishes `cmd.qual.check` to RabbitMQ with `metadata.correlation_id = req-123`.
4.  **BFF**: Subscribes to `evt.qual.checked` (Ephemeral Reply Queue or filtered subscription).
5.  **BFF**: Holds the HTTP request open (Long Polling) OR upgrades to WebSocket.

### 3.2 WebSocket Strategy (Recommended)
For complex flows like "Checkout", a pure WebSocket connection is preferred to stream updates.

**Flow**:
1.  Client connects `ws://api/events`.
2.  Client sends `cmd.order.submit` (over HTTP or WS).
3.  Backend (POCV) emits multiple events: `evt.order.validating`, `evt.order.pricing`, `evt.order.created`.
4.  BFF streams these events to the Client via WS.
5.  Client UI updates the progress bar in real-time.

---

## 4. Message Envelopes (Raw JSON + Headers)
Validation against existing services (`product-catalog`, `customer-mgmt`) confirms **Raw JSON** payloads are used. We DO NOT use CloudEvents envelopes in the body. Context is passed via AMQP Headers.

### 4.1 Payload Structure
```json
// The Body is the pure Domain Logic
{
  "id": "PO_1",
  "status": "Active"
}
```

### 4.2 AMQP Headers (Context Propagation)
Middleware MUST inject these headers:
*   `X-Correlation-ID`: `req-123` (Traceability)
*   `user`: `user-uuid` (Identity)
*   `Authorization`: `Bearer eyJ...` (Security context for downstream)

## 5. Error Handling in EDA
### 5.1 Dead Letter Queues (DLQ)
*   Every consumer queue MUST have a corresponding DLQ (e.g., `q.cart.DLQ`).
*   Config: `x-dead-letter-exchange` on RabbitMQ.

### 5.2 Retry Policy
*   **Transient Errors** (DB timeout): Retry with Exponential Backoff (3 times).
*   **Permanent Errors** (Invalid JSON): Reject immediately (Sent to DLQ).

---
