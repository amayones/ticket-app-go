#!/usr/bin/env sh
# Hentikan app background via PID file (universal: bash/zsh/Git Bash).
# Urutan: pakai binary stop bila ada (graceful) -> kill PID tercatat.
# Tidak pernah kill by-name: hanya PID yang tercatat yang disinyal.
set -eu
. "$(dirname "$0")/common.sh"

is_windows() {
  case "$(uname -s 2>/dev/null)" in
    MINGW*|MSYS*|CYGWIN*|Windows_NT) return 0 ;;
    *) return 1 ;;
  esac
}

msg() {
  if [ "${STOP_QUIET:-0}" != "1" ]; then echo "$1"; fi
}

# 1. Cara terbaik: helper stop (graceful SIGINT + tunggu + verifikasi PID).
if [ -x "$STOP_BIN" ]; then
  exec "$STOP_BIN"
fi

# 2. Fallback: sinyal langsung ke PID di pidfile.
# Cari di semua kandidat temp dir (Git Bash: $TMPDIR bisa kosong sementara
# app Windows memakai %TEMP% -> %TEMP% != /tmp).
tmp_candidates() {
  [ -n "${TMPDIR:-}" ] && printf '%s\n' "$TMPDIR"
  printf '/tmp\n'
  if is_windows; then
    if command -v cygpath >/dev/null 2>&1; then
      [ -n "${TEMP:-}" ] && cygpath -u "$TEMP" 2>/dev/null
      [ -n "${TMP:-}" ] && [ "${TMP:-}" != "${TEMP:-}" ] && cygpath -u "$TMP" 2>/dev/null
    fi
    [ -n "${TEMP:-}" ] && printf '%s\n' "$TEMP"
    [ -n "${TMP:-}" ] && [ "${TMP:-}" != "${TEMP:-}" ] && printf '%s\n' "$TMP"
  fi
}

stopped=0
seen=""
for dir in $(tmp_candidates | sort -u); do
for pf in "$dir/go-core.pid" "$dir/golang-backend.pid"; do
  case "$seen" in *"|$pf|"*) continue ;; esac
  seen="$seen|$pf|"
  if [ ! -f "$pf" ]; then continue; fi
  pid="$(tr -d ' \r\n' < "$pf")"
  rm -f "$pf"
  case "$pid" in
    ''|*[!0-9]*)
      msg "PID file $pf tidak valid, dilewati" >&2
      continue
      ;;
  esac
  alive=0
  if is_windows; then
    if tasklist //FI "PID eq $pid" //NH 2>/dev/null | grep -qi "INFO:"; then
      : # tidak ada proses tsb
    else
      alive=1
    fi
  else
    if kill -0 "$pid" 2>/dev/null; then alive=1; fi
  fi
  if [ "$alive" -eq 1 ]; then
    if is_windows; then
      taskkill //PID "$pid" //F >/dev/null 2>&1 || true
    else
      kill "$pid" 2>/dev/null || true
    fi
    msg "Stopped PID $pid"
    stopped=1
  else
    msg "PID $pid sudah tidak jalan"
  fi
done
done

if [ "$stopped" -eq 0 ]; then
  msg "Tidak ada app yang jalan (PID file tidak ditemukan)."
  msg "Jalankan dulu: ./scripts/run.sh --hide  (lalu hentikan via ./scripts/stop.sh)"
fi
