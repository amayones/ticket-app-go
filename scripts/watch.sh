#!/usr/bin/env sh
# Auto rebuild tiap ada file berubah (prod-like loop, universal).
# Polling via find -newer (tanpa dependensi tambahan). Ctrl+C untuk berhenti.
set -eu
. "$(dirname "$0")/common.sh"

MARKER="$ROOT/.watch-marker"
touch "$MARKER"

cleanup() {
  echo ""
  echo "[watch] stopping app..."
  sh "$(dirname "$0")/stop.sh" >/dev/null 2>&1 || true
  rm -f "$MARKER"
  exit 0
}
trap cleanup INT TERM

rebuild_restart() {
  echo ""
  echo "[watch] change detected, rebuilding..."
  if ! sh "$(dirname "$0")/build.sh"; then
    echo "[watch] build failed, will retry on next change"
    return 0
  fi
  echo "[watch] restarting..."
  sh "$(dirname "$0")/stop.sh" >/dev/null 2>&1 || true
  sleep 1
  if [ -x "$APP_BIN" ]; then
    (cd "$ROOT" && "$APP_BIN" >/dev/null 2>&1 &)
    echo "[watch] app restarted -> http://localhost:$(app_port)/"
  fi
}

echo "Watching for changes... (Ctrl+C to stop)"
echo "Go + frontend/src -> auto rebuild + restart"
rebuild_restart

while true; do
  sleep 2
  if [ -n "$(find "$ROOT" \
      -path "$ROOT/.git" -prune -o \
      -path "$ROOT/frontend/node_modules" -prune -o \
      -path "$ROOT/frontend/dist" -prune -o \
      -type f \( -name '*.go' -o -name '*.js' -o -name '*.jsx' -o -name '*.css' -o -name '*.html' -o -name 'go.mod' -o -name 'go.sum' \) \
      -newer "$MARKER" -print -quit 2>/dev/null)" ]; then
    touch "$MARKER"
    sleep 1
    rebuild_restart
    touch "$MARKER"
  fi
done
