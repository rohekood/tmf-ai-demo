---
name: go-test-coverage-validation
description: "Validate that Go tests meet a 90% coverage bar using the repository's actual Go module layout and `go test` coverage output. Use for coverage gates, pre-PR validation, service-level test quality checks, and identifying low-coverage Go packages."
argument-hint: "Optional: target service, package, or module path; strictness; and whether to suggest the smallest test additions needed to reach 90%"
---

# Go Test Coverage Validation

Use this skill when the user wants to verify that Go code has enough test coverage, check whether a service or module meets a 90% coverage bar, or identify the Go packages most responsible for missing the target.

## Outcome

Produce an evidence-based Go coverage validation that:

- Runs the repository's real Go test tooling instead of a generic coverage estimate
- Measures coverage for the requested Go scope within this workspace
- Applies a 90% coverage gate and reports pass or fail explicitly
- Identifies the packages or files most responsible for missing the threshold
- Distinguishes tooling limitations from actual test coverage gaps

Default to a validation mindset: pass or fail first, then the main deficits and next actions.

## Primary Sources

Inspect these before concluding:

1. `go.work`
2. The target module's `go.mod`
3. The target module's `cmd/` and `internal/` directories where applicable
4. The target module's `*_test.go` files
5. Repository instructions that define the quality bar, especially `.github/instructions/post-implementation-review.instructions.md`

Relevant Go modules in this repository currently include:

- `pkg`
- `services/customer-management`
- `services/demo-ui/bff`
- `services/party-management`
- `services/pocv`
- `services/product-catalog-management`
- `services/qualification`
- `services/shopping-cart`
- `tests/e2e`

If the user asks for repo-wide Go coverage, verify first whether the request should exclude `tests/e2e` and other non-unit-test scopes. If unspecified, default to validating a single Go module or service rather than the entire workspace aggregate.

## Procedure

### 1. Frame the Coverage Scope

Determine:

- Whether the user wants a single service, a single package, a whole Go module, or a broader repo-wide check
- Whether the user wants check-only validation or validation plus remediation suggestions
- Whether the user wants current behavior only, or also wants tooling changes that enforce the threshold automatically

If unspecified, default to the smallest meaningful Go module or service scope rather than attempting a workspace-wide aggregate.

### 2. Confirm the Repository Coverage Workflow

Use repository-defined Go tooling as the source of truth.

For this repo, verify:

- `go.work` includes the target module
- The target path has its own `go.mod`
- Coverage is collected from the target module directory so package resolution is correct

Preferred commands:

For a module or service:

```bash
cd <target-module> && go test ./... -coverprofile=coverage.out -covermode=atomic
cd <target-module> && go tool cover -func=coverage.out
```

For a narrower package scope:

```bash
cd <target-module> && go test ./path/to/package -coverprofile=coverage.out -covermode=atomic
cd <target-module> && go tool cover -func=coverage.out
```

If the repository later adds a dedicated coverage script or Make target, prefer that script over an ad hoc command.

### 3. Run the Coverage Check

Execute the coverage run from the target Go module directory.

Capture:

- The `go test` result for the selected scope
- The total statement coverage reported by `go tool cover -func`
- Package-level or file-level hot spots that are materially below the threshold
- Whether any packages were skipped, failed, or produced no coverage data

Do not infer a pass from passing tests alone. The coverage decision must come from the actual coverage report.

### 4. Apply the 90% Gate

Validate the result against a 90% threshold.

Default interpretation:

- Use total statement coverage as the authoritative gate because that is what standard Go tooling reports directly
- Require at least 90% total statement coverage for a pass decision
- If the user asks for strict validation, also call out any important subpackages that fall materially below 90% even when the module aggregate passes

If repository tooling does not enforce the threshold automatically, report that as a tooling gap rather than pretending the gate is automated.

### 5. Separate Meaningful Gaps From Noise

Inspect what is dragging coverage down.

Check for:

- Core domain or use-case packages with weak `*_test.go` coverage
- Thin transport or config packages that distort the aggregate despite low business risk
- Generated, mock, or wiring-heavy files that may need explicit exclusion if they are not meaningful coverage targets
- Integration-only packages where low statement coverage reflects missing unit tests rather than missing end-to-end tests

Do not silently normalize results by excluding files unless the user explicitly asks for a normalized report or the repository already defines exclusions.

### 6. Produce the Validation Result

When reporting results, provide:

1. Overall pass or fail against the 90% target
2. The exact total coverage percentage used for the decision
3. The scope that was tested
4. Whether the threshold is enforced in tooling or only evaluated manually
5. The main low-coverage packages or files
6. Minimal next actions if the user asked for remediation

## Decision Points

- If the user asks for a service-level check, run coverage from that service's module directory rather than the repo root.
- If the user asks for a package-level check, report that scope separately and do not confuse it with module-wide coverage.
- If the user asks for repo-wide Go coverage, make clear whether the result is an aggregate of multiple module runs or a single command scope.
- If the user asks for changed-file coverage specifically, explain that vanilla `go test` coverage is strongest at package and aggregate statement level; file-specific or changed-line gating may require extra analysis or tooling.
- If the repo lacks an enforced 90% threshold in CI or scripts, treat that as a standards gap and recommend adding it.

## Completion Checks

Validation is complete only if all are true:

- The coverage command was run from the correct Go module directory
- The response includes the coverage percentage used for the decision
- The response makes an explicit 90% pass or fail call
- The response states whether the threshold is actually enforced in tooling
- The main low-coverage packages or files are identified with evidence when the target fails or is near the threshold

## Repo-Specific Notes

Current repository evidence shows:

- This workspace uses a multi-module Go layout coordinated by `go.work`
- Go services and shared packages each have their own `go.mod`
- The repository review instructions require test coverage of 90% or higher for changed files unless explicitly approved otherwise
- No repository-level, dedicated Go coverage gate script has been confirmed yet

That means Go coverage validation may currently be a reporting workflow rather than an enforced CI-quality gate unless the repo adds explicit automation.

## Example Prompts

- `Validate whether shopping-cart meets 90% Go test coverage.`
- `Run a strict 90% Go coverage gate for services/qualification.`
- `Check which packages are keeping customer-management below 90% coverage.`
- `Validate Go coverage for pkg and suggest the smallest test additions needed to reach 90%.`
- `Tell me whether the repository actually enforces a 90% Go coverage gate or only reports it manually.`