#!/usr/bin/env sh
# Build + jalankan binary di foreground (universal).
#   ./scripts/run.sh [--hide]   (--hide = sembunyikan console, Windows)
set -eu
. "$(dirname "$0")/common.sh"

sh "$(dirname "$0")/build.sh"
echo "==> Running $APP_BIN ..."
exec "$APP_BIN" "$@"
