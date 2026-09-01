# Compatibility contract

This document is the migration guardrail between `nesquena/hermes-webui` and Hermes Web Studio.

## Kanban contract

Kanban uses native Hermes state through a server-side transport adapter. The
CLI is the default transport; an authenticated Hermes Dashboard may be used
when configured and healthy. The browser never receives the CLI path, Dashboard
session token, provider key, subprocess details, or native action output. Native
Native board, task, create, dispatch, and action responses use stable allowlisted projections; arbitrary CLI fields (including workspace paths, secrets, and subprocess details) never reach browser. Action success responses contain only stable action status fields.

The canonical statuses are `triage`, `todo`, `scheduled`, `ready`, `running`,
`blocked`, `review`, `done`, and `archived`. Studio stores the selected board
locally and passes it explicitly to CLI requests, so normal board selection does
not change Hermes' shared current-board pointer. A Space is resolved to an
upstream workspace value; no native `space` field is sent.

CLI mode supports board reads, task creation, dispatch, assignment, comments,
links, and named Hermes transitions. Generic task editing, arbitrary status
patches, plugin bulk mutations, and live event streaming are capability-gated
until the Dashboard transport is available. The earlier local `/api/kanban`
placeholder handlers remain deprecated compatibility routes while migration
and live acceptance are incomplete.

## Chat browser contract (implemented)

### Start

`POST /api/chat/start`

```json
{
  "session_id": "optional-stable-id",
  "message": "Hello Hermes",
  "model": "optional-model",
  "provider": "optional-provider"
}
```

Response:

```json
{"stream_id":"opaque-id","session_id":"stable-id"}
```

### Stream

`GET /api/chat/stream?stream_id=<opaque-id>` returns `text/event-stream`.

| Event | Required payload | Meaning |
|---|---|---|
| `token` | `{"text":"..."}` | Assistant text delta |
| `reasoning` | `{"text":"..."}` | Reasoning delta/summary |
| `tool` | `{"name":"...","args":{}}` | Tool started |
| `tool_complete` | `{"name":"...","is_error":false}` | Tool completed |
| `done` | `{"answer":"...","session_id":"..."}` | Successful terminal event |
| `cancel` | `{"message":"..."}` | User cancellation |
| `apperror` | `{"code":"...","message":"..."}` | Safe terminal error |

### Cancel

`POST /api/chat/cancel?stream_id=<opaque-id>` cancels the upstream request context and emits `cancel` when possible.

## Gateway contract (implemented)

- Base URL: `HERMES_WEBUI_GATEWAY_BASE_URL`, default `http://127.0.0.1:8642`
- Key: `HERMES_WEBUI_GATEWAY_API_KEY`, fallback `API_SERVER_KEY`
- Request: `POST /v1/chat/completions`
- Headers: `Authorization: Bearer …`, `X-Hermes-Session-Id`, `X-Hermes-Session-Key`
- Body: OpenAI-style `model`, `stream: true`, `messages`; optional `provider`

Supported upstream frames:

- OpenAI chunks: `choices[0].delta.content`
- Hermes `message.delta` / `message.complete`
- Hermes `reasoning.available` / `reasoning.delta` / `thinking.delta`
- Hermes `tool.start` / `tool.progress` / `tool.complete` (plus legacy `tool.started` / `tool.completed`)
- Hermes `subagent.start` / `subagent.progress` / `subagent.complete` and related lifecycle aliases
- Hermes `approval.request`
- Hermes `run.completed` / `run.failed`
- `[DONE]`

## Compatibility phases

Unimplemented routes must not be improvised. Inventory the original route, request shape, response shape, persistence effects, auth policy, and tests before implementation. Track that work in `TASKS.md`.

## Personal WebUI MVP baseline

Replacement parity is frozen against the read-only personal source at
`3caeca14064cec36c9c7b4f83ffade9a92cf2aee` plus the production behavior
audited on 2026-08-30. `MVP.md` defines the required browser surfaces,
feature matrix, evidence, appearance, responsive states, resource budgets, and
cutover gate.

