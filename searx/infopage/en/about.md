# About GoreeCloud Search

GoreeCloud Search is a privacy-first metasearch service maintained as part of the GoreeCloud platform. It sends a search to selected external search providers, combines the returned results, and presents them through a GoreeCloud-controlled interface without requiring a search-provider account.

The service is built from the open-source [SearXNG] project. GoreeCloud maintains its own product identity, interface, deployment defaults, validation, and operational documentation while preserving SearXNG attribution and the terms of the upstream license.

## Privacy model

GoreeCloud Search is designed to reduce unnecessary disclosure during ordinary web search. The application does not need to build a personal search profile in order to provide results, and GoreeCloud-specific defaults are intended to minimize avoidable tracking-oriented behavior.

Search results still come from external providers. Those providers remain independent services with their own availability, content, policies, rate limits, and network visibility. GoreeCloud Search cannot make an external provider private simply by querying it on your behalf.

## Search-provider choice

The available providers and categories are visible in {{link('Preferences', 'preferences')}}. Provider availability can change, and an individual engine may temporarily fail, throttle requests, or return no usable results. GoreeCloud Search is designed so that the metasearch experience does not depend on one provider being permanently available.

The {{link('stats page', 'stats')}} provides engine-performance information from the running instance.

## Browser integration

GoreeCloud Search supports [OpenSearch], allowing compatible browsers to add the service as a search provider. Browser behavior differs by platform, so adding or selecting a default search provider may require browser-specific steps.

## Open-source foundation

GoreeCloud Search remains open source and traceable to its upstream foundation. SearXNG provides the metasearch engine, provider integrations, and much of the underlying application architecture. GoreeCloud-specific work focuses on product identity, Glaze UI integration, privacy-oriented defaults, deployment hardening, testing, and operation within the GoreeCloud environment.

For upstream SearXNG documentation, development, translations, and community information, use the official [SearXNG project] resources.

## Service boundary

GoreeCloud Search is a search interface, not an authoritative archive of the web. Results can change or disappear as external sites and providers change. Important information should be preserved separately when long-term retention matters.

[SearXNG]: https://github.com/searxng/searxng
[SearXNG project]: https://docs.searxng.org/
[OpenSearch]: https://github.com/dewitt/opensearch/blob/master/opensearch-1-1-draft-6.md
