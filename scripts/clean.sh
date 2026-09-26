#!/usr/bin/env sh
# Hapus binary + output build frontend (universal).
set -eu
. "$(dirname "$0")/common.sh"

rm -f "$ROOT/app" "$ROOT/app.exe" "$ROOT/stop" "$ROOT/stop.exe"
rm -f "$ROOT/app~" "$ROOT/app.exe~" "$ROOT/stop~" "$ROOT/stop.exe~"
rm -rf "$ROOT/frontend/dist/assets" "$ROOT/frontend/dist/index.html"
echo "Cleaned."
