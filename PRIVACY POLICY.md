# GoreeCloud Search — Privacy Policy

## Current development implementation

Version `0.1.0.dev10` contains a local query parser, source planner, bounded provider-execution engine, versioned Search ↔ Index contract models with pagination, a transport-injected Index adapter boundary, result normalization/deduplication, content-policy hooks, deterministic ranking, Search-local Lens reranking, and a portable Lens data format. The repository code in this version:

- Ships one opt-in Development external provider adapter for Brave Web Search API. When explicitly selected, the minimized provider-facing query is disclosed to Brave; Search-local `lens:` and `source:` controls are not included in that provider query string.
- Executes explicitly injected provider adapters only within the planner-approved source plan. The Brave API credential is supplied from `BRAVE_SEARCH_API_KEY`, is not placed in query parameters or result evidence, and is sent only to the fixed Brave HTTPS endpoint; redirects are refused.
- Supports optional per-query disclosure budgets before execution.
- Rejects Moderate/Strict SafeSearch before provider execution when no configured hook can enforce it.
- Resolves a requested Lens before provider execution and removes the `lens:` selection from the provider-facing parsed query.
- Applies Lens boost/lower/exclude rules only after normalized results pass content policy and baseline ranking.
- Can serialize/parse Lens configuration as deterministic JSON locally; the implementation performs no file write, upload, publication, synchronization, or remote retrieval of Lens documents.
- Does not persist search history or Lens state.
- The Development HTTP API binds only to `127.0.0.1`, emits no CORS allow-origin header, sends `Cache-Control: no-store`, and suppresses default HTTP request-target logging so query strings are not written by this server's request logger.
- Does not contain advertising/tracking code, build behavioral profiles, or use click history/advertising identifiers/paid placement for ranking.

## Portable Lens privacy

Portable Lens documents intentionally contain only format version, user-visible Lens name/description, and explicit rules. The v1 schema rejects unknown fields rather than silently carrying arbitrary identifiers or hidden metadata. A Lens document can still reveal user preferences through its rules if a user chooses to share it, so future sharing/synchronization workflows must provide clear disclosure and consent boundaries.

The Brave adapter does not persist queries, API responses, or search history. External-provider operation remains subject to the provider's own service and privacy terms and is not equivalent to GoreeCloud-only search.

The Development HTTP search endpoint uses GET query parameters. A browser, local client, intermediary debugging tool, or operating-system component outside this server may retain requested URLs according to its own settings. The server's `no-store` response and suppressed request-target logging do not control those external histories. A production Browser integration must define the final history/private-search behavior explicitly.

## Future network features

Any live network provider transport, remote suggestion, Lens sharing/discovery/synchronization, account storage, analytics, AI, Browser, Index, or other network-capable feature must update this policy to match verified behavior and apply the relevant GoreeCloud Identity, Privacy Shield, Wardveil Security, retention, consent, minimization, and disclosure controls before production acceptance.

## Development warning

This document describes the current repository implementation. It does not claim that a future hosted deployment has the same data flows until that deployment is separately verified.
