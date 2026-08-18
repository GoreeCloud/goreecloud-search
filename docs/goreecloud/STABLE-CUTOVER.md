# GoreeCloud Search Stable Cutover

## Purpose

This record defines the controlled transition that occurs only after GoreeCloud Search is approved as its first Stable release. Until that gate is satisfied, the existing production compatibility names and rollback material remain in place even though the running application is already GoreeCloud Search.

## Stable gate

Stable approval requires all of the following evidence:

- Source CI passes for the exact release revision.
- Mobile and desktop Glaze UI acceptance passes without viewport overflow or blocking usability defects.
- General, Images, Videos, and News search paths are validated with representative queries.
- Provider failures degrade safely and do not break the application shell.
- Private AdGuard DNS resolution, NetBird access, Caddy HTTPS routing, and backend Docker routing remain validated.
- Uptime monitoring is healthy under the GoreeCloud Search product identity.
- Backup and rollback material is verified and a recovery path is documented.
- The release image is pinned by immutable digest.
- GoreeCloud Browser integration passes with GoreeCloud Search configured as the only and default browser search engine.
- Address-bar, new-tab, and browser search-field queries resolve through GoreeCloud Search without direct or silent fallback to an external browser-level search provider.
- Browser failure handling surfaces GoreeCloud Search connectivity/service state instead of bypassing the private search service.
- Required AGPL licensing, source availability, and SearXNG upstream attribution remain intact.

## Stable cutover scope

After the Stable gate passes, GoreeCloud Search becomes the sole active GoreeCloud search-service identity. The maintenance change should migrate internal runtime names from the historical SearXNG-derived deployment names to GoreeCloud Search names, including the Compose project, stack directory, application-data directory, container names, internal Docker network, Caddy backend target, and monitoring label.

Target names:

- `/srv/docker/stacks/goreecloud-search`
- `/srv/docker/appdata/goreecloud-search`
- `goreecloud-search`
- `goreecloud-search-valkey`
- `goreecloud-search-internal`
- Compose project `goreecloud-search`

The current compatibility names must not be removed until the Stable migration is validated and rollback evidence is captured.

## Documentation transition

After Stable acceptance, active GoreeCloud documentation must use **GoreeCloud Search** as the canonical search product and service. Current-state inventories, strategies, architectures, AI integration records, monitoring records, VPS records, software-portfolio records, deployment instructions, Browser integration records, and task records must no longer present upstream SearXNG as an independently operated GoreeCloud service.

Historical change logs must remain historically accurate. References required to explain provenance, upstream maintenance, licensing, compatibility, or migration history must remain. GoreeCloud Search is a maintained fork derived from SearXNG, so legal and technical upstream attribution must not be erased.

The post-Stable documentation pass must explicitly inspect records that previously treated SearXNG as an active GoreeCloud component, including local-AI/research architecture, software portfolio, VPS runtime inventory, listening/routing records, Caddy references, monitoring records, Docker records, GoreeCloud Browser search-provider configuration, and Search project/change-log material. Those records must be updated in the same cutover window so the operational documentation cannot drift from the deployed service identity.

## Runtime retirement

Once the Stable cutover and rollback window are complete:

1. Confirm the new GoreeCloud Search stack is healthy and serving private HTTPS successfully.
2. Confirm persistent configuration/cache/Valkey data is present at the approved GoreeCloud Search paths.
3. Confirm Caddy targets the new GoreeCloud Search container name.
4. Confirm monitoring uses the GoreeCloud Search identity and retains required history.
5. Confirm GoreeCloud Browser uses GoreeCloud Search as its only/default search engine and has no external browser-level fallback search provider.
6. Confirm no active Compose stack or container depends on the historical SearXNG runtime names.
7. Remove obsolete upstream `searxng/searxng` container images from the VPS if no rollback dependency remains.
8. Remove obsolete historical stack/application-data paths only after validation and retention requirements are satisfied.
9. Retain required source attribution, AGPL material, upstream documentation, and migration history.

## Cutover verification

The cutover is not complete until both runtime and documentation audits pass.

Runtime audit must confirm:

- no running container uses the historical SearXNG service/container name;
- no active Docker network uses the historical SearXNG deployment name;
- no Caddy backend target references the retired container name;
- no active monitoring target presents SearXNG as the GoreeCloud search service;
- no active stack or application-data path remains authoritative under the historical name;
- GoreeCloud Browser exposes GoreeCloud Search as its only/default search engine;
- Browser address-bar, new-tab, and search-field queries route through GoreeCloud Search;
- Browser failure behavior never silently bypasses GoreeCloud Search with an external search provider;
- the public/private user-facing product identity remains GoreeCloud Search on home, results, Preferences, About, errors, PWA metadata, OpenSearch metadata, and browser integration surfaces.

Documentation audit must classify every remaining `SearXNG` reference as one of the following:

- required upstream attribution or licensing;
- maintained-fork provenance or upstream-sync guidance;
- historically accurate change/migration history;
- implementation-level compatibility terminology that cannot yet be renamed safely.

Any remaining reference that describes SearXNG as the current GoreeCloud search product, VPS service, user-facing application, monitoring identity, browser search provider, or canonical architecture component fails the cutover and must be corrected before the retirement is considered complete.

## Safety boundary

Do not interpret “remove SearXNG” as deleting upstream license notices, copyright notices, source provenance, upstream links needed for attribution, or implementation-level compatibility identifiers required by the maintained fork. The goal is to remove SearXNG as an active GoreeCloud service identity and VPS deployment, not to conceal the software lineage of GoreeCloud Search.
