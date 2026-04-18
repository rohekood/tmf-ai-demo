---
name: ui-architecture-review
description: "Review the demo UI architecture against the repository's frontend, BFF-boundary, routing, state, and testing expectations."
argument-hint: "Optional: specify scope such as routing, state management, design system, API layer, testing, or compare docs vs implementation"
---

# UI Architecture Review

Use this skill when the user wants a frontend architecture review, a UI compliance check, a review of the React app structure, or a gap analysis between the documented UI design and the current implementation.

## Outcome

Produce a UI architecture review that:

- Identifies the UI application's structure, boundaries, and responsibilities
- Checks the frontend against the repository's documented UI architecture
- Verifies how the UI integrates with the BFF and avoids leaking backend concerns into the browser
- Reviews routing, feature modularity, state management, API access, design-system usage, and testing approach
- Distinguishes verified code facts from documented intent
- Reports concrete findings first, with evidence and architectural impact

Default to a review mindset: findings first, ordered by severity.

## Primary Sources

Review these first unless the user narrows the scope:

1. `services/demo-ui/docs/ARCHITECTURE.md`
2. `services/demo-ui/docs/ANALYSIS.md`
3. `services/demo-ui/ui/package.json`
4. `services/demo-ui/ui/src/router.tsx`
5. `services/demo-ui/ui/src/main.tsx`
6. `services/demo-ui/ui/src/api/`
7. `services/demo-ui/ui/src/features/`
8. `services/demo-ui/ui/src/components/`
9. `services/demo-ui/ui/src/design-system/`
10. `services/demo-ui/ui/src/test/`
11. `services/demo-ui/ui/.storybook/`
12. `services/demo-ui/ui/README.md`

If the review extends into the BFF boundary, also consult:

- `docs/ARCHITECTURE.md`
- `services/demo-ui/bff/`

## Procedure

### 1. Frame the Review

Determine:

- Whether the user wants a full UI architecture review or a focused review
- Whether the review is code-only, docs-plus-code, or standards compliance
- Whether the target is the React UI only, or the UI plus BFF boundary

If the request is broad, review the React UI first and include the BFF boundary where it materially affects the frontend architecture.

### 2. Establish the Expected Standard

Extract the expected frontend architecture from the repo docs. At minimum, check for:

- React 19 plus TypeScript in strict mode
- Vite-based application structure
- Separation between UI and BFF deployments
- UI access to backend capabilities only through the BFF, not direct microservice coupling
- TanStack Query for server state
- React hooks, Context, reducer patterns, or local state for client-side UI state
- Feature-oriented organization under `src/features/`
- Shared layout and reusable components separated from feature code
- Design-system components and Storybook support for reusable UI primitives
- Lazy-loaded routes and coherent routing structure
- Accessible component and test patterns
- Vitest and React Testing Library for component behavior testing
- Stateless frontend deployment assumptions
- Environment-based configuration for API endpoints and runtime integration

Do not turn these into assumptions of compliance. Use them as the standard to test the implementation against.

### 3. Inspect the Current UI Shape

Establish how the frontend is actually organized:

- Entry point and router setup
- Feature folders and whether pages, components, API modules, and types are colocated sensibly
- Shared layout components and common utilities
- Design-system primitives versus app-level components
- API client setup, auth token propagation, and HTTP boundaries
- Test coverage patterns and Storybook usage
- Environment configuration and build tooling

Check whether the code reflects the documented architecture or has drifted.

### 4. Review by Concern Area

Check each of these areas explicitly.

#### A. App Shell and Routing

- Is there a clear application entry point and router composition?
- Are routes organized coherently by feature or domain?
- Is lazy loading used appropriately for page-level code splitting?
- Are layout concerns separated from route content?
- Are dead or unused app entry artifacts present and likely to confuse maintenance?

#### B. Feature Modularity

- Are feature modules grouped under `src/features/`?
- Are domain-specific pages, components, API functions, and types kept together or at least consistently organized?
- Is shared code separated from feature code rather than duplicated?
- Do feature boundaries remain readable as the UI grows?

#### C. API and BFF Boundary

- Does the UI communicate only with the BFF-facing API layer?
- Is HTTP access centralized in `src/api/` or another clear boundary?
- Are auth headers or tokens injected in one place instead of scattered across features?
- Is the browser insulated from RabbitMQ and backend service topology?
- Are environment variables used for base URLs and deployment-dependent configuration?

#### D. State Management

- Is server state handled with TanStack Query rather than custom ad hoc fetching patterns?
- Is client UI state kept local unless it must be shared?
- If Context or reducers are used, are they narrowly scoped and justified?
- Are mutation side effects, cache invalidation, and optimistic updates handled consistently?

#### E. Design System and Reuse

- Are reusable components separated from feature-specific UI?
- Does `src/design-system/` contain primitives rather than business-specific compositions?
- Is Storybook used to exercise reusable components?
- Are styling patterns consistent across shared and feature-level components?

#### F. Accessibility and UX Architecture

- Are accessible primitives or role-based patterns evident in components and tests?
- Do tests favor user-visible behavior over implementation details?
- Are loading, empty, error, and auth states handled explicitly in page flows?

#### G. Test and Tooling Architecture

- Are Vitest and React Testing Library set up coherently?
- Do tests exist across shared layout, auth, and major feature areas?
- Are Storybook and test infrastructure maintained as part of the architecture rather than as dead configuration?
- Are linting, build, and test scripts aligned with the documented frontend stack?

### 5. Separate Verified Facts From Documented Intent

Use explicit language such as:

- "UI architecture doc expects..."
- "Current frontend code shows..."
- "No evidence found for..."
- "Configuration exists, but usage is not yet verified"
- "This appears to be planned rather than fully implemented"

If a pattern is described in docs but not clearly visible in code, report that gap rather than assuming it exists.

### 6. Produce the Review

When the user asks for a review, present findings first.

Recommended structure:

1. Findings, ordered by severity
2. Open questions or ambiguous areas
3. Short compliance summary by concern area
4. Optional remediation suggestions if the user asks for them

Each finding should include:

- The frontend architecture expectation being violated or not evidenced
- The code or structure indicating the issue
- Why it matters for maintainability, scalability, user experience, or boundary clarity

## Decision Points

- If the user asks for a strict review, report only evidence-backed findings.
- If the user asks for a checklist, summarize concern areas instead of writing a narrative review.
- If the user asks for docs-versus-code, compare `services/demo-ui/docs/ARCHITECTURE.md` directly to the current React app structure.
- If the user asks about UI architecture broadly, include the BFF boundary but do not drift into a full backend review.
- If the user asks about component architecture, emphasize design-system boundaries, reuse, accessibility, and Storybook coverage.
- If the user asks about data flow, emphasize TanStack Query usage, API modules, auth token handling, and page-level loading or mutation patterns.

## Quality Bar

The review is complete only if it:

- Uses the repository's UI architecture docs and actual frontend structure as the standard
- Checks the BFF boundary rather than treating the UI as an isolated SPA
- Reviews routing, feature organization, API boundaries, state management, reuse, and testing
- Separates documented intent from verified implementation
- Surfaces concrete findings rather than generic frontend advice

## Example Prompts

- `Review the demo UI architecture against the repo's standards.`
- `Compare the documented UI architecture with the current React implementation.`
- `Do a strict architecture review of the demo-ui frontend.`
- `Check whether the UI respects the BFF boundary and frontend state-management rules.`
- `Review the design-system and feature-module architecture in the UI.`