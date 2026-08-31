# Hermes WebUI → Hermes Studio Parity Migration Plan

> Status: approved planning baseline; implementation not started by this plan
> Source: `/home/adityahimaone/hermes-webui` current working tree, including uncommitted changes
> Target: `/home/adityahimaone/apps/hermes-web-studio`, branch `dev`
> Scope priority: Chat, Kanban, Spaces, Profile
> Strategy: full rewrite/migration into Go + React Studio, preserving observable workflow and state contracts while allowing implementation changes

## Goal

Make Hermes Studio a replacement for the personal Hermes WebUI for Chat, Kanban, Spaces, and Profile. Preserve feature responsibility, visual hierarchy, component positions, keyboard behavior, responsive behavior, persistence, error/loading/empty states, and Hermes-side effects. Do not claim 100% parity from route existence or mock tests alone.

Visual rule: clone reference geometry first. Any visual or interaction deviation needs an explicit decision record and acceptance evidence. Studio may use its own React/shadcn implementation, but users must recognize equivalent regions, controls, ordering, density, and state transitions.

## Baseline and evidence rules

- Freeze source before each slice. Initial source audit references the current working tree, not only the last commit.
- Source commit observed during planning: `c7c1a8ff069f1a7a9f317c61871a24f280b829f6`.
- Relevant source working-tree fingerprint recorded during planning: `a85c9756a6e2156bda60fa3f4590771dc3f6c80e8309bc9746f344772efea893`.
- Target current branch: `dev`.
- Existing Studio contracts remain authoritative guardrails: `MVP.md`, `TASKS.md`, `docs/compatibility-contract.md`, `docs/shell-contract.md`.
- Private production state, credentials, session names, provider keys, and real workspace paths never become fixtures.
- A parity item is `[x]` only after automated evidence plus browser/manual side-by-side evidence where visual or live Hermes behavior matters. Use `[~]` while evidence is incomplete.
- Keep original WebUI available through certification and rollback window. Never migrate destructively in place.

## Current architecture

### Source responsibilities

The personal WebUI is a vanilla HTML/CSS/JavaScript application with large behavior modules:

- `static/index.html`: complete shell, primary rail, titlebar, panels, dialogs, responsive markup.
- `static/style.css`: visual tokens, geometry, responsive breakpoints, panel states, modal/drawer styling.
- `static/panels.js`: panel switching plus Kanban, Spaces/workspace registry, Profiles, settings/control-center behavior.
- `static/sessions.js`: session list, search, grouping, source projections, actions, lineage, archive/pin/project/tag state.
- `static/messages.js`: composer, streaming, reconnect/replay, approvals, clarification, tool/subagent cards, branching, recovery.
- `static/workspace.js`: tree, preview, CRUD, upload/download, artifact projection, path safety.
- `static/terminal.js`: terminal lifecycle and resize behavior.
- `api/routes.py`, `api/kanban_bridge.py`, `api/workspace.py`, `api/profiles.py`, and related modules: server contracts, persistence, safety, and Hermes integration.

Source scale is large: `static/panels.js` ~14.6k lines, `static/sessions.js` ~9.5k, `static/messages.js` ~9.3k, `api/routes.py` ~29.6k. Do not port by superficial file copying. Extract observable contracts and migrate vertical slices.

### Studio ownership

- App/shell: `frontend/src/app.tsx`, `frontend/src/app.css`, `frontend/src/components/layout/{sidebar,context-rail,view-shell}.tsx`.
- Chat: `frontend/src/components/chat/{session-rail,message-list,activity-cards,composer,connection-status}.tsx`, `frontend/src/hooks/use-chat.ts`, `frontend/src/lib/{chat-contract,conversation-runtime,turn-control,markdown}.ts(x)`.
- Workspace: `frontend/src/components/workspace/workspace-panel.tsx`, `frontend/src/hooks/use-workspace.ts`, `frontend/src/lib/workspace-contract.ts`.
- Control center/Kanban: `frontend/src/components/control/{control-center,kanban-view}.tsx`, `frontend/src/lib/kanban-client.ts`, `frontend/src/lib/control-client.ts`.
- Profiles: `frontend/src/components/auth/identity-controls.tsx`, `frontend/src/lib/profile-contract.ts`, profile handlers in `backend/internal/httpapi/auth_profiles.go`.
- Backend routes: `backend/internal/httpapi/server.go`, `workspace.go`, `operator.go`, `kanban_native.go`, `auth_profiles.go`; state services in `backend/internal/{session,workspace,control,auth,gateway}`.
- Existing tests: frontend `*.test.ts(x)`, `frontend/e2e/*.mjs`; backend `*_test.go`.

