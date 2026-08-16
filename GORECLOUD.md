# GoreeCloud Search

GoreeCloud Search is a GoreeCloud-maintained fork of SearXNG and the planned private search, metasearch, discovery, and AI research gateway for GoreeCloud.

## Role

GoreeCloud Search provides one controlled interface for searching approved external providers while reducing direct dependence on any single commercial search service. It is designed for interactive search, browser integration, GoreeCloud application integrations, and approved local-AI research workflows.

## Product boundary

GoreeCloud Search does not replace a browser, GoreeCloud Research Library, a local knowledge base, retrieval-augmented generation, source verification, or independent research judgment.

GoreeCloud Search provides discovery. GoreeCloud Research Library preserves and manages sources after discovery.

## Upstream relationship

Upstream project: SearXNG

Upstream repository: https://github.com/searxng/searxng

Initial GoreeCloud baseline: `b2da6b90f2f8446557c91f67d6be5064ab785ecd`

GoreeCloud development repository: https://github.com/GoreeCloud/goreecloud-search

The fork should preserve a clean upstream relationship. GoreeCloud-specific changes should be isolated and documented where practical, and upstream changes must be reviewed before incorporation into the GoreeCloud release line.

## Design

The user-facing application follows the GoreeCloud Glaze UI design language. The first design pass focuses on product identity, layered surfaces, readable search results, responsive behavior, keyboard accessibility, light and dark themes, reduced-motion behavior, and strong fallbacks when translucency is unavailable.

## Privacy

GoreeCloud Search should minimize query retention, profiling, telemetry, and unnecessary logging. External search providers can still observe requests originating from the GoreeCloud infrastructure and can enforce their own policies, rate limits, captchas, regional behavior, and result restrictions.

## Operational boundary

Source development does not authorize production replacement. Production deployment requires separate validation for configuration, secrets, networking, Caddy, DNS, monitoring, backup and restore, provider behavior, accessibility, and target-environment acceptance.

## Licensing

GoreeCloud Search remains subject to the GNU Affero General Public License v3.0 or later and all applicable upstream notices and source-availability obligations. GoreeCloud branding does not remove or supersede upstream copyright and licensing requirements.
