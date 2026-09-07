# periscope Development Patterns

> Synthesized from ECC repository-analysis runs (PR #16 harmonization + PR #18
> additive replay) and verified against the repository's documented architecture
> and test layout.

## Overview

Periscope is a Go service and CLI backed by SQLite/FTS5, with a Svelte 5 and
TypeScript frontend embedded in the Go binary. It also includes a Tauri desktop
wrapper and optional PostgreSQL synchronization. Apply conventions within the
language and subsystem being changed rather than treating the repository as a
TypeScript-only project.

## Coding Conventions

### File and Symbol Naming

- Follow the surrounding subsystem's established naming.
- Frontend TypeScript files use **kebab-case** where applicable.
- TypeScript functions use **camelCase**, classes use **PascalCase**, and
  constants use **SCREAMING_SNAKE_CASE**.
- Go files, packages, and symbols follow idiomatic Go conventions.

### TypeScript Imports and Exports

- Prefer relative imports for project-local frontend modules.
- Prefer named exports where the surrounding module follows that pattern.

```typescript
import { fetchData } from "./data-fetcher";

export function getUserProfile(id: string) {
  return fetchData(id);
}
```

### Commit Messages

- Use Conventional Commits.
- Choose the prefix that describes the change; observed prefixes include
  `build`, `chore`, `docs`, `feat`, `fix`, and scoped variants such as
  `fix(frontend)` and `fix(desktop)`.
- Keep the subject concise and include a scope when it adds useful context.

```text
build(deps): update frontend dependencies
fix(desktop): honor PERISCOPE_VERSION override
chore(git): sync attribution guard scripts
docs: align layer-2 synthesis analysis tip SHA
fix(frontend): remove stale ActivityMinimap and pass svelte-check
feat(parser): add session discovery for a new agent
```

## Workflows

### Code Contribution

**Trigger:** When adding or updating code

**Guide:** `/contribute`

1. Read `AGENTS.md` and the conventions nearest to the files being changed.
2. Follow the naming, import, and export patterns of that subsystem.
3. Add or update tests for new features and bug fixes.
4. Run the targeted checks, then the broader affected suite.
5. Use a Conventional Commit prefix that matches the change.

### Dependency Update

**Trigger:** When updating a package or language ecosystem

**Guide:** `/update-dependencies`

**Instinct pair (keep both):**

- `periscope-workflow-dependency-update` — numbered workflow steps; trigger:
  "when doing dependency update".
- `periscope-instinct-dependency-update` — concise action summary; trigger:
  "When updating dependencies for a package or language ecosystem".

1. Update the relevant manifest, such as `frontend/package.json`,
   `desktop/package.json`, `desktop/src-tauri/Cargo.toml`, or `go.mod`.
2. Regenerate the corresponding lockfile or module metadata with the native
   package manager.
3. Review both manifest and generated dependency changes for unintended drift.
4. Run the checks for each affected subsystem.
5. Commit the manifest and lockfile together with a `build(deps)` subject.

### Integration Analysis Doc

**Trigger:** When updating layer-2 integrative synthesis verification or tip SHA

**Guide:** `/update-integration-analysis`

**Instinct:** `periscope-workflow-update-integration-analysis-doc` — numbered
workflow steps; trigger: "when doing update integration analysis doc".

1. Edit `docs/INTEGRATION-SYNTHESIS-LAYER2-ANALYSIS.md` for verification gates,
   branch references, or tip SHA alignment.
2. Use a `docs:` Conventional Commit subject that names the alignment work.
3. Keep the analysis consistent with the current `merged` integration line.

### Add New Agent Integration

**Trigger:** When adding support for a new AI agent parser/discovery path

**Guide:** `/add-new-agent-integration` (instinct:
`periscope-workflow-add-new-agent-integration`)

1. Implement a parser in `internal/parser/{agent}.go` with
   `internal/parser/{agent}_test.go`.
2. Add discovery logic in `internal/parser/discovery.go`.
3. Update normalization in `internal/parser/types.go` and `taxonomy.go`.
4. Wire the agent into `internal/sync/engine.go` and integration tests.
5. Update agent directory/env configuration in `internal/config/config.go`.
6. Update frontend agent metadata in `frontend/src/lib/utils/agents.ts` and
   related UI filters.
7. Update documentation (`README.md`, `AGENTS.md`) when user-facing behavior
   changes.

### Add New Feature (Full Stack)

**Trigger:** When implementing a user-facing feature spanning backend and UI

**Guide:** `/add-new-feature-full-stack` (instinct:
`periscope-workflow-add-new-feature-full-stack`)

