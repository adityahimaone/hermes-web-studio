# M5 beta and cutover checklist

## Parallel beta

- [ ] Choose a disposable state directory and backup the original state.
- [ ] Start original Hermes WebUI and Hermes Web Studio against the same Gateway.
- [ ] Compare session listing, resume, chat streaming, attachments, workspace,
  profile switch, and control-center state.
- [ ] Record differences in `TASKS.md` with route and fixture evidence.
- [ ] Exercise stop, reconnect, auth failure, workspace traversal, backup, and
  restore scenarios.
- [ ] Confirm no Gateway key appears in browser requests or built artifacts.

## Cutover

- [ ] Confirm release artifacts for the target OS and checksum them.
- [ ] Create and verify a state backup with `make migrate-backup`.
- [ ] Run smoke and browser acceptance against the release binary.
- [ ] Keep the original WebUI available for rollback.
- [ ] Announce the new binary and rollback command to operators.
- [ ] Archive the original implementation only after beta sign-off.
