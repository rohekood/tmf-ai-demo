# Customer Management Service (TMF629)

## Core Concept: Customer vs. Party
It is crucial to distinguish between a **Party** and a **Customer**:
- **Party (TMF632)**: An individual or an organization. It represents "who" they are (name, legal ID, contact details).
- **Customer (TMF629)**: A role played by a Party in the context of buying products/services. It represents "how" they interact with the business (customer status, credit profile, segment).

*A Party can exist without being a Customer, but a Customer must be linked to a Party.*

## Use Cases

### 1. Onboard Customer
**Goal**: Register a new customer in the system.
- **Pre-condition**: The underlying Party (Individual or Organization) should ideally exist, or can be created during this flow (though separation of concerns suggests Party should typically be managed first).
- **Input**: Party Reference, Customer Segment (e.g., B2C, B2B), Initial Status.
- **Output**: A new Customer entity with a unique ID.

### 2. Retrieve Customer
**Goal**: View customer details.
- **Input**: Customer ID.
- **Output**: Full customer profile including status, credit score (if applicable), and reference to the associated Party.

### 3. Update Customer
**Goal**: Manage the customer lifecycle.
- **Scenarios**:
    - Update Customer Status (e.g., Active -> Suspended).
    - Change Customer Segment.
    - Update Credit Profile.
- **Note**: Changes to name or contact info should be done via Party Management, not here.

### 4. Find Customers
**Goal**: Search for customers based on criteria.
- **Criteria**: Name (via Party), ID, Status.

## Technical Implementation

This service is built using Go and follows a hexagonal architecture pattern.

- **Persistence**: PostgreSQL with GORM.
- **Messaging**: RabbitMQ (AMQP 0.9.1).
- **Architecture**:
    - `cmd/server/main.go`: Application entry point and dependency injection.
    - `internal/domain`: Core entities and repository interfaces.
    - `internal/infrastructure/postgres`: GORM repository implementations and SQL migrations.
    - `internal/transport/rabbitmq`: Message consumers (Listener) and producers (Publisher).

## Messaging Specification

### Exchange: `tmf.commands` (Topic)

| Function | Routing Key | Payload |
| :--- | :--- | :--- |
| Onboard Customer | `cmd.customer.onboard` | `OnboardCustomerPayload` |
| Update Customer | `cmd.customer.update` | `UpdateCustomerPayload` |
| Get Customer | `query.customer.get` | `GetCustomerPayload` |

### Exchange: `tmf.events` (Topic)

| Event | Routing Key | Description |
| :--- | :--- | :--- |
| Customer Created | `evt.customer.created` | Emitted after successful onboarding. |
| Customer Updated | `evt.customer.updated` | Emitted after any profile change. |
| State Change | `evt.customer.stateChange` | Emitted when customer status changes. |
| Party Sync | Subscribes to `evt.party.*` | Reacts to identity changes in TMF632. |

## Getting Started

1. **Prerequisites**: Docker & Docker Compose.
2. **Database**: The service automatically runs migrations on startup.
3. **Run**: `go run cmd/server/main.go` Or use Docker Compose at the root.

## TMF Alignment
This service implements **TMF629 Customer Management**.
- **Key Resources**: `Customer`
- **Related Resources**: `CustomerAccount` (potentially out of scope for initial order creation, but relevant for billing).
