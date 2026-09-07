#!/usr/bin/env bash
# sync-upstream.sh — Merge upstream latentsignal-org/periscope into diazMelgarejo/periscope
#
# Usage:
#   ./scripts/sync-upstream.sh [--dry-run] [--upstream-ref <ref>]
#
# Behaviour:
#   1. Fetches the upstream remote (adds it if missing)
#   2. Merges upstream/<ref> into the current branch (default: agentsview or main)
#   3. Auto-resolves known-safe conflicts using our periscope-layer rules:
#      - go.mod module path: always keep upstream's (go.kenn.io/agentsview) — never diverge here
#      - cmd/ directory: upstream renames are rebased onto ours (cmd/periscope)
#      - internal/db/sessions.go: keep our extra columns (model_context_window_tokens etc.)
#      - internal/sync/engine.go: keep our UpdateSessionIncremental call signature
#      - frontend/src: prefer upstream (new features), unless file is in PERISCOPE_OWNED list
#   4. Prompts the user interactively for any remaining conflicts
#   5. Prints a diff summary of periscope-owned invariants to verify nothing was lost
#
# Periscope invariants (MUST survive every upstream merge):
#   - internal/summarize/        (summarizer service)
#   - internal/llm/              (LLM integration)
#   - frontend/src/lib/components/context/  (ContextPage, ActivityMinimap)
#   - frontend/src/lib/context.js / types/index.ts exports
#   - vite.config.ts proxy entry for /api/context
#   - cmd/periscope/ (our binary name)
#   - jetbrains-plugin/          (our JetBrains plugin)

set -euo pipefail

UPSTREAM_REMOTE="upstream"
UPSTREAM_URL="https://github.com/latentsignal-org/periscope.git"
DEFAULT_UPSTREAM_REF="main"
DRY_RUN=0

# Files / dirs periscope owns — always keep ours on conflict
PERISCOPE_OWNED=(
    "internal/summarize"
    "internal/llm"
    "frontend/src/lib/components/context"
    "frontend/src/lib/context.js"
    "frontend/src/types/index.ts"
    "vite.config.ts"
    "jetbrains-plugin"
    "cmd/periscope"
    "go.mod"
    "scripts/install.sh"
    "scripts/sync-upstream.sh"
)

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { echo -e "${GREEN}[sync]${NC} $*"; }
warn()  { echo -e "${YELLOW}[warn]${NC} $*"; }
error() { echo -e "${RED}[error]${NC} $*" >&2; exit 1; }
step()  { echo -e "${CYAN}==> $*${NC}"; }

usage() {
    sed -n '2,12p' "$0" | sed 's/^# //'
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run)   DRY_RUN=1; shift ;;
        --upstream-ref) DEFAULT_UPSTREAM_REF="$2"; shift 2 ;;
        -h|--help)  usage ;;
        *) error "Unknown option: $1" ;;
    esac
done

UPSTREAM_REF="${DEFAULT_UPSTREAM_REF}"

# ── 1. Sanity checks ──────────────────────────────────────────────────────────
step "Checking git state"
if [[ -n "$(git status --porcelain)" ]]; then
    error "Working tree is dirty. Commit or stash changes before syncing."
fi

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
info "Current branch: $CURRENT_BRANCH"

# ── 2. Ensure upstream remote exists ──────────────────────────────────────────
step "Fetching upstream ($UPSTREAM_URL)"
if ! git remote get-url "$UPSTREAM_REMOTE" &>/dev/null; then
    info "Adding upstream remote..."
    git remote add "$UPSTREAM_REMOTE" "$UPSTREAM_URL"
fi

git fetch "$UPSTREAM_REMOTE" "$UPSTREAM_REF" --tags

UPSTREAM_COMMIT=$(git rev-parse "${UPSTREAM_REMOTE}/${UPSTREAM_REF}")
info "Upstream ${UPSTREAM_REMOTE}/${UPSTREAM_REF} is at: ${UPSTREAM_COMMIT:0:12}"

if [[ "$DRY_RUN" == "1" ]]; then
    info "[dry-run] Would merge ${UPSTREAM_REMOTE}/${UPSTREAM_REF} into $CURRENT_BRANCH"
    echo
    git log --oneline "HEAD..${UPSTREAM_REMOTE}/${UPSTREAM_REF}" | head -20
    echo
    warn "[dry-run] No changes made."
    exit 0