The contract applies to observable behavior and state effects, not Python
implementation details. Go routes may use internal names that fit the new
architecture, but the reference browser workflow, validation, error state,
persistence effect, and compatibility aliases must remain available until the
replacement is proven. Shallow list/CRUD placeholders do not satisfy richer
personal-WebUI contracts such as cron delivery/history, Skills editing, Memory
sources, profile isolation, or session recovery.

The complete route/state/feature ownership map is maintained in
`docs/personal-webui-feature-manifest.md`. A manifest row is not an acceptance
claim; it is a traceability record linking the reference capability to its
Studio owner, parity task, and required evidence.

The primary rail and titlebar semantics are frozen in
`docs/shell-contract.md`. It defines stable ownership selectors, keyboard and
focus behavior, dialog/menu semantics, and responsive boundaries. Visual
geometry remains an evidence task and is not inferred from source inspection.

The composer may select an active profile and model from `/api/profiles`; the
selected model/provider is sent only in the server-bound chat start request.
Workspace selection opens the server-owned workspace panel. Context usage is
rendered only when normalized Gateway usage fields include both a total and a
context limit, so the UI never invents a quota or percentage.

### Model catalog contract

`GET /api/models/catalog` returns HTTP `200` with `{ "status": "ready", "models": [...] }`
when the local Gateway catalog is valid. Each model is sanitized and includes
an ID, display name, provider, aliases, and `available: true`. Gateway HTTP
errors, invalid JSON, oversized responses, and catalogs over the explicit
1,000-model item limit return HTTP `200` with
`{ "status": "unavailable", "models": [], "message": "..." }`; the browser
must not invent fallback or fake models. Browser state distinguishes `loading`
(request pending), `error` (request failed), `unavailable` (server returned no
usable catalog), and `ready`, including an explicit empty model list. Upstream
response bodies are bounded
to 1 MiB before parsing. IDs, providers, aliases, and other catalog text are
trimmed, control-character stripped, byte-bounded, and deduplicated by the
server.

If an active profile's model/provider is absent or unavailable in a normalized
ready catalog, the composer preserves that profile selection, visibly marks it
unavailable, and blocks submission until a valid catalog model or `default` is
selected. While catalog status is `loading`, `unavailable`, or `error`, any
non-`default` profile model is treated as stale and cannot be submitted. A
ready empty catalog is shown as an explicit empty state; it does not create
fake models. A valid profile keeps normal default selection behavior.

Production credentials, session names, model inventories, workspace paths, and
other private runtime data are never fixtures. Deterministic tests use sanitized
state derived from the frozen source contracts. If frozen source and production
behavior disagree, the task remains open until the difference has an explicit
disposition.

## M1 session persistence inventory

The frozen upstream session store is file-based and remains the compatibility source for the first M1 slice:

- State directory: `HERMES_WEBUI_STATE_DIR`, then `$HERMES_HOME/webui`, then POSIX `~/.hermes/webui`.
- Session files: `sessions/<session_id>.json`.
- Listing index: `sessions/_index.json`, with a `*.json` scan fallback when the index is missing, invalid, or empty.
- Session JSON: top-level metadata plus a `messages` array. Unknown top-level fields are retained by the reader.
- Safety: session IDs are single path components. Traversal, slash, and backslash forms are rejected.

The Go implementation uses the same JSON files as its durable session store. Writes are atomic, use `0600` files, preserve unknown top-level fields, and rebuild `_index.json` without transcript messages. SQLite or CLI session data is not treated as compatible until those formats receive their own inventory and tests.

Per-session mutations are serialized by the BFF's per-session lock,
including transcript append, truncate, metadata updates, duplicate, and delete.
Session-list/search reads (`GET /api/sessions`) read `_index.json` and may fall
back to a directory scan; scans may load session files without a per-session
read lock. Index rebuilds are serialized by a store-wide process-local mutex
and replace `_index.json` atomically, which serializes rebuild writers only.
Index and fallback reads do not provide a cross-file snapshot and may observe
mixed filesystem state. Separate processes still require filesystem locking or
CAS to provide cross-process coordination. A session detail read loads its
session file directly.

