---
name: react-unit-test-coverage-validation
description: "Validate that the demo UI React unit test suite meets an 80% coverage bar using the repository's Vitest setup. Use for coverage gates, pre-PR validation, and identifying gaps in UI unit test coverage."
argument-hint: "Optional: target path within the UI, strictness, and whether to suggest remediation for files below threshold"
---

# React Unit Test Coverage Validation

Use this skill when the user wants to verify that the React UI has enough unit test coverage, check whether the UI meets an 80% coverage bar, or identify the main TSX areas that need additional tests.

## Outcome

Produce an evidence-based coverage validation that:

- Runs the repository's actual UI unit test tooling instead of a generic React workflow
- Measures coverage for the React UI in `services/demo-ui/ui`
- Applies an 80% coverage gate and reports pass or fail explicitly
- Highlights the files and feature areas most responsible for missing the target
- Distinguishes configuration gaps from test coverage gaps

Default to a validation mindset: pass or fail first, then the main deficits and next actions.

## Primary Sources

Inspect these before concluding:

1. `services/demo-ui/ui/package.json`
2. `services/demo-ui/ui/vitest.config.ts`
3. `services/demo-ui/ui/src/test/`
4. `services/demo-ui/ui/src/**/*.test.tsx`
5. `services/demo-ui/ui/src/features/`
6. `services/demo-ui/ui/src/components/`
7. `services/demo-ui/ui/src/design-system/`

If the user asks for a narrower scope, still confirm how that scope fits into the main UI package and Vitest config.

## Procedure

### 1. Frame the Coverage Scope

Determine:

- Whether the user wants the full React UI checked or a specific feature path
- Whether the user wants check-only validation or validation plus remediation suggestions
- Whether the user wants repo-current behavior only, or also wants configuration changes that enforce the threshold in tooling

If unspecified, default to full UI coverage validation for `services/demo-ui/ui`.

### 2. Confirm the Repository Coverage Workflow

Use repository-defined tooling as the source of truth.

For this repo, verify:

- `package.json` uses Vitest for unit testing
- `vitest.config.ts` defines the UI test environment and setup files
- Coverage is collected with the package-local command:

```bash
cd services/demo-ui/ui && yarn vitest --coverage --run
```

If the repo later adds a dedicated coverage script such as `yarn test:coverage`, prefer that script over an ad hoc command.

### 3. Run the Coverage Check

Execute the coverage run from `services/demo-ui/ui` so Vitest resolves the local config and dependencies correctly.

Capture:

- Overall statement coverage
- Overall branch coverage
- Overall function coverage
- Overall line coverage
- File-level coverage hot spots with notably low values

Do not infer a pass from test count alone. Coverage must come from the actual coverage report.

### 4. Apply the 80% Gate

Validate the result against an 80% threshold.

By default, treat the gate as applying to the overall UI coverage result and state clearly which metric passed or failed.

Recommended default interpretation:

- Require at least 80% line coverage for a minimum pass decision
- If statements, branches, or functions are materially below 80%, call that out as a quality risk even if lines pass
- If the user asks for strict mode, require 80% across statements, branches, functions, and lines

If repository config does not enforce thresholds in `vitest.config.ts`, report that as a tooling gap rather than pretending the gate is automated.

### 5. Separate Meaningful Gaps From Noise

Inspect what is dragging coverage down.

Check for:

- Large TSX feature modules with low coverage
- Shared UI components that are untested or weakly tested
- API adapter code that lacks unit tests
- CSS files or non-executable assets appearing in the report and skewing totals
- Type-only modules or trivial entry files included in coverage without a clear reason

If non-executable assets are counted, note whether the reported failure is partly a coverage-configuration issue. Do not silently ignore them unless the user explicitly asks for a normalized report.

### 6. Produce the Validation Result

When reporting results, provide:

1. Overall pass or fail against the 80% target
2. The exact coverage figures used for the decision
3. Whether the threshold is enforced in tooling or only evaluated manually
4. The biggest low-coverage files or areas
5. Minimal next actions if the user asked for remediation

## Decision Points

- If the user asks for a strict gate, fail unless statements, branches, functions, and lines are all at least 80%.
- If the user asks for pragmatic validation, use line coverage as the primary gate and report the other metrics separately.
- If the repo has no enforced threshold in Vitest config, treat that as a standards gap and recommend adding it.
- If the user asks for feature-specific coverage, report that scope separately and do not confuse it with whole-app coverage.
- If CSS, stories, or other non-unit-test targets materially affect totals, call out the distortion and explain whether config exclusions should be tightened.

## Completion Checks

Validation is complete only if all are true:

- The UI coverage command was run from `services/demo-ui/ui`
- The response includes the coverage percentages used for the decision
- The response makes an explicit 80% pass or fail call
- The response states whether the threshold is actually enforced in config
- The main low-coverage files or feature areas are identified with evidence

## Repo-Specific Notes

Current repository evidence shows:

- UI tests run through Vitest in `services/demo-ui/ui`
- Coverage can be collected with `yarn vitest --coverage --run`
- `vitest.config.ts` currently defines environment and setup, but does not yet show an explicit coverage threshold gate

That means coverage validation may currently be a reporting workflow rather than an enforced CI-quality gate unless the config is extended.

## Example Prompts

- `Validate whether the React UI meets 80% unit test coverage.`
- `Run a strict 80% coverage gate for services/demo-ui/ui.`
- `Check which demo-ui React files are keeping us below 80% coverage.`
- `Validate UI coverage and suggest the smallest set of tests needed to reach 80%.`
- `Review the current Vitest setup and tell me whether the 80% UI coverage gate is actually enforced.`