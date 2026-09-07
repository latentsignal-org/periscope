---
name: frontend-component-update-with-i18n-and-tests
description: Workflow command scaffold for frontend-component-update-with-i18n-and-tests in periscope.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /frontend-component-update-with-i18n-and-tests

Use this workflow when working on **frontend-component-update-with-i18n-and-tests** in `periscope`.

## Goal

Adds or updates a frontend Svelte component, updates internationalization files, and adds or updates corresponding tests.

## Common Files

- `frontend/src/lib/components/**/*.svelte`
- `frontend/src/lib/components/**/*.test.ts`
- `frontend/messages/*.json`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Edit or create .svelte component in frontend/src/lib/components/
- Edit or create corresponding .test.ts file for the component
- Update i18n message files (frontend/messages/*.json) for localization
- Update other related frontend files if needed

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.