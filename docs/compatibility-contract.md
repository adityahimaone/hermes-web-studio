# Compatibility contract

This document is the migration guardrail between `nesquena/hermes-webui` and Hermes Web Studio.

## Chat browser contract (implemented)

### Start

`POST /api/chat/start`

```json
{
  "session_id": "optional-stable-id",
  "message": "Hello Hermes",
  "model": "optional-model",
  "provider": "optional-provider"
}
```

Response:

```json
{"stream_id":"opaque-id","session_id":"stable-id"}
```

### Stream

`GET /api/chat/stream?stream_id=<opaque-id>` returns `text/event-stream`.

| Event | Required payload | Meaning |
|---|---|---|
| `token` | `{"text":"..."}` | Assistant text delta |
| `reasoning` | `{"text":"..."}` | Reasoning delta/summary |
| `tool` | `{"name":"...","args":{}}` | Tool started |
| `tool_complete` | `{"name":"...","is_error":false}` | Tool completed |
| `done` | `{"answer":"...","session_id":"..."}` | Successful terminal event |
| `cancel` | `{"message":"..."}` | User cancellation |
| `apperror` | `{"code":"...","message":"..."}` | Safe terminal error |

### Cancel

`POST /api/chat/cancel?stream_id=<opaque-id>` cancels the upstream request context and emits `cancel` when possible.

## Gateway contract (implemented)

- Base URL: `HERMES_WEBUI_GATEWAY_BASE_URL`, default `http://127.0.0.1:8642`
- Key: `HERMES_WEBUI_GATEWAY_API_KEY`, fallback `API_SERVER_KEY`
- Request: `POST /v1/chat/completions`
- Headers: `Authorization: Bearer …`, `X-Hermes-Session-Id`, `X-Hermes-Session-Key`
- Body: OpenAI-style `model`, `stream: true`, `messages`; optional `provider`

Supported upstream frames:

- OpenAI chunks: `choices[0].delta.content`
- Hermes `message.delta`
- Hermes `reasoning.available`
- Hermes `tool.started` / `tool.completed`
- Hermes `run.completed` / `run.failed`
- `[DONE]`

## Compatibility phases

Unimplemented routes must not be improvised. Inventory the original route, request shape, response shape, persistence effects, auth policy, and tests before implementation. Track that work in `TASKS.md`.