## Non-negotiable contracts

1. Browser never receives Hermes Gateway keys, provider credentials, CLI paths, subprocess details, OAuth secrets, or OS absolute workspace paths.
2. Gateway events normalize into Studio `ChatEvent`; UI never parses provider-specific event variants.
3. Session, Space, Kanban, and Profile state must have one authoritative owner. No competing local fallback that can diverge from server state.
4. Readers ship before writers. Legacy unknown fields survive migration. Writes are atomic and idempotent.
5. Relative workspace paths only. Resolve symlinks before containment checks. Reject traversal, absolute paths, backslashes, hostile names, oversized content, and unsafe uploads.
6. Preserve loading, empty, success, error, cancellation, reconnect, replacement, and permission-denied states.
7. Every visible essential action stays visible; no hover-only copy/edit/delete/stop controls.
8. Desktop and mobile use intentional layouts, not only hide/show rules. Touch targets minimum 44 CSS px.
9. Live Hermes behavior cannot be certified from mocks. Record live model/provider and sanitized evidence.
10. No broad redesign mixed with protocol/data migration. First match reference structure; polish only after parity gate.

## Shared shell and visual parity foundation

### Reference geometry to measure

Capture source and Studio at required viewports: desktop wide, desktop narrow, tablet, 375×812 mobile, and 390×844 mobile. Capture expanded/collapsed rail, open/closed session sidebar, open/closed workspace panel, active stream, modal/dialog, empty list, long list, and profile/Space detail states.

Measure and assert:

- 64 px desktop icon rail and its active/glowing badge treatment.
- 52–64 px titlebar/navbar, centered active context, right-side controls.
- 270–300 px contextual sidebar/session rail.
- Chat transcript max width/density, composer width, bottom offsets, and panel seams.
- Workspace panel width and resize bounds.
- Mobile drawer, five-tab/bottom navigation where applicable, safe-area padding, keyboard behavior.
- Typography roles, color tokens, border/radius/shadow, focus ring, disabled and status indicators.

### Shared implementation

Before feature slices, stabilize shared primitives and hooks:

- `frontend/src/components/ui/{button,dialog,dropdown-menu,input,select,textarea,tooltip,badge}.tsx`.
- Shared navigation/sidebar/context rail ownership and collapse seam.
- Stable `data-testid`/ARIA hooks for rail, titlebar, panel, composer, drawer, modal, active profile, active Space, and Kanban board.
- One scroll owner per region. Streaming must not reset sidebar or workspace scroll.
- Theme boot before React; System/Dark/Light semantics; reduced-motion handling.

Tests:

- Extend shell tests in existing frontend suites and Playwright E2E.
- Add screenshot/DOM/computed-style matrix runner with sanitized fixtures.
- Verify keyboard focus, Escape, Enter/Shift+Enter, dialog focus restoration, drawer transitions, and 44 px targets.

## Phase 1 — Chat parity

### Source-to-target map

