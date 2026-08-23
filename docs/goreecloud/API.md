# GoreeCloud Search API Boundary

## Status

Phase 1 of the GoreeCloud-owned integration API is implemented on the current development line through `GET /api/v1/status`.

GoreeCloud Search already provides `GET /healthz` through the SearXNG-derived web application. This work preserves that established health endpoint and its `text/plain` `OK` response rather than replacing or silently changing its semantics.

The general GoreeCloud-facing machine-readable Search API remains planned and unaccepted. The status contract does not enable JSON, RSS, or CSV search-result formats and does not promote the current first-Stable candidate or production runtime to Stable.

## Purpose

GoreeCloud Search will provide stable, documented interfaces for approved GoreeCloud applications, Browser integration, local AI systems, research agents, and approved automation workflows. Consumers should depend on a GoreeCloud-owned versioned contract rather than on incidental SearXNG internals.

The first implemented GoreeCloud-owned API boundary is intentionally narrow: clients can verify GoreeCloud Search identity, API-contract version, basic capabilities, implementation-foundation transparency, and canonical service paths without receiving query, result, preference, engine, plugin, or secret data.

## Phase 1 — `GET /api/v1/status`

The versioned read-only status endpoint returns:

- API contract version;
- GoreeCloud Search product and service identity;
- basic service status;
- the SearXNG foundation name and runtime version for implementation transparency;
- capability flags for HTML search, OpenSearch, and the future machine-readable Search API;
- canonical paths for health, OpenSearch, and interactive search.

The endpoint deliberately excludes:

- search query text;
- search results or provider-response content;
- search history;
- user preferences;
- enabled-engine inventories;
- plugin inventories;
- credentials, tokens, keys, secret configuration, or environment values.

The response uses `Cache-Control: no-store`, `Pragma: no-cache`, and `X-GoreeCloud-API-Version: 1`. Query-string input is ignored and is not echoed into the response.

### Example response shape

```json
{
  "api_version": "1",
  "product": "GoreeCloud Search",
  "service": "search",
  "status": "ok",
  "foundation": {
    "name": "SearXNG",
    "version": "<runtime version>"
  },
  "capabilities": {
    "html_search": true,
    "opensearch": true,
    "machine_readable_search_api": false
  },
  "endpoints": {
    "health": "/healthz",
    "opensearch": "/opensearch.xml",
    "search": "/search"
  }
}
```

## Existing process health — `GET /healthz`

The API status contract advertises the existing `/healthz` application-process health path already implemented in the SearXNG-derived web application.

A successful response is HTTP 200, `text/plain`, with body:

```text
OK
```

This PR deliberately preserves that established contract. It does not add a duplicate health route and does not reinterpret process health as full dependency readiness.

`/healthz` proves that the GoreeCloud Search application process can serve its health response. It does **not** by itself prove that external search providers, DNS, Caddy, monitoring, backups, or every dependency are healthy. Any richer readiness or dependency-diagnostic contract should use a separately governed endpoint rather than silently changing `/healthz` semantics.

## Versioning rule

Clients must use the versioned `/api/v1/` namespace rather than treating upstream internal JSON surfaces such as `/config` as a permanent GoreeCloud application contract.

Compatible additive changes may be made within API version 1. Breaking field or semantic changes require a new API version or an explicit migration contract.

## Future machine-readable Search contract

The future interface must:

- preserve a versioned or otherwise governed response contract;
- normalize provider-specific result details where practical;
- expose source and engine provenance needed for research and debugging;
- distinguish partial-provider failure from complete request failure;
- avoid exposing secrets or administrative configuration;
- apply bounded request limits and operational safeguards;
- avoid retaining query history unless separately approved;
- remain replaceable if SearXNG is later reduced or removed from the backend.

A future response may include:

- query and normalized search parameters;
- result category and result type;
- title, URL, snippet, and published date when available;
- source engine or engines;
- score or ordering metadata only when its meaning can be documented;
- request timing and partial-failure information suitable for diagnostics;
- pagination or continuation data;
- optional structured answers or infobox data.

Before `machine_readable_search_api` can become `true`, GoreeCloud Search must separately define and validate request/response schema versioning, result normalization, category and preference semantics, provider degradation behavior, rate limiting, abuse controls, query-data minimization, error/retry contracts, authorization if required, integration tests, and production/Stable lifecycle evidence.

## Security boundary

The current GoreeCloud runtime example exposes HTML search only. JSON, RSS, and CSV search-result formats remain disabled until the API access model, authentication or network restriction, rate limiting, versioning, abuse controls, privacy boundaries, monitoring expectations, and lifecycle acceptance are approved.

The Phase 1 status endpoint is intentionally non-sensitive and read-only. It is not an administrative API and must not grow secret-bearing configuration fields merely for convenience.

## Compatibility

No application should be written against undocumented SearXNG template objects, internal Python classes, provider-specific structures, or the broad upstream `/config` representation when a GoreeCloud-owned abstraction can reasonably be introduced.
