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
- [~] Run live Hermes smoke test using operator credentials (current probe blocked by provider HTTP 402 insufficient credits)
- [~] Record live proof: date, Hermes version, model/provider, first-token latency, final response (current probe blocked by provider HTTP 402 insufficient credits; live chat remains unclaimed)

Production baseline is available with `docker build -t hermes-web-studio:local .`. Configure the Gateway URL and key as container environment variables. The image serves the built frontend through Nginx and keeps the Go BFF and Gateway credentials server-side.

**Gate:** do not begin broad feature migration until a real Hermes response is visible in the new UI.

## M1 — Chat parity

- [x] Persist sessions using existing `~/.hermes/webui` data where possible (JSON CRUD, metadata-preserving edits, attachment sidecars, and chat transcript append tested; no chat SQLite/CLI source is present in the resolved state directory)
- [x] Add read-only sessions API contract and integration tests
- [~] Session CRUD, search, date grouping, rename, pin, archive, tags, projects (CRUD/action routes, search/date grouping, session actions, inline tag/project metadata, overflow actions visible for keyboard and touch, and isolated live parity runner verified; original-WebUI side-by-side proof remains)
- [x] Load and resume existing session history
- [~] Edit/regenerate/retry and queue while processing (server-side edit truncation awaited before edit/retry, queued text/attachment delivery and live persisted branch semantics verified; original-UI side-by-side click proof remains)
- [x] Markdown, code highlight/copy, Mermaid, safe links
- [~] Tool cards, subagent cards, reasoning blocks (canonical Runs lifecycle aliases, bounded redaction for browser-facing previews/summaries and task/name strings, and live tool events verified; current Gateway emitted no structured subagent event)
- [~] Runs API approval request/response parity (BFF forwarding and UI decision path implemented; live Runs API proof remains)
- [~] Attachments and multimodal messages (validated upload/download, MIME-specific Gateway transport, and attachment turns safely stay on chat-completions when Runs mode is enabled; current live model completion proof blocked by provider HTTP 402)
- [~] Reconnect/replay cursor and duplicate suppression (SSE IDs, `Last-Event-ID`/`after`, bounded replay, and Playwright interruption/reconnect coverage verified; current live replay proof blocked by provider HTTP 402)
- [x] Harden chat turn failure boundaries (per-session turn serialization, store-wide same-process session-index write serialization, owned rollback error reporting, nested HTTP-200 error rejection, terminal SSE validation, and assistant persistence failure handling; live provider completion remains blocked by current HTTP 402)
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
- [~] OIDC and trusted-header SSO (trusted-header authentication disabled fail-closed until a server-validated proxy boundary/identity contract exists; OIDC remains configuration/discovery-only until issuer client registration is supplied)

## M4 — Control center feature parity

- [~] Tasks/cron and history (upstream-compatible create/list/run/pause/resume/history routes are implemented; wall-clock scheduler, delivery, and alerts remain)
- [x] Skills (safe local discovery surface)
- [~] Memory and external notes (safe local discovery surface; external source adapters remain)
- [x] Todos and goals
- [x] Spaces/projects (profile-scoped active/list/use state; remote spaces remain unavailable to local Kanban execution)
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
- [~] P006 Match the Chat session panel: WebUI/CLI sources, content search, project chips, date groups, archive/pin, channel badges, batch mode, and overflow actions (source/project/channel/batch UI, server-backed transcript search, duplicate, Markdown export, and ArrowUp/ArrowDown session selection implemented; selection focus waits for React rerender with bounded retry; import/share remain)
- [~] P007 Match the titlebar and composer control ownership, ordering, responsive overflow, model/profile/workspace pickers, context ring, and Stop/Send states (runtime pickers and usage indicator implemented; visual titlebar evidence and grouped live model discovery remain)
- [~] P008 Match the demand-driven workspace panel shell, resize/open persistence, Files/Artifacts/Todos tabs, and mobile slide-over (Files/Todos tabs and persisted open/width implemented; Artifacts backend and full browser resize/mobile evidence remain)
- [~] P009 Rebuild the Control Center section structure and primary-rail order/visibility customization without shipping non-functional controls (rail visibility/order dialog implemented; full reference section registry and broader panel evidence remain)
- [~] P010 Implement System/Dark/Light boot without flash and the complete frozen built-in skin registry through lazy theme assets (pre-React theme boot and 21-name registry implemented; per-skin visual tuning and screenshot matrix remain)
- [~] P011 Match mobile drawer, five-tab bottom navigation, 44px targets, safe-area/keyboard composer behavior, and panel transitions (five-tab navigation, 44px targets, safe-area composer spacing, reduced-motion panel transition, and browser navigation coverage implemented; full device/IME evidence remains)
- [~] P012 Add Playwright screenshot, DOM, computed-style, keyboard, and responsive shell matrices at the `MVP.md` viewports (five-viewport screenshot/geometry matrix passes; reference pixel-diff and full keyboard matrix remain)