| Reference behavior | Source | Studio owner |
|---|---|---|
| Chat shell/titlebar/rail | `static/index.html`, `static/style.css` | `app.tsx`, layout components, `app.css` |
| Session list/search/grouping/actions | `static/sessions.js`, `static/index.html` | `session-rail.tsx`, session API/hooks |
| Transcript/rendering | `static/messages.js`, `static/smd.min.js`, KaTeX/Mermaid assets | `message-list.tsx`, `markdown.tsx`, `mermaid.tsx` |
| Composer/send/queue/interrupt/steer | `static/messages.js`, `static/terminal.js` | `composer.tsx`, `use-chat.ts`, `turn-control.ts` |
| Stream/reconnect/replay/cancel | `static/messages.js`, `api/streaming.py`, `api/gateway_chat.py` | `gateway`, `server.go`, chat reducer/runtime |
| Tool/reasoning/subagent/approval/clarify | `static/messages.js`, `api/route_approvals.py`, `api/clarify.py` | `activity-cards.tsx`, normalized `ChatEvent` |
| Sessions persistence/actions | `api/agent_sessions.py`, `api/session_*`, `api/webui_session_db.py` | `backend/internal/session`, `server.go` |

### Required behavior

1. Session rail: WebUI/CLI/external source labels, search, content/title matching, date groups, project/tag chips, pin/archive, duplicate, export, delete, batch selection/actions, unread/streaming indicators, lineage/fork visibility, and active-row preservation.
2. Titlebar: active profile/context, session title, turn/state controls, workspace toggle, usage/context indicator, new chat, overflow ownership, and responsive ordering.
3. Transcript: user/assistant/system/tool/reasoning/subagent/approval/clarification cards; Markdown, code copy/highlight, safe links, tables, Mermaid, KaTeX, media tokens, timestamps, JSON/YAML behavior where reference exposes it.
4. Lifecycle: send, stop, retry/regenerate, edit/truncate, queue, interrupt, steer, attachment add/remove, session switch, reload, reconnect backoff, cursor replay, duplicate suppression, compression barrier, partial-turn recovery, branch/fork, clear watermark.
5. Live events: Runs-first normalization for message, reasoning, tool progress/result, subagent, approval, clarification, usage, compression, recovery, terminal, completed, failed, and cancel. Unknown events remain visible as safe diagnostic state, never silently dropped.
6. Composer: model/provider/profile/workspace selectors, saved prompts/commands, selected-text reply, attachments, slash autocomplete, terminal/runtime controls, mode persistence and visible pending intent.

### Implementation order

- Freeze deterministic sanitized sessions and event fixtures from source contracts.
- Complete backend session reader/writer and compatibility aliases before richer UI.
- Finish normalized event reducer and lifecycle state machine.
- Implement session rail and titlebar geometry against captured reference.
- Implement transcript cards and rendering hardening.
- Implement composer modes and attachments.
- Add replay/reconnect/reload/session-switch browser tests.
- Run live Gateway side-by-side matrix before marking complete.

### Chat files likely to change

- `frontend/src/app.tsx`, `frontend/src/app.css`.
- `frontend/src/components/chat/*.tsx`.
- `frontend/src/components/layout/*.tsx`.
- `frontend/src/hooks/use-chat.ts`.
- `frontend/src/lib/chat-contract.ts`, `conversation-runtime.ts`, `turn-control.ts`, `markdown.tsx`, `api-client.ts`.
- `backend/internal/gateway/*`, `backend/internal/httpapi/server.go`.
- `backend/internal/session/*`, attachment handling, tests.
- `frontend/e2e/*`, frontend chat tests, backend server/gateway/session tests.
- `TASKS.md`, `docs/compatibility-contract.md`, `docs/decisions.md`.

### Chat acceptance gate

- Sanitized normal/error/cancel/reconnect/reload/switch/retry/approval/clarify/tool/subagent rows pass.
- Browser geometry and keyboard matrix matches reference within recorded tolerances.
- Live Hermes stream, stop, reconnect/replay, attachment, and restored history pass.
- No secrets in built JS, network payloads visible to browser, logs, snapshots, or errors.

## Phase 2 — Spaces and Workspace parity

### Source-to-target map

