---
name: add-or-update-database-feature
description: Workflow command scaffold for add-or-update-database-feature in periscope.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /add-or-update-database-feature

Use this workflow when working on **add-or-update-database-feature** in `periscope`.

## Goal

Implements a new database-backed feature or updates existing database logic, including schema changes, new tables, triggers, and related Go logic.

## Common Files

- `internal/db/schema.sql`
- `internal/db/*.go`
- `internal/db/*_test.go`
- `cmd/agentsview/*_test.go`
- `docs/superpowers/specs/*.md`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Edit or add SQL schema files (e.g., internal/db/schema.sql).
- Update Go code for database access and logic (e.g., internal/db/*.go).
- Add or update trigger DDL and migration logic in Go (e.g., internal/db/db.go).
- Write or update tests for new/changed database logic (e.g., internal/db/*_test.go, cmd/agentsview/*_test.go).
- Update or add documentation/specs if the feature is significant (e.g., docs/superpowers/specs/...).

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.