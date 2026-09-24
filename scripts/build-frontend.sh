#!/usr/bin/env sh
# Build React -> frontend/dist (universal: bash/zsh/Git Bash).
set -eu
. "$(dirname "$0")/common.sh"

echo "==> Building frontend..."
cd "$ROOT/frontend"
if [ ! -d node_modules ]; then
  npm ci --no-audit --no-fund
fi
npm run build
# vite emptyOutDir menghapus placeholder embed; buat ulang agar
# //go:embed all:frontend/dist tetap compile di fresh clone.
printf '*\n!.gitignore\n' > "$ROOT/frontend/dist/.gitignore"
echo "==> Frontend done: frontend/dist"