| Reference behavior | Source | Studio owner |
|---|---|---|
| Space list/detail/create/edit/delete/switch | `static/panels.js` workspace functions, `api/workspace.py`, `api/routes.py` | `control-center.tsx`, `workspace-panel.tsx`, `server.go`, control store |
| Workspace tree/breadcrumb/preview | `static/workspace.js`, `static/index.html`, `api/workspace.py` | `workspace-panel.tsx`, `use-workspace.ts`, `workspace` service |
| Git badges and operations | `static/workspace.js`, `api/workspace_git.py` | backend workspace/git adapter, workspace UI |
| Terminal | `static/terminal.js`, `api/terminal.py` | contained terminal contract; never fabricate unavailable state |
| Artifacts/Todos projection | `static/workspace.js`, `api/todo_state.py` | workspace tabs and control store |
| Remote Mac/Tailscale execution | source workspace/runner/worktree modules and deployment behavior | server-side transport adapter; explicit remote health/ownership state |

### Space model

Introduce a server-owned registry with stable ID, display name, location kind, workspace reference, host/transport metadata, ordering, active flag, health, last error, and profile ownership. Keep OS path/SSH details server-side. Active Space must resolve for Chat workspace selection, Kanban workspace, file operations, terminal, artifacts, and profile-local defaults.

Space operations:

- list registered Spaces in stable order;
- create/update/delete with validation and active/deletion protection;
- activate one Space and persist per profile where reference behavior does so;
- health probe with loading/healthy/degraded/offline/permission states;
- preserve selected Space on reload and session switch;
- distinguish local workspace, remote Mac via configured SSH/Tailscale transport, and unavailable transport;
- never silently fall back from remote Mac to VPS-local path.

### Workspace behavior

- Tree lazy loading, hidden-file rule, breadcrumb, filemap, copy relative path, open/reveal status, preview text/code/Markdown/image/binary, download.
- Create file/folder, edit/write, rename/move, delete, upload, paste/extract only when safe contract exists.
- Files/Artifacts/Todos tabs with independent tab state and stream-safe preview persistence.
- Git branch/status/diff initially read-only; stage/discard/commit/push/pull/stash/checkout require confirmation, ownership, hostile-path tests, and rollback.
- Terminal start/input/output/resize/close and cleanup on success/error/cancel/replacement/disconnect. If transport is unavailable, show explicit unavailable state.
- Remote execution canary must verify actual remote host identity, cwd, path, and process ownership before accepting a Space as healthy.

### Space acceptance gate

- Same Space list/detail positions and controls at desktop/mobile.
- Active Space inheritance proven across Chat, Kanban, workspace preview, terminal/artifacts, and Profile.
- Root containment, symlink, traversal, upload, size, permission, and hostile-path tests pass.
- Remote Mac path never executes in VPS-local cwd; real SSH/Tailscale canary passes with sanitized host identity.
- Reload, session switch, profile switch, disconnect, and transport failure preserve or clear state according to reference contract.

## Phase 3 — Kanban parity

### Source-to-target map

| Reference behavior | Source | Studio owner |
|---|---|---|
| Kanban panel/list/filter layout | `static/index.html`, `static/style.css`, `static/panels.js` | `kanban-view.tsx`, shared contextual shell |
| Board/task CRUD and detail | `static/panels.js`, `api/kanban_bridge.py`, related API modules | `kanban-client.ts`, `backend/internal/httpapi/kanban_native.go` |
| Dispatch/worker/runtime | `api/kanban_bridge.py`, agent runtime/runner modules | server-side Hermes CLI/Dashboard adapter |
| Space/workspace binding | source workspace/kanban functions | Space resolver shared with workspace/chat |
| Events/stats/history | source Kanban routes and UI renderers | backend event adapter + Kanban view |

### Required behavior

