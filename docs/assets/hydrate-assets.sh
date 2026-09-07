#!/usr/bin/env bash
# Populate ignored docs asset directories from orphan asset branches.
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
docs_root="$(cd "$script_dir/.." && pwd)"
repo_root="$(cd "$docs_root/.." && pwd)"
static_branch="${AGENTSVIEW_DOCS_ASSETS_BRANCH:-docs-assets}"
generated_branch="${AGENTSVIEW_DOCS_GENERATED_ASSETS_BRANCH:-docs-generated-assets}"
use_local_branches="${AGENTSVIEW_DOCS_USE_LOCAL_ASSET_BRANCHES:-false}"
assets_upstream="${AGENTSVIEW_DOCS_ASSETS_UPSTREAM:-https://github.com/kenn-io/agentsview.git}"
assets_upstream_remote="${AGENTSVIEW_DOCS_ASSETS_UPSTREAM_REMOTE:-docs-assets-upstream}"

static_target="$docs_root/assets/static"
generated_target="$docs_root/assets/generated"

static_assets=(
  "architecture.svg"
  "og-image.png"
)

generated_assets=(
  "screenshots/about-dialog.png"
  "screenshots/activity-breakdowns.png"
  "screenshots/activity-concurrency.png"
  "screenshots/activity-insight.png"
  "screenshots/activity-page.png"
  "screenshots/activity-sessions.png"
  "screenshots/activity-timeline.png"
  "screenshots/activity-week.png"
  "screenshots/agent-comparison.png"
  "screenshots/analytics-model-filter.png"
  "screenshots/block-filter.png"
  "screenshots/code-block-copy-btn.png"
  "screenshots/command-palette.png"
  "screenshots/dashboard.png"
  "screenshots/date-range.png"
  "screenshots/focused-transcript.png"
  "screenshots/follow-latest-toggle.png"
  "screenshots/grade-badge.png"
  "screenshots/heatmap-filtered.png"
  "screenshots/heatmap.png"
  "screenshots/hour-of-week.png"
  "screenshots/import-button.png"
  "screenshots/import-modal-chatgpt.png"
  "screenshots/import-modal-claude.png"
  "screenshots/in-session-search.png"
  "screenshots/layout-compact.png"
  "screenshots/layout-stream.png"
  "screenshots/machine-labels.png"
  "screenshots/message-copy-btn.png"
  "screenshots/message-viewer.png"
  "screenshots/project-breakdown.png"
  "screenshots/publish-modal.png"
  "screenshots/recent-edits.png"
  "screenshots/resync-modal.png"
  "screenshots/search-grouped.png"
  "screenshots/search-results.png"
  "screenshots/semantic-search-setup.png"
  "screenshots/session-filtered.png"
  "screenshots/session-filters-active.png"
  "screenshots/session-filters.png"
  "screenshots/session-health.png"
  "screenshots/session-insight-action.png"
  "screenshots/session-list.png"
  "screenshots/session-shape.png"
  "screenshots/session-vital-signs.png"
  "screenshots/settings-remote.png"
  "screenshots/settings.png"
  "screenshots/shortcuts-modal.png"
  "screenshots/signal-panel.png"
  "screenshots/starred-session.png"
  "screenshots/subagent-tree.png"
  "screenshots/summary-cards.png"
  "screenshots/theme-dark.png"
  "screenshots/theme-light.png"
  "screenshots/thinking-blocks.png"
  "screenshots/token-usage.png"
  "screenshots/tool-block-copy-btn.png"
  "screenshots/tool-blocks.png"
  "screenshots/tool-groups.png"
  "screenshots/tool-usage.png"
  "screenshots/top-sessions.png"
  "screenshots/top-skills.png"
  "screenshots/trends.png"
  "screenshots/usage-attribution.png"
  "screenshots/usage-cache-efficiency.png"
  "screenshots/usage-cost-trend.png"
  "screenshots/usage-filter-dropdown.png"
  "screenshots/usage-page.png"
  "screenshots/usage-summary-cards.png"
  "screenshots/usage-toolbar.png"
  "screenshots/usage-top-sessions.png"
  "screenshots/velocity.png"
  "screenshots/vital-signs-panel.png"
  "screenshots/worktree-mappings.png"
)

