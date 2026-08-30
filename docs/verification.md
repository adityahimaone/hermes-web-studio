# Verification record

## 2026-08-30 — M0 local verification

Environment: Linux amd64, Go 1.27.0, Node 24.19.0, pnpm 11.19.0.

| Check | Result |
|---|---|
| `go test ./...` | Pass: Gateway parser and HTTP integration suites |
| `pnpm test` | Pass: 3 frontend event reducer tests |
| `pnpm build` | Pass |
| Frontend production output | JS 365.87 KiB / 114.27 KiB gzip; CSS 23.43 KiB / 5.24 KiB gzip |
| Real Hermes Gateway | Not available in this workspace; live gate remains open |

The backend integration suite uses an HTTP mock that implements the Hermes Gateway streaming boundary. It verifies the outgoing bearer/session headers and body, incoming SSE normalization, safe authentication failure, and upstream cancellation.

Live verification must be added as a separate dated entry using `scripts/smoke-hermes.sh`. Do not convert the live task to complete based on this mock result.

## 2026-08-30 - Live verification attempt

| Check | Result |
|---|---|
| `./scripts/smoke-hermes.sh` | Blocked: could not connect to the local BFF at `127.0.0.1:8787` |
| Hermes Gateway probe | Blocked: `127.0.0.1:8642` was not reachable |

The live M0 gate remains open. Re-run after starting both the Hermes Gateway and Hermes Web Studio API, then record the Hermes version, model/provider, first-token latency, and final response here.

## 2026-08-30 - M0 live proof

| Check | Result |
|---|---|
| Hermes Web Studio health | Pass: `configured=true`, `reachable=true` |
| Live smoke script | Pass: `./scripts/smoke-hermes.sh` |
| Model | `default` |
| Provider | Gateway default |
| First-token latency | 4588 ms, measured from stream request to first `token` event |
| Terminal event | Pass: exactly one `done` event |
| Final response | Pass: `HERMES_CONNECTED` |
| Hermes version | Not exposed by the available Gateway health metadata |

This closes the M0 live gate. The version is intentionally recorded as unavailable rather than inferred.

## 2026-08-30 - M1 automated acceptance

| Check | Result |
|---|---|
| Backend unit/integration tests | Pass: Gateway normalization, session store/API, cancellation, attachment upload and multimodal payload, approval forwarding, replay cursor |
| Backend vet and formatting | Pass |
| Frontend reducer/contract tests | Pass: 7 tests |
| Frontend production build | Pass: Vite output generated successfully |
| Secret boundary | Pass: Gateway credentials are only configured/read in Go; attachment responses expose metadata/opaque IDs only |
| M1 UI contract check | Pass: `bash scripts/check-m1-ui-contract.sh` verifies labels, keyboard copy, focus rings, mobile navigation, tap sizing, and stream URL usage |
| M1 browser acceptance | Pass: `node frontend/e2e/m1-browser-check.mjs` with temporary Chromium verifies mobile navigation, focus/labels, desktop sidebar, Shift+Enter, Enter-to-send, queued attachment delivery, and one final answer over mocked SSE |
| Browser SSE interruption/reconnect | Pass: the same Playwright script closes the first mocked SSE delivery after a token, observes EventSource reconnect, then verifies the continued reply and terminal event without duplicate prefix |
| Live approval probe | Inconclusive safely: a non-destructive `printf approval-check` request returned a textual confirmation prompt and no normalized `approval` event; no command was executed |
| Live Hermes chat | Pass: M0 smoke proof above |
| Current-source live attachment upload | Pass: multipart README upload returned opaque ID, canonical `text/plain`, original filename, and size |
| Current-source live multimodal completion | Pass: source BFF started through Makefile with configured credential, real Gateway returned a streamed README summary and exactly one `done` |
| Current-source live replay | Pass: reconnect with `Last-Event-ID: 29` replayed only event `id: 30` (`done`) |
| Opt-in live Runs API stream | Pass: `HERMES_WEBUI_USE_RUNS_API=true` returned token/reasoning events and exactly one non-duplicated `done` answer; payload-level `run.completed` output was deduplicated |
| Opt-in live Runs approval-shaped prompt | Pass: reasoning snapshot/full token duplication is suppressed; Gateway returned a textual confirmation rather than a structured approval event, so no command was run |
| Live approval interaction | Pending operator verification; safe prompt did not create a Runs approval event; tracked as `[~]` in `TASKS.md` |

## 2026-08-30 — M1 duplicate-response and queued-attachment regression

| Check | Result |
|---|---|
| Browser reconnect/finalization | Pass: terminal handling now reduces each event once; the acceptance script confirms the streamed prefix and final answer are not rendered as duplicate assistant replies |
| Queued attachment | Pass: an attachment selected while the first turn is streaming is retained in the queued turn and forwarded after the reconnecting turn completes |
| Canonical Runs activity events | Pass: fixture coverage normalizes `tool.start/progress/complete`, `subagent.start/complete`, and reasoning delta aliases without duplicate activity cards |
