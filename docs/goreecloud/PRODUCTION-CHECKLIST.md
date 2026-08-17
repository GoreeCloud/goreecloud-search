# GoreeCloud Search Production Checklist

Use this as the short operational companion to `TARGET-ACCEPTANCE.md`.

- [ ] Exact source SHA recorded.
- [ ] Exact immutable image reference/digest recorded.
- [ ] Existing SearXNG runtime, configuration, networks, mounts, image and Caddy backend recorded for rollback.
- [ ] Staged GoreeCloud Search is healthy without replacing production.
- [ ] `target_acceptance.sh` passes against the staged target-host instance.
- [ ] Representative `general`, `images`, `news`, `videos`, `it`, and `science` searches are accepted or provider-specific failures are classified.
- [ ] Glaze UI is accepted on deployed Compact, Medium, Expanded and Wide layouts.
- [ ] Private DNS resolves the approved endpoint.
- [ ] Caddy serves a trusted certificate for `search.goreecloud.com`.
- [ ] Approved NetBird clients succeed and unapproved sources are denied.
- [ ] GoreeCloud Search publishes no unnecessary public backend port.
- [ ] Privacy/security response headers and logging behavior are accepted.
- [ ] Monitoring and alert delivery are verified.
- [ ] Required configuration is in backup scope.
- [ ] Representative restore/recreation succeeds in isolation.
- [ ] Caddy validation succeeds before cutover.
- [ ] Rollback procedure is verified and previous SearXNG state remains recoverable.
- [ ] Production hostname passes post-cutover identity, health and representative-search checks.
- [ ] Authoritative GoreeCloud inventories, monitoring names and Search change log are updated after successful cutover.
