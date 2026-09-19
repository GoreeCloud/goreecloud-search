# GoreeCloud Search

GoreeCloud Search is the privacy-first, self-hostable search and information-discovery service for the GoreeCloud ecosystem.

> **Lifecycle:** Development  
> **Version:** `0.1.0.dev3`  
> **Current scope:** Native query parsing, privacy-aware source planning, a versioned GoreeCloud Index contract boundary, Search-owned result normalization/deduplication, and deterministic explainable ranking. Live web retrieval is not implemented in this revision.

## What exists now

This repository currently contains the first native implementation foundation:

- A typed search-query model.
- Parsing for `site:`, `-domain:`, `filetype:`, `ext:`, `before:`, `after:`, `language:`, `region:`, `source:`, `category:`, and `lens:`.
- Quoted phrases and excluded terms.
- Search categories and deployment source modes.
- A deterministic source planner for Index First, Federated, GoreeCloud Only, External Only, and Offline / Local Index operation.
- A replaceable provider contract for future GoreeCloud Index and federated adapters.
- A privacy invariant that prevents external-provider inclusion in GoreeCloud Only and Offline / Local modes.
- Versioned Search ↔ GoreeCloud Index v1 contract models and a transport-injected first-party adapter boundary.
- Conservative URL canonicalization, canonical-URL/content-hash deduplication, source agreement, and result provenance.
- Deterministic ranking with inspectable scoring signals and human-readable result explanations.
- Unit tests and pull-request CI.

The current code does **not** contact external search engines, GoreeCloud Index, GoreeCloud Identity, Privacy Shield, Wardveil Security, Mesh, or any deployed service.

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

The current development candidates implement the query/parser, source-planning, Search ↔ Index contract, normalization/deduplication, and initial deterministic ranking layers. See `FEATURE-ROADMAP.md` for planned work and `FEATURES.md` for current implementation state.

## Privacy boundary

No live provider execution exists in this revision. The planning layer distinguishes third-party query disclosure and fails closed for source modes that prohibit it. Future network-capable adapters must integrate applicable GoreeCloud Identity, Privacy Shield, Wardveil Security, and other platform controls before production acceptance.

## Status integrity

A branch, pull request, passing CI run, configuration declaration, or documented plan does not mean a feature is released, deployed, production-accepted, or Stable.

## Current implementation expansion

The current stacked development candidate adds the versioned `goreecloud.search-index.v1` contract, a first-party Index provider adapter boundary, Search-owned result normalization/deduplication, and deterministic ranking. Live Index connectivity and runtime Platform-System enforcement remain unimplemented and are not implied by these interfaces.
