# Hermes Web Studio UI/UX audit and design contract

Status: Proposed design authority for the next UI convergence pass  
Target: `adityahimaone/hermes-web-studio`  
Reference: supplied Chat and Skills screenshots plus `adityahimaone/hermes-webui-personal` (`master`)  
Audit date: 2026-08-31

## 1. Outcome

Hermes Web Studio should feel like a dense, calm desktop workbench: navigation is compact, lists use the available height, the central canvas receives most of the width, and controls appear where they are needed without competing with the user's work.

This document is a visual and interaction audit, not permission to remove or simulate a feature. Existing Hermes flows, server contracts, keyboard behavior, data compatibility, loading/error/empty states, and the browser credential boundary remain governed by `AGENTS.md`, `MVP.md`, and `docs/compatibility-contract.md`.

### Success criteria

- No clipped search input, label, placeholder, select value, or title at supported viewports and 200% zoom.
- Chat, Skills, Memory, Profiles, and Settings use one predictable shell and one primary scrolling list per context rail.
- Search fields fill the available sidebar width.
- Long Chat and Skills lists stay within the viewport and scroll independently from the content pane.
- Dense desktop controls use 32–40px visual heights; 44px targets are retained for touch/coarse pointers.
- Session and skill rows use 12–13px labels rather than 14–16px display text.
- The main content width grows beyond the current 768px limit on large screens.
- Filters no longer consume three oversized rows in Chat.
- Each screen exposes one primary create action, not duplicate actions in the rail and content header.
- Empty space is intentional and bounded; it must not be caused by narrow arbitrary max-widths or oversized controls.

## 2. Evidence and current-state audit

### 2.1 Screenshot findings

| Area | Current observation | Impact | Priority |
| --- | --- | --- | --- |
| Primary rail | Icons and active tiles carry more visual weight than the work content. | Navigation dominates the page and reduces focus. | High |
| Context rail search | Search is visibly narrower than the list and its placeholder is clipped. | Search looks broken and wastes horizontal space. | Critical |
| Chat filters | Source filters, project filters, and Select mode occupy separate tall rows. | Too much sidebar height is spent before the first conversation. | High |
| Session rows | Labels are visually large and rows are tall relative to their information. | Fewer sessions are visible; scanning is slower. | High |
| Skills list | Large one-line rows use substantial vertical space. | Long skill collections require excessive scrolling. | High |
| Skills actions | “New” appears in the context rail and “New skill” appears again in the content header. | Duplicate hierarchy and unnecessary visual noise. | High |
| Main canvas | Empty state, transcript, composer, and utility pages are constrained to a narrow central column. | Wide displays show large unusable gutters. | Critical |
| Composer | A tall input plus many always-visible controls creates a heavy block. | The composer competes with the conversation. | High |
| Header | Runtime, profile, password, status, reset, and panel actions compete in one row. | Weak prioritization and likely clipping at zoom/narrow widths. | High |
| Skills preview | Raw content is placed in a bounded card with another internal scroll. | Content feels boxed-in and hard to read or edit. | Medium |

### 2.2 Source-level causes in Hermes Web Studio

- `frontend/src/components/ui/input.tsx` has no `w-full`; a standalone `Input` therefore uses intrinsic width. This directly explains the narrow search field.
- `frontend/src/components/layout/context-rail.tsx` makes the whole child region scrollable, while `session-rail.tsx` also creates an inner `overflow-y-auto` list. This creates competing/nested scroll ownership.
- `session-rail.tsx` uses `min-h-11` for each source filter, project chip, Select control, batch action, and session target. A 44px minimum is correct for touch but unnecessarily large for dense desktop pointer use.
- `message-list.tsx`, `composer.tsx`, Profiles, Tasks, Spaces, and several control views use `max-w-3xl`, capping useful content at 768px.
- `control-center.tsx` renders the Skills/Memory create action twice.
- `control-center.tsx` renders the Skills preview as a raw `<pre>` capped at `32rem`, even when the content pane has ample space.
- The titlebar uses a raw lower-case `view` string for non-chat pages, which produces inconsistent page naming.
- Solid primary styling is used for filters and active choices. This makes secondary state controls compete with the actual primary action.

