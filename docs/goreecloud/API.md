# GoreeCloud Search API Boundary

## Status

Phase 1 of the GoreeCloud-owned integration API is implemented on the current development line through two read-only endpoints:

- `GET /api/v1/status` for stable service identity, capabilities, and canonical paths;
- `GET /api/v1/readiness` for bounded local-application readiness.

GoreeCloud Search also provides `GET /healthz` through the SearXNG-derived web application. This work preserves that established health endpoint and its `text/plain` `OK` response rather than replacing or silently changing its semantics.

The general GoreeCloud-facing machine-readable Search API remains planned and unaccepted. The Phase 1 service contracts do not enable JSON, RSS, or CSV search-result formats and do not promote the current first-Stable candidate or production runtime to Stable.

## Purpose

GoreeCloud Search will provide stable, documented interfaces for approved GoreeCloud applications, Browser integration, local AI systems, research agents, monitoring, and approved automation workflows. Consumers should depend on a GoreeCloud-owned versioned contract rather than on incidental SearXNG internals.

The first implemented GoreeCloud-owned API boundary is intentionally narrow. Clients can verify GoreeCloud Search identity, API-contract version, basic capabilities, implementation-foundation transparency, canonical service paths, and bounded local-application readiness without receiving query, result, preference, engine, plugin, provider-response, or secret data.

## Phase 1 — `GET /api/v1/status`

The versioned read-only status endpoint returns:

- API contract version;
- GoreeCloud Search product and service identity;
- basic service status;
- the SearXNG foundation name and runtime version for implementation transparency;
- capability flags for HTML search, OpenSearch, and the future machine-readable Search API;
- canonical paths for health, readiness, OpenSearch, and interactive search.

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
    "readiness": "/api/v1/readiness",
    "search": "/search"
  }
}
```

## Phase 1 — `GET /api/v1/readiness`

The readiness endpoint answers one deliberately narrow question: **is the local GoreeCloud Search application configured and wired well enough to accept approved local application traffic?**

It verifies only deterministic local application conditions:

- the configured instance identity is `GoreeCloud Search`;
- HTML search remains enabled;
- the existing `/healthz` route is registered;
- the OpenSearch route is registered;
- the interactive `/search` route is registered;
- the versioned `/api/v1/status` route is registered.

When every local check passes, the endpoint returns HTTP 200 with `status: "ready"` and `ready: true`. If any local check fails, it returns HTTP 503 with `status: "not_ready"` and `ready: false`.

The readiness response uses the same `Cache-Control: no-store`, `Pragma: no-cache`, and `X-GoreeCloud-API-Version: 1` controls as the status response. Query-string input is ignored and is not echoed.

### Example ready response shape

```json
{
  "api_version": "1",
  "product": "GoreeCloud Search",
  "service": "search",
  "status": "ready",
  "ready": true,
  "readiness_scope": "local_application",
  "checks": {
    "service_identity": true,
    "html_search_enabled": true,
    "health_route_registered": true,
    "opensearch_route_registered": true,
    "search_route_registered": true,
    "status_route_registered": true
  },
  "not_evaluated": [
    "external_search_providers",
    "dns",
    "reverse_proxy",
    "monitoring_and_alert_delivery",
    "backup_restore_and_rollback"
  ]
}
```

### Readiness boundary

`/api/v1/readiness` is intentionally **not** a substitute for final-candidate, production, or Stable acceptance. It does not contact external search providers and does not prove:

- representative General, Images, Videos, News, or Files provider success;
- DNS or private-network reachability;
- Caddy or other reverse-proxy behavior;
- monitoring or approved alert delivery;
- backup, restore, or rollback capability;
- GoreeCloud Browser runtime integration;
- production cutover or Stable authorization.

Those remain separate evidence and governance requirements. A consumer may use this endpoint to decide whether the local application itself is ready, but release governance must continue to evaluate the broader system independently.

## Existing process health — `GET /healthz`

The API status contract advertises the existing `/healthz` application-process health path already implemented in the SearXNG-derived web application.

A successful response is HTTP 200, `text/plain`, with body:

```text
OK
```

The Phase 1 API deliberately preserves that established contract. It does not add a duplicate health route and does not reinterpret process health as application readiness or full dependency readiness.

`/healthz` proves that the GoreeCloud Search application process can serve its health response. `/api/v1/readiness` adds bounded local application/configuration checks. Neither endpoint by itself proves that external search providers, DNS, Caddy, monitoring, backups, recovery, or every dependency are healthy.

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

The Phase 1 status and readiness endpoints are intentionally non-sensitive and read-only. They are not administrative APIs and must not grow secret-bearing configuration, provider inventories, query data, or response content merely for convenience.

## Compatibility

No application should be written against undocumented SearXNG template objects, internal Python classes, provider-specific structures, or the broad upstream `/config` representation when a GoreeCloud-owned abstraction can reasonably be introduced.
