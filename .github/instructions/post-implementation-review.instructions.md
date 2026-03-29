---
description: "Use when implementation work is finished; enforce a mandatory review pass before final response or handoff."
name: "Post-Implementation Review"
applyTo: "**"
---

# Post-Implementation Review

After implementation is done, always perform a review pass before declaring completion.

## Mandatory Review Checklist

- Re-check changed files for logic bugs, regressions, and mismatches with requirements.
- Validate alignment with project architecture and analysis documents in the relevant `docs` folders.
- always ensure you have found relevant docs. If you haven't, ask for help or clarification.
- Confirm tests are added or updated for new behavior and bug fixes.
- Run or report relevant validation steps (tests, lint, and static checks) for touched areas.
- For changes in Go services, run and report: all tests, `golangci-lint`, and `go vet`.
- For changes in `services/demo-ui/ui`, run and report: `yarn test` and lint.
- Call out known risks, assumptions, or gaps explicitly in the final response.
- Test coverage must be 90% or higher for changed files, unless explicitly approved otherwise.

## Response Behavior

- Do not present implementation as complete until the review checklist is executed.
- If validation cannot be run, explicitly explain what was not run and why.
- Summarize review findings first, then provide a short change summary.