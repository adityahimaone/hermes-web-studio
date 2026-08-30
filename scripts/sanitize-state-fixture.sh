#!/bin/sh
set -eu

output="${1:-/tmp/hermes-web-studio-fixture}"
case "$output" in
  */.hermes/*|*/.hermes) echo "refusing Hermes production state path" >&2; exit 1;;
esac
mkdir -p "$output"
chmod 700 "$output"
cat > "$output/control.json" <<'EOF'
{"tasks":[],"todos":[],"goals":[],"spaces":[],"preferences":{"theme":"dark","skin":"default"},"task_history":[]}
EOF
cat > "$output/auth.json" <<'EOF'
{"password_hash":"sha256$fixture-only","secret":"fixture-secret-not-a-production-credential"}
EOF
chmod 600 "$output/control.json" "$output/auth.json"
printf '%s\n' "sanitized fixture created at $output"