1. Design and implement DB schema changes in `internal/db/schema.sql` and
   `internal/db/*.go`.
2. Add or update backend logic in `internal/server/` and `internal/sync/`.
3. Expose API endpoints in `internal/server/`.
4. Implement Svelte components/stores under `frontend/src/lib/`.
5. Update frontend types under `frontend/src/lib/api/types/`.
6. Add or update unit and e2e tests for the touched subsystems.
7. Update docs when operator or contributor behavior changes.

### Backend Bugfix with Test

**Trigger:** When fixing backend logic

**Guide:** `/backend-bugfix-with-test` (instinct:
`periscope-instinct-backend-bugfix-with-test`)

1. Fix the bug in the relevant Go package under `internal/`.
2. Add or update a regression test in the colocated `*_test.go` file.
3. Run targeted and affected package tests before handoff.

### Frontend Bugfix with Test

**Trigger:** When fixing frontend UI or store logic

**Guide:** `/frontend-bugfix-with-test` (instinct:
`periscope-instinct-frontend-bugfix-with-test`)

1. Fix the bug in the relevant Svelte/TypeScript module.
2. Add or update a Vitest or Playwright test as appropriate.
3. Run `cd frontend && npm test` and any affected e2e journeys.

### Refactor and Test Split

**Trigger:** When test files become large or mixed across concerns

**Guide:** `/refactoring` (command scaffold:
`.claude/commands/refactoring.md`)

1. Identify oversized or mixed test files.
2. Split into per-agent or per-feature test files.
3. Move or create fixtures under `internal/parser/testdata/` when needed.
4. Verify all tests still pass.

### Database Schema Migration and Sync

**Trigger:** When changing or extending the SQLite schema

**Guide:** `/database-migration` (command scaffold:
`.claude/commands/database-migration.md`)

1. Edit `internal/db/schema.sql` and related Go structs/queries.
2. Update sync logic in `internal/sync/` when ingestion behavior changes.
3. Add or update tests in `internal/db/*_test.go` and sync integration tests.
4. Handle resync/migration semantics without destroying archived session data.
5. Update frontend types when API payloads change.

### Feature Development Scaffold

**Trigger:** When starting a multi-file feature with unclear boundaries

**Guide:** `/feature-development` (command scaffold:
`.claude/commands/feature-development.md`)

Use the scaffold to sequence discovery, smallest coherent change, verification,
and handoff notes before expanding scope.

### Testing

**Trigger:** When verifying correctness

**Guide:** `/test`

1. Add tests in the location used by the affected subsystem.
2. Run the smallest relevant test target while iterating.
3. Run the broader affected suite before handoff.
4. For Go changes, run `go fmt ./...` and `go vet ./...`.

## Testing Patterns

- Go unit tests are colocated with packages as `*_test.go`; table-driven tests
  are preferred.
- Frontend unit tests are colocated as `*.test.ts` and run with Vitest.
- Browser journeys live in `frontend/e2e/` and run with Playwright.
- PostgreSQL integration tests use the `pgtest` build tag and a dedicated test
  database.
- Use `t.TempDir()` for isolated Go test data.
- Use `internal/parser/testdata/` for parser fixtures.

## Verified Commands

| Command | Purpose |
| --- | --- |
| `make test-short` | Run fast Go tests |
| `make test` | Run the full Go test suite |
| `cd frontend && npm test` | Run frontend Vitest tests |
| `make e2e` | Run Playwright end-to-end tests |
| `make vet` | Run Go static checks |
| `make lint` | Run configured Go linters |

## Workflow Command Scaffolds

| Command | Scaffold path |
| --- | --- |
| `/database-migration` | `.claude/commands/database-migration.md` |
| `/feature-development` | `.claude/commands/feature-development.md` |
| `/refactoring` | `.claude/commands/refactoring.md` |
| `/add-or-update-backend-feature-with-tests` | `.claude/commands/add-or-update-backend-feature-with-tests.md` |
| `/frontend-component-update-with-i18n-and-tests` | `.claude/commands/frontend-component-update-with-i18n-and-tests.md` |
| `/documentation-and-blueprint-update` | `.claude/commands/documentation-and-blueprint-update.md` |
| `/add-or-update-database-feature` | `.claude/commands/add-or-update-database-feature.md` |
| `/add-new-parser-or-provider` | `.claude/commands/add-new-parser-or-provider.md` |
| `/feature-development-with-tests-and-docs` | `.claude/commands/feature-development-with-tests-and-docs.md` |