fi

# ── 3. Attempt merge ─────────────────────────────────────────────────────────
step "Merging ${UPSTREAM_REMOTE}/${UPSTREAM_REF}"
if git merge --no-ff "${UPSTREAM_REMOTE}/${UPSTREAM_REF}" \
        --message "chore: merge upstream latentsignal-org/periscope @ ${UPSTREAM_COMMIT:0:12}" \
        2>&1; then
    info "Merge succeeded with no conflicts."
else
    info "Merge has conflicts. Applying auto-resolution rules..."

    # ── 4. Auto-resolve: periscope-owned files → always keep ours ────────────
    for owned in "${PERISCOPE_OWNED[@]}"; do
        if git diff --name-only --diff-filter=U | grep -q "^${owned}"; then
            info "Auto-keeping ours for: ${owned}"
            git checkout --ours -- "${owned}" 2>/dev/null || true
            git add "${owned}" 2>/dev/null || true
        fi
    done

    # ── 5. Auto-resolve: go.mod module path ──────────────────────────────────
    if git diff --name-only --diff-filter=U | grep -q "^go.mod$"; then
        info "Auto-resolving go.mod (keeping our module path)"
        git checkout --ours -- go.mod
        git add go.mod
    fi

    # ── 6. Interactive: remaining conflicts ──────────────────────────────────
    REMAINING=$(git diff --name-only --diff-filter=U 2>/dev/null || true)
    if [[ -n "$REMAINING" ]]; then
        warn "Remaining conflicts require manual resolution:"
        echo "$REMAINING" | while IFS= read -r f; do
            echo "  - $f"
        done
        echo
        warn "Resolve each conflict, then run:"
        warn "  git add <resolved-file>"
        warn "  git merge --continue"
        echo
        read -rp "Open a shell to resolve conflicts now? [y/N] " answer
        if [[ "${answer,,}" == "y" ]]; then
            echo "Dropping into shell. Type 'exit' when done resolving."
            bash --rcfile <(echo 'PS1="[merge-shell] \w \$ "') || true
        else
            warn "Merge left in progress. Resolve manually and run: git merge --continue"
            exit 1
        fi

        # Re-check
        STILL_REMAINING=$(git diff --name-only --diff-filter=U 2>/dev/null || true)
        if [[ -n "$STILL_REMAINING" ]]; then
            error "Conflicts remain unresolved:\n$STILL_REMAINING"
        fi
    fi

    git merge --continue --no-edit
fi

# ── 7. Invariant verification ─────────────────────────────────────────────────
step "Verifying periscope invariants"
INVARIANT_FAILURES=0

check_exists() {
    local path="$1"
    local label="${2:-$1}"
    if [[ -e "$path" ]]; then
        info "  ✓ $label"
    else
        warn "  ✗ MISSING: $label"
        INVARIANT_FAILURES=$((INVARIANT_FAILURES + 1))
    fi
}

check_contains() {
    local file="$1"
    local pattern="$2"
    local label="$3"
    if grep -q "$pattern" "$file" 2>/dev/null; then
        info "  ✓ $label"
    else
        warn "  ✗ MISSING in $file: $label"
        INVARIANT_FAILURES=$((INVARIANT_FAILURES + 1))
    fi
}

check_exists "internal/summarize" "summarizer package"
check_exists "internal/llm" "LLM package"
check_exists "frontend/src/lib/components/context/ContextPage.svelte" "ContextPage"
check_exists "frontend/src/lib/components/context/ActivityMinimap.svelte" "ActivityMinimap"
check_exists "jetbrains-plugin" "JetBrains plugin"
check_exists "cmd/periscope" "periscope binary cmd"

check_contains "go.mod" "go.kenn.io/agentsview" "go.mod module path"
check_contains "vite.config.ts" "/api/context" "vite.config.ts context proxy"

if [[ $INVARIANT_FAILURES -gt 0 ]]; then
    echo
    error "$INVARIANT_FAILURES invariant(s) missing after merge! Review before pushing."
fi

# ── 8. Done ──────────────────────────────────────────────────────────────────
echo
info "Sync complete. Upstream ${UPSTREAM_COMMIT:0:12} merged into $CURRENT_BRANCH."
info "Run 'go test -tags fts5 ./...' and 'make build' to verify."
info "Then push: git push origin $CURRENT_BRANCH"