# Screenshots periscope's own Playwright suite (docs/screenshots/tests/
# screenshots.spec.ts: "full insights page", "insight content") can still
# generate today -- the Insights page they cover is real, live app code
# (internal/insight/, frontend/src/lib/stores/insights.svelte.ts) -- but
# that regeneration requires the Docker-based pipeline
# (docs/screenshots/run.sh --push), which was not available when this
# split was introduced. Tracked as pending, not silently dropped: warn
# instead of hard-failing CI so the gap stays visible without blocking
# every unrelated docs change. Move an entry back into generated_assets
# once docs/screenshots/run.sh --push has actually republished it.
pending_generated_assets=(
  "screenshots/insight-content.png"
  "screenshots/insights.png"
)

has_expected_assets() {
  local target="$1"
  shift

  local asset
  for asset in "$@"; do
    [[ -f "$target/$asset" ]] || return 1
  done
}

in_git_worktree() {
  git -C "$repo_root" rev-parse --is-inside-work-tree >/dev/null 2>&1
}

fetch_asset_branch() {
  local remote="$1"
  local branch="$2"

  git -C "$repo_root" fetch --force --depth=1 "$remote" \
    "+refs/heads/$branch:refs/remotes/$remote/$branch" >/dev/null
}

ensure_upstream_assets_remote() {
  if git -C "$repo_root" remote get-url "$assets_upstream_remote" >/dev/null 2>&1; then
    git -C "$repo_root" remote set-url "$assets_upstream_remote" "$assets_upstream"
    return 0
  fi

  git -C "$repo_root" remote add "$assets_upstream_remote" "$assets_upstream"
}

resolve_asset_ref() {
  local branch="$1"

  if [[ "$use_local_branches" == "1" || "$use_local_branches" == "true" ]]; then
    if git -C "$repo_root" rev-parse --verify --quiet "$branch" >/dev/null; then
      printf '%s\n' "$branch"
      return 0
    fi
  fi

  if fetch_asset_branch origin "$branch"; then
    if git -C "$repo_root" rev-parse --verify --quiet "origin/$branch" >/dev/null; then
      printf 'origin/%s\n' "$branch"
      return 0
    fi
  fi

  if ensure_upstream_assets_remote && fetch_asset_branch "$assets_upstream_remote" "$branch"; then
    if git -C "$repo_root" rev-parse --verify --quiet \
      "$assets_upstream_remote/$branch" >/dev/null; then
      printf '%s/%s\n' "$assets_upstream_remote" "$branch"
      return 0
    fi
  fi

  if [[ "$use_local_branches" == "1" || "$use_local_branches" == "true" ]] &&
    git -C "$repo_root" rev-parse --verify --quiet "$branch" >/dev/null; then
    printf '%s\n' "$branch"
    return 0
  fi

  printf 'docs assets not hydrated: failed to fetch %s from origin or upstream\n' "$branch" >&2
  return 1
}

hydrate_branch() {
  local branch="$1"
  local target="$2"
  shift 2

  if ! in_git_worktree; then
    if has_expected_assets "$target" "$@"; then
      return 0
    fi

    printf 'docs assets not hydrated: no git worktree found and expected assets are missing\n' >&2
    return 1
  fi

  local asset_ref
  if ! asset_ref="$(resolve_asset_ref "$branch")"; then
    printf 'docs assets not hydrated: %s branch unavailable\n' "$branch" >&2
    return 1
  fi

  rm -rf "$target"
  mkdir -p "$target"
  git -C "$repo_root" archive "$asset_ref" | tar -xf - -C "$target"

  if ! has_expected_assets "$target" "$@"; then
    printf 'docs assets not hydrated: %s is missing expected assets\n' "$branch" >&2
    return 1
  fi
}

warn_missing_pending_assets() {
  local target="$1"
  shift

  local asset
  for asset in "$@"; do
    if [[ ! -f "$target/$asset" ]]; then
      printf 'docs assets: pending regeneration, not blocking CI: %s (see docs/assets/hydrate-assets.sh)\n' "$asset" >&2
    fi
  done
}

hydrate_branch "$static_branch" "$static_target" "${static_assets[@]}"
hydrate_branch "$generated_branch" "$generated_target" "${generated_assets[@]}"
warn_missing_pending_assets "$generated_target" "${pending_generated_assets[@]}"
