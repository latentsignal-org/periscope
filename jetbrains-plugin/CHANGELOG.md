<!-- Keep a Changelog guide -> https://keepachangelog.com -->

# Periscope JetBrains Plugin Changelog

## [Unreleased]

## [0.29.2-periscope.2] - 2026-05-13
### Added
- `PeriscopeProcessManager`: ref-counted singleton that starts the `periscope` binary on project
  open and stops it when the last project closes. Auto-discovers a free port from 8080.
- `MyToolWindowFactory`: JCEF-based tool window wired to `PeriscopeProcessManager.serverUrl()`
  instead of a hardcoded `localhost:5173`. Opens in the right sidebar under **Periscope**.
- `MyProjectActivity`: starts the process manager on project open; registers
  `ProjectManagerListener` to stop it on project close.
- Balloon notification group `Periscope` — shown when server is ready or binary is missing.
- `gradle.properties` version set to `0.29.2-periscope.2` matching the Go binary release.

[Unreleased]: https://github.com/diazMelgarejo/periscope/compare/v0.29.2-periscope.2-579ca34...HEAD
[0.29.2-periscope.2]: https://github.com/diazMelgarejo/periscope/releases/tag/v0.29.2-periscope.2-579ca34