1. Board selector/list/create/rename/archive and stable selected-board persistence.
2. Dashboard/list/board modes, lane ordering, canonical statuses: `triage`, `todo`, `scheduled`, `ready`, `running`, `blocked`, `review`, `done`, `archived`.
3. Search, assignee, tenant/Space, archived, only-mine, status filters; filter state survives refresh as reference dictates.
4. Task create/edit/detail/delete/archive, title/description, status, priority, assignee, parent/child, skills, model/provider, Space/workspace, schedule, retry, error, output, timestamps, comments, links, and metadata.
5. Drag/drop lane transitions with keyboard alternative, optimistic state only when rollback/error is explicit, and no duplicate writes.
6. Bulk selection/actions, dispatcher preview/nudge/run, pause/resume/cancel where supported, worker ownership, event stream/watch, stats, history, and failure recovery.
7. Task detail drawer/modal position, scroll ownership, focus, Escape, mobile sheet behavior, and action ordering match reference.
8. Native Hermes state is canonical. Do not use local fallback for production Kanban. CLI is default only where its capability is proven; Dashboard transport may unlock richer mutation/events.

### Implementation order

- Freeze sanitized boards/tasks/status fixtures and inspect every source action.
- Complete native transport capability matrix and explicit unavailable states.
- Implement board/task reads and selected board/Space binding.
- Implement task create/detail/action flows with server-side validation.
- Implement board lanes, filters, drag/drop + keyboard movement.
- Implement comments/links/bulk actions and dispatcher lifecycle.
- Implement event stream, stats, history, and runtime recovery.
- Run remote Space Kanban canary, then side-by-side visual matrix.

### Kanban files likely to change

- `frontend/src/components/control/kanban-view.tsx`, `control-center.tsx`.
- `frontend/src/components/layout/context-rail.tsx`, shared dialogs/selects.
- `frontend/src/lib/kanban-client.ts`, control/state hooks.
- `backend/internal/httpapi/kanban_native.go`, `operator.go`, `server.go`.
- New narrowly scoped backend Kanban transport/event files only if current native file becomes multi-responsibility.
- `backend/internal/control/*`, tests; frontend Kanban tests and Playwright E2E.
- `docs/compatibility-contract.md`, `docs/kanban-plan.md`, `TASKS.md`.

### Kanban acceptance gate

- Board/task state reads from Hermes-owned transport, not browser/local fallback.
- Every unsupported mutation says unavailable and does not pretend success.
- CRUD, filters, drag/drop, keyboard movement, drawer/modal, dispatch, events, stats, error/retry, and Space binding pass fixtures and live capability checks.
- Remote Mac assignment proves remote cwd/host and no VPS-local execution.

## Phase 4 — Profile parity

### Source-to-target map

| Reference behavior | Source | Studio owner |
|---|---|---|
| Profile list/status/detail | `static/panels.js` profile functions, `static/index.html`, CSS | `control-center.tsx`, `identity-controls.tsx` |
| Create/edit/clone/delete | `static/panels.js`, `api/profiles.py` | profile feature components + `auth_profiles.go` |
| Switch and titlebar context | `static/panels.js`, `static/sessions.js` | `app.tsx`, titlebar/sidebar, profile contract |
| Provider/model/personalities | `api/profiles.py`, `api/providers.py`, settings surfaces | backend safe metadata + profile editor |
| Profile-local sessions/Spaces/workspace/state | source config/state resolution modules | server-side profile scope and active resolver |

### Required behavior

- Profile cards/list density, status/health label, active marker, model/provider labels, action placement, empty/loading/error states.
- Create, edit, clone/duplicate, delete with active/last-profile protection and confirmation dialog.
- Switch profile updates titlebar, Chat session projection, model/provider defaults, Space/workspace defaults, Skills/Memory visibility, and profile-local state without leaking another profile's data.
- Provider/model picker: grouped/searchable live metadata, custom endpoint state, aliases, reasoning/toolset fields, quota/cost/TPS only when authoritative, unavailable/error state otherwise.
- Personalities/prompt/soul/skills configuration only through validated server-owned paths and safe metadata.
- OAuth linking, credentials, and secrets remain server-side; UI shows `has_key`/health, never raw values.
- Concurrent tabs/profile switches must not apply stale responses to active profile.
- Profile-local persistence survives reload and is isolated in copied-state tests.

