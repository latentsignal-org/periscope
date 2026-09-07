#!/bin/bash
# periscope installer — diazMelgarejo/periscope fork
# Usage: curl -fsSL https://raw.githubusercontent.com/diazMelgarejo/periscope/merged/scripts/install.sh | bash
#
# Installs the Periscope product binary. Accepts legacy agentsview release
# archive names and environment variables for compatibility during transition.

set -euo pipefail

REPO="diazMelgarejo/periscope"
BINARY_NAME="periscope"
LEGACY_BINARY_NAME="agentsview"
INSTALL_LEGACY_SYMLINK="${PERISCOPE_LEGACY_SYMLINK:-1}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() { echo -e "${GREEN}$1${NC}"; }
warn() { echo -e "${YELLOW}$1${NC}"; }
error() { echo -e "${RED}$1${NC}" >&2; exit 1; }

skip_checksum_enabled() {
    [ "${PERISCOPE_SKIP_CHECKSUM:-0}" = "1" ] || [ "${AGENTSVIEW_SKIP_CHECKSUM:-0}" = "1" ]
}

detect_os() {
    case "$(uname -s)" in
        Darwin) echo "darwin" ;;
        Linux) echo "linux" ;;
        *) error "Unsupported OS: $(uname -s). periscope supports macOS and Linux." ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

find_install_dir() {
    if [ -w "/usr/local/bin" ]; then
        echo "/usr/local/bin"
    else
        mkdir -p "$HOME/.local/bin"
        echo "$HOME/.local/bin"
    fi
}

download() {
    local url="$1"
    local output="$2"
    if command -v curl &>/dev/null; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget &>/dev/null; then
        wget -q "$url" -O "$output"
    else
        error "Neither curl nor wget found"
    fi
}

get_latest_version() {
    local url="https://github.com/${REPO}/releases/latest"
    local final_url=""
    if command -v curl &>/dev/null; then
        final_url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$url") || return 1
    elif command -v wget &>/dev/null; then
        final_url=$(wget --spider -S "$url" 2>&1 \
            | awk 'tolower($1)=="location:" {print $2}' \
            | tail -1 \
            | tr -d '\r\n') || return 1
    else
        return 1
    fi
    case "$final_url" in
        */releases/tag/*) echo "${final_url##*/releases/tag/}" ;;
        *) return 1 ;;
    esac
}

release_archive_candidates() {
    local version="$1"
    local os="$2"
    local arch="$3"
    local platform="${os}_${arch}"
    local version_no_v="${version#v}"
    printf '%s\n' \
        "${BINARY_NAME}_${version_no_v}_${platform}.tar.gz" \
        "${BINARY_NAME}_${version}_${platform}.tar.gz" \
        "${LEGACY_BINARY_NAME}_${version_no_v}_${platform}.tar.gz" \
        "${LEGACY_BINARY_NAME}_${version}_${platform}.tar.gz"
}

verify_checksum() {
    local file="$1"
    local checksums_file="$2"
    local filename="$3"

    if skip_checksum_enabled; then
        warn "Checksum verification skipped (PERISCOPE_SKIP_CHECKSUM or AGENTSVIEW_SKIP_CHECKSUM=1)"
        return 0
    fi

    if [ ! -f "$checksums_file" ]; then
        error "Checksum file not available. Set PERISCOPE_SKIP_CHECKSUM=1 to bypass."
    fi

    local expected
    expected=$(awk -v f="$filename" '{gsub(/^\*/, "", $2); if ($2==f) {print $1; exit}}' "$checksums_file")
    if [ -z "$expected" ]; then
        error "No checksum found for $filename in SHA256SUMS"
    fi

    local actual
    if command -v sha256sum &>/dev/null; then
        actual=$(sha256sum "$file" | cut -d' ' -f1)
    elif command -v shasum &>/dev/null; then
        actual=$(shasum -a 256 "$file" | cut -d' ' -f1)
    else
        error "No sha256 tool available. Install coreutils or set PERISCOPE_SKIP_CHECKSUM=1 to bypass."
    fi

    if [ "$expected" != "$actual" ]; then
        error "Checksum verification failed!\n  Expected: $expected\n  Actual:   $actual"
    fi

    info "Checksum verified"
}

