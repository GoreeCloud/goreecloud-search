# GoreeCloud Search API Boundary

## Status

The GoreeCloud-owned native Search service exposes a versioned API v1 development contract. These endpoints are implemented in `native/cmd/searchd` and remain **pre-Stable** with `production_approved: false` where lifecycle state is material.

Current native endpoints are:

- `GET /healthz` — bounded native process/service identity health;
- `GET /api/v1/status` — GoreeCloud Search API/service identity and capability discovery;
- `GET /api/v1/readiness` — bounded local-native-application readiness;
- `GET /api/v1/search` — machine-readable execution through the same native engine used by the HTML results surface;
- `GET /api/v1/preferences/definitions` — read-only preference schema/definition discovery;
- `GET /api/v1/providers/definitions` — read-only sanitized provider/capability discovery;
- `GET /api/v1/sync/capabilities` — read-only Search Sync capability discovery;
- `GET /api/v1/platform/status` — read-only sanitized Privacy Shield, Wardveil Security, and Everkeep source/runtime-evidence state.

The inherited SearXNG-derived tree still contains its earlier GoreeCloud `/api/v1/status` and `/api/v1/readiness` plugin contract for transitional continuity. That legacy implementation is not the target architecture and must not be described as the native API implementation.

API source availability does not authorize unrestricted production consumers. Stable API use still requires exact-release acceptance, access/trust boundaries, rate/resource controls, provider acceptance, privacy/security evidence, monitoring, recovery, and compatibility governance appropriate to each consumer.

## Versioning and response controls

Native API endpoints use the `/api/v1/` namespace and return `X-GoreeCloud-API-Version: 1`.

Compatible additive changes may be made within version 1. Breaking field or semantic changes require a new version or an explicit migration contract.

The native service applies `Cache-Control: no-store` through its application security middleware. API endpoints do not intentionally echo unrelated query-string input into status or readiness responses.

Consumers must not depend on undocumented internal Go structures, transitional SearXNG objects, provider-specific implementation details, or the broad upstream `/config` representation when a GoreeCloud-owned contract exists.

## `GET /healthz`

The native health endpoint returns a bounded JSON response identifying:

- `service: goreecloud-search`;
- `implementation: native-development-foundation`; and
- `production_approved: false`.

This is process/service health only. It does not prove live-provider functionality, private DNS, reverse proxy behavior, monitoring, recovery, physical-device acceptance, production cutover, or Stable qualification.

The transitional SearXNG-derived runtime has a different historical `/healthz` representation (`text/plain` `OK`). Consumers must bind health expectations to the implementation/release they are actually using rather than assuming the two runtimes are byte-for-byte identical.

## `GET /api/v1/status`

The native status endpoint provides bounded first-party service discovery.

Current fields include:

- `api_version: "1"`;
- `product: "GoreeCloud Search"`;
- `service: "search"`;
- `status: "ok"`;
- `implementation: "native"`;
- `lifecycle: "pre-stable"`;
- `production_approved: false`;
- capability flags; and
- canonical native endpoint paths.

Current native capability flags identify source-level availability of:

- HTML search;
- the machine-readable Search API;
- preference definitions;
- provider definitions;
- Sync capability discovery; and
- sanitized platform-status discovery.

`machine_readable_search_api: true` means the native endpoint is implemented in the current source. `platform_status: true` similarly means the sanitized evidence-state endpoint exists in source. Neither flag means the endpoint or any underlying platform integration has received production or Stable approval.

### Representative response shape

```json
{
  "api_version": "1",
  "product": "GoreeCloud Search",
  "service": "search",
  "status": "ok",
  "implementation": "native",
  "lifecycle": "pre-stable",
  "production_approved": false,
  "capabilities": {
    "html_search": true,
    "machine_readable_search_api": true,
    "preferences_definitions": true,
    "provider_definitions": true,
    "sync_capabilities": true,
    "platform_status": true
  }
}
```

## `GET /api/v1/readiness`

The native readiness endpoint answers one deliberately narrow question: **is the local native application initialized well enough to accept its current source-level application traffic?**

Current local checks are:

- native engine initialized; and
- General category locally executable under the native engine contract.

When both checks pass, the endpoint returns HTTP 200 with `status: "ready"`, `ready: true`, and `readiness_scope: "local_native_application"`.

If the native engine is unavailable or the required local General path is not ready, it fails closed with HTTP 503, `status: "not_ready"`, and `ready: false`.

The readiness response explicitly marks `production_approved: false` and lists material areas it does not evaluate, including:

- external search providers;
- production provider credentials;
- private DNS and reverse proxy;
- monitoring and alert delivery;
- backup, restore, and rollback;
- physical-device acceptance; and
- production cutover.

`/healthz` and `/api/v1/readiness` are complementary local signals. Neither authorizes production deployment or Stable promotion. Platform-status presentation does not expand the readiness scope.

## `GET /api/v1/platform/status`

The platform-status endpoint exposes a minimized application-side view of platform integration evidence without manufacturing runtime or production truth.

The current Development snapshot identifies three authoritative systems:

