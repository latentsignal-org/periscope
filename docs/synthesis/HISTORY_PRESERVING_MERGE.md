# History-preserving synthesis policy

**Branch:** `cursor/agentsview-plus-periscope-f559`  
**PR:** [#29](https://github.com/diazMelgarejo/periscope/pull/29)

## Policy (2026-07-29)

Per [PERISCOPE_MODERNIZATION_PURIFIED_INTEGRATION](https://github.com/diazMelgarejo/Perpetua-Tools/blob/main/.agent/memory/working/PERISCOPE_MODERNIZATION_PURIFIED_INTEGRATION_2026-07-29.md) § "never synthesize SHAs":

1. **Prefer `git merge` with two real parents** — preserves all commits from `merged` and `purified+PR26` with original SHAs.
2. **Minimize synthetic commits** — one merge commit + docs; no path-scoped replay stacks on the integration branch.
3. **Tree resolution** uses matryoshka layers, not wholesale `-X ours/theirs`.
4. **Synthetic pass work** archived on `cursor/agentsview-plus-periscope-synthetic-pass-f559` (reference only).

## Merge technique

```bash
git merge origin/cursor/agentsview-purified-onto-kenn-f559 --no-commit --no-ff
git read-tree --reset -u $(git rev-parse origin/cursor/agentsview-purified-onto-kenn-f559^{tree})
git checkout HEAD -- <PERISCOPE_OWNED paths from ARCHITECTURE.md Layer 2+3>
git commit  # records both parent SHAs
```

## Layer overlay (merged paths restored onto purified base tree)

Purified+PR26 (`ff8cd5b3`) is the passing base tree. Overlay from `merged` only where
purified lacks fork-specific value — see `docs/synthesis/OVERLAY_FIX.md`.

| Layer | Paths from `merged` |
|-------|----------------------|
| 3 | `scripts/sync-upstream.sh`, `jetbrains-plugin/` |
| ECC | `.claude/`, `.agents/`, `.codex/` |

**Do not overlay** from `merged`: scripts (`dev-backend-build.sh`, `e2e-server.sh`),
`README.md`/`docs/`, `frontend/src/lib/components/context/`, `internal/summarize/`,
`internal/llm/` — purified already has Layer 1–2 + PR #26 and passes CI.

Purified tree retained for upstream parser, sync, postgres, artifact, duckdb, huma routes, kit-ui context, etc.

## Progress

| Step | Status |
|------|--------|
| PR #26 merged into purified | done (`ff8cd5b3`) |
| History-preserving merge commit | done |
| `go build -tags fts5 ./cmd/periscope` | pass |
| CI gate on PR #29 | pending |
| Experiment B (#28) | probe only — do not merge |

## Abandoned approach

Path-scoped `synthesis(passN)` commits on the integration branch **violate** SHA preservation policy. They remain on `cursor/agentsview-plus-periscope-synthetic-pass-f559` for reference only.