## M8 — Conversation runtime and session parity

- [~] P013 Freeze personal chat/SSE/turn-journal/session contracts and deterministic fixtures before changing the Gateway adapter (validated inflight journal parser and deterministic fixtures; full personal-source fixture freeze remains)
- [~] P014 Make Runs-first normalization cover message, reasoning, tool, progress, subagent, approval, clarification, usage, compression, recovery, and terminal events (existing normalized event path retained; clarification/compression/recovery coverage remains)
- [~] P015 Preserve inflight turns across session switches, hard reload, reconnect backoff, cursor replay, and full-session poll fallback (epoch/session/stream/pump ownership guards, hard-reload restore claim before stream assignment, generic external-error normalization, abortable pending pump replacement/cancel, restore controller cleanup, queued baseline handoff, timeout without invented messages, and identity-preserving restored replies implemented; live proof remains)
- [~] P016 Implement Queue/Interrupt/Steer modes, including attachments, visible pending intent, cancellation, replacement, and cleanup on every exit (modes, pending intent, attachment names, removal, replacement, and cleanup implemented; live/browser proof remains)
- [~] P017 Implement compact worklog, transparent stream, and final-answer-only modes with per-turn disclosure persistence and settled-history behavior (disclosure helpers and render predicates added; UI persistence and live stream modes remain)
- [~] P018 Add compaction barriers, `/compact`, exhausted-context recovery, partial-message crash recovery, retry, and safe transcript repair/audit (barrier and repair helpers added; command integration and crash/live recovery remain)
- [~] P019 Complete message copy, edit/regenerate, clear watermark, branch/fork from any valid turn, duplicate, undo, and lineage reporting (branch/fork/undo lineage helpers added; full message UI and server effects remain)
- [~] P020 Complete WebUI/CLI/cron/webhook/gateway session projections, canonical deduplication, active recency, clusters, external read-only rules, and import (projection/source/dedup helpers added; backend projections and import remain)
- [~] P021 Complete title/content search, projects, tags, archived visibility, pinning, batch delete/move/archive, and adaptive/LLM title regeneration (batch archive/delete preserve successful actions and report failed session IDs in an accessible result; move/import/title regeneration remain)
- [~] P022 Complete Markdown, syntax, JSON/YAML tree, Mermaid toolbar, KaTeX, tables, safe links, media tokens, timestamps, and routing metadata (safe export/runtime helpers added; renderer and routing parity remain)
- [~] P023 Complete searchable provider-grouped model selection, live custom endpoint discovery, auxiliary models, aliases, quota/cost/TPS, reasoning, and toolsets (server-owned `/api/models/catalog` validates local Gateway URLs, disables redirects, bounds/normalizes/deduplicates catalog values by provider+ID, returns explicit unavailable state without fallback models, and composer validates catalog response shape/status while requiring available selections; provider-aware collision-safe Select identities and colon-safe backend tuple dedupe are covered; auxiliary/runtime metadata, live custom endpoints, and browser/live proof remain)
- [~] P024 Complete slash-command registry/autocomplete and reference built-ins with transparent pass-through for unknown commands (typed `/help` and `/clear` registry, keyboard suggestions, and unknown-command pass-through added; full reference command catalog and live command effects remain)
- [~] P025 Complete transcript/session Markdown, JSON, and self-contained HTML export; import; public share creation/revocation; and safe download behavior (safe Markdown/JSON export endpoints and HTML/runtime helpers added; frontend exposes Markdown only; import/share endpoints remain)
- [~] P026 Add deterministic normal/error/cancel/switch/reload/reconnect/compression/recovery lifecycle rows plus live Hermes side-by-side acceptance (lifecycle rows helper added; live side-by-side evidence remains)

## M9 — Workbench and operator parity

