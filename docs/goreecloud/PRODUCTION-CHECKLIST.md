# GoreeCloud Search Production Checklist

Use this as the short operational companion to `TARGET-ACCEPTANCE.md`, `STABLE-CUTOVER.md`, and `RECOVERY-ACCEPTANCE.md`.

- [ ] Exact release source SHA recorded.
- [ ] Exact immutable candidate image reference/digest recorded.
- [ ] Candidate release evidence passes and remains non-authorizing for production cutover.
- [ ] Current known-good Search runtime, configuration, networks, mounts, image and Caddy backend are recorded for rollback.
- [ ] Source-controlled rollback baseline matches the verified known-good image identity or any difference is investigated before acceptance.
- [ ] Staged GoreeCloud Search is healthy without replacing production.
- [ ] Candidate-bound `bash goreecloud/target_acceptance.sh` passes against the staged target-host instance with exact digest/source identity.
- [ ] Sanitized target-runtime evidence is retained with the candidate evidence set.
- [ ] Representative `general`, `images`, `news`, `videos`, `files`, `it`, and `science` searches are accepted or provider-specific failures are classified.
- [ ] All five first-Stable required provider categories—General, Images, Videos, News, and Files—have acceptable final-candidate evidence.
- [ ] Glaze UI is accepted on deployed Compact, Medium, Expanded and Wide layouts.
- [ ] Physical Android/mobile Preferences review is complete.
- [ ] Desktop regression/final visual review is complete.
- [ ] GoreeCloud Browser runtime integration uses GoreeCloud Search as the only/default browser search provider and has no silent external fallback.
- [ ] Private DNS resolves the approved endpoint.
- [ ] Caddy serves a trusted certificate for `search.goreecloud.com`.
- [ ] Approved NetBird clients succeed and unapproved sources are denied.
- [ ] GoreeCloud Search publishes no unnecessary public backend port.
- [ ] Privacy/security response headers and logging behavior are accepted.
- [ ] Monitoring identifies GoreeCloud Search at the appropriate migration point.
- [ ] Approved alert delivery is verified.
- [ ] Authoritative Search deployment/configuration is in approved backup scope.
- [ ] Protected runtime-configuration recovery path is verified without placing secret values in evidence.
- [ ] Representative application-level restore/recreation succeeds in isolation.
- [ ] Completed recovery evidence validates against the exact release evidence, target-runtime evidence, and rollback baseline.
- [ ] Caddy validation succeeds before any route change.
- [ ] Rollback procedure is verified and previous known-good Search state remains recoverable.
- [ ] Production hostname passes post-cutover identity, health and representative-search checks when cutover is actually performed.
- [ ] Authoritative GoreeCloud inventories, monitoring names and Search change log are updated after successful cutover.

No individual checklist item or evidence artifact independently authorizes production cutover. Stable promotion requires the complete reviewed evidence set and an explicit release decision.
