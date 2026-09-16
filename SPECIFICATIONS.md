# GoreeCloud Search — Repository Specifications

## Lifecycle

- Product: GoreeCloud Search
- Version: `0.1.0.dev7`
- Lifecycle: Development
- Stable: No
- License: `AGPL-3.0-or-later`

## Current implementation boundary

The current native development candidates provide a query parser and privacy-aware source planner, bounded asynchronous provider execution, a versioned Search ↔ Index contract boundary with pagination, Search-owned normalization/deduplication, content policy, deterministic baseline ranking, and an initial transparent Search-local Lens reranking layer.

### Query model

The parser currently supports free-text terms, double-quoted phrases, excluded terms, `site:`, `-domain:`, `filetype:`, `ext:`, `before:`, `after:`, `language:`, `region:`, `source:`, `category:`, and `lens:`.

Unsupported syntax must remain ordinary query text or be rejected explicitly rather than causing hidden network behavior.

### Source modes and privacy

The source planner implements `index_first`, `federated`, `goreecloud_only`, `external_only`, and `offline_local`. The executor may run only providers named by the plan. `QueryDisclosureBudget` can cap distinct third-party providers before execution. `goreecloud_only` and `offline_local` must never produce a plan that discloses the query to a third-party provider.

### Provider and Index boundaries

Provider implementations conform to typed descriptors/search contracts. Execution is bounded by concurrency and timeout controls with cancellation and failure isolation. `goreecloud.search-index.v1` defines the current transport-injected Search ↔ Index model with category/capability negotiation, cursor pagination, provenance, and degraded-state propagation. No authenticated live Index client ships yet.

### Ranking boundary

Baseline ranking uses transparent deterministic query/result signals. GoreeCloud Index presence has zero ranking weight. Provider rank, click history, advertising payment, cross-query profiles, and hidden behavioral signals are not used.

### Lens boundary

A `Lens` is a Search-local deterministic set of rules. The current rule targets are domain, filetype, and language; actions are boost, lower, and exclude. Boost/lower weights are explicit and bounded. Matching subdomains are recognized for domain rules. Every applied score adjustment becomes an inspectable `RankingSignal`, and every exclusion is returned as Lens exclusion evidence.

Lens registry names use a stable case-insensitive lookup form while preserving the user-visible Lens name. An unknown requested Lens fails before provider execution. Search removes `query.filters.lens` from the provider-facing query before execution, so the selected local Lens is not sent to GoreeCloud Index or federated adapters through this contract.

Lens persistence, remote discovery, import/export, sharing, synchronization, trust/signature handling, and UI are not implemented.

### Content-policy boundary

Content policy is evaluated after normalization and before ranking. Hooks return attributable allow/warn/block decisions. Block dominates warn, and hook exception/spoofed provenance fails closed. The built-in `DomainPolicyHook` is explicit administrator domain policy, not a production content classifier. Moderate/Strict SafeSearch is rejected before provider execution unless at least one configured hook declares enforcement.

## Not yet implemented

- Authenticated live GoreeCloud Index transport and runtime integration.
- Approved external search-provider adapters.
- HTTP API and health/readiness endpoints.
- Snippet generation beyond provider-supplied snippets.
- Production SafeSearch classification sources and verified Wardveil runtime safety feeds.
- Lens persistence, export/import, sharing, synchronization, and UI.
- Private View, GoreeCloud Browser integration, and GoreeCloud AI answer integration.
- Identity-authenticated service requests, Privacy Shield runtime authorization, GoreeCloud Mesh discovery/events, and GoreeCloud Manager administration.
- User/admin Glaze UI surfaces, persistent search history/GoreeCloud Sync, and production deployment artifacts.

These remain planned and must not be represented as implemented until code and verification evidence exist.
