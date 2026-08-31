# GoreeCloud Search Features

This file records Search functionality and implementation state. A listed capability is not a production-readiness or Stable claim unless current runtime and acceptance evidence support it.

## Implemented native foundations

### Native search engine

- GoreeCloud-owned Go search engine under `native/internal/search`.
- Bounded 512-rune query validation.
- Explicit General, Images, Videos, News, and Files category model.
- General-category development execution even with no configured providers.
- Specialized categories fail closed unless an executable provider path exists.
- Concurrent provider execution under a bounded request timeout; the request returns degraded timeout evidence even when an integration adapter fails to honor context cancellation.
- Bounded provider availability/timeout status without returning raw provider error messages.
- HTTP/HTTPS result URL normalization with fragments removed.
- Result URLs with embedded user-info credentials rejected.
- Provider identities normalized and bounded to 128 runes; blank/control-character provider names are not advertised or executed.
- Sanitized provider capability definitions without credentials, endpoints, runtime errors, or mutable controls.
- GoreeCloud-owned deterministic ranking v2 that evaluates the submitted query against result titles, snippets, and URL text locally for each request.
- Strong title/exact-match relevance with lower-weight snippet and URL coverage.
- Provider scores bounded to weak supporting evidence instead of being trusted as one universal cross-provider scale.
- Exact-URL de-duplication that preserves sorted source provenance and bounded multi-provider consensus evidence.
- Query-relevant representative selection when duplicate providers return different titles/snippets for the same URL.
- First-viewport hostname diversity when multiple relevant domains are available, with site/domain-directed queries exempt from diversity reshuffling.
- Deterministic final ordering and bounded ranking scores.

### Native GoreeCloud product experience

- GoreeCloud-owned native service entry point and web-presentation foundation.
- First-party homepage/results presentation work.
- Scan-first native result list that prioritizes source, title, URL, and snippet hierarchy over repeated elevated-card chrome.
- Compact persistent query field and category navigation on results.
- Multi-provider source-agreement disclosure without exposing internal numeric ranking scores to users.
- Source-health presentation separated from the primary result-reading flow.
- Adaptive wide/narrow layouts plus reduced-motion, increased-contrast, forced-colors, and reduced-transparency fallbacks in the native results stylesheet.
- Script-free Go-template result rendering with automatic escaping of provider/query/result content.
- First-party preference state and organization.
- GoreeCloud product identity rather than an upstream-only shell.
- Native-first development direction governed by the latest applicable Stable Glaze UI contract; Glaze UI 2.0.0 is the current required production target, with application-specific acceptance still incomplete.

### Privacy-first behavior

- No GoreeCloud advertising or sponsored-ranking business model.
- No intended behavioral profiling.
- Native ranking does not use click history, behavioral profiles, advertising signals, or remote ranking telemetry.
- Explicit query bounds and minimized native state.
- No hidden fallback requirement that bypasses GoreeCloud Search as configured Browser search authority.
- Provider errors reduced to bounded status codes.
- Result URLs containing embedded credentials rejected before presentation.
- Internal numeric relevance scores are not displayed as user-facing trust or quality claims.

### GoreeCloud Sync foundations

- Application-owned `search.history` capability contract.
- Exact dataset/schema validation.
- 512-byte record/cursor bounds.
- Authenticated submission requirements.
- Payload-free deletion tombstones.
- Bounded paginated retrieval.
- Client-side Ed25519 record-proof verification against the canonical GoreeCloud Sync vector.

These are source contracts; they do not imply that account search-history synchronization is enabled in a production environment.

## Transitional retained capabilities

The repository still contains the inherited SearXNG-derived runtime for continuity and migration. While it remains in use it may provide broader mature metasearch/provider/category functionality than the current native implementation.

Inherited capability presence is transitional evidence, not the target application architecture. Features must be retained, replaced, improved, or explicitly approved for retirement before the inherited runtime is removed.

## Planned / incomplete native capabilities

- Production-approved native external-provider adapters and credentials integration.
- Accepted native provider coverage for every category selected for release.
- Query correction/spelling assistance backed by an approved local or privacy-preserving implementation.
- Richer intent routing and specialized result understanding where justified by accepted native provider metadata.
- Freshness-aware ranking for result types that carry trustworthy publication/update timestamps.
- Optional semantic retrieval/reranking only if implemented through an approved privacy-preserving GoreeCloud-controlled path and validated against deterministic fallback behavior.
- Completed feature-parity migration from the inherited runtime.
- Full native Glaze UI 2.0 visual/accessibility/device acceptance.
- Complete Privacy Shield, Wardveil Security, Everkeep, GoreeCloud Identity, and GoreeCloud Mesh runtime/evidence integration where applicable.
- Production-approved account search history, saved searches, synchronization, or personalization.
- Governed machine-readable Search API for approved first-party/AI/research consumers.
- Complete monitoring, rate limiting, abuse protection, target-host deployment, restore, rollback, migration, and operational evidence.
- Stable lifecycle promotion of the native application.

## Feature-governance rule

Planning, source implementation, CI validation, runtime integration, production acceptance, migration cutover, and Stable promotion are separate states. Search documentation must preserve those distinctions.
