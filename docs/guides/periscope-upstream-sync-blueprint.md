# Periscope upstream sync blueprint

**Captured:** 2026-07-28
**Status:** operational report; review before any remote mutation
**Scope:** keep `diazMelgarejo/periscope` current with AgentsView while
preserving Periscope innovations on the integration line.

This report records the repository and memory analysis performed before the
2026-07-28 nondestructive modernization probe. The authoritative policy is the
2026-07-28 revalidation in orama-system. Older branch tables in this repository
remain useful for fork invariants and conflict examples, but they do not
override the current three-branch model.

## Canonical document hierarchy

| Priority | Path | Role |
| --- | --- | --- |
| 1 | `orama-system/docs/plans/2026-05-24-periscope-l4-integration-plan.md` | Primary plan. Read Design canon, 2026-07-28 revalidation, and Epilogue. The historical May execution appendix is superseded. |
| 2 | `orama-system/docs/reference/periscope-cursor-repo-rules.md` | Branch model, Cursor doctrine, and ordered integration policy. |
| 3 | `orama-system/scripts/periscope/cursor-rules-templates/openclaw-fork-guide.mdc` | Periscope Cursor rule template. |
| 4 | `orama-system/docs/plans/2026-07-28-periscope-lineage-modernization-epic.md` | Optional clean semantic patch-stack epic. |
| 5 | `orama-system/bin/orama-system/skills/oramasys-method/references/integrative-merge.md` | Mandatory conflict doctrine: synthesize; never amputate. |
| 6 | `orama-system/bin/orama-system/skills/git-history-surgery/references/path-scoped-pr-replay-reference-card.md` | Replay technique when `merged` already contains overlapping content. |
| 7 | `orama-system/bin/orama-system/skills/git-history-surgery/SKILL.md` | Branch-state and history-surgery decision flow. |
| 8 | `orama-system/bin/orama-system/skills/periscope-ecc/SKILL.md` | ECC mirror verification and overlapping-PR replay. |
| 9 | `Perpetua-Tools/.agent/memory/working/PERISCOPE_DUAL_PEDIGREE_REANCHOR_2026-07-28.md` | Executed dual-pedigree recovery evidence and repeatable procedure. |
| 10 | `Perpetua-Tools/.agent/memory/working/WORKSPACE_PR_BASE_BRANCHES_2026-07-28.md` | PR-base policy: Periscope agent work targets `merged`. |
| 11 | `periscope/docs/ARCHITECTURE.md` | Fork invariants, owned paths, and conflict examples. Branch roles live in this blueprint. |
| 11b | `periscope/docs/guides/periscope-modernization-status.md` | Current modernization snapshot (replaces stale `PROGRESS.md` ledger). |
| 12 | `periscope/scripts/sync-upstream.sh` | Operational latentsignal upstream merge helper. |

## Canonical branch roles

| Branch | Source | Role | Agent PR target |
| --- | --- | --- | --- |
| `agentsview` | `kenn-io/agentsview:main` | Exact AgentsView mirror | No |
| `main` | `latentsignal-org/periscope:main` | Exact upstream Periscope mirror | No |
| `merged` | Both mirror pedigrees plus fork innovations | Integration and shipping line | Yes |

Hard invariants:

- Never merge `merged` into `main`.
- Never land build or integration work on either mirror branch.
- Never use mirror `main` as a wholesale salvage source.
- Preserve both mirror tips in `merged` ancestry.
- Keep product innovations as additive patches over the current upstream
  foundation.
- Do not merge stale or overlapping PR branches wholesale; replay their proven
  semantic or path-scoped delta onto fresh `merged`.

## Required sequence

### 0. Read policy before Git writes

Read:

1. The orama-system L4 plan's 2026-07-28 revalidation.
2. The integrative-merge reference.
3. The git-history-surgery skill when ancestry or a prior rewrite is involved.
4. PT `.agent` memory for dual-pedigree recovery, rename policy, and ECC replay.