### 2.3 Useful patterns from Hermes WebUI Personal

The reference should guide composition and density, not be copied line-for-line.

- The reference sidebar is 300px and its search input is explicitly `width: 100%`.
- Session rows use 13px text, 8px padding, 2px inter-row spacing, and 11px metadata.
- Skill rows use 12px text with 8px vertical padding.
- The session and skill lists own `flex: 1`, `min-height: 0`, and `overflow-y: auto`.
- The titlebar is visually quiet, and the sidebar header owns local actions.
- The reference composer uses a responsive width rather than a fixed 768px ceiling.
- Create actions appear once in the relevant panel header.

## 3. Information hierarchy

The desktop shell has four possible layers. Only the first three are present by default.

1. Primary rail: global destination switching.
2. Context rail: search, filtering, and selection for the active destination.
3. Main canvas: reading, editing, and primary work.
4. Workspace panel: demand-driven files, artifacts, and todos.

The main canvas must receive all remaining width. A rail may collapse, but collapsing it must never discard selection, query, draft, or scroll state.

### Ownership rules

- Global destination actions belong in the primary rail.
- Collection actions such as New chat or New skill belong once in the context-rail header.
- Item actions such as rename, archive, duplicate, edit, and delete belong in row overflow menus or the detail header.
- Runtime status belongs in the titlebar, but configuration belongs in a menu/panel.
- Conversation controls belong in the composer, progressively disclosed by available width.
- Workspace actions belong in the workspace panel, not duplicated in the titlebar and composer.

## 4. Design tokens

### 4.1 Typography

Use Inter Variable with system fallbacks. Use a monospace stack only for code, paths, IDs, logs, and raw skill source.

| Token | Size / line-height | Weight | Use |
| --- | --- | --- | --- |
| `display` | 28/34px desktop, 24/30px small | 650 | Empty-state prompt only |
| `page-title` | 22/28px | 600 | Utility page title |
| `section-title` | 16/22px | 600 | Cards and detail sections |
| `body` | 14/21px | 400 | Message and explanatory copy |
| `control` | 13/18px | 500 | Buttons, selects, list titles |
| `compact` | 12/17px | 450 | Skills and dense utility rows |
| `meta` | 11/15px | 450 | Dates, projects, providers, status |
| `eyebrow` | 10/14px | 600 | Uppercase section labels; tracking 0.12em |

Rules:

- Do not use 16px+ text in context-rail rows.
- Do not use `text-[10px]` for essential controls or values; reserve it for nonessential labels and badges.
- Truncated text must expose the full value through `title`, tooltip, or an accessible detail surface.
- Body text must remain at least 14px and use a maximum reading line of about 80 characters.

### 4.2 Spacing

Use a 4px base scale: 4, 8, 12, 16, 24, 32, and 48px. Avoid untracked one-off margins unless needed for optical alignment.

- Dense row gap: 4px.
- Control group gap: 8px.
- Context-rail padding: 12px.
- Main content padding: 24px desktop, 16px tablet, 12px mobile.
- Section gap: 24px.
- Large empty-state gap: 16px, not 24–32px between every element.

### 4.3 Radius and borders

- Inputs/buttons: 8px.
- List rows: 6–8px.
- Cards: 10–12px.
- Composer: 14px.
- Modal: 14px.
- Use one 1px semantic border. Avoid stacking a bordered rail, bordered card, and bordered pre unless each level conveys distinct hierarchy.

### 4.4 Color and emphasis

- Solid purple is reserved for the main action and final Send action.
- Selected filters and list rows use a subtle accent background plus a small indicator/check, not a full solid-purple tile.
- Hover, selected, focus, streaming, warning, and destructive states must be distinguishable without color alone.
- Muted text must meet WCAG AA contrast against its actual surface.
- Focus rings remain visible at 2px with a 2px offset when they would otherwise blend into a border.

