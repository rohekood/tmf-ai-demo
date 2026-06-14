# Analysis — Catalog Detail Views & Detail-Page Styling

Review date: 2026-06-14. Scope: catalog feature (`ui/src/features/catalog`).

This document analyses four reported defects and turns each into actionable tasks.
All four share two underlying causes:

1. **Catalogs and Categories have no read-only detail page** — their "view" route is wired to the *edit* form.
2. **The catalog detail pages reference CSS classes that are never defined**, so spacing
   between blocks (and between an icon and its label) collapses.

---

## Root-cause findings

### A. No detail page for Catalogs / Categories

The router maps the bare `:id` route to the **edit** component for both resources:

- [`router.tsx:150-156`](../ui/src/router.tsx#L150-L156) — `catalog/catalogs/:id` → `CatalogEditPage`
- [`router.tsx:187-193`](../ui/src/router.tsx#L187-L193) — `catalog/categories/:id` → `CategoryEditPage`

Contrast with Specifications and Offerings, which *do* have a detail component on `:id`
([`router.tsx:224-230`](../ui/src/router.tsx#L224-L230), [`router.tsx:261-267`](../ui/src/router.tsx#L261-L267)).

Consequences:

- **Issue 1 (Catalog).** [`CatalogListPage.tsx:69`](../ui/src/features/catalog/CatalogListPage.tsx#L69)
  renders a "View" (eye) button to `/catalog/catalogs/:id`, but that route is the edit form —
  identical to the "Edit" button next to it. Inside the form,
  [`CatalogEditPage.tsx:38-41`](../ui/src/features/catalog/CatalogEditPage.tsx#L38-L41) shows a
  **"Back to Details"** link pointing at `/catalog/catalogs/:id` — i.e. back to *itself*, so it
  appears to do nothing. There is no details view to go back to.
- **Issue 2 (Categories).** Same wiring (`:id` → `CategoryEditPage`), and worse:
  [`CategoryListPage.tsx:79-82`](../ui/src/features/catalog/CategoryListPage.tsx#L79-L82) has only
  **Edit** and **Delete** buttons — no "View" button at all.

### B. Catalog detail pages import a stylesheet missing their layout classes

`SpecificationDetailPage` and `OfferingDetailPage` both `import './SpecificationListPage.css'`
([`SpecificationDetailPage.tsx:4`](../ui/src/features/catalog/SpecificationDetailPage.tsx#L4),
[`OfferingDetailPage.tsx:5`](../ui/src/features/catalog/OfferingDetailPage.tsx#L5)), but that file
does **not** contain the classes the markup uses. An audit of every `.css` under `ui/src`:

| Class | Defined? |
|---|---|
| `detail-grid` | only in `parties/PartyDetailPage.css` (lazy-loaded, not imported here) |
| `section-header` | yes (2 files) |
| `main-content`, `sidebar-content`, `detail-section`, `info-card`, `stats-list`, `stat-item`, `metadata-grid`, `metadata-item`, `char-item`, `char-header`, `char-title`, `char-type-badge`, `values-cloud`, `value-tag` | **none — undefined** |

Because `.detail-grid` lives only in [`PartyDetailPage.css:191-196`](../ui/src/features/parties/PartyDetailPage.css#L191-L196)
(`display:grid; grid-template-columns: repeat(2,1fr); gap:1.5rem`), the two-column layout and its
`gap` only apply when that stylesheet happens to already be loaded. On a fresh navigation to a
catalog detail page it is not, so the columns/blocks render as plain stacked `<div>`s with **no gap**.

- **Issue 3 (Specification detail).** "Characteristics" (in `.main-content`) and "Quick Stats"
  (in `.sidebar-content`) have no spacer because `.detail-grid` gap is absent and
  `.main-content`/`.sidebar-content` are unstyled.
- **Issue 4a (Offering detail).** Same cause: no spacer between "Pricing" and "Quick Stats".

### C. Offering price line: dollar icon glued to a non-dollar amount

[`OfferingDetailPage.tsx:137-139`](../ui/src/features/catalog/OfferingDetailPage.tsx#L137-L139)
hardcodes a `<DollarSign>` icon, then renders `formatMoney(price)`, which formats with the price's
*actual* currency via `Intl.NumberFormat`
([`OfferingDetailPage.tsx:13-20`](../ui/src/features/catalog/OfferingDetailPage.tsx#L13-L20)). So a
EUR price shows a **`$` icon immediately followed by `€10.00`** — wrong symbol *and* no space,
because `.char-title` (which should supply `display:flex; gap; align-items`) is undefined (cause B).

- **Issue 4b.** The "dollar and euro with no space" is two bugs: a misleading hardcoded `$` icon,
  and the missing `.char-title` flex/gap styling.

---

## Tasks

**Status: implemented, tested, and verified via a live Playwright walkthrough on 2026-06-14.**

### Task 1 — Add a read-only Catalog detail page (Issue 1) ✅
- [x] Create `ui/src/features/catalog/CatalogDetailPage.tsx` (read-only), mirroring the
      `SpecificationDetailPage` layout: back link → `/catalog/catalogs`, an "Edit Catalog" button →
      `/catalog/catalogs/:id/edit`, and cards for overview/metadata (name, description, lifecycle
      status, validity, last update) using the catalog fields from `useCatalog`.
- [x] In [`router.tsx`](../ui/src/router.tsx#L151-L157), point `catalog/catalogs/:id` at the new
      `CatalogDetailPage` (added the lazy import); `:id/edit` → `CatalogEditPage`.
- [x] [`CatalogEditPage.tsx:38-41`](../ui/src/features/catalog/CatalogEditPage.tsx#L38-L41) "Back to
      Details" link now resolves to the real detail page (no code change needed — route now exists).
- [x] Add unit tests (`CatalogDetailPage.test.tsx`): renders fields, loading, not-found,
      Edit link target.

### Task 2 — Add a read-only Category detail page + View button (Issue 2) ✅
- [x] Create `ui/src/features/catalog/CategoryDetailPage.tsx` (read-only): name, description,
      lifecycle status, root flag, parent (link to parent detail), validity period; "Edit Category"
      button → `/catalog/categories/:id/edit`; back link → `/catalog/categories`.
- [x] In [`router.tsx`](../ui/src/router.tsx#L189-L195), point `catalog/categories/:id` at the new
      `CategoryDetailPage`; `:id/edit` → `CategoryEditPage`.
- [x] In [`CategoryListPage.tsx`](../ui/src/features/catalog/CategoryListPage.tsx#L79-L85),
      add a "View" (Eye) action linking to `/catalog/categories/:id`, before the Edit action.
- [x] Add unit tests (`CategoryDetailPage.test.tsx`) and extend `CategoryListPage.test.tsx` for the
      new View button.

### Task 3 — Define the missing catalog detail-page styles (Issues 3 & 4a) ✅
- [x] Create a shared `ui/src/features/catalog/CatalogDetail.css` defining the currently-undefined
      classes used by all four detail pages: `detail-grid` (2-col grid, `gap: 1.5rem`, single
      column under 1024px), `main-content`, `sidebar-content`, `detail-section`, `info-card`,
      `stats-list`, `stat-item`, `metadata-grid`, `metadata-item`, `char-item`, `char-header`,
      `char-title` (`display:flex; align-items:center; gap:.5rem`), `char-type-badge`,
      `values-cloud`, `value-tag`. Does **not** rely on `PartyDetailPage.css`.
- [x] Import this stylesheet from `SpecificationDetailPage`, `OfferingDetailPage`, and the two new
      detail pages, so spacing is guaranteed regardless of load order.
- [x] Verified with Playwright: a clear spacer is visible between Characteristics↔Quick Stats (spec)
      and Pricing↔Quick Stats (offering) on a direct/cold navigation.
- [x] Run `yarn lint:classes` (dead-class guard) — passes (`No banned utility classes found`).

### Task 4 — Fix the offering price symbol/spacing (Issue 4b) ✅
- [x] In [`OfferingDetailPage.tsx`](../ui/src/features/catalog/OfferingDetailPage.tsx#L137-L139),
      replaced the hardcoded `<DollarSign>` icon (which contradicted the formatted currency) with a
      currency-neutral `<Coins>` icon; the amount already includes the correct symbol via
      `Intl.NumberFormat`.
- [x] `.char-title` gap (Task 3) now provides spacing between the icon and the amount.

### Cross-cutting validation
- [x] `yarn test` (222 passed) and `yarn lint` (no new issues — 22 pre-existing errors in ordering
      tests / parties pages, none in changed files); `tsc -b` clean.
- [x] Coverage ≥ 90% on changed files (per CLAUDE.md rule 5): CatalogDetailPage 100%,
      CategoryDetailPage 100% lines / 96.7% branch, OfferingDetailPage 100% lines / 91.1% branch,
      SpecificationDetailPage 100% lines / 92.3% branch, CategoryListPage 96.7% lines / 93.8% branch
      (all metrics ≥ 90%).
- [x] Playwright walkthrough: Catalog list → View → Edit → Back to Details (round-trips);
      Category list → View → detail; Specification & Offering detail spacing confirmed; offering
      price renders `€21.00` with a neutral coins icon and no stray `$`.

---

## Suggested order
Task 3 first (unblocks correct rendering for the new pages and fixes Issues 3/4a), then Task 4,
then Tasks 1 & 2 (which reuse the Task 3 styles). Tasks 1 and 2 are independent of each other.
