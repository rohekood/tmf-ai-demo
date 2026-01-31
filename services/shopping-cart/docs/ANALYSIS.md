# Analysis: Shopping Cart Management (TMF663)

## 1. Goal & Scope
The Shopping Cart Service implements **TMF663** to manage the transient state of customer selections. It serves as a temporary container for products, services, and resources that a customer intends to purchase, acting as the staging area before the **Product Ordering** process.

**Key Characteristic**: The Cart is "Fast & Dumb". It maintains state and structure but delegates complex logic (Pricing, Eligibility, Promotions) to specialized external services via asynchronous events.

## 2. Core Entities

### 2.1 Shopping Cart
The aggregate root representing the session.
*   **Attributes**: `id`, `validFor`, `status` (Active, Abandoned, CheckedOut).
*   **References**: `relatedParty` (Customer), `billingAccount`.
*   **Pricing**: `cartTotalPrice` (Aggregated by Pricing Service).

### 2.2 Cart Item
A specific entry in the cart.
*   **Product**: `productOfferingRef` (Catalog ID), `product` (Configuration/Characteristics).
*   **Quantity**: Number of units.
*   **Pricing**: `itemPrice` (Calculated unit price).
*   **Hierarchy**: Supports parent/child items (e.g., Bundles).

### 2.3 Price Alteration (Cart Level)
Discounts or surcharges applied to the cart as a whole (e.g., "10% off total").

## 3. Key Use Cases

### 3.1 Cart Lifecycle Management
*   **Create Cart**: Explicitly create a new cart or implicitly on the first item addition.
*   **Retrieve Cart**: Fetch the current state, including the latest calculated prices.
*   **Delete/Empty Cart**: Clear contents or remove the cart entirely.
*   **Merge Carts**: (Advanced) Merge an anonymous "Guest Cart" into a "User Cart" upon login.

### 3.2 Item Management
*   **Add Item**: Add a Product Offering to the cart.
    *   *Validation*: Check if Offering exists and is active (via Catalog RPC or Cache).
*   **Update Item**: Change quantity or product characteristics (e.g., Color, Bandwidth).
*   **Remove Item**: Delete an entry.

### 3.3 Pricing Choreography (The "Ping-Pong")
The Cart does **not** calculate prices.
1.  **Trigger**: On any change (Add/Update/Remove), Cart emits `evt.cart.session.updated`.
2.  **External Process**: Pricing Service calculates costs.
3.  **Callback**: Cart receives `cmd.cart.price.update`.
4.  **Update**: Cart updates its `cartTotalPrice` and `itemPrice` fields and emits `evt.cart.session.priced`.

### 3.4 Checkout (Order Conversion)
*   **Initiate Checkout**: User signals intent to buy.
*   **Validation**: Ensure Cart is not empty and prices are fresh.
*   **Handover**: Emit `cmd.order.checkout.submit` (or similar trigger) to the **POCV Service** (Saga Orchestrator).
*   **Locking**: The Cart should be locked or marked as `CheckedOut` to prevent further modifications during the transaction.

## 4. State Machine

| State | Description | Transitions |
| :--- | :--- | :--- |
| **Active** | Open for modifications. | -> `Pricing`, `CheckedOut` |
| **Pricing** | Waiting for Pricing Service update (Optimistic Lock). | -> `Active` |
| **CheckedOut** | Successfully converted to an Order. Immutable. | Final State |
| **Abandoned** | Expired after TTL (Time To Live). | Final State |

## 5. TMF Alignment & Gaps

| Feature | TMF663 Standard | TMF Project Implementation | Strategy |
| :--- | :--- | :--- | :--- |
| **Persistence** | Persistent | Transient (Redis/Postgres) | Use Postgres for reliability + Outbox. |
| **Pricing** | Synchronous or Internal | Asynchronous External | **Event Choreography**: Cart emits events, Pricing responds with commands. |
| **Validation** | Synchronous | Asynchronous | Validation errors (e.g., "Out of Stock") are returned as async error events. |
| **Promotions** | Internal | External | Promotion Service listens to events and applies discounts via Pricing Service. |

## 6. Business Rules

### 6.1 Versioning & Optimistic Locking
To handle the async pricing loop:
1.  Cart maintains a `version` (integer) or `etag`.
2.  `evt.cart.session.updated` includes `version: N`.
3.  `cmd.cart.price.update` must include `for_version: N`.
4.  If Cart is at `version: N+1` when price arrives, the price update is **discarded** (Stale).

### 6.2 Item Uniqueness
*   Items are identified by a generated UUID, not just the Offering ID (allowing multiple entries of the same product with different configs).

### 6.3 Customer Context
*   The Cart must store the `customerId` (if known) to allow the Pricing Service to apply loyalty rules or regional pricing.