## 5. Responsive shell specification

| Region | Desktop ≥1280px | Compact 1024–1279px | Tablet 768–1023px | Mobile <768px |
| --- | --- | --- | --- | --- |
| Titlebar | 52px | 52px | 56px | 56px + safe area |
| Primary rail | 64px | 56px | Drawer | Bottom nav + drawer |
| Context rail | 300px default; resize 264–360px | 280px | Overlay/drawer | Full-width sheet |
| Workspace panel | 360px default; resize 300–520px | 320px | Overlay sheet | Full-width sheet |
| Main padding | 24–32px | 20–24px | 16px | 12px |

At 200% browser zoom, layout behavior is based on the reduced CSS viewport: context and workspace rails switch to overlay/drawer behavior before content clips.

### Primary rail

- Desktop width: 64px; compact width: 56px.
- Brand mark: 40px; navigation buttons: 36px visual size; icons: 18–20px.
- Use a 4px vertical gap and an independently scrollable navigation group when height is insufficient.
- Keep Settings anchored at the bottom.
- Use tooltips with a short delay and always provide `aria-label`.
- On coarse pointers, button hit areas become at least 44×44px without increasing desktop visual density.

### Titlebar

- One line only.
- Left: page/conversation title and optional short metadata.
- Right: connection status, compact profile/account menu, and one workspace toggle.
- Move password/account actions into the profile menu.
- Put Reset/Clear under a conversation overflow menu and require confirmation if transcript data will be removed.
- Hide low-priority labels before controls clip. Status may collapse to an icon with tooltip below 1180px.

### Context rail structure

Refactor `ContextRail` into explicit zones:

1. Header: 48px, title, optional count, one create action, collapse action.
2. Tools: non-scrolling search/filter region.
3. Body: the only vertically scrolling list region.
4. Optional footer: selection/batch summary.

Required layout behavior:

```css
.context-rail {
  height: calc(100dvh - 52px);
  max-height: calc(100dvh - 52px);
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.context-rail__body {
  flex: 1 1 auto;
  min-height: 0;
  max-height: 100%;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
}
```

Do not make both `ContextRail` children and an inner list scroll. Search and filters remain visible while the user scrolls Chat or Skills.

### Main canvas widths

- Chat transcript: `width: min(100%, 960px)`.
- Composer: same horizontal alignment as transcript; `width: min(100%, 1040px)`.
- Empty-state suggestions: `width: min(100%, 760px)`.
- Utility/detail pages: `width: min(100%, 1120px)`.
- Data-heavy pages (Kanban, logs, insights): use the full content pane up to 1280px.
- When the workspace panel opens, widths shrink fluidly; never create horizontal page scroll.

These values replace the blanket `max-w-3xl` pattern.

## 6. Component specification

### Inputs and search

- Add `w-full` to the shared `Input` primitive or require an explicit width variant. All context-rail searches must resolve to `width: 100%`.
- Desktop search height: 36px; mobile/coarse pointer: 44px.
- Search text: 13px; icon: 14–16px; horizontal padding: 10–12px.
- Use `type="search"`, a visible clear action when nonempty, and Escape to clear once before focus leaves.
- Debounce remote content search by 250–300ms; local list filtering is immediate.
- Announce result counts with a polite live region without announcing every keystroke.

### Buttons

| Variant | Desktop visual height | Use |
| --- | --- | --- |
| Icon compact | 32–36px | Rail/header/row actions |
| Small | 32px | Filters and inline actions |
| Default | 40px | Forms and main page actions |
| Touch/coarse pointer | min 44px target | All interactive controls |

Avoid applying `h-11` to every desktop button. Destructive actions remain visually neutral until confirmation, except inside the confirmation dialog.

### Select and combobox

