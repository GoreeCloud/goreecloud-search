# GoreeCloud Search Readiness

This document defines the readiness boundary for GoreeCloud Search. It separates process health, local application readiness, deterministic source/CI readiness, and target-environment release acceptance so a green endpoint or development branch is never mistaken for an authorized production deployment or Stable release.

## Readiness layers

GoreeCloud Search uses distinct readiness layers that must not be collapsed into one signal.

### Process health — `GET /healthz`

The existing SearXNG-derived `/healthz` route returns HTTP 200 with `text/plain` body `OK` when the application process can serve its health response.

Process health does not prove that the local application is correctly configured for GoreeCloud, that external providers work, or that the intended DNS/Caddy/private-network path, monitoring, backup, restore, rollback, or Browser integration is healthy.

### Local application readiness — `GET /api/v1/readiness`

The GoreeCloud-owned API v1 readiness endpoint performs deterministic local checks only. It verifies GoreeCloud Search instance identity, HTML-search enablement, and registration of required health, OpenSearch, search, and API-status routes.

It returns HTTP 200 with `ready: true` when those local checks pass and HTTP 503 with `ready: false` when a local check fails.

The endpoint explicitly does not evaluate external search providers, DNS, reverse proxy behavior, monitoring or alert delivery, backup/restore, rollback, or actual GoreeCloud Browser runtime behavior. It is a local application readiness contract for approved consumers, not a release-governance shortcut.

### Release and production readiness

Release and production readiness remain evidence-driven governance states. They require the broader source, target-runtime, provider, recovery, monitoring, device/desktop, Browser, and rollback acceptance defined by the first-Stable process.

Neither `/healthz` nor `/api/v1/readiness`, alone or together, authorizes production cutover or Stable promotion.

## Governing release boundary

GoreeCloud Search is a GoreeCloud-maintained SearXNG fork. A release candidate must preserve upstream AGPL obligations and required attribution while presenting GoreeCloud Search as the controlled product identity.

A Stable GoreeCloud Search release requires all applicable GoreeCloud production gates to be satisfied, including:

- security and safe-operation readiness;
- private-access and individual-user boundary readiness;
- Glaze UI 1.1.0 source and final visual conformance;
- immutable candidate identity and deployment/rollback readiness;
- exact target-runtime identity acceptance;
- monitoring and approved alert-delivery validation;
- backup, restore, and recovery readiness;
- representative General, Images, Videos, News, and Files provider acceptance;
- physical-device and desktop acceptance;
- actual GoreeCloud Browser runtime integration acceptance;
- integration acceptance for every enabled consumer;
- completed first-Stable evidence binding and a separate explicit Stable decision.

Source readiness does not authorize DNS, Caddy, NetBird, firewall, container, hostname, production-image, or compatibility-name changes.

## Current automated source gates

The repository provides deterministic gates that are safe to run on every proposed source revision.

### Foundation

Validates GoreeCloud product markers, AGPL preservation, upstream provenance, Python syntax, privacy defaults, Glaze UI 1.1.0 semantics, deployment safeguards, provider-test boundaries, and Compose/source contracts.

### API v1 service contract

Validates the first-party versioned `/api/v1/status` and `/api/v1/readiness` source contracts, preservation of the existing `/healthz` semantics, privacy/non-cache behavior, and the explicit boundary that general machine-readable search remains disabled.

### Runtime smoke

Starts the exact source revision with GoreeCloud runtime settings and validates application startup, `/healthz`, `/api/v1/status`, `/api/v1/readiness`, core pages, search behavior, privacy-facing headers, and product identity without relying on live external providers.

### Browser acceptance

Runs Chromium against Compact, Medium, Expanded, and Wide Glaze UI layouts. It validates GoreeCloud identity, browser metadata, OpenSearch discovery, manifest identity, minimum practical target sizing, keyboard behavior, About and Preferences surfaces, and horizontal-overflow safety.

### Container acceptance

Builds the GoreeCloud Search OCI image from the exact source revision and validates startup, health, product identity, and container metadata.

### First-Stable evidence contracts

The repository contains fail-closed source-side validation for immutable candidate provenance, target-runtime identity, provider result integrity, recovery/rollback evidence, visual and Browser evidence, readiness reporting, and final evidence binding. These gates validate evidence structure and consistency; they do not manufacture missing real-world evidence.

### Upstream Integration

Retains the SearXNG integration suite so GoreeCloud branding, UI, API, and operational changes do not silently break upstream engine, settings, unit, Robot, lint, or theme contracts.

## Manual provider acceptance

External providers are intentionally not part of required deterministic pull-request CI because they can throttle, block, rate-limit, challenge, or change independently of GoreeCloud Search.

The candidate-bound real-provider workflow must be run before first-Stable governance review. Required representative categories are General, Images, Videos, News, and Files. Acceptance should record sanitized evidence for:

- useful result completion for every required category;
- latency, timeout, access-denied, CAPTCHA, and rate-limit behavior where encountered;
- image-proxy behavior where relevant;
- bounded partial-provider degradation;
- exact staged immutable Search candidate identity before and after requests;
- absence of persisted query text and provider-response content in the acceptance artifact.

