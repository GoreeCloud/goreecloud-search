# GoreeCloud Search Readiness

This document defines the readiness boundary for GoreeCloud Search. It separates source and CI readiness from target-environment acceptance so a green development branch is not mistaken for an authorized production deployment.

## Governing release boundary

GoreeCloud Search is a GoreeCloud-maintained SearXNG fork. A release candidate must preserve upstream AGPL obligations and required attribution while presenting GoreeCloud Search as the controlled product identity.

A stable GoreeCloud Search release requires all applicable GoreeCloud production gates to be satisfied:

- security and safe-operation readiness;
- private-access and individual-user boundary readiness;
- Glaze UI 1.0 conformance;
- deployment and rollback readiness;
- monitoring and health validation;
- backup, restore, and recovery readiness;
- representative provider acceptance;
- browser and supported-form-factor acceptance;
- integration acceptance for every enabled consumer.

Source readiness does not authorize DNS, Caddy, NetBird, firewall, container, or hostname changes.

## Current automated source gates

The repository provides deterministic gates that are safe to run on every proposed source revision.

### Foundation

Validates GoreeCloud product markers, AGPL preservation, upstream provenance, Python syntax, privacy defaults, Glaze UI 1.0 semantics, deployment safeguards, provider-test boundaries, and Compose syntax.

### Runtime smoke

Starts the exact source revision with GoreeCloud runtime settings and validates application startup, `/healthz`, core pages, search behavior, privacy-facing headers, and product identity without relying on live external providers.

### Browser acceptance

Runs Chromium against Compact, Medium, Expanded, and Wide Glaze UI layouts. It validates GoreeCloud identity, browser metadata, OpenSearch discovery, manifest identity, minimum practical target sizing, keyboard focus behavior, About and Preferences surfaces, and horizontal-overflow safety.

### Container acceptance

Builds the GoreeCloud Search OCI image from the exact source revision and validates startup, health, product identity, and container metadata.

### Upstream Integration

Retains the SearXNG integration suite so GoreeCloud branding and UI changes do not silently break the upstream engine, settings, unit, Robot, lint, or theme contracts.

## Manual provider acceptance

External providers are intentionally not part of required deterministic pull-request CI because they can throttle, block, rate-limit, challenge, or change independently of GoreeCloud Search.

The manual provider workflow must be run before a stable production promotion. Acceptance should include representative queries across the enabled search categories and should record:

- which engines returned useful results;
- latency and timeout behavior;
- access-denied, CAPTCHA, and rate-limit behavior;
- image-proxy behavior where relevant;
- whether failure of one engine remains isolated from the overall service;
- whether automated traffic remains reasonable and non-abusive.

A provider failure is not automatically an application failure. Persistent low-value or unstable engines should be reviewed for disablement.

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
- machine-readable JSON, CSV, and RSS formats disabled until an integration contract is approved.

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

Machine-readable search formats remain disabled in the foundation profile until that contract is accepted. Enabling JSON or another API format is therefore a deliberate deployment change, not a default product feature.

## Monitoring acceptance

Production acceptance should verify at least:

- container/process availability;
- `/healthz` from the intended internal network path;
- Caddy HTTPS success;
- private DNS resolution;
- certificate validity;
- representative search completion;
- latency and repeated engine failures;
- resource usage appropriate to actual demand;
- an actionable alert path for a sustained service outage.

A healthy process alone is not enough when the application cannot perform a search through the intended private route.

## Backup and recovery acceptance

The authoritative configuration and recovery material must be identified before production promotion. At minimum this includes the reviewed settings, deployment definition, protected secret recovery method, private publication configuration, and any persistent state that the selected runtime actually requires.

A stable deployment must have a documented rebuild/restore procedure and a representative recovery test. Cache-only data does not need to be preserved merely because it exists; recovery priority should follow whether the data is unique and required to restore the service capability.

## Upgrade and rollback acceptance

Before a material upgrade:

1. Record the deployed image/source revision.
2. Verify current recovery material and required credentials.
3. Review upstream release and dependency changes.
4. Build and validate the candidate revision.
5. Run representative provider and browser acceptance.
6. Preserve the previous known-good image and configuration.
7. Deploy through the approved private path.
8. Validate DNS, HTTPS, health, search, monitoring, and integrations.
9. Roll back immediately if the release does not satisfy the acceptance boundary.

The previous deployment must not be destroyed merely because the candidate starts successfully.

## Glaze UI 1.0 acceptance

GoreeCloud Search targets Glaze UI 1.0. Stable visual acceptance requires:

- canonical semantic colors and spacing behavior;
- Canvas, Solid, Raised, Glaze, and Overlay hierarchy where appropriate;
- recognizable GoreeCloud identity without removing required upstream attribution;
- light and dark appearance quality;
- Compact, Medium, Expanded, and Wide adaptive behavior;
- visible keyboard focus and practical target sizing;
- reduced-motion, reduced-transparency, increased-contrast, and forced-colors resilience;
- readable solid fallbacks when blur is unavailable;
- local or application-owned UI assets without analytics, trackers, remote fonts, or remote icon runtimes.

Rounded cards or translucency by themselves do not satisfy Glaze UI conformance.

## Stable-release decision

A GoreeCloud Search source revision may be described as source-stable or release-candidate-ready when the deterministic repository gates pass.

It may be described as production-ready only after the target deployment has separately passed private-access, provider, browser, monitoring, backup/restore, integration, and rollback acceptance and those results have been recorded in the appropriate GoreeCloud documentation.
