#!/usr/bin/env sh
# SATU perintah untuk pengembangan harian, cukup 1 terminal:
#   ./scripts/start.sh   (atau: task start)
# Backend (go run .) jalan di background, Vite HMR di foreground.
# Ctrl+C menghentikan keduanya.
set -eu
. "$(dirname "$0")/common.sh"

PORT="$(app_port)"
echo "==> Backend (go run .) starting in background..."
cd "$ROOT"
go run . &
BACKEND_PID=$!

# Matikan Vite yang mungkin masih hidup (kasus sinyal grup: npm selamat,
# shell sudah lanjut ke trap). Berbasis port agar tidak salah bunuh proses.
kill_frontend() {
  case "$(uname -s 2>/dev/null)" in
    MINGW*|MSYS*|CYGWIN*|Windows_NT)
      for p in $(netstat -ano 2>/dev/null | grep ':5173' | grep LISTENING | awk '{print $5}' | sort -u); do
        taskkill //PID "$p" //F >/dev/null 2>&1 || true
      done
      ;;
    *)
      pkill -f "vite" 2>/dev/null || true
      ;;
  esac
}

cleanup() {
  echo "==> Stopping backend ($BACKEND_PID)..."
  kill "$BACKEND_PID" 2>/dev/null || true
  wait "$BACKEND_PID" 2>/dev/null || true
  # Fallback: go run bisa meninggalkan child exe (terutama di Windows/Git Bash).
  STOP_QUIET=1 sh "$(dirname "$0")/stop.sh" >/dev/null 2>&1 || true
  kill_frontend
  echo "==> Stopped."
}
trap cleanup INT TERM EXIT

# Tunggu backend sehat (maks 30 detik) sebelum menyalakan Vite.
i=0
while [ "$i" -lt 30 ]; do
  i=$((i + 1))
  sleep 1
  if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
    echo "backend gagal start" >&2
    exit 1
  fi
  if curl -sf -o /dev/null "http://localhost:$PORT/healthz" 2>/dev/null; then
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "backend tidak sehat setelah 30 detik" >&2
    exit 1
  fi
done
echo "==> Backend OK di http://localhost:$PORT"

cd "$ROOT/frontend"
if [ ! -d node_modules ]; then
  echo "==> npm ci (pertama kali)..."
  npm ci --no-audit --no-fund
fi
echo "==> Frontend (Vite HMR) di http://localhost:5173 - Ctrl+C untuk berhenti"
npm run dev
