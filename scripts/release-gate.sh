#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
go_cache=${GOCACHE:-/tmp/hermes-webui-release-go-cache}

cd "$root"
git diff --check
sh scripts/check-secret-boundary.sh
sh scripts/check-performance-budget.sh frontend/dist

(cd backend && GOCACHE="$go_cache" go test ./... && GOCACHE="$go_cache" go vet ./... && GOCACHE="$go_cache" go build -trimpath -ldflags='-s -w' -o ../hermes-web-studio ./cmd/hermes-web-studio)
sh scripts/check-m5-artifact.sh ./hermes-web-studio

(cd frontend && pnpm test && pnpm build)
sh scripts/check-performance-budget.sh frontend/dist
echo "release gate passed"
