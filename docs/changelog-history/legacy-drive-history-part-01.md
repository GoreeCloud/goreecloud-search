# GoreeCloud Search — Migrated Historical Changelog — Part 01

> **Historical provenance only.** This normalized archive preserves every dated entry from the retired Google Drive `Change Log — Search.docx` in chronological order. Each entry retains source-derived evidence excerpts; current repository state and lifecycle are controlled by the current root records.

**Legacy Drive source ID:** `1uEAtCFrxl8D3HnVxzRInMe92lRAiJKBv`  
**Migration date:** 2026-09-22  
**Historical entries represented:** 1–10 of 77  

## August 16, 2026 at 4:06 PM CDT — Initial GoreeCloud Search Maintained-Fork Foundation, Glaze UI Product Shell, CI, and Draft PR #1

- Change type or category: Maintained open-source fork development; repository foundation; upstream provenance; Glaze UI; accessibility; continuous integration; source validation; product documentation.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/stable-foundation; draft pull request #1; SearXNG upstream baseline; no production runtime changes.
- Initial upstream baseline: searxng/searxng commit b2da6b90f2f8446557c91f67d6be5064ab785ecd.
- The baseline was selected by exact commit SHA because the upstream repository did not expose a conventional GitHub release or tag sequence suitable as the initial GoreeCloud source boundary.
- • Created development branch agent/stable-foundation from the exact upstream/fork master baseline.

## August 16, 2026 — GoreeCloud Search Results Experience, Runtime Baseline, API Boundary, and Deployment Preparation

- Change type or category: Glaze UI expansion; search-results experience; runtime configuration; API governance; deployment documentation; continuous integration.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/stable-foundation; draft pull request #1; no production runtime changes.
- • Added docs/goreecloud/DEPLOYMENT.md to define Docker/private-service deployment expectations and the required production-acceptance checks before replacing the existing SearXNG runtime.
- • Updated draft pull request #1 to describe the expanded product layer and remaining production-acceptance work.
- Validation state: The branch remains mergeable.

## August 16, 2026 — GoreeCloud Search Runtime Smoke, Container Acceptance, Compose Baseline, and OCI Product Metadata

- Change type or category: Runtime validation; container build acceptance; deployment baseline; OCI metadata; compatibility correction; continuous integration.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/stable-foundation; draft pull request #1; GitHub Actions acceptance environment; no production runtime changes.
- • Runtime validation detected that current SearXNG requires brand.docs_url, brand.public_instances, and brand.wiki_url to be strings.
- Corrected the GoreeCloud settings example from boolean disable values to empty strings before deployment.
- • Added .github/workflows/goreecloud-container-build.yml to build the maintained fork from source, run the resulting image, validate product identity, verify the health endpoint, and inspect GoreeCloud OCI metadata.

## August 16, 2026 — GoreeCloud Search Glaze UI 1.0 Source-Stable Finalization, Maintained-Fork Governance, and Full Acceptance

- Change type or category: Glaze UI finalization; accessibility and adaptive layout; privacy and security hardening; browser/PWA/OpenSearch identity; maintained-fork governance; provider acceptance; localization/product identity; continuous integration; source stabilization.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/stable-foundation; pull request #1; exact source candidate b07728e1d882772006e1f9e46e04b4a18794bbe5; GitHub Actions acceptance environment; no production runtime changes.
- Purpose: Complete the first source-stable GoreeCloud Search release candidate as a comprehensively rebranded, privacy-oriented, Glaze UI 1.0 maintained fork of SearXNG while preserving upstream attribution, upstream synchronization, and a separate production-authorization boundary.
- • Expanded manual provider acceptance into a representative general, images, news, videos, IT, and science suite with single-category diagnostic mode while keeping external-provider behavior out of deterministic pull-request CI.
- • Replaced upstream-only CONTRIBUTING.rst with a maintained-fork contribution model that distinguishes GoreeCloud-owned product/Glaze/deployment work from generally applicable upstream-capable engine and backend changes.

## August 17, 2026 at 6:10 PM CDT — GoreeCloud Search Production Cutover, Immutable Image Validation, Private-Path Acceptance, and Secret Rotation

