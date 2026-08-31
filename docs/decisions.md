# Architecture decisions

## ADR-001 — Gateway-first Hermes integration

**Status:** accepted for MVP

The Go service connects to the Hermes Gateway/API Server instead of importing the Python agent runtime or spawning the Hermes CLI. This is the cleanest runtime boundary for a non-Python rewrite and matches the bridge already supported by the original WebUI.

The adapter uses `POST /v1/chat/completions` with streaming enabled. Approval decisions use the Gateway Runs API at `POST /v1/runs/{run_id}/approval`; run creation remains owned by Hermes and is not duplicated in the BFF.

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

## ADR-011 - Preserve multimodal transport in opt-in Runs mode

The current Runs adapter sends a text `input` to `POST /v1/runs`; it does not
encode attachment parts. When `HERMES_WEBUI_USE_RUNS_API=true`, turns carrying
attachments therefore continue through the established chat-completions
adapter. This prevents an opt-in feature flag from silently dropping user
files while leaving the text-only Runs path available for approval parity work.
## ADR-013 - Server-owned identity boundary

Local password setup and signed session cookies are owned by the Go BFF. The
browser receives only authentication state, never password hashes, cookie
signing material, Gateway keys, provider credentials, or raw profile config.
Remote binding is refused until authentication is configured. Trusted-header
identity is opt-in for an already authenticated reverse proxy, while OIDC and
WebAuthn remain explicit capability states until provider registration is
configured.

## ADR-014 - Embedded frontend with recoverable state migration

The release binary embeds the Vite output so production does not require a
separate frontend server. `make build` refreshes the embedded package after a
locked frontend build; a tiny checked-in fallback keeps direct Go compilation
valid. Legacy Hermes state is never migrated in place: the separate migrator
creates a timestamped private backup and can restore the newest backup.

## ADR-015 - Separate liveness and readiness

`/health` remains a cheap liveness probe. `/ready` reports only local BFF
initialization and deliberately does not call Hermes Gateway. Gateway health
is an operator diagnostic because the BFF can still serve sessions, auth, and
safe offline UI while Hermes is restarting.

## ADR-017 - Shared UI primitives and Chat-owned session navigation

Visible controls use the repository's copy-owned shadcn-style `Button`,
`Input`, `Select`, and `Textarea` primitives so focus, sizing, and disabled
states stay consistent across the workspace. The file chooser remains a
hidden native input because it is the browser upload transport. Recent
sessions belong to the Chat navigation context and are hidden while another
control-center menu is active; session actions and search behavior remain
unchanged. The rail lives inside the Chat content area, leaving the primary
navigation sidebar focused on product sections.

## ADR-018 - Copy-owned dialogs for destructive and naming actions

Session rename/delete and workspace create/rename/delete use the shared Dialog
primitive instead of browser-native prompt/confirm surfaces. This keeps focus,
copy, cancellation, and destructive intent visible and consistent. The API
actions remain unchanged, so the compatibility boundary and rollback behavior
are preserved.

## ADR-019 - Icon rail with Chat-owned session toggle

The primary navigation is a compact icon-only rail on desktop. Each icon keeps
an accessible label and a focusable tooltip, while the mobile navigation drawer
retains readable labels. Chat remains a navigation destination, but activating
Chat again while already there toggles the secondary Recent sessions rail.
This preserves the session-first workflow while giving the conversation canvas
the full available width when the session list is not needed.

## ADR-020 - Session actions use a per-row overflow menu

Session rows keep the title and metadata as the primary scan target. Rename,
pin, archive, and delete remain available but are grouped under a three-dot
menu per row. The shared menu closes on selection, outside click, or Escape;
destructive confirmation continues to use the shared Dialog.

## ADR-016 - Separate Hermes runtime state from Web Studio metadata

The Web Studio state directory remains the owner of sessions, attachments,
control collections, auth, and preferences. Hermes-owned Skills and memory
must be read from `HERMES_HOME`, otherwise a valid Hermes installation appears
empty whenever the UI state directory is separate. Discovery is read-only,
bounded, and path-checked; the browser never receives credentials.

## ADR-021 - Personal WebUI defines replacement MVP parity

The earlier allowance for a visually distinct dashboard is superseded for the
replacement MVP. The read-only personal WebUI source at
`3caeca14064cec36c9c7b4f83ffade9a92cf2aee` and audited production behavior
define the required information architecture, workflows, responsive behavior,
appearance system, and backend state effects. React and Go remain the target
stack, but lightweight implementation work must use lazy loading, bounded state,
and small dependencies rather than removing features. Existing M0-M6 records
remain implementation history and do not certify personal-WebUI parity.