A provider failure is not automatically an application failure. Persistent low-value or unstable engines should be reviewed separately rather than allowing one provider to redefine the readiness of the entire product.

## Private access and individual-user boundary

GoreeCloud Search is intended to remain private. The approved user path is through GoreeCloud private access controls such as NetBird, private DNS, and Caddy rather than a directly exposed application port.

The application container must remain bound to loopback or an otherwise explicitly approved private network. Public port forwarding is not part of the approved deployment model.

Before production promotion, the target environment must confirm that each approved person reaches the service through an individually attributable access identity or individually enrolled private-access device boundary. GoreeCloud Search itself does not claim to provide a standalone account database. If the target access architecture cannot provide the required individual boundary, the deployment is not production-ready.

## Privacy and logging

The baseline intentionally keeps:

- public-instance mode disabled;
- SearXNG metrics disabled;
- query text out of page titles;
- image proxying enabled;
- `noindex, nofollow` response directives;
- a `no-referrer` policy;
- machine-readable JSON, CSV, and RSS search-result formats disabled until the broader integration contract is approved;
- API v1 status/readiness responses non-cacheable and free of query, result, provider-inventory, preference, credential, and secret data.

Search queries can contain sensitive information. Reverse-proxy, container, application, network, and monitoring logs must be reviewed so query strings are not retained unnecessarily.

## AI and automation integrations

Open WebUI, AnythingLLM, local research assistants, and approved automation may use GoreeCloud Search only after the integration has a documented contract covering:

- requesting service and purpose;
- required search categories and response format;
- query sensitivity;
- network path and access boundary;
- expected request volume and concurrency;
- failure and timeout behavior;
- logging and retention;
- recovery and disablement procedure.

The API v1 status/readiness contracts do not enable machine-readable search results. General machine-readable search remains disabled until its separate schema, privacy, authorization/network restriction, rate-limit, abuse-control, provider-degradation, and lifecycle requirements are accepted.

## Monitoring acceptance

Production acceptance should verify at least:

- container/process availability;
- `/healthz` from the intended internal network path;
- `/api/v1/readiness` from the intended internal network path;
- Caddy HTTPS success;
- private DNS resolution;
- certificate validity;
- representative search completion;
- latency and repeated engine failures;
- resource usage appropriate to actual demand;
- an actionable approved alert path for a sustained service outage.

A healthy process and locally ready application are still insufficient if the intended private route cannot perform representative searches or alert on sustained failure.

## Backup and recovery acceptance

The authoritative configuration and recovery material must be identified before production promotion. At minimum this includes the reviewed settings, deployment definition, protected secret recovery method, private publication configuration, immutable candidate and rollback identities, and any persistent state that the selected runtime actually requires.

A Stable deployment must have documented and evidenced rebuild/restore and rollback procedures with representative application-level recovery testing. Cache-only data does not need to be preserved merely because it exists; recovery priority should follow whether the data is unique and required to restore service capability.

## Upgrade and rollback acceptance

Before a material upgrade:

1. Record the deployed image/source revision.
2. Verify current recovery material and required credentials.
3. Review upstream release and dependency changes.
4. Build and validate the exact candidate revision.
5. Run deterministic source, Browser, and evidence-contract gates.
6. Preserve the previous known-good immutable image and configuration.
7. Stage the candidate through the approved private/loopback acceptance path.
8. Validate exact target-runtime identity and required real-provider behavior.
9. Validate DNS, HTTPS, local readiness, monitoring/alerts, restore, rollback, and integrations.
10. Assemble and validate the required first-Stable evidence artifacts.
11. Perform production cutover only after separate explicit authorization.
12. Roll back immediately if an authorized release does not satisfy the acceptance boundary.

The previous deployment must not be destroyed merely because the candidate starts successfully.

## Glaze UI 1.1.0 acceptance

GoreeCloud Search targets canonical Glaze UI 1.1.0. Stable visual acceptance requires:

- canonical semantic colors, typography, spacing, state, and layering behavior;
- Canvas, Solid, Raised, Glaze, and Overlay hierarchy where appropriate;
- recognizable GoreeCloud identity without removing required upstream attribution;
- successful Compact light, Compact dark, Expanded light, and Expanded dark final-candidate review;
- physical Android Preferences review;
- desktop runtime/regression review;
- persisted-theme preference behavior;
- visible keyboard focus and practical target sizing;
- reduced-motion, reduced-transparency, increased-contrast, and forced-colors resilience;
- readable solid fallbacks when blur is unavailable;
- local or application-owned UI assets without analytics, trackers, remote fonts, or remote icon runtimes.

Rounded cards or translucency by themselves do not satisfy Glaze UI conformance.

## Stable-release decision

A GoreeCloud Search source revision may be described as source-valid or release-candidate-ready only when the applicable deterministic repository gates pass.

It may be described as ready for final governance review only when the required candidate-bound evidence validates and cross-binds successfully.

It may be described as production-ready or Stable only after the target deployment has separately passed private-access, provider, Browser, monitoring/alert, backup/restore, recovery, rollback, device/desktop, integration, and final evidence acceptance and an explicit GoreeCloud release-governance decision authorizes that state.
