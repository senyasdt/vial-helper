# vial-helper

Cross-platform companion daemon for a Vial/QMK macropad with a custom Raw HID helper API.

It reads layer/layout data from firmware and writes:

- `state.json` for fast-changing active layer state
- `layout.json` for the cached full layout snapshot

## Requirements

Firmware must expose the helper Raw HID commands used by this daemon (`0x40` through `0x49`).

## Config

`config.toml` is created automatically in the OS user config directory:

- Windows: `%AppData%\vial-helper\config.toml`
- macOS: `~/Library/Application Support/vial-helper/config.toml`
- Linux: `${XDG_CONFIG_HOME:-~/.config}/vial-helper/config.toml`

`state.json` and `layout.json` are written there by default.

`VIAL_HELPER_CONFIG_DIR` overrides the config directory.

## Local Build

Recommended Go toolchain: `1.23.x`.

Use the `Makefile`:

```bash
make build win
make build linux
make build macos
```

This creates:

- `dist/local-windows`
- `dist/local-linux`
- `dist/local-macos`

Each folder contains the binary plus matching `install-*` and `uninstall-*` scripts.

Notes:

- Windows build target is `windows/amd64`
- Linux build target is `linux/amd64`
- macOS build target is `darwin/arm64`
- `go-hid` requires `cgo`
- local builds need a working C toolchain with `gcc` available in `PATH`
- for Windows local builds use `CGO_ENABLED=1`

Print embedded build info:

```bash
vial-helperd --version
```

## Install

From a release archive or from a locally built `dist/local-*` directory, run the installer that matches the platform:

Linux:

```bash
./install-linux.sh
```

macOS:

```bash
./install-macos.sh
```

Windows:

```powershell
.\install-windows.bat
```

Uninstall with the matching `uninstall-*` script.

The installer expects the binary to be next to the script in the same directory.

Background launch is quiet on all platforms:

- Linux runs as a `systemd --user` service
- macOS runs as a LaunchAgent
- Windows runs through a hidden PowerShell wrapper started by Scheduled Task

Error logs:

- Linux: `${XDG_CONFIG_HOME:-~/.config}/vial-helper/daemon.err.log`
- macOS: `~/Library/Application Support/vial-helper/daemon.err.log`
- Windows: `%APPDATA%\vial-helper\daemon.err.log`

On each daemon start, a simple rotation keeps one previous file as `daemon.err.log.1` when the current error log grows beyond `256 KB`.

## Commands

Main commands:

- `run`
- `init`
- `paths`
- `refresh-layout`
- `refresh-now`
- `doctor`
- `status`
- `version`

Examples:

```bash
vial-helperd --command run
vial-helperd --command doctor
vial-helperd --command status
vial-helperd --command version
```

Use `--raw` with `doctor` to print the decoded layout snapshot too.

## CI And Releases

GitHub Actions in [`.github/workflows/build.yml`](.github/workflows/build.yml):

- runs on every push to `master`
- runs on tags `v*`
- builds:
  - Windows x64
  - Linux x64
  - macOS ARM64
- validates install/uninstall scripts
- uploads build artifacts
- publishes release assets on tag pushes

Packaged artifacts include platform-specific install and uninstall scripts.

## Versioning

Suggested release flow:

1. Merge work into `master`.
2. CI produces snapshot artifacts automatically.
3. Create release tags as `vMAJOR.MINOR.PATCH`.

Use:

- `PATCH` for fixes
- `MINOR` for backward-compatible features
- `MAJOR` for breaking CLI, config, JSON, or protocol changes

Snapshot builds on `master` use:

```text
<last-tag-or-0.1.0>-dev.<github-run-number>+<short-sha>
```

Tag builds embed the exact tag version and publish release assets for all target platforms.
