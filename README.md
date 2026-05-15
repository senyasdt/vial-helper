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

Inspect the latest written `state.json` / `layout.json` files without talking to the device:

```bash
vial-helperd --command status
```

Use a custom config file:

```bash
vial-helperd --config /path/to/config.toml --command run
```

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
