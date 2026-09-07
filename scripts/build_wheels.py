from __future__ import annotations

"""Build Python wheels for PyPI distribution from pre-built periscope binaries.

Takes release archives (tar.gz/zip) and packages them into platform-specific
Python wheels that can be uploaded to PyPI.
"""

import argparse
import base64
import hashlib
import io
import os
import re
import stat
import sys
import tarfile
import zipfile
from pathlib import Path

# ---------------------------------------------------------------------------
# Version helpers
# ---------------------------------------------------------------------------


def normalize_wheel_version(version: str) -> str:
    """Convert a release tag version to a PEP 440-compatible wheel version.

    Our release tags follow the convention:
        {semver}-{suffix}-{commit}  →  0.29.2-periscope.2-258218f5

    Wheel filenames use ``-`` as a field delimiter, so the version field must
    not contain dashes.  PEP 440 local identifiers (``+``) are the standard
    escape hatch:

        0.29.2-periscope.2-258218f5  →  0.29.2+periscope.2.258218f5

    A plain semver (e.g. ``0.29.2``) is returned as-is.
    """
    m = re.match(r"^(\d+\.\d+\.\d+)-(.+)$", version)
    if m:
        local = m.group(2).replace("-", ".")
        return f"{m.group(1)}+{local}"
    return version


# ---------------------------------------------------------------------------
# Package identity
# ---------------------------------------------------------------------------

# PyPI project name. The upstream name "periscope" is registered by an
# unrelated project (a subtitle downloader), so this fork publishes under
# a distinct distribution name. The installed console script and the
# internal module/import name stay "periscope" -- only the PyPI project
# identity (and therefore the wheel filename and dist-info dir) differ.
PYPI_DIST_NAME = "periscope-agentsview"
PYPI_DIST_NAME_NORMALIZED = PYPI_DIST_NAME.replace("-", "_")

# ---------------------------------------------------------------------------
# Platform constants
# ---------------------------------------------------------------------------

PLATFORM_MAP: dict[str, dict[str, str]] = {
    "linux_amd64": {
        "wheel_tag": "manylinux_2_28_x86_64",
        "binary_name": "periscope",
    },
    "linux_arm64": {
        "wheel_tag": "manylinux_2_28_aarch64",
        "binary_name": "periscope",
    },
    "darwin_amd64": {
        "wheel_tag": "macosx_11_0_x86_64",
        "binary_name": "periscope",
    },
    "darwin_arm64": {
        "wheel_tag": "macosx_11_0_arm64",
        "binary_name": "periscope",
    },
    "windows_amd64": {
        "wheel_tag": "win_amd64",
        "binary_name": "periscope.exe",
    },
    "windows_arm64": {
        "wheel_tag": "win_arm64",
        "binary_name": "periscope.exe",
    },
}

_ARCHIVE_RE = re.compile(
    r"^periscope_(?P<version>[^_]+)_(?P<platform>[^.]+)\.(?:tar\.gz|zip)$"
)

# ---------------------------------------------------------------------------
# Filename parsing
# ---------------------------------------------------------------------------


def parse_archive_filename(filename: str) -> tuple[str, str] | None:
    """Parse a release archive filename into (platform_key, version).

    Recognizes filenames of the form:
        periscope_<version>_<platform>.(tar.gz|zip)

    Returns None for unrecognized filenames or unknown platforms.
    """
    m = _ARCHIVE_RE.match(filename)
    if m is None:
        return None
    platform_key = m.group("platform")
    if platform_key not in PLATFORM_MAP:
        return None
    return (platform_key, m.group("version"))


# ---------------------------------------------------------------------------
# Archive extraction
# ---------------------------------------------------------------------------


def extract_binary(archive_path: Path, binary_name: str) -> bytes:
    """Extract a named binary from a .tar.gz or .zip archive.

    Searches for an entry whose basename matches binary_name (handles
    nested paths inside the archive).

    Raises FileNotFoundError if the binary is not found.
    """
    suffix = archive_path.name
    if suffix.endswith(".tar.gz"):
        return _extract_from_targz(archive_path, binary_name)
    if suffix.endswith(".zip"):
        return _extract_from_zip(archive_path, binary_name)
    raise ValueError(f"Unsupported archive format: {archive_path}")


