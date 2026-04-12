#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ARTIFACTS_WORKFLOW="$REPO_ROOT/.github/workflows/desktop-artifacts.yml"
RELEASE_WORKFLOW="$REPO_ROOT/.github/workflows/desktop-release.yml"
DOC_FILE="$REPO_ROOT/docs/desktop-release-setup.md"

assert_contains() {
  local file="$1"
  local needle="$2"
  local message="$3"

  if ! grep -Fq "$needle" "$file"; then
    echo "assertion failed: $message" >&2
    echo "missing: $needle" >&2
    echo "file: $file" >&2
    exit 1
  fi
}

assert_not_contains() {
  local file="$1"
  local needle="$2"
  local message="$3"

  if grep -Fq "$needle" "$file"; then
    echo "assertion failed: $message" >&2
    echo "unexpected: $needle" >&2
    echo "file: $file" >&2
    exit 1
  fi
}

assert_contains "$ARTIFACTS_WORKFLOW" "name: Linux (arm64)" \
  "desktop artifacts workflow should build Linux arm64 in PR CI"
assert_contains "$ARTIFACTS_WORKFLOW" "os: ubuntu-22.04-arm" \
  "desktop artifacts workflow should use the Ubuntu arm runner"
assert_contains "$ARTIFACTS_WORKFLOW" "target_triple: aarch64-unknown-linux-gnu" \
  "desktop artifacts workflow should target Linux arm64"
assert_contains "$ARTIFACTS_WORKFLOW" "artifact_name: agentsview-desktop-linux-arm64" \
  "desktop artifacts workflow should upload a distinct Linux arm64 artifact"
assert_contains "$ARTIFACTS_WORKFLOW" "xdg-utils" \
  "desktop artifacts workflow should install xdg-utils for AppImage bundling"

assert_contains "$RELEASE_WORKFLOW" 'name: Desktop Build (Linux ${{ matrix.arch }})' \
  "desktop release workflow should matrix Linux builds by arch"
assert_contains "$RELEASE_WORKFLOW" 'runs-on: ${{ matrix.os }}' \
  "desktop release workflow should select Linux runner per arch"
assert_contains "$RELEASE_WORKFLOW" "os: ubuntu-22.04-arm" \
  "desktop release workflow should ship Linux arm64 from an arm runner"
assert_contains "$RELEASE_WORKFLOW" "artifact_name: agentsview-desktop-linux-arm64" \
  "desktop release workflow should upload a distinct Linux arm64 release artifact"
assert_contains "$RELEASE_WORKFLOW" 'create_updater_artifacts: "false"' \
  "desktop release workflow should disable updater artifacts for Linux arm64"
assert_contains "$RELEASE_WORKFLOW" "linux-x86_64" \
  "desktop release workflow should keep Linux x86_64 updater support"
assert_contains "$RELEASE_WORKFLOW" "xdg-utils" \
  "desktop release workflow should install xdg-utils for AppImage bundling"
assert_not_contains "$RELEASE_WORKFLOW" 'linux-aarch64' \
  "desktop release workflow should not add Linux arm64 to latest.json"

assert_contains "$DOC_FILE" "AgentsView_x.y.z_aarch64.AppImage" \
  "desktop release docs should mention the Linux arm64 AppImage"

echo "desktop workflow checks passed"
