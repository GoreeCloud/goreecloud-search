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
- The Brave Web Search adapter is an opt-in Development path; its presence is not production provider acceptance.
- The local HTTP API is fixed to IPv4 loopback and is a Development boundary only; production authentication, abuse controls, readiness, UI, and deployment acceptance remain open.
- The verified GoreeCloud hardware inventory identifies the current laptop as a Lenovo running Zorin OS Linux. The repository now contains a matching systemd user-service Development profile, but no actual laptop deployment or SearXNG retirement has been verified.
- This version is Development only.

## Open decisions

- Final production service/runtime framework.
- Public source licensing/rights model.
- Production acceptance criteria and provider set beyond the Development Brave adapter.
- Production safety/content classification sources and Wardveil integration.
- Exact current Stable Platform-System contract versions at integration time.
- Production deployment topology and persistence model.

## Stacked implementation candidates

The current stack continues through the Lens, snippet, and external-provider candidates: `feature/native-search-core-foundation` → `feature/index-contract-normalization` → `feature/deterministic-ranking-explanations` → `feature/provider-execution-pipeline` → `feature/query-disclosure-budget` → `feature/content-policy-hooks` → licensing/Lens candidates → `feature/snippet-generation` → `feature/brave-web-provider` → `feature/local-http-api` → `feature/zorin-user-service`. Each child must be retargeted/reconciled and revalidated when a parent changes or merges.
