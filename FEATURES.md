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
- At most 512 results from any one provider enter native sanitization/ranking for a request; oversized provider result slices are copied down to the bounded working set and the provider status records `truncated=true`.
- HTTP/HTTPS result URL normalization with fragments removed.
- Result URLs with embedded user-info credentials rejected.
- Provider identities normalized and bounded to 128 runes; blank/control-character provider names are not advertised or executed.
- Sanitized provider capability definitions without credentials, endpoints, runtime errors, or mutable controls.
- GoreeCloud-owned deterministic ranking v2 that evaluates the submitted query against result titles, snippets, and URL text locally for each request.
- Strong title/exact-match relevance with lower-weight snippet and URL coverage.
- Local query-intent parsing for `site:`, domain-directed, `filetype:`/`ext:`, quoted-phrase, and temporal signals without rewriting the provider query.
- Explicit site/domain matches and requested file extensions receive bounded ranking preference while mismatched explicit operators are demoted.
- Quoted multi-word phrases receive bounded title/snippet/URL ordering boosts.
- Bounded one-edit or adjacent-transposition tolerance for query tokens of at least five runes, with substantially lower weight than exact title/token relevance.
- Conservative user-visible query correction derived only from result-title evidence already returned for the request. Search proposes at most one eligible token change, requires agreement from at least two independent normalized providers, fails closed on ambiguous alternatives, excludes quoted/operator/domain targets, and never rewrites the provider query.
- Short query tokens are excluded from fuzzy matching to reduce false-positive relevance.
- Provider scores bounded to weak supporting evidence instead of being trusted as one universal cross-provider scale.
- Explicit provider-level publication-timestamp authority: only adapters that declare `PublishedAt` authoritative from a trustworthy upstream publication/update field can retain that metadata for output and freshness ranking.
- Untrusted, zero, pre-Unix, and implausibly future publication timestamps are stripped before aggregation; retained metadata carries provider provenance through `published_at_source`.
- Bounded request-local freshness ranking for explicit temporal queries and News-category ranking, with no recency bias on ordinary General searches and a maximum 1,200-point contribution.
- Clear unquoted freshness modifiers such as `latest`, `recent`, `today`, `breaking`, `newest`, and `this week`/`this month` are separated from ordinary lexical relevance so they do not dilute the real subject terms.
- Quoted temporal language remains literal relevance text, while leading `current` can request freshness without misclassifying noun uses such as `electric current`; content-bearing `news`, `updated`, and `updates` remain lexical terms even when they activate freshness.
- Freshness is derived only from accepted provider timestamp metadata; Search does not infer publication time from snippets, URLs, crawl order, or arbitrary provider scores.
- Exact-URL de-duplication that preserves sorted source provenance and bounded multi-provider consensus evidence.
- Query-relevant representative selection when duplicate providers return different titles/snippets for the same URL.
- First-viewport hostname diversity when multiple relevant domains are available, with site/domain-directed queries exempt from diversity reshuffling.
- Domain detection distinguishes actual domain targets from dotted version-like tokens so queries such as `1.5.0` do not accidentally disable result diversity.
- Deterministic final ordering and bounded ranking scores.

### Native GoreeCloud product experience

- GoreeCloud-owned native service entry point and web-presentation foundation.
- First-party homepage/results presentation work.
- Scan-first native result list that prioritizes source, title, URL, and snippet hierarchy over repeated elevated-card chrome.
- Compact persistent query field and category navigation on results.
- Multi-provider source-agreement disclosure without exposing internal numeric ranking scores to users.
- Trusted retained publication timestamps are shown as concise human-readable dates with semantic machine-readable `<time datetime>` values; results without accepted timestamps do not receive synthetic dates.
- A user-visible “Search instead for” correction link is rendered only when the conservative local provider-agreement correction contract produces a single unambiguous alternative.
- The results ranking explanation discloses that trustworthy freshness is used when requested rather than presenting recency as an undisclosed ranking signal.
- Source-health presentation separated from the primary result-reading flow, including visible “limit applied” disclosure when a provider exceeds the native per-request processing ceiling.
- Adaptive wide/narrow layouts plus reduced-motion, increased-contrast, forced-colors, and reduced-transparency fallbacks in the native results stylesheet.
- Current native results acceptance enforces a 48px minimum target floor across Compact, Medium, Expanded, and Wide browser viewports.
- Script-free Go-template result rendering with automatic escaping of provider/query/result content.
- First-party preference state and organization.
- GoreeCloud product identity rather than an upstream-only shell.
- Native-first development direction governed by the latest applicable Stable Glaze UI contract; Glaze UI 2.1.0 is the current required production target. Glaze UI 2.2 is Candidate/design-reference work and is not a Stable consumer target. Application-specific whole-application and physical-device acceptance remain incomplete.

### Privacy-first behavior

- No GoreeCloud advertising or sponsored-ranking business model.
- No intended behavioral profiling.
- Native ranking does not use click history, behavioral profiles, advertising signals, or remote ranking telemetry.
- Intent parsing, typo-tolerant ranking, user-visible correction suggestions, and freshness scoring operate request-locally and do not send correction or recency lookups to another service.
- Freshness metadata is fail-closed unless its provider explicitly satisfies the publication-timestamp authority contract.
- Explicit query and provider-result processing bounds and minimized native state.
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
- Additional intent routing and specialized result understanding beyond the implemented site/domain/file-type/quoted-phrase/temporal ranking signals, where justified by accepted native provider metadata.
- Production-reviewed timestamp authority and live-provider freshness acceptance for result classes where recency is required; the source-level freshness contract is implemented but no provider is production-approved by that fact alone.
- Optional semantic retrieval/reranking only if implemented through an approved privacy-preserving GoreeCloud-controlled path and validated against deterministic fallback behavior.
- Completed feature-parity migration from the inherited runtime.
- Full native Glaze UI 2.1 whole-application visual/accessibility/device acceptance beyond the bounded native results evidence already implemented.
- Complete Privacy Shield, Wardveil Security, Everkeep, GoreeCloud Identity, and GoreeCloud Mesh runtime/evidence integration where applicable.
- Production-approved account search history, saved searches, synchronization, or personalization.
- Governed machine-readable Search API for approved first-party/AI/research consumers.
- Complete monitoring, rate limiting, abuse protection, target-host deployment, restore, rollback, migration, and operational evidence.
- Stable lifecycle promotion of the native application.

## Feature-governance rule

Planning, source implementation, CI validation, runtime integration, production acceptance, migration cutover, and Stable promotion are separate states. Search documentation must preserve those distinctions.