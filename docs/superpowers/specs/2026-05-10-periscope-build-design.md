# Periscope fork — build system design spec

- **Status:** Historical design (May 2026) with 2026-07-28 modernization epilogue
- **Date:** 2026-05-10 (original), epilogue 2026-07-28
- **Branch:** `merged` in `diazMelgarejo/periscope`
- **Architecture ref:** [`docs/ARCHITECTURE.md`](../../ARCHITECTURE.md)
- **Canonical sync policy:** [`docs/guides/periscope-upstream-sync-blueprint.md`](../../guides/periscope-upstream-sync-blueprint.md)

---

## Context

`diazMelgarejo/periscope` integrates two canonical upstreams:

- **Layer 1:** `kenn-io/agentsview` (AgentsView foundation)
- **Layer 2:** `latentsignal-org/periscope` (Periscope product features)

This spec covers build system, release pipeline, and fork tooling (Layer 3). It
does **not** specify Periscope product UX — see `docs/periscope-spec.md` and
`docs/periscope-v1-plan.md`.

---

## Scope (Layer 3)

### In scope

1. **Product install/release branding** — `periscope` binary and artifacts on
   `merged`, with legacy `agentsview` compatibility during transition
2. **`scripts/sync-upstream.sh`** — dry-run, simulate, and guarded merge helper
3. **`scripts/install.sh`** — diazMelgarejo release download with checksum verify
4. **GitHub Actions release workflow** — platform binaries on `v*` tags
5. **`PeriscopeProcessManager.kt`** — JetBrains plugin lifecycle
6. **Versioning** — `v0.(upstream_minor+1).2-periscope.2-{commit}`

### Out of scope

- Periscope V1/V2 feature implementation
- macOS notarization / Tauri signing details → `docs/internal/desktop-release-setup.md`
- Full dual-pedigree replay automation (blueprint + disposable worktrees)

---

## Architecture summary

See [`docs/ARCHITECTURE.md`](../../ARCHITECTURE.md) for the matryoshka diagram.

| Layer | Source | Integration branch |
| --- | --- | --- |
| 1 | `kenn-io/agentsview` | Mirrored on `agentsview` |
| 2 | `latentsignal-org/periscope` | Mirrored on `main` |
| 3 | `diazMelgarejo/periscope` | `merged` |

---

## Decision log

| # | Decision | Choice | Rationale |
| --- | --- | --- | --- |
| D1 | Upstream sync strategy | Integrative replay + guarded merge helper | Proven conflict playbook; blueprint for dual pedigree |
| D2 | Unknown conflicts | Stop for operator (`--non-interactive` aborts) | Safety over silent auto-merge |
| D3 | Canonical AgentsView remote | `kenn-io/agentsview` | Current upstream identity (was `wesm/agentsview`) |
| D4 | Product binary | `periscope` | Product name; keep `agentsview` compat during transition |
| D5 | Go module (2026-07) | `go.kenn.io/agentsview` on modernization line | Matches current AgentsView foundation import path |
| D6 | Versioning | `v0.(upstream_minor+1).2-periscope.2-{sha}` | Readable base + immutable release pointer |
| D7 | Release binaries | GitHub Actions on `v*` tags | Canonical distribution |
| D8 | Remote mutation | Backup tag + lease SHA + explicit authorization | No automatic push/force from scripts |
| D9 | Deprecated code | Preserve when Periscope features depend on it | Correctness over tidiness |
| D10 | Session UI invariant | `SessionVitals` + `ActivityLane` | Replaces legacy `ActivityMinimap` |

---

## Component designs

### 1. Product binary and install path

**Target state on `merged`:**

- Release archives: `periscope_{version}_{os}_{arch}.tar.gz`
- Installed command: `periscope`
- Compatibility: accept legacy `agentsview_*` archives; optional `agentsview`
  symlink to `periscope`
- Build tree may still emit `agentsview` from `cmd/agentsview/` until Makefile
  rename debt clears (see rename catalogue)

**Verification:**

```bash
./scripts/install_test.sh
make build && ./agentsview --version   # during transition
```

---

### 2. `scripts/sync-upstream.sh`

**Modes:**

| Flag | Effect |
| --- | --- |
| `--dry-run` | Fetch, report incoming commits, run invariant checks — no merge |
| `--simulate` | Merge probe in disposable worktree — no branch mutation |
| `--source latentsignal\|kenn` | Select upstream (default: `latentsignal`) |
| `--non-interactive` | Abort on unresolved conflicts (no prompts) |
| `--authorize-remote-update` + `--lease-sha` | Required together to refresh a mirror ref |

**Never does:** `git push`, unqualified `--force`, or mirror updates without
authorization.

**Invariant checks after merge:** summarizer, LLM, ContextPage, SessionVitals,
JetBrains plugin, context API proxy.

---

### 3. `scripts/install.sh`

- Repo: `diazMelgarejo/periscope`, branch `merged` for curl entrypoint
- Version resolution: GitHub `/releases/latest` redirect (not `api.github.com`)
- Checksum env: `PERISCOPE_SKIP_CHECKSUM` with `AGENTSVIEW_SKIP_CHECKSUM` legacy
- Archive preference: `periscope_*` then `agentsview_*`

---

### 4. Release pipeline

Triggered by operator-pushed `v*` tags. Builds platform archives and attaches
checksums. Desktop and JetBrains artifacts follow workflow-specific naming
(`periscope-desktop-*` target state).

**Signing (from 72431d3b):** Desktop release steps are conditional on secrets;
unsigned builds disable updater artifacts when signing keys are absent.

---

### 5. JetBrains plugin lifecycle

`PeriscopeProcessManager.kt`:

- Resolve binary: `$PERISCOPE_BIN` → `~/.local/bin/periscope` → bundled resource
- Start backend, poll health, load JBCefBrowser
- SIGTERM on close; notification with install instructions when binary missing

---

## 2026-07-28 modernization epilogue

The May 2026 spec assumed:

- Module path `github.com/latentsignal-org/periscope`
- `cmd/periscope/` and wholesale `wesm/agentsview` mirror force-push workflow
- `ActivityMinimap` as a preserved UI invariant

**Current policy supersedes those items:**

| May 2026 assumption | 2026-07-28 state |
| --- | --- |
| `wesm/agentsview` canonical | `kenn-io/agentsview` canonical |
| Force-update `agentsview` mirror in sync script | Mirror refresh requires backup + lease + authorization; use blueprint |
| `ActivityMinimap` invariant | `SessionVitals` + `ActivityLane` |
| `PROGRESS.md` task ledger | `docs/guides/periscope-modernization-status.md` |
| Auto-push after sync | Never — operator reviews candidate locally |

First-release fork operations (install, sync helper, tag convention) are
**ported and modernized**; full module/cmd rename remains tracked in the rename
catalogue and modernization status doc.

---

## References

- Implementation plan: [`2026-05-10-periscope-build-system.md`](../plans/2026-05-10-periscope-build-system.md)
- Operator blueprint: [`periscope-upstream-sync-blueprint.md`](../../guides/periscope-upstream-sync-blueprint.md)
- Rename catalogue: [`agentsview-to-periscope-rename-catalogue.md`](../../guides/agentsview-to-periscope-rename-catalogue.md)