The current resolved state directory was inventoried on 2026-08-30: chat data is present as `sessions/*.json` plus `_index.json`; the separate `kanban.db` is not a chat session source and is not read by this service.

### Session API

`GET /api/sessions` returns `{ "sessions": [...] }` with compact metadata at the top level. `GET /api/sessions/{session_id}` returns the metadata plus the original ordered `messages` array. `POST /api/sessions` creates a session, `PATCH /api/sessions/{session_id}` updates metadata, and `DELETE /api/sessions/{session_id}` removes it. Action-compatible routes are available at `/rename`, `/pin`, and `/archive`. Missing sessions return `404`; unsafe IDs return `400`. Create and update preserve the legacy top-level JSON shape and unknown fields. Chat turns append user and assistant messages to the same transcript, allowing later loads to resume the stored history.

## M1 streaming, approvals, and attachments

Start requests may include `attachment_ids`, returned by `POST /api/attachments`. Stream events carry monotonically increasing numeric SSE IDs. Clients may reconnect with `Last-Event-ID` or `after=<id>`; completed turns remain replayable for five minutes. The browser also ignores any event whose ID is not newer than its local cursor, protecting the transcript from duplicate delivery during reconnect.

The normalized event set also includes `subagent`, `approval`, and `usage`. Approval decisions use `POST /api/runs/{run_id}/approval` with `{"choice":"once"}`, `{"choice":"session"}`, `{"choice":"always"}`, or `{"choice":"deny"}`. The BFF also accepts the older `decision: approved|denied` aliases and maps them to `once|deny`. It forwards the canonical choice using its server-side Gateway credential.

Set `HERMES_WEBUI_USE_RUNS_API=true` to create text-only turns through `POST /v1/runs` and subscribe to `/v1/runs/{run_id}/events`, the Gateway path that exposes structured approval events. Turns with attachments intentionally remain on legacy chat completions because the current adapter does not yet have an upstream Runs multimodal input contract. The default remains legacy chat completions for session continuity and multimodal compatibility until Runs API live parity is approved.

Attachment uploads are multipart `file` requests, limited to 10 MiB and image/PDF/plain-text MIME types. Files are stored with `0600` permissions under the resolved WebUI state directory and are addressed by opaque IDs. The Gateway adapter maps images to `image_url`, PDFs to `file`, and text files to text content. Browser code never receives Gateway or provider credentials.

Editing a persisted user message uses `POST /api/sessions/{session_id}/truncate` with `{"count":<message-count>}`. The BFF keeps the transcript prefix, preserves session metadata, and the next composer send creates the replacement branch.

Retry/regenerate uses the same prefix operation and immediately starts a replacement turn, preventing the previous assistant branch from being appended as a second answer. Turns serialize per session inside one BFF process; failed turns roll back only state they own. Rollback failures emit `apperror` and preserve state rather than claiming cleanup. Assistant persistence failure never emits `done`.

Gateway HTTP 200 completion frames must contain a non-empty answer and valid terminal SSE completion (`[DONE]` or recognized completion event). Truncated streams, scanner EOF, nested error fields, and empty completion payloads fail without assistant persistence. Current provider HTTP 402 insufficient-credit responses remain upstream limitations; they are reported as controlled errors and do not prove live completion or rollback beyond the safely owned session prefix.

## M2 workspace contract

Spaces are backed by a server-owned registry. `GET /api/spaces` returns stable IDs, names, local/remote location kinds, sanitized workspace references, ordering, active state, profile ownership, and health. `POST /api/spaces` validates registration; `POST /api/spaces/active` persists activation; `DELETE /api/spaces/{id}` protects the active Space. Remote Spaces return only an opaque `remote:<id>` reference and explicit `unavailable` health; raw configured remote references never return to the browser, and remote references are never resolved or silently mapped to local paths. Legacy or unsafe references return `unavailable`.

