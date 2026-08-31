#!/bin/sh
set -eu

# Safe local acceptance for the M7-M12 server-owned surfaces. This runner only
# reads metadata and creates/removes one isolated session when the BFF is live.
# It never treats an offline Gateway as a passing live-chat result.

api_url=${HERMES_WEB_STUDIO_URL:-http://127.0.0.1:8787}
tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/hermes-web-studio-live.XXXXXX")
trap 'rm -rf "$tmp_dir"' EXIT

hermes_bin=${HERMES_BIN:-$(command -v hermes || true)}
if [ -n "$hermes_bin" ] && [ -x "$hermes_bin" ]; then
  version=$($hermes_bin --version 2>/dev/null | head -1 || true)
  [ -n "$version" ] || { echo 'Local Hermes CLI did not report a version.' >&2; exit 1; }
  marker=HERMES_LOCAL_ACCEPTANCE
  cli_output=$($hermes_bin chat --query "Reply with exactly: $marker" --quiet --no-restore-cwd --max-turns 2 2>&1)
  printf '%s' "$cli_output" | grep -Fq "$marker" || {
    echo 'Local Hermes CLI did not return the expected acceptance marker.' >&2
    exit 1
  }
  echo "PASS local Hermes CLI/provider: $version"
else
  echo 'LOCAL SKIP Hermes CLI is unavailable; BFF-only checks may still run.'
fi

get_json() {
  path=$1
  body="$tmp_dir/body"
  headers="$tmp_dir/headers"
  status=$(curl -sS -D "$headers" -o "$body" -w '%{http_code}' "$api_url$path") || {
    echo "BFF unavailable: $api_url$path" >&2
    return 2
  }
  case "$status" in
    2*) ;;
    *) echo "GET $path returned HTTP $status: $(head -c 240 "$body")" >&2; return 1 ;;
  esac
  grep -Eiq '^Content-Type: application/json([;[:space:]]|$)' "$headers" || {
    echo "GET $path did not return application/json" >&2
    return 1
  }
  node -e 'const fs=require("fs"); JSON.parse(fs.readFileSync(process.argv[1], "utf8"))' "$body" || {
    echo "GET $path returned invalid JSON: $(head -c 240 "$body")" >&2
    return 1
  }
  echo "PASS $path"
}

if ! get_json /health; then
  echo 'LOCAL SKIP Hermes Web Studio BFF is offline; no local API evidence claimed.'
  exit 0
fi
get_json /ready
get_json /api/health/hermes

for path in \
  /api/operator/health \
  /api/operator/logs \
  /api/operator/diagnostics \
  /api/operator/insights \
  /api/operator/version \
  /api/operator/update \
  /api/capabilities \
  /api/settings/capabilities \
  /api/profiles \
  /api/providers \
  /api/preferences \
  /api/skills \
  /api/memory \
  /api/spaces \
  /api/terminal \
  /api/plugins \
  /api/extensions \
  /api/mcp/servers \
  /api/kanban \
  /api/crons \
  /api/crons/history \
  /api/sessions
do
  get_json "$path"
done

diagnostics=$(curl -fsS "$api_url/api/operator/diagnostics")
printf '%s' "$diagnostics" | node -e '
  let data = "";
  process.stdin.on("data", chunk => data += chunk);
  process.stdin.on("end", () => {
    const value = JSON.parse(data);
    const serialized = JSON.stringify(value);
    for (const forbidden of ["api_key", "authorization", "password", "token", "secret"]) {
      if (serialized.toLowerCase().includes(forbidden)) {
        throw new Error(`diagnostics contains forbidden credential field: ${forbidden}`);
      }
    }
    if (!value.components || !value.counts) throw new Error("diagnostics lacks component/count snapshots");
  });
'
echo 'PASS /api/operator/diagnostics sanitized shape'

health=$(curl -fsS "$api_url/api/health/hermes")
reachable=$(printf '%s' "$health" | node -e '
  let data = "";
  process.stdin.on("data", chunk => data += chunk);
  process.stdin.on("end", () => process.stdout.write(JSON.parse(data).reachable ? "true" : "false"));
')

if [ "$reachable" = true ]; then
  sh "$(dirname "$0")/smoke-hermes.sh"
  sh "$(dirname "$0")/m1-live-parity.sh"
  echo 'LIVE PASS local Hermes smoke and session parity'
else
  echo 'LIVE SKIP Hermes Gateway is offline; P059 live chat/tool/subagent/approval rows remain unclaimed.'
fi

echo 'Local M7-M12 read-only acceptance completed.'
