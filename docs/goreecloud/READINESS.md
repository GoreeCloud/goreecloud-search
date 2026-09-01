# GoreeCloud Search Readiness

This document defines the current readiness boundary for the GoreeCloud-owned native Search application. It separates source/CI health, local native runtime behavior, integration readiness, target-environment acceptance, migration cutover, and Stable promotion so no lower lifecycle state is represented as a higher one.

## Current product boundary

GoreeCloud Search is undergoing a native migration.

- Target application: GoreeCloud-owned Go service under `native/`.
- Transitional dependency: the inherited SearXNG-derived tree remains only for continuity, migration, compatibility, feature-preservation decisions, and applicable upstream security maintenance until controlled retirement.
- Current lifecycle: pre-Stable native migration.
- Current Stable design-system consumer target: Glaze UI 2.1.0.
- Glaze UI 2.2: Candidate/design-reference line, not a Stable consumer target.

A successful source build, green pull request, working development page, or healthy transitional production service does not establish Stable qualification for the native application.

## Native readiness layers

### Process health — `GET /healthz`

The native `searchd` process exposes `GET /healthz` as a bounded JSON health response. The current development response identifies:

- `service: goreecloud-search`;
- the native development implementation state; and
- `production_approved: false`.

A successful `/healthz` response proves only that the native process can serve that route. It does not prove provider availability, production configuration, private routing, monitoring, recovery, platform integrations, migration parity, or Stable qualification.

### Service status — `GET /api/v1/status`

The native status endpoint exposes the GoreeCloud-owned API version, product/service identity, native implementation identity, pre-Stable lifecycle, source-level capability availability, canonical endpoint paths, and `production_approved: false`.

It is service-discovery evidence, not production approval. In particular, `machine_readable_search_api: true` means the native `/api/v1/search` endpoint exists in the current source; it does not mean arbitrary production consumers have been approved.

### Local application readiness — `GET /api/v1/readiness`

The native readiness endpoint evaluates a deliberately bounded local-native-application scope.

Current checks are:

- native engine initialized; and
- General category locally executable under the native engine contract.

When those checks pass, the endpoint returns HTTP 200 with `ready: true` and `readiness_scope: local_native_application`. If the native engine is unavailable or the required General path is not ready, it fails closed with HTTP 503 and `ready: false`.

The endpoint explicitly marks `production_approved: false` and does not evaluate external providers, production provider credentials, private DNS/reverse proxy, monitoring and alert delivery, backup/restore/rollback, physical-device acceptance, or production cutover.

Neither `/healthz` nor `/api/v1/readiness`, individually or together, authorizes a production deployment or Stable promotion.

### Provider capability readiness — `GET /api/v1/providers/definitions`

The native provider-definition endpoint exposes a sanitized, deterministic capability view. It may report:

- configured provider identity;
- supported categories;
- categories currently executable by the native engine;
- whether publication timestamps are authoritative for a provider;
- deployment-controlled management scope; and
- `production_approved: false` while production acceptance is incomplete.

The endpoint must not expose provider credentials, mutable management controls, raw provider errors, authorization headers, or secret endpoint configuration.

A provider interface or advertised capability is not production approval. A provider is release-ready only after its transport, privacy/terms boundary, credentials, timeouts, response/body/result bounds, degradation behavior, timestamp authority where applicable, target-runtime behavior, and representative live queries are accepted.

### Search execution readiness

Native Search currently includes bounded query validation, explicit category capability handling, concurrent provider execution, timeout/degradation handling, bounded per-provider post-processing, result URL sanitization, deterministic local ranking, bounded freshness, conservative local correction suggestions, source provenance, and first-party HTML/JSON response surfaces.

General may retain a development empty-provider path. Specialized categories fail closed when the current native engine has no executable provider for the requested category.

This source behavior is not equivalent to live-provider or production readiness.

### Preferences and local state readiness

