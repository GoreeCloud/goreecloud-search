# GoreeCloud Search — Current Features

**Lifecycle:** Development  
**Version:** `0.1.0.dev14`

This file records implemented behavior in the current development candidate. It does not claim release, deployment, production acceptance, or Stable qualification.

## Implemented in the current development candidate

- Native Python package under `src/goreecloud_search`.
- Typed search categories and source modes.
- Query parser for the initial operator set with ISO date validation and normalized repeatable filters.
- Deterministic provider eligibility filtering and Index-first, Federated, GoreeCloud-only, External-only, and Offline/local planning.
- Explicit third-party query-disclosure signal.
- Optional per-query disclosure budget limiting the number of distinct third-party providers admitted to a plan.
- Plan evidence for selected third-party provider count, applied budget, and providers omitted by the budget.
- Fail-closed privacy invariant for GoreeCloud-only and offline/local modes.
- Replaceable provider protocol with typed provider search batches.
- Bounded asynchronous provider execution with configurable per-provider timeout and maximum concurrency.
- Outer cancellation propagation and provider failure isolation.
- Explicit `available`, `degraded`, and `unavailable` execution states.
- Index-first fallback execution only when the primary stage does not fill the requested raw-result target.
- Fail-closed rejection of provider batches that spoof another provider's provenance.
- Versioned `goreecloud.search-index.v1` Search ↔ Index contract and first-party adapter boundary with injected transport.
- Index category/capability negotiation, page-size enforcement, cursor pagination, repeated-cursor protection, and degraded/warning propagation.
- Dedicated cycle-safe Index-originated delegation path using external-only primary providers with Index re-entry and fallback disabled.\n- Bounded Index-originated HTTP source boundary with capability discovery, JSON-body POST, health/readiness endpoints, injected GoreeCloud Identity requester verification, injected Privacy Shield capability-reference verification/consumption, strict request bounds, and generic fail-closed authorization errors.\n- Capability evidence is deliberately `production_accepted=false`; the source boundary is Development evidence rather than runtime or Production acceptance.
- Enriched result candidates with canonical URL, source ID, content hash, language, crawl timestamp, and provider-contract provenance.
- Conservative URL canonicalization, canonical-URL/content-hash deduplication, multi-provider source agreement, and provenance aggregation.
- Fail-closed rejection of candidates from undeclared providers.
- Deterministic Search-owned ranking using quoted-phrase, title-term, snippet-term, explicit site/filetype/language, and bounded source-agreement signals.
- Per-result ranking signal records and human-readable “Why this result?” explanations.
- No behavioral-history, advertising-payment, click-profile, or hidden provider-specific ranking boost.
- Development CLI for query parsing without network access.
- Unit tests and CI for Python 3.11 and 3.12.
- Explicit internal application version.

## Not implemented

No approved live external provider, credential issuer, live Identity/Privacy Shield verifier transport, user-facing web UI, Browser integration, AI synthesis, persistent history, or production deployment ships in this candidate. The authenticated Index HTTP boundary is source/CI infrastructure only and is not automatically started or registered.
