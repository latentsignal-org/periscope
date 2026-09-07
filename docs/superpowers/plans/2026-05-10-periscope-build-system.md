# Periscope fork build system implementation plan

> **Status:** Historical plan (May 2026). First-release tasks completed on the
> original fork line; 2026-07-28 modernization replays tooling onto the current
> AgentsView foundation. **Do not treat task checkboxes as live operator truth**
> — see [`docs/guides/periscope-modernization-status.md`](../../guides/periscope-modernization-status.md).

**Goal:** Ship Periscope product identity, release tooling, and upstream sync
helpers on the `merged` integration line while preserving dual-pedigree mirrors.

**Architecture:** Three-layer matryoshka — see [`docs/ARCHITECTURE.md`](../../ARCHITECTURE.md).

**Spec:** [`2026-05-10-periscope-build-design.md`](../specs/2026-05-10-periscope-build-design.md)

**Out of scope here:** Periscope V1/V2 product features (`docs/periscope-v1-plan.md`,
`docs/periscope-v2-llm-plan.md`).

---

## File map (original + modernization notes)

| Action | File | Notes |
| --- | --- | --- |
| Keep | `go.mod` | Module `go.kenn.io/agentsview` on modernization line |
| Transition | `cmd/agentsview/` | Product binary `periscope`; rename tracked in catalogue |
| Modify | `Makefile` | Install/build artifact names → `periscope` where product-facing |
| Modify | `scripts/install.sh` | `diazMelgarejo/periscope` releases + legacy compat |
| Create | `scripts/sync-upstream.sh` | Dry-run / simulate / guarded merge |
| Modify | `.github/workflows/release.yml` | Product archive names (incremental) |
| Keep | `jetbrains-plugin/…/PeriscopeProcessManager.kt` | Lifecycle manager |
| Create | `docs/ARCHITECTURE.md` | Modernized matryoshka reference |
| Create | `docs/guides/periscope-modernization-status.md` | Replaces stale `PROGRESS.md` ledger |

---

## Task ledger (May 2026 first release — historical)

| # | Task | Original status | Modernization note |
| --- | --- | --- | --- |
| 1 | Go module rename | Done (May fork) | Superseded: stay on `go.kenn.io/agentsview` until integrative rename PR |
| 2 | Binary + Makefile rename | Done (May fork) | Partial on 3way line — install/script branding first |
| 3 | Version string in main.go | Done | Keep ldflags injection pattern |
| 4 | Update `install.sh` | Done | Re-ported with redirect version + dual archive names |
| 5 | Create `sync-upstream.sh` | Done | Re-ported with dry-run/simulate/safety gates |
| 6 | Update `release.yml` | Done (May) | CI still transitioning artifact names on 3way line |
| 7 | `PeriscopeProcessManager.kt` | Done | Unchanged |
| 8 | Wire JetBrains lifecycle | Done | Unchanged |
| 9 | Build verification + tag | Done | Tag convention: `v{semver}-{8-char-sha}` (5bd2e8a) |
| 10 | E2E install test | Done | `scripts/install_test.sh` |

---

## Release tag convention (5bd2e8a)

Tags embed the short commit hash:

```
v{semver}-{8-char-commit}   e.g.  v0.29.2-periscope.2-657a1090
```

```bash
COMMIT=$(git rev-parse --short=8 HEAD)
VERSION="v0.29.2-periscope.2"   # bump per versioning scheme
git tag -a "${VERSION}-${COMMIT}" -m "Release ${VERSION}-${COMMIT}"
# Operator pushes after review — tooling never auto-pushes.
```

---

## Desktop signing resilience (72431d3b)

When preparing desktop releases on the fork:

- Guard macOS certificate import on `APPLE_CERTIFICATE` presence
- Split signed vs unsigned Tauri builds on `TAURI_SIGNING_PRIVATE_KEY`
- Gate updater manifest upload on detected signatures
- Updater endpoint targets `diazMelgarejo/periscope`

Required secrets when signing: `TAURI_SIGNING_PRIVATE_KEY`, `PERISCOPE_UPDATER_PUBKEY`.

---

## Modernization replay checklist (2026-07-28)

Use on the current `merged` candidate — not the May worktree paths.

- [ ] Read [`periscope-upstream-sync-blueprint.md`](../../guides/periscope-upstream-sync-blueprint.md)
- [ ] Prove mirror exactness (`agentsview` ↔ kenn-io, `main` ↔ latentsignal)
- [ ] `./scripts/sync-upstream.sh --dry-run --source kenn`
- [ ] `./scripts/sync-upstream.sh --simulate --source kenn` in disposable worktree
- [ ] Replay Periscope innovations integratively onto current AgentsView tree
- [ ] `./scripts/install_test.sh`
- [ ] `go fmt ./...` · `go vet ./...` · `CGO_ENABLED=1 go test -tags fts5 ./...`
- [ ] `make build`
- [ ] Update [`periscope-modernization-status.md`](../../guides/periscope-modernization-status.md)

---

## Resume pointer

Do **not** resume from historical `PROGRESS.md` paths (`/tmp/periscope-work/...`).
Start from:

```bash
cd /path/to/periscope
git checkout merged
cat docs/guides/periscope-modernization-status.md
./scripts/sync-upstream.sh --help
```
