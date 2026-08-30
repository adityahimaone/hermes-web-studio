#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)
app="$root/frontend/src/app.tsx"
rail="$root/frontend/src/components/layout/sidebar.tsx"
tooltip="$root/frontend/src/components/ui/tooltip.tsx"
dialog="$root/frontend/src/components/ui/dialog.tsx"
menu="$root/frontend/src/components/ui/dropdown-menu.tsx"
button="$root/frontend/src/components/ui/button.tsx"

rg -q 'data-testid="primary-rail"' "$rail"
rg -q 'data-testid="primary-navigation"' "$rail"
rg -q 'data-testid="titlebar"' "$app"
rg -q 'id="main-content"' "$app"
rg -q 'role="tooltip"' "$tooltip"
rg -q "aria-describedby" "$tooltip"
rg -q 'aria-modal="true"' "$dialog"
rg -q "event.key === 'Escape'" "$dialog"
rg -q 'role="menuitem"' "$menu"
rg -q "ArrowDown" "$menu"
rg -q 'focus-visible:ring-2' "$button"

echo "M7 shell contract passed."