### Implementation order

- Lock profile identity and safe response schema.
- Add profile scope to sessions, Spaces, workspace, Kanban, Skills/Memory, and preferences where reference requires it.
- Complete profile CRUD/clone/delete/switch and active-protection tests.
- Add provider/model discovery and custom provider validation with redaction.
- Match profile UI geometry and titlebar switch behavior.
- Add auth/OAuth/passkey integration only behind configured capability states.
- Run concurrent isolation, reload, failure, and live provider matrices.

### Profile files likely to change

- `frontend/src/components/auth/identity-controls.tsx` and new feature-local profile components following `src/features/<name>/{components,hooks,api,types}` when extraction is needed.
- `frontend/src/components/layout/sidebar.tsx`, `view-shell.tsx`, `app.tsx`.
- `frontend/src/lib/profile-contract.ts`, API client/state hooks.
- `backend/internal/httpapi/auth_profiles.go`, auth/config/provider files, session/control/workspace stores.
- `backend/internal/auth/*`, tests, migration fixtures.
- `TASKS.md`, `docs/compatibility-contract.md`, `docs/decisions.md`.

### Profile acceptance gate

- Profile create/edit/clone/delete/switch works with safe validation and active protection.
- Session, Space, workspace, Kanban, Skills, Memory, preferences, and model selection isolation proven with two profiles and concurrent requests.
- No credentials in response, browser storage, bundles, logs, snapshots, or error text.
- Password, trusted header, OIDC, WebAuthn, CSRF/origin, cookie rotation, and reverse-proxy behavior match configured capability states.

## Cross-feature dependency order

1. Shared shell geometry and stable selectors.
2. Chat session/runtime contracts.
3. Profile scope primitives, because active profile affects Chat and Spaces.
4. Space registry and resolver, because Chat, Kanban, workspace, and terminal consume it.
5. Full Chat parity.
6. Workspace/remote transport parity.
7. Kanban parity using the same Space resolver.
8. Profile UI/provider/auth completion.
9. Cross-feature isolation and certification.

If implementation discovers a dependency that changes this order, write a decision record before proceeding. Do not bypass a failing earlier gate with a mock or placeholder.

## Data migration and rollback

- Back up `~/.hermes/webui/`, `HERMES_HOME`, profile state, control state, and registered Space metadata before any writer runs.
- Use copied sanitized state for all migration rehearsals. Never point rehearsal commands at production state.
- Version migration schema and record source version/fingerprint.
- Preserve unknown JSON fields and legacy session shape.
- Make migration idempotent; rerun must not duplicate sessions, boards, Spaces, attachments, or profiles.
- Add dry-run, report, backup, restore, and rollback commands before destructive migration.
- Keep old WebUI and Studio deployable side by side with separate ports/state locks until explicit cutover approval.
- Verify rollback by restoring copied state and reopening Chat history, active Profile, active Space, Kanban board/task, attachments, and workspace metadata.

## Testing and verification matrix

### Automated

- Go: `gofmt`, `go test ./...`, race-sensitive store/concurrency tests where supported, `go vet ./...`.
- Frontend: `pnpm test`, `pnpm build`, lint/type checks according to repository scripts.
- Browser: Playwright desktop/narrow/tablet/mobile tests for route, geometry, keyboard, focus, dialogs, drawers, stream/reconnect, CRUD, and error states.
- Security: secret scan, path containment/symlink tests, upload limits, origin/CSRF/auth tests, safe Markdown/XSS tests, hostile Kanban/Space IDs.
- Performance: initial JS gzip, idle RSS, startup, first token, long transcript/list (1,000 sessions / 5,000 messages), no full refetch per token.

### Side-by-side evidence

For each feature, capture source and Studio with identical sanitized fixtures and compare:

- DOM region order and stable labels.
- Bounding boxes and computed styles for rail, sidebar, header, transcript, cards, composer, drawers, dialogs, lanes, cards, profile rows, and Space rows.
- Keyboard tab order/focus restoration.
- State transitions for loading, empty, success, error, cancel, reconnect, permission failure, mobile drawer, and reduced motion.
- Screenshot diff with documented tolerance. Pixel-perfect claims require reference captures; source inspection alone is insufficient.

