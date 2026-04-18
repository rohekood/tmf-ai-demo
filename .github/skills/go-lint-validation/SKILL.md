---
name: go-lint-validation
description: "Validate Go code against repository linting standards. Use for golangci-lint compliance checks, pre-PR quality gates, and finding standards violations across Go modules."
argument-hint: "Optional: target path/module, strictness level, and whether to include go vet"
---

# Go Lint Validation

Use this skill when the user asks to validate Go code quality, run lint checks, identify standards violations, or verify Go readiness before merge.

## Outcome

Produce an evidence-based lint validation report that:

- Identifies all Go module scope in the workspace or requested target path
- Runs repository-standard lint checks instead of generic assumptions
- Reports violations with clear file-level evidence and severity
- Distinguishes blocking errors from advisory warnings
- Confirms pass/fail status with explicit completion criteria

## Primary Sources

Inspect these before concluding:

1. `README.md`
2. `go.work`
3. `.github/instructions/general-development.instructions.md`
4. `.github/instructions/post-implementation-review.instructions.md`
5. All relevant `go.mod` files in target scope
6. Any `.golangci*` config files in target scope

If the user asks for full-workspace validation, enumerate all modules from `go.work` and include standalone modules that are outside `go.work` only if explicitly requested.

## Procedure

### 1. Frame the Validation Scope

Determine:

- Target scope: specific files, package path, module, service, or full workspace
- Validation mode: check-only or check-plus-fix suggestions
- Whether `go vet` must be included (default: include)
- Reporting mode: concise summary or full violation list

If unspecified, default to repository-wide Go lint validation for all modules in `go.work`.

### 2. Discover Applicable Lint Standards

For each target Go module:

- Confirm module root via `go.mod`
- Confirm workspace/module inclusion via `go.work`
- Detect local lint configuration (`.golangci.yml`, `.golangci.yaml`, `.golangci.toml`, `.golangci.json`)
- Determine repository-preferred linter invocation

Use the project-defined command as source of truth. In this repo, prefer:

- `$(go env GOPATH)/bin/golangci-lint run` when that path convention is required
- Otherwise `golangci-lint run` if binary is on `PATH`

### 3. Run Validation Using Project Tooling

Execute linting from each module root so local module context and excludes are honored.

Recommended command pattern:

```bash
cd <module-root> && $(go env GOPATH)/bin/golangci-lint run
```

Fallback pattern:

```bash
cd <module-root> && golangci-lint run
```

If `go vet` is in scope, run it after lint for the same module:

```bash
cd <module-root> && go vet ./...
```

For file-scoped requests, run module-level lint and report findings filtered to the requested files; do not claim full-module pass from partial-file execution.

### 4. Classify and Analyze Findings

For each finding, capture:

- File path and line reference
- Linter or analyzer rule (for example `govet`, `staticcheck`, `ineffassign`, `errcheck`)
- Error severity and whether it is blocking
- Whether fix is likely safe and mechanical

Do not claim fixes were applied unless they were actually applied and the checks were rerun.

### 5. Apply Optional Fix Loop (Only If Requested)

If the user requests remediation:

1. Apply minimal safe fixes first
2. Rerun lint (and `go vet` if in scope)
3. Report remaining issues requiring manual changes

Keep edits small and avoid unrelated refactoring.

### 6. Produce Validation Report

When reporting results, provide:

1. Overall status: pass or fail
2. Per-module status summary
3. Blocking errors summary
4. Warning or advisory summary
5. `go vet` summary (if run)
6. Rerun commands for verification

## Decision Points

- If multiple modules exist, validate each independently and report per-module status.
- If lint config differs per module, respect module-local config and do not normalize flags globally.
- If lint tool is missing, report environment blocker and provide the exact install command used in this environment.
- If user requests strict mode, treat all warnings and advisories as fail criteria in the final report.
- If user requests CI parity, run the same commands and flags used by repo scripts/config rather than custom shortcuts.

## Completion Checks

Validation is complete only if all are true:

- Lint command ran in every requested Go module scope
- `go vet` ran when requested or by default policy
- Results were captured with per-module pass/fail evidence
- Report explicitly states final pass/fail decision
- Any claimed fix was followed by a rerun of relevant checks
- No assumptions were made without command evidence

## Example Prompts

- `Validate all Go linting in this workspace.`
- `Run strict Go lint validation for services/qualification.`
- `Check Go lint issues in services/shopping-cart and include go vet.`
- `Lint all modules from go.work and suggest minimal safe fixes.`
