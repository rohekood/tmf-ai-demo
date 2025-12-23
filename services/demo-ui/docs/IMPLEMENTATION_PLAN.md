# Implementation Plan: Demo UI for Party & Customer Management

## Goal

Implement the Demo UI based on the [ANALYSIS.md](./ANALYSIS.md) document, delivering:

1. **Party Management UI** - Full CRUD for Individuals and Organizations
2. **Customer Management UI** - Onboard, Update, Search, Delete customers
3. **RabbitMQ Debug View** - Real-time message monitoring via WebSocket

---

## Current State

### BFF (Golang)
- ✅ RPC Client implemented (`internal/transport/rabbitmq/`)
- ✅ Customer handlers: `GetCustomers`, `CreateCustomer`, `DeleteCustomer`
- ❌ Party handlers: Not implemented
- ❌ WebSocket for Debug View: Not implemented
- ❌ Redis for session externalization: Not implemented

### UI (React)
- ✅ Base setup with Vite + TypeScript
- ✅ Dependencies: TanStack Query, TanStack Table, Radix UI, Lucide
- ✅ Customer feature skeleton (`features/customers/CustomerList`)
- ❌ Party feature: Not implemented
- ❌ Routing: Not implemented (single-page)
- ❌ Debug View: Not implemented

---

## Proposed Changes

### Phase 1: BFF - Party Management Handlers

Add Party Management endpoints to the BFF.

---

#### [NEW] `bff/internal/transport/http/party_handlers.go`

New handler file for Party operations:

```go
// Routes to add:
// GET    /api/parties          -> SearchParties (query.party.search)
// GET    /api/parties/:id      -> GetParty (query.party.get)
// POST   /api/parties          -> CreateParty (cmd.party.create)
// PUT    /api/parties/:id      -> UpdateParty (cmd.party.update)
// PATCH  /api/parties/:id      -> PatchParty (cmd.party.patch)
// DELETE /api/parties/:id      -> DeleteParty (cmd.party.delete)
```

---

#### [MODIFY] `bff/internal/transport/http/handlers.go`

- Integrate party routes in `RegisterRoutes()`
- Add `GetCustomer` (single customer by ID) endpoint
- Add `UpdateCustomer` endpoint

---

### Phase 2: UI - Routing & Layout

Set up React Router and create the main layout with navigation.

---

#### [NEW] `ui/src/router.tsx`

Define all application routes:

```tsx
// Routes:
// /parties           -> PartyListPage
// /parties/new       -> PartyFormPage (create mode)
// /parties/:id       -> PartyDetailPage
// /parties/:id/edit  -> PartyFormPage (edit mode)
// /customers         -> CustomerListPage
// /customers/new     -> CustomerOnboardPage
// /customers/:id     -> CustomerDetailPage
// /customers/:id/edit-> CustomerEditPage
// /debug             -> DebugConsolePage
```

---

#### [NEW] `ui/src/components/layout/Layout.tsx`

Main layout with sidebar navigation and content area.

---

#### [NEW] `ui/src/components/layout/Sidebar.tsx`

Navigation sidebar with links to Parties, Customers, and Debug Console.

---

#### [MODIFY] `ui/src/App.tsx`

Replace current content with `<RouterProvider>` and layout wrapper.

---

#### [MODIFY] `ui/src/main.tsx`

Add `QueryClientProvider` wrapper for TanStack Query.

---

### Phase 3: UI - Party & Customer Features

Implement all pages and components for Party and Customer management.

---

#### Party Feature

##### [NEW] `ui/src/features/parties/`

New feature directory with:

| File | Purpose |
|:-----|:--------|
| `api.ts` | TanStack Query hooks for party API calls |
| `types.ts` | TypeScript interfaces (Party, Individual, Organization, ContactMedium, etc.) |
| `PartyListPage.tsx` | Data table with search, filter, actions |
| `PartyDetailPage.tsx` | Read-only view with tabs (Overview, Contacts, IDs, Related) |
| `PartyFormPage.tsx` | Create/Edit form with sub-resource sections |
| `components/PartyTable.tsx` | TanStack Table configuration |
| `components/PartyTypeSelector.tsx` | Individual/Organization toggle |
| `components/ContactMediumForm.tsx` | Repeatable contact medium inputs |
| `components/IdentificationForm.tsx` | Repeatable identification inputs |
| `components/RelatedPartyForm.tsx` | Related party selector |

