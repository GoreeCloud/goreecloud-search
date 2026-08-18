# GoreeCloud Search — Browser Integration Contract

## Status

This document defines the canonical integration boundary between **GoreeCloud Browser** and **GoreeCloud Search**.

GoreeCloud Search is the only and default search engine for GoreeCloud Browser. The browser must treat Search as a GoreeCloud platform service rather than exposing individual upstream search providers as browser-level search engines.

## Canonical service

- Product name: `GoreeCloud Search`
- Canonical private URL: `https://search.goreecloud.com/`
- Search endpoint: `https://search.goreecloud.com/search?q={searchTerms}`
- OpenSearch discovery: provided by GoreeCloud Search through the product page's `rel="search"` metadata
- Access model: private GoreeCloud service, resolved and reached through the approved GoreeCloud private networking and DNS path

## Browser requirements

GoreeCloud Browser must:

1. Configure GoreeCloud Search as its default search engine.
2. Expose GoreeCloud Search as the only user-selectable browser search engine.
3. Route address-bar searches through GoreeCloud Search.
4. Route new-tab and dedicated search-field queries through GoreeCloud Search.
5. Use the GoreeCloud Search product identity, icon, and OpenSearch metadata where browser integration requires search-provider metadata.
6. Never silently fall back to Google, Bing, DuckDuckGo, Brave Search, or another external browser-level search provider when GoreeCloud Search is unavailable.
7. Present GoreeCloud Search unavailability as a GoreeCloud service/connectivity state rather than bypassing the service.
8. Keep upstream engine selection and provider policy behind the GoreeCloud Search boundary.

## Search-provider boundary

Individual engines and upstream providers are implementation details of GoreeCloud Search. GoreeCloud Browser does not directly configure or depend on them.

The architectural flow is:

```text
GoreeCloud Browser
        |
        v
GoreeCloud Search
        |
        v
Configured search engines/providers
```

This separation allows Search to change provider selection, privacy controls, failure handling, ranking, proxy behavior, and future GoreeCloud-native search capabilities without requiring a Browser search-provider migration.

## Privacy and failure behavior

Browser integration must preserve the GoreeCloud Search privacy model. In particular:

- Search queries are sent to GoreeCloud Search, not directly to a browser-vendor search provider.
- Browser-level telemetry must not add search-query reporting to external providers.
- Provider degradation inside GoreeCloud Search must not cause browser-level provider fallback.
- If the private DNS, NetBird path, Caddy route, or Search service is unavailable, Browser should show a clear GoreeCloud Search connectivity/error state.

## Stable-release acceptance

Before the first Stable Search release and the corresponding Browser integration are considered complete, acceptance must verify:

- the OpenSearch metadata identifies `GoreeCloud Search`;
- the canonical search endpoint accepts encoded address-bar queries;
- compact, desktop, and browser-integrated search flows preserve the same product identity;
- no upstream search provider is exposed as the GoreeCloud Browser default or fallback engine;
- provider failures remain contained within GoreeCloud Search;
- the private service remains unavailable through the unintended public DNS path;
- GoreeCloud Browser can recover cleanly when Search becomes reachable again.

## Compatibility terminology

The maintained codebase may retain upstream SearXNG-compatible internal names, environment variables, paths, or APIs while migration work is in progress. Those implementation details must not be presented to GoreeCloud Browser as alternative search products.

After the first Stable cutover, active GoreeCloud Browser and Search documentation must describe **GoreeCloud Search** as the current search service. SearXNG references may remain only where required for upstream provenance, licensing, compatibility, source history, or migration records.
