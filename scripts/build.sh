#!/usr/bin/env sh
# Build penuh: frontend + backend -> single binary (universal).
set -eu
sh "$(dirname "$0")/build-frontend.sh"
sh "$(dirname "$0")/build-backend.sh"