- Closed trigger height: 36px desktop, 44px touch.
- Use 13px labels and values.
- Lists with more than eight options become searchable comboboxes.
- Dropdown max height: 288px with internal scroll.
- Match trigger width to its context; in toolbars use a bounded width, in forms use full width.
- Show the full selected value through tooltip when the visible value is truncated.

### Segmented controls and filters

- Source (`All`, `WebUI`, `CLI`) is one 32px segmented control, not three large primary buttons.
- Project, tags, channel, archived state, and status live in one `Filter` popover or compact row.
- The Filter button shows an active count, for example `Filter · 2`.
- Include Clear all and preserve filters while switching sessions within Chat.
- Use a single-line horizontally scrollable chip row only for genuinely frequent project shortcuts; it must not be the only way to access projects.

### List rows

- Default row height: 40px; rows with metadata: 44–48px.
- Title: 13px/18px; metadata: 11px/15px; icon: 14–16px.
- Row horizontal padding: 8px; vertical padding: 6px; gap: 6–8px.
- Use one-line truncation. Do not allow a long skill/session name to increase row height.
- Overflow actions appear on hover/focus but remain reachable from keyboard and touch.
- Active state uses a subtle fill and 2px accent marker or check.
- Virtualize or window the list beyond 200 items; preserve keyboard navigation and active-row visibility.

### Cards

- Use cards for separate conceptual objects, not as a default wrapper around every region.
- A detail view may use a borderless reading surface with section dividers.
- Card padding: 16px compact, 20px default.
- Avoid nested cards more than two levels deep.

### Dialogs and menus

- Restore focus to the trigger on close.
- Escape closes the topmost non-blocking surface.
- Destructive dialogs name the affected item and explain whether recovery is possible.
- Menus close on selection, outside click, and Escape.
- Keep menu labels as verb + object: Rename conversation, Archive conversation, Delete skill.

## 7. Chat page audit and target flow

### 7.1 Context rail

Recommended order:

1. Header: Chat, optional result/session count, New conversation icon, collapse.
2. Full-width search.
3. One compact toolbar: segmented source + Filter button + Select mode.
4. Optional active-filter chips on one line.
5. Date-grouped session list owning the remaining height.
6. Batch footer only while Select mode is active.

Session rows:

- First line: title, pin/stream status, overflow action.
- Second line only when useful: project/tag/channel and relative time.
- Do not display an empty metadata line.
- Search results may show one 11px matching-content excerpt; normal browsing should not.
- Pinned group appears before Today, followed by Yesterday and older ranges.
- Archived sessions are hidden by default and exposed through Filter.

### 7.2 Chat empty state

- Center within available conversation space, excluding composer height.
- Icon: 48px container / 22–24px glyph.
- Heading: 28px desktop, 24px small.
- Supporting copy: one or two lines, maximum 560px.
- Suggestions: maximum four; 40–44px rows, 13px text.
- Hide marketing-style copy once a session has content.

### 7.3 Transcript

- Increase max width from 768px to 960px.
- Use 14px/22px message body; code 13px/20px.
- Assistant content is mostly borderless; user messages use a restrained bubble capped at 80% width.
- Message group gap: 24px, not 32px.
- Keep tool, reasoning, approval, clarification, subagent, usage, and recovery states close to the assistant turn that owns them.
- Timestamps and routing metadata are progressive disclosure, not permanent high-contrast labels.
- Copy, retry, edit, branch, and overflow controls appear on hover/focus and remain keyboard accessible.

### 7.4 Composer

The composer should start compact and grow with content.

- Initial textarea: 44px; maximum text area: 160px before internal scroll.
- Container padding: 8px; radius: 14px.
- Top/primary row: textarea and Send/Stop.
- Bottom/secondary row: attachment, tools, turn mode, profile/model, workspace, usage.
- At constrained widths, move profile/model/workspace/reasoning/toolsets into one `Conversation settings` popover rather than wrapping into multiple rows.
- Keep Send at the lower-right edge and Stop in the same stable position during streaming.
- Show keyboard hint only when sufficient width exists.
- Queue items appear in a bounded flyout/dock above the composer, not below the page disclaimer.
- Approval and clarification surfaces dock above the composer so the requested decision remains visible.

