#!/usr/bin/env bash
set -euo pipefail

# Wrapper that restores tauri.conf.json after `tauri` exits,
# undoing the version patch applied by prepare-sidecar.sh.
# Uses the .orig backup instead of git checkout to preserve
# any pre-existing uncommitted edits.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONF="$SCRIPT_DIR/../src-tauri/tauri.conf.json"

cleanup() {
  if [ -f "$CONF.orig" ]; then
    mv "$CONF.orig" "$CONF"
  fi
}
trap cleanup EXIT INT TERM

tauri "$@"
