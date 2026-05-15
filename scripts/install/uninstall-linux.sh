#!/usr/bin/env sh
set -eu

DRY_RUN=0
if [ "${1:-}" = "--dry-run" ]; then
    DRY_RUN=1
fi

INSTALL_DIR="$HOME/.local/bin"
TARGET_BIN="$INSTALL_DIR/vial-helperd"

SYSTEMD_USER_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
SERVICE_PATH="$SYSTEMD_USER_DIR/vial-helper.service"

run() {
    if [ "$DRY_RUN" -eq 1 ]; then
        printf '[dry-run] %s\n' "$*"
        return 0
    fi
    "$@"
}

echo
echo "========================================"
echo " Vial Helper - Linux Uninstaller"
echo "========================================"
echo

if ! command -v systemctl >/dev/null 2>&1; then
    echo "[ERROR] systemctl is required for the Linux uninstaller."
    exit 1
fi

echo "[1/5] Stopping daemon..."
run systemctl --user stop vial-helper.service >/dev/null 2>&1 || true
run pkill -f "$TARGET_BIN --command run" >/dev/null 2>&1 || true
run pkill -f "vial-helperd --command run" >/dev/null 2>&1 || true

echo "[2/5] Disabling autostart..."
run systemctl --user disable vial-helper.service >/dev/null 2>&1 || true

echo "[3/5] Removing systemd user unit..."
if [ -f "$SERVICE_PATH" ]; then
    run rm -f "$SERVICE_PATH"
fi
run systemctl --user daemon-reload

echo "[4/5] Removing binary..."
if [ -f "$TARGET_BIN" ]; then
    run rm -f "$TARGET_BIN"
fi

echo "[5/5] Finished."
echo
echo "Removed:"
echo "  $TARGET_BIN"
echo "  $SERVICE_PATH"
echo
echo "Config and JSON files were left intact:"
echo "  ${XDG_CONFIG_HOME:-$HOME/.config}/vial-helper"
echo
