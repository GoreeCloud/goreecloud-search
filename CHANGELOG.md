# Changelog

All notable verified changes to GoreeCloud Search will be recorded here.

## Unreleased

### Added

- Initial native query model and parser.
- Privacy-aware source planning.
- Replaceable provider contract.
- Development CLI.
- Baseline unit tests and CI.
- Required repository documentation and truthful Platform-System status declaration.
- Versioned Search ↔ GoreeCloud Index v1 contract and capability model.
- First-party GoreeCloud Index adapter boundary with injected transport.
- Conservative URL canonicalization, content-hash/canonical-URL deduplication, source agreement, and result provenance primitives.
- Deterministic Search-owned ranking with inspectable scoring signals and initial “Why this result?” explanations.
- Bounded asynchronous provider execution with configurable concurrency and per-provider timeouts.
- Cancellation propagation and partial-provider failure isolation.
- Explicit available/degraded/unavailable execution state and provider attempt evidence.
- Index-first conditional fallback execution.
- GoreeCloud Index cursor pagination, repeated-cursor protection, and degraded/warning propagation.
- End-to-end SearchCore pipeline from parse → plan → execute → normalize/deduplicate → rank.

No release, deployment, production acceptance, or Stable qualification is claimed by these unreleased changes.
