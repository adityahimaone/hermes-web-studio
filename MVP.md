# Hermes Web Studio parity MVP

> Planning status: baseline inventoried, implementation not certified
> Reference source: `/Users/adityahimawan/Development/hermes-webui-personal`
> Frozen source commit: `3caeca14064cec36c9c7b4f83ffade9a92cf2aee`
> Production reference: `https://hermes.adityahimaone.space/`
> Production asset observed: `exp-v0.52.264-13-gc7c1a8ff-dirty-d6a2c40f`
> Audit date: 2026-08-30

## 1. MVP definition

The MVP is not a smaller feature subset of Hermes WebUI. It is the point where
Hermes Web Studio can replace the personal Hermes WebUI for its built-in browser
workflows without a user-visible or data-contract regression.

The rewrite may use React, TypeScript, Tailwind, copy-owned shadcn primitives,
and Go, but the lighter stack is an implementation constraint rather than a
reason to remove behavior. Until MVP certification, the personal WebUI remains
available side by side as the rollback implementation.

Parity has four independent dimensions:

1. **Visual parity:** information architecture, panel positions, dimensions,
   density, typography, tokens, empty/loading/error states, themes, skins, and
   responsive transformations match the reference.
2. **Interaction parity:** keyboard behavior, focus order, menus, dialogs,
   progressive disclosure, session switching, composer controls, mobile
   navigation, and recovery actions have the same discoverability and result.
3. **Frontend feature parity:** every built-in browser workflow has an owned
   implementation, including non-happy-path states and persistence.
4. **Backend contract parity:** route behavior, validation, auth, state effects,
   streaming/recovery semantics, and legacy data compatibility are proven even
   when the Go route name differs internally.

Build success, mocked API success, or a visually similar empty state does not
certify parity by itself.

## 2. Authority and conflict rules

- The frozen local source is the authority for backend contracts, edge cases,
  persistence, and tests.
- The production deployment is the authority for currently visible navigation,
  configured runtime behavior, density, and responsive presentation.
- Dynamic production content such as session names, profile names, model lists,
  workspaces, timestamps, and provider status is test data, not product copy.
- If production behavior and the frozen source disagree, record the difference
  in `TASKS.md`; do not silently choose one.
- Credentials and private production content must never be committed as a
  fixture, screenshot, log, or documentation example.

## 3. Required browser information architecture

### Primary rail

The desktop rail order and ownership are:

1. Chat
2. Tasks
3. Kanban
4. Skills
5. Memory
6. Spaces
7. Agent profiles
8. Todos
9. Insights
10. Optional Hermes Dashboard link when configured
11. Logs
12. Settings

Chat and Settings are fixed. Configurable visibility/order for the other tabs
is part of parity. Mobile uses the same destinations through the reference
drawer/bottom-navigation behavior.

### Chat composition

- Compact primary icon rail.
- Collapsible Chat session panel with new-conversation action, title/content
  search, WebUI/CLI source tabs, project chips, archived sessions, grouped dates,
  external-channel badges, and per-session overflow actions.
- Conversation-first center surface with a contextual titlebar.
- Composer footer owns attachments, bookmark/context controls, profile,
  workspace, model/provider, reasoning, queue/interrupt/steer mode, context
  usage, Stop, and Send according to available width and runtime state.
- Demand-driven right workspace panel with Files, Artifacts, and optional Todos
  tabs, resize behavior, preview/edit controls, and a mobile slide-over.

### Settings composition

The Control Center retains these sections: Conversation, Appearance,
Preferences, Providers, Plugins, Extensions, System, and Help. A generic
theme/locale form is not equivalent to this surface.

## 4. Required feature parity matrix

| Domain | Personal WebUI contract required for MVP | Studio status at planning audit |
|---|---|---|
| Chat lifecycle | streaming, cancel, queue/interrupt/steer, inflight session switching, reconnect/poll fallback, partial recovery, clarification, approvals, compaction/recovery | Partial |
| Transcript | Markdown, syntax, Mermaid, KaTeX, media, timestamps, model/provider routing, compact worklog, transparent stream, final-only mode | Partial |
| Conversation controls | copy, edit/regenerate, clear with watermark, branch/fork from turn, retry, undo/recovery | Partial |
| Sessions | WebUI + CLI sources, CRUD, content search, projects, tags, archive, pin, duplicate, import/export, batch actions, lineage, gateway sessions | Partial |
| Composer/runtime | searchable grouped models, custom providers, reasoning effort, toolsets, slash commands, attachments, voice/TTS, context/cost/TPS, mode controls | Partial |
| Workspace | multi-space registration, tree, broad preview matrix, edit/file ops, upload/extract, terminal, git operations/status, artifacts, worktrees | Partial |
| Tasks/cron | create/edit/delete, schedule builder, delivery, skills, run/pause/resume, history, alerts, live status | Placeholder-level |
| Kanban | boards, tasks, bulk actions, links, assignees, dispatch, event stream, stats | Missing |
| Skills | grouped search, content and linked files, create/edit/delete/toggle, usage | Read-only partial |
| Memory/notes | MEMORY/USER rendering, source search, timestamps, create/edit entries, external notes | Read-only partial |
| Profiles | create/switch/update/delete, profile-local state/workspaces, OAuth linking, runtime refresh and isolation | Switch-only partial |
| Todos/goals | native lifecycle, current-goal integration, workspace projection, completion state | Partial and structurally different |
| Insights | state synchronization, usage/cost/provider history, operational summaries | Missing |
| Logs/health | logs, crash visibility, agent/gateway/system health, restart/rollback/update diagnostics | Missing |
| Settings/providers | full appearance/preferences, provider CRUD/discovery/quota, model aliases/auxiliary routing, personality, notification controls | Minimal partial |
| Auth/onboarding | password, passkeys, OIDC, OAuth, trusted proxy behavior, branded provider setup and diagnostics | Partial |
| Gateway channels | Telegram/Discord/Slack/WeChat/Signal/SMS sessions, handoff dock, summary, routing metadata, approval runs | Missing/partial |
| MCP/plugins/extensions | MCP CRUD/tools, plugin settings, extension gallery/install/toggle/uninstall, sidecar consent, custom themes/TTS/nav | Missing |
| PWA/i18n/accessibility | service worker, installability, 15 locales, RTL/CJK/IME, keyboard/focus/mobile/touch contracts | Minimal partial |
| Distribution/operations | embedded binary, subpath, Docker topologies, migration/rollback, health, updates, multi-platform artifacts | Partial |

