---
name: context-visualizer-ui-feature-workflow
description: Workflow command scaffold for context-visualizer-ui-feature-workflow in periscope.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /context-visualizer-ui-feature-workflow

Use this workflow when working on **context-visualizer-ui-feature-workflow** in `periscope`.

## Goal

Implements or migrates context visualizer UI features, including bug fixes, refactors, and kit-ui migrations.

## Common Files

- `frontend/src/lib/components/context/*.svelte`
- `frontend/src/lib/components/context/*.test.ts`
- `frontend/messages/*.json`
- `frontend/scripts/generate-api-client.mjs`
- `docs/context-session-visualizer-mvp-plan.md`
- `docs/context-session-visualizer-roadmap.md`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Edit or create Svelte components under frontend/src/lib/components/context/
- Update or add related test files (*.test.ts) for those components
- Update frontend/messages/*.json for i18n if needed
- Update or fix docs/context-session-visualizer-*.md and docs/context-visualizer-ui-recommendation.md as needed
- Update frontend/scripts/generate-api-client.mjs if API changes

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.