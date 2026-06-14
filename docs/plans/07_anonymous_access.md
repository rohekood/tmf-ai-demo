# Plan 07 — Anonymous Access for Check Availability

## Goal

Allow unauthenticated ("anonymous") visitors to use the **Check Availability**
(service qualification) flow only. Everything else — catalog management, parties,
customers, cart, checkout — stays behind authentication. Anonymous users see
offerings with **generic (base) prices**; segment/tier-specific pricing is only
applied for authenticated customers. The moment an anonymous user tries to add an
offering to the cart, the login flow starts.

This is a demo project, so the emphasis is a clean, correct authorization boundary
rather than a full identity/entitlement system.

## Requirements (from request)

1. If not logged in, **do not** show the login screen — show the ordering section
   with Check Availability instead.
2. Offerings are visible to anonymous users, but with **generic prices** (the
   segment-specific price cannot be computed without a customer; the generic/base
   price is always available).
3. Adding an offering to the cart requires authentication — that action **starts
   the login flow**.
4. The menu/sidebar shows a **Login** button for anonymous users.
5. Backend APIs are guarded by authorization. Anonymous users are **blocked** from
   every call except those needed for Check Availability.

## Current State (as built)

### Frontend (`services/demo-ui/ui`)

- [`components/layout/Layout.tsx`](../../services/demo-ui/ui/src/components/layout/Layout.tsx)
  gates the **entire** app: if `!isAuthenticated` it renders `<LoginPage />` and
  nothing else is reachable. This is the primary blocker for requirement 1.
- [`components/layout/Sidebar.tsx`](../../services/demo-ui/ui/src/components/layout/Sidebar.tsx)
  always renders all nav sections (Management, Product Catalog, Ordering,
  Developer) and already has a `LoginButton`/`LogoutButton` in the footer.
- [`components/auth/AuthTokenSync.tsx`](../../services/demo-ui/ui/src/components/auth/AuthTokenSync.tsx)
  sets the bearer token only when authenticated; anonymous requests carry no token.
- [`features/ordering/QualifyPage.tsx`](../../services/demo-ui/ui/src/features/ordering/QualifyPage.tsx)
  is the Check Availability page. `handleSelectOffering` adds to cart and navigates
  to `/order/cart` with no auth gate.
- Auth wiring is already in place: `AuthProvider`, `useAuth()` context exposing
  `isAuthenticated`, `loginWithRedirect`, etc.

### Backend (`services/demo-ui/bff`)

- [`cmd/server/main.go`](../../services/demo-ui/bff/cmd/server/main.go) wraps the mux
  so that **any** path starting with `/api` requires a valid JWT
  (`auth.EnsureValidToken`). There is no per-route exemption.
- Qualification endpoints registered in
  [`internal/transport/http/order_handlers.go`](../../services/demo-ui/bff/internal/transport/http/order_handlers.go):
  - `POST /api/qualification/check`
  - `GET  /api/qualification/session/{sessionId}`
- Cart / checkout endpoints (must stay protected): `POST /api/cart/items`,
  `GET /api/cart/{cartId}`, `DELETE /api/cart/{cartId}/items/{itemId}`,
  `POST /api/orders/checkout`, `GET /api/orders/saga/{sagaId}`.

### Pricing (`services/qualification`)

- [`internal/usecase/check_eligibility.go`](../../services/qualification/internal/usecase/check_eligibility.go)
  only builds `qualifiedOffers` when `cmd.CustomerID != ""`. With an empty
  CustomerID, **no offers (and no prices) are returned at all**.
- [`internal/usecase/pricing.go`](../../services/qualification/internal/usecase/pricing.go)
  `CalculatePrice` always fetches the customer (`GetCustomer`) to apply a tier
  discount — it cannot run without a customer ID. There is no "generic price" path.

## Gap Analysis → Design Decisions

| # | Requirement | Gap | Decision |
|---|---|---|---|
| 1 | Show ordering, not login, when anonymous | Layout hard-blocks on `isAuthenticated` | Replace the global block with **per-route** protection. Anonymous users land on `/order/qualify`. |
| 2 | Offerings with generic price | Qualification skips pricing when no customer | Add a **generic price path**: when `CustomerID == ""`, return offers priced at base price (no tier discount). |
| 3 | Add-to-cart requires login | No gate on add-to-cart | In `QualifyPage.handleSelectOffering`, if anonymous → `loginWithRedirect` (preserve intent: address + offering). |
| 4 | Menu shows Login when anonymous | Sidebar already conditionally shows it; nav needs trimming | Keep Login button; hide/disable protected nav items for anonymous users. |
| 5 | Backend blocks anonymous except Check Availability | All `/api` requires JWT | Convert auth to an **allowlist**: qualification endpoints are public, everything else requires JWT. |

