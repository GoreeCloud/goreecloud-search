# GoreeCloud Search

GoreeCloud Search is a GoreeCloud-maintained fork of SearXNG and the planned private search, metasearch, discovery, and AI research gateway for GoreeCloud.

## Role

GoreeCloud Search provides one controlled interface for searching approved external providers while reducing direct dependence on any single commercial search service. It is designed for interactive search, browser integration, GoreeCloud application integrations, and approved local-AI research workflows.

## Product boundary

GoreeCloud Search does not replace a browser, a local knowledge base, retrieval-augmented generation, source verification, or independent research judgment.

GoreeCloud Search provides discovery and remains independent from downstream knowledge-management applications.

## Upstream relationship

Upstream project: SearXNG

Upstream repository: https://github.com/searxng/searxng

Initial GoreeCloud baseline: `b2da6b90f2f8446557c91f67d6be5064ab785ecd`

GoreeCloud development repository: https://github.com/GoreeCloud/goreecloud-search

The fork should preserve a clean upstream relationship. GoreeCloud-specific changes should be isolated and documented where practical, and upstream changes must be reviewed before incorporation into the GoreeCloud release line.

## Design and Glaze UI conformance

The user-facing application targets the GoreeCloud Glaze UI 1.0 design language. GoreeCloud-specific styling is maintained in `searx/static/themes/simple/goreecloud.css` so the product layer remains identifiable during upstream review and resynchronization.

The current conformance contract includes recognizable GoreeCloud identity, semantic surface and motion tokens, deliberate surface hierarchy, visible keyboard focus, practical 44-pixel interaction targets, responsive compact-layout behavior, light and dark themes, reduced-motion behavior, reduced-transparency handling, increased-contrast handling, forced-colors compatibility, and solid fallbacks when backdrop filtering is unavailable.

Glaze UI conformance is guarded by the GoreeCloud foundation workflow and complemented by rendered browser acceptance. Passing source guards alone is not sufficient evidence of visual acceptance; representative desktop and mobile rendering must also pass the browser gate.

## Privacy

GoreeCloud Search should minimize query retention, profiling, telemetry, and unnecessary logging. External search providers can still observe requests originating from the GoreeCloud infrastructure and can enforce their own policies, rate limits, captchas, regional behavior, and result restrictions.

## Validation layers

The repository intentionally separates deterministic validation from external-provider acceptance.

- Foundation validation checks GoreeCloud product markers, browser metadata, Glaze UI resilience requirements, privacy-oriented defaults, deployment structure, licensing, and source syntax.
- Runtime smoke validation starts the application with GoreeCloud settings and checks the rendered product shell and privacy-facing behavior.
- Browser acceptance exercises representative desktop and mobile application behavior, responsive containment, browser metadata, and install/search integration.
- Container acceptance builds and runs the GoreeCloud-derived image and verifies runtime identity from the built artifact.
- Upstream Integration preserves compatibility pressure from SearXNG linting, unit, Robot/browser, documentation, shell, localization, and theme tests.
- Real-provider acceptance is manual by design because third-party throttling, captchas, regional behavior, and availability are external conditions that should not make routine pull-request CI nondeterministic.

## Release boundary

A green source branch is a candidate for deployment acceptance, not automatic production authorization. A stable release requires the deterministic repository gates to pass at the same revision and then requires target-environment validation of configuration, secrets, networking, Caddy, DNS, provider behavior, monitoring, backup and restore, accessibility, rollback, and recovery.

Production replacement must not be inferred merely from a successful image build, browser test, or pull request. Real-provider behavior and target-environment recovery remain separate acceptance requirements.

## Operational boundary

Source development does not authorize production replacement. Production deployment requires separate validation for configuration, secrets, networking, Caddy, DNS, monitoring, backup and restore, provider behavior, accessibility, and target-environment acceptance.

## Licensing

GoreeCloud Search remains subject to the GNU Affero General Public License v3.0 or later and all applicable upstream notices and source-availability obligations. GoreeCloud branding does not remove or supersede upstream copyright and licensing requirements.
