---
name: feature-development-with-tests-and-docs
description: Workflow command scaffold for feature-development-with-tests-and-docs in periscope.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /feature-development-with-tests-and-docs

Use this workflow when working on **feature-development-with-tests-and-docs** in `periscope`.

## Goal

Implements a new feature or significant enhancement, accompanied by tests and documentation/specs.

## Common Files

- `internal/*/*.go`
- `internal/*/*_test.go`
- `cmd/agentsview/*_test.go`
- `docs/superpowers/specs/*.md`
- `docs/configuration.md`
- `frontend/src/lib/utils/agents.ts`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Implement feature in Go (e.g., internal/<area>/*.go).
- Write or update tests (e.g., internal/<area>/*_test.go, cmd/agentsview/*_test.go).
- Update or add documentation/specs (e.g., docs/superpowers/specs/*.md, docs/configuration.md).
- Update frontend if feature is user-facing (e.g., frontend/src/lib/utils/agents.ts).

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.