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

Session overflow actions remain visible by default, while preserving existing hover/focus-within styling, so keyboard and touch users can reach them without hover.

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

## ADR-040: Keep rail controls compact but legible

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Cap the desktop primary rail at 72px with 48px controls and
  20px icons. Keep the brand mark slightly larger than navigation controls, but
  avoid using the larger mobile-drawer scale on the desktop rail.
- **Reason:** The first visual pass made the rail consume too much horizontal
  space and gave outline icons more visual weight than the chat content. The
  reference relies on a compact rail with clear hit areas, not oversized
  decoration.
- **Consequence:** Navigation remains keyboard and touch friendly while the
  chat canvas and session list recover the intended visual balance. Small
  session filters and titlebar metadata stay compact; action labels remain
  readable through the shared control sizes.

## ADR-041: Share the collapsible context rail across utility views

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Reuse one `ContextRail` shell for Chat sessions, Skills, and
  Settings categories. Each view owns its data and selection behavior, while
  the shell owns width, header action placement, collapse affordance, density,
  and the desktop-only expand control.
- **Reason:** Skills and Settings were consuming content width with bespoke
  list columns and larger rows. The Chat session rail already established the
  intended utility-panel rhythm and interaction pattern.
- **Consequence:** Utility panes now gain the same predictable collapse/expand
  behavior and row scale. Mobile keeps the existing bottom navigation and
  inline list fallback because the desktop rail is intentionally hidden below
  the large breakpoint.

## ADR-042: Track literal design convergence separately from feature parity

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Treat the supplied Hermes Studio design-system spec as a
  separate M12 visual-convergence track. Existing shared-rail and typography
  work remains valid implementation history, but does not certify the literal
  88px rail, 64px navbar, 552px sidebar, seam-control ownership, Profiles row
  anatomy, or screenshot/computed-style thresholds.
- **Reason:** Feature parity and visual parity have different evidence gates.
  Collapsing a sidebar and matching a few screenshots is not proof that every
  page uses the same shell or that dimensions, tokens, and states match.
- **Consequence:** P067-P072 must be completed and evidenced independently;
  MVP exit remains blocked by the open visual and Profiles rows even though the
  current utility rail implementation is reusable.

## ADR-043: Integrate M7-M12 work as bounded parity slices

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Integrate independently testable M7-M12 slices when their local
  contract is proven, while keeping the parent milestone partial until its
  live, browser, hosted, or reference-comparison gates pass.
- **Reason:** Parallel implementation can improve coverage quickly, but a
  slash-command registry, diagnostics endpoint, security hardening, or shell
  geometry probe does not prove the complete personal WebUI workflow.
- **Consequence:** The task log records each slice and its evidence separately;
  open rows, Profiles parity, full visual matrices, live Hermes checks, hosted
  artifacts, beta, and cutover approval remain explicit blockers.

## ADR-044: Converge Profiles through the shared ContextRail

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Render Profiles inside the existing `ContextRail` and keep
  profile selection/activation state in the Profiles view. Use one compact
  status-row presentation plus a feature slot before the list; provide a
  mobile inline fallback because the desktop rail is intentionally hidden on
  small screens.
- **Reason:** Profiles was the remaining control-center view with a bespoke
  card grid. Reusing the rail preserves the established collapse behavior and
  makes profile metadata scannable without changing the server contract.
- **Consequence:** Profile API behavior and activation remain compatible. The
  slice improves shell parity without claiming the full M12 visual matrix or
  exact reference dimensions are complete.
## ADR-045: Expose extension and update status before enabling control

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Add server-owned read-only plugin/extension registry and
  version/update diagnostics. Return explicit unavailable states for install,
  execution, trust, update apply, shutdown, restart, and lock recovery.
- **Reason:** The settings and operations surfaces need truthful status data,
  but browser-triggered extension execution and process lifecycle control
  would be unsafe without sandbox, consent, signed-update, and supervisor
  contracts.
- **Consequence:** P050/P052 gain testable platform contracts while sandboxed
  extensions, MCP integration, signed updates, and lifecycle actions remain
  open acceptance work.

## ADR-046: Keep MCP and plugin management metadata-only until sandbox proof

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Store validated MCP server/tool metadata and plugin settings in
  the server-owned control store. Do not launch MCP processes or execute plugin
  code from the BFF. Filter secret-shaped keys before persistence and response.
- **Reason:** Configuration management is useful for parity and can be tested
  without granting the browser or an untrusted integration process execution,
  network, or credential authority.
