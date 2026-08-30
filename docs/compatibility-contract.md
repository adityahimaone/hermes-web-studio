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

## M1 session persistence inventory

The frozen upstream session store is file-based and remains the compatibility source for the first M1 slice:

- State directory: `HERMES_WEBUI_STATE_DIR`, then `$HERMES_HOME/webui`, then POSIX `~/.hermes/webui`.
- Session files: `sessions/<session_id>.json`.
- Listing index: `sessions/_index.json`, with a `*.json` scan fallback when the index is missing, invalid, or empty.
- Session JSON: top-level metadata plus a `messages` array. Unknown top-level fields are retained by the reader.
- Safety: session IDs are single path components. Traversal, slash, and backslash forms are rejected.

The Go implementation uses the same JSON files as its durable session store. Writes are atomic, use `0600` files, preserve unknown top-level fields, and rebuild `_index.json` without transcript messages. SQLite or CLI session data is not treated as compatible until those formats receive their own inventory and tests.

### Session API

`GET /api/sessions` returns `{ "sessions": [...] }` with compact metadata at the top level. `GET /api/sessions/{session_id}` returns the metadata plus the original ordered `messages` array. `POST /api/sessions` creates a session, `PATCH /api/sessions/{session_id}` updates metadata, and `DELETE /api/sessions/{session_id}` removes it. Action-compatible routes are available at `/rename`, `/pin`, and `/archive`. Missing sessions return `404`; unsafe IDs return `400`. Create and update preserve the legacy top-level JSON shape and unknown fields. Chat turns append user and assistant messages to the same transcript, allowing later loads to resume the stored history.

## M1 streaming, approvals, and attachments

Start requests may include `attachment_ids`, returned by `POST /api/attachments`. Stream events carry monotonically increasing numeric SSE IDs. Clients may reconnect with `Last-Event-ID` or `after=<id>`; completed turns remain replayable for five minutes.

The normalized event set also includes `subagent`, `approval`, and `usage`. Approval decisions use `POST /api/runs/{run_id}/approval` with `{"decision":"approved"}` or `{"decision":"denied"}`. The BFF forwards this using its server-side Gateway credential.

Attachment uploads are multipart `file` requests, limited to 10 MiB and image/PDF/plain-text MIME types. Files are stored with `0600` permissions under the resolved WebUI state directory and are addressed by opaque IDs. The Gateway adapter maps images to `image_url`, PDFs to `file`, and text files to text content. Browser code never receives Gateway or provider credentials.

Editing a persisted user message uses `POST /api/sessions/{session_id}/truncate` with `{"count":<message-count>}`. The BFF keeps the transcript prefix, preserves session metadata, and the next composer send creates the replacement branch.
