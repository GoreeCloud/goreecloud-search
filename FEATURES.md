# GoreeCloud Search — Current Features

**Lifecycle:** Development  
**Version:** `0.1.0.dev4`

This file records implemented behavior in the current development candidate. It does not claim release, deployment, production acceptance, or Stable qualification.

## Implemented in the current development candidate

- Native Python package under `src/goreecloud_search`.
- Typed search categories.
- Typed source modes.
- Query parser for the initial operator set.
- ISO date-range validation.
- Domain and file-extension normalization.
- Duplicate filter normalization.
- Deterministic provider eligibility filtering.
- Index-first primary/fallback planning.
- Federated planning.
- GoreeCloud-only planning.
- External-only planning.
- Offline/local planning.
- Explicit third-party query-disclosure signal.
- Fail-closed privacy invariant for GoreeCloud-only and offline/local modes.
- Replaceable provider protocol with typed provider search batches.
- Bounded asynchronous provider execution with configurable per-provider timeout and maximum concurrency.
- Outer cancellation propagation that cancels in-flight provider work.
- Provider failure isolation: one timeout/error does not fail successful providers.
- Explicit `available`, `degraded`, and `unavailable` execution states.
- Index-first fallback execution that invokes external fallback sources only when the primary stage does not fill the requested raw-result target.
- Fail-closed rejection of provider batches that spoof another provider's provenance.
- Versioned `goreecloud.search-index.v1` Search ↔ Index contract.
- First-party GoreeCloud Index adapter boundary with injected transport.
- Index category/capability negotiation, page-size enforcement, cursor pagination, repeated-cursor protection, and degraded/warning propagation.
- Enriched result-candidate data contract with canonical URL, source ID, content hash, language, crawl timestamp, and provider-contract provenance.
- Conservative URL canonicalization that removes fragments and known tracking parameters without collapsing HTTP/HTTPS semantics.
- Canonical-URL and content-hash deduplication.
- Multi-provider source-agreement counts and provenance aggregation.
- Fail-closed rejection of candidates from undeclared providers.
- Deterministic Search-owned ranking using quoted-phrase, title-term, snippet-term, explicit site/filetype/language, and bounded source-agreement signals.
- Per-result ranking signal records and human-readable “Why this result?” explanations.
- No behavioral-history, advertising-payment, click-profile, or hidden provider-specific ranking boost.
- Development CLI for query parsing without network access.
- Unit tests and CI for Python 3.11 and 3.12.
- Explicit internal application version.

## Not implemented

No authenticated live network provider ships in this candidate. The execution engine can run explicitly injected adapters, but the GoreeCloud Index adapter still has no authenticated runtime transport and no approved external network provider is included. No user-facing web UI, Browser integration, AI synthesis, persistent history, or production deployment exists yet.
