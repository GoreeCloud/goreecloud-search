# GoreeCloud Search — Repository Notes

## Current development baseline

The repository's authoritative `main` line was initialized on September 16, 2026. Current implementation work remains on stacked development candidates and is not accepted on `main`.

## Design decisions

- Python 3.11+ keeps the initial core dependency-light.
- Provider execution is policy-following rather than provider-selecting: only source-plan-approved providers can run.
- Query-disclosure budgets narrow the plan before execution and never expand provider eligibility.
- Provider exceptions/timeouts become explicit execution state; caller cancellation propagates.
- Content policy is a distinct stage after normalization and before ranking.
- Policy hooks are attributable; exceptions or spoofed decision provenance fail closed.
- Moderate/Strict SafeSearch must be enforceable before any provider call or Search rejects the request.
- The built-in domain hook is administrator policy only, not a semantic/safety classifier.
- This version is Development only.

## Open decisions

- Final production service/runtime framework.
- Public source licensing/rights model.
- Approved production provider set.
- Production safety/content classification sources and Wardveil integration.
- Exact current Stable Platform-System contract versions at integration time.
- Production deployment topology and persistence model.

## Stacked implementation candidates

The current stack is `feature/native-search-core-foundation` → `feature/index-contract-normalization` → `feature/deterministic-ranking-explanations` → `feature/provider-execution-pipeline` → `feature/query-disclosure-budget` → `feature/content-policy-hooks`. Each child must be retargeted/reconciled and revalidated when a parent changes or merges.
