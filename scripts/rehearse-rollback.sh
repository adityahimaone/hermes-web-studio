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
backup="$output/rollback-backup"
restored="$output/restored"
rm -rf "$backup" "$restored"
cp -R "$source" "$backup"
chmod -R go-rwx "$backup"
printf '%s\n' '{"rollback_probe":"original","messages":[]}' > "$backup/rollback-probe.json"
cp -R "$backup" "$restored"
printf '%s\n' '{"rollback_probe":"mutated","messages":[]}' > "$restored/rollback-probe.json"
cp "$backup/rollback-probe.json" "$restored/rollback-probe.json"
cmp -s "$backup/rollback-probe.json" "$restored/rollback-probe.json"
chmod 600 "$backup/rollback-probe.json" "$restored/rollback-probe.json"
test "$(stat -f '%Lp' "$backup/rollback-probe.json" 2>/dev/null || stat -c '%a' "$backup/rollback-probe.json")" = 600
echo "P064/P066 copied-state rollback rehearsal passed: backup=$backup restored=$restored"
