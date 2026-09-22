# GoreeCloud Search — Migrated Historical Changelog — Part 02

> **Historical provenance only.** This normalized archive preserves every dated entry from the retired Google Drive `Change Log — Search.docx` in chronological order. Each entry retains source-derived evidence excerpts; current repository state and lifecycle are controlled by the current root records.

**Legacy Drive source ID:** `1uEAtCFrxl8D3HnVxzRInMe92lRAiJKBv`  
**Migration date:** 2026-09-22  
**Historical entries represented:** 11–20 of 77  

## August 18, 2026 at 8:49 AM CDT — Search Landing Page and Preferences Glaze UI Refinement and PR #8 Integration

- Change type or category: GoreeCloud Search user-interface refinement; Glaze UI; landing page; Preferences; accessibility; compact-layout containment; continuous integration; stabilization integration.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/landing-preferences-polish-v1; pull request #8 targeting agent/mobile-stabilization; final integrated stabilization commit a1daa697cea5c8380cdf043b79a4a3f4968aeef4.
- No live production Search deployment was changed.
- Purpose: Continue post-cutover visual stabilization after reviewing the current production landing page and Preferences surface.
- • Browser acceptance run 32143870110 exposed that the newly refined landing search button rendered at approximately 40–43 pixels across adaptive viewport classes, below the 44-pixel minimum.

## August 18, 2026 at 9:08 AM CDT — Preferences Webapp Contract Correction, Ranking Gate Hardening, and PR #9 Integration

- Change type or category: GoreeCloud Search stabilization; upstream unit-test compatibility; continuous-integration hardening; Preferences contract correction.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; follow-up branch agent/landing-preferences-ranking-fix; pull request #9; agent/mobile-stabilization.
- No live production deployment was changed.
- Purpose: Correct the post-merge ranking/upstream-unit regression detected after PR #8 and ensure future landing-page and Preferences changes cannot bypass the reviewed upstream unit-test gate.
- Root cause: The PR #8 Preferences form intentionally added class="goreecloud-preferences-form". tests/unit/test_webapp.py still asserted the older exact opening fragment in which method="post" immediately followed id="search_form".

## August 18, 2026 at 12:26 PM CDT — Preferences Keyboard Accessibility and All-Tab Acceptance PR #10 Integration

- Change type or category: GoreeCloud Search stabilization; Preferences accessibility; keyboard navigation; adaptive browser acceptance; CI validation; source integration.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/preferences-keyboard-acceptance; pull request #10 targeting agent/mobile-stabilization; parent draft pull request #3.
- No live production Search deployment was changed.
- Purpose: Close a remaining Preferences accessibility and automated acceptance gap by ensuring the six top-level settings surfaces are usable through native keyboard navigation and are continuously validated across the full Glaze UI adaptive layout range.
- • Extended the existing browser-acceptance workflow to run the focused Preferences acceptance after the shared adaptive browser suite, and expanded the page-polish source contract to fail closed on loss of the native keyboard semantics.

## August 18, 2026 at 7:51 PM CDT — Deterministic Provider-Failure Degradation Gate Integrated Through PR #11

- Change type or category: Stable-readiness engineering; deterministic provider degradation; browser/search boundary validation; CI hardening; failure-path testing; source integration.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; agent/provider-degradation-acceptance; agent/mobile-stabilization; pull request #11; GitHub Actions; no production runtime changes.
- Purpose: I added a deterministic acceptance gate that proves an individual upstream-style provider failure remains contained inside GoreeCloud Search while a healthy sibling provider still returns a normal result and the Browser/Search privacy boundary does not silently introduce a direct external search fallback.
- Implementation: I added searx/engines/goreecloud_acceptance.py as a disabled-by-default offline acceptance provider, goreecloud/provider_degradation_settings.yml as an isolated two-provider CI runtime, goreecloud/provider_degradation_acceptance.py as the response and fallback validator, and .github/workflows/goreecloud-provider-degradation.yml as the dedicated pull-request, integration-push, and manual acceptance workflow.
- Validation: PR #11 exact head ce13d13113e96aeb67b7b9afde4d612f85297b5e passed GoreeCloud provider degradation run 32202607855.

## August 18, 2026 at 8:01 PM CDT — Immutable Candidate Identity and Image-Level Rollback Evidence Hardening Integrated Through PR #12

- Change type or category: Stable-readiness engineering; release engineering; immutable container identity; rollback evidence; CI hardening; source integration.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; agent/release-evidence-hardening; agent/mobile-stabilization; pull request #12; GitHub Actions; no production runtime changes.
- Purpose: I strengthened the first-Stable release path so an accepted candidate must be bound to its exact source revision and immutable GHCR digest, while preserving a separate, explicit known-good production rollback image identity and preventing image-level rehearsal from being misrepresented as full target-host rollback acceptance.
- Implementation: I added goreecloud/release_baseline.json with the current known-good production GoreeCloud Search source revision 3584da535f7ed7c3b4b8dc73cf0424fb4bdf1949 and immutable image ghcr.io/goreecloud/goreecloud-search@sha256:30ec99e3311fa9dcc934ac267aef123bb2541e0cd0165b360c4a7dc4fa29e3d5.
- I added goreecloud/release_evidence.py to validate the baseline, reject mutable rollback tags, require 40-character source revisions, require GoreeCloud Search GHCR sha256 image references, require candidate OCI revision to equal the exact source SHA, and generate structured release evidence whose rollback scope explicitly keeps target-environment configuration rollback, data restoration, and production cutover authorization false.