Profile isolation is partial: legacy Spaces with empty `profile_id` remain visible to every profile for compatibility. New Spaces are assigned to the active profile and profile-scoped Spaces are filtered by ownership; full isolation requires migrating legacy records.

The workspace surface is backed by one server-owned root. Set
`HERMES_WEBUI_DEFAULT_WORKSPACE` to choose it; when unset, the backend creates
`~/.hermes/webui/workspace` with private permissions. Browser paths are always
relative to that root. Absolute paths, traversal, backslash-separated paths,
and symlinks resolving outside the root are rejected. This is a pre-operation containment check; residual symlink replacement TOCTOU remains possible between validation and filesystem use, so the contract does not claim absolute elimination.

The BFF exposes `GET /api/workspace/tree?path=.`,
`GET /api/workspace/preview?path=...`, `GET /api/workspace/download?path=...`,
and `GET /api/workspace/git?path=...`. Mutations use `POST /api/workspace/item`,
`PUT /api/workspace/file`, `POST /api/workspace/rename`,
`DELETE /api/workspace/item`, and multipart `POST /api/workspace/upload`.
Text previews and writes are limited to 1 MiB; uploads are limited to 10 MiB.
Files are written with `0600` and created directories with `0700`.

Text/code/Markdown previews return `content` and `editable`; images and other
binary files return `binary: true` and are only served through the download
endpoint. Git output is informational and never blocks file access. Workspace
state is independent from the chat stream, so an active stream does not clear
the selected preview.

The workspace panel owns Files, Artifacts, and Todos tabs. Files uses the
workspace contract above, Todos reads the server-owned control collection, and
Artifacts remains an honest empty state until an artifact route is available.
Panel open state and width are browser layout preferences persisted locally;
workspace contents and mutations remain server-owned.

Primary navigation visibility and order are layout preferences. The
customization dialog persists those preferences locally, keeps Chat visible,
and only exposes destinations that have a working Studio view. Full reference
section registry parity remains a later task. The compact primary rail does not
duplicate Settings or expose Customize navigation; that action is owned by the
Settings navigation sidebar. New conversation is owned by the Chat/session
sidebar so session actions stay adjacent to recent sessions.

Theme boot reads local layout preferences before the React module loads and
sets `data-theme` and `data-skin` on the document root. Preferences also
persist the selected theme, skin, and locale through the server-backed
preferences route. The built-in registry contains all 21 frozen skin names;
visual tuning and screenshot evidence remain separate acceptance work.

Mobile exposes five working bottom-navigation destinations: Chat, Tasks,
Skills, Memory, and Settings. Each target is at least 44 CSS pixels. The
composer reserves the bottom safe-area/navigation space, while the primary
rail remains a labeled drawer and the workspace remains a mobile slide-over.
Reduced-motion users do not receive the panel transition.

## M3 identity and profile contract

`GET /api/onboarding` reports whether local password setup is complete and
which identity providers are configured. The first-run action is
`POST /api/onboarding/password` with a password of at least 12 characters.
`POST /api/auth/login` issues a signed, HttpOnly, SameSite cookie; the password
and signing material remain in the server state directory. Login attempts are
rate limited per client address. `POST /api/auth/logout` clears the cookie and
`GET /api/auth/me` reports the current identity without exposing credentials.

When password auth is enabled, protected requests require the signed cookie.
Mutating requests with an `Origin` header must match the request host. A
non-loopback server refuses protected traffic until authentication is set up.
`HERMES_WEBUI_TRUSTED_USER_HEADER` does not enable client-supplied trusted-header
authentication. The signed HttpOnly session cookie remains authoritative. SSO
remains capability-only until server-side proxy identity and boundary validation
exists.

`GET /api/profiles` returns only safe profile IDs, names, models, providers,
and health labels. `POST /api/profiles/active` switches the active profile.
Profiles are supplied through server-only `HERMES_WEBUI_PROFILES_JSON`; secrets
and arbitrary fields are discarded during decoding. OIDC issuer discovery and
WebAuthn ceremonies remain explicit capability states until deployment
configuration is present; unavailable providers are not shown as working login
paths.

