.PHONY: dev dev-api dev-ui test test-api test-ui build build-ui build-api clean

dev:
	@trap 'kill 0' INT TERM EXIT; $(MAKE) dev-api & $(MAKE) dev-ui

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

build-api:
	cd backend && go build -trimpath -ldflags='-s -w' -o ../hermes-web-studio ./cmd/hermes-web-studio

clean:
	rm -rf frontend/dist frontend/coverage hermes-web-studio
