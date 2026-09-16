# GoreeCloud Search — Current Features

**Lifecycle:** Development  
**Version:** `0.1.0.dev7`

This file records implemented behavior in the current development candidate. It does not claim release, deployment, production acceptance, or Stable qualification.

## Implemented in the current development candidate

- Native Python package under `src/goreecloud_search`.
- Typed search categories and source modes.
- Query parser for the initial operator set, including `lens:` selection.
- ISO date-range validation, domain/file-extension normalization, and duplicate filter normalization.
- Deterministic provider eligibility filtering for Index First, Federated, GoreeCloud Only, External Only, and Offline/Local modes.
- Explicit third-party query-disclosure signal and optional per-query disclosure budgets with plan evidence.
- Content-policy hooks, SafeSearch intent modes, administrator domain controls, and fail-closed enforcement checks.
- Replaceable provider protocol with bounded asynchronous provider execution, timeout/concurrency controls, cancellation, failure isolation, fallback behavior, provenance checks, and explicit availability state.
- Versioned `goreecloud.search-index.v1` Search ↔ Index contract with capability negotiation, cursor pagination, repeated-cursor protection, and degraded/warning propagation.
- Conservative URL canonicalization, canonical-URL/content-hash deduplication, multi-provider source agreement, and provenance aggregation.
- Deterministic Search-owned baseline ranking with inspectable “Why this result?” signals and no behavioral/ad/payment/provider-rank boost.
- Transparent Search-local GoreeCloud Lenses with domain, filetype, and language boost/lower/exclude rules.
- Stable Lens-name lookup, explicit Lens ranking signals/exclusion evidence, and fail-fast unknown Lens handling before provider execution.
- Provider-facing queries strip Search-local Lens selection so provider adapters do not receive the selected Lens name from this layer.
- Development CLI for query parsing without network access.
- Unit tests and CI for Python 3.11 and 3.12.
- Explicit internal application version.

## Not implemented

No authenticated live network provider ships in this candidate. The GoreeCloud Index adapter still has no authenticated runtime transport and no approved external network provider is included. Lens persistence, export/import, sharing, synchronized Lens state, Lens UI, user-facing web UI, Browser integration, AI synthesis, persistent history, and production deployment are not implemented yet.
