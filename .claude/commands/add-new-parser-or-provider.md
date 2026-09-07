---
name: add-new-parser-or-provider
description: Workflow command scaffold for add-new-parser-or-provider in periscope.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /add-new-parser-or-provider

Use this workflow when working on **add-new-parser-or-provider** in `periscope`.

## Goal

Adds support for a new agent, data source, or provider to the system, including schema detection, discovery, and integration with sync and test coverage.

## Common Files

- `internal/parser/<agent>_provider.go`
- `internal/parser/<agent>.go`
- `internal/parser/types.go`
- `internal/parser/provider.go`
- `internal/parser/provider_migration.go`
- `internal/parser/testdata/<agent>/*`
- `internal/parser/<agent>_test.go`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Implement parser/provider logic in `internal/parser` (e.g., `<agent>_provider.go`, `<agent>.go`).
- Add or update types in `internal/parser/types.go`.
- Update provider registration in `internal/parser/provider.go` and/or `provider_migration.go`.
- Add test fixtures and coverage in `internal/parser/testdata/<agent>/*` and `<agent>_test.go`.
- Integrate with sync engine (`internal/sync/engine.go`, `engine_test.go`, `integration_test.go`).
- Update documentation and format sources (e.g., `docs/internal/session-format-sources.md`, `docs/configuration.md`).

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.