### Security note (do not skip)

Frontend route gating is **UX only, not security**. The real authorization
boundary is the BFF allowlist (task group B). The frontend must never be the only
thing preventing a protected call. Anonymous bearer-less requests to any non-
qualification `/api` route must return `401`.

### Authenticated qualification still needs a customer ID

Today the UI never sends `customerId` on qualification, so tier pricing never
triggers in practice. Out of scope to fully fix here, but the generic-price change
(task A?) must not regress the authenticated path: when a customer ID *is* present,
tier pricing still applies. The BFF should derive `customerId` from the JWT `sub`
when the caller is authenticated (noted as a follow-up, not required for the five
requirements above).

## Task List

### Group A — Backend authorization boundary (BFF) — **highest priority**

- [x] **A1.** In [`cmd/server/main.go`](../../services/demo-ui/bff/cmd/server/main.go),
  replace the "all `/api` requires auth" wrapper with an **allowlist** of public
  routes. Public set (exact): `POST /api/qualification/check`,
  `GET /api/qualification/session/{sessionId}`. Everything else under `/api`
  requires a valid JWT. — **Done** (`main.go` 3-way switch + `auth.IsPublicRoute`).
- [x] **A2.** Make the public-route matcher **method-aware** and prefix-safe (a path
  like `/api/qualification/checkfoo` must not slip through; cart/checkout must stay
  protected). Centralize the predicate in one testable function. — **Done**
  (`IsPublicRoute` in `internal/auth/middleware.go`).
- [x] **A3.** For public routes, still attempt to parse the JWT if present so an
  authenticated user gets their `customerId` (JWT `sub`) injected into context —
  but never reject when the token is missing/invalid on a public route. — **Done**
  (`auth.OptionalToken` injects the `sub` into context; never rejects). Note: the
  qualification handler does not yet copy it into the command payload — see
  follow-ups.
- [x] **A4.** Tests: anonymous `POST /api/qualification/check` → allowed;
  anonymous `POST /api/cart/items`, `GET /api/cart/...`, `POST /api/orders/checkout`,
  catalog/party/customer routes → `401`; authenticated requests still pass.
  Cover the prefix/method edge cases from A2. — **Done** (`middleware_test.go`:
  `TestIsPublicRoute`, `TestOptionalToken_*`).

### Group B — Qualification generic pricing (qualification service)

- [x] **B1.** In [`internal/usecase/pricing.go`](../../services/qualification/internal/usecase/pricing.go),
  add a generic-price path that returns the offering **base price** without a
  customer lookup (e.g. `CalculateGenericPrice(ctx, offeringID)` or have
  `CalculatePrice` short-circuit when `customerID == ""`). — **Done**
  (`CalculateGenericPrice`).
- [x] **B2.** In [`internal/usecase/check_eligibility.go`](../../services/qualification/internal/usecase/check_eligibility.go),
  build `qualifiedOffers` for **both** anonymous (generic price) and authenticated
  (tier price) cases. Remove the `cmd.CustomerID != ""` guard around offer
  construction; branch on it only for which pricing call to use. — **Done**
  (now guarded only by `isQualified`; branches on `CustomerID == ""`).
- [x] **B3.** (Optional, for clarity) Mark generic-priced offers so the UI can label
  them, e.g. add `priceType: "generic" | "customer"` to the offer payload. — **Done**
  (`domain.PriceType*` constants + `QualifiedOffer.PriceType`; passes through the BFF).
- [x] **B4.** Tests: anonymous qualification returns offers at base price;
  authenticated returns tier-discounted price; verify ≥90% coverage on changed files.
  — **Done** (`pricing_test.go` generic cases + `check_eligibility_test.go`
  anonymous/authenticated offer assertions; full suite green). Coverage % not
  formally measured this round.

### Group C — Frontend routing & access control

- [x] **C1.** In [`router.tsx`](../../services/demo-ui/ui/src/router.tsx), introduce a
  `<RequireAuth>` wrapper (or route loader) around protected route groups
  (parties, customers, catalog, cart, checkout, order status/confirmation, debug).
  Leave `/order/qualify` public. — **Done** (pathless `RequireAuth` layout route;
  `order/qualify` lifted out as public).
- [x] **C2.** Change the index redirect: anonymous → `/order/qualify`; authenticated
  → existing default (`/parties`). — **Done** (`IndexRedirect`).
- [x] **C3.** In [`Layout.tsx`](../../services/demo-ui/ui/src/components/layout/Layout.tsx),
  remove the global `!isAuthenticated → <LoginPage />` block. Keep the `isLoading`
  spinner. The Layout now always renders the shell; route guards handle protection.
  — **Done** (`LoginPage` import removed; only `isLoading` spinner remains).
