# GoreeCloud Search Production Checklist

Use this as the short operational companion to `TARGET-ACCEPTANCE.md`, `STABLE-CUTOVER.md`, `RECOVERY-ACCEPTANCE.md`, and `FINAL-ACCEPTANCE.md`.

- [ ] Exact release source SHA recorded.
- [ ] Exact immutable candidate image reference/digest recorded.
- [ ] Candidate release evidence passes and remains non-authorizing for production cutover.
- [ ] Current known-good Search runtime, configuration, networks, mounts, image and Caddy backend are recorded for rollback.
- [ ] Source-controlled rollback baseline matches the verified known-good image identity or any difference is investigated before acceptance.
- [ ] Staged GoreeCloud Search is healthy without replacing production.
- [ ] Candidate-bound `bash goreecloud/target_acceptance.sh` passes against the staged target-host instance with exact digest/source identity.
- [ ] Sanitized target-runtime evidence is retained with the candidate evidence set.
- [ ] Representative `general`, `images`, `news`, `videos`, `files`, `it`, and `science` searches are accepted or provider-specific failures are classified.
- [ ] Candidate-bound real-provider evidence is generated against the same loopback-staged container rather than the pre-cutover production hostname.
- [ ] Provider evidence proves the staged container is the exact immutable candidate before and after provider requests, including matching image reference/image ID and OCI source revision.
- [ ] Candidate-bound real-provider evidence JSON is retained for the exact final source/image.
- [ ] All five first-Stable required provider categories—General, Images, Videos, News, and Files—have passing final-candidate evidence.
- [ ] Glaze UI 1.1 is accepted on deployed Compact, Medium, Expanded and Wide layouts.
- [ ] Compact light and Compact dark exact-final-candidate visual evidence is retained.
- [ ] Expanded light and Expanded dark exact-final-candidate visual evidence is retained.
- [ ] The retained visual-review artifact has an immutable `sha256:<digest>` recorded in `visual-evidence.json`.
- [ ] Physical Android/mobile Preferences review is complete and has an evidence reference.
- [ ] Desktop regression/final visual review is complete and has an evidence reference.
- [ ] Persisted light/dark theme preference behavior is verified and has an evidence reference.
- [ ] `visual-evidence.json` identifies the exact final Search source/image and keeps production cutover unauthorized.
- [ ] GoreeCloud Browser runtime integration uses GoreeCloud Search as the only/default browser search provider and has no silent external fallback.
- [ ] Exact GoreeCloud Browser source revision is recorded in `browser-evidence.json`.
- [ ] The retained actual Browser-runtime artifact has an immutable `sha256:<digest>` recorded in `browser-evidence.json`.
- [ ] Browser runtime evidence identifies the same exact Search candidate source/image used by all other final evidence.
- [ ] Browser address-bar, new-tab, and dedicated search-field queries all route through GoreeCloud Search.
- [ ] Browser Search-unavailability and recovery behavior is verified without external provider bypass.
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
- [ ] Final-candidate schema-version 2 manifest rejects provider artifacts without verified staged runtime binding.
- [ ] Final-candidate manifest cryptographically binds exactly six JSON artifacts: release, target runtime, recovery, provider, visual, and Browser evidence.
- [ ] Final-candidate manifest rejects changed visual or Browser evidence after assembly.
- [ ] Final-candidate manifest rejects visual/Browser evidence that identifies a different Search source/image.
- [ ] Final-candidate manifest records Glaze UI 1.1 and actual Browser runtime acceptance as complete while keeping `production_cutover_authorized` false.
- [ ] Caddy validation succeeds before any route change.
- [ ] Rollback procedure is verified and previous known-good Search state remains recoverable.
- [ ] Production hostname passes post-cutover identity, health and representative-search checks when cutover is actually performed.
- [ ] Authoritative GoreeCloud inventories, monitoring names and Search change log are updated after successful cutover.

No individual checklist item or evidence artifact independently authorizes production cutover. Stable promotion requires the complete reviewed evidence set, a passing six-artifact final manifest, and an explicit release decision.
