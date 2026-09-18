# GoreeCloud Search — Current Features

**Lifecycle:** Development  
**Version:** `0.1.0.dev11`

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
- Bounded query-aware snippet generation for already-authorized plain text.
- Opt-in Brave Web Search API adapter for General search, with fixed-endpoint HTTPS transport, environment-only credentials, query minimization, region/language mapping, bounded date ranges, provider-side SafeSearch selection, response-size bounds, redirect refusal, and malformed-result degradation.
- Development CLI supports local parsing, explicit Brave-backed search, a loopback-only local service command, and a distinct container-network service command with a bounded Host-header allowlist.
- Development HTTP API exposes `/healthz` and versioned `/api/v1/search`. The local command remains fixed to IPv4 loopback; the container command binds inside the container network for Caddy access and accepts only loopback plus the configured service hostname. Both omit CORS authorization, use `no-store`/anti-sniffing/frame/referrer headers, bound request/query/limit handling, return generic internal errors, and suppress request-target logging.
- VPS Docker candidate includes a non-root image, read-only-root compatible runtime, built-in health check, runtime file secret support, no host-port requirement, capability drop/no-new-privileges Compose controls, Caddy-compatible `searxng-core:8080` naming, and immutable-image-reference enforcement in the example Compose service.
- Unit tests and CI for Python 3.11 and 3.12.
- Explicit internal application version.

## Not implemented

One opt-in external live provider adapter (Brave Web Search API) ships in this Development candidate, but it is not production-accepted. The GoreeCloud Index adapter still has no authenticated runtime transport. Portable Lens serialization, a loopback Development HTTP API, and a VPS Docker deployment candidate are implemented, but production service authentication/authorization, rate/abuse controls, readiness semantics, published immutable artifact acceptance, live VPS provider validation, file persistence, automatic import/export UI, remote sharing/discovery, signatures/trust, synchronized Lens state, hosted Lens registry, Lens UI, Browser integration, AI synthesis, persistent history, and verified production deployment are not implemented yet.
