---
name: database-schema-and-storage-migration
description: Workflow command scaffold for database-schema-and-storage-migration in periscope.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /database-schema-and-storage-migration

Use this workflow when working on **database-schema-and-storage-migration** in `periscope`.

## Goal

Adds or modifies database tables, triggers, or storage logic, including migrations and related Go code/tests.

## Common Files

- `internal/db/schema.sql`
- `internal/db/*.go`
- `internal/db/*_test.go`
- `internal/artifact/*.go`
- `internal/artifact/*_test.go`
- `internal/parser/*.go`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Edit internal/db/schema.sql or equivalent Go migration logic
- Update or add Go files in internal/db/ (e.g., artifact_publication.go, sessions.go, usage_events.go)
- Update or add related test files in internal/db/
- Edit or add storage logic in internal/artifact/ or internal/parser/ as needed
- Update docs/specs or plans if schema or storage design changes

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.