### 1. Bootstrap repository safeguards

```bash
bash scripts/git/apply-attribution-guard-all-repos.sh
bash scripts/git/check_identity.sh
```

Install or refresh Periscope-specific Cursor rules from orama-system when
needed:

```bash
export PERISCOPE_REPO=/path/to/periscope
bash /path/to/orama-system/scripts/periscope/install-cursor-rules.sh
```

### 2. Detect upstream movement

Fetch the two real sources into distinct remote-tracking refs:

```bash
git fetch https://github.com/kenn-io/agentsview.git \
  main:refs/remotes/upstream-kenn/main
git fetch https://github.com/latentsignal-org/periscope.git \
  main:refs/remotes/upstream-latentsignal/main
git fetch origin agentsview main merged
```

Prove mirror trees:

```bash
git diff --quiet origin/agentsview upstream-kenn/main
git diff --quiet origin/main upstream-latentsignal/main
```

Record remote lease SHAs before any proposed remote update:

```bash
git ls-remote --heads origin main agentsview merged
```

Do not diagnose rewritten branch health from ahead/behind counts. Use the
tree-twin scanner:

```bash
/path/to/orama-system/scripts/git/reanchor_scan.sh \
  /path/to/periscope origin/merged remotes
```

### 3. Refresh mirrors only when their source moves

Before force-updating a mirror:

1. Obtain explicit operator authorization.
2. Preserve the old remote tip under an immutable
   `backup/pre-reanchor-<branch>-<timestamp>` tag.
3. Record the exact remote SHA as the lease.
4. Prove the candidate mirror tree equals the source tree.
5. Use ref-specific `--force-with-lease`; never use unqualified `--force`.

### 4. Advance `merged` integratively

`merged` has dual pedigree. Do not blindly rebase it onto one mirror and lose
the other.

For ancestry repair with no intended content change, simulate in a disposable
worktree:

```bash
git checkout -B <probe> origin/main
git merge -s ours --no-ff origin/agentsview \
  -m "integrative: dual-pedigree anchor"
git read-tree --reset -u origin/merged^{tree}
git commit --amend -m "reanchor: restore merged tree on dual-pedigree base"
git diff --quiet origin/merged HEAD
git merge-base --is-ancestor origin/main HEAD
git merge-base --is-ancestor origin/agentsview HEAD
```

For actual AgentsView code movement, preserving ancestry is not enough. Replay
or synthesize the fork's unique innovations on the current
`upstream-kenn/main` tree in a disposable worktree. Resolve conflicts using the
six integrative modes in order:

1. additive
2. union
3. superset
4. synthesize
5. architecturally correct
6. API correct

Do not use whole-file `ours` or `theirs` without reading both sides.

### 5. Preserve fork invariants

Always preserve these Periscope capabilities:

- Context page and context/transcript navigation.
- SessionVitals and ActivityLane (supersedes legacy ActivityMinimap).
- Context timeline, token accounting, and model-context capacity.
- Context-guidance signals.
- Summarizer worker and LLM client.
- Guidance client/model/cache.
- Periscope context and summarize API routes.
- Dev proxy behavior needed by embedded clients.
- Context API type exports.
- JetBrains plugin.
- Periscope binary, desktop sidecar, release, and artifact contracts.
- Product branding and compatibility migrations.

Conflict guidance:

| File kind | Resolution |
| --- | --- |
| Fork-owned feature directories | Preserve fork feature; adapt it to current upstream APIs. |
| Upstream parsers, sync, and storage improvements | Prefer current upstream, then reapply required Periscope extensions. |
| Shared structs and API exports | Union fields/exports; do not replace one feature set with another. |
| Frontend shared composition | Synthesize upstream UI additions with ContextPage, SessionVitals, ActivityLane, and guidance blocks. |
| Database data version | Use the higher valid upstream version and add Periscope migrations additively. |
| Module/binary identity | Preserve the Periscope product contract where functional. |
| Tests | Build a unified suite exercising both upstream and fork behavior. |