“Partial” means a route or UI exists but the reference workflow and its
acceptance matrix are not yet proven. Existing M0-M6 completion records remain
implementation history; they are not retroactive MVP parity certification.

## 5. Appearance contract

- Base modes: System, Dark, Light.
- Built-in skins from the frozen personal source: default, ares, catppuccin,
  charizard, codex, geist-contrast, github, graphite, hepburn, mono, neon,
  neon-paint, neon-soft, nous, poseidon, sienna, sisyphus, slate, terracotta,
  verdigris, and zeus.
- Font sizes, no-flash boot, server/local persistence, system-mode live updates,
  and extension-registered skins are required behavior.
- The conversation remains the visual priority; activity/tool details are
  progressively disclosed and action-required approval/error states stay
  prominent.

### 5.1 Reference UI convergence backlog

The literal Hermes Studio design-system audit is now tracked separately in
M12/P067-P072. The current implementation has a shared collapsible rail for
Chat, Skills, Memory, and Settings, compact utility typography, and a slimmer
non-chat titlebar. It is not yet visual parity certification: Profiles still
needs the shared shell and profile-specific rows, the page-layout-owned seam
control is not unified across every page, and the audited 88px/64px/552px
geometry and exact token map require screenshot and computed-style evidence.
The reference sidebar remains a 552px target even though the current
transitional implementation uses a narrower compact width while convergence is
in progress.

## 6. Locale and responsive contract

Required locales: English, Italian, Japanese, Russian, Spanish, German,
Simplified Chinese, Traditional Chinese, Portuguese, Korean, French, Czech,
Turkish, Polish, and Vietnamese.

Every visual slice must be checked at minimum at 1440x1000, 1280x800, 1024x768,
768x1024, and 390x844. Required states include both side panels open/closed,
long session lists, long transcripts, active streaming, menus/dialogs, and
mobile keyboard-safe composer behavior.

## 7. Evidence and acceptance thresholds

Each parity task needs all applicable evidence:

- contract fixture derived from the frozen personal source;
- Go unit/integration coverage for route and state effects;
- frontend unit/component coverage for state transitions;
- Playwright behavior coverage against deterministic fixtures;
- live Hermes/Gateway proof for runtime-dependent behavior;
- screenshot plus DOM/computed-style evidence for visual work;
- empty, loading, success, failure, cancellation, reconnect, and recovery states;
- secret-boundary and migration/rollback checks where applicable.

For stable UI regions, key geometry must match within 1 CSS pixel and computed
typography/color/radius values must match the approved baseline. Screenshot
comparison should remain within 0.5% changed pixels after masking dynamic
session text, timestamps, cursors, and live status. Any intentional deviation
requires an ADR and explicit approval.

## 8. Lightweight constraints

- One Go binary embeds production frontend assets.
- No Python runtime is required after cutover.
- Gateway/provider credentials remain server-side.
- Frontend initial JavaScript stays below 250 KiB gzip through route-level lazy
  loading; large renderers and operator panels must not inflate Chat startup.
- Go BFF idle RSS remains below 80 MiB excluding Hermes Gateway.
- Streaming does not refetch the complete transcript per token.
- First normalized-token overhead remains below 50 ms beyond local upstream.
- Large-session tests cover 1,000 sessions and a 5,000-message transcript.

## 9. MVP exit gate

MVP parity is complete only when:

1. every matrix row has an owner, contract, implementation, and evidence link;
2. no required built-in personal WebUI workflow remains missing or placeholder;
3. desktop, laptop, tablet, and mobile visual matrices pass in all base modes
   and built-in skins selected for release certification;
4. session, profile, workspace, cron, auth, and preference migration/rollback
   pass against copied legacy state;
5. a side-by-side beta completes with no unresolved critical or major parity gap;
6. performance and secret budgets pass from the release artifact;
7. the user explicitly approves cutover;
8. the personal WebUI remains recoverable until the rollback window closes.
