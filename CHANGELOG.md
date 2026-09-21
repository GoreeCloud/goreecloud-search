# Changelog

All notable verified changes to GoreeCloud Search will be recorded here.

## Unreleased

### Added

- Initial native query model and parser, privacy-aware source planning, provider contract, development CLI, tests, and CI.
- Versioned Search ↔ GoreeCloud Index v1 contract, capability model, transport-injected adapter, cursor pagination, and degraded/warning propagation.
- Conservative URL canonicalization, canonical-URL/content-hash deduplication, source agreement, and result provenance primitives.
- Deterministic Search-owned ranking with inspectable scoring signals and initial “Why this result?” explanations.
- Bounded asynchronous provider execution with concurrency/timeouts, cancellation propagation, partial-provider failure isolation, availability states, and Index-first fallback.
- End-to-end SearchCore pipeline from parse → plan → execute → normalize/deduplicate → rank.
- Optional per-query third-party disclosure budgets enforced during source planning, before any provider execution.
- Source-plan evidence recording third-party provider count, applied disclosure cap, and omitted providers.
- Cycle-safe Index-originated Search delegation contract and external-only execution path that prohibits recursive Index re-entry and fallback.

No release, deployment, production acceptance, or Stable qualification is claimed by these unreleased changes.
