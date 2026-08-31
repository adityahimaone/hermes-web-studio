#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
base_url="${BASE_URL:-http://127.0.0.1:5173}"

echo 'P056 contract/state effects'
sh "$root/scripts/check-contract-matrix.sh"
echo 'M7-M12 local Hermes/API acceptance'
sh "$root/scripts/local-hermes-acceptance.sh"
echo 'P057 visual matrix'
(cd "$root/frontend" && node e2e/p057-visual-matrix.mjs)
echo 'M11/M12 bounded shell geometry'
(cd "$root/frontend" && node e2e/m11-m12-shell-geometry.mjs)
echo 'P058 accessibility matrix'
(cd "$root/frontend" && node e2e/p058-accessibility-check.mjs)
echo 'P060 performance evidence'
sh "$root/scripts/check-performance-evidence.sh" "$root/frontend/dist"
echo 'P061 security boundary'
sh "$root/scripts/check-secret-boundary.sh"
sh "$root/scripts/check-platform-contract.sh"
echo 'P062 local release artifact checks'
sh "$root/scripts/check-release-artifacts.sh" "${BINARY:-$root/hermes-web-studio}"
echo 'P064/P066 copied-state rollback evidence'
fixture=$(mktemp -d /tmp/hermes-web-studio-acceptance.XXXXXX)
trap 'rm -rf "$fixture"' EXIT
sh "$root/scripts/sanitize-state-fixture.sh" "$fixture/source"
sh "$root/scripts/rehearse-rollback.sh" "$fixture/source" "$fixture/evidence"
echo "Local P056-P066 acceptance runners passed where locally provable. BASE_URL=$base_url"
echo 'Non-claims: P059 live Hermes, P062 hosted OS/Docker/Nix, P063 beta, P065 cutover approval, and rollback-window closure remain open.'
