# GoreeCloud Search Native Experience Revamp

## Objective

GoreeCloud Search is being rebuilt as original GoreeCloud-owned native software. The native application will preserve useful Search capabilities while replacing the transitional SearXNG-derived product architecture, interface, preferences experience, and general application logic through controlled migration slices.

The target experience is visually distinctive, responsive, private by default, accessible, and conformant with Glaze UI, Wardveil Security, Privacy Shield, and Everkeep.

## Homepage

The homepage is a first-party GoreeCloud Search surface rather than a reproduction of the inherited interface.

It must provide:

- a polished Glaze UI composition across Compact, Medium, Expanded, and Wide layouts;
- prominent GoreeCloud Search identity and a focused primary search field;
- keyboard-first and touch-first query entry;
- clear access to General, Images, Videos, News, Files, and every other approved retained category;
- excellent Light, Dark, and Deep Dark presentation under the current Stable Glaze UI appearance contract;
- reduced-motion, reduced-transparency, increased-contrast, and forced-colors behavior;
- responsive empty, loading, degraded-provider, offline, error, and no-results states;
- accessible suggestions/autocomplete when enabled;
- an understandable private-search posture without advertising or marketing clutter;
- no behavioral tracking, click tracking, advertising, or browser-level provider bypass.

## Results

The native results experience preserves and improves useful Search behavior:

- federated multi-provider search behind the GoreeCloud Search boundary;
- deterministic aggregation, URL normalization, deduplication, ranking, and bounded provider degradation;
- General, Images, Videos, News, Files, and approved specialized categories;
- category-aware result cards and media layouts;
- language, time, SafeSearch, and other approved query filters;
- useful Special Queries equivalents;
- keyboard navigation and mobile interaction;
- bounded provider-state evidence that never exposes raw backend error strings;
- result provenance/provider identity where appropriate;
- safe external-link handling;
- copy/share actions that avoid adding tracking parameters.

## Preferences information architecture

The inherited Preferences surface will be replaced with a first-party settings application. Settings must be organized by user intent rather than inherited implementation modules.

### Search

- default category;
- result density and layout;
- language and region;
- SafeSearch;
- time and filter defaults;
- autocomplete behavior.

### Sources

- provider enable/disable controls;
- category/provider relationships;
- provider availability and degraded-state presentation;
- advanced provider options where supported.

### Appearance

- system, Light, Dark, and Deep Dark appearance;
- Glaze UI density and responsive presentation options;
- result-card and media presentation preferences;
- appropriate application-level accessibility presentation controls.

### Privacy

- local recent-query/history controls;
- autocomplete privacy;
- proxy and media privacy behavior;
- request/data-minimization explanations;
- clear/reset controls for locally retained preference data.

### Security

- Wardveil-backed Search security state where implemented;
- safe-link and external-content protections;
- security diagnostics that do not expose secrets.

### Data & resilience

- preference export, import, and reset;
- schema-versioned portability;
- Everkeep portability and recovery state where implemented.

### Advanced

- Special Queries;
- appropriate developer/diagnostic controls;
- version, build, and provider diagnostics.

Preferences must also support search-within-settings, stable deep links to sections, clear defaults/reset behavior, keyboard accessibility, mobile layouts, and an explicit scope indicator for local, account-scoped, deployment-controlled, and unavailable settings.

## Additional native enhancements

The rebuild should add the following when supported by implementation and acceptance evidence:

- keyboard shortcuts and a command surface for search/category navigation;
- query syntax/help;
- per-category remembered view preferences without behavioral profiling;
- provider health and degraded-state UI;
- privacy-preserving local recent searches with explicit controls;
- preference export/import with schema versioning;
- accessible first-run/onboarding explanation;
- native OpenSearch and GoreeCloud Browser integration;
- installability/PWA support only when compatible with the approved delivery model;
- strong content-security and outbound-request boundaries;
- homepage and results performance budgets;
- first-party components and tokens aligned with the current Glaze UI contract.

## Current Stable Glaze UI V1.1 source-adoption tranche

The native browser source now explicitly targets **GLAZE UI V1.1 / 1.1.0** while remaining a Development application.

The source-level shell adoption includes:

- `data-glaze-version="1.1"` on native Home, Preferences, and Results documents;
- the current Stable Deep Teal (`#0f6b6f`) and Soft Amber (`#d9a35f`) environmental identity, with neutral structure remaining dominant;
- bounded upper-left teal and restrained warm-amber atmospheric treatment rather than the previous violet application identity;
- explicit Light, Dark, and Deep Dark mappings plus system-following behavior;
- Glaze density mapping where Search `comfortable` maps to the Glaze comfortable profile and Search `compact` maps to the Glaze productive profile;
- one small same-origin `appearance.js` bootstrap that reads only the schema-versioned local Search preference envelope and performs no network request;
- a 48 px minimum primary interactive-target contract in shared and results-specific browser acceptance;
- reduced-motion, reduced-transparency, increased-contrast, and forced-colors source fallbacks;
- durable content surfaces that avoid unnecessary nested backdrop-filter materials;
- automated Compact, Medium, Expanded, and Wide browser checks across Light, Dark, and Deep Dark for Home, Preferences, General Results, and Image Results; and
- existing dedicated image-viewer keyboard/touch/focus acceptance alongside the shell checks.

This is **source implementation and automated CI acceptance scope**, not a whole-application or production Glaze conformance claim. Fresh manual contextual visual review, remaining resilience evidence including RTL and 200% text, representative target-form-factor validation, target-environment acceptance, and production acceptance are still required before the `glaze_ui` platform-system result can become conformant.

## Feature-preservation gate

Before an inherited user-facing capability is removed, it must be inventoried and classified as one of:

- `retain` — preserve behavior in the native application;
- `replace` — provide a native equivalent before migration;
- `improve` — replace it with a materially better native capability;
- `retire` — remove only with an explicit project decision and documented rationale.

The inventory includes search categories, filters, provider controls, privacy controls, accessibility behavior, preferences, specialized results, browser integration, import/export behavior, and operational diagnostics.

No inherited feature may disappear merely because the native implementation has not reached it yet.

## Migration boundary

The transitional production implementation remains available until the native replacement satisfies the applicable production-readiness and migration gates. Native development does not itself authorize production deployment or Stable qualification.

Stable remains blocked until required Glaze UI, Wardveil Security, Privacy Shield, Everkeep, provider, runtime, recovery, accessibility, and real-device acceptance evidence is complete.

## Delivery sequence

1. Native web shell and homepage design system.
2. Native Preferences information architecture and preference schema.
3. Provider-adapter contract and retained-feature inventory.
4. Native results pages, categories, filters, and specialized cards.
5. Suggestions and advanced query capabilities.
6. Privacy, security, and resilience integrations.
7. Preference/data portability and migration compatibility.
8. GoreeCloud Browser and OpenSearch integration.
9. Performance, accessibility, and real-device acceptance.
10. Controlled production migration and retirement of the transitional SearXNG-derived runtime.
