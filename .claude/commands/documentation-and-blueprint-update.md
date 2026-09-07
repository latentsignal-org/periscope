---
name: documentation-and-blueprint-update
description: Workflow command scaffold for documentation-and-blueprint-update in periscope.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /documentation-and-blueprint-update

Use this workflow when working on **documentation-and-blueprint-update** in `periscope`.

## Goal

Updates or adds documentation, blueprints, or architecture plans, often in docs/ or scripts/guides/ directories.

## Common Files

- `docs/**/*.md`
- `docs/guides/**/*.md`
- `docs/superpowers/**/*.md`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Edit or create markdown files in docs/ or docs/guides/
- Optionally update related scripts or blueprints
- Commit with a message referencing documentation or blueprint

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.