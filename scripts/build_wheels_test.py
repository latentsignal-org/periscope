from __future__ import annotations

import io
import os
import stat
import tarfile
import zipfile
from pathlib import Path

import pytest

from build_wheels import (
    PLATFORM_MAP,
    build_all_wheels,
    build_wheel,
    extract_binary,
    normalize_wheel_version,
    parse_archive_filename,
)


# ---------------------------------------------------------------------------
# Wheel version normalization
# ---------------------------------------------------------------------------


class TestNormalizeWheelVersion:
    def test_plain_semver_unchanged(self) -> None:
        assert normalize_wheel_version("0.29.2") == "0.29.2"

    def test_commit_suffixed_tag(self) -> None:
        assert (
            normalize_wheel_version("0.29.2-periscope.2-258218f5")
            == "0.29.2+periscope.2.258218f5"
        )

    def test_prerelease_with_commit_suffix(self) -> None:
        assert (
            normalize_wheel_version("1.0.0-rc.1-abc12345")
            == "1.0.0+rc.1.abc12345"
        )


# ---------------------------------------------------------------------------
# Platform mapping
# ---------------------------------------------------------------------------


class TestPlatformMap:
    def test_all_required_platforms_present(self) -> None:
        required = {
            "linux_amd64",
            "linux_arm64",
            "darwin_amd64",
            "darwin_arm64",
            "windows_amd64",
            "windows_arm64",
        }
        assert set(PLATFORM_MAP.keys()) == required

    def test_each_entry_has_wheel_tag(self) -> None:
        for key, entry in PLATFORM_MAP.items():
            assert "wheel_tag" in entry, f"{key} missing wheel_tag"
            assert isinstance(entry["wheel_tag"], str)
            assert entry["wheel_tag"]

    def test_each_entry_has_binary_name(self) -> None:
        for key, entry in PLATFORM_MAP.items():
            assert "binary_name" in entry, f"{key} missing binary_name"
            assert isinstance(entry["binary_name"], str)
            assert entry["binary_name"]

    def test_windows_binary_has_exe_extension(self) -> None:
        assert PLATFORM_MAP["windows_amd64"]["binary_name"] == "periscope.exe"
        assert PLATFORM_MAP["windows_arm64"]["binary_name"] == "periscope.exe"

    def test_unix_binaries_have_no_extension(self) -> None:
        for key in ("linux_amd64", "linux_arm64", "darwin_amd64", "darwin_arm64"):
            assert PLATFORM_MAP[key]["binary_name"] == "periscope"

    def test_manylinux_wheel_tags(self) -> None:
        assert PLATFORM_MAP["linux_amd64"]["wheel_tag"] == "manylinux_2_28_x86_64"
        assert PLATFORM_MAP["linux_arm64"]["wheel_tag"] == "manylinux_2_28_aarch64"

    def test_macos_wheel_tags(self) -> None:
        assert PLATFORM_MAP["darwin_amd64"]["wheel_tag"] == "macosx_11_0_x86_64"
        assert PLATFORM_MAP["darwin_arm64"]["wheel_tag"] == "macosx_11_0_arm64"

    def test_windows_wheel_tag(self) -> None:
        assert PLATFORM_MAP["windows_amd64"]["wheel_tag"] == "win_amd64"
        assert PLATFORM_MAP["windows_arm64"]["wheel_tag"] == "win_arm64"


# ---------------------------------------------------------------------------
# Archive filename parsing
# ---------------------------------------------------------------------------


class TestParseArchiveFilename:
    def test_parse_linux_amd64_tar_gz(self) -> None:
        result = parse_archive_filename("periscope_0.15.0_linux_amd64.tar.gz")
        assert result == ("linux_amd64", "0.15.0")

    def test_parse_darwin_arm64_tar_gz(self) -> None:
        result = parse_archive_filename("periscope_1.2.3_darwin_arm64.tar.gz")
        assert result == ("darwin_arm64", "1.2.3")

    def test_parse_windows_amd64_zip(self) -> None:
        result = parse_archive_filename("periscope_0.15.0_windows_amd64.zip")
        assert result == ("windows_amd64", "0.15.0")

    def test_parse_windows_arm64_zip(self) -> None:
        result = parse_archive_filename("periscope_0.15.0_windows_arm64.zip")
        assert result == ("windows_arm64", "0.15.0")

    def test_parse_darwin_amd64_tar_gz(self) -> None:
        result = parse_archive_filename("periscope_2.0.0_darwin_amd64.tar.gz")
        assert result == ("darwin_amd64", "2.0.0")

    def test_unrecognized_filename_returns_none(self) -> None:
        assert parse_archive_filename("somethingelse_0.1.0_linux_amd64.tar.gz") is None

    def test_unknown_platform_returns_none(self) -> None:
        assert parse_archive_filename("periscope_0.1.0_freebsd_amd64.tar.gz") is None

    def test_no_extension_returns_none(self) -> None:
        assert parse_archive_filename("periscope_0.1.0_linux_amd64") is None

    def test_sha256sums_returns_none(self) -> None:
        assert parse_archive_filename("periscope_0.15.0_SHA256SUMS") is None

    def test_path_with_directory_uses_basename(self) -> None:
        result = parse_archive_filename(
            "releases/periscope_0.15.0_linux_arm64.tar.gz"
        )
        # parse_archive_filename only accepts basenames, so paths return None
        assert result is None

        # The caller is responsible for passing just the filename
        result = parse_archive_filename("periscope_0.15.0_linux_arm64.tar.gz")
        assert result == ("linux_arm64", "0.15.0")


