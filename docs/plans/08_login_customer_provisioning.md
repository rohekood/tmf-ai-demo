# Plan 08 — Auto-Provision Customer (and Party) on Login

## Goal

When a user logs in, they must be backed by a **Customer** entity. If the logged-in
user is not yet a customer, one is created. The link between the logged-in user and
the customer is the **email address**. Because a Customer is a Party playing the
customer role, a **Party** must also exist — if no party exists for that email, it
is created too. When extra information beyond what the login token provides is
required, a dedicated profile form is shown to the logged-in user; on save, the
Party and Customer entities are persisted.

This builds directly on [Plan 07 — Anonymous Access](07_anonymous_access.md): the
provisioning endpoints are authenticated-only and therefore already protected by
the allowlist boundary introduced there.

## Requirements (from request)

1. On login, the user should be created as a **Customer** if not already one.
2. The reference between Customer and the logged-in user is the **email**.
3. If no Customer exists for the email, create one. If no Party exists for the
   email, create the Party too.
4. When additional customer info is required, show a **separate form** to the
   logged-in user where they fill in the fields.
5. After save, the **Party and Customer** entities are persisted to the database.

## Current State (as built)

### Entities & linkage

- **Party** ([party.go](../../services/party-management/internal/domain/party.go)) is
  the base entity (`parties` table); `Individual`/`Organization` subtypes. Email is
  stored as a `ContactMedium` (`party_contact_mediums`, `mediumType = "email"`).
- **Customer** ([customer.go](../../services/customer-management/internal/domain/customer.go))
  references the party via `PartyID` (`gorm:"not null"`). A customer cannot exist
  without a party. Email may also appear as a customer `ContactMedium`, but the
  authoritative person-level email lives on the Party.

### Creation commands (async via RabbitMQ)

- Party: `cmd.party.create` (exchange `tmf.party`). Payload `CreatePartyPayload`
  with `@type` + `Individual{givenName, familyName, contactMediums[…]}`. Emits
  `evt.party.created`.
- Customer: `cmd.customer.onboard` (exchange `tmf.customer`). Payload
  `OnboardCustomerPayload` requires **`partyId`** and validates the party exists
  via `PartyChecker` (RPC to party service). It does **not** create the party.
  Emits `evt.customer.created`.

### Lookup capabilities

- Customer search (`query.customer.search`) supports: `id`, `name`, `status`,
  **`party_id`**, and a generic `search` (name/status/party_id/party_type). **No
  email criterion.**
- Party search (`query.party.search`) supports: `id`, `type`, `status`, generic
  `search` (names), `externalReference`, `given_name`, `family_name`,
  `trading_name`. **No email criterion** — email is not joined from
  `party_contact_mediums`.

### Identity available at the BFF

- The BFF auth middleware injects only the JWT **`sub`** (user id) into context;
  `getHeaders` forwards it as the `user` header
  ([customer_handlers.go](../../services/demo-ui/bff/internal/transport/http/customer_handlers.go)).
- **Email is not currently available server-side.** Auth0 access tokens do not
  include `email` by default. The frontend *does* have it via
  `useAuth().user.email` (from the ID token,
  [context.ts](../../services/demo-ui/ui/src/auth/context.ts)).

## Gap Analysis → Design Decisions

| # | Need | Gap | Decision |
|---|---|---|---|
| Email server-side | Map user → email | `sub` only; no email claim | Add a **namespaced email claim** to the Auth0 access token (Action) and extract it in the BFF into context/header. Demo fallback: trust client-sent email from the verified session (documented caveat). |
| Find party by email | Resolve existing party | Party search has no email filter | **New** `email` criterion on party search (JOIN `party_contact_mediums` where `medium_type='email'`). |
| Find customer by email | Resolve existing customer | Customer search has no email filter | **Reuse**: find party by email → take `partyId` → `query.customer.search?party_id=…` (already supported). Avoids a second new query. |
| Create party if missing | Provision base entity | `cmd.party.create` exists | Reuse; include the email as a preferred `email` contact medium. |
| Create customer | Provision role entity | `cmd.customer.onboard` exists (needs `partyId`) | Reuse, passing the new/found `partyId`. |
| Extra-info form | Collect required fields | No such form/flow | New **"Complete your profile"** UI, prefilled with email + name from the token. |

### Provisioning flow (resolve-or-create, keyed by email)

```
On login (authenticated):
  GET /api/me/customer            # BFF resolves the caller's customer by email
    ├─ find Party by email
    │    ├─ found → find Customer by party_id
    │    │           ├─ found    → return {status: "ready", customer}
    │    │           └─ none     → return {status: "needs_customer", partyId}
    │    └─ none  → return {status: "needs_party"}
    └─ UI:
         ├─ "ready"          → proceed (cache customerId)
         └─ needs_* → route to "Complete your profile" form (prefill email+name)

On form submit:
  POST /api/me/provision {email, givenName, familyName, contactMediums, …}
    1. find Party by email
    2. if no party → cmd.party.create (Individual + email contact medium) → partyId
    3. find Customer by partyId
    4. if no customer → cmd.customer.onboard {partyId, name, contactMediums, …}
    5. return {customer}
```