### 7.5 Chat flow states

| State | Required behavior |
| --- | --- |
| New | Empty state visible; composer focused after explicit New action. |
| Searching | List updates without moving search focus; show result count and clear action. |
| Loading session | Keep row selected; show transcript skeleton; preserve composer draft. |
| Sending | Send becomes Stop; pending intent is visible; session list shows streaming status. |
| Queue / Interrupt / Steer | Mode and consequence are visible before submit; queued items can be removed. |
| Approval / clarification | Decision surface is focusable, announced, and adjacent to composer. |
| Error | Keep user input and attachments recoverable; show Retry and copyable diagnostic ID. |
| Reconnect | Preserve transcript and cursor state; show compact non-blocking banner. |
| Cancelled | Preserve partial output and offer Retry/Edit. |
| Session switch | Active stream and draft remain recoverable according to existing contract. |

## 8. Skills and Memory audit and target flow

### Context rail

- One create action in the context-rail header. Remove the duplicated New skill/New note button from the main header.
- Full-width 36px search.
- Skills list receives all remaining rail height and is the only scroll owner.
- Skill row: 40px, 12–13px title, optional 11px domain/source line only when present.
- Add grouping only when backed by real metadata, for example Enabled, User, Built-in, or domain. Do not infer or fake groups.
- If enable/disable exists, use a compact switch and preserve row selection; otherwise do not display a nonfunctional toggle.

### Detail pane

- Before selection: compact empty state with title, explanation, and optional create action only when the context header action is not visible (mobile).
- After selection: detail header with name, status/source metadata, and overflow actions.
- Default Preview tab renders safe Markdown/structured content.
- Raw tab shows monospace source with copy and line wrapping options.
- Linked files and usage appear as tabs/sections only when server-backed.
- Remove the arbitrary `32rem` reading cap on desktop. The main content pane owns page scrolling; only raw code blocks scroll internally when necessary.
- Edit opens an editor with dirty-state protection. Delete names the skill and explains consequences.
- After create, select the new item and move focus to its detail heading. After delete, select the nearest remaining item or return to the empty state.

## 9. Other feature surfaces

### Profiles

- Keep the shared context rail and 40–48px status rows.
- Show active/health/provider metadata without duplicating the same status in every region.
- Detail pane uses a 2-column metadata grid at desktop and one column on mobile.
- Switching profiles must clearly state whether it affects the current conversation or only new conversations.

### Settings

- Context rail: full-width search followed by compact 36–40px categories.
- Search filters sections and highlights matching labels/descriptions; clearing restores all sections.
- Main content max width: 960–1120px.
- Preference rows should use label/description on the left and control on the right at desktop; stack on mobile.
- Save state must be visible: saving, saved, failed, and unsaved changes.
- Capability/unavailable cards remain honest but visually secondary.

### Tasks, Todos, Goals, and Spaces

- Replace blanket centered 768px columns with up to 1120px.
- Creation forms stay one line only when labels and errors fit; otherwise stack.
- Keep filters/search sticky above long lists.
- Use table/list density appropriate to operator work: 40–48px rows and 13px labels.
- Empty, loading, error, permission, watch/stream, and recovery states must not shift the page shell.

### Kanban

- Use the full content pane.
- Columns scroll horizontally only within the board; the shell must not horizontally scroll.
- Column headers remain sticky inside the board.
- Drag-and-drop must have keyboard alternatives and clear live announcements.

### Insights and Logs

- Use responsive grids with a maximum 1280px canvas.
- Charts require text/table equivalents and truthful unavailable states.
- Logs use monospace 12–13px, sticky controls, severity filters, follow/pause, and bounded virtualization.