- Privacy Shield — authority contract `contracts/privacy-shield.platform-evidence.runtime-acceptance.json`;
- Wardveil Security — authority contract `contracts/wardveil.status.schema.json`;
- Everkeep — authority contract `contracts/continuity.status.schema.json`.

Current source truth deliberately remains conservative:

- Privacy Shield source integration is present, runtime evidence is unavailable, state is `unknown`, `positive_claim` is false, and `production_accepted` is false.
- Wardveil source integration is present, runtime evidence is unverified, state is `unknown`, `positive_claim` is false, and `production_accepted` is false.
- Everkeep currently has a presentation boundary only, runtime evidence is unavailable, state is `unknown`, `positive_claim` is false, and `production_accepted` is false.
- The aggregate snapshot reports `production_approved: false`.

The response also explicitly reports that it contains no user content or query text and exposes no credentials. Unrelated query-string input is ignored and is not echoed.

Source presence, green CI, schema parsing, Glaze UI presentation, or Mesh evidence transport cannot create a positive platform claim. A future runtime adapter may strengthen state only from current authoritative evidence that satisfies the relevant producer contract and the exact Search application/release boundary.

See `docs/goreecloud/PLATFORM-STATUS.md` for the complete evidence boundary.

## `GET /api/v1/search`

The native Search API executes through the same `search.Engine` used by first-party HTML search.

Current source behavior includes:

- bounded query validation;
- explicit category validation;
- fail-closed handling when a specialized category has no executable provider path;
- bounded concurrent provider execution;
- timeout/degraded provider status;
- bounded per-provider result processing;
- URL sanitization and credential-bearing URL rejection;
- deterministic GoreeCloud-owned ranking;
- bounded trustworthy freshness where applicable;
- conservative local correction metadata where supported by current result evidence; and
- source provenance.

General may return a valid empty development response when no native production provider is configured. Images, Videos, News, and Files return a bounded not-implemented response when the current native engine has no executable provider for the requested category rather than silently falling back to General or to the transitional runtime.

Production provider availability and category parity remain separate release gates.

## `GET /api/v1/preferences/definitions`

This read-only endpoint exposes the native preference-definition contract, including schema version, recognized sections, and first-party preference definitions.

It is descriptive. It does not provide an administrator write path, override deployment policy, or authorize browser-local preferences to modify Privacy Shield, Wardveil Security, Everkeep, Identity, Mesh, or provider-management authority.

## `GET /api/v1/providers/definitions`

This read-only endpoint exposes a sanitized deterministic view of configured native provider capability.

It includes applicable:

- configured provider count;
- sanitized provider identities;
- supported categories;
- categories executable by the current native engine;
- publication-timestamp authority capability;
- deployment-controlled management scope;
- `credentials_exposed: false`; and
- `production_approved: false`.

It must not expose provider credentials, secret headers, raw provider error strings, mutable management controls, or unnecessary endpoint configuration.

A provider appearing here is not production approval.

## `GET /api/v1/sync/capabilities`

This read-only endpoint describes Search-owned Sync datasets/capabilities such as `search.preferences`, `search.history`, and `search.sources` according to the current source contract.

It exposes no credentials and remains `production_approved: false` while end-to-end Sync/Identity/runtime acceptance is incomplete.

## Error and degradation contract

Native API errors should remain bounded and safe for first-party consumers.

The API must not expose reusable credentials, authorization headers, private keys, provider secrets, raw internal provider exceptions, or unbounded diagnostic payloads merely for debugging convenience.

Provider partial failure should be distinguishable from whole-request failure where the native engine has usable results from other sources.

## Privacy boundary

Search queries may contain sensitive information. API consumers, reverse proxies, monitoring systems, logs, and integrations must minimize unnecessary retention of query strings and result content.

The native API does not use click-history or behavioral-profile ranking and must not grow advertising or sponsored-result authority.

External providers may observe requests from GoreeCloud infrastructure. API availability does not imply external-provider anonymity.

## Security and authorization boundary

The current native development service is not an unrestricted public API product.

Before a consumer is production-approved, its contract must define applicable:

- access/network boundary;
- authentication and authorization;
- least privilege;
- request volume and concurrency;
- rate/resource limits;
- timeout and retry behavior;
- provider degradation behavior;
- logging and retention;
- privacy classification;
- compatibility/versioning;
- monitoring;
- disablement; and
- recovery.

Protected administrative or account-bound functions must use the appropriate GoreeCloud trusted authority rather than treating API versioning as authorization.

## Transitional compatibility

The SearXNG-derived plugin implementation remains in the repository only while transitional continuity requires it. Its source contract may continue to be tested so migration work does not accidentally break the active transitional service before cutover.

New first-party consumers should target the documented native API contract for the release they are approved to use. Migration must preserve or explicitly version useful GoreeCloud-owned contracts instead of making consumers depend on unstable backend internals.

## Stable API qualification

The native API may be described as source-implemented when exact-revision tests pass.

It may be described as production-ready or Stable only when the exact candidate has passed the applicable GoreeCloud Stable gates, including provider/category acceptance, Privacy Shield, Wardveil Security, Everkeep recovery, Identity/Mesh integration where applicable, monitoring, target-runtime validation, supported consumer integration, release provenance, and explicit production authorization.