## August 18, 2026 at 8:14 PM CDT — Exact Target Runtime Identity Acceptance Integrated Through PR #13

- Change type or category: Stable-readiness engineering; target-environment acceptance; immutable runtime identity; OCI provenance; sanitized evidence; CI hardening; source integration.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; agent/target-runtime-evidence; agent/mobile-stabilization; pull request #13; GitHub Actions; no production runtime changes.
- Purpose: I strengthened target-host acceptance so a staged or future production candidate cannot be accepted merely because a GoreeCloud Search instance is healthy.
- The acceptance procedure can now prove that the running Docker container is the exact immutable GHCR candidate produced from the exact release source revision while keeping backup, data-restore, rollback, and production authorization as separate gates.
- Existing home, Preferences, About, health, privacy-header, and representative-provider acceptance remains intact.

## August 18, 2026 at 8:22 PM CDT — Required Files Real-Provider Coverage Integrated Through PR #14

- Change type or category: Stable-readiness acceptance coverage; real-provider test planning; CI contract hardening; source integration.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; agent/provider-suite-files; agent/mobile-stabilization; pull request #14; GitHub Actions; no production runtime changes.
- Purpose: I corrected a mismatch between the documented first-Stable gate and the executable representative real-provider suite.
- I defined RELEASE_REQUIRED_CATEGORIES as General, Images, Videos, News, and Files and added fail-closed validation so a future edit cannot silently remove one of those categories.
- It validates Python syntax, the mandatory first-Stable category set, unique representative categories, one non-empty Files case, continued Files configuration in searx/settings.yml, and Files visibility in CLI help.

## August 19, 2026 at 6:46 AM CDT — Glaze UI 1.1 Conformance, Stable-Gate Reconciliation, and PR #16 Integration

- Change type or category: GoreeCloud Search source stabilization; Glaze UI 1.1 adoption; accessibility and adaptive-layout hardening; release-governance correction; CI validation; pull-request integration.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/glaze-ui-1.1-conformance; pull request #16 targeting agent/mobile-stabilization; parent draft pull request #3.
- No live production Search deployment or VPS configuration was changed.
- Purpose: I brought the current first-Stable Search stabilization line into explicit conformance with the canonical Glaze UI 1.1.0 semantic contract while preserving the Search-specific composition that had already passed adaptive browser acceptance.
- I also corrected a stale Stable-release provider gate so Files remains a mandatory first-Stable representative category.

## August 19, 2026 at 7:03 AM CDT — Final-Candidate Acceptance Evidence Contract and PR #17 Integration

- Change type or category: First-Stable release governance; candidate-bound evidence; real-provider acceptance; Glaze UI 1.1 final review; GoreeCloud Browser runtime integration; CI hardening; documentation reconciliation.
- Affected project and environment: GoreeCloud Search; GoreeCloud/goreecloud-search; branch agent/final-candidate-evidence; pull request #17 targeting agent/mobile-stabilization; integration commit 1bb26fba7a4b1658442e1dd872a0ea170abe141f.
- No live production Search runtime change was performed.
- Purpose: I closed the remaining source-side evidence-format gap before first-Stable final-candidate execution.
- The repository could already validate source, immutable release identity, target runtime identity, provider coverage, recovery, monitoring, and rollback contracts, but the real-provider runner emitted only console output and the required physical Glaze UI and actual GoreeCloud Browser runtime reviews had no candidate-bound final manifest.

## August 19, 2026 at 7:47 AM CDT — Controlled First-Stable Candidate Publication Path Integrated Repository state - PR #18, “Add explicit first-Stable candidate publication control,” was squash-merged into agent/mobile-stabilization

- . - Exact integrated candidate-request source revision: d00a29f5924073e34bdfadc55454c80888f79b44. - The merge commit is verified and its direct parent is the reviewed stabilization revision 1bb26fba7a4b1658442e1dd872a0ea170abe141f.
- Candidate publication control - Added goreecloud/candidate_request.json as the explicit first-Stable publish-and-rehearse request marker. - The request keeps production cutover, Stable release, and target-host changes unauthorized. - Added a fail-closed validator that rejects malformed identity, reviewed-base drift, candidate-parent drift, unknown fields, authorization drift, and secret-like evidence fields. - Added a marker-only agent/mobile-stabilization push workflow.
- Ordinary stabilization changes do not publish candidate images. - Made the existing immutable candidate-image workflow reusable while preserving its production-acceptance and manual-dispatch entry points. - The reusable path still performs exact-source build, OCI identity validation, GHCR publication, immutable digest retrieval, candidate and known-good rollback image pulls, isolated loopback runtime rehearsal, and release-evidence artifact generation.
- Validation - Initial release-evidence CI failed because the standalone candidate-request test did not include the repository root on Python’s import path.
- The candidate request validator and rollback baseline had already passed. - The test harness path was corrected without weakening any release rule. - Final PR #18 head 33bcbc275fde93e8f1610f2dab8585b666c76d39 passed release-evidence run 32254006150, final-acceptance-evidence run 32254006095, and browser-acceptance run 32254006131. - Full adaptive Glaze UI and Preferences keyboard/all-tab browser acceptance passed on that exact head.
