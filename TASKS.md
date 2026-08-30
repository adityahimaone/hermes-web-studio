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
- [x] Verify production frontend build (114.22 KiB gzip JS at M0)
- [ ] Run live Hermes smoke test using operator credentials
- [ ] Record live proof: date, Hermes version, model/provider, first-token latency, final response

**Gate:** do not begin broad feature migration until a real Hermes response is visible in the new UI.

## M1 — Chat parity

- [ ] Persist sessions using existing `~/.hermes/webui` data where possible
- [ ] Session CRUD, search, date grouping, rename, pin, archive, tags, projects
- [ ] Load and resume existing session history
- [ ] Edit/regenerate/retry and queue while processing
- [ ] Markdown, code highlight/copy, Mermaid, safe links
- [ ] Tool cards, subagent cards, reasoning blocks
- [ ] Runs API approval request/response parity
- [ ] Attachments and multimodal messages
- [ ] Reconnect/replay cursor and duplicate suppression
- [ ] Context/token usage indicator
- [ ] Chat keyboard, focus, screen-reader, and mobile acceptance tests

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
