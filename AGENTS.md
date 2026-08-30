# Agent execution contract

Read `plan.md`, `TASKS.md`, and `docs/compatibility-contract.md` before changing code.

## Non-negotiable rules

1. A milestone is complete only when its acceptance tests pass.
2. Do not remove an original Hermes WebUI flow. Map it to a task and preserve its contract until the replacement is proven.
3. UI redesign is allowed; workflow, information architecture, keyboard behavior, responsive behavior, and data compatibility must remain recognizable.
4. Browser code never receives Hermes API keys or provider credentials.
5. Normalize external Gateway events in `backend/internal/gateway`; UI components consume only `ChatEvent`.
6. Keep dependencies small. Prefer platform APIs and copy-owned shadcn primitives.
7. Update `TASKS.md` in the same change as implementation. Record decisions in `docs/decisions.md`.
8. Never mark live-Hermes verification complete from mock tests alone.

## Definition of done

- Formatting, unit tests, integration tests, and production build pass.
- Error, loading, empty, cancellation, and reconnect states are covered.
- No secret appears in frontend bundles, logs, snapshots, or error payloads.
- New behavior has a compatibility note and an acceptance test.

