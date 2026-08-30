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

- [x] Embed `frontend/dist` in the Go binary
- [x] Legacy state migration and rollback tool
- [x] One-command developer install
- [~] Linux/macOS/Windows release matrix (workflow committed; hosted matrix execution remains)
- [x] Multi-stage Docker image and compose compatibility baseline
- [~] Nix package (package definition committed; Nix runtime verification remains)
- [~] Visual regression matrix and performance budgets (performance check passes; visual baseline comparison remains)
- [~] Security review and threat model (threat model and automated boundary checks pass; independent review remains)
- [~] Parallel beta against original WebUI (runbook committed; operator comparison remains)
- [~] Cutover checklist; archive Python only after parity sign-off (checklist committed; sign-off remains)

M5 implementation status: embedded frontend, recoverable backup/restore migration,
install script, Docker/Compose integration, and release artifacts are
implemented. Hosted release matrix execution, Nix runtime verification, visual
regression comparison, formal security review, beta, and cutover remain `[~]`
until their external evidence is recorded.

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

## Parity convergence note

M0-M6 record implemented Studio slices. They do not certify replacement parity
with `hermes-webui-personal`. The following tasks use `MVP.md` and the frozen
personal source as the new acceptance baseline.

## M7 — Baseline freeze and shell parity

- [x] P001 Freeze the read-only personal source commit and record the production build marker without committing credentials or private state
- [x] P002 Publish `MVP.md` with visual, interaction, frontend, backend, evidence, and lightweight-runtime definitions
- [~] P003 Capture the production shell at desktop; settings, narrow, tablet, mobile, and all panel states remain to be captured with dynamic-data masks
- [x] P004 Build a route/state/feature manifest from the frozen personal source and map every entry to a Studio owner, test, and disposition
- [x] P005 Freeze DOM, computed-style, keyboard, focus, tooltip, menu, dialog, and responsive contracts for the primary rail and titlebar
- [~] P006 Match the Chat session panel: WebUI/CLI sources, content search, project chips, date groups, archive/pin, channel badges, batch mode, and overflow actions (source/project/channel/batch UI, server-backed transcript search, duplicate, and Markdown export implemented; import/share remain)
- [~] P007 Match the titlebar and composer control ownership, ordering, responsive overflow, model/profile/workspace pickers, context ring, and Stop/Send states (runtime pickers and usage indicator implemented; visual titlebar evidence and grouped live model discovery remain)
- [~] P008 Match the demand-driven workspace panel shell, resize/open persistence, Files/Artifacts/Todos tabs, and mobile slide-over (Files/Todos tabs and persisted open/width implemented; Artifacts backend and full browser resize/mobile evidence remain)
- [~] P009 Rebuild the Control Center section structure and primary-rail order/visibility customization without shipping non-functional controls (rail visibility/order dialog implemented; full reference section registry and broader panel evidence remain)
- [~] P010 Implement System/Dark/Light boot without flash and the complete frozen built-in skin registry through lazy theme assets (pre-React theme boot and 21-name registry implemented; per-skin visual tuning and screenshot matrix remain)
- [~] P011 Match mobile drawer, five-tab bottom navigation, 44px targets, safe-area/keyboard composer behavior, and panel transitions (five-tab navigation, 44px targets, safe-area composer spacing, reduced-motion panel transition, and browser navigation coverage implemented; full device/IME evidence remains)
- [ ] P012 Add Playwright screenshot, DOM, computed-style, keyboard, and responsive shell matrices at the `MVP.md` viewports

## M8 — Conversation runtime and session parity

- [~] P013 Freeze personal chat/SSE/turn-journal/session contracts and deterministic fixtures before changing the Gateway adapter (validated inflight journal parser and deterministic fixtures; full personal-source fixture freeze remains)
- [~] P014 Make Runs-first normalization cover message, reasoning, tool, progress, subagent, approval, clarification, usage, compression, recovery, and terminal events (existing normalized event path retained; clarification/compression/recovery coverage remains)
- [~] P015 Preserve inflight turns across session switches, hard reload, reconnect backoff, cursor replay, and full-session poll fallback (journal restore, transcript restore, cursor reconnect, and stale-stream protection implemented; expired-stream polling fallback and live proof remain)
- [ ] P016 Implement Queue/Interrupt/Steer modes, including attachments, visible pending intent, cancellation, replacement, and cleanup on every exit
- [ ] P017 Implement compact worklog, transparent stream, and final-answer-only modes with per-turn disclosure persistence and settled-history behavior
- [ ] P018 Add compaction barriers, `/compact`, exhausted-context recovery, partial-message crash recovery, retry, and safe transcript repair/audit
- [ ] P019 Complete message copy, edit/regenerate, clear watermark, branch/fork from any valid turn, duplicate, undo, and lineage reporting
- [ ] P020 Complete WebUI/CLI/cron/webhook/gateway session projections, canonical deduplication, active recency, clusters, external read-only rules, and import
- [ ] P021 Complete title/content search, projects, tags, archived visibility, pinning, batch delete/move/archive, and adaptive/LLM title regeneration
- [ ] P022 Complete Markdown, syntax, JSON/YAML tree, Mermaid toolbar, KaTeX, tables, safe links, media tokens, timestamps, and routing metadata
- [ ] P023 Complete searchable provider-grouped model selection, live custom endpoint discovery, auxiliary models, aliases, quota/cost/TPS, reasoning, and toolsets
- [ ] P024 Complete slash-command registry/autocomplete and reference built-ins with transparent pass-through for unknown commands
- [ ] P025 Complete transcript/session Markdown, JSON, and self-contained HTML export; import; public share creation/revocation; and safe download behavior
- [ ] P026 Add deterministic normal/error/cancel/switch/reload/reconnect/compression/recovery lifecycle rows plus live Hermes side-by-side acceptance