- [~] P027 Replace the single-root workspace abstraction with reference-compatible registered Spaces, active inheritance, ordering, health, suggestions, and profile-local state (server-owned registry validates configured profile ownership, rejects URI/SSH/private refs, resolves symlinks before local containment, redacts remote/legacy refs, uses numeric ordering, and atomically persists create/activate/delete; UI validates primitive normalized entries with opaque workspace_ref Kanban mapping, supports keyboard lane transitions, visible Space mutation network errors, task-detail action errors, delete/error states, and touch-safe controls; profile CRUD persistence/local inheritance, suggestions/update UI, health/canary/live/browser evidence, and rendered Kanban tests remain open; remote Space references remain opaque and unavailable for local Kanban execution, with raw remote refs never returned to browser; profile-local isolation is partial because legacy `profile_id == ""` Spaces remain visible to every profile for compatibility; full isolation remains open; profile CRUD persistence remains outside this slice and in-memory; containment checks reduce symlink risk but residual replacement TOCTOU remains possible and is not absolutely eliminated)
- [~] P028 Complete lazy tree/breadcrumb/filemap, hidden-file rules, broad preview matrix, copy paths, open/reveal, create/edit/move/rename/delete, upload/paste/extract, and size limits (tree/preview/edit/basic CRUD/upload/download/copy/open and safe containment added; paste/extract/lazy filemap/reveal and broader preview matrix remain)
- [~] P029 Complete workspace Artifacts and optional Todos projections with independent panel state and stream-safe persistence (server-owned Todos projection and explicit Artifacts empty state retained; Artifacts contract remains)
- [~] P030 Add contained terminal start/input/output/resize/close and prove process ownership/cleanup on success, error, cancel, replacement, and disconnect (explicit unavailable capability and safe UI state added; contained process lifecycle remains)
- [~] P031 Add git branch/status/diff/stage/unstage/discard/commit/fetch/pull/push/stash-checkout flows with confirmation and hostile-path tests (read-only branch/status/diff and hostile-path containment added; mutating Git flows remain)
- [~] P032 Replace task placeholders with cron create/edit/delete, schedule builder, delivery options, skill picker, run/pause/resume, output/history, alerts, and watch status (run/pause/resume/delete/history added; builder, delivery, alerts, and watch remain)
- [~] P033 Implement Kanban boards/tasks/assignees/links/bulk actions/dispatch/events/stats and map its persisted state without local fallback (native CLI-backed board/read/create/dispatch/action slice, Chat-like Kanban rail, scroll-safe reference task modal, Spaces-tab registry binding, parent/runtime/skills fields, task detail drawer, preview, desktop drag targets, and rendered browser acceptance implemented; Dashboard upgrade, full detail editing, links, bulk, events, stats parity, migration, remote execution proof, and authenticated Dashboard proof remain)
- [~] P034 Complete Skills grouping/search/content/linked files/create/edit/delete/toggle/usage and profile-aware refresh (filesystem-category skill groups with working collapse state, compact category headers, session-density two-line rows, bounded sidebar list scrolling with persistent filters, icon-only sidebar add action, full-width content, and wrapped bounded SKILL.md preview added; CRUD, linked files, usage, and profile refresh remain)
- [~] P035 Complete Memory/USER/external-note sources, search, timestamps, inline create/edit, and path/source safety (server-owned flat note navigation, search/create/edit/delete, and protected SOUL.md added; external sources and timestamps remain)
- [~] P036 Complete Todos and Goals lifecycle plus current-goal/workspace projections using Hermes-owned state (server-owned lifecycle retained and workspace Todos projection added; current-goal integration remains)
- [~] P037 Implement Insights for usage, cost/provider history, state synchronization, and operational summaries (server-owned usage/provider/synchronization view added with explicit cost-unavailable state; live cost source remains)
- [~] P038 Implement Logs, crash visibility, background status, agent/gateway/system health, safe restart, and rollback diagnostics (operator health/log/diagnostics routes now expose sanitized read-only capability and count snapshots; crash/background/restart/rollback flows remain)
- [~] P039 Implement external-channel sessions, handoff dock, round summaries, per-conversation identity, routing metadata, and model-switch warnings (external projection, channel/identity/routing metadata, and explicit unavailable handoff dock added; live channel rows, summaries, and model-switch semantics remain)
- [~] P040 Add browser and backend matrices for every rail panel covering loading, empty, CRUD, watch/stream, permission failure, cancellation, and recovery (backend operator/control coverage and shell matrix added; complete rail matrix remains)

