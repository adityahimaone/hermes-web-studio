#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

rg -q 'aria-label="Message Hermes"' "$root/frontend/src/components/chat/composer.tsx"
rg -q 'aria-label="Open navigation"' "$root/frontend/src/app.tsx"
rg -q 'aria-label="Close navigation"' "$root/frontend/src/components/layout/sidebar.tsx"
rg -q 'focus-visible:ring-2' "$root/frontend/src/components/layout/sidebar.tsx"
rg -q 'Enter.*Shift Enter' "$root/frontend/src/components/chat/composer.tsx"
rg -q 'min-h-11' "$root/frontend/src/components/layout/sidebar.tsx"
rg -q 'new EventSource\(streamUrl' "$root/frontend/src/hooks/use-chat.ts"

printf '%s\n' 'M1 UI contract checks passed.'
