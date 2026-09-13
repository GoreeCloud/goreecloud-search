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

## Failure behavior

A non-success HTTP response, incompatible API version, malformed response, mismatched query/category, timeout, or unavailable Search client must not be represented as successful Internet retrieval. Index should surface the Search provider as unavailable or failed through its own provider-health contract and preserve valid results from unrelated providers.

## Current acceptance boundary

This contract is intentionally narrow. It does not add an Index network transport, service discovery, authentication mechanism, production endpoint configuration, Android Internet permission, live end-to-end acceptance, or production authorization claim. Those require separately reviewed implementation and runtime evidence.