### Workspace panel

- Open on demand; default 360px, resizable 300–520px.
- Preserve active tab, width, and selected file without reducing chat below a usable width.
- At compact widths, overlay the main canvas instead of squeezing it.
- Files, Artifacts, and Todos own independent scroll positions.

## 10. Responsive and mobile behavior

- Below 1024px, the context rail becomes an overlay/drawer rather than a permanently squeezed column.
- Below 768px, use bottom navigation for the highest-frequency working destinations and a drawer for the full destination list.
- Mobile context rail becomes a full-width sheet with search and filters fixed above the list.
- The composer accounts for safe-area and virtual-keyboard insets.
- Do not autofocus search, model, or filter popovers on coarse pointers if doing so opens the keyboard unexpectedly.
- Long press/swipe actions may supplement but never replace visible or keyboard-accessible actions.
- Preserve draft, selection, open panel, and active filters across rotation and breakpoint changes.

## 11. Accessibility contract

- Meet WCAG 2.2 AA for contrast, keyboard access, focus visibility, labels, status announcements, and target sizing.
- Maintain logical landmarks: primary navigation, context navigation, main workspace, complementary workspace panel.
- Every icon-only control has an accessible name and tooltip.
- Arrow keys navigate listbox/combobox options; Home/End work; Escape closes the topmost surface.
- List selection and batch selection are separate states and separately announced.
- Loading uses `role="status"`; destructive and blocking errors use appropriate alert semantics without repeated announcements.
- Focus returns to the initiating control after dialog, menu, drawer, or sheet closure.
- Respect reduced motion. No essential state is conveyed only through animation.
- Test screen reader naming and ordering, 200% zoom, high contrast, RTL, CJK IME, and touch/coarse pointer layouts.

## 12. Performance and long-list behavior

- Session and skill search must not perform a full app rerender per keystroke.
- Debounce server content search and cancel/ignore stale responses.
- Window session lists above 200 rows and message history above the existing bounded threshold.
- Preserve scroll anchor during streamed updates, filters, rename, pin, archive, and live status changes.
- Avoid backdrop blur on every nested surface; one shell layer is sufficient.
- Lazy-load heavy renderers such as Mermaid, syntax, KaTeX, media preview, and data charts.
- Do not add a large component library for layout changes; extend copy-owned primitives.

## 13. Implementation map

| Area | Primary files | Required change |
| --- | --- | --- |
| Input width/density | `frontend/src/components/ui/input.tsx` | Add full-width behavior and density variants. |
| Button density | `frontend/src/components/ui/button.tsx` | Add compact/default/touch-aware sizes; stop using 44px for every desktop control. |
| Select density/overflow | `frontend/src/components/ui/select.tsx` | Compact trigger, bounded menu, searchable variant for long lists. |
| Shared context shell | `frontend/src/components/layout/context-rail.tsx` | Add header/tools/body/footer slots and one scroll owner. |
| Chat sidebar | `frontend/src/components/chat/session-rail.tsx` | Consolidate filters, compact rows, fixed tools, scrollable list, batch footer. |
| Transcript | `frontend/src/components/chat/message-list.tsx` | Wider content, tighter rhythm, interaction states. |
| Composer | `frontend/src/components/chat/composer.tsx` | Wider compact layout, progressive overflow, stable Send/Stop, dock queued work. |
| Header/shell | `frontend/src/app.tsx`, `frontend/src/components/layout/sidebar.tsx` | Compact rail/titlebar, prioritize actions, normalize titles, zoom behavior. |
| Skills/Memory/Settings/Profiles | `frontend/src/components/control/control-center.tsx` | Remove duplicate actions, compact lists, improve detail widths and flow. |
| Semantic tokens | `frontend/src/app.css` | Add density, width, surface, focus, and responsive tokens. |
| Browser evidence | `frontend/e2e` | Add geometry, scrolling, clipping, keyboard, zoom, and long-list coverage. |

