# M5 security threat model

## Assets

- Hermes Gateway and provider credentials
- Legacy session, attachment, control, and auth state
- Workspace files and generated artifacts
- Authentication cookies and profile metadata

## Boundaries and controls

| Boundary | Threat | Control | Evidence |
|---|---|---|---|
| Browser to BFF | credential exposure | credentials only loaded by Go config | artifact scan, API tests |
| Browser to BFF | CSRF and cross-origin writes | SameSite cookie plus Origin/host check | auth integration test |
| Client to login | brute force | per-client bounded login attempts | auth unit test |
| BFF to workspace | traversal/symlink escape | relative paths and resolved-root containment | workspace tests |
| State migration | destructive overwrite | timestamped backup and explicit restore command | migrate test |
| Embedded release | stale or secret frontend artifact | locked build and binary scan | M5 artifact script |
| Runtime execution | arbitrary shell/plugin execution | no terminal/plugin execution until sandbox contract exists | capability endpoint |

## Residual risks before remote release

- Run the release matrix on hosted Linux, macOS, and Windows runners.
- Perform an independent review of cookie rotation, reverse-proxy headers, and
  deployment TLS configuration.
- Run side-by-side beta against the original WebUI with rollback rehearsal.
- Do not enable trusted-header SSO unless the proxy strips client-supplied
  identity headers and authenticates upstream requests.
