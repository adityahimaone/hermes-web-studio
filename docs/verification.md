# Verification record

## 2026-08-30 — M0 local verification

Environment: Linux amd64, Go 1.27.0, Node 24.19.0, pnpm 11.19.0.

| Check | Result |
|---|---|
| `go test ./...` | Pass: Gateway parser and HTTP integration suites |
| `pnpm test` | Pass: 3 frontend event reducer tests |
| `pnpm build` | Pass |
| Frontend production output | JS 365.87 KiB / 114.27 KiB gzip; CSS 23.43 KiB / 5.24 KiB gzip |
| Real Hermes Gateway | Not available in this workspace; live gate remains open |

The backend integration suite uses an HTTP mock that implements the Hermes Gateway streaming boundary. It verifies the outgoing bearer/session headers and body, incoming SSE normalization, safe authentication failure, and upstream cancellation.

Live verification must be added as a separate dated entry using `scripts/smoke-hermes.sh`. Do not convert the live task to complete based on this mock result.
