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

## 2026-08-31 - M7-M12 local acceptance coverage

- Added `scripts/local-hermes-acceptance.sh`, a safe read-only route matrix for
  the BFF-owned M7-M12 surfaces: readiness, diagnostics, capabilities,
  profiles/providers, preferences, discovery, control collections, operator
  views, and sessions. It validates JSON content types and rejects credential
  field names in the diagnostics payload.
- The runner invokes the existing live smoke/session probes only when
  `/api/health/hermes` reports `reachable: true`; otherwise it records an
  explicit live skip and leaves P059 open. It creates no state unless the
  existing isolated live parity probe is actually eligible to run.
- Local execution on this date: the route matrix and diagnostics sanitization
  checks passed. A first smoke attempt exposed that the previous smoke script
  accepted an empty `done` answer; the script now requires the exact answer
  marker. The follow-up probe received `answer:""` and the local BFF/Gateway
  process was no longer available, so no live chat, tool, approval, subagent,
  or side-by-side claim was made. The visual browser runner remains
  environment-blocked when Chromium cannot launch in the sandbox.
- `scripts/local-acceptance.sh` now includes this runner before the visual
  matrix, so future local Hermes sessions produce a single repeatable evidence
  trail without conflating offline capability states with passing live gates.

### Follow-up live observation

- After the local Hermes runtime became available, the strengthened runner
  passed the CLI/provider marker (`Hermes Agent v0.19.0`, upstream marker
  `d604141d`), all read-only M7-M12 route JSON checks, diagnostics sanitization,
  the real BFF smoke marker, and isolated session CRUD/chat/history/truncate
  parity. This is local Hermes evidence only.
- P059 remains open because no structured approval, subagent, external-channel,
  scheduled-work, compression, or recovery side-by-side row was observed in
  this run. P062/P063/P065 and reference visual gates remain open as well.

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

## 2026-08-30 — P009 navigation customization

- Added a functional Customize navigation Dialog with persisted visibility and
  order controls. Chat remains locked visible, and missing reference surfaces
  are not presented as dead links.
- Added browser coverage for the visibility toggle and Dialog completion.
- Frontend tests and production build pass. P009 remains `[~]` pending the
  complete reference Control Center section registry and broader panel matrix.

## 2026-08-30 — P010 theme boot and skin registry

- Added pre-React System/Dark/Light theme boot and validated `data-theme` and
  `data-skin` root attributes through browser acceptance.
- Added a validated registry for all 21 frozen skin names and shared token
  variants for the tuned palette groups.
- Preferences now preview and persist theme, skin, and locale through the
  server route plus local layout storage.
- Frontend tests and production build pass. P010 remains `[~]` pending
  per-skin visual tuning and the P012 screenshot matrix.

## 2026-08-30 — P011 mobile navigation and safe-area shell

- Added five working mobile bottom tabs for Chat, Tasks, Skills, Memory, and
  Settings, with 44px minimum targets.
- Added safe-area-aware composer spacing, mobile panel transition, and reduced
  motion fallback while preserving the drawer and workspace slide-over.
- Browser acceptance covers mobile tab navigation and target sizing.
- P011 remains `[~]` pending device-level IME, safe-area, and full transition
  evidence.

## 2026-08-30 — Desktop mobile-nav visibility and composer toolbar

- Fixed the custom bottom navigation CSS overriding `lg:hidden`, so it is
  hidden at the desktop breakpoint while remaining available on mobile.
- Split composer controls from composer actions to keep profile, model,
  workspace, hint, cancel, and send controls readable without overlap.
- Added desktop browser coverage for the hidden bottom navigation state.

## 2026-08-30 — Responsive drawer regression

- Closed the mobile navigation state when `matchMedia('(min-width: 1024px)')`
  changes to desktop, preventing the fixed drawer from leaking into desktop.
- Added browser acceptance for mobile-open followed by desktop resize.
- P011 remains `[~]` pending device-level IME, safe-area, and full transition
  evidence.

## 2026-08-30 — Parallel parity wave: controls, recovery, and discovery states

- Replaced visible native selects with a shadcn-style trigger/listbox while
  preserving controlled form and change-event behavior.
- Added a validated inflight turn journal, transcript restore, cursor-aware
  reconnect, and stale-stream protection for session switch and hard reload.
- Separated Skills/Memory preview loading, empty, success, and error states.
- P013-P015 and P034 remain `[~]`; full source fixtures, expired-stream polling,
  live reload proof, and complete Skills parity are still open.

## 2026-08-30 — Parallel remaining-task wave

- Added Queue/Interrupt/Steer turn planning, visible pending intent, attachment
  names, per-item removal, replacement behavior, and exit cleanup.
