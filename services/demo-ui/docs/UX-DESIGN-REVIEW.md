# UX & Design Review — Demo UI

**Date:** 2026-06-13
**Method:** Live walkthrough of the running dev app (`http://localhost:5173`) driven with Playwright, logged in as the `demo` user, plus source/CSS inspection.
**Scope requested:** Visual design & consistency · Usability & flows · Accessibility · Responsiveness.

> ⚠️ **Caveat — moving target.** The frontend was being actively refactored *during* this review (a Tailwind → semantic-CSS migration). Several findings about "dead Tailwind classes" that were true at the start were fixed mid-review. This document reflects the state observed on the date above; verify each item against current `main` before acting. Screenshots are in [`ux-review-assets/`](./ux-review-assets/).

---

## TL;DR

The app is in **good shape**. It has a genuinely cohesive dark theme built on a well-structured design-token system ([`src/index.css`](../ui/src/index.css)), a consistent app shell, sensible empty states, and working mobile responsiveness. It is **not** using Tailwind (no install, no config — utility classes are inert); the real styling is a hand-built semantic CSS design system, and the few remaining Tailwind utility classes are migration leftovers being actively removed.

The highest-value improvements are about **consistency and robustness**, not a redesign: standardize empty states, fix the Debug Console's runaway WebSocket reconnect loop, disambiguate the two "connection" indicators, and tidy a handful of accessibility details.

---

## What's working well

- **Cohesive visual language.** A single dark theme with a proper token layer (surfaces, borders, text tiers, semantic colors, radius, shadows) in [`index.css`](../ui/src/index.css). Components (`.btn`, `.card`, `.form`, `.badge`, `.alert`, `.data-list`) are reusable and consistent.
- **Clear app shell.** Left sidebar with logical grouping (Management · Product Catalog · Ordering · Developer), a persistent top banner, and a user card with avatar + connection status. Easy to scan. See [parties](./ux-review-assets/ux-02-parties.png).
- **Good empty states with next-step CTAs** on most list pages ("Create your first party / customer / offering").
- **Comprehensive, well-sectioned forms.** The Create Party page groups Personal Info, Contact Mediums, Identifications, Tax Exemptions, External References, Attachments, and Related Parties into discrete cards with their own "Add" actions — [screenshot](./ux-review-assets/ux-03-party-new.png).
- **Responsive.** At 375 px the sidebar collapses to a hamburger, the header stacks, and content goes full-width — [mobile](./ux-review-assets/ux-12-mobile-parties.png).
- **Login page** has been refactored to the design system and looks polished (glassmorphism card, primary button) — [login](./ux-review-assets/ux-01-login.png).

---

## Findings (prioritized)

### P1 — High impact

**1. Debug Console WebSocket: infinite reconnect loop + permanent error spam.**
The Debug Console opens `ws://localhost/ws/debug` and the handshake returns HTTP 200 instead of 101, so it fails and immediately retries — **88+ console errors** accumulated during a short session, and the panel sits permanently on a red "Disconnected" badge ([debug screenshot](./ux-review-assets/ux-11-debug.png)).
Root cause is a dev/prod origin mismatch: the WS URL targets `localhost` (port 80, the nginx/prod build) while the dev server runs on `:5173`.
- Derive the WS URL from the current origin (or proxy `/ws` through Vite), don't hardcode the prod host.
- Add capped exponential backoff and stop the tight retry loop.
- Surface a real "Reconnecting… (attempt N)" state instead of a static "Disconnected".
- File: [`src/features/debug/useDebugWebSocket.ts`](../ui/src/features/debug/useDebugWebSocket.ts).

**2. Two different "connection" concepts share one visual language.**
The sidebar shows a green **"Connected"** while the Debug Console simultaneously shows a red **"Disconnected"** — both rendered as the same status-dot pattern. Users can't tell these are different channels (app/API liveness vs. the debug event stream).
- Label them distinctly: e.g. "API: Connected" and "Event stream: Disconnected", or use different iconography.

**3. Empty states are inconsistent across the app.**
Three different treatments observed:
| Page | Icon? | CTA style |
|---|---|---|
| Offerings | ✅ cart icon | prominent `btn-primary` |
| Parties / Customers | ❌ none | prominent `btn-primary` |
| Shopping Cart | ❌ none | weak `btn-link` text link |
The cart empty state is also an oversized card holding two short lines ([cart](./ux-review-assets/ux-09-cart.png)).
- Extract a single `<EmptyState icon message action />` component and use it everywhere. Give the cart a real primary CTA, an icon, and right-size the card.