# ---------------------------------------------------------------------------
# Archive extraction
# ---------------------------------------------------------------------------


def _make_targz(binary_name: str, content: bytes) -> bytes:
    """Create an in-memory .tar.gz with a single binary file."""
    buf = io.BytesIO()
    with tarfile.open(fileobj=buf, mode="w:gz") as tf:
        info = tarfile.TarInfo(name=binary_name)
        info.size = len(content)
        tf.addfile(info, io.BytesIO(content))
    return buf.getvalue()


def _make_zip(binary_name: str, content: bytes) -> bytes:
    """Create an in-memory .zip with a single binary file."""
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w") as zf:
        zf.writestr(binary_name, content)
    return buf.getvalue()


class TestExtractBinary:
    def test_extract_from_tar_gz(self, tmp_path: Path) -> None:
        content = b"fake-binary-content"
        archive = tmp_path / "periscope_0.15.0_linux_amd64.tar.gz"
        archive.write_bytes(_make_targz("periscope", content))
        result = extract_binary(archive, "periscope")
        assert result == content

    def test_extract_from_zip(self, tmp_path: Path) -> None:
        content = b"fake-binary-exe"
        archive = tmp_path / "periscope_0.15.0_windows_amd64.zip"
        archive.write_bytes(_make_zip("periscope.exe", content))
        result = extract_binary(archive, "periscope.exe")
        assert result == content

    def test_missing_binary_raises_file_not_found(self, tmp_path: Path) -> None:
        archive = tmp_path / "periscope_0.15.0_linux_amd64.tar.gz"
        archive.write_bytes(_make_targz("wrong_name", b"data"))
        with pytest.raises(FileNotFoundError, match="periscope"):
            extract_binary(archive, "periscope")

    def test_missing_binary_in_zip_raises_file_not_found(self, tmp_path: Path) -> None:
        archive = tmp_path / "periscope_0.15.0_windows_amd64.zip"
        archive.write_bytes(_make_zip("wrong.exe", b"data"))
        with pytest.raises(FileNotFoundError, match="periscope.exe"):
            extract_binary(archive, "periscope.exe")

    def test_nested_path_in_tar_gz(self, tmp_path: Path) -> None:
        """Binary may be inside a subdirectory in the archive."""
        content = b"nested-binary"
        buf = io.BytesIO()
        with tarfile.open(fileobj=buf, mode="w:gz") as tf:
            info = tarfile.TarInfo(name="periscope_0.15.0_linux_amd64/periscope")
            info.size = len(content)
            tf.addfile(info, io.BytesIO(content))
        archive = tmp_path / "periscope_0.15.0_linux_amd64.tar.gz"
        archive.write_bytes(buf.getvalue())
        result = extract_binary(archive, "periscope")
        assert result == content


# ---------------------------------------------------------------------------
# Wheel assembly
# ---------------------------------------------------------------------------


