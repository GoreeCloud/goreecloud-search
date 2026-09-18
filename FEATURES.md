# GoreeCloud Search — Current Features

**Lifecycle:** Development  
**Version:** `0.1.0.dev9`

This file records implemented behavior in the current development candidate. It does not claim release, deployment, production acceptance, or Stable qualification.

## Implemented in the current development candidate

- Native Python package under `src/goreecloud_search`.
- Typed search categories and source modes.
- Query parser for the initial operator set, including `lens:` selection.
- Deterministic provider eligibility filtering and optional per-query disclosure budgets with plan evidence.
- Content-policy hooks, SafeSearch intent modes, administrator domain controls, and fail-closed enforcement checks.
- Replaceable provider protocol with bounded asynchronous execution, timeout/concurrency controls, cancellation, failure isolation, fallback behavior, provenance checks, and explicit availability state.
- Versioned `goreecloud.search-index.v1` Search ↔ Index contract with capability negotiation, cursor pagination, repeated-cursor protection, and degraded/warning propagation.
- Conservative URL canonicalization, canonical-URL/content-hash deduplication, multi-provider source agreement, and provenance aggregation.
- Deterministic Search-owned baseline ranking with inspectable “Why this result?” signals and no behavioral/ad/payment/provider-rank boost.
- Transparent Search-local GoreeCloud Lenses with domain, filetype, and language boost/lower/exclude rules.
- Stable Lens-name lookup, explicit Lens ranking signals/exclusion evidence, and fail-fast unknown Lens handling before provider execution.
- Provider-facing queries strip Search-local Lens selection so adapters do not receive the selected Lens name from this layer.
- Versioned `goreecloud.search-lens.v1` portable JSON representation with deterministic export and strict import validation.
- Portable Lens parser rejects unknown fields/versions, duplicate JSON keys, unsupported rule enums, non-finite values, oversized documents, excessive rules, and invalid weights.
- Bounded query-aware snippet generation for already-authorized plain text.\n- Opt-in Brave Web Search API adapter for General search, with fixed-endpoint HTTPS transport, environment-only credentials, query minimization, region/language mapping, bounded date ranges, provider-side SafeSearch selection, response-size bounds, redirect refusal, and malformed-result degradation.\n- Development CLI supports local parsing and explicit Brave-backed search.
- Unit tests and CI for Python 3.11 and 3.12.
- Explicit internal application version.

## Not implemented

One opt-in external live provider adapter (Brave Web Search API) ships in this Development candidate, but it is not production-accepted. The GoreeCloud Index adapter still has no authenticated runtime transport. Portable Lens serialization is implemented, but file persistence, automatic import/export UI, remote sharing/discovery, signatures/trust, synchronized Lens state, hosted Lens registry, Lens UI, Browser integration, AI synthesis, persistent history, and production deployment are not implemented yet.
