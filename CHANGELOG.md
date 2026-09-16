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

No release, deployment, production acceptance, or Stable qualification is claimed by these unreleased changes.
