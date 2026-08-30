#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
binary="${1:-$root/hermes-web-studio}"
test -x "$binary"
sh "$root/scripts/check-m5-artifact.sh" "$binary"
test -f "$root/deploy/nginx.conf"
test -f "$root/deploy/start.sh"
echo "release artifact acceptance passed"