The personal implementation remains side by side through migration, beta, and
rollback proof. Any intentional visual or behavioral deviation requires a
recorded decision and explicit approval before MVP certification.
## ADR-022: Use a parity manifest before broad feature implementation

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Freeze a sanitized route/state/feature manifest for the
  personal WebUI before implementing P005 onward. Each capability maps to a
  Studio owner, task, disposition, and evidence type.
- **Reason:** The current Studio already contains several shallow equivalents;
  without traceability, broad migration could falsely mark a menu complete
  while losing workflow behavior such as CLI sessions, delivery history,
  external channels, or profile-local state.
- **Consequence:** Manifest rows are planning evidence only. Completion still
  requires the acceptance tests and live/browser proof defined by `MVP.md`.
  Private production state remains excluded from fixtures.

## ADR-023: Freeze shell semantics before visual convergence

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Give the primary rail, titlebar, menus, dialogs, and responsive
  shell stable semantic selectors and keyboard contracts before changing their
  visual geometry.
- **Reason:** Pixel work against an unstable DOM makes parity regressions hard
  to localize and can hide dead controls behind polished styling.
- **Consequence:** P005 source checks verify the contract, while screenshot and
  computed-style measurements remain separate P012 evidence.

## ADR-024: Keep composer selectors server-backed

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Profile and model controls in the composer read `/api/profiles`
  and send the selected values through the existing chat start contract. The
  workspace control opens the existing workspace service, and usage is shown
  only when Gateway supplies bounded usage values.
- **Reason:** The personal WebUI exposes runtime controls in the composer, but
  hardcoded models, quotas, or paths would create a misleading UI and violate
  the server credential boundary.
- **Consequence:** Grouped provider discovery, auxiliary models, and richer
  context-ring geometry remain P023/P012 work.

## ADR-025: Keep workspace panel tabs capability-driven

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Ship Files against the existing workspace service and Todos
  against the control API. Render Artifacts as an explicit empty state until
  its server contract exists. Persist only panel layout state in browser
  storage.
- **Reason:** The personal WebUI has a multi-surface right panel, but a fake
  artifact list would violate the requirement that every visible data row be
  real.
- **Consequence:** Artifact persistence and stream projection remain P029
  work; file and todo state are not duplicated in local storage.

## ADR-026: Navigation customization cannot create dead destinations

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Persist only the order and visibility of existing Studio
  destinations. Chat remains locked visible; unavailable reference panels are
  not added as fake navigation targets.
- **Reason:** The personal WebUI supports navigation customization, but a
  visible item without a working destination is a broken promise.
- **Consequence:** Kanban, Insights, Logs, and other missing sections remain
  mapped in the parity manifest and will enter the rail only with their own
  backend/UI slice.

## ADR-027: Apply theme before React paint

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Read theme and skin layout preferences in a minimal inline boot
  script, then let the shared theme utility validate and apply the same values
  after React starts.
- **Reason:** The personal WebUI supports System/Dark/Light and skins; waiting
  for an asynchronous preferences request would flash the wrong surface.
- **Consequence:** Only non-secret layout preferences are read before the app
  loads. Skin-specific visual certification remains P012 work.

## ADR-028: Keep mobile navigation to working destinations

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Use five mobile tabs backed by existing Studio views and keep
  the remaining destinations in the drawer until their panels are implemented.
  Reserve safe-area space for the composer and honor reduced motion.
- **Reason:** The reference has a compact mobile navigation model, but adding
  unsupported destinations would create dead controls and obscure the active
  chat workflow.
- **Consequence:** Device-specific IME, PWA, and complete transition evidence
  remain P011/P048 work.

## ADR-029: Close the responsive drawer at the desktop breakpoint

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Close the mobile navigation drawer when the viewport enters
  the desktop breakpoint, and keep the desktop primary rail compact.
- **Reason:** A drawer opened on mobile must not remain fixed and expanded after
  a resize or device rotation into desktop layout.
- **Consequence:** Returning to mobile requires the normal explicit menu action;
  no navigation destination or workflow is removed.

## ADR-030: Keep select behavior shadcn-owned while preserving form contracts

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Render visible selects as an accessible shadcn-style trigger and
  listbox, while retaining a visually hidden native select for form serialization
  and the existing controlled `onChange` contract.
- **Reason:** The interface must avoid browser-native visible controls without
  breaking existing form fields or keyboard expectations at current call sites.
- **Consequence:** Option children remain the source of truth; future grouped
  options and portal positioning need their own acceptance coverage.

## ADR-031: Persist one recoverable inflight turn journal

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Persist only the active stream and session identifiers in local
  storage, restore the server transcript on reload, and reconnect with the SSE
  cursor when available.