extract_binary_name() {
    local dir="$1"
    if [ -f "$dir/${BINARY_NAME}" ]; then
        echo "$BINARY_NAME"
    elif [ -f "$dir/${LEGACY_BINARY_NAME}" ]; then
        echo "$LEGACY_BINARY_NAME"
    else
        return 1
    fi
}

install_binary() {
    local extracted_name="$1"
    local install_dir="$2"
    local dest="$install_dir/${BINARY_NAME}"

    if [ -w "$install_dir" ]; then
        mv "$extracted_name" "$dest"
    else
        sudo mv "$extracted_name" "$dest"
    fi
    chmod +x "$dest"

    if [ "$INSTALL_LEGACY_SYMLINK" = "1" ] && [ "$BINARY_NAME" != "$LEGACY_BINARY_NAME" ]; then
        local legacy_path="$install_dir/${LEGACY_BINARY_NAME}"
        if [ -w "$install_dir" ]; then
            ln -sf "$BINARY_NAME" "$legacy_path"
        else
            sudo ln -sf "$BINARY_NAME" "$legacy_path"
        fi
    fi
}

install_from_release() {
    local os="$1"
    local arch="$2"
    local install_dir="$3"

    info "Fetching latest release..."
    local version
    version=$(get_latest_version)

    if [ -z "$version" ]; then
        return 1
    fi

    info "Found version: $version"

    local base_url="https://github.com/${REPO}/releases/download/${version}"
    local tmpdir
    tmpdir=$(mktemp -d)
    trap "rm -rf $tmpdir" EXIT

    local filename=""
    while IFS= read -r candidate; do
        [ -n "$candidate" ] || continue
        info "Trying ${candidate}..."
        if download "${base_url}/${candidate}" "$tmpdir/release.tar.gz"; then
            filename="$candidate"
            break
        fi
    done < <(release_archive_candidates "$version" "$os" "$arch")

    if [ -z "$filename" ]; then
        return 1
    fi

    if ! skip_checksum_enabled; then
        if ! download "${base_url}/SHA256SUMS" "$tmpdir/SHA256SUMS"; then
            error "Failed to download SHA256SUMS. Cannot verify binary integrity."
        fi
        verify_checksum "$tmpdir/release.tar.gz" "$tmpdir/SHA256SUMS" "$filename"
    else
        warn "Checksum verification skipped"
    fi

    info "Extracting..."
    tar -xzf "$tmpdir/release.tar.gz" -C "$tmpdir"

    local extracted_name
    extracted_name=$(extract_binary_name "$tmpdir") || error "Binary not found in archive"
    install_binary "$tmpdir/$extracted_name" "$install_dir"

    if [ "$os" = "darwin" ] && [ -f "$install_dir/${BINARY_NAME}" ]; then
        codesign -s - "$install_dir/${BINARY_NAME}" 2>/dev/null || true
    fi

    return 0
}

main() {
    info "Installing periscope..."
    echo

    local os arch install_dir
    os=$(detect_os)
    arch=$(detect_arch)
    install_dir=$(find_install_dir)

    info "Platform: ${os}/${arch}"
    info "Install directory: ${install_dir}"
    echo

    if install_from_release "$os" "$arch" "$install_dir"; then
        info "Installed from GitHub release"
    else
        error "Installation failed. Please check https://github.com/${REPO}/releases for available builds."
    fi

    echo
    info "Installation complete!"
    echo

    if ! echo "$PATH" | grep -q "$install_dir"; then
        warn "Add this to your shell profile:"
        echo "  export PATH=\"\$PATH:$install_dir\""
        echo
    fi

    echo "Get started:"
    echo "  periscope serve    # Start the server and open browser"
    echo "  periscope update   # Check for and install updates"
    if [ "$INSTALL_LEGACY_SYMLINK" = "1" ]; then
        echo "  agentsview serve   # Legacy compatibility symlink"
    fi
}

if [[ "${BASH_SOURCE[0]-}" == "${0}" || -z "${BASH_SOURCE[0]-}" ]]; then
    main "$@"
fi