## M9 — Workbench and operator parity

- [ ] P027 Replace the single-root workspace abstraction with reference-compatible registered Spaces, active inheritance, ordering, health, suggestions, and profile-local state
- [ ] P028 Complete lazy tree/breadcrumb/filemap, hidden-file rules, broad preview matrix, copy paths, open/reveal, create/edit/move/rename/delete, upload/paste/extract, and size limits
- [ ] P029 Complete workspace Artifacts and optional Todos projections with independent panel state and stream-safe persistence
- [ ] P030 Add contained terminal start/input/output/resize/close and prove process ownership/cleanup on success, error, cancel, replacement, and disconnect
- [ ] P031 Add git branch/status/diff/stage/unstage/discard/commit/fetch/pull/push/stash-checkout flows with confirmation and hostile-path tests
- [ ] P032 Replace task placeholders with cron create/edit/delete, schedule builder, delivery options, skill picker, run/pause/resume, output/history, alerts, and watch status
- [ ] P033 Implement Kanban boards/tasks/assignees/links/bulk actions/dispatch/events/stats and map its persisted state without local fallback
- [~] P034 Complete Skills grouping/search/content/linked files/create/edit/delete/toggle/usage and profile-aware refresh (preview loading/empty/error state separation implemented; CRUD, linked files, usage, and profile refresh remain)
- [ ] P035 Complete Memory/USER/external-note sources, search, timestamps, inline create/edit, and path/source safety
- [ ] P036 Complete Todos and Goals lifecycle plus current-goal/workspace projections using Hermes-owned state
- [ ] P037 Implement Insights for usage, cost/provider history, state synchronization, and operational summaries
- [ ] P038 Implement Logs, crash visibility, background status, agent/gateway/system health, safe restart, and rollback diagnostics
- [ ] P039 Implement external-channel sessions, handoff dock, round summaries, per-conversation identity, routing metadata, and model-switch warnings
- [ ] P040 Add browser and backend matrices for every rail panel covering loading, empty, CRUD, watch/stream, permission failure, cancellation, and recovery

## M10 — Identity, settings, extensibility, and platform parity

- [ ] P041 Complete profile create/update/switch/delete, concurrent isolation, profile-local workspace/state, runtime refresh, and OAuth linking
- [ ] P042 Complete provider CRUD, self-hosted provider setup, live models, custom OpenAI-compatible endpoints, personalities, and credential redaction
- [ ] P043 Complete password lifecycle, passkey registration/login/delete, OIDC start/callback, trusted headers, cookie rotation, CSRF/CORS, and reverse-proxy threat cases
- [ ] P044 Match branded onboarding for existing config, provider setup, OAuth polling/cancel, probes, completion, skip rules, and failure recovery
- [ ] P045 Complete Control Center Conversation, Appearance, Preferences, Providers, Plugins, Extensions, System, and Help sections with searchable settings
- [ ] P046 Implement all reference preference behaviors, including send key, activity modes, scroll, rendering, outline, notifications, token controls, voice/TTS, and tab/composer customization
- [ ] P047 Add 15 locales with key parity, fallback accounting, RTL chat layout, CJK/IME behavior, and localized login/onboarding/error surfaces
- [ ] P048 Add PWA manifest/service worker/offline shell, install flow, icons, update invalidation, bfcache behavior, and subpath support
- [ ] P049 Implement MCP server/tool management and plugin settings/status using server-owned validation and secret filtering
- [ ] P050 Implement extension registry/install/toggle/uninstall/settings, skin/TTS/nav capability registration, sidecar consent/proxy, iframe tab, and trust boundaries
- [ ] P051 Complete push-to-talk, hands-free mode, raw audio/transcription capability, browser/server TTS engines, rate/pitch/voice, and unsupported states
- [ ] P052 Complete update check/apply/lock recovery, shutdown/restart controls, version/health summaries, and release diagnostics
- [ ] P053 Match Docker single/two/three-container topologies, subpath reverse proxy, multi-arch healthchecks, installer behavior, and state migration/rollback
- [ ] P054 Add auth, onboarding, settings, locale, PWA, extension, and distribution acceptance matrices

## M11 — MVP certification and reversible cutover

- [ ] P055 Build sanitized personal-state fixtures plus a copied-state migration rehearsal; never run migration against the production state directory
- [ ] P056 Run full contract comparison for required personal route/state effects and record every intentional alias or deviation
- [ ] P057 Run visual matrices across required viewports, System/Dark/Light, all release skins, long lists/transcripts, active streaming, dialogs, and both side panels
- [ ] P058 Run accessibility checks for keyboard-only flows, focus restoration, screen readers, contrast, reduced motion, touch targets, RTL, and IME
- [ ] P059 Run live Hermes/Gateway side-by-side rows for chat, tools, subagents, approvals, clarification, compression, recovery, external channels, and scheduled work
- [ ] P060 Prove 1,000-session and 5,000-message behavior, JS gzip, idle RSS, startup, first-token overhead, stream rendering, and no per-token full refetch
- [ ] P061 Run independent security review for auth/proxy, SSRF, XSS/Markdown, workspace/terminal/git, uploads/extraction, extensions/sidecars, migration, and secret boundaries
- [ ] P062 Run hosted Linux/macOS/Windows artifact, Docker, Nix, PWA/subpath, and browser matrices from release candidates
- [ ] P063 Complete parallel beta with the personal WebUI available for rollback and resolve every critical/major parity gap
- [ ] P064 Rehearse backup, migration, release validation, rollback, and operator communication using copied state
- [ ] P065 Request explicit user cutover approval; do not archive or modify the personal WebUI before approval
- [ ] P066 Keep rollback artifacts and the personal implementation available until the agreed rollback window closes