def _extract_from_targz(archive_path: Path, binary_name: str) -> bytes:
    with tarfile.open(archive_path, "r:gz") as tf:
        for member in tf.getmembers():
            if os.path.basename(member.name) == binary_name:
                f = tf.extractfile(member)
                if f is not None:
                    return f.read()
    raise FileNotFoundError(
        f"Binary '{binary_name}' not found in {archive_path}"
    )


def _extract_from_zip(archive_path: Path, binary_name: str) -> bytes:
    with zipfile.ZipFile(archive_path) as zf:
        for name in zf.namelist():
            if os.path.basename(name) == binary_name:
                return zf.read(name)
    raise FileNotFoundError(
        f"Binary '{binary_name}' not found in {archive_path}"
    )


# ---------------------------------------------------------------------------
# Wheel assembly
# ---------------------------------------------------------------------------

_INIT_PY_UNIX = """\
from __future__ import annotations

import os
import stat
import sys
from pathlib import Path


def main() -> None:
    bin_path = str(Path(__file__).parent / "bin" / "periscope")
    mode = os.stat(bin_path).st_mode
    if not (mode & stat.S_IXUSR):
        os.chmod(bin_path, mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)
    os.execvp(bin_path, [bin_path] + sys.argv[1:])
"""

_INIT_PY_WINDOWS = """\
from __future__ import annotations

import subprocess
import sys
from pathlib import Path


def main() -> None:
    bin_path = Path(__file__).parent / "bin" / "periscope.exe"
    sys.exit(subprocess.call([str(bin_path)] + sys.argv[1:]))
"""

_MAIN_PY = """\
from periscope import main

main()
"""


def _sha256_record_hash(data: bytes) -> str:
    """Return a url-safe base64 sha256 hash in RECORD format."""
    digest = hashlib.sha256(data).digest()
    return "sha256=" + base64.urlsafe_b64encode(digest).rstrip(b"=").decode()


def build_wheel(
    binary_content: bytes,
    output_dir: Path,
    version: str,
    platform_key: str,
    readme: str | None = None,
) -> Path:
    """Build a Python wheel containing the periscope binary.

    Args:
        binary_content: Raw bytes of the platform binary.
        output_dir: Directory where the wheel file will be written.
        version: Package version string (e.g. "0.15.0").
        platform_key: One of the keys in PLATFORM_MAP.
        readme: Optional README content to embed in METADATA.

    Returns:
        Path to the created .whl file.
    """
    platform_info = PLATFORM_MAP[platform_key]
    wheel_tag = platform_info["wheel_tag"]
    binary_name = platform_info["binary_name"]
    is_windows = platform_key.startswith("windows")

    # Normalize version for PEP 440 compliance (dashes → local identifier)
    whl_version = normalize_wheel_version(version)
    dist_info = f"{PYPI_DIST_NAME_NORMALIZED}-{whl_version}.dist-info"
    whl_name = f"{PYPI_DIST_NAME_NORMALIZED}-{whl_version}-py3-none-{wheel_tag}.whl"

    output_dir.mkdir(parents=True, exist_ok=True)
    whl_path = output_dir / whl_name

    init_py = _INIT_PY_WINDOWS if is_windows else _INIT_PY_UNIX
    main_py = _MAIN_PY
    metadata = _build_metadata(whl_version, readme)
    wheel_meta = _build_wheel_file(wheel_tag)
    entry_points = "[console_scripts]\nperiscope = periscope:main\n"

    # Collect (arcname, data) pairs to write, then build RECORD
    entries: list[tuple[str, bytes, int]] = []

    def _add(arcname: str, data: bytes, unix_mode: int = 0o644) -> None:
        entries.append((arcname, data, unix_mode))

    _add("periscope/__init__.py", init_py.encode())
    _add("periscope/__main__.py", main_py.encode())
    _add(f"periscope/bin/{binary_name}", binary_content, 0o755)
    _add(f"{dist_info}/METADATA", metadata.encode())
    _add(f"{dist_info}/WHEEL", wheel_meta.encode())
    _add(f"{dist_info}/entry_points.txt", entry_points.encode())

    # Build RECORD content (RECORD itself is listed as empty)
    record_lines: list[str] = []
    for arcname, data, _ in entries:
        record_lines.append(f"{arcname},{_sha256_record_hash(data)},{len(data)}")
    record_lines.append(f"{dist_info}/RECORD,,")
    record_content = "\n".join(record_lines) + "\n"

    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        for arcname, data, unix_mode in entries:
            info = zipfile.ZipInfo(arcname)
            info.compress_type = zipfile.ZIP_DEFLATED
            # Include S_IFREG so pip recognizes the file type and
            # applies permissions (including +x) during installation
            info.external_attr = (stat.S_IFREG | unix_mode) << 16
            zf.writestr(info, data)
        # Write RECORD last
        record_info = zipfile.ZipInfo(f"{dist_info}/RECORD")
        record_info.compress_type = zipfile.ZIP_DEFLATED
        record_info.external_attr = (stat.S_IFREG | 0o644) << 16
        zf.writestr(record_info, record_content.encode())

    whl_path.write_bytes(buf.getvalue())
    return whl_path


