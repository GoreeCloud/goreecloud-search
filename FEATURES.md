# Features

## Purpose

I use this record to maintain a current, product-level inventory of GoreeCloud Search capabilities. I distinguish features that are already part of the active product from first-Stable Release Candidate capabilities and future work that remains unaccepted.

A feature appearing here does not silently promote a Release Candidate to Stable. Release lifecycle and runtime acceptance remain governed separately.

## Active product capabilities

### Private self-hosted metasearch

- GoreeCloud-controlled SearXNG-derived metasearch service.
- Aggregation across approved external search providers instead of dependence on one direct Browser search engine.
- General web search and supported specialized search categories.
- Clear engine and result-source visibility.
- Failure isolation when individual providers are unavailable or degraded where the underlying engine supports it.

### Search categories and controls

- General web search.
- Image search.
- Video search.
- News search.
- Technical and documentation-oriented discovery.
- Software/package and research-oriented discovery where enabled providers support it.
- Search-language, time-range, safe-search, category, and engine controls inherited from the maintained search foundation where applicable.
- Representative provider-acceptance tooling for General, Images, Videos, News, Files, IT, and Science coverage, with the first five treated as first-Stable mandatory categories in the current release-candidate evidence model.

### GoreeCloud product experience

- GoreeCloud Search identity rather than an upstream-only product shell.
- Glaze UI treatment across the home page, search header, results, result cards, sidebars, pagination, Preferences, About, empty states, errors, and recovery surfaces.
- Compact, Medium, Expanded, and Wide adaptive layouts.
- Light and dark presentation.
- Keyboard navigation and visible focus behavior.
- Reduced-motion support.
- Reduced-transparency support and solid no-blur fallbacks.
- Increased-contrast and forced-colors resilience.
- Mobile-friendly interaction targets and result presentation.
- GoreeCloud-owned favicon, web-app manifest identity, and OpenSearch provider metadata.

### Privacy-first defaults

- No GoreeCloud advertising or sponsored-result business model.
- No intended behavioral user profiling.
- Minimal persistent user-search state.
- Image proxying in the controlled runtime baseline.
- Query text excluded from browser page titles in the controlled runtime baseline.
- `noindex, nofollow` search-service behavior.
- `no-referrer` behavior.
- Frame-embedding denial.
- Camera, microphone, and geolocation browser capabilities disabled for the Search surface.
- Public-instance behavior and SearXNG usage metrics disabled in the GoreeCloud runtime baseline.
- HTML-only response format by default while machine-readable integration access remains separately governed.

### Browser and client integration

- OpenSearch integration for compatible browsers.
- GoreeCloud Browser policy direction that makes GoreeCloud Search the sole/default managed Browser search authority on the current acceptance line.
- Candidate-bound Browser evidence tooling that keeps normal Search policy separate from acceptance-only loopback staging.
- Search outage behavior designed so direct URL navigation can remain available while Browser searches do not silently fall back to an unrelated external search provider.

### Operational and deployment capabilities

- Docker and Docker Compose deployment model.
- GoreeCloud container image build and runtime validation.
- Health endpoint validation.
- Loopback/private-publication-oriented deployment controls.
- Persistent supporting cache/runtime integration through Valkey in the current production architecture.
- Controlled configuration examples with secret separation.
- Bounded container logging in the GoreeCloud Compose example.
- Target-runtime evidence tooling.
- Provider-acceptance evidence tooling.
- Provider-degradation evidence tooling.
- Backup, recovery, and rollback evidence contracts.
- Exact source, image, artifact, and release-evidence identity controls.

### Maintained-fork governance

- Preserved SearXNG provenance and AGPL licensing obligations.
- Recorded exact initial upstream baseline.
- Controlled upstream-update policy.
- GoreeCloud-specific contribution and security-reporting boundaries.
- Automated checks for product identity, privacy behavior, Glaze UI contracts, provider contracts, deployment configuration, source provenance, and licensing preservation.
- Retained upstream integration testing rather than bypassing inherited compatibility checks.

## First-Stable Release Candidate capabilities

The current first-Stable line is Release Candidate material and remains separate from Stable promotion.

Candidate #07 adds or binds the following first-Stable capabilities and acceptance controls:

- Glaze UI 1.1 first-Stable source-conformance work.
- Exact immutable candidate publication and registry retrieval.
- Candidate-bound release evidence.
- Exact-image Compact/Expanded light/dark visual evidence.
- Deterministic GoreeCloud result-treatment and ranking acceptance contracts on the stabilization line.
- Provider-suite and provider-degradation acceptance contracts.
- Target-runtime identity evidence.
- Recovery and rollback evidence contracts.
- Actual GoreeCloud Browser runtime-evidence contract.
- Final-acceptance schema version 2 that cryptographically binds exactly six completed JSON artifacts: release, target runtime, recovery, provider, visual/device review, and actual GoreeCloud Browser runtime evidence.
- Fail-closed rejection of mismatched candidate identities, mutated review evidence, incomplete physical-device/desktop/persisted-theme/Browser acceptance, and production-cutover self-authorization.

These controls improve release integrity but do not independently authorize production cutover or Stable promotion.

## Planned or not-yet-accepted capabilities

The following remain future or incomplete product capabilities and must not be represented as fully accepted merely because related scaffolding exists:

- A stable GoreeCloud-facing machine-readable Search API for approved applications and AI/research consumers.
- JSON, RSS, CSV, or other machine-readable response formats enabled as a general product contract.
- Account-based search history, saved searches, synchronization, or personalization.
- Production-approved local-AI and research-agent integration using the governed Search API.
- Completed actual GoreeCloud Browser runtime acceptance against the exact first-Stable candidate.
- Completed physical Android Preferences acceptance.
- Completed desktop runtime/regression acceptance.
- Completed persisted theme-preference acceptance.
- Completed target-host provider, monitoring, alert-delivery, restore, and rollback evidence for the exact first-Stable candidate.
- Stable lifecycle promotion for the current first-Stable candidate.
- A fully GoreeCloud-native search backend replacing SearXNG.

## Feature-governance rule

I add a capability to the accepted product inventory only to the level supported by source, release, runtime, and lifecycle evidence. Planning, source scaffolding, CI success, candidate publication, production deployment, production acceptance, and Stable promotion remain distinct states.
