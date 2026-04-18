---
name: service-architecture-review
description: "Review a single TMF service against the repository's architecture, deployability, communication, and persistence standards."
argument-hint: "Specify the service name and optionally whether to check code only, docs plus code, or focus areas such as persistence, messaging, or structure"
---

# Service Architecture Review

Use this skill when the user wants an architecture review of one service, a compliance check against repo standards, or a gap analysis between a service implementation and the documented architecture.

## Outcome

Produce a service-level architecture review that:

- Identifies the service's role, boundaries, and dependencies
- Checks the service against the repository's mandatory architecture standards
- Distinguishes verified implementation facts from documented requirements
- Reports concrete findings, risks, and missing elements with evidence
- Highlights what is compliant, what is missing, and what is unclear

Default to a code-review mindset: findings first, ordered by severity, with concise supporting context.

## Primary Sources

Review these before concluding:

1. `docs/ARCHITECTURE.md`
2. `docs/design/01_architecture_eda.md`
3. `docs/architecture/GENERAL_STANDARDS.md`
4. `docs/architecture/POSTGRES_STANDARDS.md`
5. `services/<service>/README.md`
6. `services/<service>/docs/`
7. `services/<service>/cmd/`
8. `services/<service>/internal/`
9. `services/<service>/go.mod`
10. `services/<service>/Dockerfile`

If the service path is ambiguous, resolve it before reviewing. If the service is missing expected files, treat that as a finding rather than silently working around it.

## Procedure

### 1. Frame the Review

Determine:

- Which service is being reviewed
- Whether the user wants code-only review, docs-plus-code review, or standards compliance review
- Whether the user wants a full review or a focused review such as messaging, persistence, deployability, or layering

If the request is broad, perform a full service-level review.

### 2. Establish the Expected Standard

Extract the service obligations from the repo docs. At minimum, the review must check for:

- Independent deployability under `services/<service>`
- Presence of `go.mod`, `Dockerfile`, and expected server entrypoint
- Environment-variable based configuration
- Strict asynchronous inter-service communication through RabbitMQ
- No direct access to another service's database
- Structured logging with `log/slog`
- Propagation of `X-Correlation-ID` and `X-User-ID`
- Clean or hexagonal layering with dependencies flowing inward
- `context.Context` as the first parameter across layers
- Domain isolation from transport, ORM, and broker packages
- Domain-specific error handling and wrapped lower-level errors
- Graceful shutdown and retry-aware infrastructure setup
- Unit and integration test expectations
- Persistence rules where applicable: no `AutoMigrate`, migration files, explicit DAO or domain separation, transactional outbox, transaction management, audit support, and environment-sourced credentials

Do not assume all services currently comply. The purpose is to test the service against the standard, not to defend the current implementation.

### 3. Inspect the Service Shape

Establish the current implementation structure:

- Entry points in `cmd/`
- Package organization under `internal/`
- Presence of handlers, repositories, gateways, use cases, and domain packages
- Messaging clients and topic usage
- Persistence code, migrations, DAOs, and repositories
- Tests and test scope

This repository contains more than one internal layout convention, for example:

- `adapter/core/infrastructure/usecase`
- `config/domain/infrastructure/transport`

Do not mark a service non-compliant just because it uses a different naming convention. Check whether the dependency direction and separation of concerns still satisfy the architectural rules.

### 4. Review by Concern Area

Check each of these areas explicitly.

#### A. Deployability

- Is the service independently deployable?
- Does it have a `Dockerfile`?
- Does it have a Go module?
- Does it expose a server entrypoint in the expected shape?
- Is configuration sourced from environment variables rather than hardcoded values?

#### B. Layering and Dependency Direction

- Is business logic separated from transport and infrastructure?
- Do dependencies point inward?
- Does domain code avoid importing `gorm`, `amqp`, `gin`, or equivalent framework packages?
- Are interfaces and adapters placed coherently?

#### C. Communication Model

- Does the service communicate with other backend services via RabbitMQ rather than direct HTTP?
- Are commands, events, and queries used consistently with the documented model?
- Are correlation and user context propagated across boundaries?
- If the service interacts with the UI or BFF, is that boundary explicit and appropriate?

#### D. Persistence and Transactions

- Does the service avoid `gorm.AutoMigrate()`?
- Are migrations present in the expected location?
- Are domain models kept free of `gorm` tags?
- Are DAOs and mapping boundaries explicit?
- Is transaction ownership in the use case or service layer rather than buried in repositories?
- Is a transactional outbox used for state changes that emit events?
- Are audit hooks and `app.current_user` handling present where required?

#### E. Reliability and Operations

- Does the service support graceful shutdown?
- Are broker and database connections initialized with retry and pooling considerations?
- Does logging include enough structured context for tracing?

#### F. Testing

- Are there unit tests for domain or isolated logic?
- Are there integration tests for messaging, database, or use-case flows?
- Are there obvious gaps where architecture-critical behavior is untested?

### 5. Separate Verified Facts From Expected Standards

Use language that makes the evidence clear:

- "Repo standard requires..."
- "Service implementation shows..."
- "No evidence found for..."
- "Docs imply..., but current code does not yet confirm it"

If you cannot verify an item from the available code or docs, say so directly instead of inferring compliance.

### 6. Produce the Review

When the user asks for a review, present findings first.

Recommended structure:

1. Findings, ordered by severity
2. Open questions or unclear areas
3. Short compliance summary by concern area
4. Optional remediation suggestions if the user asks or if the fixes are obvious

Each finding should include:

- What standard is being violated or not evidenced
- What in the service indicates the problem
- Why it matters architecturally or operationally

## Decision Points

- If the user asks for a strict review, report only evidence-backed findings and avoid best-effort approval language.
- If the user asks for a checklist, summarize compliance area by area instead of writing a narrative review.
- If the user asks for a gap analysis, compare documented expectations to current implementation and call out deltas.
- If the service has no persistence layer, mark persistence rules as not applicable rather than failed.
- If the service is a BFF or UI-adjacent component, adapt the communication review to account for frontend-facing HTTP or WebSocket boundaries while still enforcing async backend integration.

## Quality Bar

The review is complete only if it:

- Uses the repository architecture docs as the standard rather than generic microservice advice
- Verifies the actual service layout before judging compliance
- Distinguishes naming differences from real architectural violations
- Checks deployability, layering, communication, persistence, reliability, and testing
- Clearly separates verified implementation facts from undocumented assumptions
- Surfaces concrete findings instead of a vague "looks good"

## Example Prompts

- `Review the qualification service against the repo's architecture standards.`
- `Give me a service-level architecture review of shopping-cart.`
- `Check whether customer-management complies with the persistence and messaging standards.`
- `Compare the pocv service implementation with the documented architecture rules.`
- `Do a strict architecture compliance review of party-management.`