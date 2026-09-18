# Changelog

All notable verified changes to GoreeCloud Search will be recorded here.

## Unreleased

### Added

- Native query model/parser, privacy-aware source planning, provider contract, development CLI, tests, and CI.
- Versioned Search ↔ GoreeCloud Index v1 contract, capability model, transport-injected adapter, cursor pagination, and degraded/warning propagation.
- URL normalization/deduplication, source agreement, and result provenance.
- Deterministic Search-owned ranking with inspectable signals and initial “Why this result?” explanations.
- Bounded asynchronous provider execution with concurrency/timeouts, cancellation, partial-failure isolation, availability states, and Index-first fallback.
- Optional per-query third-party disclosure budgets enforced before provider execution, with plan evidence.
- Content-policy hook engine between normalization and ranking, with decision provenance and allow/warn/block outcomes.
- Off/Moderate/Strict SafeSearch intent plus fail-closed pre-execution enforcement-availability checks.
- Built-in administrator domain allow/block controls with subdomain matching.
- Transparent Search-local GoreeCloud Lens primitives for domain/filetype/language boost, lower, and exclusion rules, including explicit ranking signals and non-disclosure of Lens selection to provider adapters.
- Versioned `goreecloud.search-lens.v1` portable JSON import/export with deterministic round-trip serialization, strict field/version validation, duplicate-key and non-finite-number rejection, and bounded document/rule sizes.
- Project-specific licensing decision selecting `AGPL-3.0-or-later` for GoreeCloud Search, plus repository license and rights notices.
- Bounded query-aware snippet generation for already-authorized plain text.
- Opt-in Brave Web Search API provider and CLI search path with fixed-endpoint HTTPS, environment-only credentials, minimized provider queries, bounded response parsing, redirect refusal, result-URL validation, region/language controls, bounded date-range mapping, and provider-side SafeSearch selection.
- Loopback-only Development HTTP API with health and versioned search endpoints, bounded request parsing, privacy-preserving request logging behavior, no CORS exposure, and defensive response headers.
- Non-root Zorin OS/Linux systemd user-service deployment profile with a protected runtime environment file, source-direct installation, Python 3.10+ runtime gating, service hardening, health-gated startup, and rollback-safe removal that leaves SearXNG unchanged.
- Server-rendered loopback Search UI and OpenSearch discovery with escaped result/query content, source-disclosure state, ranking explanations, no client-side JavaScript or remote UI assets, responsive/adaptive styling, accessibility fallbacks, and stronger browser-facing response isolation. Glaze UI 1.5.1 is the target, but repository-local consumer acceptance remains unresolved.
- Python 3.10 support candidate for the Zorin laptop path, including package metadata, installer runtime gate, and CI matrix coverage alongside Python 3.11 and 3.12.

No release, deployment, production acceptance, or Stable qualification is claimed by these unreleased changes.
