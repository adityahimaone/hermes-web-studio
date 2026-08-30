# Personal WebUI parity manifest

This manifest maps the frozen personal WebUI baseline to the Studio surface,
owner, acceptance evidence, and current disposition. It is intentionally based
on contracts and sanitized behavior, not on private production state.

## Baseline

- Source: `hermes-webui-personal@3caeca14064cec36c9c7b4f83ffade9a92cf2aee`
- Production behavior audit: 2026-08-30
- Studio baseline: M0-M6 implementation in this repository
- Required evidence: automated contract tests plus browser DOM/computed-style
  or live-Hermes evidence where the row says so

## Shell and navigation

| Reference surface/state | Studio owner | Current state | Task | Evidence required |
|---|---|---:|---|---|
| Compact primary rail: Chat, Tasks, Kanban, Skills, Memory, Spaces, Profiles, Todos, Insights, Logs, Settings | `frontend/src/components/sidebar.tsx`, `frontend/src/app.tsx` | partial | P005/P009 | DOM order, tooltip/focus, responsive screenshots |
| Chat rail toggle and secondary session panel | `frontend/src/components/session-rail.tsx`, `frontend/src/app.tsx` | partial | P005/P006 | open/collapse keyboard and narrow viewport |
| Chat titlebar: profile, connection, model/runtime controls | `frontend/src/components/top-bar.tsx` | partial | P007 | computed geometry and interaction matrix |
| Files/Artifacts/Todos workspace panel | `frontend/src/components/workspace-panel.tsx` | partial | P008 | independent panel state during stream |
| Control Center sections and customizable navigation | `frontend/src/components/control-center.tsx` | partial | P009 | section and persistence contract |
| Dark/Light/System and built-in skins | `frontend/src/styles.css`, preferences API | partial | P010 | boot-without-flash and skin matrix |
| Mobile drawer, bottom navigation, safe-area composer | `frontend/src/app.tsx`, responsive styles | partial | P011 | 44px targets, IME, screenshots |

## Conversation and sessions

| Reference capability | Studio owner | Current state | Task | Evidence required |
|---|---|---:|---|---|
| WebUI/CLI session sources, title/content search, date groups, project filters | session API and `SessionRail` | partial | P006/P020 | sanitized fixture list, transcript index, and browser workflow |
| Create/load/resume, rename, pin, archive, delete, tags, projects | `backend/internal/httpapi`, session store | partial | P006/P021 | CRUD, persistence, permission/error rows |
| Overflow actions: duplicate, fork, import/export, share, hide | session API and chat actions | partial | P019/P025 | state effects and safe downloads |
| Streaming message, reasoning, tools, subagents, approvals, usage | `backend/internal/gateway`, chat reducer | partial | P013/P014 | normalized fixtures and live side-by-side |
| Reconnect, replay, cancel, switch-session recovery | stream service and frontend chat state | partial | P015/P026 | interruption/reconnect lifecycle |
| Queue, interrupt, steer, clarify, compact, recovery | chat runtime | gap | P016/P018 | deterministic lifecycle rows |
| Markdown, code, Mermaid, KaTeX, tables, media, safe links | message renderer | partial | P022 | sanitized rendering and XSS tests |
| Profile/workspace/model/toolset composer controls | titlebar/composer | partial | P007/P023 | keyboard, responsive overflow, live model data |

## Workbench and operator surfaces

| Reference capability | Studio owner | Current state | Task | Evidence required |
|---|---|---:|---|---|
| Registered Spaces, inheritance, ordering, health | workspace service and Spaces UI | partial | P027 | multi-space fixture and isolation test |
| Tree, previews, edit/create/move/rename/delete/upload/extract | workspace API/UI | partial | P028 | containment, MIME, size, hostile-path tests |
| Artifacts, Todos, terminal, git operations | workspace/operator services | partial | P029-P031 | stream-safe panel and cleanup matrix |
| Cron/tasks schedule, delivery, skills, run/history/alerts | cron API/UI | partial | P032 | scheduler and delivery integration |
| Kanban boards, tasks, dispatch, events, stats | Kanban API/UI | gap | P033 | persisted state and event matrix |
| Skills grouping, content, linked files, CRUD, toggle, usage | skills API/UI | partial | P034 | profile-aware refresh and CRUD |
| Memory, USER, external notes, search, inline edits | memory API/UI | partial | P035 | source/path safety and timestamps |
| Insights, Logs, health, crash/background status, rollback | operator API/UI | gap | P037/P038 | safe diagnostics and failure states |
| Telegram/Discord/Slack/WeChat/Signal/SMS sessions and handoff | Gateway projections/UI | gap | P039 | live channel fixture and live Hermes proof |

## Identity, settings, extensions, and distribution

| Reference capability | Studio owner | Current state | Task | Evidence required |
|---|---|---:|---|---|
| Profiles, local state, workspace isolation, OAuth linking | profile/auth services | partial | P041/P044 | concurrent profile isolation |
| Provider CRUD, custom endpoints, model discovery, redaction | profile/provider services | partial | P023/P042 | no-secret bundle/log scan |
| Password, passkeys, OIDC, trusted headers, CSRF/CORS | auth middleware | partial | P043 | threat and browser acceptance matrix |
| Conversation/Appearance/Preferences/Providers/Plugins/Extensions/System/Help | Control Center/settings API | partial | P045/P046 | searchable settings and persistence |
| 15 locales, fallback, RTL, CJK/IME | frontend i18n | gap | P047 | locale and input matrix |
| PWA, offline shell, subpath, update invalidation | distribution assets | partial | P048/P053 | hosted/browser release matrix |
| MCP and extension registry/settings/sidecar trust boundary | extension services | gap | P049/P050 | consent, validation, secret filtering |
| Voice/TTS and unsupported capability states | browser/server media services | partial | P051 | permission and fallback rows |
| Docker, installer, multi-arch, Nix, migration/rollback | root `Dockerfile`, `scripts/`, release docs | partial | P053/P062-P066 | release artifacts and copied-state rehearsal |

## Disposition rules

1. `partial` means an equivalent Studio surface exists but does not yet prove
   the full personal workflow. It must not be marked complete by a list/CRUD
   mock alone.
2. `gap` means the surface or its state contract has not been implemented.
3. Native desktop/mobile clients are outside the web replacement boundary;
   their web-visible session and API behavior remains in scope.
4. Production credentials, private session names, model inventories, and
   workspace paths are never copied into fixtures.
5. A row may move to complete only when its task acceptance evidence is linked
   from `TASKS.md` and the relevant compatibility note is updated.

## Execution order

P004-P012 establish the shell and evidence harness. P013-P026 then close the
conversation runtime before P027-P040 expand the workbench. P041-P054 finish
identity/platform surfaces. P055-P066 are certification and reversible
cutover gates, not implementation shortcuts.
