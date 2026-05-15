#!/usr/bin/env sh
set -eu

DRY_RUN=0
if [ "${1:-}" = "--dry-run" ]; then
    DRY_RUN=1
fi

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
SOURCE_BIN="$SCRIPT_DIR/vial-helperd"

INSTALL_DIR="$HOME/.local/bin"
TARGET_BIN="$INSTALL_DIR/vial-helperd"

SYSTEMD_USER_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
SERVICE_PATH="$SYSTEMD_USER_DIR/vial-helper.service"
LOG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/vial-helper"
ERROR_LOG="$LOG_DIR/daemon.err.log"
MAX_LOG_SIZE=262144

run() {
    if [ "$DRY_RUN" -eq 1 ]; then
        printf '[dry-run] %s\n' "$*"
        return 0
    fi
    "$@"
}

echo
echo "========================================"
echo " Vial Helper - Linux Installer"
echo "========================================"
echo

if [ ! -f "$SOURCE_BIN" ]; then
    echo "[ERROR] vial-helperd not found next to this installer:"
    echo "$SOURCE_BIN"
    echo
    exit 1
fi

if ! command -v systemctl >/dev/null 2>&1; then
    echo "[ERROR] systemctl is required for the Linux installer."
    exit 1
fi

echo "[1/8] Stopping existing daemon..."
run systemctl --user stop vial-helper.service >/dev/null 2>&1 || true
run pkill -f "$TARGET_BIN --command run" >/dev/null 2>&1 || true
run pkill -f "vial-helperd --command run" >/dev/null 2>&1 || true

echo "[2/8] Creating install directory..."
run mkdir -p "$INSTALL_DIR"

echo "[3/8] Installing binary..."
run cp "$SOURCE_BIN" "$TARGET_BIN"
run chmod +x "$TARGET_BIN"

echo "[4/8] Initializing config..."
run "$TARGET_BIN" --command init

echo "[5/8] Creating systemd user unit..."
run mkdir -p "$LOG_DIR"
run mkdir -p "$SYSTEMD_USER_DIR"

if [ "$DRY_RUN" -eq 1 ]; then
    echo "[dry-run] write $SERVICE_PATH"
else
    cat > "$SERVICE_PATH" <<EOF
[Unit]
Description=Vial Helper daemon
After=default.target

[Service]
Type=simple
ExecStart=/bin/sh -c 'if [ -f "$1" ] && [ "$(wc -c < "$1")" -ge "$2" ]; then rm -f "$1.1"; mv "$1" "$1.1"; fi; exec "$0" --command run 1>/dev/null 2>>"$1"' "$TARGET_BIN" "$ERROR_LOG" "$MAX_LOG_SIZE"
Restart=always
RestartSec=2

[Install]
WantedBy=default.target
EOF
fi

echo "[6/8] Reloading systemd user units..."
run systemctl --user daemon-reload

echo "[7/8] Enabling autostart..."
run systemctl --user enable vial-helper.service

echo "[8/8] Starting daemon..."
run systemctl --user restart vial-helper.service

echo
echo "========================================"
echo " Installation complete"
echo "========================================"
echo
echo "Binary:"
echo "  $TARGET_BIN"
echo
echo "systemd user unit:"
echo "  $SERVICE_PATH"
echo
echo "Config and JSON files:"
echo "  ${XDG_CONFIG_HOME:-$HOME/.config}/vial-helper"
echo
echo "Error log:"
echo "  $ERROR_LOG"
echo
echo "Useful commands:"
echo "  \"$TARGET_BIN\" --command doctor"
echo "  \"$TARGET_BIN\" --command status"
echo "  systemctl --user status vial-helper.service"
echo
