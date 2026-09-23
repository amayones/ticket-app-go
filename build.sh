#!/bin/sh
set -e
echo "==> Building frontend..."
cd frontend
if [ ! -d "node_modules" ]; then npm install; fi
npm run build
cd ..
echo "==> Building Go binary (embedded frontend/dist)..."
go vet ./...
go build -o app ./...
echo "==> Done: ./app - single binary, embedded frontend"
echo "    Run: ./app  (frontend + API di satu port)"