### Key decisions / trade-offs

- **Orchestration in the BFF.** The resolve-or-create sequence is thin glue over
  existing commands/queries — no business rules beyond "find by email, else
  create". This matches the BFF's existing multi-step RPC orchestration (the
  qualification poll). Alternative considered: a dedicated provisioning **saga** or
  a customer-management use case that owns party creation. Rejected for now as
  over-engineered for a demo onboarding; revisit if provisioning gains real
  invariants. Keep the BFF logic idempotent so it isn't a hidden domain layer.
- **Email is the idempotency key.** Each step re-resolves by email/party_id, so a
  retry after partial failure (party created, customer not) simply finds the party
  and creates only the customer. A concurrent double-login could still race two
  creates — mitigate with a unique email/party guard (see Open Questions).
- **Authorization.** `/api/me/*` are **not** in the Plan 07 public allowlist, so
  they require a valid JWT automatically. No anonymous provisioning.
- **PartyType.** Provisioned parties are `Individual`; given/family name are the
  minimal required fields the token usually cannot supply → the form collects them.
- **Trust boundary.** Prefer the server-derived email claim over client-sent email.
  If the demo trusts the client value, the form's email field must be read-only and
  the caveat documented (a malicious client could claim another email).

## Task List

### Group A — Server-side email identity (BFF + Auth0)

- [~] **A1.** Configure an Auth0 **Action** to add a namespaced `email` custom
  claim to the **access token**. *Tenant config — must be applied in the Auth0
  dashboard (cannot be done from this repo).* The BFF expects the claim key
  `https://tmf-demo/email` (see `auth.EmailClaimKey`). Add a **Login / M2M Action**:

  ```js
  // Auth0 Action: "Add email to access token"
  exports.onExecutePostLogin = async (event, api) => {
    if (event.authorization) {
      api.accessToken.setCustomClaim('https://tmf-demo/email', event.user.email);
    }
  };
  ```

  Until applied, the BFF falls back to the client-provided email (A3).
- [x] **A2.** In the BFF auth middleware
  ([middleware.go](../../services/demo-ui/bff/internal/auth/middleware.go)), extract
  the email claim from validated claims and inject it into context. — **Done**
  (`CustomClaims` + `withIdentity` inject email in both `EnsureValidToken` and
  `OptionalToken`; `EmailFromContext` reads it). **Deviation:** email is kept
  **BFF-local** (read directly by the provisioning handlers) rather than forwarded
  as an RPC header — downstream services receive the email inside command payloads
  (contact medium / search criterion), so no transport header is needed.
- [x] **A3.** Demo fallback: accept a client-provided email in the provisioning
  request body **only** when no verified claim is present; log when the fallback is
  used. Tests for both paths. — **Done** (`Provision` prefers `EmailFromContext`,
  falls back to body email with a warning log; `TestProvision_FallbackToBodyEmail`).

### Group B — Party lookup by email (party-management)

- [x] **B1.** Add an `email` criterion to `SearchParties`
  ([party_repository.go](../../services/party-management/internal/infrastructure/postgres/party_repository.go)):
  match `party_contact_mediums` with `medium_type = 'email'` and `value = ?`. —
  **Done** (parameterized `IN (subquery)`, avoids JOIN row-multiplication; safe).
- [x] **B2.** Surface the criterion through the search handler/DTO so
  `query.party.search` accepts `{ "email": "…" }`. — **Done** (`SearchPartyPayload.Email`).
- [x] **B3.** Tests: party found by email; no match returns empty; email filter
  combines safely with others. — **Done** (`TestSearchParties_ByEmail`: exact match,
  non-email medium excluded, unknown email empty).

### Group C — Provisioning orchestration (BFF)

- [x] **C1.** New `me_handlers.go`: `GET /api/me/customer` — resolve the caller's
  customer by email (party-by-email → customer-by-party_id). Returns a status
  discriminator (`ready` / `needs_party` / `needs_customer`). — **Done**.
- [x] **C2.** `POST /api/me/provision` — orchestrate find-or-create party
  (`cmd.party.create` with email contact medium) then `cmd.customer.onboard` with
  the resolved `partyId`. Idempotent and safe to retry. — **Done**.
- [x] **C3.** Register both routes; confirm they are **outside** the Plan 07 public
  allowlist (auth required). Derive email from context (A2), not the URL/body. —
  **Done** (registered via `MeHandler.RegisterRoutes`; not in `IsPublicRoute`).
