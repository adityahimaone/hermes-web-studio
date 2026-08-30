#!/bin/sh
set -eu
binary="${1:-./hermes-web-studio}"
test -x "$binary"
strings "$binary" | grep -q 'Build the frontend with' || { echo "embedded frontend marker missing" >&2; exit 1; }
if [ -n "${HERMES_WEBUI_GATEWAY_API_KEY:-}" ] && strings "$binary" | grep -Fq "$HERMES_WEBUI_GATEWAY_API_KEY"; then echo "configured gateway credential found" >&2; exit 1; fi
echo "embedded frontend and secret scan passed"