### 6. Apply the AgentsView rename policy

AgentsView mirror identity is intentional; product identity is Periscope.

| Signal | Action |
| --- | --- |
| Branch or remote named `agentsview` | Keep |
| Historical upstream references and fixtures | Keep |
| Binary, command, CI artifact, desktop sidecar, product branding | Rename to Periscope |
| Legacy environment variables | Read for compatibility; prefer Periscope names |
| Legacy localStorage keys | Dual-read and migrate |

After each upstream replay, rediscover literals and classify each affected path:

```bash
rg -l 'agentsview|AgentsView|AGENTSVIEW' \
  --glob '!**/node_modules/**' --glob '!**/Cargo.lock'
```

### 7. Replay open PRs only after the candidate base is proven

For each open `merged`-based PR:

1. Preserve its old head and old base.
2. Start from the repaired or modernized `merged` candidate.
3. Replay only the old-base-to-old-head semantic delta.
4. For overlapping bundles, copy only harmonized unique paths.
5. Exclude timestamp-only generated metadata unless intentional.
6. Verify tests and final-tree intent before proposing a remote update.

## Verification gates

| Gate | Criterion |
| --- | --- |
| Mirror exactness | `origin/agentsview` equals `upstream-kenn/main`; `origin/main` equals `upstream-latentsignal/main`. |
| Dual pedigree | Both mirror tips are ancestors of the candidate integration line. |
| Upstream foundation | Candidate starts from or incorporates the current AgentsView source tree, not merely an `ours` ancestry marker. |
| Innovation preservation | Every valid fork capability above has code and test anchors. |
| Rename policy | Functional product contracts say Periscope; compatibility and historical references are retained intentionally. |
| Conflict hygiene | No conflict markers; no silent path deletion. |
| Go formatting | `go fmt ./...` |
| Go analysis | `go vet ./...` |
| Go tests | `CGO_ENABLED=1 go test -tags fts5 ./...` |
| Frontend | `npm run check`, relevant unit tests, and production build |
| Full binary | `make build` |
| Desktop | Sidecar/workflow contract tests and Rust tests where toolchain supports them |
| ECC | Agents and Claude skill mirrors are byte-identical |
| Git attribution | Identity and banned-token guards pass |
| Remote safety | No push until operator reviews the disposable candidate |

## Destructive and authorization constraints

- No force-push without explicit current-user authorization.
- No remote update without preserving backup tags and exact lease SHAs.
- Never force-update or merge into mirror `main` with fork content.
- Never merge `merged` into `main`.
- Never merge `agentsview` wholesale into `merged` merely because its tip moved.
- Never rewrite open PR branches wholesale when their base contains overlapping
  content.
- Unknown conflicts stop for integrative review.
- Preserve vulnerability records, lessons, audits, and review ledgers
  additively.
- Keep the modernization candidate local and unpushed until operator review.

## 2026-07-28 preflight evidence

Before this report was saved:

- `origin/agentsview` and `upstream-kenn/main` both resolved to
  `6c3317ad69eb1383928833dda006957a7a2d1f0d`.
- `origin/main` and `upstream-latentsignal/main` both resolved to
  `852b8e381ead918dc70e64c25233c124b8ecb5e1`.
- Both mirror trees were identical to their source trees.
- Both mirror tips were ancestors of `origin/merged`.
- The existing `merged` tree nevertheless differed materially from current
  AgentsView, confirming that the prior `ours` ancestry anchor did not itself
  integrate the newer AgentsView product tree.
- A disposable worktree was created from current AgentsView for semantic replay.

## Operational summary

Keep `agentsview` tree-identical to `kenn-io/agentsview:main`, keep `main`
tree-identical to `latentsignal-org/periscope:main`, and maintain `merged` as the
current AgentsView foundation plus additive Periscope innovations. Use
disposable-worktree simulation, semantic replay, and integrative conflict
resolution. Prove the candidate locally before proposing any update to
`origin/merged`.