## M4 control-center contract

Tasks, todos, goals, and spaces use server-owned JSON state at `control.json`.
Collections expose `GET /api/control/{collection}`,
`POST /api/control/{collection}`, `PATCH /api/control/{collection}/{id}`, and
`DELETE /api/control/{collection}/{id}`. `GET` and `PUT /api/preferences` store
non-secret display preferences; keys containing token or secret are discarded.
`GET /api/skills` and `GET /api/memory` discover local server-state entries and
return empty arrays when no source is configured. `GET /api/capabilities`
declares runtime-dependent surfaces such as terminal, voice, background work,
and extensions instead of exposing non-functional controls.

Skills and memory now use the Hermes home selected by `HERMES_HOME` (default
`~/.hermes`), independent of the Web Studio metadata state directory. Skills
are discovered recursively from `skills/**/SKILL.md` with safe frontmatter
metadata; `GET /api/skills?name=<relative-SKILL.md>` returns bounded content for
preview. Memory lists `memories/MEMORY.md`, `memories/USER.md`, and root
`SOUL.md`; `GET /api/memory?name=<allowed-file>` returns bounded content.

## M5 distribution contract

`make build` produces a Go binary that serves the compiled frontend from an
embedded filesystem. `make migrate-backup` creates a timestamped private copy
of the state directory and `make migrate-restore` restores the newest copy;
neither command performs an implicit destructive migration. `install.sh`
installs the binary under `$HOME/.local/bin` by default. Release CI covers
Linux amd64, macOS arm64, and Windows amd64 artifacts. Docker and Nix consume
the same frontend build contract, while hosted visual/security/beta sign-off
remains a release process rather than a local mock claim.

The upstream scheduled-task route names are retained as compatibility aliases:
`GET /api/crons`, `GET /api/crons/history`, `POST /api/crons/create`,
`POST /api/crons/run`, `POST /api/crons/pause`, and
`POST /api/crons/resume`. Jobs are persisted in the same server-owned control
state and run history is bounded to the latest 100 records. A wall-clock
scheduler and external delivery remain deferred until the Gateway task runtime
contract is available. `/api/settings` GET/POST aliases map to the safe
preferences store, and `/api/plugins` returns sanitized visibility metadata.

## M6 operations contract

`GET /health` is a liveness check and returns `200` while the HTTP process is
running. `GET /ready` checks locally initialized session, workspace, auth, and
control services without contacting Hermes Gateway. It returns `200` with
`ready: true` when the service can accept traffic, and `503` with per-check
details when initialization failed. Gateway reachability remains a separate
`/api/health/hermes` diagnostic so a temporarily offline Gateway does not make
the BFF process appear dead.

The mobile navigation drawer is an ephemeral responsive surface: when an open
drawer crosses into the desktop breakpoint, it closes and the compact primary
rail is restored. Resizing back to mobile still requires the explicit menu
action, so desktop layout state cannot leak into the mobile drawer.

The fixed mobile bottom navigation follows the same `1024px` breakpoint and
must be hidden on desktop; the desktop primary rail remains the only persistent
navigation surface at that width.

The composer supports Queue, Interrupt, and Steer turn modes. Pending turns
retain attachment names for visible intent, can be removed individually, and
are cleared on session switch, reset, cancellation, and terminal stream states.
The current browser acceptance covers the shell and responsive boundaries;
live mode semantics remain a separate Hermes verification gate.

Skills discovery supports server-owned query, create, update, and delete
operations at `/api/skills`, with relative-path and symlink containment checks.
The UI keeps preview and editor loading, empty, success, and error states
separate. Grouping, activation, linked files, usage, and profile-aware refresh
remain unimplemented until their upstream contracts are inventoried.

The server also exposes registered Spaces and operator diagnostics without
putting workspace paths or credentials in browser-owned state. Profile and
provider CRUD responses contain only safe metadata; provider secrets are
represented by `has_key` and are never returned. The control surface reports
15 supported locale identifiers with English fallback.

