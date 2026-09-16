# GoreeCloud Search — Privacy Policy

## Current development implementation

Version `0.1.0.dev8` contains a local query parser, source planner, bounded provider-execution engine, versioned Search ↔ Index contract models with pagination, a transport-injected Index adapter boundary, result normalization/deduplication, content-policy hooks, deterministic ranking, Search-local Lens reranking, and a portable Lens data format. The repository code in this version:

- Ships no authenticated live network transport or approved external network provider. The development CLI performs no network access.
- Executes explicitly injected provider adapters only within the planner-approved source plan.
- Supports optional per-query disclosure budgets before execution.
- Rejects Moderate/Strict SafeSearch before provider execution when no configured hook can enforce it.
- Resolves a requested Lens before provider execution and removes the `lens:` selection from the provider-facing parsed query.
- Applies Lens boost/lower/exclude rules only after normalized results pass content policy and baseline ranking.
- Can serialize/parse Lens configuration as deterministic JSON locally; the implementation performs no file write, upload, publication, synchronization, or remote retrieval of Lens documents.
- Does not persist search history or Lens state.
- Does not contain advertising/tracking code, build behavioral profiles, or use click history/advertising identifiers/paid placement for ranking.

## Portable Lens privacy

Portable Lens documents intentionally contain only format version, user-visible Lens name/description, and explicit rules. The v1 schema rejects unknown fields rather than silently carrying arbitrary identifiers or hidden metadata. A Lens document can still reveal user preferences through its rules if a user chooses to share it, so future sharing/synchronization workflows must provide clear disclosure and consent boundaries.

## Future network features

Any live network provider transport, remote suggestion, Lens sharing/discovery/synchronization, account storage, analytics, AI, Browser, Index, or other network-capable feature must update this policy to match verified behavior and apply the relevant GoreeCloud Identity, Privacy Shield, Wardveil Security, retention, consent, minimization, and disclosure controls before production acceptance.

## Development warning

This document describes the current repository implementation. It does not claim that a future hosted deployment has the same data flows until that deployment is separately verified.
