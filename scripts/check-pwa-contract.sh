#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
rg -q "hermes:bfcache-restore|event.persisted" frontend/src/main.tsx
rg -q "registration.update|hermes:sw-update" frontend/src/main.tsx
rg -q "VERSION = 'hermes-studio-shell-v2'|SKIP_WAITING" frontend/public/sw.js
test -f frontend/public/manifest.webmanifest
echo "PWA update and bfcache contract passed"