- **Consequence:** P049 has a proven local API/UI status slice, while live tool
  discovery, plugin execution, consent, and sandbox trust remain explicit gates.

## ADR-047: Use the responsive design contract over the superseded literal grid

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Implement the design contract with a 64px desktop icon rail,
  52–64px titlebar, and responsive 280–320px ContextRail rather than the
  superseded 88px/64px/552px values from the earlier audit draft. The shared
  rail owns its header, tools, body, and collapse seam; only the body owns
  vertical scrolling.
- **Reason:** The current reference interaction requires usable content width
  at laptop breakpoints and the existing app already establishes a compact
  rail. Applying 552px universally would regress the chat composer and utility
  panes while contradicting the responsive conflict-resolution section in
  `design.md`.
- **Consequence:** P067–P071 can converge incrementally on one shell without
  changing API or navigation behavior. Exact screenshot, computed-style, and
  hosted personal-WebUI comparison evidence remains a separate P072 gate.

## ADR-048: Give every utility view the same contextual sidebar shell

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Route Tasks, Spaces, Todos, Goals, Terminal, and Insights
  through the shared `ViewShell` used by the existing Skills, Memory, Profiles,
  and Settings surfaces. The shell owns the contextual rail, collapse seam,
  and content width; view components retain their existing data and actions.
- **Reason:** Utility views previously bypassed the sidebar system and appeared
  as unrelated narrow pages. Consistent shell ownership improves orientation and
  preserves the product's three-region information architecture without adding
  non-functional navigation controls or changing API behavior.
- **Consequence:** All utility views now share the same desktop rail and
  responsive content treatment. The non-chat workspace is height-bounded so
  the contextual rail reaches the bottom of the titlebar area while the main
  pane owns overflow. Mobile continues to use the existing drawer and bottom
  navigation contract; visual parity against the personal WebUI remains an open
  P072 evidence gate.

## ADR-049: Keep authentication entry in Settings, not the titlebar

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** The global titlebar exposes profile context only; password setup,
  sign-in, and sign-out controls render in the Settings account section. The
  existing `/api/onboarding/password`, `/api/auth/login`, and logout flows are
  unchanged.
- **Reason:** Password creation is configuration, not conversation runtime
  state. Removing its form from the titlebar reduces header competition and
  gives the security-sensitive action a clear labeled home.
- **Consequence:** Unconfigured users must open Settings to create a password.
  The browser still never receives provider credentials or Gateway keys.

## ADR-050: Use a CLI-first native Kanban transport with Dashboard capability upgrade

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Kanban reads and safe named actions use `hermes kanban` by
  default. When an authenticated Dashboard plugin is configured and healthy,
  Studio may upgrade automatically to its richer REST/WebSocket contract.
  Unsupported CLI operations are capability-gated in the UI.
- **Reason:** The installed Hermes CLI is the safest available default but does
  not expose the Dashboard's generic task patch, plugin bulk, or live event
  surface. Direct SQLite/Python access would duplicate Hermes behavior and
  violate the BFF boundary.
- **Consequence:** CLI-only deployments receive a useful, honest Kanban slice;
  full editing, live events, and Dashboard-only orchestration require the
  authenticated Dashboard service. Board selection remains browser-local and
  does not mutate Hermes' shared CLI current-board pointer.

## ADR-051: Bound Skills navigation and preview surfaces

- **Status:** Accepted
- **Date:** 2026-08-31
- **Decision:** Group discovered skills by their first filesystem category
  directory, place root-level skills in `(General)`, and make each group
  expandable with a count. Keep the Skills search filter fixed while only the
  results list scrolls, expose creation from the sidebar as an icon action,
  and let the Skills content pane use the available width. SKILL.md previews
  wrap long tokens and remain bounded by a scrollable preview region.
- **Reason:** Large Hermes skill collections otherwise make filters and the
  add action disappear during navigation, while narrow content and unbroken
  Markdown lines can make the page overflow. These changes improve orientation
  and resilience without changing the discovery API or CRUD workflow.
- **Consequence:** Category grouping is client-side presentation derived from
  the server-provided path; server filtering, selection, preview, edit, and
  delete contracts remain unchanged. The rail is explicitly height-bounded so
  its list, rather than the page, owns scrolling; rows follow the compact Chat
  session rhythm, with the skill name above an optional meaningful description.
  Placeholder-only descriptions are hidden. Mobile keeps its inline grouped
  list and existing navigation.
