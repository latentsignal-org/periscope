# AgentsView → Periscope rename catalogue

**Status:** living operator guide for upstream merges
**Base:** `candidate@2b6e5128` (2026-07-28) — refreshed from prior snapshot `origin/merged@6cf2f38f`
**Machine index:** [`agentsview-rename-index.json`](./agentsview-rename-index.json) (844 files, 5165 matches)

> **Historical snapshot:** the milestone index at `origin/merged@6cf2f38f` tracked 96
> files / 514 matches before the synthesis replay landed. Counts below reflect the
> current candidate tree; category notes and `session_fixes_2026_07_28` are preserved.

## Why this exists

Periscope is a **dual-pedigree fork**:

| Branch | Upstream | Role |
| --- | --- | --- |
| `main` | `latentsignal-org/periscope` | Layer 2 mirror — do not agent-PR here |
| `agentsview` | `kenn-io/agentsview` (was `wesm/agentsview`) | Layer 1 mirror — **keep branch name** |
| `merged` | integrative build line | Layer 3 — **rename upstream imports here** |

Every sync from `agentsview` reintroduces `agentsview` strings. This guide is the
canonical merge checklist so agents do not treat the mirror branch as the product
name or miss CI/desktop rename debt.

See also: [`docs/ARCHITECTURE.md`](../ARCHITECTURE.md) (matryoshka model),
[`scripts/sync-upstream.sh`](https://github.com/diazMelgarejo/periscope/blob/merged/scripts/sync-upstream.sh).

## Decision tree (use on every upstream merge)

```text
Incoming hunk from agentsview upstream
│
├─ Git branch / remote named "agentsview"?
│  └─ KEEP — mirror branch name is intentional
│
├─ Test fixture path or session JSON cwd like /Users/.../agentsview?
│  └─ KEEP — simulates real user data; not product branding
│
├─ Historical doc in docs/superpowers/* describing the rename?
│  └─ KEEP — archive context
│
├─ Product binary, module, cmd path, CI artifact, sidecar filename?
│  └─ RENAME → periscope (see tables below)
│
├─ Env var AGENTSVIEW_* or AGENT_VIEWER_DATA_DIR?
│  └─ COMPAT — accept upstream name in merge, ensure PERISCOPE_* /
│     legacy fallbacks still read old names (see config.go)
│
├─ Frontend localStorage key agentsview-*?
│  └─ MIGRATE — add periscope-* key; read legacy on load (do not
│     silently drop user prefs)
│
└─ Unsure?
   └─ Grep this catalogue + agentsview-rename-index.json; default to
      product-facing rename, data-fixture keep
```

## Rename map (high priority)

Apply these on **`merged`** (and PR branches targeting `merged`), not on the
`agentsview` mirror branch.

| Upstream (agentsview) | Periscope (merged) | Notes |
| --- | --- | --- |
| `cmd/agentsview/` | `cmd/periscope/` | Done on merged |
| binary `agentsview` | `periscope` | Makefile, CI, installers |
| module `github.com/wesm/agentsview` | `go.kenn.io/agentsview` (unchanged from upstream) | Reverted 2026-08-03; see note below |
| `AGENTSVIEW_DATA_DIR` | `PERISCOPE_DATA_DIR` | Keep legacy read in `config.go` |
| `AGENT_VIEWER_DATA_DIR` | `PERISCOPE_DATA_DIR` | Legacy alias still supported |
| `AGENTSVIEW_VERSION` | `PERISCOPE_VERSION` | `prepare-sidecar.sh` uses PERISCOPE_* |
| `AGENTSVIEW_TARGET_TRIPLE` | `PERISCOPE_TARGET_TRIPLE` | Prefer in workflows; release.yml already uses PERISCOPE_* |
| sidecar `binaries/agentsview-<triple>` | `binaries/periscope-<triple>` | `prepare-sidecar.sh` emits periscope-* |
| Tauri `sidecar("agentsview")` | `sidecar("periscope")` | Must match `tauri.conf.json` `externalBin` |
| artifact `agentsview-desktop-*` | `periscope-desktop-*` | `desktop-release.yml` done; `desktop-artifacts.yml` pending on merged |
| User-Agent `agentsview` | `periscope` | `internal/server/export.go` |
| UI title `AgentsView` | `Periscope` | `internal/web/fallback/index.html`, docs |
| PG schema default (product) | `periscope` | CLI default; many pg tests still say `agentsview` |

**Module path note (2026-08-03):** a Cursor Cloud Agent inferred the module
rename beyond its authorized scope — only `cmd/agentsview/` → `cmd/periscope/`
was authorized. Module identity always tracks upstream and is never forked.

## Category index (844 files)

| Category | Files | Action |
| --- | ---: | --- |
| `rename_residual` | 704 | Comments, tests, docs — review per hunk |
| `rename_pg_tests` | 82 | PG integration test schema names |
| `rename_localstorage` | 16 | localStorage keys — migrate with legacy read |
| `rename_build_release` | 14 | Makefile, install scripts, wheels, e2e |
| `rename_desktop` | 11 | Sidecar, Tauri spawn, desktop tests, README |
| `keep_test_fixtures` | 6 | Parser/sync integration fixtures |
| `keep_historical` | 4 | `docs/superpowers/*` plans |
| `rename_docs_operator` | 3 | AGENTS.md, README, desktop-release-setup |
| `rename_product_strings` | 2 | Export User-Agent / HTML footer |
| `compat_env` | 1 | `internal/config/config.go` — keep legacy env reads |
| `keep_upstream_sample` | 1 | `support/launchd/io.agentsview.pg-serve.plist` |

Full per-file rows: [`agentsview-rename-index.json`](./agentsview-rename-index.json).

## Hot spots (largest remaining debt)

| Matches | Path | Merge action |
| ---: | --- | --- |
| 216 | `docs/commands.md` | Operator doc branding |
| 129 | `docs/changelog.md` | Historical changelog references — review |
| 125 | `cmd/agentsview/recall_test.go` | Residual `cmd/agentsview/` test paths |
| 125 | `desktop/src-tauri/src/lib.rs` | `sidecar("agentsview")` → `"periscope"` |
| 118 | `README.md` | Product naming |
| 99 | `internal/db/recall_test.go` | Test/schema strings — review |
| 80 | `docs/configuration.md` | Operator doc branding |
| 71 | `docs/pg-sync.md` | Operator doc branding |
| 71 | `internal/db/recall_import_test.go` | Test strings — review |
| 70 | `scripts/build_wheels_test.py` | Wheel/archive regex → `periscope_*` |
| 69 | `docs/token-usage.md` | Operator doc branding |
| 60 | `internal/server/recall_test.go` | Test strings — review |
| 59 | `docs/session-api.md` | Operator doc branding |
| 51 | `cmd/agentsview/pg_service_test.go` | Residual `cmd/agentsview/` paths |
| 6 | `AGENTS.md` | Stale `cmd/agentsview/` paths in examples |

## CI / desktop checklist after each agentsview sync

1. `bash desktop/scripts/test-prepare-sidecar.sh`
2. `bash desktop/scripts/test-desktop-workflows.sh`
3. `bash desktop/scripts/test-startup-ui.sh`
4. Confirm smoke step uses `periscope-${target_triple}` not `agentsview-*`
5. Confirm `version` output grep uses `periscope dev` not `agentsview dev`
6. Run Desktop Artifacts workflow on PR branch before merge to `merged`

## Session fixes (2026-07-28)

Recorded on PR **#11** (`cursor/deps-npm-onto-merged-f559`); cherry-pick or
reapply onto `merged` if not yet integrated:

| File | Fix |
| --- | --- |
| `desktop/scripts/test-desktop-workflows.sh` | Assert `periscope-desktop-linux-arm64` in release workflow |
| `.github/workflows/desktop-artifacts.yml` | Sidecar path `periscope-${target_triple}`; version grep `periscope dev` |
| `desktop/src-tauri/.gitignore` | Ignore `periscope-*` sidecar binaries |

**Root cause:** Layer 3 rename updated `prepare-sidecar.sh` and
`desktop-release.yml` but CI smoke tests and `desktop-artifacts.yml` still
expected `agentsview-*` paths.

## Regenerate this catalogue

```bash
cd /path/to/periscope
git fetch origin merged
BASE=origin/merged

# Human scan
rg -l 'agentsview|AgentsView|AGENTSVIEW' \
  --glob '!**/node_modules/**' --glob '!**/Cargo.lock' | sort

# Count per file
rg -c 'agentsview|AgentsView|AGENTSVIEW' \
  --glob '!**/node_modules/**' --glob '!**/Cargo.lock' | sort -t: -k2 -nr

# Rebuild JSON (update categorizer in PT memory doc or re-run agent catalogue pass)
# See agentsview-rename-index.json → "regenerate" for the ripgrep command.

## Merge workflow reminder

```bash
# 1. Sync mirror (on agentsview branch only)
git fetch https://github.com/kenn-io/agentsview.git main
git push --force-with-lease=refs/heads/agentsview:<old-sha> \
  origin FETCH_HEAD:refs/heads/agentsview

# 2. Integrate into merged (PR base)
git checkout -B cursor/sync-agentsview-onto-merged-f559 origin/merged
git merge origin/agentsview
# Resolve using this catalogue — rename product, keep fixtures

# 3. Run rename checklist + desktop smoke scripts + CI
```

**Never** open agent PRs against `main` or `agentsview` — only `merged`.
