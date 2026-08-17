# GoreeCloud Search Deployment Boundary

## Status

Development documentation. Production cutover is not approved by this file.

The complete source-versus-production readiness contract is maintained in `docs/goreecloud/READINESS.md`.

## Runtime model

GoreeCloud Search is intended to run as a Docker-managed private web service behind the approved GoreeCloud reverse-proxy and private-network architecture. The application should bind to loopback or an otherwise explicitly approved private interface rather than exposing its application port directly to the public Internet.

The maintained fork uses SearXNG's upstream container build system rather than introducing a parallel, GoreeCloud-only container implementation. The fork's source, templates, Glaze UI layer, and GoreeCloud documentation therefore remain inside the image produced from the reviewed source revision.

The approved user path is expected to preserve GoreeCloud's private-access model through individually attributable private access, NetBird/private networking, private DNS, and Caddy or an equivalently reviewed private publication layer. The application container itself is not treated as the public access-control boundary.

## Configuration

`goreecloud/settings.yml.example` is the GoreeCloud-owned runtime override example. It intentionally inherits upstream defaults with `use_default_settings: true` and records only the GoreeCloud-specific baseline.

`goreecloud/compose.yml.example` is the GoreeCloud-owned Compose example. It requires an explicitly selected image, binds the service to loopback by default, requires the application secret at runtime, keeps the cache in a named volume, enables an init process for clean child-process handling, applies `no-new-privileges`, provides a bounded shutdown grace period, and health-checks the application through `/healthz` from inside the container.

Secrets must not be committed. The SearXNG secret key and any future provider credentials must be supplied through protected runtime configuration or environment variables.

## Initial product defaults

The development baseline establishes:

- instance identity as GoreeCloud Search;
- GoreeCloud source and upstream links;
- private-instance behavior;
- SearXNG usage metrics disabled;
- image proxying enabled for result privacy;
- query text excluded from page titles;
- `noindex, nofollow` response directives plus `noindex, nofollow, noarchive` browser metadata;
- `no-referrer` policy;
- HTML as the only enabled search response format;
- no public-instance directory integration;
- no donation link;
- loopback-only publication in the GoreeCloud Compose example;
- container health checking and bounded restart/shutdown behavior;
- browser application identity, manifest discovery, OpenSearch discovery, and light/dark browser metadata;
- canonical Glaze UI 1.0 semantic tokens and adaptive layout ranges.

## Automated validation

The development branch includes four deterministic GoreeCloud-specific validation layers in addition to the retained upstream Integration workflow:

1. `goreecloud-foundation.yml` performs source, product-marker, browser-contract, Glaze UI 1.0, privacy-default, deployment-syntax, upstream-baseline, Python syntax, and AGPL-preservation checks.
2. `goreecloud-runtime-smoke.yml` installs the reviewed source revision, loads the GoreeCloud settings file through `SEARXNG_SETTINGS_PATH`, starts the application, verifies the home/search/preferences product shell, and checks privacy-facing configuration and HTTP behavior.
3. `goreecloud-container-build.yml` builds the SearXNG-derived container image from the fork's source and runs the resulting image with the GoreeCloud runtime configuration before validating the rendered product identity and health endpoint.
4. `goreecloud-browser-acceptance.yml` runs Chromium through the canonical Glaze UI Compact, Medium, Expanded, and Wide layout classes. It validates the visible product shell, practical target sizing, keyboard behavior, About and Preferences surfaces, GoreeCloud browser metadata, OpenSearch and manifest behavior, and unintended horizontal overflow with element-level diagnostics.

The runtime smoke workflow is intentionally valuable as a schema-compatibility gate. During initial implementation it detected an invalid boolean value in the `brand` settings, which was corrected to the current string-based schema before deployment. Browser acceptance likewise detected overflow in both the landing search shell and the Preferences interface, allowing those layout defects to be corrected without hiding overflow globally.

## Real-provider acceptance

Real search providers are intentionally tested outside the required pull-request gates because third-party engines can throttle, reject, or temporarily block shared CI runners independently of GoreeCloud Search correctness.

`goreecloud/provider_acceptance.py` performs an explicit HTML search against a running GoreeCloud Search instance and requires a configurable minimum number of rendered result cards. `.github/workflows/goreecloud-provider-acceptance.yml` exposes that check through `workflow_dispatch`, with explicit query, category, and minimum-result inputs.

A failed provider-acceptance run must be investigated rather than automatically classified as an application defect. The acceptance record should distinguish application/runtime failure from provider throttling, provider blocking, engine initialization failure, or a genuinely empty result set. Target-environment provider acceptance remains mandatory before production cutover even if a GitHub-hosted provider run succeeds.

## Local acceptance commands

From a reviewed source checkout, the upstream container build path is:

```bash
make container
```

After a reviewed image has been selected, create a deployment directory containing `settings.yml` derived from `goreecloud/settings.yml.example`, set the required runtime environment values, and use `goreecloud/compose.yml.example` as the deployment starting point.

Do not use an unreviewed `latest` image tag as production provenance. Production should identify the exact reviewed GoreeCloud Search image by immutable digest or another equivalently immutable artifact reference.

After startup, verify Compose health before publishing the service through any reverse proxy or private routing layer. An unhealthy container is not ready for Caddy, DNS, or NetBird cutover even if its process is still running.

A local real-provider check can then be run explicitly, for example:

```bash
python goreecloud/provider_acceptance.py --base-url http://127.0.0.1:8888 --category general --query "GoreeCloud private search"
```

## API and integration boundary

Machine-readable formats remain disabled until the GoreeCloud Search API contract and access controls are accepted. See `docs/goreecloud/API.md`.

Open WebUI, AnythingLLM, research-agent, and automation integration must define an access boundary, request volume, query sensitivity, logging and retention behavior, timeout/failure behavior, recovery, and disablement before JSON or another machine-readable format is enabled. Browser OpenSearch integration remains supported by the normal HTML/private-user path.

## Production acceptance

Before replacing the current SearXNG runtime, the deployment must validate at minimum:

1. exact source revision and build provenance;
2. Docker build and application startup;
3. health and readiness behavior;
4. representative web, image, news, video, technical, and academic searches;
5. provider failure behavior and clear classification of external-engine failures;
6. Glaze UI behavior across Compact, Medium, Expanded, and Wide layouts;
7. light/dark, keyboard, contrast, reduced-motion, and resilience behavior;
8. browser/OpenSearch/manifest identity behavior;
9. private DNS, Caddy, NetBird, and individual-user access boundaries;
10. secret separation and privacy-conscious logging;
11. backup, representative restoration, and recovery procedures;
12. monitoring, alerting, upgrade, and rollback procedures;
13. every enabled AI, automation, browser, or portal integration.

No DNS, Caddy, firewall, NetBird, or current production SearXNG configuration should be changed merely because a development branch passes source-level or CI runtime validation.
