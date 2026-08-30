# Hermes Web Studio

A lightweight, modern UI for Hermes Agent. This repository is a compatibility-first rewrite of [`nesquena/hermes-webui`](https://github.com/nesquena/hermes-webui) using React, TypeScript, Tailwind CSS v4, shadcn-style owned primitives, and a small Go gateway bridge.

The first vertical slice is chat. It connects the browser to a locally running Hermes Gateway without exposing the Gateway API key to the browser.

## MVP status

- Modern responsive dashboard shell
- Session-shaped chat UI and streaming assistant messages
- Go BFF with `/api/chat/start`, `/api/chat/stream`, `/api/chat/cancel`
- Hermes Gateway `/v1/chat/completions` adapter
- Translation for OpenAI chunks and Hermes native stream events
- Health endpoint and connection status
- Mock Gateway integration tests
- Detailed migration plan and agent-ready task backlog

The remaining WebUI features are deliberately represented in the navigation but disabled until their parity gates are complete. See [plan.md](./plan.md) and [TASKS.md](./TASKS.md).

## Requirements

- Node.js 20+
- pnpm 11+
- Go 1.23+
- A running Hermes Gateway/API Server, normally on `http://127.0.0.1:8642`

## Quick start

```bash
cp .env.example .env
make dev
```

`make dev` starts the Go API on port `8787` and Vite on port `5173`. Vite proxies `/api` and `/health` to Go.

Configure the gateway when it differs from the defaults:

```bash
export HERMES_WEBUI_GATEWAY_BASE_URL=http://127.0.0.1:8642
export HERMES_WEBUI_GATEWAY_API_KEY=your-gateway-key
export HERMES_WEBUI_DEFAULT_MODEL=default
make dev
```

The API key is read only by the Go server. Never use a `VITE_*` variable for secrets.

## Verification

```bash
make test
make build
```

To test against a real Hermes instance:

```bash
HERMES_WEBUI_GATEWAY_BASE_URL=http://127.0.0.1:8642 \
HERMES_WEBUI_GATEWAY_API_KEY=your-gateway-key \
./scripts/smoke-hermes.sh
```

## Architecture

```mermaid
flowchart LR
  UI["React chat UI"] -->|"POST start + SSE stream"| BFF["Go BFF :8787"]
  BFF -->|"Bearer key, server-side"| GW["Hermes Gateway :8642"]
  GW -->|"OpenAI or Hermes SSE"| BFF
  BFF -->|"normalized WebUI events"| UI
```

The BFF owns credentials, cancellation, event normalization, and future persistence. The frontend consumes a stable compatibility contract rather than coupling components directly to Gateway event variants.
