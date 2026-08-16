# GoreeCloud Search Deployment Boundary

## Status

Development documentation. Production cutover is not approved by this file.

## Runtime model

GoreeCloud Search is intended to run as a Docker-managed private web service behind the approved GoreeCloud reverse-proxy and private-network architecture. The application should bind to loopback or an otherwise explicitly approved private interface rather than exposing its application port directly to the public Internet.

The maintained fork uses SearXNG's upstream container build system rather than introducing a parallel, GoreeCloud-only container implementation. The fork's source, templates, Glaze UI layer, and GoreeCloud documentation therefore remain inside the image produced from the reviewed source revision.

## Configuration

`goreecloud/settings.yml.example` is the GoreeCloud-owned runtime override example. It intentionally inherits upstream defaults with `use_default_settings: true` and records only the GoreeCloud-specific baseline.

`goreecloud/compose.yml.example` is the GoreeCloud-owned Compose example. It requires an explicitly selected image, binds the service to loopback by default, requires the application secret at runtime, keeps the cache in a named volume, enables an init process for clean child-process handling, applies `no-new-privileges`, provides a bounded shutdown grace period, and health-checks the application through `/healthz` from inside the container.

Secrets must not be committed. The SearXNG secret key and any future provider credentials must be supplied through protected runtime configuration or environment variables.

## Initial product defaults

The development baseline establishes:

- instance identity as GoreeCloud Search;
- GoreeCloud source and upstream links;
- private-instance behavior;
- image proxying enabled for result privacy;
- query text excluded from page titles;
- HTML as the only enabled search response format;
- no public-instance directory integration;
- no donation link;
- loopback-only publication in the GoreeCloud Compose example;
- container health checking and bounded restart/shutdown behavior;
- browser application identity, manifest discovery, OpenSearch discovery, and light/dark browser metadata.

## Automated validation

The development branch includes four GoreeCloud-specific validation layers:

1. `goreecloud-foundation.yml` performs source, product-marker, browser-contract, privacy-default, upstream-baseline, Python syntax, and AGPL-preservation checks.
2. `goreecloud-runtime-smoke.yml` installs the reviewed source revision, loads the GoreeCloud settings file through `SEARXNG_SETTINGS_PATH`, starts the application, verifies the home/search/preferences product shell, and checks privacy-facing HTTP behavior.
3. `goreecloud-container-build.yml` builds the SearXNG-derived container image from the fork's source and runs the resulting image with the GoreeCloud runtime configuration before validating the rendered product identity and health endpoint.
4. `goreecloud-browser-acceptance.yml` runs Chromium at desktop and mobile viewport sizes, validates the visible product shell and keyboard behavior, verifies GoreeCloud browser metadata, and fails on unintended horizontal page overflow with element-level diagnostics.

The runtime smoke workflow is intentionally valuable as a schema-compatibility gate. During initial implementation it detected an invalid boolean value in the `brand` settings, which was corrected to the current string-based schema before deployment. Browser acceptance likewise detected mobile overflow in both the landing search shell and the Preferences interface, allowing those layout defects to be corrected without hiding overflow globally.

## Local acceptance commands

From a reviewed source checkout, the upstream container build path is:

```bash
make container
```

After a reviewed image has been selected, create a deployment directory containing `settings.yml` derived from `goreecloud/settings.yml.example`, set the required runtime environment values, and use `goreecloud/compose.yml.example` as the deployment starting point.

Do not use an unreviewed `latest` image tag as production provenance. Production should identify the exact reviewed GoreeCloud Search image by immutable digest or another equivalently immutable artifact reference.

After startup, verify Compose health before publishing the service through any reverse proxy or private routing layer. An unhealthy container is not ready for Caddy, DNS, or NetBird cutover even if its process is still running.

## API boundary

Machine-readable formats remain disabled until the GoreeCloud Search API contract and access controls are accepted. See `docs/goreecloud/API.md`.

## Production acceptance

Before replacing the current SearXNG runtime, the deployment must validate at minimum:

1. exact source revision and build provenance;
2. Docker build and application startup;
3. health and readiness behavior;
4. representative web, image, news, video, technical, and academic searches;
5. provider failure behavior;
6. responsive Glaze UI behavior;
7. keyboard and accessibility behavior;
8. browser/OpenSearch/manifest identity behavior;
9. private DNS, Caddy, and NetBird routing;
10. secret separation;
11. backup and restoration procedures;
12. monitoring and rollback procedures.

No DNS, Caddy, firewall, NetBird, or current production SearXNG configuration should be changed merely because a development branch passes source-level or CI runtime validation.
