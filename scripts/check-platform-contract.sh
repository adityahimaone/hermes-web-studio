#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

test -f frontend/public/manifest.webmanifest
test -f frontend/public/sw.js
test -f frontend/public/icons/hermes-studio.svg
rg -q 'serviceWorker\.register' frontend/src/main.tsx
rg -q 'VITE_BASE_PATH' frontend/vite.config.ts
rg -q 'Content-Security-Policy' backend/internal/httpapi/server.go
rg -q 'Permissions-Policy' backend/internal/httpapi/server.go deploy/nginx.conf
echo "platform contract passed"