- **Reason:** Hard reload and session switching must not silently discard an
  active turn, while transcript contents remain server-owned and credentials
  never enter browser storage.
- **Consequence:** Concurrent inflight turns and expired-stream full-session
  polling remain explicit follow-up work.

## ADR-032: Keep parity evidence artifacts outside the repository

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** The MVP viewport matrix writes screenshots and computed geometry
  to an operator-selected local evidence directory, defaulting to `/tmp`, and
  commits only the deterministic runner and sanitized assertions.
- **Reason:** Screenshots contain environment-dependent rendering and must not
  capture personal session content or become mistaken for a frozen reference.
- **Consequence:** Reference pixel-diff approval and masked visual baselines
  remain explicit certification work.

## ADR-033: Use capability-safe local platform slices before hosted certification

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Ship locally verifiable PWA, subpath, security-header, and
  sanitized-fixture foundations while leaving hosted OS, independent security,
  live Hermes, beta, and cutover claims open.
- **Reason:** Local checks can prove wiring and boundaries but cannot substitute
  for external runtime or review evidence.
- **Consequence:** The related M10/M11 tasks remain `[~]` until their stated
  external gates are executed.

## ADR-034: Keep runtime parity helpers pure until upstream behavior is proven

- **Status:** Accepted
- **Date:** 2026-08-30
- **Decision:** Implement deterministic conversation-runtime planning,
  projection, repair, export, and lifecycle helpers as pure functions before
  wiring uncertain upstream/live semantics into the Gateway adapter.
- **Reason:** Helpers can be tested against sanitized contracts without
  inventing unsupported Hermes events or persisting browser-only state.
- **Consequence:** P017-P026 remain partial until their UI, server effects, and
  live side-by-side rows are proven.

## ADR-035: Expose honest operator capability states

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Expose Insights from server-owned facts and expose Terminal as
  explicitly unavailable until a contained process lifecycle is implemented.
- **Reason:** Operational dashboards must not invent cost or health metrics, and
  a terminal control without ownership and cleanup guarantees is unsafe.
- **Consequence:** Cost-backed Insights and terminal process operations remain
  explicit follow-up tasks with live and hostile-path acceptance.

## ADR-036: Rotate authentication state on every successful login

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Add per-login nonce entropy to signed session cookies and honor
  HTTPS reverse-proxy context only through the configured forwarding signal.
- **Reason:** Reusing a same-second deterministic cookie weakens session
  rotation, while incorrect Secure flags break or weaken deployments behind TLS
  termination.
- **Consequence:** Existing cookie verification remains compatible; passkey,
  OIDC, and independent threat review remain open.

## ADR-037: Ship read-only workspace and evidence slices before mutating parity

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Expose contained workspace browsing, safe preview/download/edit
  actions, and read-only Git status projections first. Keep reveal, mutating
  Git commands, and hosted certification explicitly unavailable until their
  ownership, confirmation, and rollback contracts are proven.
- **Reason:** Browser-visible paths and unbounded Git/process actions create
  security and recovery risk; deterministic local evidence must remain
  distinguishable from live Hermes parity.
- **Consequence:** P028/P031 and M11 certification tasks remain partial while
  paste/extract, mutating Git, screen-reader/device, hosted, and live gates are
  completed.

## ADR-038: Keep optional Settings diagnostics non-blocking

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Parse Settings responses through the shared non-JSON-safe API
  reader. Treat capability and operator diagnostics as optional status data, so
  an unavailable or stale route cannot blank the preference editor. Render the
  Capability status card only for the unfiltered view or matching searches.
- **Reason:** An embedded/older server can return an HTML fallback or a missing
  diagnostic route while the preferences route remains usable. Showing the
  capability card in every filtered Settings section also obscures the section
  the user selected.
- **Consequence:** Settings reports unavailable optional diagnostics honestly
  and remains usable during partial upgrades; preferences still fail visibly if
  their required server route is unavailable.

## ADR-039: Use one desktop rail scale and reserve its footer slot

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Use the same 24px icon scale and 56px control target for every
  primary rail item, including Settings. Keep the navigation list flexible and
  anchor Settings after it so the footer remains stable as items are reordered
  or hidden.
- **Reason:** The reference UI treats the rail as a persistent spatial anchor;
  inconsistent small icons and a moving Settings control reduce scanability.
  Chat/session controls also need a defined header grid so the New action never
  overlaps its subtitle.
- **Consequence:** The desktop rail is wider and more legible, the mobile
  drawer keeps the same control scale, and session microcopy uses readable
  small text while metadata remains intentionally compact.
