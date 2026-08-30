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
- Live check against the current source BFF proved attachment upload and a real multimodal completion; the BFF was started on port 8788 through the Makefile so the configured credential was inherited without replacing the operator process on port 8787.
- Live replay check with `Last-Event-ID` returned only the missed terminal event, confirming the bounded cursor path against the running Gateway-backed turn.
- Attachment original name/MIME metadata is now persisted in a restrictive sidecar so resumed turns keep the file contract.
- Added functional mobile navigation open/close behavior and session action controls with accessible labels and focus rings; frontend test/build remain green.
- Fixed session API response serialization so legacy metadata fields are flattened to the browser contract; added integration coverage for pinned/tags metadata.
- Retry/regenerate now truncates the persisted branch before automatically re-running the selected user prompt, matching edit branch semantics.
- Session browsing now exposes project/tag metadata inline, in addition to search matching and persisted top-level fields.
- Added explicit browser-side SSE cursor filtering in addition to server replay, so reconnects cannot duplicate token/tool events when delivery is retried.
- Resolved-state inventory confirms chat persistence is JSON under `~/.hermes/webui`; the separate `kanban.db` is unrelated to WebUI chat sessions, so no SQLite/CLI bridge is required for this M1 source contract.
- Added a repeatable static UI contract check for required labels, keyboard guidance, focus styles, mobile navigation, tap sizing, and EventSource wiring; it complements but does not replace browser-level verification.
- Added official `@playwright/test` dev tooling and a browser acceptance script; Chromium verification passes for mobile navigation, focus/labels, desktop layout, Shift+Enter, and Enter-to-send against mocked SSE.
- Extended browser coverage to simulate an interrupted SSE connection and verify automatic EventSource reconnect, cursor progression, continued output, and no duplicated prefix.
- Live approval probe used only a non-destructive `printf` request; the configured Gateway returned a textual confirmation prompt rather than a Runs approval event, so no command was executed and the approval item remains `[~]`.
- Local Gateway contract inspection confirmed `POST /v1/runs` is exposed, while the safe live prompt did not create an approval run; the BFF forwarding route remains tested and live approval interaction remains intentionally unclaimed.
- Aligned approval choices with the Gateway Runs contract (`once`, `session`, `always`, `deny`) and retained backwards-compatible approved/denied aliases; the UI now exposes all four choices.
- Added upstream approval event aliases `approval.request` and `hermes.approval.request` to the normalizer.
- Added an opt-in Runs API adapter (`HERMES_WEBUI_USE_RUNS_API`) for structured run events and approval-capable sessions while preserving the existing chat-completions default.
- Live Runs API regression exposed and fixed payload-level `run.completed` duplicate output; the corrected source returned one `RUNS_FIXED` answer and one terminal `done` event.
- A second live Runs prompt exposed reasoning-snapshot duplication with whitespace-prefixed answers; the parser now suppresses normalized snapshot repeats and returned one confirmation answer.
- Closed a multimodal safety gap in opt-in Runs mode: attachment-bearing turns now stay on the legacy chat-completions transport until a verified upstream Runs multipart/input contract exists. Added a regression acceptance test and recorded the decision in ADR-011.
- Reconciled `plan.md` with the current M1 evidence: removed stale claims that Runs approval and session persistence were unimplemented, while retaining explicit `[~]` live/manual parity gates and the M5 frontend-embedding boundary.
- Fixed the duplicate assistant rendering path by moving terminal event reduction outside the React state updater; browser acceptance now verifies one final answer across reconnect. Queued turns now retain selected attachments and are covered by the same browser check.
- Removed the edit/retry truncate race: the frontend now waits for the server-side transcript prefix operation to succeed before mutating the visible branch or starting the replacement turn, and surfaces failures instead of silently continuing.
- Expanded the Gateway normalizer against the current Hermes Runs event vocabulary (`tool.start/progress/complete`, subagent lifecycle events, and reasoning deltas), with upsert behavior preventing repeated progress frames from creating duplicate cards.
- Added `scripts/m1-live-parity.sh`, an isolated operator probe for session CRUD metadata, real no-tools chat, history persistence, and truncate semantics; it cleans up its temporary session and explicitly leaves approval/tool execution as separate live checks.
- Added frontend reducer coverage for canonical tool progress upserts and stable subagent completion, bringing the contract suite to seven passing tests.
- Live M1 probes passed for isolated session CRUD metadata, real chat/history, and transcript truncation; Runs mode also emitted normalized tool start/complete activity. Safe subagent and approval prompts produced no structured events, so those capability limits remain explicitly unclaimed.
- Live branch probe passed for edit/retry persistence: truncation removed the old assistant branch before the replacement response was appended; the remaining parity gap is original-UI side-by-side interaction proof, not server transcript semantics.
- Gateway capability discovery advertised approval and tool-progress support, but explicit harmless terminal/approval/delegation probes still produced no structured approval or subagent event; those results are recorded as observed runtime behavior rather than implementation claims.

