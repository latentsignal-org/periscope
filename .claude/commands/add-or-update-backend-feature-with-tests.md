---
name: add-or-update-backend-feature-with-tests
description: Workflow command scaffold for add-or-update-backend-feature-with-tests in periscope.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /add-or-update-backend-feature-with-tests

Use this workflow when working on **add-or-update-backend-feature-with-tests** in `periscope`.

## Goal

Implements or refactors backend features, always accompanied by corresponding test files. Typically involves Go source files and their _test.go counterparts in cmd/periscope/ or internal/ directories.

## Common Files

- `cmd/periscope/*.go`
- `cmd/periscope/*_test.go`
- `internal/**/*.go`
- `internal/**/*_test.go`
- `Makefile`
- `go.mod`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Edit or create Go implementation file in cmd/periscope/ or internal/ directories
- Edit or create corresponding _test.go file for tests
- Update Makefile or go.mod/go.sum if dependencies or build steps change
- Run tests to verify changes

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.