---

#### Customer Feature

##### [MODIFY] `ui/src/features/customers/`

Enhance existing feature with:

| File | Purpose |
|:-----|:--------|
| `api.ts` | TanStack Query hooks (add `useCustomer`, `useUpdateCustomer`) |
| `types.ts` | TypeScript interfaces (Customer, Account, CreditProfile, Consent) |
| `CustomerListPage.tsx` | Upgrade `CustomerList` to full page with routing |
| `CustomerDetailPage.tsx` | View with Party reference, accounts, consents |
| `CustomerOnboardPage.tsx` | Party selector + customer form |
| `CustomerEditPage.tsx` | Edit status, consents, tax exemptions |
| `components/CustomerTable.tsx` | TanStack Table configuration |
| `components/PartySelector.tsx` | Search and select existing party |
| `components/ConsentManager.tsx` | Privacy consent toggle cards |
| `components/TaxExemptionForm.tsx` | Tax exemption inputs |

---

### Phase 4: RabbitMQ Debug View

Implement WebSocket-based real-time message monitoring.

---

#### BFF Changes

##### [NEW] `bff/internal/transport/http/websocket.go`

WebSocket handler for `/ws/debug`:

- Upgrade HTTP to WebSocket connection
- Broadcast messages from RabbitMQ to connected clients
- Require authenticated session

---

##### [NEW] `bff/internal/transport/rabbitmq/debug_consumer.go`

Dedicated RabbitMQ consumer:

- Subscribe to all `evt.*`, `cmd.*`, `query.*` topics
- Forward messages to WebSocket hub
- Buffer last N messages for new connections

---

##### [MODIFY] `bff/cmd/server/main.go`

- Initialize WebSocket hub
- Start debug consumer
- Register `/ws/debug` route

---

#### UI Changes

##### [NEW] `ui/src/features/debug/`

| File | Purpose |
|:-----|:--------|
| `types.ts` | DebugMessage interface |
| `useDebugWebSocket.ts` | Custom hook for WebSocket connection |
| `DebugConsolePage.tsx` | Main debug page with message feed |
| `components/MessageFeed.tsx` | Virtualized list of messages |
| `components/MessageDetail.tsx` | JSON viewer with syntax highlighting |
| `components/DebugFilters.tsx` | Service, type, topic filters |

---

## Verification Plan

### Automated Tests

```bash
# BFF Tests
cd services/demo-ui/bff
go test ./...

# UI Tests
cd services/demo-ui/ui
yarn test
yarn lint
```

### Manual Verification

1. **Party CRUD Flow**:
   - Create Individual → Verify in list → View detail → Edit → Delete

2. **Customer Onboard Flow**:
   - Search for Party → Create Customer → Verify Party link → View accounts

3. **Debug View**:
   - Open `/debug` → Trigger operations → Verify messages appear in real-time
   - Test filters → Verify correlation ID linking

### Integration Test

```bash
# Run full stack
docker-compose up -d

# Verify health endpoints
curl http://localhost:8080/health
curl http://localhost:3000
```

---

## Implementation Priority

| Phase | Effort | Dependencies |
|:------|:-------|:-------------|
| Phase 1: BFF Party Handlers | ~2 hours | None |
| Phase 2: UI Routing & Layout | ~2 hours | None |
| Phase 3: Party & Customer Features | ~6 hours | Phase 1, 2 |
| Phase 4: Debug View | ~4 hours | Phase 1 (for RabbitMQ patterns) |

---

## File Summary

| Action | Count | Files |
|:-------|:------|:------|
| NEW | 20+ | Party handlers, WebSocket, React components |
| MODIFY | 5 | handlers.go, App.tsx, main.tsx, existing features |

---

*Last Updated: 2025-12-23*