## M10 — Identity, settings, extensibility, and platform parity

- [~] P041 Complete profile create/update/switch/delete, concurrent isolation, profile-local workspace/state, runtime refresh, and OAuth linking (server-owned profile CRUD/switch, active-profile protection, and unavailable-provider activation rejection added; isolation, workspace state, refresh, OAuth remain)
- [~] P042 Complete provider CRUD, self-hosted provider setup, live models, custom OpenAI-compatible endpoints, personalities, and credential redaction (provider create/list/delete, `has_key` redaction, gateway-provider protection, and active-profile dependency protection added; live endpoints/models and personalities remain)
- [~] P043 Complete password lifecycle, passkey registration/login/delete, OIDC start/callback, trusted headers, cookie rotation, CSRF/CORS, and reverse-proxy threat cases (nonce cookie rotation, HTTPS proxy flags, logout origin protection, and auth tests added; passkeys/OIDC/full threat matrix remain)
- [~] P044 Match branded onboarding for existing config, provider setup, OAuth polling/cancel, probes, completion, skip rules, and failure recovery (validation recovery tests and secure onboarding behavior added; branded/provider/OAuth flows remain)
- [~] P045 Complete Control Center Conversation, Appearance, Preferences, Providers, Plugins, Extensions, System, and Help sections with searchable settings (capability section registry, shared collapsible ContextRail, non-JSON-safe loading, scoped capability visibility, and bottom-anchored Settings navigation added; complete section behavior/search remains)
- [~] P046 Implement all reference preference behaviors, including send key, activity modes, scroll, rendering, outline, notifications, token controls, voice/TTS, and tab/composer customization (searchable server-persisted preference surface, immediate theme/skin preview, and explicit media gaps added; full behavior wiring remains)
- [~] P047 Add 15 locales with key parity, fallback accounting, RTL chat layout, CJK/IME behavior, and localized login/onboarding/error surfaces (15-locale resolver/fallback contract added; translated keys, RTL, CJK/IME, and localized surfaces remain)
- [~] P048 Add PWA manifest/service worker/offline shell, install flow, icons, update invalidation, bfcache behavior, and subpath support (manifest, service worker, install icon, production registration, and Vite subpath support implemented; update invalidation, bfcache, and install-flow evidence remain)
- [~] P049 Implement MCP server/tool management and plugin settings/status using server-owned validation and secret filtering (server-owned MCP CRUD, tool metadata, endpoint/transport validation, plugin status/settings, and secret-key filtering are implemented and tested; live Hermes MCP execution and full plugin lifecycle remain)
- [~] P050 Implement extension registry/install/toggle/uninstall/settings, skin/TTS/nav capability registration, sidecar consent/proxy, iframe tab, and trust boundaries (sanitized read-only plugin/extension registry contracts and explicit unavailable mutation/trust states added; sandboxed execution, consent, iframe, and full settings remain)
- [ ] P051 Complete push-to-talk, hands-free mode, raw audio/transcription capability, browser/server TTS engines, rate/pitch/voice, and unsupported states
- [~] P052 Complete update check/apply/lock recovery, shutdown/restart controls, version/health summaries, and release diagnostics (read-only version/update contracts expose health links, release evidence state, and supervisor-owned unavailable actions; signed update source, apply, lifecycle control, and hosted release evidence remain)
- [ ] P053 Match Docker single/two/three-container topologies, subpath reverse proxy, multi-arch healthchecks, installer behavior, and state migration/rollback
- [~] P054 Add auth, onboarding, settings, locale, PWA, extension, and distribution acceptance matrices (platform contract and shell matrix coverage added; full domain matrices remain)

## M11 — MVP certification and reversible cutover

