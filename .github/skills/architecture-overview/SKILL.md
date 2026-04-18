---
name: architecture-overview
description: "Produce a comprehensive architecture overview and requirements summary for the TMF solution from repository docs and code structure."
argument-hint: "Optional: audience, depth, specific service/domain, or current-state vs target-state focus"
---

# Architecture Overview

Use this skill when the user asks for an architecture overview, architecture requirements, service interaction summary, system design explanation, or a high-level map of how the TMF platform is intended to work.

## Outcome

Produce a comprehensive but readable architecture summary that:

- Explains the project goal and in-scope business flow
- Identifies the major services and their responsibilities
- Describes the mandatory architectural constraints and quality standards
- Explains the communication model, event topology, and request/reply patterns
- Distinguishes documented target architecture from currently verified implementation when they differ
- Calls out missing details, open questions, or design gaps instead of guessing

## Primary Sources

Review these sources first and prefer them in this order unless the user narrows the scope:

1. `docs/ARCHITECTURE.md`
2. `docs/design/01_architecture_eda.md`
3. `README.md`
4. `antigravity_context.md`
5. `docs/architecture/*.md`
6. `docs/plans/*.md`
7. `services/*/docs/`

If code and docs disagree, state that explicitly and label which facts are documented intent versus verified implementation.

## Procedure

### 1. Frame the Request

Determine these before answering:

- Is the user asking for a generic overview or a service-specific overview?
- Is the user asking for current implementation, target design, or both?
- Is the audience technical, mixed, or executive?
- Is the user asking for an overview, requirements list, or review of architectural compliance?

If the request is broad and no audience is given, default to an engineer-facing overview with moderate detail.

### 2. Extract the Core System Shape

Summarize:

- The project goal and bounded scope
- The major business domains and services
- Which components are sources of truth for which entities
- Which components are infrastructure or bridging layers

For this repository, expect to check for these recurring domains and services:

- Product catalog and service catalog
- Qualification
- Geographic information or address validation
- Party management
- Customer management
- Shopping cart and pricing
- Product ordering or POCV
- Demo UI and BFF

Do not claim a service is implemented only because it appears in planning docs. Verify against repository structure when needed.

### 3. Extract the Non-Negotiable Architecture Rules

Capture the mandatory requirements from the architecture docs. At minimum, check for and include:

- `context.Context` must be the first parameter across layers
- Domain-specific errors, error wrapping with `%w`, and adapter-layer error mapping
- Structured logging with `log/slog`
- Graceful shutdown, signal handling, and retrying infrastructure connections
- Unit tests and integration tests as merge gates
- Clean Architecture or Hexagonal structure with inward dependencies
- Domain isolation from framework and infrastructure packages
- RabbitMQ-based asynchronous communication between backend services
- No direct HTTP coupling between core backend services
- Commands for writes, events for state changes, and async request-reply for reads
- BFF as the bridge between frontend HTTP or WebSocket traffic and backend RabbitMQ traffic
- Integration events for state mutations
- Transactional outbox expectation where consistency matters
- Raw JSON message bodies with context propagated via AMQP headers
- Correlation IDs, user context, authorization propagation, DLQs, and retry policy

### 4. Explain Interaction Patterns

When relevant, describe:

- Topic naming rules such as `<type>.<domain>.<entity>.<action>`
- The difference between `cmd`, `evt`, and `query`
- RPC over RabbitMQ for read paths
- WebSocket streaming through the BFF for long-running flows
- Saga-style orchestration for distributed workflows

If the user asks for flow detail, walk the request path from UI to BFF to RabbitMQ to backend service and back.

### 5. Separate Confirmed Facts From Planned Design

Use clear language such as:

- "Documented architecture requires..."
- "Repository structure confirms..."
- "Planning docs describe a future flow where..."
- "This is not yet verifiable from the current code layout"

This distinction matters because the repository contains both implementation and design-plan material.

### 6. Produce the Response

Structure the answer in this order unless the user asks otherwise:

1. Project purpose and scope
2. Major components and responsibilities
3. Communication and runtime model
4. Mandatory architectural requirements
5. Current-state versus target-state notes
6. Risks, gaps, or open questions

Prefer short sections with concrete statements over vague architectural language.

## Decision Points

- If the user asks for a high-level overview, summarize services, communication model, and constraints without going deep into file layout.
- If the user asks for architecture requirements, emphasize mandatory rules and completion criteria.
- If the user asks for a service-specific overview, focus on that service's role, inputs, outputs, dependencies, and message contracts.
- If the user asks whether the implementation matches the architecture, compare docs against the actual repository and call out deviations.
- If the audience is non-technical, reduce implementation details but keep boundaries, responsibilities, and integration model intact.

## Quality Bar

The output is complete only if it:

- Names the architectural style and the async-first communication model
- Identifies the BFF as the frontend protocol bridge
- Explains the clean-architecture dependency direction
- Mentions context propagation, testing expectations, and operational resilience requirements
- Distinguishes documented intent from verified code when certainty is limited
- Avoids inventing flows, services, or guarantees not supported by repo sources

## Example Prompts

- `Give me an architecture overview of this TMF solution.`
- `Summarize the mandatory architecture requirements in this repo.`
- `Explain how the BFF, RabbitMQ, and backend services are supposed to interact.`
- `Compare the documented target architecture with what is currently implemented.`
- `Give me a qualification-service-focused architecture overview.`