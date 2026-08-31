#!/usr/bin/env bash
set -euo pipefail

api_url="${HERMES_WEB_STUDIO_URL:-http://127.0.0.1:8787}"

health="$(curl -fsS "${api_url}/api/health/hermes")"
if ! grep -q '"reachable":true' <<<"${health}"; then
  echo "Hermes is not reachable through Hermes Web Studio: ${health}" >&2
  exit 1
fi

start="$(curl -fsS -X POST "${api_url}/api/chat/start" \
  -H 'Content-Type: application/json' \
  --data '{"session_id":"live-smoke","message":"Reply with exactly: HERMES_CONNECTED"}')"
stream_id="$(sed -n 's/.*"stream_id":"\([^"]*\)".*/\1/p' <<<"${start}")"
if [[ -z "${stream_id}" ]]; then
  echo "Chat did not return a stream_id: ${start}" >&2
  exit 1
fi

result="$(curl -fsS -N "${api_url}/api/chat/stream?stream_id=${stream_id}")"
if ! grep -q 'event: done' <<<"${result}"; then
  echo "Hermes stream did not complete: ${result}" >&2
  exit 1
fi
if ! grep -q 'HERMES_CONNECTED' <<<"${result}"; then
  echo "Hermes stream completed without the expected answer marker: ${result}" >&2
  exit 1
fi

echo "Live Hermes chat completed successfully."
