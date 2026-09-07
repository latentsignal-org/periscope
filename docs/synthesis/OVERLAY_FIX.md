# Merge overlay correction (2026-07-29)

## Problem

History-preserving merge `32d3c281` overlaid too many paths from `merged`, regressing
the passing `ff8cd5b3` (purified+PR26) tree:

| Fault plane | Symptom | Root cause |
|-------------|---------|------------|
| Scripts | `build_script_test` failures | Merged `dev-backend-build.sh` / `e2e-server.sh` lacked `litellm-snapshot -restore` and used `cmd/agentsview` |
| Docs | `make docs-check` failed | Merged `README.md` used stale `agentsview.io/screenshots/` URLs |
| Frontend | 30 kit-ui findings | Merged context components replaced purified kit-ui migration |

## Options evaluated

| Option | Verdict |
|--------|---------|
| Cherry-pick synthetic-pass commits | **Reject** — based on `merged` (1178-file drift); pass2 duplicates purified; pass4 fights purified kit-ui |
| Rebase | **Reject** — rewrites merge parents; violates SHA-preservation policy |
| **Tree replay** | **Adopt** — `read-tree purified^{tree}` + minimal merged overlay |

## Minimal merged overlay (canonical)

Purified+PR26 (`ff8cd5b3`) is the **passing base**. Cherry-picking synthetic-pass
commits is **unnecessary** — those replay a `merged`-base tree and duplicate purified
content.

Overlay from `merged` only where purified lacks fork-specific value:

| Overlay path | Why |
|--------------|-----|
| `scripts/sync-upstream.sh` | Periscope fork sync logic (Layer 3) |
| `jetbrains-plugin/` | Fork JetBrains lifecycle (Layer 3) |
| `.claude/`, `.agents/`, `.codex/` | ECC bundle superset (PR #25 precedent) |

**Do not overlay** from `merged`:

- `scripts/dev-backend-build.sh`, `e2e-server.sh`, `desktop-dev.ps1` — purified has periscope paths + litellm restore
- `README.md`, `docs/` — purified has strict-link fixes from PR #26
- `frontend/src/lib/components/context/` — purified has kit-ui migration; merged context regresses kit-ui-check
- `internal/summarize/`, `internal/llm/` — purified already has Layer 2; `ff8cd5b3` passes without merged copies

## Fix

```bash
git read-tree --reset -u origin/cursor/agentsview-purified-onto-kenn-f559^{tree}
git checkout 32d3c281 -- jetbrains-plugin .claude .agents .codex scripts/sync-upstream.sh
```

Purified already contains PR #26 upstream stack, periscope script paths, and kit-ui context.
Merged overlay limited to ECC superset and fork `sync-upstream.sh` only.
