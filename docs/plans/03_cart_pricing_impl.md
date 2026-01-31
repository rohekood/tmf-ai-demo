# Plan 03: Cart & Pricing Implementation

**Reference**: `docs/design/03_cart_pricing_flow.md`

## 1. Cart Service (`services/shopping-cart`)
**Architecture**: Clean Architecture + Transactional Outbox.

- [ ] **Directory Structure**:
    - `internal/core/domain`: `Cart`, `CartItem`, `events.go`.
    - `internal/adapter/repository`: `GormCartRepo`, `outbox_model.go`.
    - `internal/adapter/worker`: `outbox_worker.go` (Copy logic from product-catalog).
- [ ] **Implement `AddCartItemUseCase`**:
    - Start Transaction.
    - Save Cart to Postgres (GORM).
    - Save `OutboxEvent{"evt.cart.updated"}`.
    - Commit.
- [ ] **Implement `UpdatePriceUseCase`**:
    - Input: `UpdateCartPriceCmd`.
    - Logic: Update Prices in DB.
    - Output: Save `OutboxEvent{"evt.cart.finalized"}` (if complete).

## 2. Pricing Service (`services/pricing`)
The calculator.

- [ ] **Scaffold Service**: Stateless.
- [ ] **Implement `CalculateCartUseCase`**:
    - Listen to `evt.cart.updated`.
    - Loop through items.
    - **Stub**: Lookup Price in DB (or hardcoded map for Phase 1).
    - **Logic**: Apply "Bundle Discount" (e.g. if Internet+TV, -10 EUR).
    - Publish `cmd.cart.update_price`.

## 3. Integration Test
- [ ] Send `cmd.cart.add` via BFF.
- [ ] Verify `evt.cart.updated` is emitted.
- [ ] Verify `cmd.cart.update_price` is emitted by Pricing.
- [ ] Verify Cart State in Redis has the new price.
