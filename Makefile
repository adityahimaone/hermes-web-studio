.PHONY: dev dev-api dev-ui test test-api test-ui build build-ui build-api release-gate migrate-backup migrate-restore prod-image clean

ifneq (,$(wildcard .env))
include .env
export
endif

dev:
	@$(MAKE) dev-api & api_pid=$$!; trap 'kill $$api_pid 2>/dev/null || true' EXIT INT TERM; $(MAKE) dev-ui

dev-api:
	cd backend && go run ./cmd/hermes-web-studio

dev-ui:
	cd frontend && pnpm dev

test: test-api test-ui

test-api:
	cd backend && go test ./...

test-ui:
	cd frontend && pnpm test

build: build-ui build-api

build-ui:
	cd frontend && pnpm install --frozen-lockfile && pnpm build

build-api: build-ui
	cp -R frontend/dist/. backend/internal/web/dist/
	cd backend && go build -trimpath -ldflags='-s -w' -o ../hermes-web-studio ./cmd/hermes-web-studio

release-gate:
	sh scripts/release-gate.sh

migrate-backup:
	cd backend && go run ./cmd/hermes-web-studio-migrate -state-dir "$${HERMES_WEBUI_STATE_DIR:-$$HOME/.hermes/webui}"

migrate-restore:
	cd backend && go run ./cmd/hermes-web-studio-migrate -state-dir "$${HERMES_WEBUI_STATE_DIR:-$$HOME/.hermes/webui}" -restore-latest

prod-image:
	docker build -t hermes-web-studio:local .

clean:
	rm -rf frontend/dist frontend/coverage hermes-web-studio
