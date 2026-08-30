# Execution backlog

Status legend: `[x]` implemented and verified with automated tests, `[~]` implemented but needs live/manual proof, `[ ]` not started.

## M0 — Foundation and chat gate

- [x] Create Vite + React + TypeScript + Tailwind v4 project
- [x] Add copy-owned shadcn-style primitives and responsive dashboard shell
- [x] Add typed chat event contract
- [x] Add Go configuration with server-only Gateway credentials
- [x] Add `/health` and `/api/health/hermes`
- [x] Add `/api/chat/start`, `/api/chat/stream`, `/api/chat/cancel`
- [x] Translate OpenAI and Hermes Gateway SSE variants
- [x] Add mock Gateway integration tests for streaming, errors, headers, and cancel
- [x] Add frontend parser/reducer tests
- [x] Verify production frontend build (114.27 KiB gzip JS at M0)
- [x] Suppress duplicate completion output when Gateway sends deltas and `run.completed.output`
- [x] Add graceful shutdown, server timeouts, and a production container/reverse proxy baseline
- [x] Run live Hermes smoke test using operator credentials
- [x] Record live proof: date, Hermes version, model/provider, first-token latency, final response

Production baseline is available with `docker build -t hermes-web-studio:local .`. Configure the Gateway URL and key as container environment variables. The image serves the built frontend through Nginx and keeps the Go BFF and Gateway credentials server-side.

**Gate:** do not begin broad feature migration until a real Hermes response is visible in the new UI.

## M1 — Chat parity

- [x] Persist sessions using existing `~/.hermes/webui` data where possible (JSON CRUD, metadata-preserving edits, attachment sidecars, and chat transcript append tested; no chat SQLite/CLI source is present in the resolved state directory)
- [x] Add read-only sessions API contract and integration tests
- [~] Session CRUD, search, date grouping, rename, pin, archive, tags, projects (CRUD/action routes, search/date grouping, session actions, inline tag/project metadata, and isolated live parity runner verified; original-WebUI side-by-side proof remains)
- [x] Load and resume existing session history
- [~] Edit/regenerate/retry and queue while processing (server-side edit truncation awaited before edit/retry, queued text/attachment delivery and live persisted branch semantics verified; original-UI side-by-side click proof remains)
- [x] Markdown, code highlight/copy, Mermaid, safe links
- [~] Tool cards, subagent cards, reasoning blocks (canonical Runs lifecycle aliases, upserted progress cards, redaction, and live tool events verified; current Gateway emitted no structured subagent event)
- [~] Runs API approval request/response parity (BFF forwarding and UI decision path implemented; live Runs API proof remains)
- [x] Attachments and multimodal messages (validated upload/download, MIME-specific Gateway transport, live model completion verified, and attachment turns safely stay on chat-completions when Runs mode is enabled)
- [x] Reconnect/replay cursor and duplicate suppression (SSE IDs, `Last-Event-ID`/`after`, bounded replay, live replay proof, and Playwright interruption/reconnect coverage)
- [x] Context/token usage indicator
- [x] Chat keyboard, focus, screen-reader, and mobile acceptance tests (Playwright coverage passes for mobile navigation, focus/labels, desktop rail, Shift+Enter, and Enter-to-send)

## M2 — Workspace composition

- [x] Secure workspace root discovery and containment
- [x] Tree, breadcrumb, text/code/Markdown/image preview
- [x] Create, edit, rename, delete, upload, download
- [x] Git status badges
- [x] Resizable desktop panel and mobile sheet
- [x] Preserve preview during active chat streams

## M3 — Profiles, auth, and onboarding

- [x] Profile discovery, model/provider picker, profile switch
- [x] Password auth and secure signed cookies
- [x] Rate limits, CSRF/origin rules, security headers
- [x] First-run Gateway configuration and diagnostics
- [~] Passkeys/WebAuthn (capability is reported as unavailable until WebAuthn RP configuration is supplied)
- [~] OIDC and trusted-header SSO (trusted-header mode is implemented; OIDC remains configuration/discovery-only until issuer client registration is supplied)

## M4 — Control center feature parity

- [~] Tasks/cron and history (upstream-compatible create/list/run/pause/resume/history routes are implemented; wall-clock scheduler, delivery, and alerts remain)
- [x] Skills (safe local discovery surface)
- [~] Memory and external notes (safe local discovery surface; external source adapters remain)
- [x] Todos and goals
- [x] Spaces/projects
- [~] Slash commands and voice input (local /help and /clear commands plus browser speech input are implemented; provider transcription remains)
- [~] Preferences, skins, locale, update status (theme/locale persistence and `/api/settings` compatibility aliases are implemented; skins/update registry remains)
- [~] Background tasks and wakeups (capability reported unavailable until a durable scheduler contract exists)
- [~] Extensions/plugins and terminal surfaces (capability reported unavailable until sandboxed runtime contracts exist)

## M5 — Distribution and cutover

- [ ] Embed `frontend/dist` in the Go binary
- [ ] Legacy state migration and rollback tool
- [ ] One-command developer install
- [ ] Linux/macOS/Windows release matrix
- [ ] Multi-stage Docker image and compose compatibility
- [ ] Nix package
- [ ] Visual regression matrix and performance budgets
- [ ] Security review and threat model
- [ ] Parallel beta against original WebUI
- [ ] Cutover checklist; archive Python only after parity sign-off

M5 implementation status: embedded frontend, recoverable backup/restore migration, install script, Docker integration, and release artifacts are implemented. Hosted release matrix execution, Nix runtime verification, visual regression, formal security review, beta, and cutover remain pending evidence.

M5 verification follow-up: performance budget script, browser control-center acceptance script, formal threat model, and beta/cutover runbook are now committed. Hosted release execution, Docker/Nix runtime checks, visual capture comparison, beta, and cutover sign-off remain pending.

## M6 — Release hardening and operations

- [x] Separate liveness and local readiness probes, with deployment routing and tests
- [x] Add one-command release gate for formatting, tests, vet, build, artifact, performance, and secret-boundary checks
- [x] Enforce the browser credential boundary with a regression scan
- [~] Run backend artifact and browser control-center acceptance checks in CI (workflow committed; hosted run pending)
- [x] Document M6 compatibility, decision, and operator verification flow

M6 status: implementation and local automated checks are complete. Hosted CI
execution and live Hermes deployment verification remain evidence gates, not
claims inferred from local tests.

Control-center audit follow-up: empty collections, missing Goals navigation,
Profiles fallback, and incorrect Skills/Memory roots were fixed. Hermes Skills
and memory runtime verification passed locally; live scheduler execution and
external provider-backed integrations remain `[~]` under M4.

UI follow-up: visible controls now use shared shadcn-style primitives and
Recent sessions is a secondary sidebar inside Chat content without removing
session actions.

UI polish follow-up: Chat now owns a dedicated Recent sessions rail inside its
content area, and session/workspace modal actions use the shared Dialog
primitive instead of browser-native prompt/confirm UI.

Navigation polish follow-up: the primary rail is icon-only on desktop with
keyboard/focus tooltips. Chat toggles its in-content Recent sessions rail when
activated while already on Chat; the panel can also be collapsed from its
header, and the session title/action layout reserves space for the toolbar.

Navigation refinement follow-up: primary rail targets are now larger, the
Gateway-first status card is removed from the rail, the session panel is
narrower, and per-session actions are grouped under an overflow menu.
