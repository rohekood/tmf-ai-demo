# Architecture: Shopping Cart Service (TMF663)

## Overview
The Shopping Cart Service is designed as a high-throughput, event-driven component responsible for managing the customer's intent-to-buy state. It strictly adheres to the project's **Clean Architecture** and **Transactional Outbox** patterns.

## 1. Core Principles

### 1.1 Context Propagation
**Requirement**: All methods MUST accept `context.Context` as the first parameter to ensure timeout control, cancellation propagation, and trace ID transport.

### 1.2 Asynchronous First
*   **No Synchronous Calculations**: Pricing, Inventory Checks, and Eligibility are performed asynchronously by external services.
*   **Eventual Consistency**: The Cart State displayed to the user may be "Pending Pricing" until the `evt.cart.priced` event is processed.

### 1.3 Data Consistency
*   **Transactional Outbox**: State changes (Cart DB) and Events (Outbox DB) are persisted in a single ACID transaction.
*   **Optimistic Locking**: Uses a `version` field to manage concurrent updates and the async pricing loop.

## 2. Communication Pattern

### 2.1 Ingress (Commands)
The service consumes commands via RabbitMQ (from BFF or Pricing Service).

| Topic / Subject | Payload Schema | Source | Description |
| :--- | :--- | :--- | :--- |
| `cmd.cart.session.create` | `CreateCartCmd` | BFF | Create a new session. |
| `cmd.cart.item.add` | `AddCartItemCmd` | BFF | Add an offering to the cart. |
| `cmd.cart.item.update` | `UpdateCartItemCmd` | BFF | Change quantity/config. |
| `cmd.cart.item.remove` | `RemoveCartItemCmd` | BFF | Remove an item. |
| `cmd.cart.price.update` | `UpdateCartPriceCmd` | Pricing Svc | Apply calculated prices to items/total. |
| `cmd.cart.session.checkout` | `CheckoutCartCmd` | BFF | Finalize and trigger Order Saga. |

### 2.2 Egress (Events)
The service emits events to broadcast state changes.

| Topic / Subject | Payload Schema | Trigger |
| :--- | :--- | :--- |
| `evt.cart.session.created` | `Cart` | New cart created. |
| `evt.cart.session.updated` | `Cart` | Items added/removed/modified. **Triggers Pricing**. |
| `evt.cart.session.priced` | `Cart` | Prices applied successfully. |
| `evt.cart.session.checkout_initiated` | `CartCheckoutNotification` | Handover to POCV Saga. |
| `evt.cart.session.deleted` | `CartDeleteNotification` | Cart emptied/expired. |

### 2.3 Queries (RPC)
*   **Queue**: `query.cart.session.get`
*   **Request**: `{"cartId": "uuid"}`
*   **Response**: `Cart` JSON.

## 3. Internal Architecture (Hexagonal)

```text
internal/
├── core/
│   ├── domain/             # Entities (Cart, CartItem, Price)
│   │   ├── cart.go         # Aggregate Root
│   │   └── errors.go       # Domain Errors
│   └── ports/              # Interfaces
│       ├── repository.go   # CartRepository
│       ├── publisher.go    # EventPublisher
│       └── usecases.go     # Primary Ports
│
├── usecase/                # Application Logic
│   ├── manage_items.go     # Add/Update/Remove logic
│   ├── update_price.go     # Handling Pricing callbacks
│   └── checkout.go         # State transition logic
│
└── adapter/
    ├── handler/            # RabbitMQ Consumers
    ├── repository/         # Postgres Implementation (GORM)
    └── worker/             # Outbox Worker
```

## 4. Data Model (PostgreSQL)

### 4.1 `carts` Table
| Column | Type | Notes |
| :--- | :--- | :--- |
| `id` | UUID | PK |
| `customer_id` | UUID | Nullable (Guest) |
| `status` | VARCHAR | Active, Pricing, CheckedOut |
| `version` | INT | For Optimistic Locking |
| `total_price_amount` | DECIMAL | |
| `total_price_currency` | CHAR(3) | |
| `valid_for_end` | TIMESTAMP | TTL |

### 4.2 `cart_items` Table
| Column | Type | Notes |
| :--- | :--- | :--- |
| `id` | UUID | PK |
| `cart_id` | UUID | FK -> carts |
| `offering_id` | UUID | Reference to Catalog |
| `quantity` | INT | |
| `product_config` | JSONB | Selected characteristics |
| `item_price` | JSONB | Cached price details |

### 4.3 `outbox_events` Table
Standard Outbox table for reliable messaging.

## 5. Technology Stack
*   **Language**: Go 1.24+
*   **Database**: PostgreSQL 16+ (GORM)
*   **Broker**: RabbitMQ (AMQP 0.9.1)
*   **Serialization**: JSON

## 6. Race Condition Handling
**Scenario**: User adds Item A. Cart emits `evt.cart.session.updated` (v1). User immediately adds Item B. Cart emits `evt.cart.session.updated` (v2). Pricing Service processes v1 and sends `cmd.cart.price.update` (for v1).

**Resolution**:
1.  Cart Service receives `cmd.cart.price.update` (for v1).
2.  UseCase loads Cart (Current Version = 2).
3.  Comparison: `cmd.for_version (1) != cart.version (2)`.
4.  Action: **Ignore/Drop** the command.
5.  Outcome: Pricing Service will eventually process v2 and send a new command (for v2), which will be accepted.
