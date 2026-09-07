# Periscope — JetBrains Plugin

[![Build](https://github.com/diazMelgarejo/periscope/actions/workflows/release.yml/badge.svg)](https://github.com/diazMelgarejo/periscope/releases)

Embedded session viewer for the [Periscope](https://github.com/diazMelgarejo/periscope) local AI-agent session database.

<!-- Plugin description -->
**Periscope** embeds a full session-history viewer directly inside your IDE.

- Launches the `periscope` binary automatically when a project opens and shuts it down when it closes.
- Displays the Periscope web UI in a dedicated **Periscope** tool window (right sidebar) via JCEF.
- Auto-discovers a free port starting at 8080 — no port conflicts, no manual setup.
- Shows a balloon notification when the server is ready or if the binary cannot be found.

Requires the `periscope` binary to be on `PATH` or installed in `~/.local/bin`, `~/bin`, or `/usr/local/bin`.
Install with: `curl -fsSL https://raw.githubusercontent.com/diazMelgarejo/periscope/merged/scripts/install.sh | bash`
<!-- Plugin description end -->

## Installation

- **Manually (recommended for now):**

  Download `periscope-jetbrains-plugin-*.zip` from the
  [latest release](https://github.com/diazMelgarejo/periscope/releases/latest) and install via
  <kbd>Settings</kbd> > <kbd>Plugins</kbd> > <kbd>⚙️</kbd> > <kbd>Install plugin from disk...</kbd>

## Requirements

- IntelliJ IDEA 2025.1 or later (Community or Ultimate)
- `periscope` binary installed and on `PATH`

---
Part of the [diazMelgarejo/periscope](https://github.com/diazMelgarejo/periscope) fork.
