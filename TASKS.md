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

- [~] Persist sessions using existing `~/.hermes/webui` data where possible (JSON CRUD and chat transcript append implemented; legacy CLI/SQLite bridge deferred)
- [x] Add read-only sessions API contract and integration tests
- [~] Session CRUD, search, date grouping, rename, pin, archive, tags, projects (CRUD/action routes, search/date grouping, and session actions implemented; tags/projects remain metadata-compatible without dedicated UI)
- [x] Load and resume existing session history
- [~] Edit/regenerate/retry and queue while processing (server-side edit truncation, queue, and retry implemented; live parity proof remains)
- [x] Markdown, code highlight/copy, Mermaid, safe links
- [~] Tool cards, subagent cards, reasoning blocks (normalized rendering and redaction implemented; live upstream variants remain)
- [~] Runs API approval request/response parity (BFF forwarding and UI decision path implemented; live Runs API proof remains)
- [~] Attachments and multimodal messages (validated upload/download and Gateway multimodal transport implemented; live model acceptance remains)
- [~] Reconnect/replay cursor and duplicate suppression (SSE IDs, `Last-Event-ID`/`after`, bounded replay, and terminal dedup implemented; browser interruption proof remains)
- [x] Context/token usage indicator
- [~] Chat keyboard, focus, screen-reader, and mobile acceptance tests (keyboard/focus states and functional mobile navigation implemented; browser matrix remains)

## M2 — Workspace composition

- [ ] Secure workspace root discovery and containment
- [ ] Tree, breadcrumb, text/code/Markdown/image preview
- [ ] Create, edit, rename, delete, upload, download
- [ ] Git status badges
- [ ] Resizable desktop panel and mobile sheet
- [ ] Preserve preview during active chat streams

## M3 — Profiles, auth, and onboarding

- [ ] Profile discovery, model/provider picker, profile switch
- [ ] Password auth and secure signed cookies
- [ ] Rate limits, CSRF/origin rules, security headers
- [ ] First-run Gateway configuration and diagnostics
- [ ] Passkeys/WebAuthn
- [ ] OIDC and trusted-header SSO

## M4 — Control center feature parity

- [ ] Tasks/cron and history
- [ ] Skills
- [ ] Memory and external notes
- [ ] Todos and goals
- [ ] Spaces/projects
- [ ] Slash commands and voice input
- [ ] Preferences, skins, locale, update status
- [ ] Background tasks and wakeups
- [ ] Extensions/plugins and terminal surfaces

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
