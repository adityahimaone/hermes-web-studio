# Architecture decisions

## ADR-001 — Gateway-first Hermes integration

**Status:** accepted for MVP

The Go service connects to the Hermes Gateway/API Server instead of importing the Python agent runtime or spawning the Hermes CLI. This is the cleanest runtime boundary for a non-Python rewrite and matches the bridge already supported by the original WebUI.

The adapter uses `POST /v1/chat/completions` with streaming enabled. Runs API support (`POST /v1/runs`) will be added when approval parity starts.

## ADR-002 — Compatibility BFF

The browser uses the original two-step shape: start a turn, then subscribe to a stream. The BFF normalizes Gateway-specific frames to `token`, `reasoning`, `tool`, `tool_complete`, `done`, `cancel`, and `apperror` events.

## ADR-005 - Completion de-duplication at the Gateway boundary

Gateway implementations can stream token deltas and then include the complete answer in `run.completed.output`. The adapter treats that terminal output as a completion snapshot: it emits only the missing suffix relative to already streamed text. This keeps the browser contract deterministic and avoids duplicating a response in the chat timeline.

## ADR-006 - Production process and container baseline

The Go process uses bounded request-header/read/idle timeouts and graceful SIGINT/SIGTERM shutdown. A multi-stage Docker image builds the Vite frontend, runs the Go BFF with loopback-only container binding, and exposes only Nginx on port 8080. Nginx proxies `/api` and `/health`, disables buffering for SSE, and serves the frontend fallback. Gateway credentials are supplied only as backend environment variables.

## ADR-007 - Read legacy sessions before introducing a new store

M1 starts with a read-only reader for the upstream JSON session store. It resolves `HERMES_WEBUI_STATE_DIR` and `HERMES_HOME` using the upstream precedence, prefers `_index.json` for metadata listing, and falls back to session sidecar scans. Writes and SQLite remain deferred until the session JSON, CLI database, and route behavior have separate compatibility tests.

## ADR-003 — Modern visual system, preserved composition

Pixel parity is no longer a release requirement. We preserve the three-region composition, session-first navigation, central conversation flow, optional workspace surface, and mobile navigation. Visuals use a modern neutral shadcn dashboard language powered by Tailwind v4 tokens.

## ADR-004 — Single binary remains the distribution target

The repository starts with independently served frontend and Go API for fast iteration. Before M3, the frontend production output must be embedded into the Go binary. This sequencing prevents embed mechanics from blocking the chat integration spike.

## ADR-008 - Bounded replay journal

The BFF keeps normalized SSE events in memory, assigns monotonic IDs, and replays events after `Last-Event-ID` or `after`. A completed turn remains available for five minutes so EventSource reconnect can recover its cursor, then expires. Durable session transcripts remain the source for history.

## ADR-009 - Server-owned attachment ingress

The browser uploads files to the BFF using opaque attachment IDs. The BFF validates size and MIME, writes restrictive files, and converts them to multimodal Gateway payloads. Credentials remain server-side.

## ADR-010 - Edit as transcript prefix truncation

Edit/regenerate is represented as a deterministic prefix operation against the durable JSON transcript. The browser sends a message count, the BFF truncates only the transcript array, and all other top-level legacy metadata remains intact. This avoids inventing message IDs in the persisted compatibility format.