The native Preferences surface provides device/browser-local preference behavior and bounded import/export contracts. Local preferences do not grant deployment authority and cannot override Privacy Shield, Wardveil Security, Everkeep, Identity, Mesh, or administrator-controlled provider policy.

### Sync readiness

Native Search contains source-level `search.history` capability/schema, bounded record/cursor, authenticated transport, envelope-validation, tombstone, pagination, and Ed25519 record-proof contracts.

These contracts do not establish production account-history synchronization. Production use still requires accepted Identity/session authority, deployed Sync transport, privacy controls, recovery behavior, and end-to-end integration evidence.

## Deterministic source and CI gates

The repository uses exact-revision CI as source evidence. Applicable gates currently include native foundation tests, runtime smoke, native application rendered browser acceptance, container build/runtime checks, platform-integration checks, workflow supply-chain checks, documentation checks, and retained transitional compatibility/integration validation.

The native application browser acceptance uses deterministic representative data rather than live providers. It validates, as applicable:

- homepage, Preferences, and results surfaces;
- Compact, Medium, Expanded, and Wide layouts;
- light and dark appearances;
- a 48px minimum interaction-target floor on tested actionable controls;
- visible keyboard focus;
- horizontal-overflow safety;
- scan-first result composition;
- trusted publication-date presentation;
- explicit local correction presentation;
- source-agreement and provider-health disclosure;
- local Preferences interaction and privacy-first defaults;
- empty/error states;
- Reduced Motion;
- Increased Contrast;
- Forced Colors; and
- reduced-transparency fallbacks in source styling.

These gates can establish exact-head source and bounded rendered acceptance. They do not manufacture live-provider, physical-device, target-host, recovery, or production evidence.

## Glaze UI 2.1 consumer acceptance

Glaze UI 2.1.0 is the current Stable Search consumer target.

Stable application conformance requires Search-specific evidence for the exact release candidate, including applicable:

- current semantic/material hierarchy;
- typography, spacing, geometry, focus, state, and interaction behavior;
- 48px general interaction-target floor;
- Compact, Medium, Expanded, and Wide behavior;
- Light, Dark, and supported appearance behavior;
- Reduced Motion;
- Reduced Transparency;
- Increased Contrast;
- Forced Colors;
- effects-free/solid fallbacks;
- representative homepage, results, Preferences, error, empty, degraded, security, privacy, and recovery states;
- keyboard and pointer behavior;
- representative physical-device/native acceptance where browser automation is insufficient; and
- exact-revision rendered evidence plus required human visual review.

The expanded native browser gate provides substantially broader deterministic application evidence than the earlier results-only gate. It still does not by itself prove physical-device or production conformance.

Glaze UI 2.2 documentation may guide future Search design work, but Candidate behavior cannot be used to satisfy the current Stable consumer gate.

## Privacy Shield readiness

Stable Search must prove applicable Privacy Shield behavior at the actual runtime boundary, including:

- privacy-first defaults;
- no GoreeCloud advertising or sponsored-result ranking;
- no click-history/behavioral-profile ranking dependency;
- minimized query/history retention;
- privacy-safe provider diagnostics;
- logging and telemetry boundaries;
- no hidden provider bypass of the configured Search authority;
- no unnecessary credential or private-content exposure; and
- explicit controls before account history, personalization, or similar persistent convenience features are enabled.

External providers may still observe requests originating from GoreeCloud infrastructure. Search must not claim anonymity from those providers.

## Wardveil Security readiness

Stable Search requires application-specific Wardveil evidence appropriate to the deployed release, including secure defaults, request/input validation, provider trust boundaries, secret handling, dependency/security maintenance, authorization for protected functionality, safe failure behavior, deployment hardening, and production security validation.

Source tests or Wardveil platform validation alone do not certify the Search deployment.

## Everkeep and recovery readiness

Stable Search requires evidence-backed recovery rather than backup assumptions. Applicable release evidence must identify and validate:

- exact source and release artifact;
- native deployment configuration;
- provider policy/configuration required to restore service;
- protected secret-recovery path without copying reusable secrets into ordinary evidence;
- reverse-proxy/private-publication configuration;
- user-controlled persistent state included in the selected release;
- backup coverage;
- isolated or representative restore procedure;
- post-restore integrity/service validation;
- previous known-good rollback target; and
- migration rollback while the transitional runtime remains available.

Rebuildable cache data is not authoritative simply because it exists.

## Provider and category acceptance

Before Stable promotion, every category selected for the release must have production-approved native execution coverage.

Provider acceptance must record sanitized evidence for applicable:

- useful result completion;
- result integrity and URL safety;
- category correctness;
- latency and timeout behavior;
- rate limiting, access denial, CAPTCHA, and provider outages where encountered;
- bounded provider response/body/result processing;
- degradation when one provider fails;
- timestamp-authority correctness where freshness is used;
- absence of raw credentials or private query/result content in release evidence; and
- exact release candidate/runtime identity during the test.

No provider becomes trusted merely because it implements the native interface.

## Monitoring and operational readiness

The production candidate requires privacy-conscious monitoring appropriate to Search, including applicable:

- process/container availability;
- native `/healthz` and `/api/v1/readiness` availability through the intended private route;
- HTTPS/reverse-proxy success;
- private DNS resolution;
- certificate validity;
- representative search completion;
- provider failure/timeout trends without collecting unnecessary query text;
- latency and resource pressure;
- supporting-runtime health where required; and
- verified actionable alert delivery for sustained failure.

A healthy process and locally ready application are not sufficient when the intended private user path cannot perform representative searches.

## Identity and private-access readiness

Search must use GoreeCloud Identity for account/session/authorization authority wherever the selected release includes account-bound functionality. Search must not invent a parallel identity authority.

Where Search remains private without an application-level account boundary, the deployed access model must still provide the approved private-access and individual-attribution boundary required by GoreeCloud policy. Direct public application-port exposure is not an approved shortcut.

## Mesh and consumer integration readiness

GoreeCloud Mesh is authoritative for cross-application capability coordination. Approved Search consumers must have documented and tested integration contracts covering trust, authentication/authorization where applicable, query sensitivity, request volume, timeouts, failure/degraded behavior, logging, compatibility, recovery, and disablement.

The existence of `/api/v1/search` in native source does not by itself make the machine interface Stable for unrestricted production consumers.

## Migration and cutover readiness

The inherited SearXNG-derived runtime remains a controlled transitional dependency until the native release has sufficient accepted capability and rollback evidence.

Before native production cutover:

1. Identify one exact release candidate and immutable artifact.
2. Complete exact-head source/CI gates.
3. Complete required native category/provider live acceptance.
4. Complete Glaze UI 2.1 whole-application and required device acceptance.
5. Complete Privacy Shield and Wardveil runtime/evidence acceptance.
6. Complete Everkeep backup/restore/migration/rollback acceptance.
7. Validate required Identity, Mesh, Browser, AI, and other consumer integrations.
8. Validate target-host private networking, DNS, reverse proxy, TLS, monitoring, alert delivery, and resource/abuse boundaries.
9. Preserve the previous known-good transitional deployment until rollback is proven.
10. Perform a controlled cutover only after explicit production authorization.
11. Verify the deployed artifact/configuration matches the exact approved candidate.
12. Re-run representative production acceptance after cutover.

If the candidate or relevant configuration changes after final validation, it becomes a new candidate and applicable validation must be repeated.

## Stable-release decision

A native Search revision may be described as source-valid when its applicable deterministic gates pass.

It may be described as a release candidate only when an exact candidate identity and artifact are established and no source-level release blocker remains for the selected scope.

It may be described as production-ready or Stable only when all applicable GoreeCloud Stable qualification gates have passed for that exact release, artifact, supported platform, and deployment boundary, and no unresolved release blocker remains.

Stable is not a synonym for working, green CI, source-complete, deployed, or long-running.