# GoreeCloud Search — Glaze UI 1.1 Conformance

## Status

- Target design-system version: **Glaze UI 1.1.0**
- Canonical Glaze UI source revision: `5c8320de4f770614a3e2bcf9de2a27f7fcfd920c`
- Source-conformance status: Enforced by exact-head CI before integration
- Stable conformance status: Pending visual acceptance
- Product: GoreeCloud Search
- Maintained-fork foundation: SearXNG

## Purpose

This record binds GoreeCloud Search to the canonical Glaze UI 1.1 semantic contract without replacing the Search-specific composition that has already passed adaptive browser acceptance. Glaze UI is a shared semantic and interaction language, not a requirement that every GoreeCloud product use identical page composition.

The existing GoreeCloud Search Glaze styles remain the accepted product-specific foundation. `goreecloud-glaze-1.1.css` is loaded as the final GoreeCloud semantic layer and adds the Glaze UI 1.1 roles that were not part of the original Search 1.0 adaptation.

## Canonical mapping

The Search web client records Glaze UI `1.1.0` in runtime metadata and maps the current canonical release into local CSS semantics for:

- light and dark canvas, surface, text, line, accent, status, on-accent, information, and scrim roles;
- hover, pressed, focus, and selected state-layer strengths;
- spacing, radius, icon, density, blur, opacity, shadow, typography, and layout roles;
- minimum and comfortable interaction targets;
- Compact, Medium, Expanded, and Wide adaptive ranges;
- navigation, raised, scrim, overlay, and toast layering semantics;
- safe-area insets for viewport-bounded mobile rendering;
- disabled-state feedback and explicit focus treatment;
- reduced transparency and forced-colors resilience.

No remote font, remote icon, remote stylesheet, analytics runtime, or third-party user-interface dependency is introduced by the Glaze UI 1.1 layer.

## Product-specific implementation

GoreeCloud Search retains its existing product personality and information architecture:

- the focused private-search landing page;
- the query masthead and category navigation;
- provider filters and search result cards;
- Files and Videos specialized result presentations;
- Preferences and About surfaces;
- empty, error, and recovery states.

The 1.1 adoption therefore augments the semantic contract instead of replacing the already accepted Search composition with a generic shared shell.

## Accessibility and resilience

The current source contract preserves and extends the established Search requirements for:

- keyboard access and visible focus;
- practical 44-pixel minimum actionable targets;
- semantic labels and native control behavior;
- reduced motion through existing GoreeCloud Search layers;
- reduced transparency and solid fallbacks;
- increased-contrast and forced-colors support;
- page-level horizontal containment;
- intentional local scrolling for dense controls and tables;
- safe-area-aware Compact viewport behavior;
- useful content and critical actions when blur or nonessential presentation effects are unavailable.

## Automated evidence

`tests/unit/test_goreecloud_glaze_ui_11.py` is the application-level fail-closed Glaze UI 1.1 source contract. It verifies:

- exact runtime version metadata;
- exact canonical Glaze UI source revision binding;
- final stylesheet ordering;
- the expanded 1.1 token subset used by Search;
- state-layer, disabled-state, safe-area, adaptive, reduced-transparency, and forced-colors behavior;
- absence of remote UI dependencies in the 1.1 layer;
- this version-specific conformance record.

The browser-acceptance workflow executes that source contract before application startup and adaptive browser acceptance. Integration is permitted only after the exact pull-request head satisfies the required Search checks; the durable status above records that governance rule rather than a transient workflow state.

## Stable visual acceptance gate

Source conformance is not equivalent to Stable visual acceptance. Before GoreeCloud Search may claim **Glaze UI 1.1 conformant** for a Stable release, I must complete representative visual review of the integrated final candidate in:

- Compact light appearance;
- Compact dark appearance;
- Expanded light appearance;
- Expanded dark appearance.

The existing physical Android/mobile review and desktop regression review required by the first-Stable Search gate satisfy this requirement only when they are performed against the exact final candidate carrying this Glaze UI 1.1 contract.

Any material exception discovered during that review must be recorded with the affected rule, reason, user impact, fallback, and review condition before Stable promotion.

## Production boundary

This source adoption does not deploy a new Search image, modify the Docker stack, rename the historical runtime, change Caddy, DNS, AdGuard Home, NetBird, provider configuration, monitoring, backups, or authorize Stable production cutover. Those remain separate target-environment acceptance gates.
