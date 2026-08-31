#!/usr/bin/env bash
set -euo pipefail

api_url="${HERMES_WEB_STUDIO_URL:-http://127.0.0.1:8787}"
session_id="m1-live-$(date +%s)-$$"

cleanup() {
  curl -fsS -X DELETE "${api_url}/api/sessions/${session_id}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

health="$(curl -fsS "${api_url}/api/health/hermes")"
if ! grep -q '"reachable":true' <<<"${health}"; then
  echo "Hermes is not reachable through Hermes Web Studio: ${health}" >&2
  exit 1
fi

create="$(curl -fsS -X POST "${api_url}/api/sessions" \
  -H 'Content-Type: application/json' \
  --data "{\"session_id\":\"${session_id}\",\"title\":\"M1 live parity\"}")"
grep -q "\"session_id\":\"${session_id}\"" <<<"${create}"

detail="$(curl -fsS "${api_url}/api/sessions/${session_id}")"
grep -q '"title":"M1 live parity"' <<<"${detail}"

renamed="$(curl -fsS -X POST "${api_url}/api/sessions/${session_id}/rename" \
  -H 'Content-Type: application/json' --data '{"title":"M1 live renamed"}')"
grep -q '"title":"M1 live renamed"' <<<"${renamed}"

pinned="$(curl -fsS -X POST "${api_url}/api/sessions/${session_id}/pin" \
  -H 'Content-Type: application/json' --data '{"pinned":true}')"
grep -q '"pinned":true' <<<"${pinned}"

archived="$(curl -fsS -X POST "${api_url}/api/sessions/${session_id}/archive" \
  -H 'Content-Type: application/json' --data '{"archived":true}')"
grep -q '"archived":true' <<<"${archived}"

updated="$(curl -fsS -X PATCH "${api_url}/api/sessions/${session_id}" \
  -H 'Content-Type: application/json' --data '{"project":"m1-live","tags":["parity"]}')"
grep -q '"project":"m1-live"' <<<"${updated}"
grep -q '"tags":\["parity"\]' <<<"${updated}"

start="$(curl -fsS -X POST "${api_url}/api/chat/start" \
  -H 'Content-Type: application/json' \
  --data "{\"session_id\":\"${session_id}\",\"message\":\"Reply with exactly M1_LIVE_OK. Do not use tools.\"}")"
stream_id="$(python3 -c 'import json, sys; print(json.load(sys.stdin)["stream_id"])' <<<"${start}")"
stream="$(curl -fsS -N "${api_url}/api/chat/stream?stream_id=${stream_id}")"
grep -q 'event: done' <<<"${stream}"
if ! grep -q 'M1_LIVE_OK' <<<"${stream}"; then
  echo "Live chat completed without the expected answer marker: ${stream}" >&2
  exit 1
fi

history="$(curl -fsS "${api_url}/api/sessions/${session_id}")"
grep -q 'M1_LIVE_OK' <<<"${history}"
truncated="$(curl -fsS -X POST "${api_url}/api/sessions/${session_id}/truncate" \
  -H 'Content-Type: application/json' --data '{"count":1}')"
grep -q '"messages"' <<<"${truncated}"

echo "M1 live session/chat parity checks passed; temporary session cleanup is scheduled."