- Change type or category: Production migration; maintained-fork deployment; immutable Docker image; rollback preparation; private DNS and HTTPS validation; secret rotation; stabilization.
- Purpose: Replace the active upstream SearXNG application image with the validated GoreeCloud Search maintained-fork image while preserving the existing private-service topology, persistent configuration, Valkey dependency, Caddy route, DNS rewrite, and rollback capability.
- • Created a pre-migration rollback snapshot at /srv/docker/backups/goreecloud-search/pre-migration-20260817-091213 containing the SearXNG stack configuration, protected environment-file copy, application settings, Caddyfile copy, container inspection records, upstream image inspection record, and SHA-256 checksums.
- • Verified the existing production topology before cutover: searxng-core and searxng-valkey under /srv/docker/stacks/searxng; core attached to searxng-internal and proxy; Valkey isolated to searxng-internal; no host-published Search backend port; Caddy reverse proxy target searxng-core:8080.
- • Pulled the exact GoreeCloud Search candidate ghcr.io/goreecloud/goreecloud-search:3584da535f7ed7c3b4b8dc73cf0424fb4bdf1949 and resolved immutable registry digest sha256:30ec99e3311fa9dcc934ac267aef123bb2541e0cd0165b360c4a7dc4fa29e3d5.

## August 17, 2026 at 8:40 PM CDT — Mobile Stabilization, Shared Glaze UI Compact Shell Correction, Browser Integration Contract, and Stable-Cutover Gating

- Change type or category: GoreeCloud Search stabilization; Glaze UI compact/mobile layout; browser acceptance; GoreeCloud Browser integration contract; Stable-release governance; GitHub Actions.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/mobile-stabilization; draft pull request #3 targeting agent/production-acceptance; searx/static/themes/simple/goreecloud-mobile.css; docs/goreecloud/BROWSER-INTEGRATION.md; docs/goreecloud/STABLE-CUTOVER.md; GoreeCloud browser-acceptance workflow.
- No live VPS runtime rename or removal was performed by this work.
- Purpose: Continue first-Stable stabilization after production cutover by eliminating narrow-viewport page-level overflow, formalizing the Search-side contract required for GoreeCloud Browser, and making Browser/Search integration an explicit Stable-release gate.
- • Mobile stabilization had already corrected Preferences-specific overflow, but exact-head browser acceptance subsequently exposed the same inherited compact-page offset on the About surface.

## August 18, 2026 — Deterministic Ranking and Result-Interface Improvement Follow-Up

- Change type or category: GoreeCloud Search ranking; result-interface refinement; Glaze UI; deterministic relevance scoring; source validation; Browser integration support.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/search-ranking-results-v2; pull request #4 targeting agent/mobile-stabilization.
- No live Search production deployment was changed.
- Changes present in pull request #4:
- Validation follow-up completed in this conversation:

## August 18, 2026 at 12:57 AM CDT — Ranking and Result-Interface PR #4 Integrated into Search Stabilization

- Change type or category: GoreeCloud Search ranking; result-interface refinement; Glaze UI; deterministic relevance; source validation; stabilization integration.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; pull request #4; source head a303067fc1ddbd6feab845205d81cd3840f7b0f5; target branch agent/mobile-stabilization.
- No live production Search deployment was changed.
- Changes and validation completed:
- • Exact-head GoreeCloud browser-acceptance workflow run 32102949377 completed successfully.

## August 18, 2026 at 7:20 AM CDT — Search Query Shell, Category Navigation, Iconography, and Glaze UI Refinement

- Change type or category: GoreeCloud Search user-interface refinement; Glaze UI; search masthead; query field; category navigation; iconography; adaptive layout; accessibility; continuous integration; stabilization.
- No live production Search deployment was changed by this source work.
- Purpose: Correct production visual-review issues in the results-page search area by improving the structure, alignment, iconography, interaction hierarchy, and responsive organization of the query bar, product identity, category navigation, and filters while preserving the SearXNG metasearch backend and GoreeCloud privacy boundary.
- • Rebuilt the results/search masthead around an explicit Glaze UI composition that aligns the GoreeCloud Search wordmark and query field to one stable layout boundary rather than relying on inherited upstream positioning.
- • The first browser-validation attempt invoked pytest even though the existing acceptance environment did not install pytest.

## August 18, 2026 at 8:31 AM CDT — Search Filters, File Results, and Video Results Glaze UI Polish

- Change type or category: GoreeCloud Search user-interface refinement; Glaze UI; search filters; specialized result templates; file results; video results; accessibility; source validation.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/result-card-polish-v3; pull request #7 targeting agent/mobile-stabilization.
- No live production deployment was changed.
- Purpose: Continue post-cutover visual stabilization after reviewing current General, Files, and Videos searches in production.
- • Reworked Videos results from inherited float geometry into an explicit responsive media-card grid with a stable 16:9 thumbnail, separate title/metadata/action/description/source regions, readable duration overlays, and single-column compact behavior.
