#!/bin/sh
set -eu

# The only allowed browser reference is the server-owned environment name in
# documentation or backend code. VITE_* must never carry a Gateway/provider key.
if rg -n --glob '!frontend/package-lock.json' --glob '!frontend/pnpm-lock.yaml' \
  'VITE_[A-Z0-9_]*(KEY|TOKEN|SECRET|PASSWORD)|import\.meta\.env\.VITE_.*(KEY|TOKEN|SECRET|PASSWORD)' \
  frontend backend; then
  echo "frontend secret boundary violation" >&2
  exit 1
fi

if rg -n --glob '!frontend/package-lock.json' --glob '!frontend/pnpm-lock.yaml' \
  "HERMES_WEBUI_GATEWAY_API_KEY\\s*[:=]\\s*[\"']|API_SERVER_KEY\\s*[:=]\\s*[\"']" \
  frontend; then
  echo "frontend contains a configured Gateway credential" >&2
  exit 1
fi

echo "frontend secret boundary passed"
