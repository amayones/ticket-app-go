#!/usr/bin/env sh
# Alias pengembangan harian -> sama dengan start.sh (1 terminal).
# (Dulu versi .ps1 membuka 2 jendela Windows; versi universal cukup 1 terminal.)
set -eu
sh "$(dirname "$0")/start.sh"
