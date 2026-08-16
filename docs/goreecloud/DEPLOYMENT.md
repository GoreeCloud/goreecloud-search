# GoreeCloud Search Deployment Boundary

## Status

Development documentation. Production cutover is not approved by this file.

## Runtime model

GoreeCloud Search is intended to run as a Docker-managed private web service behind the approved GoreeCloud reverse-proxy and private-network architecture. The application should bind to loopback or an otherwise explicitly approved private interface rather than exposing its application port directly to the public Internet.

## Configuration

`goreecloud/settings.yml.example` is the GoreeCloud-owned runtime override example. It intentionally inherits upstream defaults with `use_default_settings: true` and records only the GoreeCloud-specific baseline.

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
- no direct application-port publication.

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
8. private DNS, Caddy, and NetBird routing;
9. secret separation;
10. backup and restoration procedures;
11. monitoring and rollback procedures.

No DNS, Caddy, firewall, NetBird, or current production SearXNG configuration should be changed merely because a development branch passes source-level CI.
