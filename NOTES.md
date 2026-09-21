# GoreeCloud Search — Repository Notes

## Current development baseline

The repository's authoritative `main` line was initialized on September 16, 2026. The current implementation work remains on stacked development candidates and is not accepted on `main`.

## Design decisions

- Python 3.11+ is used for the initial core because the slice requires no runtime third-party dependencies and can be validated quickly.
- Provider execution is separated behind a protocol so GoreeCloud Index and optional external sources can be integrated without coupling the parser/ranking core to a specific provider.
- The executor is intentionally policy-following rather than provider-selecting: only source-plan-approved providers can run.
- Provider exceptions and timeouts become explicit execution state; caller cancellation propagates and cancels in-flight work.
- Index-first fallback is conditional and cannot bypass planner privacy policy.
- No external provider is silently substituted when a privacy-restrictive source mode is selected.
- This version is Development only.

## Index-originated cycle-safety decision

The dedicated Index-originated path is intentionally not the ordinary Search `INDEX_FIRST` path. It is fixed to external-only planning so an Index request cannot be routed back into the GoreeCloud Index provider or another first-party/local provider and cannot fall back into Index after an external failure.

This is a source-level readiness contract only. Search still has no authenticated live Index-originated HTTP transport on authoritative `main`.

Authoritative `main` now carries the accepted Platform Contract 0.4 nine-system declaration from PR #22. This cycle-safety candidate inherits that control plane without changing its Development/nonconformant status; runtime Platform-System acceptance remains separate and blocked.

## Open decisions

- Final production service/runtime framework.
- Public source licensing/rights model.
- Approved production provider set.
- Exact current Stable Platform-System contract versions at the time each integration is implemented.
- Production deployment topology and persistence model.

## Stacked implementation candidates

`feature/index-contract-normalization` is based on `feature/native-search-core-foundation`; `feature/deterministic-ranking-explanations` is stacked on it; `feature/provider-execution-pipeline` is stacked on the ranking branch; and `feature/query-disclosure-budget` is stacked on the execution branch. Each child must be retargeted/reconciled and revalidated when a parent changes or merges.

## Ranking baseline

The initial ranker intentionally avoids provider-specific hidden boosts and behavioral signals. GoreeCloud Index presence is recorded as provenance but contributes zero score. Source agreement is bounded so federation consensus cannot dominate lexical/query-intent relevance.

## Query-disclosure budget candidate

The budget is enforced in source planning, before execution. It can reduce or eliminate third-party providers but cannot add a provider that the selected source mode, category, source filter, or provider configuration would otherwise exclude. A zero budget prevents external fallback execution rather than contacting an external provider and discarding its response afterward.

## 2026-09-21 — Authenticated Index transport source candidate

A fresh candidate from authoritative Contract 0.4 `main` adds the server-side HTTP half of Index → Search delegation while preserving `goreecloud.search-index-delegation.v1`, `external_only`, no Index re-entry, and no fallback. GoreeCloud Identity and Privacy Shield remain injected producer-authoritative verifier interfaces. Capability evidence is production-shaped but explicitly not production accepted.

No live verifier, approved external provider, deployment, Production acceptance, or Stable qualification is implied.
