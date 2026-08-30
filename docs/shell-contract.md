# Shell contract

This is the P005 acceptance contract for the primary rail and titlebar. It
freezes behavior and semantics before the richer parity panels are migrated.

## Stable DOM responsibilities

| Selector/role | Responsibility |
|---|---|
| `[data-testid="primary-rail"]` | Desktop rail, hidden below the desktop breakpoint and opened as the mobile drawer |
| `[data-testid="primary-navigation"]` | Real destinations for every visible primary action |
| `[data-testid="titlebar"]` | Current view title, connection state, identity controls, reset, and workspace toggle |
| `#main-content` | Conversation or control-center content, with `min-width: 0` to prevent horizontal overflow |
| `role="tooltip"` | Keyboard/focus and pointer description for icon-only rail actions |
| `role="menu"` / `role="menuitem"` | Session overflow actions, dismissible with Escape and navigable with arrows/Home/End |
| `role="dialog"` / `aria-modal="true"` | Session/workspace confirmations, dismissible by close, backdrop, or Escape |

## Keyboard and focus

- Every icon-only control has an accessible name and a focus-visible ring.
- Focused rail controls expose their tooltip through `aria-describedby`.
- The session overflow menu opens from its trigger and supports Arrow Up/Down,
  Home, End, Enter, Space, and Escape through native buttons plus menu key
  handling.
- Dialogs focus their close control on open and close on Escape or backdrop
  click. Destructive actions remain explicit buttons inside the Dialog.
- Enter sends a composer message and Shift+Enter inserts a newline; this
  contract remains owned by the composer and is covered by the M1 browser row.

## Responsive contract

- At desktop width, the primary rail is icon-only and the Chat session rail is
  independently visible/collapsible.
- Below the desktop breakpoint, the primary rail is a labeled drawer opened by
  the titlebar menu button. Closing it returns focus to the menu trigger through
  the existing browser interaction flow.
- Content columns use `min-width: 0`; session titles truncate inside their own
  flex region so overflow actions never cover the title.
- Controls that are intended for touch use at least 44 CSS pixels in the rail,
  session list, and mobile navigation states.

## Evidence gate

P005 is complete only when `scripts/check-m7-shell-contract.sh`,
`frontend/e2e/m7-shell-check.mjs`, frontend unit tests, the production build,
and `git diff --check` pass. Screenshots and computed-style measurements are
evidence for P012, not inferred from a build.
