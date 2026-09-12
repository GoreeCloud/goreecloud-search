# GoreeCloud Search Native Provider Runtime

## Status

This document describes the native provider runtime implemented for GoreeCloud Search Development source. It is not production-provider approval, provider-selection authority, deployment evidence, or Stable acceptance.

The native application remains pre-Stable. No provider is enabled by default, no credential is committed to source, and the shipped runtime continues to report `production_approved=false`.

## Purpose

The provider runtime gives the GoreeCloud-owned Search engine a deployment-controlled executable provider path without making the native application depend on SearXNG product architecture or exposing provider credentials, endpoints, or raw failure details to users.

It implements narrow GoreeCloud provider contracts that can be used by separately reviewed provider adapters. Provider-specific privacy, terms, rate limits, authentication, timestamp authority, live behavior, and production acceptance remain separate requirements.

## Provider contract v1

The built-in `goreecloud-http-v1` adapter sends an HTTPS `POST` request with JSON:

```json
{
  "schema_version": 1,
  "query": "example query",
  "category": "general"
}
```

The provider endpoint returns HTTP 200 with `application/json` (or a `+json` subtype):

```json
{
  "schema_version": 1,
  "results": [
    {
      "title": "Example result",
      "url": "https://example.org/result",
      "snippet": "Optional summary",
      "score": 0,
      "published_at": "2026-09-04T18:00:00Z",
      "media": {
        "kind": "image",
        "thumbnail_url": "https://cdn.example.org/thumb.jpg",
        "content_url": "https://cdn.example.org/full.jpg",
        "mime_type": "image/jpeg",
        "width": 1200,
        "height": 800,
        "alt": "Optional description"
      }
    }
  ]
}
```

The native Search engine remains authoritative for provider attribution, result URL sanitization, media normalization, timestamp-authority enforcement, result clustering, ranking, source agreement, freshness, and user-facing provider status.

## TinyFish Search adapter v1

The native Development source also supports the provider-specific adapter `tinyfish-search-v1`. This adapter is deliberately narrower than the generic provider contract and binds directly to TinyFish's documented Search API rather than requiring an intermediate translation service.

The adapter:

- permits only the official `https://api.search.tinyfish.ai` endpoint;
- authenticates only with the `X-API-Key` header from a deployment-controlled environment variable;
- supports GoreeCloud `general` through TinyFish `domain_type=web`;
- supports GoreeCloud `news` through TinyFish `domain_type=news`;
- does not claim Images, Videos, or Files support;
- maps TinyFish result position into a bounded provider-score hint while leaving final ranking to GoreeCloud Search;
- ignores TinyFish result dates for publication-time authority because timestamp authority has not been approved for this adapter;
- preserves GoreeCloud Search URL sanitization, result bounds, deduplication, ranking, provider degradation, and presentation authority;
- treats unknown TinyFish response fields as forward-compatible external metadata while consuming only the documented fields required by this adapter.

Example Development configuration:

```json
{
  "schema_version": 1,
  "providers": [
    {
      "name": "TinyFish Search",
      "adapter": "tinyfish-search-v1",
      "endpoint": "https://api.search.tinyfish.ai",
      "categories": ["general", "news"],
      "credential_env": "TINYFISH_API_KEY"
    }
  ]
}
```

The environment value is never stored in this configuration. A missing `credential_env`, missing API key, non-official endpoint, unsupported category, or attempt to mark TinyFish timestamps authoritative fails closed during provider configuration.

This source adapter is Development capability only. It does not by itself select TinyFish as a production provider, install a production credential, authorize production query egress, establish Privacy Shield authorization, establish Wardveil protected state, authorize production deployment, or satisfy Stable qualification.

## Deployment configuration

Provider configuration is disabled unless `GOREECLOUD_SEARCH_PROVIDER_CONFIG_FILE` points to an explicitly supplied JSON file. The configuration is schema-versioned and fail-closed.

Example with no credential for the generic adapter:

```json
{
  "schema_version": 1,
  "providers": [
    {
      "name": "Example Search Adapter",
      "adapter": "goreecloud-http-v1",
      "endpoint": "https://provider.example.org/search",
      "categories": ["general", "images"],
      "published_at_authoritative": false
    }
  ]
}
```

A generic provider that requires a bearer credential references only the environment-variable name:

```json
{
  "schema_version": 1,
  "providers": [
    {
      "name": "Authenticated Adapter",
      "adapter": "goreecloud-http-v1",
      "endpoint": "https://provider.example.org/search",
      "categories": ["general"],
      "credential_env": "GOREECLOUD_SEARCH_EXAMPLE_PROVIDER_TOKEN",
      "published_at_authoritative": false
    }
  ]
}
```

The referenced environment value must be supplied outside source control. Missing configured credentials fail startup rather than silently disabling or bypassing the provider.

## Security and privacy boundary

The native provider transport:

- requires HTTPS;
- rejects URL-embedded credentials, query strings, fragments, and non-443 explicit ports;
- rejects localhost, `.local`, loopback, private, link-local, multicast, documentation, benchmark, carrier-grade NAT, and other reserved address ranges;
- re-resolves the configured host at connection time and fails if any resolved address is non-public;
- rejects redirects;
- ignores ambient HTTP proxy environment configuration;
- sends no browser cookies or referrer state;
- supports only explicitly configured credentials loaded from named environment variables;
- limits a provider response to 4 MiB;
- bounds result title, snippet, URL, and result-count processing before returning data to the Search engine;
- preserves request cancellation and the engine's overall provider deadline;
- never publishes configured endpoint URLs or credentials through `/api/v1/providers/definitions`.

Search queries necessarily leave GoreeCloud when a configured external provider is used. The provider runtime does not claim anonymity and must not be presented as Privacy Shield authorization or Wardveil protected state merely because these transport controls exist.

## Timestamp authority

`published_at_authoritative` defaults to false. It may be set true only after the specific adapter/provider field has been reviewed as trustworthy publication/update metadata under the native `PublishedAtProvider` contract. Search strips untrusted or implausible timestamps regardless of provider output.

The TinyFish Search adapter is explicitly non-authoritative for publication timestamps in this Development slice.

## Lifecycle and acceptance

This runtime establishes source-level provider execution capability only. Before a provider can support production or Stable qualification, GoreeCloud still requires provider-specific approval and evidence for, as applicable:

- provider identity and purpose;
- privacy and data-use behavior;
- terms and automated-access compatibility;
- credentials and secret separation;
- supported categories and representative result quality;
- rate limits, retry/degradation behavior, and abuse controls;
- response/body/result bounds;
- timestamp authority where used;
- live-provider acceptance;
- Privacy Shield and Wardveil runtime evidence;
- observability without query-history telemetry;
- recovery and rollback requirements;
- exact release and target-runtime acceptance.

The current required GoreeCloud presentation target is GLAZE UI V1.3 / 1.3.0 — Adaptive Resonance. Provider-runtime source work does not by itself establish Search application conformance with that design system.
