# ECC for Codex CLI

This supplements the root `AGENTS.md` with a repo-local ECC baseline.

## Repo Skill

- Repo-generated Codex skill: `.agents/skills/periscope/SKILL.md`
- Claude-facing companion skill: `.claude/skills/periscope/SKILL.md`
- Keep user-specific credentials and private MCPs in `~/.codex/config.toml`, not in this repo.

## MCP Baseline

Treat `.codex/config.toml` as the default ECC-safe baseline for work in this repository.
The generated baseline enables GitHub, Context7, Exa, Memory, Playwright, and Sequential Thinking.

## Multi-Agent Support

- Explorer: read-only evidence gathering
- Reviewer: correctness, security, and regression review
- Docs researcher: API and release-note verification

## Workflow Files

- `.claude/commands/database-migration.md`
- `.claude/commands/feature-development.md`
- `.claude/commands/refactoring.md`
- `.claude/commands/add-or-update-backend-feature-with-tests.md`
- `.claude/commands/frontend-component-update-with-i18n-and-tests.md`
- `.claude/commands/documentation-and-blueprint-update.md`
- `.claude/commands/add-or-update-database-feature.md`
- `.claude/commands/add-new-parser-or-provider.md`
- `.claude/commands/feature-development-with-tests-and-docs.md`

- `.claude/commands/context-visualizer-ui-feature-workflow.md`
- `.claude/commands/database-schema-and-storage-migration.md`
Use these workflow files as reusable task scaffolds when the detected repository workflows recur.
Pair them with the synthesized instincts in
`.claude/homunculus/instincts/inherited/periscope-instincts.yaml` and the workflow
sections in the repo skill files above.
