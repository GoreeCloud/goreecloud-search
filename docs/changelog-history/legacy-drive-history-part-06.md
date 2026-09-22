# GoreeCloud Search — Migrated Historical Changelog — Part 06

> **Historical provenance only.** This normalized archive preserves every dated entry from the retired Google Drive `Change Log — Search.docx` in chronological order. Each entry retains source-derived evidence excerpts; current repository state and lifecycle are controlled by the current root records.

**Legacy Drive source ID:** `1uEAtCFrxl8D3HnVxzRInMe92lRAiJKBv`  
**Migration date:** 2026-09-22  
**Historical entries represented:** 51–60 of 77  

## August 28, 2026 at 7:30 PM CDT — Paginated Native Search Sync Retrieval Integrated

- Change ID: GC-2026-08-28-SEARCH-PAGINATION-01 Component: GoreeCloud Search Type: Native Sync integration; bounded retrieval; fail-closed continuation validation; exact-head source integration Summary: Native Search previously fetched search.history as one bounded response.
- On agent/paginated-sync-retrieval-20260828, PR #87 changed FetchHistory to request 256-record pages using an exclusive after cursor while preserving the bearer credential only at the HTTP boundary.
- Impact: PR #87 exact head 490d91e2537e027570156cf1ed845b3b3cb68b16 passed all nine attached exact-head workflows, including Search Native Foundation, platform integrations, foundation, runtime smoke, upstream container boundary, browser acceptance, Documentation, container build, and Integration.
- With no review comments and a clean merge state, the unchanged candidate was squash-merged with expected-head protection.
- No production deployment or Stable claim occurred.

## August 28, 2026 at 8:55 PM CDT — Authenticated Search Sync Record-ID Contract Integrated

- Change ID: GC-2026-08-28-SEARCH-SYNC-ID-AUTH-01 Component: GoreeCloud Search Type: Native Sync authentication and record-boundary hardening Summary: Native Search now uses one 512-byte Sync record-ID limit across signing, submission, retrieval, and pagination continuation.
- Impact: Branch agent/sync-record-id-auth-contract-20260828, PR #88, exact head 515b3fb9dfd65d30d640706caf85686e8c95b393.
- All nine attached exact-head workflows succeeded: upstream container boundary, foundation, platform integrations, runtime smoke, Search Native Foundation, browser acceptance, Integration, container build, and Documentation.
- PR had no comments and was mergeable.
- Squash merge with expected-head protection produced authoritative master b518894db2ce9ffeec6e9e5658d195fb7cbaeb8e.

## August 28, 2026 at 9:23 PM CDT — Search Sync Envelope Conformance Candidate

- Change ID: GC-2026-08-28-SEARCH-SYNC-ENVELOPE-CANDIDATE-01 Component: Search Type: Native Sync contract hardening; capability alignment; authenticated fail-closed transport; Privacy Shield deletion minimization; review candidate Summary: Branch agent/sync-envelope-conformance-20260828 / PR #89 completes the native search.history Sync envelope boundary on exact head 5962812203bc5445c49f60615b6f465f08b5799e.
- Direction-neutral envelope validation enforces the 512-byte record-ID bound, revision/timestamp/origin metadata, application payload for live records, and payload-free tombstones.
- Validation: Regression coverage was added for oversized tombstone IDs, unnegotiated schemas, invalid outbound envelope shape, invalid inbound schema/payload shape, and valid payload-free tombstones.
- Fresh exact-head Search Native Foundation, Integration, runtime, foundation, platform-integration, upstream-boundary, container-build, browser-acceptance, and Documentation runs are queued at this checkpoint; no exact-head success is claimed and PR #89 remains open/unmerged pending the complete applicable matrix.
- No production deployment, migration, or Stable compatibility claim is made.

## August 28, 2026 at 9:50 PM CDT — Search Sync Exact Envelope Conformance Integrated

- Change ID: GC-2026-08-28-SEARCH-SYNC-ENVELOPE-01 Component: GoreeCloud Search Type: Native Sync contract hardening; authenticated submission; Privacy Shield data minimization; exact-schema enforcement; pull-request integration Summary: PR #89 centralized the search.history Sync envelope boundary around Search's advertised capability.
- Gmail surfaced the failed GitHub notification; exact-head job logs confirmed the Go compile error.
- Evidence: Final exact PR head c2238f67bc53f94225c15e143b10f17967c11f84 passed all nine applicable workflows: upstream container boundary #291, platform integrations #279, foundation #996, container build #911, Search Native Foundation #38, runtime smoke #915, browser acceptance #1004, Documentation #326, and Integration #326.
- PR #89 had no review comments and was squash-merged with expected-head protection as authoritative master 8cbe5137bac382117fdc4f0f26d65f12e3b164d0.
- Final state: Native Search now fails closed on unnegotiated schema, invalid envelope metadata/payload shape, oversized deletion IDs, unauthenticated submission, and incompatible operation capability.

## August 28, 2026 at 9:56 PM CDT — Canonical Native-First Search README Integrated

