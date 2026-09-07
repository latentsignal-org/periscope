#!/bin/bash
# Tests for install.sh version parsing and archive candidate logic.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck source=install.sh
source "$SCRIPT_DIR/install.sh"

PASS=0
FAIL=0

assert_eq() {
    local desc="$1" expected="$2" actual="$3"
    if [ "$expected" = "$actual" ]; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc"
        echo "    expected: '$expected'"
        echo "    actual:   '$actual'"
        FAIL=$((FAIL + 1))
    fi
}

MOCK_FINAL_URL=""
MOCK_CURL_EXIT=0
curl() {
    if [ "$MOCK_CURL_EXIT" != "0" ]; then
        return "$MOCK_CURL_EXIT"
    fi
    printf '%s' "$MOCK_FINAL_URL"
}
export -f curl

echo "=== get_latest_version parsing ==="

MOCK_FINAL_URL="https://github.com/diazMelgarejo/periscope/releases/tag/v0.29.2-periscope.2-657a1090"
assert_eq "periscope release tag" "v0.29.2-periscope.2-657a1090" "$(get_latest_version)"

MOCK_FINAL_URL="https://github.com/diazMelgarejo/periscope/releases/tag/v0.30.1"
assert_eq "two-digit minor" "v0.30.1" "$(get_latest_version)"

MOCK_FINAL_URL="https://github.com/diazMelgarejo/periscope/releases"
assert_eq "no releases returns empty" "" "$(get_latest_version || true)"

MOCK_FINAL_URL="https://github.com/diazMelgarejo/periscope/releases/latest"
assert_eq "unresolved latest returns empty" "" "$(get_latest_version || true)"

MOCK_FINAL_URL=""
MOCK_CURL_EXIT=22
assert_eq "curl failure returns empty" "" "$(get_latest_version || true)"
MOCK_CURL_EXIT=0

echo
echo "=== release_archive_candidates ==="

candidates="$(release_archive_candidates "v0.29.2-periscope.2" "darwin" "arm64")"
assert_eq "prefers periscope archive" "periscope_0.29.2-periscope.2_darwin_arm64.tar.gz" "$(echo "$candidates" | head -1)"
assert_eq "includes legacy agentsview archive" "agentsview_v0.29.2-periscope.2_darwin_arm64.tar.gz" "$(echo "$candidates" | tail -1)"

echo
echo "=== skip_checksum_enabled ==="

PERISCOPE_SKIP_CHECKSUM=0 AGENTSVIEW_SKIP_CHECKSUM=0
assert_eq "default off" "1" "$([ "$(skip_checksum_enabled && echo 1 || echo 0)" = "0" ] && echo 1 || echo 0)"

PERISCOPE_SKIP_CHECKSUM=1 AGENTSVIEW_SKIP_CHECKSUM=0
assert_eq "periscope env" "1" "$(skip_checksum_enabled && echo 1 || echo 0)"

PERISCOPE_SKIP_CHECKSUM=0 AGENTSVIEW_SKIP_CHECKSUM=1
assert_eq "legacy agentsview env" "1" "$(skip_checksum_enabled && echo 1 || echo 0)"

unset PERISCOPE_SKIP_CHECKSUM AGENTSVIEW_SKIP_CHECKSUM

echo
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]

bash "$SCRIPT_DIR/make_install_test.sh"
