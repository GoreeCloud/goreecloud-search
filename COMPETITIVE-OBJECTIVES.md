# GoreeCloud Search Competitive Objectives

## Purpose

GoreeCloud Search benchmarks mature search and metasearch products while remaining a distinct GoreeCloud-owned, privacy-first application. Benchmarks guide capability and quality targets; they do not become product dependencies or permanent architectural authorities.

## Benchmark products and projects

Current useful benchmarks include:

- **SearXNG** — retained as a transitional compatibility/migration reference for mature metasearch aggregation, category support, and self-hosted operation.
- **DuckDuckGo** — approachable privacy-oriented general-search UX.
- **Brave Search** — independent search infrastructure and modern result presentation.
- **Startpage** — privacy-oriented access to mainstream results.
- **Kagi** — result-quality controls, filtering, and user-oriented workflows.
- **Whoogle** — simple self-hosted privacy mediation around external search results.

The target GoreeCloud architecture is native. SearXNG is no longer the intended permanent backend/product architecture even while inherited source remains during migration.

## Capabilities to match

GoreeCloud Search should meet mature-search expectations for:

- fast query submission and clear results;
- useful general web and approved specialized search categories;
- source/provider visibility;
- language, region, safe-search, time-range, and provider controls where supported by approved adapters;
- keyboard, touch, mobile, and desktop usability;
- accessible light/dark Glaze UI behavior;
- Browser/OpenSearch integration;
- bounded provider failures and partial-result behavior;
- understandable preferences and diagnostics that do not expose provider secrets or inherited implementation details.

## Capabilities to exceed

Where practical, GoreeCloud Search should exceed mainstream hosted search in:

- **Privacy by default** — no GoreeCloud ads, sponsored ranking, or behavioral-profiling business model.
- **Native product ownership** — GoreeCloud owns search orchestration, provider contracts, preferences, presentation, platform integration, migration, and release state.
- **Provider transparency** — provider identity and status are explicit, bounded, and sanitized.
- **Failure honesty** — degraded providers are represented rather than hidden behind unrelated fallback behavior.
- **Inspectability** — deterministic result treatment and source-visible behavior instead of paid/click-tracking influence.
- **Security of result metadata** — malformed provider names, raw provider errors, and credential-bearing URLs must not leak into normal result evidence/presentation.
- **Recovery and portability** — source, configuration, provider policy, migration, backup/restore, rollback, and exact-release evidence are first-class requirements.
- **GoreeCloud integration** — Search remains a stable product boundary for Browser and approved future first-party consumers without coupling them to inherited SearXNG internals.

## Capabilities intentionally rejected

GoreeCloud Search does not intend to adopt:

- advertising or sponsored-result placement;
- paid ranking priority;
- behavioral profiling or cross-site tracking for search personalization;
- mandatory hosted accounts for ordinary search use;
- hidden external fallback engines that bypass the configured Search authority;
- forced AI summaries that obscure source provenance or replace ordinary results by default;
- unnecessary telemetry/query retention;
- dark patterns around privacy controls;
- permanent architectural dependence on one external search provider or on the inherited transitional runtime.

## Native migration objective

Every user-facing inherited capability needed by the selected release must be classified as retain, replace, improve, or explicitly approved retire. New application-defining functionality should be native unless a documented temporary compatibility requirement makes transitional work necessary.

The migration is complete only when native feature parity for the release, user migration, recovery/rollback, platform-system integration, representative runtime/device evidence, and production acceptance have been completed. Source migration percentage alone is not a completion metric.

## Privacy and security objectives

Privacy Shield remains the data-use/minimization authority and Wardveil Security remains the security-acceptance authority. Provider credentials stay outside source; raw provider errors must not become user diagnostics; result destinations are validated before presentation; and administrative/provider integration must be explicitly governed.

External providers may still observe requests from GoreeCloud infrastructure. Search improves privacy boundaries and operational control but does not claim anonymity from those providers.

## User-experience objectives

GoreeCloud-owned Search surfaces should be focused, quick to understand, adaptive, keyboard/touch friendly, and compliant with the latest approved Stable Glaze UI contract. Production acceptance includes applicable focus, text scaling, reduced motion/transparency, contrast, forced colors, responsive layout, and representative device/runtime evidence.

## Performance and reliability objectives

Search should use bounded provider execution, predictable resource limits, deterministic aggregation, explicit degraded status, health/readiness diagnostics, monitoring, and safe failure behavior. Production limits and SLOs require target-runtime evidence rather than source assumptions.

## Long-term differentiator

The defining objective is a GoreeCloud-owned discovery application whose providers and implementation foundations are replaceable. The product should preserve privacy, source transparency, deterministic behavior, platform integration, continuity, and administrative control even as external providers or transitional dependencies change.
