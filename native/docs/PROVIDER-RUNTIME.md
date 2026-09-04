# GoreeCloud Search Native Provider Runtime

## Status

This document describes the native provider runtime implemented for GoreeCloud Search Development source. It is not production-provider approval, provider-selection authority, deployment evidence, or Stable acceptance.

The native application remains pre-Stable. No provider is enabled by default, no credential is committed to source, and the shipped runtime continues to report `production_approved=false`.

## Purpose

The provider runtime gives the GoreeCloud-owned Search engine a deployment-controlled executable provider path without making the native application depend on SearXNG product architecture or exposing provider credentials, endpoints, or raw failure details to users.

It implements a narrow GoreeCloud provider contract that can be used by separately reviewed provider adapters. Provider-specific privacy, terms, rate limits, authentication, timestamp authority, live behavior, and production acceptance remain separate requirements.

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

## Deployment configuration

Provider configuration is disabled unless `GOREECLOUD_SEARCH_PROVIDER_CONFIG_FILE` points to an explicitly supplied JSON file. The configuration is schema-versioned and fail-closed.

Example with no credential:

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

A provider that requires a bearer credential references only the environment-variable name:

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
- supports only an explicitly configured bearer credential loaded from a named environment variable;
- limits a provider response to 4 MiB;
- requires the exact v1 JSON contract and rejects unknown top-level/result fields;
- bounds result title, snippet, URL, and result-count processing before returning data to the Search engine;
- preserves request cancellation and the engine's overall provider deadline;
- never publishes configured endpoint URLs or credentials through `/api/v1/providers/definitions`.

Search queries necessarily leave GoreeCloud when a configured external provider is used. The provider runtime does not claim anonymity and must not be presented as Privacy Shield authorization or Wardveil protected state merely because these transport controls exist.

## Timestamp authority

`published_at_authoritative` defaults to false. It may be set true only after the specific adapter/provider field has been reviewed as trustworthy publication/update metadata under the native `PublishedAtProvider` contract. Search strips untrusted or implausible timestamps regardless of provider output.

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

The current Stable GoreeCloud presentation baseline is GLAZE UI V1.1 / 1.1.0. Provider-runtime source work does not by itself establish Search application conformance with that design system.
