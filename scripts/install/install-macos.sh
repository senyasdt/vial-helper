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

LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"
PLIST_PATH="$LAUNCH_AGENTS_DIR/com.vial-helper.daemon.plist"
LABEL="com.vial-helper.daemon"
LOG_DIR="$HOME/Library/Application Support/vial-helper"
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
echo " Vial Helper - macOS Installer"
echo "========================================"
echo

if [ ! -f "$SOURCE_BIN" ]; then
    echo "[ERROR] vial-helperd not found next to this installer:"
    echo "$SOURCE_BIN"
    echo
    exit 1
fi

echo "[1/7] Stopping existing daemon..."
run launchctl bootout "gui/$(id -u)" "$PLIST_PATH" >/dev/null 2>&1 || true
run pkill -f "$TARGET_BIN --command run" >/dev/null 2>&1 || true
run pkill -f "vial-helperd --command run" >/dev/null 2>&1 || true

echo "[2/7] Creating install directory..."
run mkdir -p "$INSTALL_DIR"

echo "[3/7] Installing binary..."
run cp "$SOURCE_BIN" "$TARGET_BIN"
run chmod +x "$TARGET_BIN"

echo "[4/7] Initializing config..."
run "$TARGET_BIN" --command init

echo "[5/7] Creating LaunchAgent..."
run mkdir -p "$LOG_DIR"
run mkdir -p "$LAUNCH_AGENTS_DIR"

if [ "$DRY_RUN" -eq 1 ]; then
    echo "[dry-run] write $PLIST_PATH"
else
    cat > "$PLIST_PATH" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>$LABEL</string>

    <key>ProgramArguments</key>
    <array>
        <string>/bin/sh</string>
        <string>-c</string>
        <string>if [ -f "$1" ] &amp;&amp; [ "$(wc -c &lt; "$1")" -ge "$2" ]; then rm -f "$1.1"; mv "$1" "$1.1"; fi; exec "$0" --command run 1&gt;/dev/null 2&gt;&gt;"$1"</string>
        <string>$TARGET_BIN</string>
        <string>$ERROR_LOG</string>
        <string>$MAX_LOG_SIZE</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>

</dict>
</plist>
EOF
fi

echo "[6/7] Loading LaunchAgent..."
run launchctl bootstrap "gui/$(id -u)" "$PLIST_PATH"

echo "[7/7] Starting daemon..."
run launchctl kickstart -k "gui/$(id -u)/$LABEL"

echo
echo "========================================"
echo " Installation complete"
echo "========================================"
echo
echo "Binary:"
echo "  $TARGET_BIN"
echo
echo "LaunchAgent:"
echo "  $PLIST_PATH"
echo
echo "Config and JSON files:"
echo "  $HOME/Library/Application Support/vial-helper"
echo
echo "Error log:"
echo "  $ERROR_LOG"
echo
echo "Useful commands:"
echo "  \"$TARGET_BIN\" --command doctor"
echo "  \"$TARGET_BIN\" --command status"
echo