def _build_metadata(version: str, readme: str | None) -> str:
    lines = [
        "Metadata-Version: 2.1",
        f"Name: {PYPI_DIST_NAME}",
        f"Version: {version}",
        "Summary: Local web viewer for AI agent sessions",
        "Home-page: https://github.com/latentsignal-org/periscope",
        "Author: Wes McKinney",
        "License: MIT",
        "Requires-Python: >=3.9",
        "Classifier: License :: OSI Approved :: MIT License",
        "Classifier: Programming Language :: Python :: 3",
    ]
    if readme is not None:
        lines.append("Description-Content-Type: text/markdown")
        lines.append("")
        lines.append(readme)
    return "\n".join(lines) + "\n"


def _build_wheel_file(wheel_tag: str) -> str:
    return (
        "Wheel-Version: 1.0\n"
        "Generator: periscope-build-wheels\n"
        "Root-Is-Purelib: false\n"
        f"Tag: py3-none-{wheel_tag}\n"
    )


# ---------------------------------------------------------------------------
# Top-level build orchestration
# ---------------------------------------------------------------------------


EXPECTED_PLATFORMS = frozenset(PLATFORM_MAP.keys())


def build_all_wheels(
    input_dir: Path,
    output_dir: Path,
    version: str,
    readme: str | None = None,
    require_all: bool = False,
) -> list[Path]:
    """Scan input_dir for release archives and build a wheel for each.

    Args:
        input_dir: Directory containing release archives.
        output_dir: Directory where wheel files will be written.
        version: Version string to embed in wheels (overrides archive version).
        readme: Optional README content to embed in METADATA.
        require_all: If True, raise if any expected platform is missing.

    Returns:
        List of paths to the created wheel files.
    """
    output_dir.mkdir(parents=True, exist_ok=True)
    wheels: list[Path] = []
    found_platforms: set[str] = set()

    for archive_path in sorted(input_dir.iterdir()):
        parsed = parse_archive_filename(archive_path.name)
        if parsed is None:
            continue
        platform_key, archive_version = parsed
        if archive_version != version:
            raise RuntimeError(
                f"{archive_path.name}: archive version {archive_version}"
                f" does not match --version {version}"
            )
        found_platforms.add(platform_key)
        binary_name = PLATFORM_MAP[platform_key]["binary_name"]
        binary_content = extract_binary(archive_path, binary_name)
        whl = build_wheel(binary_content, output_dir, version, platform_key, readme)
        wheels.append(whl)

    if require_all:
        missing = EXPECTED_PLATFORMS - found_platforms
        if missing:
            raise RuntimeError(
                f"Missing archives for platforms: {', '.join(sorted(missing))}"
            )

    return wheels


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def _parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Build Python wheels from pre-built periscope binaries."
    )
    parser.add_argument("--version", required=True, help="Package version (e.g. 0.15.0)")
    parser.add_argument(
        "--input-dir",
        required=True,
        type=Path,
        help="Directory containing release archives",
    )
    parser.add_argument(
        "--output-dir",
        default=Path("dist"),
        type=Path,
        help="Directory to write wheels to (default: ./dist)",
    )
    parser.add_argument(
        "--readme",
        type=Path,
        default=None,
        help="Path to README.md to embed in wheel METADATA",
    )
    parser.add_argument(
        "--require-all",
        action="store_true",
        help="Fail if any expected platform archive is missing",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> None:
    args = _parse_args(argv)
    readme: str | None = None
    if args.readme is not None:
        readme = args.readme.read_text(encoding="utf-8")

    wheels = build_all_wheels(
        args.input_dir,
        args.output_dir,
        args.version,
        readme,
        require_all=args.require_all,
    )
    if not wheels:
        print("Error: no wheels were built", file=sys.stderr)
        sys.exit(1)
    for whl in wheels:
        print(whl)
    print(f"Built {len(wheels)} wheel(s).")


if __name__ == "__main__":
    main()
