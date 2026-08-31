# Kanban implementation record

Kanban is M9/P033. The implementation follows the frozen personal WebUI's
board hierarchy and interaction model while using the current Hermes Agent
status superset: `triage`, `todo`, `scheduled`, `ready`, `running`, `blocked`,
`review`, `done`, and `archived`.

## Current slice

The first vertical slice is implemented with a server-side CLI transport:

- native board and task reads from `hermes kanban ... --json`;
- board-local selection persisted in browser storage;
- Chat-like collapsible Kanban context rail;
- search, assignee, tenant, archived, loading, empty, and error states;
- task creation with title, body, priority, assignee, workspace, triage, skills,
  runtime, retry, goal, parent, and idempotency fields;
- dispatch endpoint and safe named task actions;
- capability reporting for CLI limitations;
- deprecated local placeholder routes retained during migration.

## Follow-up slices

Dashboard detection and server-side REST/WebSocket proxying will add full task
editing, links, bulk partial-failure results, orchestration, diagnostics,
attachments, board metadata, and live event SSE. The frontend will retain one
normalized `/api/kanban/*` contract and capability-gate controls unavailable in
CLI-only mode. Legacy Studio boards/cards will receive an explicit dry-run and
idempotent import path; legacy data will not be deleted automatically.

Completion requires backend/frontend tests, responsive and keyboard coverage,
secret-boundary checks, and live proof against both CLI-backed native state and
an authenticated Hermes Dashboard. Mock fixtures alone do not complete P033.
