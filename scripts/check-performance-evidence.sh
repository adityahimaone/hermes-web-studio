#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist="${1:-$root/frontend/dist}"
test -f "$dist/index.html"
js_gzip=$(find "$dist/assets" -type f -name '*.js' -exec gzip -c {} + | wc -c | awk '{print $1 + 0}')
js_raw=$(find "$dist/assets" -type f -name '*.js' -exec wc -c {} + | awk 'END { print $1 + 0 }')
[ "$js_gzip" -le $((250 * 1024)) ] || { echo "initial JS gzip budget exceeded: $js_gzip > 256000" >&2; exit 1; }
rg -q 'EventSource|fetch\(' "$root/frontend/src/hooks/use-chat.ts"
if rg -n 'fetch\([^\n]*session|/api/sessions' "$root/frontend/src/hooks/use-chat.ts"; then
  echo 'chat hook appears to refetch sessions during stream' >&2
  exit 1
fi
echo "P060 local performance evidence passed: js_raw=${js_raw}B js_gzip=${js_gzip}B; RSS/startup/live token latency require a release-host run."