## 2026-08-30 - M2 workspace composition

- Added a server-owned workspace root with path containment after symlink resolution, restrictive file permissions, bounded previews/uploads, and traversal/symlink tests.
- Added normalized workspace tree, preview, download, mutation, upload, and Git status API routes. Errors use safe codes without exposing filesystem contents.
- Added a resizable desktop workspace panel and mobile sheet with breadcrumbs, text/code/Markdown/image/binary preview states, editing, file operations, upload, download, and Git branch/status context.
- Workspace state is kept outside chat state so the selected preview remains visible while a response streams. Build and acceptance verification are tracked with the M2 checklist.
- Added server-owned local password setup, iterative hashing, signed HttpOnly/SameSite cookies, login rate limiting, same-origin protection, remote-host setup gating, logout, and safe identity status.
- Added profile discovery from server-only JSON, model/provider picker data, active profile switching, onboarding diagnostics, trusted-header mode, and explicit OIDC/WebAuthn capability reporting.
- Added backend auth/profile tests and documented the credential boundary. OIDC authorization and WebAuthn ceremonies remain `[~]` until deployment-specific issuer/RP configuration is available; no unavailable provider is presented as working.

## 2026-08-30 - M4 control center foundation

- Added server-owned persistent collections for tasks, todos, goals, spaces, and filtered preferences with CRUD API routes and tests.
- Added safe skills and memory discovery endpoints plus explicit capability reporting for runtime-dependent voice, background, extension, and terminal surfaces.
- Replaced dead Soon navigation buttons with accessible control-center views for tasks, todos, goals, spaces, skills, and memory, including loading, empty, error, create, complete, and delete states.
- M4 remains partial by design: cron execution/history, external memory adapters, voice, scheduler wakeups, plugin loading, and terminal execution are not represented as working until their runtime and sandbox contracts are implemented.

## 2026-08-30 - M5 distribution baseline

- Added embedded frontend serving for the Go binary, reproducible `make build` artifact injection, standalone backup/restore migrator, install script, release CI matrix, Nix package expression, and embedded-artifact secret scan.
- M5 remains partial for hosted visual regression, external security review, parallel beta, and final cutover sign-off. Docker and Nix cannot be executed locally because those CLIs are unavailable.
- Added a performance budget gate, M5 Playwright control-center acceptance script, formal threat model, beta runbook, cutover checklist, and CI binary build coverage.
- Added local `/help` and `/clear` command affordances, browser Speech Recognition input with unsupported-browser fallback, and server-persisted theme/locale preference controls.
- Audited `hermes-webui-personal` and aligned the M4 route shape with its cron, settings, and plugin contracts. Added bounded cron run history and pause/resume lifecycle routes while keeping wall-clock execution and plugin loading explicitly gated.

