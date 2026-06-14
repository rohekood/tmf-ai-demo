# UX Improvement Task List

Derived from [UX-DESIGN-REVIEW.md](./UX-DESIGN-REVIEW.md) (review date 2026-06-13).

**Status: all P1–P3 items completed and verified on 2026-06-14.**
Validated with `yarn test` (180 passed), `yarn lint` (no new issues), `tsc -b` (clean), the new `yarn lint:classes` guard, and a live Playwright walkthrough.

---

## P1 — High impact

- [x] **Fix Debug Console WebSocket reconnect loop**
  - [x] Derive WS URL from configured API origin / page origin (exported `getWebSocketUrl`)
  - [x] Capped exponential backoff (1→2→4→8→16→30s) with a 6-attempt ceiling, then stop
  - [x] Show "Connecting…" / "Reconnecting… (N)" / "Event stream offline" + a Retry button
  - [x] Throttle error logging to the first failure of a streak (console errors dropped ~88→~7)
  - Files: `ui/src/features/debug/useDebugWebSocket.ts`, `DebugConsolePage.tsx` (+ unit tests)

- [x] **Disambiguate the two connection indicators**
  - [x] Sidebar relabelled "Connected" → "API online" (with title tooltip)
  - [x] Debug Console status now describes the event stream ("Live" / "Event stream offline")
  - Files: `ui/src/components/layout/Sidebar.tsx`, `DebugConsolePage.tsx`

- [x] **Standardize empty states**
  - [x] New shared `EmptyState` component (icon + title + description + action, `bare` variant) with tests
  - [x] Adopted on Parties, Customers, Offerings, Catalogs, Specifications, Categories
  - [x] Shopping Cart now has an icon, description, and a primary CTA (was a weak text link)
  - Files: `ui/src/design-system/components/common/EmptyState.{tsx,css,test.tsx}`, feature list pages, `CartPage.tsx`

## P2 — Medium impact

- [x] **Persist auth session across reloads** — `cacheLocation="localstorage"` + `useRefreshTokens` on `Auth0Provider`; verified a full reload no longer bounces to login. File: `ui/src/auth/client.tsx`
- [x] **Fix heading hierarchy** — banner demoted to non-heading branding; every list page title promoted to `<h1 className="page-title">`. Files: `Layout.tsx`/`.css`, all list pages
- [x] **Page headers consistent** — verified Qualify/Cart/Checkout already use the standard left-aligned `page-header` (the centered version had been refactored)
- [x] **Fix mobile search bar input collapse** — `min-width: 0` + `flex: 1 1 12rem` + `flex-wrap`; input keeps a usable width and controls wrap below on mobile. File: `ui/src/features/parties/PartyListPage.css`
- [x] **Login: report the real disabled reason** — `disabledReason` ('insecure-origin' | 'not-configured') surfaced via auth context; distinct label + hint. Files: `auth/context.ts`, `auth/client.tsx`, `LoginPage.tsx`

## P3 — Low / polish

- [x] **Finish Tailwind-class cleanup + add a guard**
  - [x] Removed inert utilities from design-system `Sidebar.tsx` (real `--collapsed`/`--center` classes + CSS) and `CategoryEditPage.tsx`
  - [x] Added `yarn lint:classes` (`scripts/check-dead-classes.mjs`) banning inert Tailwind/Bootstrap utility classes
- [x] **Hide decorative icons from screen readers** — `aria-hidden` on sidebar nav icons/arrow and `EmptyState` icon
- [x] **Fix faint-text contrast** — `--text-faint` `#64748b` (~3.95:1) → `#74829a` (~4.9:1, passes AA)
- [x] **Off-canvas mobile nav** — `visibility: hidden` when closed removes it from tab order / a11y tree (verified links non-interactive when closed). File: design-system `Sidebar.css`
- [x] **Keyboard focus visibility** — global `:focus-visible` outline for links/buttons/`[role=button]`/`[tabindex]`. File: `index.css`

## Follow-up review (still blocked on data)

- [ ] Seed demo data, then review populated list/table states (density, pagination, sorting, virtualization)
- [ ] Visually review Checkout, Order Status, Order Confirmation, and detail/edit pages (reviewed from source only)
