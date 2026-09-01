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

The live M0 gate remains open. Re-run after starting both the Hermes Gateway and Hermes Web Studio API, then record the Hermes version, model/provider, first-token latency, and final response here. Historical live entries below are scoped to their recorded date/provider state; current provider HTTP 402 insufficient-credit responses remain blocked and are not current completion evidence.

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

## 2026-08-30 — M1 live parity probes

| Check | Result |
|---|---|
| Safe M1 session/chat runner | Pass: `scripts/m1-live-parity.sh` created an isolated session, verified load/rename/pin/archive/project/tags, completed a real no-tools chat, confirmed persisted history, truncated the transcript prefix, and cleaned up the session |
| Runs tool activity | Pass: opt-in Runs BFF emitted normalized `tool` and `tool_complete` events for a read-only `pwd` request and returned the expected safe response |
| Runs subagent activity | Inconclusive capability check: the harmless delegation prompt produced tool events and a safe response but no structured `subagent` event |
| Runs approval activity | Inconclusive safely: the prompt requesting approval for `printf approval_check` produced no structured `approval` event; no approval was submitted and no command was executed |
| Live edit/retry branch semantics | Pass: an isolated session completed an old no-tools turn, was truncated to its user-message prefix, then completed a replacement turn; persisted history contained the replacement assistant response and no old assistant response |
| Gateway capability discovery | Pass: the live Gateway advertised `run_approval`, `approval_events`, `tool_progress_events`, and delegation-capable toolsets; capability advertisement is recorded separately from observed event delivery |
| Explicit live delegation probe | Inconclusive capability check: an explicit harmless delegation request produced normalized tool events and the safe result but no structured `subagent` event |

## 2026-08-30 - M5 distribution verification

| Check | Result |
|---|---|
| Go tests, vet, embedded binary compile | Pass |
| Embedded frontend and secret scan | Pass with `scripts/check-m5-artifact.sh` |
| Backup/restore migrator | Pass: dedicated unit test restores the newest backup without in-place migration |
| Frontend lockfile production rebuild | Blocked: pnpm registry DNS failed while recreating `node_modules`; existing frontend build was validated separately |
| Docker image | Not run: Docker CLI unavailable locally |
| Nix package | Not run: Nix CLI unavailable locally |

| Performance budget script | Pass against the existing frontend/dist artifact |
| M5 browser acceptance | Ready: `node frontend/e2e/m5-browser-check.mjs`; requires Vite on port 5173 and is pending hosted/manual execution |

## 2026-08-31 - M7-M12 bounded local/live follow-up

| Check | Result |
|---|---|
| Local Hermes CLI | Pass: Hermes Agent v0.19.0 returned `HERMES_LOCAL_ACCEPTANCE` for a harmless query |
| `scripts/local-hermes-acceptance.sh` | Pass: BFF readiness, health, operator, settings, discovery, profile/provider, session, Spaces, Kanban, terminal, plugin, extension, and MCP metadata routes returned JSON |
| Diagnostics sanitization | Pass: component/count snapshots were present and credential-shaped fields were rejected |
| `scripts/smoke-hermes.sh` | Pass: local BFF/Gateway returned a completed chat with the expected marker |
| `scripts/m1-live-parity.sh` | Pass: isolated live session actions, chat, persistence, truncation, and cleanup |
| Remaining M7-M12 live gates | Not claimed: structured tool/subagent/approval events, external channels, scheduled work, hosted matrices, beta, cutover, and reference visual comparison |

The local CLI result proves provider/runtime access; the smoke and session
results prove the Studio BFF path. `hermes serve` is a separate headless
JSON-RPC backend and is not substituted for the BFF Gateway contract.