- [~] P055 Build sanitized personal-state fixtures plus a copied-state migration rehearsal; never run migration against the production state directory (fixture generator refuses production Hermes paths; copied-state rehearsal remains)
- [~] P056 Run full contract comparison for required personal route/state effects and record every intentional alias or deviation (local route inventory and focused backend matrix pass; full personal state-effect comparison remains)
- [~] P057 Run visual matrices across required viewports, System/Dark/Light, all release skins, long lists/transcripts, active streaming, dialogs, and both side panels (252 local viewport/theme/skin rows pass with screenshots outside repo; shell geometry runner covers rail/sidebar/collapse/mobile visibility and capability-card filtering; long/live/reference pixel rows remain, and the parent sandbox cannot launch Chromium for the full visual runner)
- [~] P058 Run accessibility checks for keyboard-only flows, focus restoration, screen readers, contrast, reduced motion, touch targets, RTL, and IME (local keyboard/focus/reduced-motion/touch/RTL/IME audit passes; screen-reader, contrast, and device validation remains)
- [ ] P059 Run live Hermes/Gateway side-by-side rows for chat, tools, subagents, approvals, clarification, compression, recovery, external channels, and scheduled work (local runner now proves Hermes CLI/provider identity, BFF health, one real chat marker, and isolated session parity; full tool/subagent/approval/channel/scheduled-work side-by-side rows remain open)
- [~] P060 Prove 1,000-session and 5,000-message behavior, JS gzip, idle RSS, startup, first-token overhead, stream rendering, and no per-token full refetch (local JS gzip/no-refetch evidence passes; scale, RSS, startup, and live latency remain)
- [~] P061 Run independent security review for auth/proxy, SSRF, XSS/Markdown, workspace/terminal/git, uploads/extraction, extensions/sidecars, migration, and secret boundaries (CSP, Permissions-Policy, headers, and automated platform checks added; independent review and full threat matrix remain)
- [ ] P062 Run hosted Linux/macOS/Windows artifact, Docker, Nix, PWA/subpath, and browser matrices from release candidates
- [ ] P063 Complete parallel beta with the personal WebUI available for rollback and resolve every critical/major parity gap
- [~] P064 Rehearse backup, migration, release validation, rollback, and operator communication using copied state (sanitized copied-state rollback script passes; migration/release/operator communication rehearsal remains)
- [ ] P065 Request explicit user cutover approval; do not archive or modify the personal WebUI before approval
- [~] P066 Keep rollback artifacts and the personal implementation available until the agreed rollback window closes (local copied-state rehearsal is available; user-approved rollback window and hosted artifacts remain)

## M12 — Reference UI convergence

These tasks capture the remaining implementation work identified by the literal
Hermes Studio design-system audit. They are separate from the existing feature
parity tasks so visual convergence is not mistaken for backend certification.

- [x] P067 Promote the shared ContextRail and seam collapse control to the full Chat, Skills, Memory, Profiles, and Settings layout shell (shared header/body zones, seam control, responsive widths, compact chat session rail, and one body scroll owner are implemented and browser verified)
- [x] P068 Implement the reference row variants inside the shared shell: compact Chat session rows with tags/source filters/dates, Skill toggle rows, Settings/Memory navigation rows, Profile status rows, and the Profiles feature-card/info slot (Chat, tags, and Profiles density are implemented and tested)
- [x] P069 Apply the audited page grid and density contract: responsive 64px icon rail with glowing badge, 52–64px navbar with centered active context, 270–300px sidebar, independent sidebar/content scrolling, seam control, and compact floating composer (verified with browser acceptance suites)
- [x] P070 Apply the literal type/color/component tokens across navbar, rail, sidebar, empty states, settings cards, chat transcript, thinking blocks, and composer pills (semantic tokens, compact utility typography, 900px transcript, and 1000px floating composer with pill dropdowns implemented and verified)
- [x] P071 Complete navbar parity: centered active conversation badge with turn count, right-side state controls, compact workspace panel, and streamlined chat toolbar with bookmark prompts and pill dropdowns (browser tests and build verified)
- [~] P072 Add reference visual acceptance rows for Chat, Skills, Memory, Profiles, and Settings at required viewports, collapsed/expanded rails, long lists, empty/detail states, themes, and skins; record screenshot/DOM/computed-style evidence (bounded shell geometry and local API acceptance runners added; Chromium/reference visual evidence remains open)
- [x] P073 Extend the shared contextual sidebar shell to Tasks, Spaces, Todos, Goals, Terminal, and Insights; preserve each view's existing content and server-backed states (shared full-height rail/collapse seam and responsive content width implemented; full reference screenshot comparison remains part of P072)
- [x] P074 Align Skills/Memory/Settings rail tools with the Chat rail treatment and move password setup/sign-in into Settings (full-width sticky rail search and Settings account section implemented; browser screenshot comparison remains part of P072)
