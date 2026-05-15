#!/usr/bin/env sh
set -eu

DRY_RUN=0
if [ "${1:-}" = "--dry-run" ]; then
    DRY_RUN=1
fi

INSTALL_DIR="$HOME/.local/bin"
TARGET_BIN="$INSTALL_DIR/vial-helperd"

LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"
PLIST_PATH="$LAUNCH_AGENTS_DIR/com.vial-helper.daemon.plist"
LABEL="com.vial-helper.daemon"

run() {
    if [ "$DRY_RUN" -eq 1 ]; then
        printf '[dry-run] %s\n' "$*"
        return 0
    fi
    "$@"
}

echo
echo "========================================"
echo " Vial Helper - macOS Uninstaller"
echo "========================================"
echo

echo "[1/5] Stopping daemon..."
run launchctl bootout "gui/$(id -u)" "$PLIST_PATH" >/dev/null 2>&1 || true
run pkill -f "$TARGET_BIN --command run" >/dev/null 2>&1 || true
run pkill -f "vial-helperd --command run" >/dev/null 2>&1 || true

echo "[2/5] Removing LaunchAgent..."
if [ -f "$PLIST_PATH" ]; then
    run rm -f "$PLIST_PATH"
fi

echo "[3/5] Removing binary..."
if [ -f "$TARGET_BIN" ]; then
    run rm -f "$TARGET_BIN"
fi

echo "[4/5] Finished."
echo
echo "Removed:"
echo "  $TARGET_BIN"
echo "  $PLIST_PATH"
echo
echo "Config and JSON files were left intact:"
echo "  $HOME/Library/Application Support/vial-helper"
echo
