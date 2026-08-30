#!/bin/sh
set -eu

dist="${1:-frontend/dist}"
test -f "$dist/index.html"
js_bytes=$(find "$dist/assets" -type f -name '*.js' -exec wc -c {} + | awk 'END { print $1 + 0 }')
css_bytes=$(find "$dist/assets" -type f -name '*.css' -exec wc -c {} + | awk 'END { print $1 + 0 }')
max_js=$((500 * 1024))
max_css=$((100 * 1024))
[ "$js_bytes" -le "$max_js" ] || { echo "JS budget exceeded: $js_bytes > $max_js" >&2; exit 1; }
[ "$css_bytes" -le "$max_css" ] || { echo "CSS budget exceeded: $css_bytes > $max_css" >&2; exit 1; }
echo "performance budget passed: js=${js_bytes}B css=${css_bytes}B"