- Change ID: GC-2026-08-28-SEARCH-README-NATIVE-01 Component: GoreeCloud Search Type: Repository documentation; native-migration governance; platform-gate alignment; transitional compatibility documentation Summary: PR #90 added the required root README.md as the authoritative GoreeCloud Search lifecycle and architecture record.
- It now describes original native GoreeCloud Search as the long-term application, classifies the SearXNG-derived tree as a transitional continuity/migration/compatibility dependency, documents implemented native Go source and Sync protections, and explicitly makes Glaze UI, Wardveil Security, Privacy Shield, and Everkeep mandatory Stable gates.
- Reason: The repository previously had only README.rst, whose substantive content described a maintained SearXNG fork and production/RC posture inconsistent with the current native-application requirement and live README policy.
- Exact-head Foundation and container-build failures exposed two legitimate transitional dependencies: legacy validation still checks the filename and setup.py uses it for SearXNG package long_description.
- Rather than restore stale duplicate product documentation or weaken CI, README.rst was retained as a minimal compatibility shim that explicitly points to README.md as the sole authoritative lifecycle/architecture record.

## August 28, 2026 — 10:10 PM CDT

- Change ID: GC-2026-08-28-SEARCH-SYNC-PROOF-01 Component: GoreeCloud Search / native Sync submission Type: Security / protocol conformance hardening Summary: Added fail-fast validation of caller-supplied Sync RecordProof values before Search submission transport and merged PR #91 as 4954c691d4f663bcedd9c0d5cb2a3b3de15aeb6d.
- Reason: Search already validated bearer state and envelope/capability conformance, but a malformed, mismatched, or record-invalid proof could still reach the Sync server before being rejected there.
- Implementation: On exact head 84bbd490172249b08d3e4b60469aaadcc1035d0b, native/internal/syncstate/submission.go now requires proof deviceId to equal envelope originDevice, raw-URL Ed25519 public-key/signature decoding with exact sizes, and signature verification over the exact GC-SYNC-RECORD/1 message before transport.
- Server-side authenticated-peer, key-fingerprint, replay, Privacy Shield, and Wardveil verification remains authoritative.
- Validation: Exact-head workflows all succeeded: Search Native Foundation #39, platform integrations #291, upstream container boundary #301, Integration #331, foundation #1008, runtime smoke #925, container build #921, browser acceptance #1014, and Documentation #331.

## August 28, 2026 — 10:24 PM CDT

- Change ID: GC-2026-08-28-SEARCH-SYNC-PROOF-VECTOR-01 Component: GoreeCloud Search / native Sync interoperability tests Type: Test / protocol conformance hardening Summary: Pinned Search proof-message and verification behavior to the canonical GoreeCloud Sync GC-SYNC-RECORD/1 test vector and merged PR #92 as 0376abed29dd6534aa507e7e6365035562f24ca1.
- Production Search code was unchanged.
- PR #92 was held until authoritative Sync PR #14 merged as 9a20390e4ef0db6ab26a8d08fdb9310349477e8e.
- Validation: Exact head 3e7d4e2c0d7bf22b625b40223a27a4f720d401bd passed all nine workflows: Search Native Foundation #40, foundation #1012, upstream container boundary #304, platform integrations #295, runtime smoke #928, container build #924, browser acceptance #1017, Documentation #333, and Integration #333.
- PR discussion was clean and master remained 4954c691d4f663bcedd9c0d5cb2a3b3de15aeb6d at merge.

## August 30, 2026 — 10:00 PM CDT

- Change ID: GC-2026-08-30-SEARCH-NATIVE-PROVIDER-CONTRACT-01 Component: Search Type: Native provider trust boundary / documentation reconciliation Summary: Hardened the native Search provider/result boundary and reconciled the repository with the current native-first architecture; merged PR #93 as b75e78394005485aca4ccd005eb6cc753d19b611.
- Implementation: Rejected blank/control-character provider identities before execution/status/result evidence; excluded invalid identities from deterministic provider definitions; rejected result URLs containing embedded user-info credentials while retaining scheme/host validation, fragment stripping, deterministic deduplication/ranking, category gating, and bounded provider execution; added regressions.
- Impact: Native Search has a stricter provider/result trust boundary and repository-wide product documentation now reflects native-first migration without claiming production provider readiness.
- Validation: Exact head 9a256f0c3576181da567d86021dd1522468a17fc passed all nine applicable exact-head workflows: upstream container boundary, platform integrations, foundation, runtime smoke, Search Native Foundation, Documentation, Integration, container build, and browser acceptance.
- PR was mergeable and had no review comments before expected-head squash merge.

## August 30, 2026 — Central Search user-manual synchronization

- Change ID: GC-2026-08-30-SEARCH-CENTRAL-USER-MANUAL-01 Component: GoreeCloud Search Type: Documentation Summary: Created the required central User Manual — Search in GoreeCloud/User Manuals and synchronized it with the merged repository USER-MANUAL.md behavior.
- Reason: The merged native provider-contract slice added and reconciled the repository user manual, but the required central GoreeCloud/User Manuals copy was still missing.
- Impact: The central manual now describes the same current user-facing boundaries as the merged repository manual: native migration/pre-Stable status, configured Search authority without silent unrelated-provider fallback, provider/category availability, result URL safety, preferences/history Sync limitations, Privacy Shield expectations, Glaze UI/accessibility requirements, troubleshooting, and migration/recovery boundaries.
- Evidence: Repository USER-MANUAL.md is present at merged Search commit b75e78394005485aca4ccd005eb6cc753d19b611.
- Final state: Search now satisfies the two-location user-manual requirement for the current merged product behavior.

## 2026-08-31 — Draft PR #95: Native ranking v2 and results-experience overhaul Status: Source-level work under review on branch agent/native-ranking-results-v2

- .
- This entry does not establish merge, deployment, Stable promotion, Glaze UI 2.0 rendered/device acceptance, or production approval.
- • Production-approved provider adapters and full category coverage.
- • Exact-head CI completion and Glaze UI 2.0 rendered/accessibility/device acceptance.
