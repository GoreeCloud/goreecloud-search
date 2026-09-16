# GoreeCloud Search — Repository Specifications

## Lifecycle

- Product: GoreeCloud Search
- Version: `0.1.0.dev6`
- Lifecycle: Development
- Stable: No

## Current implementation boundary

The current stacked candidates provide query parsing/planning, pre-execution third-party disclosure budgeting, bounded provider execution, a versioned Search ↔ Index contract with pagination, Search-owned normalization/deduplication, local content-policy hooks, and deterministic ranking.

### Source/privacy boundary

The planner implements Index First, Federated, GoreeCloud Only, External Only, and Offline Local modes. It may also apply `QueryDisclosureBudget(max_third_party_providers=N)` before returning the plan. The executor may run only plan-approved providers and cannot widen source eligibility.

### Provider boundary

Provider adapters remain replaceable and declare category/origin. Execution is bounded by concurrency and timeout policy; cancellation propagates; provider failures are isolated; provider batches may report degradation/warnings.

### Content-policy boundary

Content policy is evaluated after normalization and before ranking. Hooks return attributable allow/warn/block decisions. Block dominates warn. Hook exceptions or spoofed decision provenance fail closed instead of silently bypassing policy.

`DomainPolicyHook` enforces explicit administrator domain allowlists/blocklists and intentionally does not claim semantic content classification. Search exposes Off, Moderate, and Strict SafeSearch intent; Moderate/Strict is rejected before any provider execution unless at least one configured hook declares SafeSearch enforcement. The hook contract is not evidence of a production classifier, Wardveil integration, or accepted safety taxonomy.

### Ranking boundary

Ranking uses transparent deterministic query/result evidence plus a bounded source-agreement bonus. GoreeCloud Index provenance has zero ranking weight. Provider rank, click history, advertising payments, cross-query profiles, and hidden behavioral signals are not used.

## Not yet implemented

- Authenticated live GoreeCloud Index transport and runtime integration.
- Approved external provider adapters.
- Production SafeSearch classification sources / Wardveil safety integration.
- HTTP API and health/readiness endpoints.
- Snippet generation and advanced ranking signals beyond the current baseline.
- Private View, GoreeCloud Browser integration, GoreeCloud AI answers.
- Identity-authenticated requests, Privacy Shield runtime authorization, GoreeCloud Mesh discovery/events, GoreeCloud Manager administration.
- User/admin Glaze UI surfaces, persistent history/Sync, and production deployment artifacts.

These remain planned and must not be represented as implemented until code and verification evidence exist.
