# Synthesis experiment A — `merged` ← `purified+PR26`

**Branch:** `cursor/agentsview-plus-periscope-f559`  
**Base for PR:** `merged`  
**Incoming:** `cursor/agentsview-purified-onto-kenn-f559` (includes merged PR #26 / upstream #23 stack)

## Intent

Canonical **Periscope fork identity** (module `go.kenn.io/agentsview`, binary
`periscope`, Layer 3 tooling) remains the merge base. Absorb the purified upstream
replay (kenn-io modernization + #1274/#1251/#1284) additively.

## Divergence (simulated 2026-07-29)

| Metric | Value |
|--------|-------|
| Merge-base | `6c3317ad` (agentsview mirror tip) |
| Commits only on purified | 25 |
| Commits only on merged | 76 |
| Files changed (tree diff) | 2133 |
| Simulated merge conflicts | **735** paths |

## Conflict hotspots (top)

| Area | Conflicted paths |
|------|------------------|
| `cmd/periscope` | 143 |
| `internal/sync` | 81 |
| `internal/server` | 81 |
| `internal/postgres` | 72 |
| `internal/db` | 51 |
| `internal/duckdb` | 36 |
| `internal/parser` | 33 |
| `scripts/git` | 16 |

## Resolution bias (oramasys-method)

Per `docs/ARCHITECTURE.md` matryoshka rules:

1. **Layer 1 (upstream agentsview):** take purified/upstream for parser, sync, postgres push.
2. **Layer 2 (Periscope features):** union — keep context visualizer, summarize, LLM routes from merged.
3. **Layer 3 (fork identity):** keep merged module path, binary name, sync-upstream, release CI.
4. **Git guards:** superset — orama canonical `verify-staged-for-commit.sh` + periscope copies.
5. **ECC bundles:** path-scoped replay (see integrative-merge PR #12 / #25 precedent).

## Status

Manifest-only commit. Full harmonization is a multi-pass integrative merge (Mode 2),
not a single `-X ours/theirs`.
