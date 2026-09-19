# GoreeCloud Search — Current Features

**Lifecycle:** Development  
**Version:** `0.1.0.dev2`

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
- Replaceable provider protocol.
- Versioned `goreecloud.search-index.v1` Search ↔ Index contract.
- First-party GoreeCloud Index adapter boundary with injected transport.
- Index category/capability negotiation and page-size enforcement.
- Enriched result-candidate data contract with canonical URL, source ID, content hash, language, crawl timestamp, and provider-contract provenance.
- Conservative URL canonicalization that removes fragments and known tracking parameters without collapsing HTTP/HTTPS semantics.
- Canonical-URL and content-hash deduplication.
- Multi-provider source-agreement counts and provenance aggregation.
- Fail-closed rejection of candidates from undeclared providers.
- Development CLI for query parsing without network access.
- Unit tests and CI for Python 3.11 and 3.12.
- Explicit internal application version.

## Not implemented

No provider performs live network search in this candidate. The GoreeCloud Index adapter has a tested contract boundary but no authenticated runtime transport. No user-facing web UI, Browser integration, AI synthesis, persistent history, or production deployment exists yet.