Any implemented slice must also update `TASKS.md` and record a decision in `docs/decisions.md` as required by `AGENTS.md`.

## 14. Priority plan

### P0 — Fix broken sizing and scroll ownership

1. Make context search inputs full width.
2. Refactor ContextRail so search/filter controls are fixed and only its list scrolls.
3. Replace blanket `max-w-3xl` in Chat, composer, and utility pages with surface-specific widths.
4. Remove duplicate Skills/Memory create actions.

### P1 — Density and hierarchy

1. Add compact desktop Button/Input/Select variants.
2. Consolidate Chat filters into segmented source + Filter + Select mode.
3. Reduce session and skill rows to the specified typography and heights.
4. Simplify titlebar and composer action ownership.

### P2 — Flow completion

1. Improve Skills preview/raw/edit/delete flow.
2. Normalize Profiles and Settings layouts.
3. Apply the same density rules to Tasks, Spaces, Todos, Goals, Logs, Insights, and Kanban.
4. Add long-list virtualization where thresholds are exceeded.

### P3 — Certification

1. Capture visual matrices for expanded/collapsed rails, empty/detail/long-list states, streaming, dialogs, workspace open, themes, and skins.
2. Run keyboard, screen-reader, contrast, zoom, RTL/IME, reduced-motion, and coarse-pointer tests.
3. Compare side by side with the reference workflows without using private production state.

## 15. Acceptance matrix

Required viewports in CSS pixels:

- 1440×900 desktop.
- 1280×800 compact desktop.
- 1024×768 tablet landscape.
- 768×1024 tablet portrait.
- 390×844 mobile.
- Repeat desktop at 200% browser zoom.

Required states:

- Chat: empty, long session list, search results, active filters, batch select, populated transcript, streaming, queued, approval, clarification, error, reconnect, workspace open.
- Skills/Memory: empty, long list, search, selected preview, raw, create, dirty edit, delete confirmation, error.
- Profiles/Settings: long lists, active selection, search, unsaved/saving/saved/error, unavailable capability.
- All rails: expanded, collapsed, insufficient viewport height, keyboard focus, touch/coarse pointer.

Geometry and behavior assertions:

- Every context search input width is within 1px of its tools container width.
- No horizontal document scroll at any required viewport.
- Context tools remain stationary while Chat/Skills list scrollTop changes.
- Chat and Skills lists never extend below the viewport or behind mobile navigation/composer.
- At least 12 single-line session rows and 14 single-line skill rows are visible at 900px desktop height when metadata is absent.
- Main chat/composer width exceeds 768px when at least 1100px is available to the main canvas.
- Text and controls remain unclipped at 200% zoom.
- Focus order follows visual order and focus is restored after overlays close.
- Touch targets are at least 44px for coarse pointers; pointer desktop controls follow compact visual sizes.

## 16. Conflict resolution with the existing M12 geometry

`TASKS.md` currently records a literal M12 target of an 88px primary rail, 64px navbar, and 552px sidebar. That geometry conflicts with the supplied screenshots, the current 72px/280px implementation, the 300px reference sidebar, and the explicit goal of reducing unused space.

Before implementation, record a superseding ADR choosing this document's responsive 64px rail, 52px titlebar, and 264–360px context-rail range. Do not mark the old literal M12 thresholds complete and do not silently change their meaning. Feature parity remains independent from this design convergence decision.

## 17. Definition of done

The audit is implemented only when:

- P0 and P1 changes are complete across Chat and Skills.
- Shared primitives and ContextRail are reused by Memory, Profiles, and Settings.
- Existing flows and server-backed state remain functional.
- Loading, empty, error, permission, cancellation, reconnect, and destructive confirmation states pass.
- Formatting, unit tests, integration tests, E2E, production build, and secret-boundary checks pass.
- Visual/geometry evidence passes the acceptance matrix.
- No live-Hermes claim is made from mock-only evidence.

