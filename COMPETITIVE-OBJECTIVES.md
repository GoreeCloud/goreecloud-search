# Competitive Objectives

## Purpose

I use this record to define the products, projects, and search experiences that GoreeCloud Search benchmarks while preserving its own privacy-first, self-hosted product direction.

GoreeCloud Search is not intended to copy one competitor. I use mature search and metasearch products to identify capabilities worth matching, weaknesses worth avoiding, and areas where a GoreeCloud-controlled service can provide stronger privacy, transparency, ownership, resilience, and administrative control.

## Primary benchmark products

I primarily benchmark:

- **SearXNG** — the current upstream metasearch foundation and the baseline for multi-engine aggregation, category support, engine configuration, and self-hosted operation.
- **DuckDuckGo** — a benchmark for approachable privacy-oriented general search and low-friction user experience.
- **Brave Search** — a benchmark for independent search infrastructure, modern search presentation, and broad web-search usability.
- **Startpage** — a benchmark for privacy-oriented access to mainstream search results.
- **Kagi** — a benchmark for search-quality controls, result presentation, filtering, and user-focused search workflows.
- **Whoogle** — an adjacent benchmark for simple self-hosted privacy mediation around external search results.

I may add or remove benchmark products when the search landscape changes. A benchmark does not become a dependency merely because I study it.

## Capabilities I intend to match

GoreeCloud Search should meet mature-search expectations for:

- Fast, clear query submission and result presentation.
- General web, Images, Videos, News, Files, technical, software, and research-oriented search categories where supported.
- Useful source and engine visibility.
- Search-language, region, safe-search, time-range, and engine preference controls where underlying providers support them.
- Keyboard-first interaction and strong mobile usability.
- Light and dark themes with accessible focus, contrast, reduced-motion, reduced-transparency, and forced-colors behavior.
- Browser/OpenSearch integration.
- Graceful handling of provider failures and partial-result conditions.
- Clear preferences and diagnostics without forcing ordinary users to understand SearXNG internals.

## Capabilities I intend to exceed

Where practical, GoreeCloud Search should exceed mainstream hosted search experiences in:

- **Privacy by default** — no advertising business model, no sponsored ranking, no behavioral profiling, no sale of user data, and no intentional query-history productization.
- **Self-hosted ownership** — the application, configuration, release history, deployment model, and operational evidence remain under GoreeCloud control.
- **Provider transparency** — result sources and engine behavior should be understandable rather than hidden behind an opaque single-provider boundary.
- **Administrative control** — approved providers, categories, privacy behavior, integration access, operational health, and future API consumers should remain explicitly governable.
- **Failure honesty** — provider outages, degraded results, and partial coverage should be represented as operational conditions instead of silently falling back to an unrelated search authority.
- **Inspectability** — GoreeCloud-specific ranking and result treatment should remain deterministic, reviewable, testable, and free from click-tracking or paid-placement influence.
- **Recovery and portability** — source, configuration, deployment documentation, immutable artifacts, rollback targets, and recovery procedures should be sufficient to rebuild the service independently.
- **GoreeCloud integration** — Search should provide a stable product boundary for GoreeCloud Browser, approved AI/research workflows, Manager, and future first-party consumers without requiring them to depend directly on unstable SearXNG internals.
- **Operational evidence** — release, runtime, provider, monitoring, recovery, visual, Browser, and rollback evidence should be treated as part of product quality rather than afterthoughts.

## Capabilities I intentionally reject

I do not intend GoreeCloud Search to adopt:

- Advertising or sponsored-result placement.
- Paid ranking priority.
- Behavioral user profiling.
- Cross-site tracking for search personalization.
- Mandatory cloud accounts for ordinary search use.
- Hidden external fallback engines that bypass the configured GoreeCloud Search authority.
- Forced AI-generated summaries that obscure source provenance or replace ordinary results by default.
- Unnecessary telemetry or query retention.
- Dark patterns that make privacy controls difficult to use.
- Product decisions that require permanent dependence on one external search provider.

## Privacy and security objectives

I require GoreeCloud Search to minimize retained query data, avoid user profiling, protect secrets, separate provider credentials from source, use approved private publication paths, preserve privacy-oriented HTTP behavior, and keep administrative capabilities appropriately restricted.

External search providers may still observe requests originating from GoreeCloud infrastructure. GoreeCloud Search therefore improves control and privacy boundaries but does not claim to make external web search anonymous.

## User-experience and accessibility objectives

I use Glaze UI as the GoreeCloud design language. Search should remain focused, fast to understand, mobile-friendly, keyboard-friendly, readable in light and dark themes, and resilient under reduced-motion, reduced-transparency, high-contrast, and forced-color settings.

The interface should feel like a deliberate GoreeCloud product rather than a lightly recolored upstream deployment.

## Performance and reliability objectives

I prioritize predictable responsiveness, bounded provider failure, healthy container/runtime behavior, clear readiness diagnostics, and resilience when individual external providers fail or throttle requests.

Search quality must not depend on hidden personalization, click tracking, sponsored placement, or remote AI reranking.

## Interoperability and administrative-control objectives

GoreeCloud Search should preserve:

- Browser and OpenSearch compatibility.
- A governed future GoreeCloud-facing API.
- Search-provider and category controls.
- Health/readiness behavior.
- Monitoring and alerting integration.
- Versioned or otherwise stable consumer contracts.
- Replaceability of the SearXNG backend when a future architecture materially improves the product.

## Data portability and recovery objectives

I require enough repository, build, configuration, deployment, artifact, monitoring, backup, restore, and rollback information to recover GoreeCloud Search even if its current upstream foundation later disappears.

## GoreeCloud differentiators

The defining competitive objective is not simply better search results. GoreeCloud Search should combine mature metasearch with:

- Direct ownership.
- Privacy-first defaults.
- No ads or sponsorships.
- Transparent provider behavior.
- Deterministic GoreeCloud result treatment.
- Glaze UI.
- GoreeCloud Browser integration.
- Future GoreeCloud API and AI/research integration.
- Evidence-driven release engineering.
- Documented backup, recovery, and rollback.
- A replaceable upstream foundation rather than permanent vendor dependence.
