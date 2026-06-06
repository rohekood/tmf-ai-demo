# Plan 06: Party Soft-Delete UX & Catalog Loading Fixes

## Summary

Six bug clusters identified across `party-management`, `customer-management`, and `product-catalog-management` services plus the `demo-ui`. This document captures root causes, affected files, and implementation steps for each.

---

## Issue 1 — Party list shows deleted parties by default

### Problem

`PartyListPage.tsx` calls `useParties()` with no status filter. Parties with `status: 'Deleted'` appear in the list alongside active parties, with no opt-in to see them.

### Root cause

- Frontend: `fetchParties()` in [api.ts:8-19](../services/demo-ui/ui/src/features/parties/api.ts#L8-L19) sends no `status` parameter.
- Backend: `SearchParties` / `ListParties` handler in party-management does not filter by status unless the caller asks.
- Types: `SearchPartyParams` in [types.ts](../services/demo-ui/ui/src/features/parties/types.ts) has no `status` field.

### Implementation plan

**Frontend**
1. Add a `showDeleted` boolean checkbox to `PartyListPage.tsx` (default `false`).
2. Pass `status` into `SearchPartyParams` and `fetchParties`:
   - When `showDeleted = false`: pass `status=Active,Initialized,DeletionPending` (exclude `Deleted`).
   - When `showDeleted = true`: pass nothing (return all).
3. Visually distinguish deleted rows (e.g. dimmed, strikethrough on the name).

**Backend**
1. In the party-management list/search handler, add a `status` filter to the repository query. When the `status` parameter is provided, apply `WHERE status IN (...)`. When absent, return all.
2. Use an explicit `switch`-case column mapping (per ARCHITECTURE standard) — never string interpolation.

---

## Issue 2 — Deleted party cannot be permanently deleted

### Problem

Soft-delete sets `party.status = 'Deleted'` but the record stays in the database forever. There is no UI button or backend endpoint to permanently purge a deleted party.

### Root cause

- `DeleteParty()` in [party_repository.go:346-371](../services/party-management/internal/infrastructure/postgres/party_repository.go#L346-L371) performs a hard delete, but it is never called by any handler. The deletion saga only transitions the status field.
- No BFF route exists for permanent deletion.
- No UI control for it.

### Implementation plan

**Backend**
1. Add a new command: `cmd.party.purge` in [handlers.go](../services/party-management/internal/transport/rabbitmq/handlers.go).
2. Handler `HandlePurgeParty`: 
   - Fetch party, reject if `status != 'Deleted'` (guard: cannot purge active parties).
   - Call `repo.DeleteParty()` which hard-deletes individual/organization + parties rows.
   - Publish `evt.party.purged` event.
3. Add BFF route: `DELETE /api/parties/{id}/purge` → `cmd.party.purge`.

**Frontend**
1. In `PartyListPage.tsx` and the party detail/view page, add a "Permanently Delete" button that is **only visible when `party.status === 'Deleted'`** (and only when "show deleted" is active).
2. Confirm dialog: _"This will permanently delete the party and cannot be undone."_
3. On success, invalidate parties query and show toast.

---

## Issue 3 — Customer can be onboarded with a deleted (or non-existent) party

### Problem

`HandleOnboardCustomer` in [customer-management handlers](../services/customer-management/internal/transport/rabbitmq/handlers.go) creates a customer with any `partyId` without verifying that the party exists and is in a valid state.

There are two sub-cases:
- **Pre-save race**: user selects a party that is already `Deleted` and clicks Save.
- **Mid-save race**: party is `Active` when the form loads, but a concurrent deletion completes before the Save message is processed.

The existing deletion saga has compensation logic (if customer creation races with `DeletionPending`, deletion is cancelled). But it has **no guard when the party is already `Deleted`** at onboarding time.

### Root cause

- No party status check in `HandleOnboardCustomer`.
- `PartySelector` / `PartyPicker` frontend components fetch all parties without filtering out `Deleted` ones, so a deleted party can be selected in the customer form.

### Implementation plan

**Backend (customer-management)**
1. After extracting `partyId` from the onboard payload, send an RPC call (`query.party.get`) to party-management to fetch the party.
2. If party not found → return `ErrPartyNotFound` (400).
3. If `party.status == 'Deleted'` → return `ErrPartyDeleted` (400).
4. If `party.status == 'DeletionPending'` → return `ErrPartyPendingDeletion` (409).
5. Only proceed if `party.status == 'Active'`.

**Frontend**
1. In `PartySelector.tsx` and `PartyPicker.tsx`, pass `status=Active` (or similar) so deleted/pending parties do not appear as selectable options in the customer onboarding form.
2. The BFF should map backend error codes to appropriate HTTP status codes (400/409) with descriptive messages, and the customer form should display them to the user.

---

## Issue 4 — Party can be soft-deleted even when it has connections

### Problem

When `HandleDeleteParty` initiates deletion, it publishes `EvtPartyDeletionInitiated` and waits for customer-management to respond. If the party has active customers, customer-management sends back a cancel signal and the deletion is aborted (good). However:

- There is **no check for other connections** before initiating the saga (e.g. orders linked to a customer linked to this party).
- The check only happens **asynchronously** via the saga, meaning users click Delete and see a spinner before being told it failed.

### Root cause

- Deletion saga is reactive (scatter-gather): it starts optimistically and relies on downstream services to veto.
- No synchronous pre-check in `HandleDeleteParty`.

### Implementation plan

**Backend (party-management)**
1. Before transitioning to `DeletionPending`, perform a synchronous RPC call to customer-management: `query.customer.search` with `partyId = <id>`.
2. If any active customers are returned, immediately return an error (do not initiate the saga).
3. Return a structured error: `{"error": "party_has_customers", "count": N}` with HTTP 409 from the BFF.

**Frontend**
1. Handle the 409 error in `handleDelete` in `PartyListPage.tsx` with a user-friendly toast: _"Cannot delete: party has N linked customer(s)."_
2. Remove the 10-second async polling loop (`checkDeletionStatus`) and replace with a simpler synchronous delete flow, since the backend now fails fast.

> **Note on architecture**: this changes the deletion from a full saga to a synchronous check + async finalization. The saga-based approach should still be kept for the finalization step (status `DeletionPending → Deleted`), but the initial guard should be synchronous to give immediate feedback.

---

## Issue 5 — Catalog UI shows "Loading catalogs..." indefinitely

### Root cause (confirmed from code)

Two defects in the catalog service's `RabbitMQHandler` (`rabbitmq_handler.go` and `rabbitmq_handler_retrieval.go`):

**Defect A — No error reply on list failure**

`handleListCatalogs` (line 290–322 of `rabbitmq_handler.go`) and `handleListCategories` (line 36–47 of `rabbitmq_handler_retrieval.go`) both `return` early when the use case returns an error **without sending any reply to the BFF**.

```go
if err != nil {
    log.Printf("Error executing ListCatalogs: %v", err)
    return  // ← BFF waits 10 s, then times out
}
```

The BFF's `ListCatalogs` handler has a `catalogRPCTimeout = 10 * time.Second` context. When no reply arrives, the context expires and the BFF returns HTTP 500. React Query retries 3 times (default) with exponential back-off, producing ~40–50 seconds of `isLoading: true` before the error state is finally rendered. Users perceive this as "infinite loading."

**Defect B — Shared AMQP channel used from two goroutines**

`NewRabbitMQHandler` creates one `h.channel`. `setupConsumers()` starts a goroutine that calls `h.handleMessage(d)` → `h.reply(...)` → `h.channel.PublishWithContext(...)`. `setupQueryConsumers()` starts a second goroutine that calls `h.handleQuery(d)` → `h.publishResponse(...)` → `h.channel.PublishWithContext(...)`.

The `amqp091-go` library serialises `PublishWithContext` calls with an internal mutex, so this does not cause data corruption. However, if one goroutine holds the mutex during a slow publish (or a network stall), the other goroutine blocks, delaying replies. Under load, this can compound with Defect A to increase the perceived loading time.

### Implementation plan

**Backend — catalog service**

1. **Fix all list/get handlers to always send a reply**, even on error. Pattern to apply to every handler in `rabbitmq_handler.go` and `rabbitmq_handler_retrieval.go`:

   ```go
   results, err := h.listCatalogsUC.Execute(ctx, ports.ListCatalogsInput{})
   if err != nil {
       log.Printf("Error: %v", err)
       h.publishResponse(ctx, d, map[string]string{"error": err.Error()})
       return
   }
   h.publishResponse(ctx, d, results)
   ```

2. **Open a dedicated reply channel** for publishing responses. In `NewRabbitMQHandler`, open a second `*amqp.Channel` (`h.replyChannel`) used exclusively for publishing replies. Protect it with a `sync.Mutex`. This eliminates the shared-channel goroutine contention.

### Issue 5.1 — Cannot create a new catalog

**Root cause**

Two compounding defects:

1. **Frontend error handling**: `CatalogEditForm.tsx` (line 56–58):
   ```tsx
   } catch (err) {
       console.error('Failed to save catalog:', err);  // ← silent, no toast
   }
   ```
   When creation fails, the user sees the spinner stop but gets no error message.

2. **BFF always returns 201**: `handleCommand` in `catalog_handlers.go` writes `http.StatusCreated` even when the backend catalog service returned `{"error": "some message"}`. The frontend calls `navigate(`/catalog/catalogs/${result.id}`)` where `result.id` is `undefined`, navigating to `/catalog/catalogs/undefined`.

3. **Lifecycle status dropped at creation**: `CatalogCreateEvent` struct (in `events.go`) has no `LifecycleStatus` field. The user's choice is silently ignored; the use case always hardcodes `"Active"`.

**Implementation plan**

**Backend — catalog service**
1. Add `LifecycleStatus string` to `CatalogCreateEvent` (same as `ProductSpecificationCreateEvent` already does).
2. Pass `LifecycleStatus` through `CreateCatalogInput` → `CreateCatalog.Execute()` instead of hardcoding `"Active"`.

**Backend — BFF**
1. In `handleCommand`, check if the response body is a JSON object containing an `"error"` key. If so, return HTTP 422/500 with the error body rather than 201.

   ```go
   // After receiving responseBytes:
   var check map[string]any
   if json.Unmarshal(responseBytes, &check) == nil {
       if errMsg, ok := check["error"].(string); ok {
           http.Error(w, errMsg, http.StatusUnprocessableEntity)
           return
       }
   }
   w.WriteHeader(http.StatusCreated)
   _, _ = w.Write(responseBytes)
   ```

**Frontend**
1. In `CatalogEditForm.tsx`, add a `useState` error message and display it in the form on catch:
   ```tsx
   } catch (err) {
       setFormError('Failed to save catalog. Please try again.');
   }
   ```

---

## Issue 6 — Category list shows "Loading categories..." indefinitely

Identical root causes as Issue 5 but for categories. The `handleListCategories` handler also returns without a reply on error.

### Implementation plan

Apply the same fixes as Issue 5 (Defects A and B) to the category handlers:
- `handleListCategories` in `rabbitmq_handler_retrieval.go:36–47`
- `handleGetCategory` in `rabbitmq_handler_retrieval.go:49–68`

### Issue 6.1 — Cannot create a new category

**Root cause**

1. `CategoryEditPage.tsx` (line 174–176): same silent error pattern as 5.1.
2. `CategoryCreateEvent` in `events.go` has no `LifecycleStatus` field — user's status selection is dropped.
3. The `CategoryForm` component sends `lifecycleStatus` in the payload, but it is stripped at the event struct boundary.
4. Same BFF 201-on-error problem as Issue 5.1.

**Implementation plan**

**Backend — catalog service**
1. Add `LifecycleStatus string` to `CategoryCreateEvent`.
2. Pass `LifecycleStatus` through `CreateCategoryInput` → `CreateCategory.Execute()`.

**Backend — BFF**
1. Same `handleCommand` fix as Issue 5.1 (already covers categories since they share the same helper).

**Frontend**
1. `CategoryEditPage.tsx`: add user-facing error state in the catch block.
2. On success, navigate to `/catalog/categories` (already done correctly at line 173).

---

## Affected Files Summary

| File | Issues |
|---|---|
| `services/demo-ui/ui/src/features/parties/PartyListPage.tsx` | 1, 2, 4 |
| `services/demo-ui/ui/src/features/parties/api.ts` | 1, 2 |
| `services/demo-ui/ui/src/features/parties/types.ts` | 1 |
| `services/demo-ui/ui/src/features/parties/PartySelector.tsx` | 3 |
| `services/demo-ui/ui/src/features/parties/PartyPicker.tsx` | 3 |
| `services/demo-ui/ui/src/features/catalog/CatalogEditForm.tsx` | 5.1 |
| `services/demo-ui/ui/src/features/catalog/CategoryEditPage.tsx` | 6.1 |
| `services/demo-ui/bff/internal/transport/http/catalog_handlers.go` | 5.1, 6.1 |
| `services/party-management/internal/transport/rabbitmq/handlers.go` | 2, 3, 4 |
| `services/party-management/internal/infrastructure/postgres/party_repository.go` | 2 |
| `services/customer-management/internal/transport/rabbitmq/handlers.go` | 3 |
| `services/product-catalog-management/internal/adapter/handler/rabbitmq_handler.go` | 5, 5.1 |
| `services/product-catalog-management/internal/adapter/handler/rabbitmq_handler_retrieval.go` | 5, 6 |
| `services/product-catalog-management/internal/core/domain/events.go` | 5.1, 6.1 |
| `services/product-catalog-management/internal/usecase/catalog/create_catalog.go` | 5.1 |
| `services/product-catalog-management/internal/usecase/category/create_category.go` | 6.1 |

---

## Implementation Order (recommended)

1. **Issue 5 + 6** (catalog/category loading) — highest UX impact, unblocks any testing of catalog features.
2. **Issue 5.1 + 6.1** (catalog/category create) — depends on 5/6 being fixed so the list refreshes after creation.
3. **Issue 1** (show deleted parties) — prerequisite for Issue 2 UI.
4. **Issue 2** (permanent delete) — depends on Issue 1 UI.
5. **Issue 4** (deletion pre-check) — depends on Issue 1 being done.
6. **Issue 3** (customer onboarding validation) — independent but important for data integrity.
