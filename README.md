# vial-helper

Cross-platform companion daemon for a Vial/QMK macropad with a custom Raw HID helper API.

The daemon reads:

- active layer state;
- full dynamic Vial keymap;
- encoder actions;
- combos;
- tap dance entries;
- macros;
- key overrides;
- alt-repeat rules;
- helper capabilities.

It writes two machine-readable files for widgets and UI helpers:

- `state.json` — fast-changing active layer state;
- `layout.json` — cached snapshot of the full Vial layout and helper metadata.

## Requirements

The firmware must expose helper Raw HID commands:

- `0x40` capabilities;
- `0x42` layer state;
- `0x43` encoder action;
- `0x44` combo entry;
- `0x45` tap dance entry;
- `0x46` macros;
- `0x47` key override entry;
- `0x48` alt-repeat entry;
- `0x49` QMK settings entry.

## Config location

The daemon creates `config.toml` automatically in the OS user config directory:

- Windows: `%AppData%\\vial-helper\\config.toml`
- macOS: `~/Library/Application Support/vial-helper/config.toml`
- Linux: `${XDG_CONFIG_HOME:-~/.config}/vial-helper/config.toml`

The JSON outputs are written to the same directory by default.

`VIAL_HELPER_CONFIG_DIR` can override the config directory.

## Build

```bash
go mod tidy
go build -o dist/vial-helperd ./cmd/vial-helperd
```

On Windows:

```powershell
go mod tidy
go build -o dist\\vial-helperd.exe .\\cmd\\vial-helperd
```

Print the embedded build version:

```bash
vial-helperd --version
```

## Commands

Run daemon:

```bash
vial-helperd --command run
```

Create config and print its path:

```bash
vial-helperd --command init
```

Print resolved paths:

```bash
vial-helperd --command paths
```

Ask a running daemon to refresh `layout.json`:

```bash
vial-helperd --command refresh-layout
```

Refresh `layout.json` immediately in a one-shot process:

```bash
vial-helperd --command refresh-now
```

Run an interactive health check that replaces the old Python probe script:

```bash
vial-helperd --command doctor
```

Print the same health check plus the decoded full layout snapshot:

```bash
vial-helperd --command doctor --raw
```

Print the embedded build metadata:

```bash
vial-helperd --command version
```

Inspect the latest written `state.json` / `layout.json` files without talking to the device:

```bash
vial-helperd --command status
```

Use a custom config file:

```bash
vial-helperd --config /path/to/config.toml --command run
```

## GitHub Actions CI

The repository includes [`.github/workflows/build.yml`](.github/workflows/build.yml), which:

- runs automatically on every push to `master`;
- also runs on tags matching `v*` and on manual `workflow_dispatch`;
- builds platform artifacts for:
  - Windows x64;
  - macOS ARM64 (`macos-15`, Apple Silicon runner);
  - Linux x64;
- uploads packaged artifacts to the workflow run;
- publishes the same packaged artifacts to a GitHub Release when the push is a semver tag.

## Suggested semver flow

A simple workflow that stays close to semver:

1. Merge regular work into `master`. CI builds snapshot artifacts automatically.
2. Tag releases as `vMAJOR.MINOR.PATCH`, for example `v0.3.2`.
3. Use:
   - `PATCH` for fixes with no intended breaking behavior;
   - `MINOR` for backward-compatible new commands, fields, or capabilities;
   - `MAJOR` for breaking CLI/config/JSON/protocol changes.

On `master`, artifacts are versioned automatically as:

```text
<last-tag-or-0.1.0>-dev.<github-run-number>+<short-sha>
```

Example:

```text
0.3.2-dev.57+1a2b3c4
```

On a tag push like `v0.3.3`, the binary embeds exactly `v0.3.3` and the workflow publishes release assets for all target platforms.

## Example `state.json`

```json
{
  "label": "BASE",
  "top": 0,
  "name": "BASE",
  "effective": "0",
  "temp": "—",
  "default": "0"
}
```

## Example consumers

- YASB widget reads `state.json`.
- Tkinter/desktop overlay reads `layout.json`.
- A future macOS menu bar app can read the same files.