- Added server-owned Skills query/create/update/delete with path containment and
  separate editor/preview states.
- Added a five-viewport MVP shell evidence runner with screenshots and DOM/
  computed geometry output stored outside the repository.
- Added PWA manifest/service worker/base-path registration, platform contract
  checks, security headers, and a sanitized fixture generator that rejects
  production state paths.
- P012, P016, P034, P048, P054, P055, and P061 remain `[~]` pending their
  reference, live, hosted, or independent-review gates.

## 2026-08-30 — Parallel parity wave two

- Added conversation-runtime helpers for disclosure modes, compaction/repair,
  branch lineage, projections, model search, safe exports, and lifecycle rows.
- Added server-owned Spaces activation/health, Kanban board/card CRUD, task
  history/delete controls, Memory note CRUD/search safety, and operator health
  and logs routes.
- Added profile/provider CRUD with credential redaction, 15-locale fallback,
  PWA update/bfcache hooks, release artifact checks, and copied-state rehearsal.
- P017-P026, P027, P029, P032-P036, P038, P040-P042, P045, P047, and P064 are
  now `[~]`; external/live/complete parity gates remain explicitly open.

## 2026-08-30 — PWA API cache boundary

- Prevented the service worker from caching `/api/*` responses so stale
  session/control data cannot leak across reloads or override live server data.
- Isolated browser acceptance contexts from service-worker cache while keeping
  the production shell cache coverage explicit.

## 2026-08-31 — Remaining-task operator and auth wave

- Added server-owned Insights with usage, provider history, synchronization,
  loading/empty/error states, and an honest cost-unavailable state.
- Added explicit Terminal unavailable capability handling and tests instead of
  exposing an unsafe fake process surface.
- Hardened auth cookie rotation, HTTPS reverse-proxy Secure behavior, logout
  origin checks, and onboarding validation recovery.
- P030, P037, P043, and P044 are now `[~]`; contained terminal lifecycle,
  authoritative cost data, passkey/OIDC, branded onboarding, and external
  review gates remain open.

## 2026-08-31 — Workspace and certification evidence wave

- Added relative-path-safe workspace tree/preview/download/edit/basic CRUD,
  upload, copy/open actions, read-only Git branch/status/diff, and hostile-path
  tests. Reveal and mutating Git operations remain explicitly unavailable.
- Added contract, performance, visual, accessibility, local acceptance, and
  copied-state rollback runners. The visual matrix passed 252 local
  viewport/theme/skin rows; frontend tests/build, backend tests/vet, contract,
  secret, platform, and gzip checks passed.
- Accessibility evidence now verifies session-action focus restoration and the
  44px target contract locally; therefore P028/P031/P046/P056/P057/P058/P060/P064/
  P066 remain `[~]` and no live, hosted, beta, or cutover gate is claimed.

## 2026-08-31 — External session projection slice

- Added a server-metadata-driven external session handoff dock in the session
  rail, including channel/source, identity, and routing labels with an honest
  unavailable handoff action.
- P039 is now `[~]`; live channel transport, round summaries, identity
  semantics, and model-switch warnings still require Gateway/live evidence.

## 2026-08-31 — Safe JSON parsing for capability panels

- Fixed Terminal, Insights, and Spaces error handling when an older API
  process returns plain-text `404 page not found` instead of JSON. The shared
  frontend response reader now preserves structured API messages and converts
  non-JSON responses into an actionable HTTP error instead of exposing a JSON
  parser exception.
- Added a regression test for the legacy plain-text proxy response. The API
  process must still be restarted after deploying routes introduced by a newer
  commit.

## 2026-08-31 — Rail and settings navigation cleanup

- Increased shared interactive targets and rail icon scale, removed duplicate
  Settings/Customize actions from the primary rail, and moved New conversation
  into the Chat recent-session sidebar.
- Added a Settings-local navigation sidebar with Conversation, Appearance,
  Preferences, Providers, Plugins, Extensions, System, Help, and Customize
  navigation entry points while preserving the existing server-backed
  preference behavior.

## 2026-08-31 — Literal design-system gap tracking

- Recorded the Claude design audit as a separate M12 convergence backlog rather
  than marking existing UI work as complete. The audit identifies missing
  Profiles rail/row parity, page-specific shared-shell row variants, unified
  seam ownership, exact 88px/64px/552px geometry, and screenshot/computed-style
  evidence.
- Existing ContextRail work is documented as partial: Chat, Skills, Memory,
  and Settings use the shared shell, while Profiles and the full reference
  visual matrix remain open.

## 2026-08-31 — M7-M12 bounded implementation wave

