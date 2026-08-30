#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
server="$root/backend/internal/httpapi/server.go"

# This is a deliberately small, reviewable inventory of the compatibility
# surface. The Go tests prove behavior; this check catches accidental route
# removal before those tests are expanded.
routes='GET /health
GET /ready
GET /api/sessions
POST /api/sessions
GET /api/sessions/{session_id}
PATCH /api/sessions/{session_id}
POST /api/sessions/{session_id}/rename
POST /api/sessions/{session_id}/pin
POST /api/sessions/{session_id}/archive
POST /api/sessions/{session_id}/truncate
POST /api/chat/start
GET /api/chat/stream
POST /api/chat/cancel
POST /api/runs/{run_id}/approval
POST /api/attachments
GET /api/workspace/tree
GET /api/workspace/preview
GET /api/workspace/download
POST /api/workspace/item
PUT /api/workspace/file
POST /api/workspace/rename
DELETE /api/workspace/item
POST /api/workspace/upload
GET /api/onboarding
POST /api/onboarding/password
POST /api/auth/login
POST /api/auth/logout
GET /api/auth/me
GET /api/profiles
POST /api/profiles/active
GET /api/providers
GET /api/control/{collection}
POST /api/control/{collection}
PATCH /api/control/{collection}/{id}
DELETE /api/control/{collection}/{id}
GET /api/preferences
PUT /api/preferences
GET /api/settings
POST /api/settings
GET /api/crons
GET /api/crons/history
POST /api/crons/create
POST /api/crons/run
POST /api/crons/pause
POST /api/crons/resume
GET /api/skills
GET /api/memory
GET /api/spaces
GET /api/operator/health
GET /api/operator/logs
GET /api/operator/insights
GET /api/capabilities
GET /api/terminal
GET /api/plugins'

missing=0
while IFS= read -r route; do
  [ -z "$route" ] && continue
  method=${route%% *}
  path=${route#* }
  pattern="HandleFunc(\"$method $path\""
  if ! grep -Fq "$pattern" "$server"; then
    echo "missing route: $route" >&2
    missing=1
  fi
done <<EOF
$routes
EOF
[ "$missing" -eq 0 ]

(cd "$root/backend" && GOCACHE="${GOCACHE:-/tmp/hermes-web-studio-contract-go-cache}" go test ./internal/httpapi ./internal/session ./internal/workspace ./internal/auth ./internal/control ./internal/gateway)
echo "P056 local contract matrix passed; live/hosted parity is not claimed."