MCP and extension surfaces are server-owned contracts. `GET/POST/PATCH/DELETE
/api/mcp/servers` stores validated stdio, HTTP, or SSE server metadata without
executing a process, and `GET /api/mcp/servers/{id}/tools` exposes only
sanitized tool metadata as read-only state. MCP settings remove
secret-shaped keys before persistence and response. `GET
/api/plugins` returns sanitized plugin metadata and explicit capability states;
`GET /api/extensions` returns registered extension directory names without
loading extension settings or executing code. Registry visibility is available,
while install, toggle, uninstall, settings, sidecar consent, iframe loading,
and trust-boundary operations return capability metadata marked unavailable
until a sandbox and consent contract exists. The existing plugin settings
mutation remains server-owned and filters secret-shaped keys; `PATCH
/api/plugins/{id}` can update discovered plugin metadata/settings without
executing plugin code. The registry response does not expose settings. `GET /api/operator/version`
reports server runtime version metadata, and `GET /api/operator/update`
reports health links, release-artifact evidence state, and explicit unavailable
check/apply/shutdown/restart/lock-recovery actions. Updates are never applied
from the browser; process lifecycle remains owned by the service supervisor.

Conversation runtime helpers define disclosure, compaction repair, branch
lineage, session projection, model search, safe export, and deterministic
lifecycle rows. These helpers do not certify the corresponding upstream/live
Hermes behavior until browser and side-by-side evidence is recorded.
Session reads, duplication, exports, mutations, and chat turns use one
process-local per-session lock, so concurrent writers are serialized within a
single BFF process; this is not a multi-process compare-and-swap guarantee.

The production frontend exposes a manifest and service worker under the Vite
base path. Registration is production-only and uses the configured base path;
API requests are network-only and are never placed in the offline cache;
offline cache invalidation and browser install/bfcache evidence remain open.

Operator Insights reads server-owned session facts and reports usage,
provider-history, synchronization, and an explicit cost-unavailable state when
no authoritative cost source exists. Terminal is reported as unavailable until
the server has a contained process contract; the UI never fabricates terminal
output or a successful capability.

Authentication rotates the signed session cookie with per-login entropy, marks
cookies Secure behind TLS or an explicitly trusted HTTPS reverse proxy, and
rejects cross-origin logout requests under the same-origin policy.

Workspace browsing is relative-path-only and resolves symlinks against the
configured root before listing, previewing, downloading, editing, renaming,
deleting, or uploading. Browser actions expose copy/open/download and an
explicit unavailable reveal state; they do not receive an operating-system
path. Git integration is currently read-only branch/status/diff projection,
with hostile-path containment tests; stage, discard, commit, sync, and
checkout actions remain deferred until a confirmation and ownership contract
exists.

The local certification runners are evidence generators, not parity approval:
the route matrix checks registered aliases, the visual matrix writes sanitized
screenshots under `/tmp`, and the accessibility audit reports open gaps instead
of treating them as passed; the current local audit passes the 44px touch-target
contract. Hosted, live-Hermes,
screen-reader, and independent-review gates remain required for M11.

External sessions are projected from server-owned session metadata. The rail
may show source, channel, identity, and routing labels and an explicit
unavailable handoff state, but it does not fabricate channel transport,
round summaries, or model-switch behavior without a Gateway channel contract.

Capability panels use the shared response reader. Structured JSON error
messages remain user-facing, while non-JSON proxy/server responses are mapped
to an HTTP status error; a parser exception is never shown as the product
state.

MCP and plugin settings are server-owned. `GET` and `POST /api/mcp/servers`
plus `PATCH` and `DELETE /api/mcp/servers/{id}` manage validated metadata;
`GET /api/mcp/servers/{id}/tools` is read-only and never executes a tool.
HTTP/SSE endpoints require HTTP(S) URLs without userinfo, stdio requires a
command, and secret-shaped settings are filtered before persistence and
response. `PATCH /api/plugins/{id}` can update settings for a discovered
plugin, while install, execution, and remote MCP discovery remain deferred.
