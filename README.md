# GoreeCloud Search

GoreeCloud Search is the privacy-first, self-hostable search and information-discovery service for the GoreeCloud ecosystem.

> **Lifecycle:** Development  
> **Version:** `0.1.0.dev5`  
> **Current scope:** Native query parsing, privacy-aware source planning, bounded provider execution orchestration, a versioned GoreeCloud Index contract boundary with pagination, Search-owned normalization/deduplication, and deterministic explainable ranking. No authenticated live provider transport ships in this revision.

## What exists now

This repository currently contains the first native implementation foundation:

- A typed search-query model.
- Parsing for `site:`, `-domain:`, `filetype:`, `ext:`, `before:`, `after:`, `language:`, `region:`, `source:`, `category:`, and `lens:`.
- Quoted phrases and excluded terms.
- Search categories and deployment source modes.
- A deterministic source planner for Index First, Federated, GoreeCloud Only, External Only, and Offline / Local Index operation.
- A replaceable provider contract for GoreeCloud Index and future federated adapters.
- Bounded asynchronous provider execution with per-provider timeouts, cancellation propagation, concurrency limits, partial-failure isolation, fallback execution, and explicit availability state.
- Optional query-disclosure budgets that cap how many third-party providers may receive one query before execution begins.
- A privacy invariant that prevents external-provider inclusion in GoreeCloud Only and Offline / Local modes.
- Versioned Search ↔ GoreeCloud Index v1 contract models, transport-injected first-party adapter boundary, cursor pagination, and provider-reported degraded-state/warning propagation.
- Conservative URL canonicalization, canonical-URL/content-hash deduplication, source agreement, and result provenance.
- Deterministic ranking with inspectable scoring signals and human-readable result explanations.
- Unit tests and pull-request CI.

The repository ships no authenticated live provider transport. The development CLI performs no network access. The execution engine can invoke explicitly injected provider adapters, so any future network-capable adapter must satisfy the applicable privacy, identity, security, and deployment controls before acceptance.

## Development use

Requires Python 3.11 or newer.

```bash
python -m pip install -e .
goreecloud-search parse 'privacy "search engine" site:example.com category:docs'
```

The command prints the parsed query as JSON. It performs no network access.

Run tests with:

```bash
python -m unittest discover -s tests -v
```

## Architecture direction

The intended separation is:

```text
Browser / AI / API clients
          |
          v
   GoreeCloud Search
   - query parser
   - source planner
   - privacy boundary
   - provider executor
   - normalization
   - deduplication
   - ranking
          |
     +----+----+
     |         |
     v         v
GoreeCloud   optional
  Index      providers
```

The current development candidates implement the query/parser, source-planning, bounded provider execution, Search ↔ Index contract/pagination, normalization/deduplication, and initial deterministic ranking layers. See `FEATURE-ROADMAP.md` for planned work and `FEATURES.md` for current implementation state.

## Cycle-safe Index-originated delegation

Search now has a dedicated source-level `SearchCore.search_from_index(...)` path for a future request that originates from GoreeCloud Index. This path is deliberately different from Search's ordinary `INDEX_FIRST` behavior:

- it always plans with `EXTERNAL_ONLY`;
- every selected step must be a primary `ProviderOrigin.EXTERNAL` provider;
- `GOREECLOUD_INDEX`, `GOREECLOUD_SERVICE`, and `LOCAL` providers are excluded;
- no fallback stage is permitted;
- an Index-only configuration, an Index-directed `source:` filter, or a zero third-party-disclosure budget fails before provider dispatch.

The source contract identifier is `goreecloud.search-index-delegation.v1`, with mode `external_only`, Index-provider re-entry disabled, and fallback disabled.

This prevents the architectural cycle `Index → Search → Index` at the Search planning/execution boundary. It does **not** add an authenticated transport, HTTP endpoint, provider credential, live external provider, Identity registration, Privacy Shield runtime acceptance, deployment, or production approval.

## Privacy boundary

The planning layer distinguishes third-party query disclosure, can apply an explicit per-query third-party-provider disclosure budget, and fails closed for source modes that prohibit disclosure. The execution layer runs only providers already admitted by that plan, applies bounded concurrency/timeouts, and reports degraded or unavailable states instead of silently substituting providers. No authenticated live network adapter is shipped; future network-capable adapters must integrate applicable GoreeCloud Identity, Privacy Shield, Wardveil Security, and other platform controls before production acceptance.

## Status integrity

A branch, pull request, passing CI run, configuration declaration, or documented plan does not mean a feature is released, deployed, production-accepted, or Stable.

## Current implementation expansion

The current stacked development candidate adds the versioned `goreecloud.search-index.v1` contract, a first-party Index provider adapter boundary with pagination, bounded provider execution/fallback behavior, Search-owned result normalization/deduplication, deterministic ranking, and pre-execution third-party query-disclosure budgeting. Authenticated live Index connectivity and runtime Platform-System enforcement remain unimplemented and are not implied by these interfaces.
