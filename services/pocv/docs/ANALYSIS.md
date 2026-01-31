# Analysis: Product Order Capture & Validation (POCV)

## 1. Goal & Scope
The **POCV Service** (Product Order Capture & Validation) serves as the **Saga Orchestrator** for the Checkout process. Its primary goal is to reliably transform a customer's **Commercial Intent** (Shopping Cart) into a **Fulfilled Order** by coordinating distributed transactions across Inventory, Billing, and Order Management systems.

**Key Characteristic**: The POCV Service is an **Orchestrator**. It owns the *process state* (The Saga) but not the *domain entities* (Stock, Payments, Orders). It implements the **Saga Pattern** to manage long-running distributed transactions without 2-Phase Commit (2PC).

## 2. Context & Role
In the context of the Shopping Cart flow:
1.  **Shopping Cart (TMF663)**: Collects and prices the items. It is the "Draft" phase.
2.  **POCV (Saga)**: The "Commit" phase. It takes the finalized Cart and attempts to make it real.
3.  **Product Ordering (TMF622)**: The "System of Record". It stores the final legal contract once validation is complete.

## 3. Core Entities

### 3.1 Saga Instance
Represents a single checkout attempt.
*   **Attributes**: `sagaId` (Correlation ID), `cartId`, `customerId`, `currentState`, `payload` (Snapshot of Cart), `createdAt`, `updatedAt`.
*   **Lifecycle**: Ephemeral but durable (persisted until completion).

### 3.2 Saga History
An audit log of every step taken within a Saga.
*   **Attributes**: `sagaId`, `step`, `status` (Started/Completed/Failed), `payload` (Command/Event data), `timestamp`.

## 4. Key Use Cases

### 4.1 Submit Checkout (Start Saga)
*   **Trigger**: User clicks "Buy Now" (BFF sends `cmd.order.submit`).
*   **Input**: `cartId` (Reference to the source of truth).
*   **Actions**:
    1.  **Fetch Cart**: Retrieve the frozen state of the Shopping Cart (RPC or Event Payload).
    2.  **Initialize Saga**: Create a `SagaInstance` in `PENDING` state.
    3.  **Validate**: Ensure Cart is `CheckedOut` (or lock it).
    4.  **Start Flow**: Transition to first step (`RESERVING_STOCK`).

### 4.2 Orchestrate Fulfillment (The Happy Path)
The service drives the workflow forward by listening to events from domain services:
1.  **Reserve Inventory**: Send `cmd.inv.reserve`. Wait for `evt.inv.reserved`.
2.  **Authorize Payment**: Send `cmd.payment.auth`. Wait for `evt.payment.authorized`.
3.  **Create Order**: Send `cmd.order.create` (to TMF622). Wait for `evt.order.created`.
4.  **Complete**: Mark Saga as `COMPLETED`. Notify User.

### 4.3 Handle Failures (Compensation / The Unhappy Path)
If any step fails, the service MUST undo previous successful steps (backward recovery).
*   **Scenario**: Inventory Reserved ✅ -> Payment Declined ❌.
*   **Action**: Trigger Compensation for Inventory.
    *   Send `cmd.inv.release`.
    *   Mark Saga as `FAILED`.
    *   Notify User: "Payment failed, stock released."

### 4.4 Timeout Management
*   **Problem**: What if Inventory never replies (Service Down / Message Lost)?
*   **Solution**: A background monitor checks for Sagas stuck in `PENDING` state for > N minutes.
*   **Action**: Mark as `TIMED_OUT` and trigger compensation for any completed steps.

## 5. Interaction with Shopping Cart
*   **Input Source**: The Cart is the *Source of Truth* for the order content.
*   **Snapshotting**: POCV must create a **Deep Copy** of the Cart at the start of the Saga. Changes to the Cart *after* checkout starts must NOT affect the ongoing Saga.
*   **Idempotency**: If the user clicks "Buy Now" twice, POCV must detect the duplicate `cartId` and return the existing Saga status instead of spawning a double order.

## 6. TMF Alignment
*   **TMF622 (Product Ordering)**: While TMF622 handles the *storage* of the order, POCV implements the *Process Flow* (C002 Order Capture).
*   **Separation of Concerns**: We separate POCV (Process) from ProductOrdering (Data) to keep the System of Record clean and highly available. The Order is only created in TMF622 when it is *guaranteed* to be valid (Stock + Payment secured).