## 2026-08-30 - M6 release hardening and operations

- Added a local-only `/ready` probe that reports BFF initialization separately
  from Hermes Gateway reachability, and routed it through the production Nginx
  baseline.
- Added `make release-gate` with backend tests/vet/build, embedded artifact
  scanning, performance budgets, secret-boundary scanning, and frontend tests/build.
- CI now runs the secret/artifact gates and a Chromium control-center acceptance
  job. Hosted execution remains the source of CI evidence.
- M6 implementation is complete locally; live Hermes deployment, hosted CI, and
  external beta/security sign-off remain explicitly outside automated claims.

## 2026-08-30 - UI primitive and Chat navigation cleanup

- Replaced visible native form controls and action buttons across chat,
  workspace, auth, control center, discovery, messages, and sidebar with the
  copy-owned shadcn-style primitives.
- Kept hidden native file inputs only where the browser file picker is the
  actual transport mechanism.
- Recent sessions now render inside the Chat navigation context, preserving
  search and session actions while keeping other menus focused.
- Frontend tests and production build pass; browser acceptance remains green.

The layout was clarified after review: Recent sessions is now a second sidebar
inside the Chat content area, not an item list in the primary navigation rail.
The acceptance script verifies it is visible only on Chat.

## 2026-08-30 - Chat rail layout and dialog cleanup

- Moved Recent sessions into a dedicated in-content Chat rail with a wider
  title column and right-aligned compact actions, preventing session names from
  being truncated by the action toolbar.
- Added a copy-owned Dialog primitive and replaced native prompt/confirm flows
  for session and workspace actions. Native file inputs remain hidden upload
  transports only.
- Frontend tests, production build, and browser acceptance pass.

## 2026-08-30 - Session toolbar refinement

- Increased primary rail icon/button targets for clearer desktop navigation
  while removing the non-navigation Gateway-first status card.
- Reduced the Chat session rail width to give the conversation canvas more
  room.
- Replaced the hover action toolbar with a per-session three-dot overflow menu;
  rename, pin, archive, and delete behavior remain unchanged and delete still
  opens the shared Dialog.
- Added Escape and outside-click dismissal to the shared dropdown menu.

## 2026-08-30 - M5 distribution checklist closure

- Revalidated the complete local `make release-gate`: backend tests/vet/build,
  frontend tests/build, embedded artifact scan, secret boundary, and
  performance budget all pass.
- Added a multi-stage Docker Compose baseline with an explicit required Gateway
  key, host Gateway mapping, restart policy, and `/ready` healthcheck.
- Synced M5 task statuses with evidence. Local distribution surfaces are
  `[x]`; hosted release matrix, Nix runtime, visual comparison, independent
  security review, beta, and cutover sign-off remain `[~]`.

## 2026-08-30 - Control-center menu audit and Hermes discovery fix

- Audited every sidebar menu and found empty control collections serialized as
  `null`, which crashed the React control view when it called `.length`.
- Fixed empty collection responses to use arrays, added the missing Goals menu,
  and gave Profiles its own data view instead of falling back to Tasks.
- Tasks now consumes the compatibility cron route and exposes run/pause/resume
  actions. Skills and Memory now read Hermes runtime state, with safe bounded
  previews for `SKILL.md` and memory files.
- Runtime verification against the local Hermes home found the installed
  Skills and memory files. Backend tests and frontend tests/build passed.

## 2026-08-30 - Icon navigation rail and collapsible Chat sessions

- Rebuilt the desktop primary navigation as a 72px icon-only rail with
  accessible labels and focus/hover tooltips; mobile navigation retains labels
  inside its drawer.
- Chat now toggles the in-content Recent sessions panel when the active Chat
  icon is clicked again. The session panel can also collapse from its header.
- Widened the session panel and reserved an action-toolbar column so titles and
  metadata remain readable while rename, pin, archive, and delete actions stay
  available.
