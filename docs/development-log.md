# Development log

This log tracks implementation progress against `plan.md` and `TASKS.md`. A task is recorded as complete only after its stated acceptance evidence is available.

## 2026-08-30

### M0 - Real Hermes chat

- Completed the Gateway completion de-duplication fix and regression test.
- Completed graceful shutdown, HTTP timeout, and production container/reverse-proxy baseline.
- Automated evidence: backend tests, backend vet/build, frontend tests, frontend production build, and diff check pass.
- Initial live attempt was blocked because neither the BFF (`127.0.0.1:8787`) nor Hermes Gateway (`127.0.0.1:8642`) was running.
- Live gate closed later the same day: smoke passed through the running BFF, one `done` event was observed, and `HERMES_CONNECTED` was returned. First-token latency was 4588 ms. Gateway version was not exposed by health metadata.
- Planning decision: M1 may now begin, starting with session persistence inventory and compatibility tests.

### Next action

Begin the first M1 slice: session persistence inventory and compatibility tests, using the frozen upstream baseline before introducing a persistence format or route.

### M1 - Session persistence inventory

- Added `backend/internal/session.LegacySessionReader` as a read-only compatibility slice.
- Supports state directory precedence, `_index.json` metadata listing, sidecar scan fallback, full transcript loading, unknown metadata fields, and path traversal rejection.
- Added fixtures covering valid sessions, corrupt index recovery, traversal safety, override precedence, and read-only behavior.
- Automated evidence: `GOCACHE=/tmp/hermes-web-studio-go-cache go test ./...` passes for all backend packages.
- Compatibility status: `[~]`. No session write route or SQLite source of truth has been introduced.
- Next task: expose a read-only sessions API contract and add browser session loading only after route shape is agreed against the upstream inventory.

### M1 - Session continuity and rich response surface

- Added session API clients and active-session history loading in the chat hook.
- Added a keyboard-usable session history list with loading, empty, error, active, and no-overflow states.
- Added queued sends while a response is streaming, terminal event de-duplication, retry entry point, and draft editing entry point.
- Added safe Markdown links, code copy controls, tool/subagent/approval activity rendering, reasoning disclosure, usage display, and local attachment selection.
- Automated evidence: 4 frontend reducer/normalization tests and frontend production build pass.
- Compatibility status: `[~]`. Mermaid rendering, real approval actions, attachment upload/multimodal payloads, true replay reconnect, and complete session search/grouping/project/tag UI remain open.

### M1 - Read-only sessions API

- Added `GET /api/sessions` and `GET /api/sessions/{session_id}` over the legacy reader.
- Added integration coverage for compact listing, ordered message loading, missing sessions, and traversal-safe routing.
- Automated evidence: full backend tests and vet pass with host loopback access; frontend tests and production build remain green.
- Compatibility status: API contract complete, persistence remains `[~]` because all session writes and browser loading are still deferred.
- Next task: add typed frontend session loading and a recognizable session-history surface without changing the legacy source format.

### M1 - Chat parity implementation pass

- Added session CRUD/append continuity, active history loading, session search/date grouping, queueing, retry/edit entry points, rich Markdown/code/Mermaid rendering, tool/subagent/reasoning/approval/usage normalization, and loading/empty/error/streaming states.
- Added server-owned attachment upload/download with 10 MiB and MIME validation, multimodal Gateway payload conversion, Runs approval forwarding, SSE event IDs, cursor replay, terminal duplicate suppression, and bounded five-minute replay retention.
- Automated evidence: backend tests, frontend tests, frontend production build, and diff checks pass. Live approval, multimodal, network interruption replay, and browser keyboard/mobile matrix proof remain required before final M1 parity sign-off.
- Follow-up acceptance coverage added for attachment validation and multimodal payloads, replay after `Last-Event-ID`, and Runs approval forwarding. A MIME parameter bug found by the attachment test was fixed by canonicalizing detected media types.
- Frontend contract coverage expanded to six tests for session search/date grouping, usage normalization, and approval identity mapping.
- Closed the edit/regenerate persistence gap with a tested truncate route and server-side transcript prefix semantics; pushed after the full backend/frontend verification pass.
- Corrected multimodal mapping so image, PDF, and plain-text attachments use distinct Gateway content parts instead of treating every file as an image.
- Live check against the current source BFF proved attachment upload; completion was safely rejected because the isolated verification process did not have the operator Gateway credential. The existing operator process was left untouched.
- Attachment original name/MIME metadata is now persisted in a restrictive sidecar so resumed turns keep the file contract.
- Added functional mobile navigation open/close behavior and session action controls with accessible labels and focus rings; frontend test/build remain green.
- Fixed session API response serialization so legacy metadata fields are flattened to the browser contract; added integration coverage for pinned/tags metadata.
- Retry/regenerate now truncates the persisted branch before automatically re-running the selected user prompt, matching edit branch semantics.
- Session browsing now exposes project/tag metadata inline, in addition to search matching and persisted top-level fields.
- Added explicit browser-side SSE cursor filtering in addition to server replay, so reconnects cannot duplicate token/tool events when delivery is retried.