- [x] **C4.** Tests: already-a-customer → `ready` (no creates); party exists but no
  customer → creates only customer; nothing exists → creates party then customer;
  partial-failure retry is idempotent. — **Done** (`me_handlers_test.go`, 8 cases).

### Group D — Frontend: provisioning gate & profile form

- [x] **D1.** After login, call `GET /api/me/customer`. Add a small provisioning
  hook exposing `{ status, customerId }`. — **Done** (`useProvisioning` +
  `ProvisioningGate` route element nested inside `RequireAuth`).
- [x] **D2.** If `status !== 'ready'`, route to a new **"Complete your profile"**
  form prefilled with `user.email` (read-only) and `user.name` (split into
  given/family). — **Done** (`CompleteProfilePage`).
- [x] **D3.** On submit, `POST /api/me/provision`; on success seed the resolve
  cache and continue to the intended destination. — **Done** (gate carries
  `returnTo`; success calls `setQueryData(meCustomerQueryKey, result)` so the gate
  passes even without the email claim configured).
- [x] **D4.** Wire `customerId` into checkout (was hardcoded `'demo-customer-id'`).
  — **Done** ([CheckoutPage.tsx](../../services/demo-ui/ui/src/features/ordering/CheckoutPage.tsx)
  now reads `useProvisioning().customerId`).
- [x] **D5.** Tests + `yarn test` / `yarn lint`. — **Done** (CheckoutPage and
  CompleteProfilePage tests; 304 frontend tests pass; no lint errors in changed files).

### Group E — Verification

- [x] **E1.** Backend: `go test ./...`, `go vet ./...` for party-management and the
  BFF (customer-management unchanged). — **Done** (all green; `golangci-lint` still
  blocked by the Plan 07 go1.25/1.26 env mismatch).
- [x] **E2.** Frontend: `yarn test` (304 pass), `yarn lint` (changed files clean). — **Done**.
- [x] **E3.** Manual/Playwright walk-through — **Done** (live stack). Verified:
  anonymous lands on Check Availability with only that nav item + a Log In button
  (Plan 07); an authenticated user with no customer is gated to **Complete your
  profile**; submitting created **one** Party (Individual "Demo User",
  email `demo@rohekood.invalid`) and **one** linked Customer (same `party_id`),
  then landed on the dashboard. A live duplicate-email insert was rejected by
  `uq_party_contact_mediums_email` (case-insensitive). Findings during the
  walk-through (all fixed):
  - **Bug:** the profile form displayed the email read-only but did not send it,
    so provisioning failed when the Auth0 email claim (A1) is absent. Fixed —
    `CompleteProfilePage` now sends `email: user.email` as the body fallback.
  - **Env:** `docker compose up` reused stale images; rebuilt bff/party/demo-ui
    with `--build`. The demo-ui nginx template needs `BFF_HOST`, which compose did
    not set — added `BFF_HOST: bff` to the `demo-ui` service in `docker-compose.yml`.
  - **Confirmed behaviour:** with the A1 email claim not configured in the tenant,
    `GET /api/me/customer` returns 400 and the gate falls back to the form (as
    designed); provisioning still completes via the body-email fallback and the
    seeded resolve cache. Fresh page reloads will re-prompt until A1 is applied.
  - **Auth0:** fresh login only works on the `http://localhost:5173` origin (the
    `http://localhost` callback is not in the tenant's allowed list).

## Suggested Order of Work

1. **Group B** (party-by-email) — the one genuinely new backend query everything
   depends on.
2. **Group A** (server email identity) — establishes the trusted email key.
3. **Group C** (BFF orchestration) — the resolve-or-create engine.
4. **Group D** (frontend gate + form) — user-facing flow.
5. **Group E** (verification) throughout.

## Open Questions / Out of Scope

- **Concurrency guard — DONE.** A partial, case-insensitive **unique index** on
  party email contact mediums hard-stops duplicate provisioning under concurrent
  logins (migration `000012_unique_party_email`:
  `UNIQUE (LOWER(value)) WHERE medium_type = 'email' AND value <> ''`). The BFF
  normalizes email to lower-case before lookups/writes, and on a create-party
  conflict it **re-resolves by email and continues** (so the losing request of a
  race still succeeds). Tests: `TestPartyEmail_UniqueConstraint` (DB) and
  `TestProvision_RaceCreatePartyConflict` (BFF). Party-level uniqueness is
  sufficient — the customer↔email link is unique transitively via `party_id`.
- **Name parsing.** Splitting a single `name` claim into given/family is lossy;
  the form lets the user correct it. Organizations are out of scope (Individual
  only for self-service provisioning).
- **Email change / re-linking.** If a user's email changes in Auth0, re-linking to
  an existing customer is out of scope here.
- **Market segment / tier.** Provisioned customers get defaults; segment-specific
  pricing (Plan 07 follow-up) depends on this `customerId` now being real and
  flowing into qualification — D4 closes part of that gap.