- Added a typed local slash-command registry for `/help` and `/clear`, keyboard
  autocomplete, and transparent pass-through for unknown slash commands.
- Added a sanitized read-only `/api/operator/diagnostics` route for component
  availability and collection/session counts without credentials or filesystem
  paths.
- Hardened profile/provider contracts: unavailable providers cannot be
  activated, the Gateway provider and active-profile dependencies cannot be
  deleted, and preference filtering rejects secret-shaped keys.
- Added the M11/M12 shell geometry runner for rail bounds, mobile nav
  visibility, shared utility rails, collapse/expand, and filtered capability
  status behavior.
- Evidence: frontend tests (31) and production build pass; backend validation
  is rerun from the backend module. The full local acceptance runner reaches
  the visual phase but Chromium is blocked by the parent macOS sandbox, so no
  full visual/live/hosted/beta/cutover gate was marked complete.
- Checklist impact: P024, P038, P041, P042, P057, P067, P069, and P071 have
  stronger bounded evidence but remain `[~]`; P068 and P072 remain open.

## 2026-08-31 — M10 capability and operations contracts

- Added server-owned read-only `/api/extensions` and strengthened `/api/plugins`
  to return sanitized metadata plus explicit registry, execution, settings,
  sidecar, iframe, and trust-boundary capability states. Plugin settings are
  not returned to the browser.
- Added `/api/operator/version` and `/api/operator/update` for runtime version,
  health-route links, release-artifact evidence state, and explicit unavailable
  update/apply/shutdown/restart/lock-recovery actions. No browser mutation or
  process-control endpoint was added.
- Added route and secret-boundary tests and updated the compatibility contract,
  feature manifest, and M10 checklist. P050/P052 remain `[~]` because a real
  sandboxed extension runtime, signed update source, supervisor integration,
  and hosted release evidence are not present.

## 2026-08-31 — M10 P049 MCP and plugin settings slice

- Added validated, server-owned MCP server metadata CRUD and read-only tool
  metadata retrieval. The BFF does not launch MCP processes or execute tools.
- Added plugin status/settings updates for discovered plugins. Secret-shaped
  settings are filtered before persistence and are never returned to the UI.
- Focused backend tests cover invalid transports/endpoints, safe settings,
  tool metadata, CRUD, and discovered plugin updates. P049 remains `[~]` for
  live MCP transport, tool discovery, plugin execution, consent, and sandbox
  trust evidence.

## 2026-08-31 — M12 Profiles convergence slice

- Moved Profiles into the existing `ContextRail` shell and reused its shared
  collapse/expand seam control. The Profiles data API, activation action,
  loading state, empty state, and safe error path remain unchanged.
- Added the Profiles reference feature slot and compact status rows: runtime
  model/provider metadata, health dot, active badge, profile selection, and a
  detail pane. A dense mobile fallback keeps profile selection reachable when
  the desktop rail is hidden.
- Added focused profile presentation contract tests. Frontend tests now cover
  33 cases and the production build passes. This is bounded local evidence;
  P067/P068 remain partial until exact reference geometry and visual/live
  acceptance evidence are complete.

## 2026-08-31 — M10 P049 MCP and plugin management slice

- Added server-owned MCP server metadata CRUD with transport, endpoint,
  command, tool-count, and input-size validation. Tool metadata is exposed by
  a separate read-only route and no MCP process or remote endpoint is started.
- Extended plugin status with persisted enabled/settings state and added a
  constrained settings update route for discovered plugins. Secret-shaped
  keys such as tokens, passwords, API keys, credentials, and private keys are
  filtered before persistence and response.
- Evidence: focused MCP/plugin HTTP tests, backend vet, frontend tests (33),
  and production build pass. Full backend test execution remains affected by
  the sandbox refusing an existing httptest listener to bind a local port;
  live Hermes MCP execution and plugin lifecycle remain unverified.
- Checklist impact: P049 is now `[~]`; P050 remains partial because install,
  execution, consent, iframe, and trust-boundary contracts are not included.
-
## 2026-08-31 — Local Hermes acceptance follow-up

- The bounded `scripts/local-hermes-acceptance.sh` runner passed the local
  Hermes CLI marker probe, BFF JSON route matrix, diagnostics sanitization,
  live smoke chat, and isolated M1 session lifecycle.
- This is partial live evidence for chat and session persistence only. No
  claim was made for structured tools, subagents, approvals, external
  channels, scheduled work, hosted artifacts, or reference visual parity.
- `hermes serve` was separately identified as a headless JSON-RPC surface,
  not the OpenAI-compatible Gateway expected by the Studio BFF.
