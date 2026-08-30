#!/bin/sh
set -eu

source="${1:?usage: $0 COPIED_STATE_DIR OUTPUT_DIR}"
output="${2:?usage: $0 COPIED_STATE_DIR OUTPUT_DIR}"
case "$source:$output" in
  *"/.hermes/"*|*"/.hermes:"*|*":/.hermes"*) echo "refusing Hermes production state path" >&2; exit 1;;
esac
test -d "$source"
test "$source" != "$output"
mkdir -p "$output"
chmod 700 "$output"
if test -f "$source/control.json"; then
  sed -E 's/("(secret|token|api_key|password_hash)"[[:space:]]*:[[:space:]]*)"[^"]*"/\1"[redacted]"/Ig' "$source/control.json" > "$output/control.json"
  chmod 600 "$output/control.json"
fi
test -f "$output/control.json"
echo "copied-state rehearsal passed: $output"