- Extended the browser acceptance check for tooltip labels and Chat
  expand/collapse behavior. Production build passed; live Hermes behavior was
  not changed by this UI-only slice.

## 2026-08-30 - Personal WebUI parity convergence planning

- Audited the read-only personal source at commit
  `3caeca14064cec36c9c7b4f83ffade9a92cf2aee` and the authenticated production
  shell without storing credentials or private runtime data in the repository.
- Confirmed the production composition includes the compact rail, WebUI/CLI
  session sources, project filters, conversation canvas, composer-owned
  profile/workspace/model controls, and demand-driven Files/Artifacts panel.
- Reclassified existing Studio control surfaces as partial where they do not yet
  implement the personal route/state workflow. Added `MVP.md` and M7-M11 tasks
  for shell, conversation runtime, operator workbench, identity/extensibility,
  and reversible certification.
- No frontend or backend implementation changed in this planning slice.
## 2026-08-30 — P004 personal WebUI parity manifest

- Added `docs/personal-webui-feature-manifest.md` from the frozen personal
  source contracts and sanitized production behavior audit.
- Mapped shell, chat/session, workbench/operator, identity/settings, extension,
  and distribution capabilities to Studio owners, tasks, dispositions, and
  evidence requirements.
- Marked P004 complete. P003 remains partial because non-desktop and dynamic
  panel baselines still need capture.
- No production credentials or private runtime state entered the repository.

## 2026-08-30 — P005 shell semantics contract

- Added `docs/shell-contract.md` for stable rail/titlebar DOM ownership,
  keyboard/focus semantics, menu/dialog behavior, and responsive boundaries.
- Added `data-testid` ownership markers, programmatic tooltip descriptions,
  dialog Escape/initial-focus behavior, and keyboard navigation for overflow
  menus.
- Added `scripts/check-m7-shell-contract.sh`; computed-style and screenshot
  comparison remain separate P012 evidence.
- Added `frontend/e2e/m7-shell-check.mjs` for rail toggle, tooltip semantics,
  overflow-menu keyboard navigation, Dialog Escape behavior, and mobile drawer.
- Marked P005 complete after source contract verification and frontend checks.

## 2026-08-30 — P006 session panel parity pass

- Added WebUI/CLI source filters, project chips, channel badges, content-aware
  search labeling, date grouping, batch selection, batch archive/delete, and
  preserved pin/archive/rename/delete overflow actions.
- Automated evidence: frontend tests and production build pass.
- Added bounded server-owned transcript search at `GET /api/sessions?q=...`;
  results remain summary-only and have reader/API regression coverage.
- Added duplicate-session and Markdown export routes with server-side tests;
  browser overflow actions now invoke real state/download behavior.
- P006 remains `[~]`: duplicate/import/share/export actions still need their
  import/share contracts and browser acceptance evidence.

## 2026-08-30 — P007 titlebar and composer runtime controls

- Added server-backed profile and model selectors to the composer, forwarding
  the selected model/provider into the chat start request.
- Added a workspace control that opens the existing workspace panel, a usage
  indicator gated on normalized Gateway totals/limits, and preserved Send,
  Queue, Stop, attachment, voice, and slash-command behavior.
- Automated evidence: frontend tests and production build pass.
- P007 remains `[~]` for titlebar visual measurements and grouped live model
  discovery, which are tracked by P012 and P023.

## 2026-08-30 — P008 workspace panel composition

- Added Files/Artifacts/Todos tabs to the demand-driven workspace panel.
- Files retains tree, breadcrumb, preview, edit, upload, download, rename,
  delete, and Git flows. Todos reads the server-owned control route; Artifacts
  remains an explicit empty state until its backend contract exists.
- Persisted panel open state and width as layout preferences while keeping
  workspace contents server-owned and stream-independent.
- Frontend production build passes. P008 remains `[~]` pending artifact API
  parity and complete browser resize/mobile evidence.
