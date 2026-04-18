---
name: typescript-react-lint-validation
description: "Validate TypeScript and React code against repository linting standards. Use for lint compliance checks, pre-PR quality gates, and finding standards violations in TS/TSX code."
argument-hint: "Optional: target path, strictness level, and whether to include autofix suggestions"
---

# TypeScript React Lint Validation

Use this skill when the user asks to validate TypeScript or React code quality, run lint checks, identify standards violations, or verify TS/TSX readiness before merge.

## Outcome

Produce an evidence-based lint validation report that:

- Identifies all TS/TSX scope in the workspace or requested path
- Runs repository-standard lint checks instead of generic assumptions
- Reports violations with clear file-level evidence and severity
- Distinguishes blocking errors from advisory warnings
- Confirms pass/fail status with explicit completion criteria

## Primary Sources

Inspect these before running conclusions:

1. `services/demo-ui/ui/package.json`
2. `services/demo-ui/ui/eslint.config.js`
3. `services/demo-ui/ui/tsconfig.json`
4. `services/demo-ui/ui/tsconfig.app.json`
5. `services/demo-ui/ui/tsconfig.node.json`

If the user asks for full-workspace validation, also detect any additional TS/TSX projects with their own lint configuration.

## Procedure

### 1. Frame the Validation Scope

Determine:

- Target scope: specific files, a directory, or full workspace
- Validation mode: check-only or check-plus-fix suggestions
- Reporting mode: concise summary or full violation list

If unspecified, default to repository-wide TS/TSX lint validation.

### 2. Discover Applicable Lint Standards

For each target TypeScript/React project:

- Locate `package.json` lint scripts
- Locate ESLint config (`eslint.config.*` or `.eslintrc*`)
- Locate TypeScript config files (`tsconfig*.json`)
- Confirm framework stack (React and TypeScript versions)

Use the project-defined lint command as the source of truth (for this repo, `yarn lint` in `services/demo-ui/ui`).

### 3. Run Validation Using Project Tooling

Execute linting from the correct package directory so local config and ignores are honored.

Recommended command pattern:

```bash
cd services/demo-ui/ui && yarn lint
```

If the user asks for file-scoped validation, run ESLint against the requested paths while preserving project config.

### 4. Classify and Analyze Findings

For each violation, capture:

- File path and rule identifier
- Error vs warning level
- Whether it is TypeScript-specific, React-hooks specific, or general ESLint
- Whether autofix is likely safe

Do not claim fixes were applied unless they were actually applied and revalidated.

### 5. Apply Optional Fix Loop (Only If Requested)

If user requests remediation:

1. Apply minimal safe fixes first
2. Re-run lint
3. Report remaining issues that require manual changes

Keep changes small and avoid unrelated refactoring.

### 6. Produce Validation Report

When reporting results, provide:

1. Overall status: pass or fail
2. Blocking errors summary
3. Warning summary
4. Top actionable fixes (if requested)
5. Re-run command for verification

## Decision Points

- If multiple TS projects exist, validate each independently and report per-project status.
- If no lint config exists for a target path, report that as a standards gap.
- If lint tool fails to run (missing deps/tooling), report environment blocker and provide exact install command.
- If user requests strict mode, treat warnings as fail criteria in the report.
- If user requests CI parity, use the same command used by repository scripts rather than custom flags.

## Completion Checks

Validation is complete only if all are true:

- Lint command was run in every requested TS/TSX scope
- Results were captured with error/warning counts
- Report explicitly states pass/fail decision
- Any claimed fix was followed by a re-run of lint
- No assumptions were made without command evidence

## Example Prompts

- `Validate all TypeScript and React linting in this workspace.`
- `Run strict lint validation for demo-ui and fail on warnings.`
- `Check only services/demo-ui/ui/src/features/cart for TS/React lint issues.`
- `Lint the UI project and suggest minimal safe fixes.`
