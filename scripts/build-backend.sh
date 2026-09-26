#!/usr/bin/env sh
# Vet + build binary backend (universal: bash/zsh/Git Bash).
set -eu
. "$(dirname "$0")/common.sh"

VERSION="${VERSION:-$(git rev-parse --short HEAD 2>/dev/null || echo dev)}"
echo "==> go vet ./..."
cd "$ROOT"
go vet ./...
echo "==> Building Go binaries (version $VERSION)..."
go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o "$APP_BIN" .
go build -trimpath -ldflags "-s -w" -o "$STOP_BIN" ./cmd/stop
# Di Windows go build menggeser binary lama ke <nama>~ lalu menulis yang baru
# ke <nama>. Setelah build, file ~ berisi versi lama yang tidak terpakai.
rm -f "$APP_BIN~" "$STOP_BIN~"
echo "==> Done: $APP_BIN + $STOP_BIN"
echo "    Run: ./scripts/run.sh | ./scripts/start.sh | ./scripts/watch.sh"
