#!/bin/bash
# Smoke tests for sync-upstream.sh (non-destructive).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SYNC="$SCRIPT_DIR/sync-upstream.sh"

PASS=0
FAIL=0

assert_exit() {
    local desc="$1" expected="$2"
    shift 2
    if "$@" >/dev/null 2>&1; then
        local actual=0
    else
        local actual=$?
    fi
    if [ "$expected" -eq "$actual" ]; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc (expected exit $expected, got $actual)"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== sync-upstream.sh smoke ==="

[ -x "$SYNC" ] || chmod +x "$SYNC"

assert_exit "--help exits 0" 0 "$SYNC" --help

output="$("$SYNC" --refresh-mirror agentsview 2>&1 || true)"
if echo "$output" | grep -q "requires --authorize-remote-update"; then
    echo "  PASS: refresh-mirror requires authorization"
    PASS=$((PASS + 1))
else
    echo "  FAIL: refresh-mirror should require authorization"
    FAIL=$((FAIL + 1))
fi

lease_sha="$(git rev-parse origin/agentsview 2>/dev/null || echo "")"
if [ -n "$lease_sha" ]; then
    assert_exit "refresh-mirror with matching lease exits 0" 0 \
        "$SYNC" --refresh-mirror agentsview \
        --authorize-remote-update --lease-sha "$lease_sha"
else
    warn_output="$("$SYNC" --refresh-mirror agentsview --authorize-remote-update --lease-sha deadbeef 2>&1 || true)"
    if echo "$warn_output" | grep -q "Lease SHA does not match"; then
        echo "  PASS: refresh-mirror rejects bad lease"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: refresh-mirror should reject bad lease"
        FAIL=$((FAIL + 1))
    fi
fi

echo
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
