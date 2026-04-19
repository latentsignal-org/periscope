# AGENTS.md

Instructions and project context for autonomous coding agents working in this
repository. This is the single source of truth — other agent instruction files
(e.g. `CLAUDE.md`) defer here.

## Scope

- Applies to all agent-driven work in this repo.
- If multiple instruction files exist, follow the most specific one for the
  files you are editing.

## Project Overview

agentsview is a local web viewer for AI agent sessions. It syncs session data
from disk into SQLite (with FTS5 full-text search), serves a Svelte 5 SPA via an
embedded Go HTTP server, and provides real-time updates via SSE. See
`internal/parser/types.go` for the full list of supported agents.

## Architecture

```
CLI (agentsview) → Config → DB (SQLite/FTS5)
                  ↓              ↓
              File Watcher → Sync Engine → Parsers (per agent)
                  ↓              ↓
              HTTP Server → REST API + SSE + Embedded SPA
                                 ↓
                           PG Push Sync → PostgreSQL (optional)
                                 ↑
              HTTP Server (pg serve) ← PostgreSQL
```

- **Server**: HTTP server with auto-port discovery (default 8080)
- **Storage**: SQLite with WAL mode, FTS5 for full-text search; optional
  PostgreSQL for multi-machine shared access
- **Sync**: File watcher + periodic sync (15min) for session directories
- **PG Sync**: On-demand push sync from SQLite to PostgreSQL via `pg push`
- **Frontend**: Svelte 5 SPA embedded in the Go binary at build time
- **Config**: `AGENT_VIEWER_DATA_DIR` plus per-agent directory overrides (see
  `EnvVar` on each entry in `internal/parser/types.go`) and CLI flags

## Project Structure

- `cmd/agentsview/` - Go server entrypoint
- `cmd/testfixture/` - Test data generator for E2E tests
- `internal/config/` - Config loading (TOML, JSON migration), flag registration
- `internal/db/` - SQLite operations (sessions, messages, search, analytics)
- `internal/postgres/` - PostgreSQL support: push sync, read-only store, schema,
  connection helpers
- `internal/parser/` - Per-agent session file parsers and content extraction
- `internal/server/` - HTTP handlers, SSE, middleware, search, export
- `internal/sync/` - Sync engine, file watcher, discovery, hashing
- `internal/timeutil/` - Time parsing utilities
- `internal/web/` - Embedded frontend (dist/ copied at build time)
- `frontend/` - Svelte 5 SPA (Vite, TypeScript)
- `scripts/` - Utility scripts (E2E server, changelog)

## Key Files

| Path                             | Purpose                                       |
| -------------------------------- | --------------------------------------------- |
| `cmd/agentsview/main.go`         | CLI entry point, server startup, file watcher |
| `cmd/agentsview/pg.go`           | pg command group (push, status, serve)        |
| `internal/server/server.go`      | HTTP router and handler setup                 |
| `internal/server/sessions.go`    | Session list/detail API handlers              |
| `internal/server/search.go`      | Full-text search API                          |
| `internal/server/events.go`      | SSE event streaming                           |
| `internal/db/db.go`              | Database open, migrations, schema             |
| `internal/db/sessions.go`        | Session CRUD queries                          |
| `internal/db/search.go`          | FTS5 search queries                           |
| `internal/sync/engine.go`        | Sync orchestration                            |
| `internal/parser/types.go`       | Agent registry (one `AgentDef` per agent)     |
| `internal/parser/*.go`           | Per-agent session parsers                     |
| `internal/postgres/connect.go`   | Connection setup, SSL checks, DSN helpers     |
| `internal/postgres/schema.go`    | PG DDL, schema management                     |
| `internal/postgres/push.go`      | Push logic, fingerprinting                    |
| `internal/postgres/sync.go`      | Push sync lifecycle                           |
| `internal/postgres/store.go`     | PostgreSQL read-only store                    |
| `internal/postgres/sessions.go`  | PG session queries (read side)                |
| `internal/postgres/messages.go`  | PG message queries, ILIKE search              |
| `internal/postgres/analytics.go` | PG analytics queries                          |
| `internal/postgres/time.go`      | Timestamp conversion helpers                  |
| `internal/config/config.go`      | Config loading, flag registration             |

## Development

```bash
make build          # Build binary with embedded frontend
make dev            # Run Go server in dev mode
make frontend       # Build frontend SPA only
make frontend-dev   # Run Vite dev server (use alongside make dev)
make install        # Build and install to ~/.local/bin or GOPATH
make install-hooks  # Install pre-commit git hooks
```

After making Go code changes, always run `go fmt ./...` and `go vet ./...`
before committing.

## Testing

**All new features and bug fixes must include unit tests.** Run tests before
committing:

```bash
make test       # Go tests (CGO_ENABLED=1 -tags fts5)
make test-short # Fast tests only (-short flag)
make e2e        # Playwright E2E tests
make lint       # golangci-lint
make vet        # go vet
```

### PostgreSQL Integration Tests

PG integration tests require a real PostgreSQL instance and the `pgtest` build
tag. The easiest way to run them is with docker-compose:

```bash
make test-postgres   # Starts PG container, runs tests, leaves container running
make postgres-down   # Stop the test container when done
```

Or manually with an existing PostgreSQL instance:

```bash
TEST_PG_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" \
  CGO_ENABLED=1 go test -tags "fts5,pgtest" ./internal/postgres/... -v
```

Tests create and drop the `agentsview` schema, so use a dedicated database or
one where schema changes are acceptable.

The CI pipeline runs these tests automatically via a GitHub Actions service
container (see `.github/workflows/ci.yml`, `integration` job).

### Test Guidelines

- Table-driven tests for Go code
- Use `testDB(t)` helper for database tests
- Frontend: colocated `*.test.ts` files, Playwright specs in `frontend/e2e/`
- All tests use `t.TempDir()` for temp directories

## Build Requirements

- **CGO_ENABLED=1** required (sqlite3 driver)
- **Build tag**: `-tags fts5` required for full-text search
- **Frontend**: Node.js + npm for Svelte build, embedded via
  `internal/web/dist/`

## Conventions

- Prefer stdlib over external dependencies.
- Tests should be fast and isolated.
- No emojis in code or output.
- **Markdown formatting**: Use `mdformat --wrap 80` to format Markdown files.
  Requires the `mdformat-tables` plugin
  (`uv tool install mdformat --with mdformat-tables`).

## Git Rules

1. Do not commit after every turn; wait for explicit user review and approval
   before committing changes.
1. Do not amend commits — always create new commits for fixes.
1. Do not change branches (create, switch, delete) without explicit user
   permission.
1. Do not push, pull, or rebase unless explicitly requested.

### Commit Expectations

- Keep commits focused and related to the requested task.
- Use conventional commit messages.
- Run relevant tests before committing when practical. If tests cannot be run,
  state that clearly in the handoff.

### Pull Requests

- PR descriptions: summary only, no test plans or checklists. Describe what the
  code does now, not how to test it.

## Safety

- Do not revert user-authored or unrelated local changes unless explicitly
  requested.
- Avoid destructive git commands unless explicitly requested.

## Data Safety

The SQLite database is a persistent archive. Never delete or recreate it to
handle data version changes. Schema changes use ALTER TABLE; parser changes
trigger a full resync (build fresh DB, sync files, copy orphaned sessions from
old DB, atomic swap). Existing session data must be preserved even when source
files no longer exist on disk.