- [x] **C4.** `<RequireAuth>` behaviour for anonymous hitting a protected route:
  trigger `loginWithRedirect` (preferred) or redirect to `/order/qualify`. Decide
  one; document it. Preserve return-to so post-login lands on the intended page.
  — **Done** (anonymous → `loginWithRedirect` with `appState.returnTo`; when Auth0
  is unavailable, falls back to `<Navigate to="/order/qualify">`).

### Group D — Frontend ordering UX

- [x] **D1.** In [`QualifyPage.tsx`](../../services/demo-ui/ui/src/features/ordering/QualifyPage.tsx)
  `handleSelectOffering`: if `!isAuthenticated`, call `loginWithRedirect` instead of
  adding to cart. Preserve intent (address + selected offering) so the user can
  resume after login — via `appState`/`returnTo` or `sessionStorage`. — **Done**
  (address stashed in `sessionStorage` under `QUALIFY_RESUME_KEY`; `loginWithRedirect`
  with `returnTo`).
- [~] **D2.** (If D1 stores intent) On return from login, resume the add-to-cart with
  the stored offering, then navigate to `/order/cart`. — **Deviated (by design).**
  On return we restore the address and **re-run qualification** (so offers are
  re-priced for the now-known customer) rather than silently re-adding the
  anonymous-priced offering. The user confirms add-to-cart against the customer
  price. Auto-adding the stored offering was intentionally not implemented.
- [x] **D3.** Surface generic vs customer pricing in
  [`OfferingCard.tsx`](../../services/demo-ui/ui/src/features/ordering/OfferingCard.tsx)
  if B3 is implemented (e.g. "from {price}" / "Sign in for your price"). — **Done**
  (`priceType === 'generic'` shows "Standard price — sign in for your pricing").

### Group E — Frontend menu/navigation

- [x] **E1.** In [`Sidebar.tsx`](../../services/demo-ui/ui/src/components/layout/Sidebar.tsx),
  hide (or visibly lock) the protected nav sections — Management, Product Catalog,
  Developer, and the Shopping Cart item — when anonymous. Keep the **Ordering →
  Check Availability** item visible. — **Done** (sections gated behind
  `isAuthenticated`; Check Availability always shown).
- [x] **E2.** Ensure the footer shows the **Login** button prominently for anonymous
  users (already present) and the user card/Logout for authenticated users. — **Done**
  (unchanged existing footer behaviour; verified it renders for anonymous).

### Group F — Verification

- [x] **F1.** Backend: `go test ./...`, `golangci-lint run`, `go vet ./...` for the
  BFF and qualification services. Report results. — **Done, with caveat.**
  `go test ./...` green for both services; `go vet` clean. `golangci-lint` could
  **not** run — pre-existing env mismatch (binary built with go1.25, project
  targets go1.26.2), unrelated to these changes.
- [x] **F2.** Frontend: `yarn test` and `yarn lint` from `services/demo-ui/ui`.
  — **Done.** `yarn test` → 302 passed; `yarn lint` → no errors in any changed file
  (22 pre-existing errors remain in the Party pages, untouched here).
- [ ] **F3.** Manual/Playwright walk-through: anonymous can qualify and see priced
  offers; add-to-cart bounces to login; after login the cart works; direct
  navigation to a protected route while anonymous is blocked; a bare
  `curl` to a protected `/api` route with no token returns `401`. — **Not done**
  (no live environment exercised this round).

## Suggested Order of Work

1. **Group A** (backend allowlist) — establishes the real security boundary first.
2. **Group B** (generic pricing) — makes anonymous qualification actually useful.
3. **Group C** + **E** (routing + menu) — unblocks the anonymous UI shell.
4. **Group D** (add-to-cart → login) — completes the conversion flow.
5. **Group F** (verification) throughout and at the end.

## Out of Scope / Follow-ups

- Wiring `customerId` into tier pricing on qualification. The middleware now
  injects the JWT `sub` into context (A3), but the qualification handler still does
  not copy it into the command payload, and `sub` (an Auth0 user ID) is not a
  catalog `customerId` — a `sub → customerId` lookup is still missing. So
  authenticated users currently also receive generic pricing.
- **Orphaned `LoginPage`**: `features/auth/LoginPage.tsx` is no longer referenced
  (login now goes straight to the Auth0 redirect). Safe to delete; left in place.
- Rate limiting / abuse protection on the now-public qualification endpoint.
- Persisting anonymous qualification sessions to a real authenticated customer on
  login (session hand-off).