### P2 — Medium impact

**4. Auth session is lost on full page reload.**
A hard refresh drops the user back to the login screen (the Auth0 client uses in-memory cache, so silent re-auth fails). This hurts perceived reliability and bookmarking/deep-linking.
- Set `cacheLocation: 'localstorage'` and `useRefreshTokens: true` on `Auth0Provider` ([`src/auth/client.tsx`](../ui/src/auth/client.tsx)).

**5. Every page's `<h1>` is identical.**
The banner renders `<h1>TMF Demo Dashboard</h1>` on every route ([`Layout.tsx:52`](../ui/src/components/layout/Layout.tsx#L52)), while the page-specific title (e.g. "Parties") is an `<h2>`. Screen readers announce the same page name everywhere, and the visual hierarchy has two competing titles.
- Make the page-specific title the `<h1>`; demote the banner to a slimmer contextual header/breadcrumb, or move it into a `<header>` without an `<h1>`.

**6. Page-header alignment is inconsistent.**
List pages left-align the title with the action button on the right; the Service Qualification page centers its heading and card ([qualify](./ux-review-assets/ux-08-qualify.png)). Pick one header pattern and apply it consistently.

**7. Mobile search bar collapses the input.**
At 375 px the search text field shrinks to a tiny unusable box squeezed between the icon and the "Search" button, hiding the placeholder ([mobile](./ux-review-assets/ux-12-mobile-parties.png)).
- Fix the flex sizing so the input keeps a sensible min-width or the controls wrap.

**8. Login button reports the wrong disabled reason.**
When disabled, the button always reads "Auth Requires HTTPS or localhost" even when the real cause is missing Auth0 config. Surface the actual reason ([`LoginPage.tsx`](../ui/src/features/auth/LoginPage.tsx) + `isAuthEnabled` in `client.tsx`).

### P3 — Low / polish

**9. Finish the Tailwind-class cleanup + guard against regressions.**
Inert utility classes remain (e.g. `justify-center px-0` in [`Sidebar.tsx:116`](../ui/src/design-system/components/layout/Sidebar.tsx#L116) — collapsed-icon centering silently does nothing). These resolve to no styles because Tailwind isn't installed. Remove the last remnants and add an ESLint rule (or a simple CI grep) banning unknown utility class names so they can't creep back in.

**10. Decorative icons add screen-reader noise.**
Lucide icons render as bare `img` nodes with no accessible name in the a11y tree. Add `aria-hidden="true"` (or `role="img"` + label where meaningful) to decorative icons.

**11. Faint text likely fails WCAG AA contrast.**
`--text-faint: #64748b` on the dark app background is ≈ 3.6:1 — below the 4.5:1 AA threshold for normal-size text (used for placeholders and hints). Lighten the token for small text or restrict its use to large text.

**12. Off-canvas mobile nav stays in the a11y tree when closed.**
The collapsed sidebar's nav links remain focusable/announced. Apply `inert` / `aria-hidden` to the menu when it's closed.

**13. Verify keyboard focus visibility.**
Inputs have a clear focus ring (`index.css`), but confirm nav links, cards, and icon-buttons (e.g. Log out, Collapse Sidebar) have visible focus states for keyboard users.

---

## Notes & limitations

- **No seed data.** All backends were empty during the review, so list/table density, pagination, sorting, row hover, and detail/edit pages were only partially exercised. A follow-up pass with populated data is worthwhile (especially `@tanstack/react-table` + virtualization screens).
- **Not visually verified:** Checkout, Order Status/Confirmation, and detail/edit pages (not reachable via the sidebar, and full-reload navigation drops the session per finding #4). Reviewed from source only.
- The app correctly uses **Radix** primitives + **lucide** icons + the semantic CSS system — a deliberate, coherent stack. No move to Tailwind is recommended; finishing the migration *away* from leftover utility classes is the right direction.

## Suggested order of work

1. Debug WebSocket loop (#1) + connection-label disambiguation (#2) — stops console spam and user confusion.
2. Shared `EmptyState` component (#3) — quick consistency win across many pages.
3. Auth persistence (#4) — removes the reload-logout papercut.
4. Heading hierarchy (#5) + accessibility polish (#10–13).
5. Header alignment (#6) and mobile search (#7) cleanups.