class TestBuildWheel:
    def test_wheel_filename_matches_convention(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64")
        assert whl.name == "periscope_agentsview-0.15.0-py3-none-manylinux_2_28_x86_64.whl"

    def test_wheel_is_valid_zip(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64")
        assert zipfile.is_zipfile(whl)

    def test_wheel_contains_expected_files(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64")
        with zipfile.ZipFile(whl) as zf:
            names = set(zf.namelist())
        assert "periscope/__init__.py" in names
        assert "periscope/__main__.py" in names
        assert "periscope/bin/periscope" in names
        assert "periscope_agentsview-0.15.0.dist-info/METADATA" in names
        assert "periscope_agentsview-0.15.0.dist-info/WHEEL" in names
        assert "periscope_agentsview-0.15.0.dist-info/entry_points.txt" in names
        assert "periscope_agentsview-0.15.0.dist-info/RECORD" in names

    def test_binary_has_executable_permissions(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64")
        with zipfile.ZipFile(whl) as zf:
            info = zf.getinfo("periscope/bin/periscope")
        unix_mode = (info.external_attr >> 16) & 0xFFFF
        assert oct(unix_mode & 0o777) == oct(0o755)

    def test_wheel_files_are_compressed(self, tmp_path: Path) -> None:
        whl = build_wheel(b"x" * 10000, tmp_path, "0.15.0", "linux_amd64")
        with zipfile.ZipFile(whl) as zf:
            for info in zf.infolist():
                assert info.compress_type == zipfile.ZIP_DEFLATED, (
                    f"{info.filename} uses {info.compress_type}, expected DEFLATED"
                )

    def test_binary_has_regular_file_type_bit(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64")
        with zipfile.ZipFile(whl) as zf:
            info = zf.getinfo("periscope/bin/periscope")
        # S_IFREG (0o100000) must be set so pip applies permissions
        unix_mode = (info.external_attr >> 16) & 0xFFFF
        assert unix_mode & stat.S_IFREG, "S_IFREG must be set"

    def test_windows_wheel_uses_exe_binary(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "windows_amd64")
        with zipfile.ZipFile(whl) as zf:
            names = set(zf.namelist())
        assert "periscope/bin/periscope.exe" in names
        assert "periscope/bin/periscope" not in names

    def test_metadata_required_fields(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64")
        with zipfile.ZipFile(whl) as zf:
            metadata = zf.read("periscope_agentsview-0.15.0.dist-info/METADATA").decode()
        assert "Metadata-Version: 2.1" in metadata
        assert "Name: periscope-agentsview" in metadata
        assert "Version: 0.15.0" in metadata
        assert "Requires-Python: >=3.9" in metadata
        assert "License: MIT" in metadata
        assert "Author: Wes McKinney" in metadata

    def test_wheel_file_has_root_is_purelib_false(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64")
        with zipfile.ZipFile(whl) as zf:
            wheel_meta = zf.read("periscope_agentsview-0.15.0.dist-info/WHEEL").decode()
        assert "Root-Is-Purelib: false" in wheel_meta
        assert "Generator: periscope-build-wheels" in wheel_meta

    def test_entry_points_correct(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64")
        with zipfile.ZipFile(whl) as zf:
            ep = zf.read("periscope_agentsview-0.15.0.dist-info/entry_points.txt").decode()
        assert "[console_scripts]" in ep
        assert "periscope = periscope:main" in ep

    def test_record_contains_hashes(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64")
        with zipfile.ZipFile(whl) as zf:
            record = zf.read("periscope_agentsview-0.15.0.dist-info/RECORD").decode()
        # Each non-RECORD entry should have a sha256 hash
        lines = [ln for ln in record.splitlines() if ln.strip()]
        record_line = None
        for line in lines:
            if "RECORD" in line:
                record_line = line
            else:
                assert "sha256=" in line, f"Missing hash in RECORD line: {line}"
        # RECORD itself listed as empty
        assert record_line is not None
        assert record_line.endswith(",,")

    def test_readme_included_in_metadata(self, tmp_path: Path) -> None:
        readme = "# periscope\nA great tool."
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64", readme=readme)
        with zipfile.ZipFile(whl) as zf:
            metadata = zf.read("periscope_agentsview-0.15.0.dist-info/METADATA").decode()
        assert "A great tool." in metadata

    def test_init_py_uses_execvp_on_unix(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "linux_amd64")
        with zipfile.ZipFile(whl) as zf:
            init = zf.read("periscope/__init__.py").decode()
        assert "os.execvp" in init

    def test_init_py_uses_subprocess_on_windows(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "0.15.0", "windows_amd64")
        with zipfile.ZipFile(whl) as zf:
            init = zf.read("periscope/__init__.py").decode()
        assert "subprocess.call" in init

    def test_wheel_filename_with_normalized_version(self, tmp_path: Path) -> None:
        whl = build_wheel(
            b"fake",
            tmp_path,
            "0.29.2-periscope.2-258218f5",
            "linux_amd64",
        )
        assert (
            whl.name
            == "periscope_agentsview-0.29.2+periscope.2.258218f5-py3-none-manylinux_2_28_x86_64.whl"
        )

    def test_wheel_filename_darwin_arm64(self, tmp_path: Path) -> None:
        whl = build_wheel(b"fake", tmp_path, "1.0.0", "darwin_arm64")
        assert whl.name == "periscope_agentsview-1.0.0-py3-none-macosx_11_0_arm64.whl"


# ---------------------------------------------------------------------------
# End-to-end: build_all_wheels
# ---------------------------------------------------------------------------


class TestBuildAllWheels:
    def _make_fake_archives(self, input_dir: Path, version: str) -> None:
        """Create fake release archives for every supported platform."""
        platforms = [
            ("linux_amd64", "periscope", ".tar.gz"),
            ("linux_arm64", "periscope", ".tar.gz"),
            ("darwin_amd64", "periscope", ".tar.gz"),
            ("darwin_arm64", "periscope", ".tar.gz"),
            ("windows_amd64", "periscope.exe", ".zip"),
            ("windows_arm64", "periscope.exe", ".zip"),
        ]
        for platform_key, binary_name, ext in platforms:
            content = f"binary-for-{platform_key}".encode()
            filename = f"periscope_{version}_{platform_key}{ext}"
            archive_path = input_dir / filename
            if ext == ".tar.gz":
                archive_path.write_bytes(_make_targz(binary_name, content))
            else:
                archive_path.write_bytes(_make_zip(binary_name, content))
        # Also add a SHA256SUMS file that should be skipped
        (input_dir / f"periscope_{version}_SHA256SUMS").write_text("checksums here")

    def test_produces_a_wheel_per_platform(self, tmp_path: Path) -> None:
        input_dir = tmp_path / "input"
        output_dir = tmp_path / "output"
        input_dir.mkdir()
        self._make_fake_archives(input_dir, "0.15.0")
        wheels = build_all_wheels(input_dir, output_dir, "0.15.0")
        assert len(wheels) == len(PLATFORM_MAP)

    def test_correct_wheel_names(self, tmp_path: Path) -> None:
        input_dir = tmp_path / "input"
        output_dir = tmp_path / "output"
        input_dir.mkdir()
        self._make_fake_archives(input_dir, "0.15.0")
        wheels = build_all_wheels(input_dir, output_dir, "0.15.0")
        names = {w.name for w in wheels}
        expected = {
            "periscope_agentsview-0.15.0-py3-none-manylinux_2_28_x86_64.whl",
            "periscope_agentsview-0.15.0-py3-none-manylinux_2_28_aarch64.whl",
            "periscope_agentsview-0.15.0-py3-none-macosx_11_0_x86_64.whl",
            "periscope_agentsview-0.15.0-py3-none-macosx_11_0_arm64.whl",
            "periscope_agentsview-0.15.0-py3-none-win_amd64.whl",
            "periscope_agentsview-0.15.0-py3-none-win_arm64.whl",
        }
        assert names == expected

    def test_unknown_platforms_skipped(self, tmp_path: Path) -> None:
        input_dir = tmp_path / "input"
        output_dir = tmp_path / "output"
        input_dir.mkdir()
        self._make_fake_archives(input_dir, "0.15.0")
        # Add an unknown platform archive
        unknown = input_dir / "periscope_0.15.0_freebsd_amd64.tar.gz"
        unknown.write_bytes(_make_targz("periscope", b"fake"))
        wheels = build_all_wheels(input_dir, output_dir, "0.15.0")
        assert len(wheels) == len(PLATFORM_MAP)  # unknown skipped

    def test_output_dir_created_if_missing(self, tmp_path: Path) -> None:
        input_dir = tmp_path / "input"
        output_dir = tmp_path / "output" / "nested"
        input_dir.mkdir()
        self._make_fake_archives(input_dir, "0.15.0")
        build_all_wheels(input_dir, output_dir, "0.15.0")
        assert output_dir.exists()

    def test_version_mismatch_raises(self, tmp_path: Path) -> None:
        input_dir = tmp_path / "input"
        output_dir = tmp_path / "output"
        input_dir.mkdir()
        self._make_fake_archives(input_dir, "0.15.0")
        with pytest.raises(RuntimeError, match="does not match"):
            build_all_wheels(input_dir, output_dir, "0.16.0")

    def test_all_wheels_are_valid_zips(self, tmp_path: Path) -> None:
        input_dir = tmp_path / "input"
        output_dir = tmp_path / "output"
        input_dir.mkdir()
        self._make_fake_archives(input_dir, "0.15.0")
        wheels = build_all_wheels(input_dir, output_dir, "0.15.0")
        for whl in wheels:
            assert zipfile.is_zipfile(whl), f"{whl.name} is not a valid zip"

    def test_require_all_fails_on_missing_platform(self, tmp_path: Path) -> None:
        input_dir = tmp_path / "input"
        output_dir = tmp_path / "output"
        input_dir.mkdir()
        # Only create linux_amd64
        (input_dir / "periscope_1.0.0_linux_amd64.tar.gz").write_bytes(
            _make_targz("periscope", b"fake")
        )
        with pytest.raises(RuntimeError, match="Missing archives"):
            build_all_wheels(
                input_dir, output_dir, "1.0.0", require_all=True
            )

    def test_require_all_passes_with_all_platforms(self, tmp_path: Path) -> None:
        input_dir = tmp_path / "input"
        output_dir = tmp_path / "output"
        input_dir.mkdir()
        self._make_fake_archives(input_dir, "1.0.0")
        wheels = build_all_wheels(
            input_dir, output_dir, "1.0.0", require_all=True
        )
        assert len(wheels) == len(PLATFORM_MAP)
