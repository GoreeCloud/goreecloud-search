# GoreeCloud Index consumer contract

**Lifecycle:** Development contract foundation only. This document does not establish production acceptance, production networking, GoreeCloud Index integration acceptance, Privacy Shield runtime acceptance, Wardveil protected state, Release Candidate status, or Stable status.

## Authority boundary

GoreeCloud Search remains the authoritative GoreeCloud service for Internet, web, and current-information discovery. GoreeCloud Index may delegate Internet discovery to Search; it must not duplicate Search provider integrations or silently bypass Search with direct third-party web providers.

The current first-party machine-readable boundary is:

`GET /api/v1/search`

For the initial Index delegation contract, Index should send only:

- `q` — the user-entered search text after Index's normal surrounding-whitespace normalization;
- `category=general` — the current Internet/general-search category;
- `limit` — an explicit integer from 1 through 100.

The `limit` parameter is additive to API v1. Existing callers that omit it retain the existing response behavior. When supplied, Search returns no more than the requested number of ranked results.

## Capability preflight

Before a future live Index transport treats a configured Search endpoint as compatible, it should obtain Search's machine-readable status from:

`GET /api/v1/status`

The `capability_evidence` collection currently publishes the first-party query capability with:

- `id=search.query`;
- `contract_version=1`;
- `authoritative=true`;
- `current=true`;
- `endpoint=/api/v1/search`;
- `max_results=100`;
- `production_accepted=false` while the native Search line remains Development.

A first-party consumer must fail closed if the required `search.query` capability is absent, is not current/authoritative, declares an unsupported contract version, or identifies an incompatible query endpoint. A consumer must never request more than the published `max_results` bound. If its desired result count exceeds that bound, it must lower the request to the accepted bound rather than assuming Search will accept a larger request.

Capability preflight is compatibility evidence, not authorization or production acceptance. Index must still apply its own provider-contract compatibility, Privacy Shield decision, local-versus-remote policy, user/provider controls, and any future service-discovery/authentication requirements before query dispatch. `production_accepted=false` must not be rewritten as a production-ready state merely because the Development endpoint is reachable.

The capability document does not authorize Index to discover arbitrary Search endpoints, follow provider-controlled URLs, or infer authentication, DNS, TLS, reverse-proxy, or deployment topology. Those remain separately governed transport/deployment concerns.

## Data minimization

Index must not attach unrelated local context to a Search delegation request. In particular, the initial contract does not require or authorize transmission of:

- local provider identifiers;
- local result sets or ranking state;
- file paths, filenames, document contents, contact data, calendar data, application inventory, or device settings;
- local snippets or source metadata;
- query history;
- GoreeCloud Identity identifiers;
- Privacy Shield evidence references;
- Wardveil evidence;
- browser-session state, Vault references, credentials, or secrets.

Authorization and privacy evidence remain local control-plane inputs unless a separately governed contract explicitly requires a minimized value to cross the service boundary.

## Response contract

Index may consume Search API v1's normalized response fields, including:

- `query`;
- `category`;
- ordered `results`;
- each result's `title`, `url`, optional `snippet`, provider attribution, and Search score;
- provider status and degraded-state evidence when needed for diagnostics.

Index should treat Search result order as Search's web-ranking output while still applying Index's own provider normalization, provenance validation, aggregation, deduplication, and final cross-provider ranking rules.

Index must validate that the response corresponds to the delegated query/category before accepting results. Invalid or unsafe result URLs must not become executable actions.

A successful Search response may be partially degraded. In that case, a consumer may preserve valid returned results while carrying forward the degraded state through its own provider-health model. A degraded response must not be silently presented as fully healthy, and degradation must not be reinterpreted as authorization, security, or result-trust evidence.

## Failure behavior

A non-success HTTP response, incompatible API version, incompatible or missing required capability evidence, malformed response, mismatched query/category, timeout, or unavailable Search client must not be represented as successful Internet retrieval. Index should surface the Search provider as unavailable or failed through its own provider-health contract and preserve valid results from unrelated providers.

## Current acceptance boundary

This contract is intentionally narrow. It does not add an Index network transport, service discovery, authentication mechanism, production endpoint configuration, Android Internet permission, live end-to-end acceptance, or production authorization claim. Those require separately reviewed implementation and runtime evidence.
