#!/bin/bash
# Static checks for install.ps1 Periscope identity and legacy fallbacks.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_PS1="$SCRIPT_DIR/install.ps1"

assert_contains() {
    local needle="$1"
    local message="$2"
    if ! grep -Fq "$needle" "$INSTALL_PS1"; then
        echo "assertion failed: $message" >&2
        echo "missing: $needle" >&2
        exit 1
    fi
}

assert_not_contains() {
    local needle="$1"
    local message="$2"
    if grep -Fq "$needle" "$INSTALL_PS1"; then
        echo "assertion failed: $message" >&2
        echo "unexpected: $needle" >&2
        exit 1
    fi
}

assert_contains "diazMelgarejo/periscope" "install.ps1 should target the periscope fork"
assert_contains "'periscope.exe'" "install.ps1 should install periscope.exe"
assert_contains "Get-ReleaseArchiveCandidates" "install.ps1 should try periscope archives first"
assert_contains "agentsview_" "install.ps1 should retain legacy archive fallback"
assert_contains "PERISCOPE_SKIP_CHECKSUM" "install.ps1 should accept periscope checksum bypass"
assert_contains "AGENTSVIEW_SKIP_CHECKSUM" "install.ps1 should accept legacy checksum bypass"
assert_contains "function Install-Periscope" "install.ps1 should expose Install-Periscope"
assert_not_contains "Install-Agentsview" "install.ps1 should not keep Install-Agentsview entrypoint"

if command -v pwsh >/dev/null 2>&1; then
    pwsh -NoProfile -File "$SCRIPT_DIR/install_ps1_test.ps1"
elif command -v powershell >/dev/null 2>&1; then
    powershell -NoProfile -ExecutionPolicy Bypass -File "$SCRIPT_DIR/install_ps1_test.ps1"
else
    echo "PowerShell not available; static install.ps1 checks passed"
fi

echo "install.ps1 checks passed"