### Live evidence

Run only against isolated/copy state and approved Hermes/Gateway configuration. Record sanitized timestamp, build, model/provider identity, feature, request/result marker, and known limitations. Never execute destructive commands during approval/remote probes. Live proof required for:

- Chat stream/stop/replay/attachments/tool/approval where available.
- Profile model/provider discovery and switch.
- Space local and remote Mac/Tailscale health/cwd/process ownership.
- Kanban board/task/dispatch/events and remote assignment.

## Milestones and exits

### P0 — Planning and contract freeze

Deliver this plan, source/target manifest, visual matrix schema, fixture policy, dependency graph, and decision log. No implementation claim.

Exit: plan reviewed, committed, pushed; implementation starts only after this baseline is available.

### P1 — Shared shell + Chat

Exit: Chat live stream and full required session/runtime matrix pass; shell geometry passes required viewports.

### P2 — Profile scope + Spaces/Workspace

Exit: profile isolation and Space inheritance pass; local and remote transport safety pass; workspace safety matrix passes.

### P3 — Kanban

Exit: native board/task/dispatch/event behavior passes; visual and mobile matrix passes; remote Space assignment proven.

### P4 — Cross-feature certification

Exit: Chat ↔ Profile ↔ Space ↔ Kanban state remains consistent across reload, switch, reconnect, concurrent tabs, errors, and rollback rehearsal.

### P5 — Reversible beta/cutover

Exit: all critical/major gaps resolved, original WebUI still available, copied-state migration and rollback rehearsed, hosted artifacts/browser/security evidence recorded, explicit user cutover approval received.

## Implementation task protocol

Each implementation task must be one small vertical slice:

1. Read `AGENTS.md`, this plan, `TASKS.md`, and relevant source contract.
2. Add/update compatibility contract and sanitized fixture first.
3. Write failing unit/integration/browser test for observable behavior.
4. Implement minimum backend/API/state path.
5. Implement UI including loading/empty/error/cancel/responsive states.
6. Run focused tests, then neighboring tests.
7. Run visual/keyboard check for visible changes.
8. Update `TASKS.md` with `[x]`, `[~]`, or `[ ]` honestly.
9. Add decision record for deviations.
10. Commit one logical slice. Do not combine unrelated feature, protocol, and migration changes.

## Risks and mitigations

- Source is large and evolving: freeze fingerprint per slice and maintain manifest.
- Existing Studio has implemented foundations but many `[~]` items: treat them as partial, not complete.
- CLI and Gateway capabilities differ: capability-gate and expose honest unavailable states.
- Profile/Space scope leakage: use server-side resolved identity and concurrent isolation tests.
- Remote Mac confusion: require explicit transport identity/cwd canary; never infer from path string.
- Visual drift from new primitives: compare computed geometry/styles, not subjective screenshots only.
- Streaming layout jumps: isolate scroll owners and update only affected transcript regions.
- VPS resource limits: no build/deploy on VPS unless explicitly requested; use Mac/CI for full frontend verification.
- Delegated work can timeout or truncate: prefer small slices and verify `git diff --stat`, file integrity, tests, and branch state after every delegation.

## Definition of done

Migration is complete only when:

- Chat, Kanban, Spaces, and Profile workflows are fully implemented, not placeholders.
- Visual/position/component/interaction/responsive parity evidence passes or each deviation has explicit approval.
- Server/API/state/security contracts pass automated and live tests.
- Profile and Space isolation passes concurrent/reload/switch scenarios.
- Remote Mac/Tailscale execution is proven remotely and never misrouted locally.
- Migration and rollback are idempotent and rehearsed on copied state.
- Performance, accessibility, security, browser, live Hermes, and hosted release gates pass.
- User explicitly approves cutover. Original WebUI remains untouched until then.
