#!/usr/bin/env sh
# Helper bersama: root proyek + nama binary lintas OS.
# Cara pakai: . "$(dirname "$0")/common.sh"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

case "$(uname -s 2>/dev/null)" in
  MINGW*|MSYS*|CYGWIN*|Windows_NT) EXE=".exe" ;;
  *) EXE="" ;;
esac
APP_BIN="$ROOT/app$EXE"
STOP_BIN="$ROOT/stop$EXE"

# Baca APP_PORT dari .env (default 1067).
app_port() {
  port="$(grep -E '^[[:space:]]*APP_PORT[[:space:]]*=' "$ROOT/.env" 2>/dev/null | tail -n 1 | cut -d= -f2 | tr -d ' \r')"
  if [ -z "$port" ]; then port="1067"; fi
  printf '%s' "$port"
}
